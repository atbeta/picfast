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
	detailsJSON := []byte("{}")
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			detailsJSON = b
		}
	}

	var actorUserID int64
	hasActor := false
	if uid, ok := r.Context().Value(domain.ContextKeyUserID).(int64); ok {
		actorUserID = uid
		hasActor = true
	}

	ip := clientip.FromRequest(r)

	if err := db.CreateAuditLog(r.Context(), sqlc.CreateAuditLogParams{
		ActorUserID:  pgtype.Int8{Int64: actorUserID, Valid: hasActor},
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
