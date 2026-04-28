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

// ConfigValidator validates the JSON configuration for a storage type.
type ConfigValidator func(json.RawMessage) error

var validatorRegistry = map[string]ConfigValidator{}

// RegisterValidator registers a config validator for a storage backend type.
func RegisterValidator(typ string, v ConfigValidator) {
	validatorRegistry[typ] = v
}

// ValidateConfig validates config JSON for the given storage type.
func ValidateConfig(typ string, cfg json.RawMessage) error {
	v, ok := validatorRegistry[typ]
	if !ok {
		return fmt.Errorf("unknown storage type: %s", typ)
	}
	return v(cfg)
}

// IsKnownType returns true if the type is a registered storage backend.
func IsKnownType(typ string) bool {
	_, ok := registry[typ]
	return ok
}
