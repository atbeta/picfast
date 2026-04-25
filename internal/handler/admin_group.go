package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pbeta/imgapi/internal/sqlc"
)

type AdminGroupHandler struct {
	db *sqlc.Queries
}

func NewAdminGroupHandler(db *sqlc.Queries) *AdminGroupHandler {
	return &AdminGroupHandler{db: db}
}

type createGroupRequest struct {
	Name      string          `json:"name"`
	IsDefault bool            `json:"is_default"`
	IsGuest   bool            `json:"is_guest"`
	Configs   json.RawMessage `json:"configs"`
}

type updateGroupRequest struct {
	Name      string          `json:"name"`
	IsDefault bool            `json:"is_default"`
	IsGuest   bool            `json:"is_guest"`
	Configs   json.RawMessage `json:"configs"`
}

func (h *AdminGroupHandler) List(w http.ResponseWriter, r *http.Request) {
	groups, err := h.db.ListGroups(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list groups")
		return
	}

	type groupRow struct {
		ID        int64           `json:"id"`
		Name      string          `json:"name"`
		IsDefault bool            `json:"is_default"`
		IsGuest   bool            `json:"is_guest"`
		Configs   json.RawMessage `json:"configs"`
		CreatedAt string          `json:"created_at"`
		UpdatedAt string          `json:"updated_at"`
	}

	items := make([]groupRow, 0, len(groups))
	for _, g := range groups {
		strategies, _ := h.db.GetGroupStrategies(r.Context(), g.ID)
		stratIDs := make([]int64, 0, len(strategies))
		for _, s := range strategies {
			stratIDs = append(stratIDs, s.ID)
		}
		items = append(items, groupRow{
			ID: g.ID, Name: g.Name, IsDefault: g.IsDefault,
			IsGuest: g.IsGuest, Configs: g.Configs,
			CreatedAt: g.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: g.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	Success(w, items)
}

func (h *AdminGroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	group, err := h.db.GetGroupByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "group not found")
		return
	}

	strategies, _ := h.db.GetGroupStrategies(r.Context(), group.ID)
	stratIDs := make([]int64, 0, len(strategies))
	for _, s := range strategies {
		stratIDs = append(stratIDs, s.ID)
	}

	Success(w, map[string]interface{}{
		"id":             group.ID,
		"name":           group.Name,
		"is_default":     group.IsDefault,
		"is_guest":       group.IsGuest,
		"configs":        json.RawMessage(group.Configs),
		"strategy_ids":   stratIDs,
		"created_at":     group.CreatedAt,
		"updated_at":     group.UpdatedAt,
	})
}

func (h *AdminGroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	if len(req.Configs) == 0 {
		req.Configs = json.RawMessage(`{}`)
	}

	group, err := h.db.CreateGroup(r.Context(), sqlc.CreateGroupParams{
		Name:      req.Name,
		IsDefault: req.IsDefault,
		IsGuest:   req.IsGuest,
		Configs:   req.Configs,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create group")
		return
	}

	if group.IsDefault {
		h.db.UnsetOtherDefault(r.Context(), group.ID)
	}
	if group.IsGuest {
		h.db.UnsetOtherGuest(r.Context(), group.ID)
	}

	Created(w, groupJSON(group))
}

func (h *AdminGroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req updateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	if len(req.Configs) == 0 {
		existing, err := h.db.GetGroupByID(r.Context(), id)
		if err != nil {
			Fail(w, http.StatusNotFound, "group not found")
			return
		}
		req.Configs = existing.Configs
	}

	group, err := h.db.UpdateGroup(r.Context(), sqlc.UpdateGroupParams{
		ID:        id,
		Name:      req.Name,
		IsDefault: req.IsDefault,
		IsGuest:   req.IsGuest,
		Configs:   req.Configs,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update group")
		return
	}

	if group.IsDefault {
		h.db.UnsetOtherDefault(r.Context(), group.ID)
	}
	if group.IsGuest {
		h.db.UnsetOtherGuest(r.Context(), group.ID)
	}

	Success(w, groupJSON(group))
}

func (h *AdminGroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	group, err := h.db.GetGroupByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "group not found")
		return
	}

	if group.IsDefault || group.IsGuest {
		Fail(w, http.StatusBadRequest, "cannot delete default or guest group")
		return
	}

	if err := h.db.DeleteGroup(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete group")
		return
	}

	SuccessMessage(w, "deleted")
}

type setGroupStrategiesRequest struct {
	StrategyIDs []int64 `json:"strategy_ids"`
}

func (h *AdminGroupHandler) SetStrategies(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req setGroupStrategiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.db.ReplaceGroupStrategies(r.Context(), id)
	for _, sid := range req.StrategyIDs {
		h.db.AddGroupStrategy(r.Context(), sqlc.AddGroupStrategyParams{
			GroupID: id, StrategyID: sid,
		})
	}

	SuccessMessage(w, "strategies updated")
}

func groupJSON(g sqlc.Group) map[string]interface{} {
	return map[string]interface{}{
		"id":         g.ID,
		"name":       g.Name,
		"is_default": g.IsDefault,
		"is_guest":   g.IsGuest,
		"configs":    json.RawMessage(g.Configs),
		"created_at": g.CreatedAt,
		"updated_at": g.UpdatedAt,
	}
}
