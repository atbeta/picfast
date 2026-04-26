package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pbeta/imgapi/internal/config"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/handler"
)

func testJWT() *handler.JWTService {
	return handler.NewJWTService(&config.JWTConfig{
		Secret:     "test-secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	})
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(domain.ContextKeyUserID).(int64)
	role, _ := r.Context().Value(domain.ContextKeyRole).(domain.UserRole)
	groupID, _ := r.Context().Value(domain.ContextKeyGroupID).(int64)
	handler.Success(w, map[string]interface{}{
		"user_id":  userID,
		"role":     string(role),
		"group_id": groupID,
	})
}

func TestAuth_ValidToken(t *testing.T) {
	jwtSvc := testJWT()
	token, _, _ := jwtSvc.GenerateAccessToken(42, domain.RoleUser, 7)

	r := chiRouter(Auth(jwtSvc), okHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if int64(data["user_id"].(float64)) != 42 {
		t.Fatalf("user_id = %v, want 42", data["user_id"])
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	jwtSvc := testJWT()
	r := chiRouter(Auth(jwtSvc), okHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	jwtSvc := testJWT()
	r := chiRouter(Auth(jwtSvc), okHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestOptionalAuth_NoToken(t *testing.T) {
	jwtSvc := testJWT()
	r := chiRouter(OptionalAuth(jwtSvc), okHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["user_id"] != nil && data["user_id"].(float64) != 0 {
		t.Fatal("expected no user_id in context")
	}
}

func TestOptionalAuth_ValidToken(t *testing.T) {
	jwtSvc := testJWT()
	token, _, _ := jwtSvc.GenerateAccessToken(99, domain.RoleAdmin, 3)

	r := chiRouter(OptionalAuth(jwtSvc), okHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if int64(data["user_id"].(float64)) != 99 {
		t.Fatalf("user_id = %v, want 99", data["user_id"])
	}
}

func TestAdmin_NonAdmin(t *testing.T) {
	r := chiRouter(Admin, okHandler)

	ctx := context.WithValue(context.Background(), domain.ContextKeyRole, domain.RoleUser)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

// chiRouter creates a minimal chi router with middleware + handler
func chiRouter(mw func(http.Handler) http.Handler, h http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h)
	return mw(mux)
}
