package events

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type Emitter interface {
	Emit(ctx context.Context, event Envelope) error
}

// EmitBestEffort emits events synchronously on a detached context derived from
// ctx (canceled requests still emit, bounded by a 5s timeout). Failures and
// panics are logged and never propagated.
func EmitBestEffort(ctx context.Context, emitter Emitter, events ...Envelope) {
	emitCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	for _, event := range events {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Warn("event emit panicked", "type", event.Type, "panic", r)
				}
			}()
			if err := emitter.Emit(emitCtx, event); err != nil {
				slog.Warn("event emit failed", "type", event.Type, "error", err)
			}
		}()
	}
}

type LogEmitter struct{}

func NewLogEmitter() *LogEmitter {
	return &LogEmitter{}
}

func (e *LogEmitter) Emit(ctx context.Context, event Envelope) error {
	slog.Info("event emitted",
		"id", event.ID,
		"type", event.Type,
		"version", event.Version,
		"idempotency_key", event.IdempotencyKey,
	)
	return nil
}

type CollectEmitter struct {
	mu     sync.Mutex
	Events []Envelope
}

func NewCollectEmitter() *CollectEmitter {
	return &CollectEmitter{}
}

func (e *CollectEmitter) Emit(ctx context.Context, event Envelope) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Events = append(e.Events, event)
	return nil
}

func (e *CollectEmitter) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Events = nil
}

type FailEmitter struct{}

func (FailEmitter) Emit(context.Context, Envelope) error {
	return errors.New("emit failed")
}
