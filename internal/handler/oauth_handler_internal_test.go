package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/oauth"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestMain(m *testing.M) {
	SetTrustedProxies([]string{"192.0.2.0/24", "127.0.0.0/8"})
	os.Exit(m.Run())
}

func TestOAuthStateSignAndVerify(t *testing.T) {
	secret := "test-state-secret-32-bytes-key!!"
	state := "abc123-random-state"

	signed := signState(state, secret)
	if !verifyState(state, signed, secret) {
		t.Fatal("verifyState failed for valid state")
	}

	if verifyState("wrong-state", signed, secret) {
		t.Fatal("verifyState should reject mismatched state")
	}

	tampered := state + "." + "00000000"
	if verifyState(state, tampered, secret) {
		t.Fatal("verifyState should reject tampered signature")
	}

	if verifyState(state, signed, "wrong-secret") {
		t.Fatal("verifyState should reject different secret")
	}
}

func TestOAuthLinkCookieSignAndVerify(t *testing.T) {
	secret := "test-link-secret-key-32-bytes"

	signed := signSignedLinkUserID(42, secret)
	uid, ok := verifySignedLinkUserID(signed, secret)
	if !ok {
		t.Fatal("verifySignedLinkUserID failed for valid signed value")
	}
	if uid != 42 {
		t.Fatalf("expected uid 42, got %d", uid)
	}

	if _, ok := verifySignedLinkUserID("not-valid", secret); ok {
		t.Fatal("verifySignedLinkUserID should reject invalid format")
	}

	tampered := "link:99." + signed[strings.LastIndex(signed, ".")+1:]
	if _, ok := verifySignedLinkUserID(tampered, secret); ok {
		t.Fatal("verifySignedLinkUserID should reject tampered payload")
	}

	if _, ok := verifySignedLinkUserID(signed, "wrong-secret"); ok {
		t.Fatal("verifySignedLinkUserID should reject different secret")
	}

	negSigned := signSignedLinkUserID(-1, secret)
	negUID, ok := verifySignedLinkUserID(negSigned, secret)
	if !ok || negUID != -1 {
		t.Fatal("should handle negative user ids (edge case)")
	}
}

func TestGenerateCodeVerifier(t *testing.T) {
	v, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("generateCodeVerifier failed: %v", err)
	}
	if len(v) < 43 || len(v) > 128 {
		t.Fatalf("code verifier length %d out of PKCE range [43,128]", len(v))
	}

	v2, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("second generateCodeVerifier failed: %v", err)
	}
	if v == v2 {
		t.Fatal("two code verifiers should be different")
	}
}

func TestOAuthHandlerRejectsFrozenUser(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	jwtSvc := NewJWTService(&cfg.JWT)

	group := testutil.SeedDefaultGroup(t, db)

	frozenUser := testutil.SeedUser(t, db, group.ID, "frozen@example.com", "password123", string(domain.RoleUser))
	freezeUser(t, db, frozenUser)

	_, err := db.CreateUserIdentity(t.Context(), sqlc.CreateUserIdentityParams{
		UserID:          frozenUser.ID,
		Provider:        "keycloak",
		ProviderSubject: "frozen-subject-1",
		Email:           "frozen@example.com",
	})
	if err != nil {
		t.Fatalf("create identity: %v", err)
	}

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "frozen-subject-1",
		Email:         "frozen@example.com",
		EmailVerified: true,
		Name:          "Frozen User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	_, err = h.findOrCreateUser(req, "keycloak", identity)
	if err == nil {
		t.Fatal("expected error for frozen user, got nil")
	}
	if err != errAccountDisabled {
		t.Fatalf("expected errAccountDisabled, got %v", err)
	}
}

func TestOAuthHandlerAutoLinkWhenEmailVerified(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	jwtSvc := NewJWTService(&cfg.JWT)

	group := testutil.SeedDefaultGroup(t, db)
	user := testutil.SeedUser(t, db, group.ID, "autolink@example.com", "password123", string(domain.RoleUser))

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "new-keycloak-subject",
		Email:         "autolink@example.com",
		EmailVerified: true,
		Name:          "AutoLink User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	result, err := h.findOrCreateUser(req, "keycloak", identity)
	if err != nil {
		t.Fatalf("auto-link failed: %v", err)
	}
	if result.ID != user.ID {
		t.Fatalf("expected existing user %d, got %d", user.ID, result.ID)
	}

	identity2, err := db.GetUserIdentityByProviderSubject(t.Context(), sqlc.GetUserIdentityByProviderSubjectParams{
		Provider:        "keycloak",
		ProviderSubject: "new-keycloak-subject",
	})
	if err != nil {
		t.Fatalf("identity not created: %v", err)
	}
	if identity2.UserID != user.ID {
		t.Fatalf("identity linked to wrong user: %d", identity2.UserID)
	}
}

func TestOAuthHandlerSkipsAutoLinkWithoutEmailVerification(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	jwtSvc := NewJWTService(&cfg.JWT)

	group := testutil.SeedDefaultGroup(t, db)
	testutil.SeedUser(t, db, group.ID, "unverified@example.com", "password123", string(domain.RoleUser))

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "unverified-subject",
		Email:         "unverified@example.com",
		EmailVerified: false,
		Name:          "Unverified User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	_, err := h.findOrCreateUser(req, "keycloak", identity)
	if err == nil {
		t.Fatal("expected error for email collision with unverified IdP email")
	}
}

func TestWriteAuditLogNilSafe(t *testing.T) {
	writeAuditLog(nil, nil, "test", "test", "", "", nil)
	writeAuditLog(nil, &http.Request{}, "test", "test", "", "", nil)
}

func newOAuthReq(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	rctx := chi.NewRouteContext()
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/auth/oauth/"), "/")
	if len(parts) >= 1 && parts[0] != "" {
		if parts[0] != "callback" && parts[0] != "providers" {
			rctx.URLParams.Add("provider", parts[0])
		}
	}
	if len(parts) >= 2 && parts[1] == "callback" && parts[0] != "" {
		rctx.URLParams.Add("provider", parts[0])
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}

func newOAuthCallbackReq(path string) *http.Request {
	req, _ := http.NewRequest("GET", path, nil)
	rctx := chi.NewRouteContext()
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/auth/oauth/"), "/")
	if len(parts) >= 1 {
		rctx.URLParams.Add("provider", parts[0])
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}

func TestOAuthCallbackRejectsMissingStateCookie(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?state=any&code=any")
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing state cookie, got %d", rec.Code)
	}
}

func TestOAuthCallbackRejectsMissingStateParam(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?code=any")
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookie,
		Value: signState("teststate", h.stateSecret()),
	})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing state param, got %d", rec.Code)
	}
}

func TestOAuthCallbackRejectsTamperedStateCookie(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?state=tampered&code=any")
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookie,
		Value: signState("original", h.stateSecret()),
	})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for tampered state, got %d", rec.Code)
	}
}

func TestOAuthCallbackHandlesProviderError(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	state := "teststate"
	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?state=" + state + "&error=access_denied")
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookie,
		Value: signState(state, h.stateSecret()),
	})
	req.AddCookie(&http.Cookie{
		Name:  oauthPKCECookie,
		Value: "test-pkce",
	})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect for provider error, got %d", rec.Code)
	}
}

func TestOAuthCallbackHandlesEmptyCode(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	state := "teststate"
	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?state=" + state)
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookie,
		Value: signState(state, h.stateSecret()),
	})
	req.AddCookie(&http.Cookie{
		Name:  oauthPKCECookie,
		Value: "test-pkce-verifier",
	})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect for missing code, got %d", rec.Code)
	}
}

func TestOAuthCallbackRejectsMissingPKCE(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	state := "teststate"
	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?state=" + state + "&code=abc")
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookie,
		Value: signState(state, h.stateSecret()),
	})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing PKCE cookie, got %d", rec.Code)
	}
}

func TestOAuthCallbackRejectsEmptyPKCE(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	state := "teststate"
	req := newOAuthCallbackReq("/api/v1/auth/oauth/keycloak/callback?state=" + state + "&code=abc")
	req.AddCookie(&http.Cookie{
		Name:  oauthStateCookie,
		Value: signState(state, h.stateSecret()),
	})
	req.AddCookie(&http.Cookie{
		Name:  oauthPKCECookie,
		Value: "",
	})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty PKCE cookie, got %d", rec.Code)
	}
}

func TestOAuthProvidersReturnsEnabledList(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthReq("GET", "/api/v1/auth/oauth/providers")
	rec := httptest.NewRecorder()
	h.Providers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for providers list, got %d", rec.Code)
	}
}

func TestOAuthStartProviderNotConfigured(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthReq("GET", "/api/v1/auth/oauth/keycloak")
	rec := httptest.NewRecorder()
	h.Start(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for provider without configured issuer, got %d", rec.Code)
	}
}

func TestOAuthStartUnknownProvider(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthReq("GET", "/api/v1/auth/oauth/nonexistent")
	rec := httptest.NewRecorder()
	h.Start(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown provider, got %d", rec.Code)
	}
}

func TestOAuthUnlinkRejectsAccountLockout(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	jwtSvc := NewJWTService(&cfg.JWT)

	group := testutil.SeedDefaultGroup(t, db)
	oidcUser := testutil.SeedUser(t, db, group.ID, "oidc-only@example.com", "password123", string(domain.RoleUser))
	// Simulate OIDC-only user: no password
	clearUserPassword(t, db, oidcUser)

	_, err := db.CreateUserIdentity(t.Context(), sqlc.CreateUserIdentityParams{
		UserID:          oidcUser.ID,
		Provider:        "keycloak",
		ProviderSubject: "sub-abc",
		Email:           "oidc-only@example.com",
	})
	if err != nil {
		t.Fatalf("create identity: %v", err)
	}

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/auth/oauth/keycloak", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "keycloak")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), domain.ContextKeyUserID, oidcUser.ID))
	h.Unlink(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for lockout-prevention unlink, got %d", rec.Code)
	}
}

func TestOAuthLinkRequiresAuth(t *testing.T) {
	cfg := testOAuthConfig()
	h := &OAuthHandler{config: cfg}

	req := newOAuthReq("POST", "/api/v1/auth/oauth/keycloak/link")
	rec := httptest.NewRecorder()
	h.Link(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without JWT context, got %d", rec.Code)
	}
}

func TestOAuthRegistrationDisabledRejectsNewUser(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	cfg.App.AllowOauthRegistration = false
	jwtSvc := NewJWTService(&cfg.JWT)

	testutil.SeedDefaultGroup(t, db)

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "new-oauth-user-disabled",
		Email:         "newuser@example.com",
		EmailVerified: true,
		Name:          "New OAuth User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	_, err := h.findOrCreateUser(req, "keycloak", identity)
	if err == nil {
		t.Fatal("expected error when oauth registration is disabled, got nil")
	}
	if !errors.Is(err, errOAuthRegistrationDisabled) {
		t.Fatalf("expected errOAuthRegistrationDisabled, got: %v", err)
	}
}

func TestOAuthRegistrationEnabledCreatesNewUser(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	jwtSvc := NewJWTService(&cfg.JWT)

	testutil.SeedDefaultGroup(t, db)

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "new-oauth-user-enabled",
		Email:         "newoauth@example.com",
		EmailVerified: true,
		Name:          "New OAuth User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	user, err := h.findOrCreateUser(req, "keycloak", identity)
	if err != nil {
		t.Fatalf("unexpected error when oauth registration is enabled: %v", err)
	}
	if user.Email != "newoauth@example.com" {
		t.Fatalf("expected new user email newoauth@example.com, got %s", user.Email)
	}
}

func TestOAuthRegistrationDisabledAllowsAutoLink(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	cfg.App.AllowOauthRegistration = false
	jwtSvc := NewJWTService(&cfg.JWT)

	group := testutil.SeedDefaultGroup(t, db)
	user := testutil.SeedUser(t, db, group.ID, "existing@example.com", "password123", string(domain.RoleUser))

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "new-oauth-subject-autolink",
		Email:         "existing@example.com",
		EmailVerified: true,
		Name:          "AutoLink User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	result, err := h.findOrCreateUser(req, "keycloak", identity)
	if err != nil {
		t.Fatalf("auto-link should succeed even when oauth registration is disabled: %v", err)
	}
	if result.ID != user.ID {
		t.Fatalf("expected existing user %d, got %d", user.ID, result.ID)
	}
}

func TestOAuthEmailCollisionRejectsUnverifiedIdentity(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	cfg := testOAuthConfig()
	jwtSvc := NewJWTService(&cfg.JWT)

	group := testutil.SeedDefaultGroup(t, db)
	testutil.SeedUser(t, db, group.ID, "collision@example.com", "password123", string(domain.RoleUser))

	h := &OAuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}

	identity := oauth.Identity{
		Subject:       "new-oauth-subject-collision",
		Email:         "collision@example.com",
		EmailVerified: false,
		Name:          "Collision User",
	}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	_, err := h.findOrCreateUser(req, "keycloak", identity)
	if err == nil {
		t.Fatal("expected error for unverified identity with existing email, got nil")
	}
	if !strings.Contains(err.Error(), "email already registered") {
		t.Fatalf("expected email collision error, got: %v", err)
	}
}

func clearUserPassword(t *testing.T, db *sqlc.Queries, user sqlc.User) {
	t.Helper()
	_, err := db.UpdateUser(t.Context(), sqlc.UpdateUserParams{
		ID:            user.ID,
		Name:          user.Name,
		Password:      pgtype.Text{},
		GroupID:       user.GroupID,
		CapacityBytes: user.CapacityBytes,
		Status:        int16(domain.UserStatusActive),
		Settings:      user.Settings,
	})
	if err != nil {
		t.Fatalf("clear password: %v", err)
	}
}

func testOAuthConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-oauth-secret-key-32bytes!",
			AccessTTL:  900,
			RefreshTTL: 604800,
		},
		App: config.AppConfig{
			Name:                  "TestApp",
			AllowRegistration:     true,
			AllowOauthRegistration: true,
			UserInitialCapacity:   524288000,
		},
		OAuth: config.OAuthConfig{
			Providers: []config.OAuthProviderConfig{
				{
					ID:          "keycloak",
					DisplayName: "Keycloak",
					Type:        "oidc",
					Enabled:     true,
				},
			},
		},
	}
}

func freezeUser(t *testing.T, db *sqlc.Queries, user sqlc.User) {
	t.Helper()
	_, err := db.UpdateUser(t.Context(), sqlc.UpdateUserParams{
		ID:            user.ID,
		Name:          user.Name,
		Password:      user.Password,
		GroupID:       user.GroupID,
		CapacityBytes: user.CapacityBytes,
		Status:        int16(domain.UserStatusFrozen),
		Settings:      user.Settings,
	})
	if err != nil {
		t.Fatalf("freeze user: %v", err)
	}
}
