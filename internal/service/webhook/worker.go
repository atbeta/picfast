package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/atbeta/picfast/internal/sqlc"
)

type Worker struct {
	db          *sqlc.Queries
	delivery    *DeliveryService
	interval    time.Duration
	concurrency int
	quit        chan struct{}
	wg          sync.WaitGroup
}

func NewWorker(db *sqlc.Queries, delivery *DeliveryService, interval time.Duration, concurrency int) *Worker {
	return &Worker{
		db:          db,
		delivery:    delivery,
		interval:    interval,
		concurrency: concurrency,
		quit:        make(chan struct{}),
	}
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go w.run()
}

func (w *Worker) Stop() {
	close(w.quit)
	w.wg.Wait()
}

func (w *Worker) run() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.quit:
			return
		case <-ticker.C:
			w.tick()
		}
	}
}

func (w *Worker) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.processOutbox(ctx); err != nil {
		slog.Warn("webhook worker: process outbox failed", "error", err)
	}

	w.dispatchPending(ctx)
}

func (w *Worker) processOutbox(ctx context.Context) error {
	events, err := w.db.ListPendingOutboxEvents(ctx, 100)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := w.createEventDeliveries(ctx, event); err != nil {
			slog.Warn("webhook worker: create deliveries failed", "event_id", event.ID, "error", err)
		}
	}

	return nil
}

func (w *Worker) createEventDeliveries(ctx context.Context, event sqlc.OutboxEvent) error {
	allWebhooks, err := w.db.ListEnabledWebhooks(ctx)
	if err != nil {
		return err
	}

	if len(allWebhooks) == 0 {
		w.db.MarkOutboxEventDispatched(ctx, event.ID)
		return nil
	}

	var matchingWebhooks []sqlc.Webhook
	for _, wh := range allWebhooks {
		if !eventOwnerMatches(event, wh) {
			continue
		}
		var subscribed []string
		if err := json.Unmarshal(wh.Events, &subscribed); err != nil {
			continue
		}
		if len(subscribed) == 0 {
			continue
		}
		if EventMatches(subscribed, event.Type) {
			matchingWebhooks = append(matchingWebhooks, wh)
		}
	}

	if len(matchingWebhooks) == 0 {
		w.db.MarkOutboxEventDispatched(ctx, event.ID)
		return nil
	}

	headersJSON, _ := json.Marshal(map[string]string{})

	for _, wh := range matchingWebhooks {
		_, err := w.db.CreateWebhookDelivery(ctx, sqlc.CreateWebhookDeliveryParams{
			WebhookID:      wh.ID,
			OutboxEventID:  event.ID,
			RequestHeaders: headersJSON,
		})
		if err != nil {
			slog.Warn("webhook worker: failed to create delivery", "webhook_id", wh.ID, "event_id", event.ID, "error", err)
		}
	}

	w.db.MarkOutboxEventDispatched(ctx, event.ID)
	return nil
}

func eventOwnerMatches(event sqlc.OutboxEvent, wh sqlc.Webhook) bool {
	if !event.OwnerUserID.Valid {
		return false
	}
	if event.OwnerUserID.Int64 == wh.UserID {
		return true
	}
	return false
}

func (w *Worker) dispatchPending(ctx context.Context) {
	limit := int32(w.concurrency)
	if limit < 1 {
		limit = 10
	}
	deliveries, err := w.db.ListPendingWebhookDeliveries(ctx, limit)
	if err != nil {
		slog.Warn("webhook worker: list pending deliveries failed", "error", err)
		return
	}

	bgCtx := context.Background()
	sem := make(chan struct{}, w.concurrency)

	for _, d := range deliveries {
		d := d
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			if err := w.delivery.Dispatch(bgCtx, d); err != nil {
				slog.Warn("webhook worker: dispatch error", "error", err)
			}
		}()
	}
}
