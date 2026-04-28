package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/router"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

type testEnv struct {
	Pool   *pgxpool.Pool
	DB     *sqlc.Queries
	JWT    *handler.JWTService
	Router http.Handler
	Config *config.Config
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	pool, db := testutil.SetupDB(t)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:    0,
			BaseURL: "http://localhost:0",
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			AccessTTL:  time.Hour,
			RefreshTTL: 24 * time.Hour,
		},
		Storage: config.StorageConfig{
			LocalRoot:    t.TempDir() + "/uploads",
			ThumbnailDir: t.TempDir() + "/thumbnails",
		},
		App: config.AppConfig{
			Name:                "TestAPI",
			AllowGuestUpload:    true,
			AllowRegistration:   true,
			UserInitialCapacity: 524288000,
		},
	}

	jwtSvc := handler.NewJWTService(&cfg.JWT)
	r := router.New(db, pool, cfg, jwtSvc, nil)

	return &testEnv{
		Pool:   pool,
		DB:     db,
		JWT:    jwtSvc,
		Router: r,
		Config: cfg,
	}
}

func (e *testEnv) seedSetup(t *testing.T) (sqlc.Group, sqlc.Strategy, sqlc.User) {
	t.Helper()
	group := testutil.SeedDefaultGroup(t, e.DB)
	strategy := testutil.SeedStrategy(t, e.DB, group.ID)
	user := testutil.SeedUser(t, e.DB, group.ID, "test@example.com", "password123", string(domain.RoleUser))
	return group, strategy, user
}

func (e *testEnv) generateToken(t *testing.T, userID int64, role domain.UserRole, groupID int64) string {
	t.Helper()
	token, _, err := e.JWT.GenerateAccessToken(userID, role, groupID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func (e *testEnv) authReq(t *testing.T, method, path string, body interface{}, userID int64, role domain.UserRole, groupID int64) *http.Request {
	t.Helper()
	req := newJSONReq(t, method, path, body)
	token := e.generateToken(t, userID, role, groupID)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func doReq(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func parseResp(t *testing.T, rec *httptest.ResponseRecorder) handler.Response {
	t.Helper()
	var resp handler.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func newJSONReq(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func uploadReq(t *testing.T, path string, fileName string, fileData []byte, token string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	part.Write(fileData)
	w.Close()

	req := httptest.NewRequest("POST", path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func respDataMap(t *testing.T, resp handler.Response) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	return m
}
