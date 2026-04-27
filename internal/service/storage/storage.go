package storage

import (
	"context"
	"encoding/json"
	"fmt"
)

type HealthResult struct {
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}

type Storage interface {
	Write(ctx context.Context, path string, data []byte) error
	Read(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	URL(pathname string) string
	HealthCheck(ctx context.Context) HealthResult
	Close() error
}

// Constructor creates a Storage from its JSON configuration.
type Constructor func(json.RawMessage) (Storage, error)

var registry = map[string]Constructor{}

// Register registers a storage backend constructor. Call it from init().
func Register(typ string, ctor Constructor) {
	registry[typ] = ctor
}

// New creates a Storage instance by type and JSON config.
func New(typ string, cfg json.RawMessage) (Storage, error) {
	ctor, ok := registry[typ]
	if !ok {
		return nil, fmt.Errorf("unknown storage type: %s", typ)
	}
	return ctor(cfg)
}
