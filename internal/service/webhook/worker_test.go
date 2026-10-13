package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
)

// newTestWorker seeds a user, a webhook subscription pointing at url, and a
// pending outbox event that matches it, then returns a worker on the pool.
func newTestWorker(t *testing.T, q *sqlc.Queries, pool *pgxpool.Pool, url string) *Worker {
	t.Helper()

	group := testutil.SeedDefaultGroup(t, q)
	user := testutil.SeedUser(t, q, group.ID, "hook-owner@example.com", "password123", "user")

	encKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(i)
	}
	svc := NewService(q, 10, encKey)
	if _, err := svc.Create(context.Background(), CreateParams{
		UserID: user.ID,
		Name:   "test-hook",
		URL:    url,
		Events: []string{"image.uploaded"},
	}); err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	payload, err := json.Marshal(events.Envelope{
		ID:             "evt-1",
		Type:           "image.uploaded",
		Version:        "1",
		OccurredAt:     time.Now(),
		IdempotencyKey: "key-1",
		Data:           map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if _, err := q.InsertOutboxEvent(context.Background(), sqlc.InsertOutboxEventParams{
		Type:           "image.uploaded",
		Version:        "1",
		IdempotencyKey: "key-1",
		Payload:        payload,
		OwnerUserID:    pgtype.Int8{Int64: user.ID, Valid: true},
	}); err != nil {
		t.Fatalf("insert outbox event: %v", err)
	}

	delivery := NewDeliveryService(q, 10*time.Second, true, true, encKey)
	return NewWorker(q, pool, delivery, time.Minute, 5, false, 1)
}

// Two concurrent RunOnce calls must not double-deliver: the second observes
// the advisory lock held and reports Skipped.
func TestRunOnceConcurrentCallsDispatchOnce(t *testing.T) {
	pool, q := testutil.SetupDB(t)

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		time.Sleep(400 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	worker := newTestWorker(t, q, pool, srv.URL)

	firstDone := make(chan Result, 1)
	go func() {
		res, err := worker.RunOnce(context.Background())
		if err != nil {
			t.Errorf("first RunOnce: %v", err)
		}
		firstDone <- res
	}()

	// Let the first call take the advisory lock and reach the slow dispatch.
	time.Sleep(150 * time.Millisecond)
	res2, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("second RunOnce: %v", err)
	}
	if !res2.Skipped {
		t.Fatalf("expected second RunOnce to be skipped, got %+v", res2)
	}

	res1 := <-firstDone
	if res1.Skipped {
		t.Fatal("first RunOnce unexpectedly skipped")
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected exactly 1 delivery request, got %d", got)
	}
}

// RunOnce must return only after in-flight deliveries finished, so the
// delivery row reflects its final status when the call completes.
func TestRunOnceWaitsForDeliveryCompletion(t *testing.T) {
	pool, q := testutil.SetupDB(t)

	received := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(received) })
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	worker := newTestWorker(t, q, pool, srv.URL)

	done := make(chan Result, 1)
	go func() {
		res, err := worker.RunOnce(context.Background())
		if err != nil {
			t.Errorf("RunOnce: %v", err)
		}
		done <- res
	}()

	select {
	case <-received:
	case <-time.After(10 * time.Second):
		t.Fatal("delivery request never arrived")
	}

	select {
	case <-done:
		t.Fatal("RunOnce returned while the delivery was still in flight")
	case <-time.After(150 * time.Millisecond):
	}

	close(release)
	res := <-done
	if res.Skipped {
		t.Fatal("RunOnce unexpectedly skipped")
	}
	if res.OutboxProcessed != 1 || res.DeliveriesAttempted != 1 {
		t.Fatalf("unexpected result counts: %+v", res)
	}

	var status string
	if err := pool.QueryRow(context.Background(),
		"SELECT status FROM webhook_deliveries LIMIT 1").Scan(&status); err != nil {
		t.Fatalf("query delivery status: %v", err)
	}
	if status != "delivered" {
		t.Fatalf("expected delivery status %q, got %q", "delivered", status)
	}
}

// A pass that fills the delivery batch must report More so the drain loop
// keeps going; the same stuck deliveries would otherwise hide the backlog.
func TestRunOnceReportsMoreWhenBatchFull(t *testing.T) {
	pool, q := testutil.SetupDB(t)
	ctx := context.Background()

	group := testutil.SeedDefaultGroup(t, q)
	user := testutil.SeedUser(t, q, group.ID, "full@example.com", "password123", "user")

	encKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(i)
	}
	wh, err := NewService(q, 10, encKey).Create(ctx, CreateParams{
		UserID: user.ID,
		Name:   "full-batch",
		URL:    "https://example.com/hook",
		Events: []string{"image.uploaded"},
	})
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	const concurrency = 20
	// Payloads unmarshal-fail so deliveries stay pending; owner is NULL so
	// processOutbox creates no additional deliveries.
	for i := 0; i < concurrency; i++ {
		evt, err := q.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
			Type:           "image.uploaded",
			Version:        "1",
			IdempotencyKey: fmt.Sprintf("full-%d", i),
			Payload:        json.RawMessage(`{"occurred_at":"bogus"}`),
			OwnerUserID:    pgtype.Int8{},
		})
		if err != nil {
			t.Fatalf("insert event %d: %v", i, err)
		}
		if _, err := q.CreateWebhookDelivery(ctx, sqlc.CreateWebhookDeliveryParams{
			WebhookID:      wh.ID,
			OutboxEventID:  evt.ID,
			RequestHeaders: json.RawMessage(`{}`),
		}); err != nil {
			t.Fatalf("create delivery %d: %v", i, err)
		}
	}

	worker := NewWorker(q, pool, NewDeliveryService(q, 10*time.Second, true, true, encKey), time.Minute, concurrency, false, 1)
	res, err := worker.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !res.More {
		t.Fatalf("More = false with a full batch: %+v", res)
	}
	if res.DeliveriesAttempted != concurrency {
		t.Fatalf("DeliveriesAttempted = %d, want %d", res.DeliveriesAttempted, concurrency)
	}
}
