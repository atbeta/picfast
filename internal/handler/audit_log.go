package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

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

	ip := realIPFromRequest(r)

	_ = db.CreateAuditLog(r.Context(), sqlc.CreateAuditLogParams{
		ActorUserID:  pgtype.Int8{Int64: actorUserID, Valid: hasActor},
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   pgText(resourceID),
		ResourceName: pgText(resourceName),
		Details:      detailsJSON,
		Ip:           ip,
		UserAgent:    r.UserAgent(),
	})
}

func pgText(v string) pgtype.Text {
	if strings.TrimSpace(v) == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: v, Valid: true}
}

func realIPFromRequest(r *http.Request) string {
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		if net.ParseIP(cfIP) != nil {
			return cfIP
		}
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			first := strings.TrimSpace(parts[0])
			if net.ParseIP(first) != nil {
				return first
			}
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
