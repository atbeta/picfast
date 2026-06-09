package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
)

func resolveUserCapacity(group sqlc.Group, app config.AppConfig) int64 {
	var gc domain.GroupConfig
	if err := json.Unmarshal(group.Configs, &gc); err != nil {
		slog.Warn("invalid group configs, falling back to default capacity", "group_id", group.ID, "error", err)
		return app.UserInitialCapacity
	}
	if gc.UserCapacityBytes < 0 {
		return 0
	}
	if gc.UserCapacityBytes > 0 {
		return gc.UserCapacityBytes
	}
	return app.UserInitialCapacity
}
