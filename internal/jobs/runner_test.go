package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/pglock"
	"github.com/atbeta/picfast/internal/service/webhook"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
)

func testRunnerConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.Storage.ThumbnailDir = t.TempDir()
	cfg.Server.ExpiredCleanupBatchSize = 100
	cfg.Webhook.Enabled = false
	return cfg
}

func TestCleanupDeletesExpiredImage(t *testing.T) {
	pool, q := testutil.SetupDB(t)
	ctx := context.Background()

	group := testutil.SeedDefaultGroup(t, q)
	strategy := testutil.SeedStrategy(t, q, group.ID)
	user := testutil.SeedUser(t, q, group.ID, "u@example.com", "password123", "user")

	img, err := q.CreateImage(ctx, sqlc.CreateImageParams{
		UserID:     domain.PgInt8(user.ID),
		GroupID:    domain.PgInt8(group.ID),
		StrategyID: domain.PgInt8(strategy.ID),
		Key:        "expired1",
		Path:       ".",
		Name:       "expired1.png",
		OriginName: "expired1.png",
		SizeBytes:  10,
		Mimetype:   "image/png",
		Extension:  "png",
		Md5:        "cccccccccccccccccccccccccccccccc",
		Sha1:       "dddddddddddddddddddddddddddddddddddddddd",
		Width:      10,
		Height:     10,
		Permission: int16(domain.PermissionPrivate),
		UploadedIp: "127.0.0.1",
		ExpiresAt:  pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("create expired image: %v", err)
	}

	runner := NewRunner(q, pool, testRunnerConfig(t))
	rep := runner.Cleanup(ctx)

	if rep.State != stateRan {
		t.Fatalf("cleanup state = %q, want %q (err=%q)", rep.State, stateRan, rep.Error)
	}
	if rep.Error != "" {
		t.Fatalf("cleanup error = %q", rep.Error)
	}
	if rep.Counts["expired_images_deleted"] != 1 {
		t.Fatalf("expired_images_deleted = %v, want 1 (counts=%v)", rep.Counts["expired_images_deleted"], rep.Counts)
	}
	if _, err := q.GetImageByID(ctx, img.ID); err != pgx.ErrNoRows {
		t.Fatalf("expired image still present (err=%v)", err)
	}
}

func TestCleanupSkippedWhenLockHeld(t *testing.T) {
	pool, q := testutil.SetupDB(t)
	ctx := context.Background()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", pglock.KeyJobsCleanup).Scan(&locked); err != nil {
		t.Fatalf("pg_try_advisory_lock: %v", err)
	}
	if !locked {
		t.Fatal("could not acquire cleanup advisory lock")
	}
	defer func() {
		if _, err := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", pglock.KeyJobsCleanup); err != nil {
			t.Errorf("unlock: %v", err)
		}
	}()

	runner := NewRunner(q, pool, testRunnerConfig(t))
	rep := runner.RunOnce(ctx)

	if rep.Cleanup.State != stateSkipped {
		t.Fatalf("cleanup state = %q, want %q", rep.Cleanup.State, stateSkipped)
	}
	if rep.Webhooks.State != stateDisabled {
		t.Fatalf("webhooks state = %q, want %q", rep.Webhooks.State, stateDisabled)
	}
}

// A delivery that fails before its status is updated stays pending and is
// re-listed by every pass — the drain loop must exit after the partial pass
// instead of spinning for the whole budget.
func TestRunOnceDoesNotSpinOnStuckDelivery(t *testing.T) {
	pool, q := testutil.SetupDB(t)
	ctx := context.Background()

	group := testutil.SeedDefaultGroup(t, q)
	user := testutil.SeedUser(t, q, group.ID, "hook@example.com", "password123", "user")

	encKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(i)
	}
	wh, err := webhook.NewService(q, 10, encKey).Create(ctx, webhook.CreateParams{
		UserID: user.ID,
		Name:   "stuck",
		URL:    "https://example.com/hook",
		Events: []string{"image.uploaded"},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	// Valid JSONB but not a valid events.Envelope (occurred_at is not a
	// timestamp): dispatch fails at unmarshal before any status update, so
	// the delivery remains pending forever.
	evt, err := q.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		Type:           "image.uploaded",
		Version:        "1",
		IdempotencyKey: "stuck-1",
		Payload:        json.RawMessage(`{"occurred_at":"bogus"}`),
		// NULL owner matches no webhook, so processOutbox marks the event
		// dispatched instead of creating a second delivery.
		OwnerUserID: pgtype.Int8{},
	})
	if err != nil {
		t.Fatalf("insert outbox event: %v", err)
	}
	d, err := q.CreateWebhookDelivery(ctx, sqlc.CreateWebhookDeliveryParams{
		WebhookID:      wh.ID,
		OutboxEventID:  evt.ID,
		RequestHeaders: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("create delivery: %v", err)
	}

	cfg := testRunnerConfig(t)
	cfg.Webhook.Enabled = true
	runner := NewRunner(q, pool, cfg)

	start := time.Now()
	report := runner.RunOnce(ctx)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("RunOnce took %v, want < 2s", elapsed)
	}
	if report.Webhooks.State != stateRan || report.Webhooks.Error != "" {
		t.Fatalf("webhooks report = %+v", report.Webhooks)
	}
	if got := report.Webhooks.Counts["deliveries_attempted"]; got != 1 {
		t.Fatalf("deliveries_attempted = %d, want 1", got)
	}

	var status string
	if err := pool.QueryRow(ctx, "SELECT status FROM webhook_deliveries WHERE id = $1", d.ID).Scan(&status); err != nil {
		t.Fatalf("query delivery status: %v", err)
	}
	if status != "pending" {
		t.Fatalf("stuck delivery status = %q, want pending", status)
	}
}
