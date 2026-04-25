package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/sqlc"
	"golang.org/x/crypto/bcrypt"
)

type AdminUserHandler struct {
	db *sqlc.Queries
}

func NewAdminUserHandler(db *sqlc.Queries) *AdminUserHandler {
	return &AdminUserHandler{db: db}
}

func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	statusStr := r.URL.Query().Get("status")
	page, pageSize := parsePagination(r)

	users, err := h.db.ListUsers(r.Context(), sqlc.ListUsersParams{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	// Apply filters in Go for simplicity (private deployment scale)
	if keyword != "" || statusStr != "" {
		filtered := make([]sqlc.User, 0)
		for _, u := range users {
			if keyword != "" && !strings.Contains(strings.ToLower(u.Email), strings.ToLower(keyword)) &&
				!strings.Contains(strings.ToLower(u.Name), strings.ToLower(keyword)) {
				continue
			}
			if statusStr != "" {
				s, err := strconv.ParseInt(statusStr, 10, 16)
				if err == nil && u.Status != int16(s) {
					continue
				}
			}
			filtered = append(filtered, u)
		}
		users = filtered
	}

	total, _ := h.db.CountUsers(r.Context())

	items := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		items = append(items, userJSON(u))
	}

	Paginated(w, items, total, page, pageSize)
}

func (h *AdminUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "user not found")
		return
	}

	Success(w, userJSON(user))
}

type adminUpdateUserRequest struct {
	Name          *string          `json:"name"`
	Password      *string          `json:"password"`
	GroupID       *int64           `json:"group_id"`
	CapacityBytes *int64           `json:"capacity_bytes"`
	Status        *int16           `json:"status"`
	Settings      *json.RawMessage `json:"settings"`
}

func (h *AdminUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req adminUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "user not found")
		return
	}

	name := user.Name
	password := user.Password
	groupID := user.GroupID
	capacityBytes := user.CapacityBytes
	status := user.Status
	settings := user.Settings

	if req.Name != nil {
		name = *req.Name
	}
	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		password = string(hash)
	}
	if req.GroupID != nil {
		groupID = domain.PgInt8(*req.GroupID)
	}
	if req.CapacityBytes != nil {
		capacityBytes = *req.CapacityBytes
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.Settings != nil {
		settings = *req.Settings
	}

	updated, err := h.db.UpdateUser(r.Context(), sqlc.UpdateUserParams{
		ID:            id,
		Name:          name,
		Password:      password,
		GroupID:       groupID,
		CapacityBytes: capacityBytes,
		Status:        status,
		Settings:      settings,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	Success(w, userJSON(updated))
}

func (h *AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role == string(domain.RoleAdmin) {
		Fail(w, http.StatusForbidden, "cannot delete admin user")
		return
	}

	if err := h.db.DeleteUser(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	SuccessMessage(w, "deleted")
}

func userJSON(u sqlc.User) map[string]interface{} {
	return map[string]interface{}{
		"id":              u.ID,
		"email":           u.Email,
		"name":            u.Name,
		"role":            u.Role,
		"group_id":        domain.PgInt8PtrVal(u.GroupID),
		"capacity_bytes":  u.CapacityBytes,
		"image_num":       u.ImageNum,
		"album_num":       u.AlbumNum,
		"status":          u.Status,
		"email_verified":  u.EmailVerified,
		"settings":        json.RawMessage(u.Settings),
		"created_at":      u.CreatedAt,
	}
}

func parsePagination(r *http.Request) (page, pageSize int32) {
	page = 1
	pageSize = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.ParseInt(p, 10, 32); err == nil && v > 0 {
			page = int32(v)
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.ParseInt(ps, 10, 32); err == nil && v > 0 && v <= 100 {
			pageSize = int32(v)
		}
	}
	return
}
