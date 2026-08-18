package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func writeAuditLog(db *sqlc.Queries, r *http.Request, action, resourceType, resourceID, resourceName string, details map[string]any) {
	if db == nil || r == nil {
		return
	}
	var actor pgtype.Int8
	if uid, ok := r.Context().Value(domain.ContextKeyUserID).(int64); ok {
		actor = pgtype.Int8{Int64: uid, Valid: true}
	}
	writeAuditLogEntry(db, r, actor, action, resourceType, resourceID, resourceName, details)
}

// writeAuditLogWithActor 显式指定操作者，用于登录回调等未经过 auth middleware、
// context 中没有 userID 的场景。
func writeAuditLogWithActor(db *sqlc.Queries, r *http.Request, actorUserID int64, action, resourceType, resourceID, resourceName string, details map[string]any) {
	writeAuditLogEntry(db, r, pgtype.Int8{Int64: actorUserID, Valid: true}, action, resourceType, resourceID, resourceName, details)
}

func writeAuditLogEntry(db *sqlc.Queries, r *http.Request, actor pgtype.Int8, action, resourceType, resourceID, resourceName string, details map[string]any) {
	if db == nil || r == nil {
		return
	}
	detailsJSON := []byte("{}")
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			detailsJSON = b
		}
	}

	ip := clientip.FromRequest(r)

	if err := db.CreateAuditLog(r.Context(), sqlc.CreateAuditLogParams{
		ActorUserID:  actor,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   pgText(resourceID),
		ResourceName: pgText(resourceName),
		Details:      detailsJSON,
		Ip:           ip,
		UserAgent:    r.UserAgent(),
	}); err != nil {
		slog.Warn("audit log insert failed", "action", action, "resource_type", resourceType, "error", err)
	}
}

func pgText(v string) pgtype.Text {
	if strings.TrimSpace(v) == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: v, Valid: true}
}
