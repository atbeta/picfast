package handler

import (
	"encoding/json"
	"log/slog"
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

func realIPFromRequest(r *http.Request) string {
	if !isTrustedProxy(originalAddr(r)) {
		return remoteHost(r.RemoteAddr)
	}
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		if ip := net.ParseIP(cfIP); ip != nil {
			return ip.String()
		}
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		if ip := rightmostUntrustedIP(forwarded); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return ip.String()
		}
	}
	return remoteHost(r.RemoteAddr)
}

func rightmostUntrustedIP(forwarded string) string {
	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		ip := net.ParseIP(strings.TrimSpace(parts[i]))
		if ip == nil {
			continue
		}
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		if !ipInTrustedCIDRs(ip) {
			return ip.String()
		}
	}
	return ""
}

func ipInTrustedCIDRs(ip net.IP) bool {
	trustedProxyMu.RLock()
	defer trustedProxyMu.RUnlock()
	for _, cidr := range trustedProxyCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err == nil && host != "" {
		return host
	}
	return addr
}
