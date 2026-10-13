// Package jobs runs PicFast's background work (webhook dispatch, hourly
// cleanup) either embedded in the server process or on demand via
// POST /internal/jobs/run and `picfast jobs run`, so the same jobs can be
// scheduled externally on platforms without persistent processes.
package jobs

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/pglock"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/service/webhook"
	"github.com/atbeta/picfast/internal/sqlc"
)

// webhookDeliveryRetentionDays is the retention for webhook delivery logs
// and outbox events, pruned by the hourly cleanup task.
const webhookDeliveryRetentionDays = 90

// webhookBudget bounds the delivery passes of a single RunOnce.
const webhookBudget = 45 * time.Second

const (
	stateRan      = "ran"
	stateSkipped  = "skipped"
	stateDisabled = "disabled"
)

// TaskReport describes the outcome of one background task in a RunOnce pass.
type TaskReport struct {
	State      string           `json:"state"` // ran, skipped, or disabled
	Counts     map[string]int64 `json:"counts,omitempty"`
	Error      string           `json:"error,omitempty"`
	DurationMS int64            `json:"duration_ms"`
}

// Report is the JSON-serializable result of Runner.RunOnce.
type Report struct {
	Webhooks TaskReport `json:"webhooks"`
	Cleanup  TaskReport `json:"cleanup"`
}

// Failed reports whether any task ended with an error.
func (r Report) Failed() bool {
	return r.Webhooks.Error != "" || r.Cleanup.Error != ""
}

// Runner owns the background jobs and their embedded scheduling.
type Runner struct {
	queries  *sqlc.Queries
	pool     *pgxpool.Pool
	cfg      *config.Config
	deleter  *service.DeleteService
	delivery *webhook.DeliveryService
	worker   *webhook.Worker

	started     atomic.Bool
	cleanupStop chan struct{}
	wg          sync.WaitGroup
}

// NewRunner builds the services shared by webhook dispatch and cleanup.
func NewRunner(queries *sqlc.Queries, pool *pgxpool.Pool, cfg *config.Config) *Runner {
	emitter := events.NewOutboxEmitter(queries)
	deleter := service.NewDeleteService(queries, pool, cfg.Storage.ThumbnailDir, emitter)
	delivery := webhook.NewDeliveryService(queries, cfg.Webhook.DeliveryTimeout, cfg.Webhook.AllowHTTP, cfg.Webhook.AllowPrivateURLs, cfg.SecretKeyBytes())
	worker := webhook.NewWorker(queries, pool, delivery, cfg.Webhook.WorkerInterval, 20, cfg.App.EnableDummyOCR, 5)
	return &Runner{
		queries:     queries,
		pool:        pool,
		cfg:         cfg,
		deleter:     deleter,
		delivery:    delivery,
		worker:      worker,
		cleanupStop: make(chan struct{}),
	}
}

// Delivery exposes the shared delivery service so HTTP handlers reuse the
// same configuration (timeouts, secret key) as the worker.
func (r *Runner) Delivery() *webhook.DeliveryService { return r.delivery }

// RunOnce runs one webhook pass (when enabled) and one cleanup pass, and
// reports both. Every task is serialized across instances by advisory locks.
func (r *Runner) RunOnce(ctx context.Context) Report {
	return Report{
		Webhooks: r.runWebhooks(ctx),
		Cleanup:  r.Cleanup(ctx),
	}
}

// runWebhooks drains the webhook queue until a pass finds nothing to do or
// the budget elapses.
func (r *Runner) runWebhooks(ctx context.Context) TaskReport {
	start := time.Now()
	rep := TaskReport{State: stateRan, Counts: map[string]int64{}}
	defer func() { rep.DurationMS = time.Since(start).Milliseconds() }()

	if !r.cfg.Webhook.Enabled {
		rep.State = stateDisabled
		return rep
	}

	deadline := start.Add(webhookBudget)
	for {
		res, err := r.worker.RunOnce(ctx)
		rep.Counts["outbox_processed"] += int64(res.OutboxProcessed)
		rep.Counts["deliveries_attempted"] += int64(res.DeliveriesAttempted)
		if err != nil {
			rep.Error = err.Error()
			return rep
		}
		if res.Skipped {
			rep.State = stateSkipped
			return rep
		}
		// Keep draining only while passes come back full. A partial pass —
		// including rows stuck pending after a failed dispatch — stops the
		// loop instead of spinning on unprocessable work. A full batch of
		// failing deliveries can still report More on the next pass (the
		// first retry delay is 0, so they are re-listed immediately);
		// webhookBudget bounds how long that may continue.
		if !res.More {
			return rep
		}
		if time.Now().After(deadline) {
			return rep
		}
	}
}

// Cleanup runs the hourly cleanup once under the cleanup advisory lock:
// expired images, webhook delivery/outbox retention, and audit-log retention.
func (r *Runner) Cleanup(ctx context.Context) TaskReport {
	start := time.Now()
	rep := TaskReport{State: stateRan, Counts: map[string]int64{}}
	defer func() { rep.DurationMS = time.Since(start).Milliseconds() }()

	acquired, err := pglock.TryWithLock(ctx, r.pool, pglock.KeyJobsCleanup, func(ctx context.Context) error {
		return r.cleanupLocked(ctx, &rep)
	})
	if err != nil {
		rep.Error = err.Error()
	}
	if !acquired {
		rep.State = stateSkipped
	}
	return rep
}

func (r *Runner) cleanupLocked(ctx context.Context, rep *TaskReport) error {
	var retErr error

	deleted, err := r.deleter.CleanExpiredImages(ctx, int32(r.cfg.Server.ExpiredCleanupBatchSize))
	if err != nil {
		slog.Warn("failed to clean expired images", "error", err)
		retErr = err
	} else if deleted > 0 {
		slog.Info("cleaned expired images", "deleted", deleted)
	}
	rep.Counts["expired_images_deleted"] = int64(deleted)

	if r.cfg.Webhook.Enabled {
		if n, err := r.queries.DeleteOldWebhookDeliveries(ctx, webhookDeliveryRetentionDays); err != nil {
			slog.Warn("failed to clean old webhook deliveries", "error", err)
			retErr = err
		} else {
			rep.Counts["webhook_deliveries_deleted"] = n
			if n > 0 {
				slog.Info("cleaned old webhook deliveries", "deleted", n)
			}
		}
		if n, err := r.queries.DeleteOldOutboxEvents(ctx, webhookDeliveryRetentionDays); err != nil {
			slog.Warn("failed to clean old outbox events", "error", err)
			retErr = err
		} else {
			rep.Counts["outbox_events_deleted"] = n
			if n > 0 {
				slog.Info("cleaned old outbox events", "deleted", n)
			}
		}
	}

	if retentionDays := int32(r.cfg.AppSnapshot().AuditLogRetentionDays); retentionDays > 0 {
		if n, err := r.queries.DeleteOldAuditLogs(ctx, retentionDays); err != nil {
			slog.Warn("failed to clean old audit logs", "error", err)
			retErr = err
		} else {
			rep.Counts["audit_logs_deleted"] = n
			if n > 0 {
				slog.Info("cleaned old audit logs", "deleted", n)
			}
		}
	}

	return retErr
}

// Start launches embedded scheduling: the webhook worker ticks every
// worker_interval (only when webhooks are enabled) and cleanup runs hourly.
// Both go through advisory locks, so co-running schedulers are safe.
func (r *Runner) Start() {
	if !r.started.CompareAndSwap(false, true) {
		return
	}
	if r.cfg.Webhook.Enabled {
		r.worker.Start()
	}
	r.wg.Add(1)
	go r.cleanupLoop()
}

// Stop halts embedded scheduling and waits for in-flight runs to finish.
func (r *Runner) Stop() {
	if !r.started.CompareAndSwap(true, false) {
		return
	}
	if r.cfg.Webhook.Enabled {
		r.worker.Stop()
	}
	close(r.cleanupStop)
	r.wg.Wait()
}

func (r *Runner) cleanupLoop() {
	defer r.wg.Done()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-r.cleanupStop:
			return
		case <-ticker.C:
			r.Cleanup(context.Background())
		}
	}
}
