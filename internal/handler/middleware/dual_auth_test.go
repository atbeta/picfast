package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

func scopeHandler(w http.ResponseWriter, r *http.Request) {
	scopes, _ := r.Context().Value(domain.ContextKeyScopes).([]string)
	handler.Success(w, map[string]interface{}{"scopes": scopes})
}

func seedAPIToken(t *testing.T, db *sqlc.Queries, userID int64, plainToken string, scopes []string) {
	t.Helper()
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		t.Fatalf("marshal scopes: %v", err)
	}
	_, err = db.CreateAPIToken(t.Context(), sqlc.CreateAPITokenParams{
		UserID:    userID,
		Name:      "test token",
		TokenHash: tokenHash,
		Scopes:    scopesJSON,
		ExpiresAt: pgtype.Timestamptz{},
	})
	if err != nil {
		t.Fatalf("CreateAPIToken: %v", err)
	}
}

func TestDualAuth(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	group := testutil.SeedDefaultGroup(t, db)
	user := testutil.SeedUser(t, db, group.ID, "dual@example.com", "password", string(domain.RoleUser))
	jwtSvc := testJWT()

	const plainToken = "img_" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	seedAPIToken(t, db, user.ID, plainToken, []string{"read", "write"})

	chain := func(h http.Handler) http.Handler {
		return DualAuth(NewJWTAuthenticator(jwtSvc), db)(h)
	}

	t.Run("jwt", func(t *testing.T) {
		token, _, _ := jwtSvc.GenerateAccessToken(user.ID, domain.RoleUser, group.ID)
		r := chiRouter(chain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("api token header", func(t *testing.T) {
		r := chiRouter(chain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-Token", plainToken)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("api token bearer", func(t *testing.T) {
		r := chiRouter(chain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+plainToken)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing credentials", func(t *testing.T) {
		r := chiRouter(chain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("expired api token", func(t *testing.T) {
		const expiredPlain = "img_" + "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		hash := sha256.Sum256([]byte(expiredPlain))
		tokenHash := hex.EncodeToString(hash[:])
		scopesJSON, _ := json.Marshal([]string{"read"})
		_, err := db.CreateAPIToken(t.Context(), sqlc.CreateAPITokenParams{
			UserID:    user.ID,
			Name:      "expired",
			TokenHash: tokenHash,
			Scopes:    scopesJSON,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		})
		if err != nil {
			t.Fatalf("CreateAPIToken: %v", err)
		}

		r := chiRouter(chain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-Token", expiredPlain)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestRequireScope(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	group := testutil.SeedDefaultGroup(t, db)
	user := testutil.SeedUser(t, db, group.ID, "scope@example.com", "password", string(domain.RoleUser))
	jwtSvc := testJWT()

	const readToken = "img_" + "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	seedAPIToken(t, db, user.ID, readToken, []string{"read"})

	writeChain := func(h http.Handler) http.Handler {
		return DualAuth(NewJWTAuthenticator(jwtSvc), db)(RequireScope("write")(h))
	}

	t.Run("api token missing write scope", func(t *testing.T) {
		r := chiRouter(writeChain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-Token", readToken)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("jwt bypasses scope check", func(t *testing.T) {
		token, _, _ := jwtSvc.GenerateAccessToken(user.ID, domain.RoleUser, group.ID)
		r := chiRouter(writeChain, scopeHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})
}

func TestOptionalDualAuth(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	group := testutil.SeedDefaultGroup(t, db)
	user := testutil.SeedUser(t, db, group.ID, "opt@example.com", "password", string(domain.RoleUser))
	jwtSvc := testJWT()

	chain := func(h http.Handler) http.Handler {
		return OptionalDualAuth(NewJWTAuthenticator(jwtSvc), db)(h)
	}

	t.Run("no credentials allowed", func(t *testing.T) {
		r := chiRouter(chain, okHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("api token sets user", func(t *testing.T) {
		const plain = "img_" + "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
		seedAPIToken(t, db, user.ID, plain, []string{"read"})

		r := chiRouter(chain, okHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-Token", plain)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})
}
