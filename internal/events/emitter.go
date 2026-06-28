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

func EmitAsync(emitter Emitter, events ...Envelope) {
	go func() {
		defer func() { recover() }()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, event := range events {
			if err := emitter.Emit(ctx, event); err != nil {
				slog.Warn("event emit failed", "type", event.Type, "error", err)
			}
		}
	}()
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
