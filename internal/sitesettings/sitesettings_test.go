package sitesettings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
)

func TestApplyMapsFields(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Name = "before"

	err := Apply(cfg, sqlc.SiteSetting{
		AppName:                "PicFast Two",
		AppUrl:                 "https://pics.example.com",
		DefaultImageTtl:        "24h",
		GuestImageTtl:          "1h",
		AllowRegistration:      true,
		ModerationMode:         "manual",
		AnalyticsConfig:        json.RawMessage(`{"a":1}`),
		ThemeConfig:            json.RawMessage(`{}`),
		AllowOauthRegistration: true,
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	app := cfg.AppSnapshot()
	if app.Name != "PicFast Two" {
		t.Fatalf("name = %q", app.Name)
	}
	if app.DefaultImageTTL != 24*time.Hour || app.GuestImageTTL != time.Hour {
		t.Fatalf("ttls = %v / %v", app.DefaultImageTTL, app.GuestImageTTL)
	}
	if !app.AllowRegistration || !app.AllowOauthRegistration || app.ModerationMode != "manual" {
		t.Fatalf("app flags not applied: %+v", app)
	}
	if cfg.ServerSnapshot().BaseURL != "https://pics.example.com" {
		t.Fatalf("base_url = %q", cfg.ServerSnapshot().BaseURL)
	}
}

// A bad TTL must fail the apply before any field is written.
func TestApplyTTLParseErrorLeavesConfigUntouched(t *testing.T) {
	cfg := &config.Config{}
	cfg.App.Name = "keep"

	err := Apply(cfg, sqlc.SiteSetting{AppName: "changed", DefaultImageTtl: "bogus"})
	if err == nil {
		t.Fatal("expected TTL parse error")
	}
	if cfg.AppSnapshot().Name != "keep" {
		t.Fatalf("config partially applied: name = %q", cfg.AppSnapshot().Name)
	}
}

func TestLoadNoRowIsNoop(t *testing.T) {
	_, q := testutil.SetupDB(t)
	cfg := &config.Config{}
	cfg.App.Name = "before"

	ts, err := Load(context.Background(), q, cfg)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !ts.IsZero() {
		t.Fatalf("updated_at = %v, want zero", ts)
	}
	if cfg.App.Name != "before" {
		t.Fatalf("config changed on empty settings: %q", cfg.App.Name)
	}
}

// A change saved by "another instance" (a direct DB write) is picked up after
// the refresh interval; requests inside the interval must not re-query.
func TestRefresherReloadsAfterInterval(t *testing.T) {
	pool, q := testutil.SetupDB(t)
	ctx := context.Background()
	cfg := &config.Config{}
	cfg.App.Name = "initial"

	if _, err := q.UpsertSiteSettings(ctx, sqlc.UpsertSiteSettingsParams{
		AppName:         "v1",
		AnalyticsConfig: json.RawMessage(`{}`),
		ThemeConfig:     json.RawMessage(`{}`),
	}); err != nil {
		t.Fatalf("upsert v1: %v", err)
	}
	seeded, err := Load(ctx, q, cfg)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppSnapshot().Name != "v1" {
		t.Fatalf("startup apply: name = %q", cfg.AppSnapshot().Name)
	}

	current := time.Now()
	f := NewRefresher(q, cfg, seeded)
	f.now = func() time.Time { return current }

	var seen string
	handler := f.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = cfg.AppSnapshot().Name
	}))
	serve := func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// First request triggers a check; the row is unchanged since seeding.
	serve()
	if seen != "v1" {
		t.Fatalf("seen = %q, want v1", seen)
	}

	// Another instance writes a newer row.
	if _, err := pool.Exec(ctx,
		"UPDATE site_settings SET app_name = 'v2', updated_at = updated_at + interval '1 second' WHERE id = 1"); err != nil {
		t.Fatalf("update to v2: %v", err)
	}

	// Inside the interval the refresher must not re-query.
	serve()
	if seen != "v1" {
		t.Fatalf("inside interval: seen = %q, want v1", seen)
	}

	// Past the interval the next request picks up v2.
	current = current.Add(2 * refreshInterval)
	serve()
	if seen != "v2" {
		t.Fatalf("after interval: seen = %q, want v2", seen)
	}
}
