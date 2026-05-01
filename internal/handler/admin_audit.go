package handler

import (
	"encoding/json"
	"net/http"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type AdminAuditHandler struct {
	db *sqlc.Queries
}

func NewAdminAuditHandler(db *sqlc.Queries) *AdminAuditHandler {
	return &AdminAuditHandler{db: db}
}

func (h *AdminAuditHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	action := r.URL.Query().Get("action")
	resourceType := r.URL.Query().Get("resource_type")

	rows, err := h.db.ListAuditLogs(r.Context(), sqlc.ListAuditLogsParams{
		Limit:        pageSize,
		Offset:       (page - 1) * pageSize,
		Action:       pgTextParam(action),
		ResourceType: pgTextParam(resourceType),
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list audit logs")
		return
	}
	total, err := h.db.CountAuditLogs(r.Context(), sqlc.CountAuditLogsParams{
		Action:       pgTextParam(action),
		ResourceType: pgTextParam(resourceType),
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to count audit logs")
		return
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		details := json.RawMessage(row.Details)
		var actorUserID any = nil
		if row.ActorUserID.Valid {
			actorUserID = row.ActorUserID.Int64
		}
		var actorEmail any = nil
		if row.ActorEmail.Valid {
			actorEmail = row.ActorEmail.String
		}
		items = append(items, map[string]any{
			"id":            row.ID,
			"actor_user_id": actorUserID,
			"actor_email":   actorEmail,
			"action":        row.Action,
			"resource_type": row.ResourceType,
			"resource_id":   row.ResourceID.String,
			"resource_name": row.ResourceName.String,
			"details":       details,
			"ip":            row.Ip,
			"user_agent":    row.UserAgent,
			"created_at":    row.CreatedAt,
		})
	}

	Paginated(w, items, total, page, pageSize)
}

func pgTextParam(v string) pgtype.Text {
	if v == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: v, Valid: true}
}
