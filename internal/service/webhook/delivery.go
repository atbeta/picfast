package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var retryDelays = []time.Duration{
	0,
	1 * time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	12 * time.Hour,
}

type DeliveryService struct {
	db             *sqlc.Queries
	timeout        time.Duration
	allowHTTP      bool
	allowPrivateIP bool
	encKey         []byte
}

func NewDeliveryService(db *sqlc.Queries, timeout time.Duration, allowHTTP, allowPrivateIP bool, encKey []byte) *DeliveryService {
	return &DeliveryService{
		db:             db,
		timeout:        timeout,
		allowHTTP:      allowHTTP,
		allowPrivateIP: allowPrivateIP,
		encKey:         encKey,
	}
}

func (d *DeliveryService) ProcessPendingDeliveries(ctx context.Context, batchSize int32) (int, error) {
	deliveries, err := d.db.ListPendingWebhookDeliveries(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("list pending deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		if err := d.dispatch(ctx, delivery); err != nil {
			slog.Warn("webhook dispatch failed", "delivery_id", delivery.ID, "error", err)
		}
	}

	return len(deliveries), nil
}

func (d *DeliveryService) Redeliver(ctx context.Context, deliveryID int64) error {
	if err := d.db.ResetWebhookDeliveryForReplay(ctx, deliveryID); err != nil {
		return fmt.Errorf("reset delivery: %w", err)
	}
	delivery, err := d.db.GetWebhookDeliveryByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("get delivery: %w", err)
	}
	return d.dispatch(ctx, delivery)
}

func (d *DeliveryService) TestDispatch(ctx context.Context, webhookID int64) error {
	wh, err := d.db.GetWebhookByID(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("get webhook: %w", err)
	}
	if !wh.Enabled {
		return fmt.Errorf("webhook is disabled")
	}
	if err := d.validateURL(wh.Url); err != nil {
		return err
	}

	envelope := events.BuildWebhookTest()
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	plainSecret, err := DecryptSecret(wh.SecretCiphertext, d.encKey)
	if err != nil {
		return fmt.Errorf("decrypt secret: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := ComputeSignature(plainSecret, timestamp, body)

	deliveryCtx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(deliveryCtx, http.MethodPost, wh.Url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PicFast-Webhook/1.0")
	req.Header.Set("X-PicFast-Event-Id", envelope.ID)
	req.Header.Set("X-PicFast-Event-Type", envelope.Type)
	req.Header.Set("X-PicFast-Timestamp", timestamp)
	req.Header.Set("X-PicFast-Signature", sig)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send test: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (d *DeliveryService) Dispatch(ctx context.Context, delivery sqlc.WebhookDelivery) error {
	return d.dispatch(ctx, delivery)
}

func (d *DeliveryService) dispatch(ctx context.Context, delivery sqlc.WebhookDelivery) error {
	wh, err := d.db.GetWebhookByID(ctx, delivery.WebhookID)
	if err != nil {
		return fmt.Errorf("get webhook: %w", err)
	}
	if !wh.Enabled {
		d.db.UpdateWebhookDeliveryStatus(ctx, sqlc.UpdateWebhookDeliveryStatusParams{
			ID:           delivery.ID,
			Status:       "dead",
			Attempt:      delivery.Attempt,
			NextRetryAt:  time.Now(),
			ErrorMessage: "webhook disabled",
		})
		return nil
	}

	outboxEvent, err := d.db.GetOutboxEventByID(ctx, delivery.OutboxEventID)
	if err != nil {
		return fmt.Errorf("get outbox event: %w", err)
	}

	var envelope events.Envelope
	if err := json.Unmarshal(outboxEvent.Payload, &envelope); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if err := d.validateURL(wh.Url); err != nil {
		d.failDelivery(ctx, delivery, err.Error())
		return nil
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	plainSecret, err := DecryptSecret(wh.SecretCiphertext, d.encKey)
	if err != nil {
		d.failDelivery(ctx, delivery, fmt.Sprintf("decrypt secret: %v", err))
		return nil
	}
	sig := ComputeSignature(plainSecret, timestamp, body)

	deliveryCtx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(deliveryCtx, http.MethodPost, wh.Url, bytes.NewReader(body))
	if err != nil {
		d.failDelivery(ctx, delivery, err.Error())
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PicFast-Webhook/1.0")
	req.Header.Set("X-PicFast-Event-Id", envelope.ID)
	req.Header.Set("X-PicFast-Event-Type", envelope.Type)
	req.Header.Set("X-PicFast-Delivery-Id", strconv.FormatInt(delivery.ID, 10))
	req.Header.Set("X-PicFast-Timestamp", timestamp)
	req.Header.Set("X-PicFast-Signature", sig)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	durationMs := time.Since(start).Milliseconds()

	if err != nil {
		d.failDelivery(ctx, delivery, err.Error())
		return nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	respBodyStr := string(respBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		d.db.UpdateWebhookDeliveryStatus(ctx, sqlc.UpdateWebhookDeliveryStatusParams{
			ID:             delivery.ID,
			Status:         "delivered",
			Attempt:        delivery.Attempt,
			NextRetryAt:    time.Now(),
			ResponseStatus: pgtype.Int4{Int32: int32(resp.StatusCode), Valid: true},
			ResponseBody:   respBodyStr,
			ErrorMessage:   "",
			DurationMs:     pgtype.Int4{Int32: int32(durationMs), Valid: true},
			Column9:        true,
		})
	} else {
		d.failDelivery(ctx, delivery, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, respBodyStr))
	}

	return nil
}

func (d *DeliveryService) failDelivery(ctx context.Context, delivery sqlc.WebhookDelivery, errMsg string) {
	nextAttempt := int(delivery.Attempt) + 1
	var nextRetry time.Time
	if nextAttempt-1 < len(retryDelays) {
		nextRetry = time.Now().Add(retryDelays[nextAttempt-1])
	}

	if nextAttempt > int(delivery.MaxAttempts) {
		d.db.UpdateWebhookDeliveryStatus(ctx, sqlc.UpdateWebhookDeliveryStatusParams{
			ID:           delivery.ID,
			Status:       "dead",
			Attempt:      int32(nextAttempt),
			NextRetryAt:  nextRetry,
			ErrorMessage: errMsg,
		})
		return
	}

	d.db.UpdateWebhookDeliveryStatus(ctx, sqlc.UpdateWebhookDeliveryStatusParams{
		ID:           delivery.ID,
		Status:       "retrying",
		Attempt:      int32(nextAttempt),
		NextRetryAt:  nextRetry,
		ErrorMessage: errMsg,
	})
}

func (d *DeliveryService) validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" && !(d.allowHTTP && u.Scheme == "http") {
		return fmt.Errorf("only https URLs are allowed")
	}
	if d.allowPrivateIP {
		return nil
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host")
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("private IP not allowed: %s", host)
		}
	} else {
		ips, err := net.LookupIP(host)
		if err != nil {
			return fmt.Errorf("failed to resolve host: %w", err)
		}
		for _, resolved := range ips {
			if resolved.IsLoopback() || resolved.IsPrivate() || resolved.IsLinkLocalUnicast() {
				return fmt.Errorf("resolved to private IP: %s", resolved.String())
			}
		}
	}
	return nil
}

func (d *DeliveryService) ListByWebhook(ctx context.Context, webhookID int64, limit, offset int32) ([]sqlc.WebhookDelivery, error) {
	return d.db.ListWebhookDeliveriesByWebhook(ctx, sqlc.ListWebhookDeliveriesByWebhookParams{
		WebhookID: webhookID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (d *DeliveryService) CountByWebhook(ctx context.Context, webhookID int64) (int64, error) {
	return d.db.CountWebhookDeliveriesByWebhook(ctx, webhookID)
}

func (d *DeliveryService) GetByID(ctx context.Context, id int64) (sqlc.WebhookDelivery, error) {
	return d.db.GetWebhookDeliveryByID(ctx, id)
}

func (d *DeliveryService) GetByIDAndUser(ctx context.Context, deliveryID, userID int64) (sqlc.WebhookDelivery, error) {
	delivery, err := d.db.GetWebhookDeliveryByID(ctx, deliveryID)
	if err != nil {
		return sqlc.WebhookDelivery{}, err
	}
	wh, err := d.db.GetWebhookByID(ctx, delivery.WebhookID)
	if err != nil {
		return sqlc.WebhookDelivery{}, err
	}
	if wh.UserID != userID {
		return sqlc.WebhookDelivery{}, fmt.Errorf("delivery not found")
	}
	return delivery, nil
}
