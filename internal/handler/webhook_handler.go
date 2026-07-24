package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/webhook"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
)

type WebhookHandler struct {
	db       *sqlc.Queries
	service  *webhook.Service
	delivery *webhook.DeliveryService
}

func NewWebhookHandler(db *sqlc.Queries, service *webhook.Service, delivery *webhook.DeliveryService) *WebhookHandler {
	return &WebhookHandler{db: db, service: service, delivery: delivery}
}

func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	webhooks, err := h.service.List(r.Context(), userID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list webhooks")
		return
	}

	items := make([]webhookItem, 0, len(webhooks))
	for _, wh := range webhooks {
		items = append(items, makeWebhookItem(wh))
	}
	Success(w, items)
}

func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name   string   `json:"name"`
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.URL == "" {
		Fail(w, http.StatusBadRequest, "url is required")
		return
	}
	if len(req.Events) == 0 {
		Fail(w, http.StatusBadRequest, "at least one event type is required")
		return
	}

	result, err := h.service.Create(r.Context(), webhook.CreateParams{
		UserID: userID,
		Name:   req.Name,
		URL:    req.URL,
		Events: req.Events,
	})
	if err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	Created(w, webhookCreateItem{
		ID:        result.ID,
		Name:      result.Name,
		URL:       result.Url,
		Events:    parseEvents(result.Events),
		Enabled:   result.Enabled,
		Secret:    result.Secret,
		CreatedAt: result.CreatedAt.UTC().Format(time.RFC3339),
	})

	writeAuditLog(h.db, r, "webhook.create", "webhook", strconv.FormatInt(result.ID, 10), result.Name, map[string]any{
		"url": result.Url,
	})
}

func (h *WebhookHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	wh, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		Fail(w, http.StatusNotFound, "webhook not found")
		return
	}

	Success(w, makeWebhookItem(wh))
}

func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	var params webhook.UpdateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wh, err := h.service.Update(r.Context(), id, userID, params)
	if err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	Success(w, makeWebhookItem(wh))

	writeAuditLog(h.db, r, "webhook.update", "webhook", strconv.FormatInt(wh.ID, 10), wh.Name, nil)
}

func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuditLog(h.db, r, "webhook.delete", "webhook", strconv.FormatInt(id, 10), strconv.FormatInt(id, 10), nil)
	SuccessMessage(w, "webhook deleted")
}

func (h *WebhookHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	secret, err := h.service.RotateSecret(r.Context(), id, userID)
	if err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	Success(w, map[string]string{"secret": secret})
}

func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	if _, err := h.service.GetByID(r.Context(), id, userID); err != nil {
		Fail(w, http.StatusNotFound, "webhook not found")
		return
	}

	page, pageSize := parsePagination(r)

	deliveries, err := h.delivery.ListByWebhook(r.Context(), id, int32(pageSize), int32((page-1)*pageSize))
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list deliveries")
		return
	}

	total, _ := h.delivery.CountByWebhook(r.Context(), id)

	Paginated(w, deliveries, total, page, pageSize)
}

func (h *WebhookHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	delivery, err := h.delivery.GetByIDAndUser(r.Context(), id, userID)
	if err != nil {
		Fail(w, http.StatusNotFound, "delivery not found")
		return
	}

	Success(w, delivery)
}

func (h *WebhookHandler) ReplayDelivery(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	if _, err := h.delivery.GetByIDAndUser(r.Context(), id, userID); err != nil {
		Fail(w, http.StatusNotFound, "delivery not found")
		return
	}

	if err := h.delivery.Redeliver(r.Context(), id); err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	SuccessMessage(w, "replay queued")
}

func (h *WebhookHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	types := []map[string]string{
		{"type": "image.uploaded", "version": "2026-06-01", "description": "An image was uploaded"},
		{"type": "image.processed", "version": "2026-06-01", "description": "Image processing pipeline completed"},
		{"type": "image.deleted", "version": "2026-06-01", "description": "An image was deleted"},
		{"type": "moderation.reviewed", "version": "2026-06-01", "description": "Moderation status changed"},
	}
	Success(w, types)
}

func (h *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	if _, err := h.service.GetByID(r.Context(), id, userID); err != nil {
		Fail(w, http.StatusNotFound, "webhook not found")
		return
	}

	if err := h.delivery.TestDispatch(r.Context(), id); err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	SuccessMessage(w, "test event sent successfully")
}

type webhookItem struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Enabled   bool     `json:"enabled"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type webhookCreateItem struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Enabled   bool     `json:"enabled"`
	Secret    string   `json:"secret"`
	CreatedAt string   `json:"created_at"`
}

func makeWebhookItem(wh sqlc.Webhook) webhookItem {
	return webhookItem{
		ID:        wh.ID,
		Name:      wh.Name,
		URL:       wh.Url,
		Events:    parseEvents(wh.Events),
		Enabled:   wh.Enabled,
		CreatedAt: wh.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: wh.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func parseEvents(raw json.RawMessage) []string {
	var events []string
	if err := json.Unmarshal(raw, &events); err != nil {
		return []string{}
	}
	return events
}
