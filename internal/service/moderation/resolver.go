package moderation

import (
	"log/slog"
	"sync"

	"github.com/atbeta/picfast/internal/sqlc"
)

// Resolver provides the Moderator for the currently configured mode. The
// moderator is rebuilt lazily when the mode string changes, so admin setting
// changes (local or propagated via the site-settings refresher) take effect
// without a restart.
type Resolver struct {
	db *sqlc.Queries

	mu            sync.Mutex
	mode          Mode
	current       Moderator
	warnedInvalid map[string]bool
}

func NewResolver(db *sqlc.Queries) *Resolver {
	return &Resolver{db: db, warnedInvalid: map[string]bool{}}
}

// For returns the Moderator for modeString. Invalid values fall back to
// ModeDisabled with a warning logged once per distinct value; construction
// failures fall back to NoopModerator (cached so it does not warn per request).
func (r *Resolver) For(modeString string) Moderator {
	mode, err := ParseMode(modeString)
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		if !r.warnedInvalid[modeString] {
			r.warnedInvalid[modeString] = true
			slog.Warn("invalid moderation mode, disabling moderation", "mode", modeString, "error", err)
		}
		mode = ModeDisabled
	}
	if mode == "" {
		mode = ModeDisabled
	}
	if mode == r.mode && r.current != nil {
		return r.current
	}
	m, err := New(mode, r.db)
	if err != nil {
		slog.Warn("failed to create moderator", "mode", mode, "error", err)
		m = NewNoopModerator()
	}
	r.mode = mode
	r.current = m
	return m
}
