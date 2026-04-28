package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
)

type AdminStrategyHandler struct {
	db *sqlc.Queries
}

func NewAdminStrategyHandler(db *sqlc.Queries) *AdminStrategyHandler {
	return &AdminStrategyHandler{db: db}
}

type createStrategyRequest struct {
	Name         string          `json:"name"`
	StrategyType string          `json:"strategy_type"`
	Configs      json.RawMessage `json:"configs"`
}

func (h *AdminStrategyHandler) List(w http.ResponseWriter, r *http.Request) {
	strategies, err := h.db.ListStrategies(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list strategies")
		return
	}
	Success(w, strategies)
}

func (h *AdminStrategyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	s, err := h.db.GetStrategyByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "strategy not found")
		return
	}

	Success(w, strategyJSON(s))
}

func (h *AdminStrategyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	if !storage.IsKnownType(req.StrategyType) {
		Fail(w, http.StatusBadRequest, "unknown strategy_type")
		return
	}

	if len(req.Configs) == 0 {
		Fail(w, http.StatusBadRequest, "configs is required")
		return
	}

	if err := storage.ValidateConfig(req.StrategyType, req.Configs); err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	s, err := h.db.CreateStrategy(r.Context(), sqlc.CreateStrategyParams{
		Name:         req.Name,
		StrategyType: req.StrategyType,
		Configs:      req.Configs,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create strategy")
		return
	}

	Created(w, strategyJSON(s))
}

func (h *AdminStrategyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req createStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := h.db.GetStrategyByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "strategy not found")
		return
	}

	name := existing.Name
	strategyType := existing.StrategyType
	configs := existing.Configs

	if req.Name != "" {
		name = req.Name
	}
	if req.StrategyType != "" {
		if !storage.IsKnownType(req.StrategyType) {
			Fail(w, http.StatusBadRequest, "unknown strategy_type")
			return
		}
		strategyType = req.StrategyType
	}
	if len(req.Configs) > 0 {
		if err := storage.ValidateConfig(strategyType, req.Configs); err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}
		configs = req.Configs
	}

	s, err := h.db.UpdateStrategy(r.Context(), sqlc.UpdateStrategyParams{
		ID:           id,
		Name:         name,
		StrategyType: strategyType,
		Configs:      configs,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update strategy")
		return
	}

	Success(w, strategyJSON(s))
}

func (h *AdminStrategyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.db.DeleteStrategy(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete strategy")
		return
	}

	SuccessMessage(w, "deleted")
}

func strategyJSON(s sqlc.Strategy) AdminStrategyResponse {
	return AdminStrategyResponse{
		ID:           s.ID,
		Name:         s.Name,
		StrategyType: s.StrategyType,
		Configs:      json.RawMessage(s.Configs),
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}
