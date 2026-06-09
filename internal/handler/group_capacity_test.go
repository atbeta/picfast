package handler

import (
	"encoding/json"
	"testing"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
)

func TestResolveUserCapacity_GroupSet(t *testing.T) {
	gc := domain.GroupConfig{UserCapacityBytes: 1073741824}
	configs, _ := json.Marshal(gc)
	group := sqlc.Group{ID: 1, Configs: configs}
	app := config.AppConfig{UserInitialCapacity: 524288000}

	got := resolveUserCapacity(group, app)
	if got != 1073741824 {
		t.Errorf("expected group value 1073741824, got %d", got)
	}
}

func TestResolveUserCapacity_GroupNotSet(t *testing.T) {
	gc := domain.GroupConfig{}
	configs, _ := json.Marshal(gc)
	group := sqlc.Group{ID: 1, Configs: configs}
	app := config.AppConfig{UserInitialCapacity: 524288000}

	got := resolveUserCapacity(group, app)
	if got != 524288000 {
		t.Errorf("expected fallback 524288000, got %d", got)
	}
}

func TestResolveUserCapacity_Unlimited(t *testing.T) {
	gc := domain.GroupConfig{UserCapacityBytes: -1}
	configs, _ := json.Marshal(gc)
	group := sqlc.Group{ID: 1, Configs: configs}
	app := config.AppConfig{UserInitialCapacity: 524288000}

	got := resolveUserCapacity(group, app)
	if got != 0 {
		t.Errorf("expected unlimited (0), got %d", got)
	}
}

func TestResolveUserCapacity_InvalidJSON(t *testing.T) {
	group := sqlc.Group{ID: 1, Configs: []byte("not-json")}
	app := config.AppConfig{UserInitialCapacity: 524288000}

	got := resolveUserCapacity(group, app)
	if got != 524288000 {
		t.Errorf("expected fallback on invalid json 524288000, got %d", got)
	}
}
