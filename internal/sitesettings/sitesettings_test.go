package sitesettings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		AppName:                   "PicFast Two",
		AppUrl:                    "https://pics.example.com",
		DefaultImageTtl:           "24h",
		GuestImageTtl:             "1h",
		AllowRegistration:         true,
		ModerationMode:            "manual",
		AnalyticsConfig:           json.RawMessage(`{"a":1}`),
		ThemeConfig:               json.RawMessage(`{}`),
		AllowOauthRegistration:    true,
		ShowPoweredBy:             false,
		ShowLoginLink:             false,
		GuestUploadNoticeTitle:    "Notice title",
		GuestUploadNoticeSubtitle: "Notice subtitle",
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
	// cfg defaults are true for the toggles below; set distinct sentinel values
	// via Apply to prove they propagated.
	if app.ShowPoweredBy || app.ShowLoginLink {
		t.Fatalf("toggles not applied: %+v", app)
	}
	if app.GuestUploadNoticeTitle != "Notice title" || app.GuestUploadNoticeSubtitle != "Notice subtitle" {
		t.Fatalf("guest notice not applied: %+v", app)
	}
}

// TestApplyCoversAllSiteSettingFields fails when a new sqlc.SiteSetting column
// is neither applied by Apply nor explicitly allowlisted here, forcing a
// decision on whether it should propagate to other instances.
func TestApplyCoversAllSiteSettingFields(t *testing.T) {
	// Columns intentionally not applied: row metadata only.
	unapplied := map[string]bool{
		"ID":        true,
		"CreatedAt": true,
		"UpdatedAt": true,
	}

	// field name -> probe asserting the applied value on the config snapshot.
	cfg := &config.Config{}
	err := Apply(cfg, sqlc.SiteSetting{
		AppName:                   "n",
		AppUrl:                    "u",
		AllowGuestUpload:          true,
		AllowRegistration:         true,
		RequireEmailVerification:  true,
		UserInitialCapacity:       123,
		DefaultImageTtl:           "2h",
		ModerationMode:            "manual",
		SiteDescription:           "d",
		AnalyticsProvider:         "plausible",
		AnalyticsConfig:           json.RawMessage(`{"x":1}`),
		FaviconUrl:                "f",
		GuestCapacityBytes:        456,
		AllowUserImageProcessing:  true,
		FooterText1:               "t1",
		FooterLink1:               "l1",
		FooterText2:               "t2",
		FooterLink2:               "l2",
		GuestImageTtl:             "3h",
		ThemeConfig:               json.RawMessage(`{"y":2}`),
		AllowOauthRegistration:    true,
		SkipImageProcessing:       true,
		ShowPoweredBy:             true,
		GuestUploadNoticeTitle:    "gt",
		GuestUploadNoticeSubtitle: "gs",
		ShowLoginLink:             true,
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	app := cfg.AppSnapshot()
	server := cfg.ServerSnapshot()

	probes := map[string]func() bool{
		"AppName":                   func() bool { return app.Name == "n" },
		"AppUrl":                    func() bool { return server.BaseURL == "u" },
		"AllowGuestUpload":          func() bool { return app.AllowGuestUpload },
		"AllowRegistration":         func() bool { return app.AllowRegistration },
		"RequireEmailVerification":  func() bool { return app.RequireEmailVerification },
		"UserInitialCapacity":       func() bool { return app.UserInitialCapacity == 123 },
		"DefaultImageTtl":           func() bool { return app.DefaultImageTTL == 2*time.Hour },
		"ModerationMode":            func() bool { return app.ModerationMode == "manual" },
		"SiteDescription":           func() bool { return app.SiteDescription == "d" },
		"AnalyticsProvider":         func() bool { return app.AnalyticsProvider == "plausible" },
		"AnalyticsConfig":           func() bool { return string(app.AnalyticsConfig) == `{"x":1}` },
		"FaviconUrl":                func() bool { return app.FaviconURL == "f" },
		"GuestCapacityBytes":        func() bool { return app.GuestCapacityBytes == 456 },
		"AllowUserImageProcessing":  func() bool { return app.AllowUserImageProcessing },
		"FooterText1":               func() bool { return app.FooterText1 == "t1" },
		"FooterLink1":               func() bool { return app.FooterLink1 == "l1" },
		"FooterText2":               func() bool { return app.FooterText2 == "t2" },
		"FooterLink2":               func() bool { return app.FooterLink2 == "l2" },
		"GuestImageTtl":             func() bool { return app.GuestImageTTL == 3*time.Hour },
		"ThemeConfig":               func() bool { return string(app.ThemeConfig) == `{"y":2}` },
		"AllowOauthRegistration":    func() bool { return app.AllowOauthRegistration },
		"SkipImageProcessing":       func() bool { return app.SkipImageProcessing },
		"ShowPoweredBy":             func() bool { return app.ShowPoweredBy },
		"GuestUploadNoticeTitle":    func() bool { return app.GuestUploadNoticeTitle == "gt" },
		"GuestUploadNoticeSubtitle": func() bool { return app.GuestUploadNoticeSubtitle == "gs" },
		"ShowLoginLink":             func() bool { return app.ShowLoginLink },
	}

	typ := reflect.TypeOf(sqlc.SiteSetting{})
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if unapplied[name] {
			continue
		}
		probe, ok := probes[name]
		if !ok {
			t.Errorf("SiteSetting.%s is neither applied nor allowlisted — update Apply or the allowlist", name)
			continue
		}
		if !probe() {
			t.Errorf("SiteSetting.%s was not applied to config", name)
		}
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
