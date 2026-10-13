// Package sitesettings applies the admin-edited site_settings row onto the
// process config — once at startup, and periodically on each instance so a
// change saved on one instance propagates to the others.
package sitesettings

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/sqlc"
)

// refreshInterval bounds how often a Refresher re-reads site_settings.
const refreshInterval = 10 * time.Second

// Apply copies a site_settings row onto cfg under the config write lock.
// Durations are parsed before taking the lock so a bad value cannot leave a
// half-applied config.
func Apply(cfg *config.Config, s sqlc.SiteSetting) error {
	defaultTTL, err := parseTTL("default_image_ttl", s.DefaultImageTtl)
	if err != nil {
		return err
	}
	guestTTL, err := parseTTL("guest_image_ttl", s.GuestImageTtl)
	if err != nil {
		return err
	}

	cfg.Update(func(c *config.Config) {
		c.App.Name = s.AppName
		c.Server.BaseURL = s.AppUrl
		c.App.SiteDescription = s.SiteDescription
		c.App.FaviconURL = s.FaviconUrl
		c.App.AllowGuestUpload = s.AllowGuestUpload
		c.App.GuestCapacityBytes = s.GuestCapacityBytes
		c.App.AllowRegistration = s.AllowRegistration
		c.App.AllowOauthRegistration = s.AllowOauthRegistration
		c.App.AllowUserImageProcessing = s.AllowUserImageProcessing
		c.App.SkipImageProcessing = s.SkipImageProcessing
		c.App.RequireEmailVerification = s.RequireEmailVerification
		c.App.UserInitialCapacity = s.UserInitialCapacity
		c.App.DefaultImageTTL = defaultTTL
		c.App.GuestImageTTL = guestTTL
		c.App.ModerationMode = s.ModerationMode
		c.App.FooterText1 = s.FooterText1
		c.App.FooterLink1 = s.FooterLink1
		c.App.FooterText2 = s.FooterText2
		c.App.FooterLink2 = s.FooterLink2
		c.App.AnalyticsProvider = s.AnalyticsProvider
		c.App.AnalyticsConfig = s.AnalyticsConfig
		c.App.ThemeConfig = s.ThemeConfig
	})
	return nil
}

func parseTTL(field, raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse persisted %s: %w", field, err)
	}
	return d, nil
}

// Load applies the persisted site_settings row to cfg; a missing row is a
// no-op. It returns the row's updated_at so callers can seed a Refresher.
func Load(ctx context.Context, queries *sqlc.Queries, cfg *config.Config) (time.Time, error) {
	settings, err := queries.GetSiteSettings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	if err := Apply(cfg, settings); err != nil {
		return time.Time{}, err
	}
	return settings.UpdatedAt, nil
}

// Refresher lets every instance notice site_settings changes saved elsewhere.
type Refresher struct {
	queries *sqlc.Queries
	cfg     *config.Config

	// now is injectable for tests.
	now         func() time.Time
	mu          sync.Mutex
	lastApplied time.Time
	lastChecked time.Time
}

// NewRefresher seeds the refresher with the updated_at observed at startup so
// the first check does not reapply needlessly.
func NewRefresher(queries *sqlc.Queries, cfg *config.Config, seeded time.Time) *Refresher {
	return &Refresher{queries: queries, cfg: cfg, lastApplied: seeded, now: time.Now}
}

// Middleware reloads site settings at most once per refreshInterval per
// instance; refresh failures are logged and never fail the request.
func (f *Refresher) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.maybeRefresh(r.Context())
		next.ServeHTTP(w, r)
	})
}

func (f *Refresher) maybeRefresh(ctx context.Context) {
	// Concurrent requests must not pile up on the DB query: only one
	// refresh runs at a time, the rest proceed unblocked.
	if !f.mu.TryLock() {
		return
	}
	defer f.mu.Unlock()

	if now := f.now(); now.Sub(f.lastChecked) < refreshInterval {
		return
	}
	f.lastChecked = f.now()

	settings, err := f.queries.GetSiteSettings(ctx)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("site settings refresh failed", "error", err)
		}
		return
	}
	if !settings.UpdatedAt.After(f.lastApplied) {
		return
	}
	if err := Apply(f.cfg, settings); err != nil {
		slog.Warn("site settings refresh apply failed", "error", err)
		return
	}
	f.lastApplied = settings.UpdatedAt
}
