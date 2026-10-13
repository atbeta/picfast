package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atbeta/picfast/internal/pglock"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
)

// outboxBatchSize caps how many pending outbox events one pass converts to
// deliveries.
const outboxBatchSize = 100

// Result summarizes a single RunOnce pass.
type Result struct {
	// Skipped is true when another instance or run already held the
	// webhook advisory lock.
	Skipped bool `json:"skipped"`
	// More is true when either list filled its batch, meaning more pending
	// rows are likely waiting for another pass. Rows that keep failing do
	// not set More — looping on them would spin the drain budget.
	More                bool `json:"more"`
	OutboxProcessed     int  `json:"outbox_processed"`
	DeliveriesAttempted int  `json:"deliveries_attempted"`
}

type Worker struct {
	db             *sqlc.Queries
	pool           *pgxpool.Pool
	delivery       *DeliveryService
	interval       time.Duration
	concurrency    int
	enableOCR      bool
	ocrConcurrency int
	quit           chan struct{}
	wg             sync.WaitGroup
}

func NewWorker(db *sqlc.Queries, pool *pgxpool.Pool, delivery *DeliveryService, interval time.Duration, concurrency int, enableOCR bool, ocrConcurrency int) *Worker {
	return &Worker{
		db:             db,
		pool:           pool,
		delivery:       delivery,
		interval:       interval,
		concurrency:    concurrency,
		enableOCR:      enableOCR,
		ocrConcurrency: ocrConcurrency,
		quit:           make(chan struct{}),
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

	if _, err := w.RunOnce(ctx); err != nil {
		slog.Warn("webhook worker: run failed", "error", err)
	}
}

// RunOnce performs one full pass — pending outbox events then pending
// deliveries — under a Postgres advisory lock so concurrent instances (or
// overlapping invocations on this instance) cannot dispatch the same
// delivery twice. It waits for every spawned goroutine before returning, so
// a delivery row is no longer pending once RunOnce completes.
func (w *Worker) RunOnce(ctx context.Context) (Result, error) {
	var res Result
	acquired, err := pglock.TryWithLock(ctx, w.pool, pglock.KeyJobsWebhooks, func(ctx context.Context) error {
		return w.runPass(ctx, &res)
	})
	if err != nil {
		return res, err
	}
	res.Skipped = !acquired
	return res, nil
}

func (w *Worker) runPass(ctx context.Context, res *Result) error {
	var retErr error
	if err := w.processOutbox(ctx, res); err != nil {
		slog.Warn("webhook worker: process outbox failed", "error", err)
		retErr = err
	}
	w.dispatchPending(ctx, res)
	return retErr
}

func (w *Worker) processOutbox(ctx context.Context, res *Result) error {
	events, err := w.db.ListPendingOutboxEvents(ctx, outboxBatchSize)
	if err != nil {
		return err
	}
	res.OutboxProcessed += len(events)
	if len(events) == outboxBatchSize {
		res.More = true
	}

	ocrSem := make(chan struct{}, w.ocrConcurrency)
	var wg sync.WaitGroup

	for _, event := range events {
		if w.enableOCR {
			event := event
			ocrSem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-ocrSem }()
				service.DummyOCRProcessor(w.db, event)
			}()
		}

		if err := w.createEventDeliveries(ctx, event); err != nil {
			slog.Warn("webhook worker: create deliveries failed", "event_id", event.ID, "error", err)
		}
	}

	wg.Wait()
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

func (w *Worker) dispatchPending(ctx context.Context, res *Result) {
	limit := int32(w.concurrency)
	if limit < 1 {
		limit = 10
	}
	deliveries, err := w.db.ListPendingWebhookDeliveries(ctx, limit)
	if err != nil {
		slog.Warn("webhook worker: list pending deliveries failed", "error", err)
		return
	}
	res.DeliveriesAttempted += len(deliveries)
	if int32(len(deliveries)) == limit {
		res.More = true
	}

	// Dispatches run detached from ctx: each HTTP request carries its own
	// delivery timeout and the status update must still be written when a
	// caller's deadline fires mid-pass. The WaitGroup keeps the advisory
	// lock held until every dispatch finishes.
	bgCtx := context.Background()
	sem := make(chan struct{}, w.concurrency)
	var wg sync.WaitGroup

	for _, d := range deliveries {
		d := d
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := w.delivery.Dispatch(bgCtx, d); err != nil {
				slog.Warn("webhook worker: dispatch error", "error", err)
			}
		}()
	}
	wg.Wait()
}
