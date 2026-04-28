package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminGroupHandler struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

func NewAdminGroupHandler(db *sqlc.Queries, pool *pgxpool.Pool) *AdminGroupHandler {
	return &AdminGroupHandler{db: db, pool: pool}
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
		ID          int64           `json:"id"`
		Name        string          `json:"name"`
		IsDefault   bool            `json:"is_default"`
		IsGuest     bool            `json:"is_guest"`
		Configs     json.RawMessage `json:"configs"`
		StrategyIDs []int64         `json:"strategy_ids"`
		CreatedAt   string          `json:"created_at"`
		UpdatedAt   string          `json:"updated_at"`
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
			IsGuest: g.IsGuest, Configs: g.Configs, StrategyIDs: stratIDs,
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
		"id":           group.ID,
		"name":         group.Name,
		"is_default":   group.IsDefault,
		"is_guest":     group.IsGuest,
		"configs":      json.RawMessage(group.Configs),
		"strategy_ids": stratIDs,
		"created_at":   group.CreatedAt,
		"updated_at":   group.UpdatedAt,
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

	var group sqlc.Group
	if err := sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		var err error
		group, err = qtx.CreateGroup(r.Context(), sqlc.CreateGroupParams{
			Name:      req.Name,
			IsDefault: req.IsDefault,
			IsGuest:   req.IsGuest,
			Configs:   req.Configs,
		})
		if err != nil {
			return err
		}
		if group.IsDefault {
			qtx.UnsetOtherDefault(r.Context(), group.ID)
		}
		if group.IsGuest {
			qtx.UnsetOtherGuest(r.Context(), group.ID)
		}
		return nil
	}); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create group")
		return
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

	var group sqlc.Group
	if err := sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		var err error
		group, err = qtx.UpdateGroup(r.Context(), sqlc.UpdateGroupParams{
			ID:        id,
			Name:      req.Name,
			IsDefault: req.IsDefault,
			IsGuest:   req.IsGuest,
			Configs:   req.Configs,
		})
		if err != nil {
			return err
		}
		if group.IsDefault {
			qtx.UnsetOtherDefault(r.Context(), group.ID)
		}
		if group.IsGuest {
			qtx.UnsetOtherGuest(r.Context(), group.ID)
		}
		return nil
	}); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update group")
		return
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

	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		qtx.ReplaceGroupStrategies(r.Context(), id)
		for _, sid := range req.StrategyIDs {
			if err := qtx.AddGroupStrategy(r.Context(), sqlc.AddGroupStrategyParams{
				GroupID: id, StrategyID: sid,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to set strategies")
		return
	}

	SuccessMessage(w, "strategies updated")
}

func groupJSON(g sqlc.Group) AdminGroupResponse {
	return AdminGroupResponse{
		ID:        g.ID,
		Name:      g.Name,
		IsDefault: g.IsDefault,
		IsGuest:   g.IsGuest,
		Configs:   json.RawMessage(g.Configs),
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}
