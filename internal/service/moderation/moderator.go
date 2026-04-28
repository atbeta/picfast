package moderation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
)

// Status represents the moderation state of an image.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Mode represents how moderation is configured.
type Mode string

const (
	ModeDisabled Mode = "disabled" // No moderation, images are immediately approved
	ModeManual   Mode = "manual"   // Human moderation required
	ModeAuto     Mode = "auto"     // AI/API moderation (future)
)

// Result is returned by a Moderator after analyzing an image.
type Result struct {
	Status   Status         `json:"status"`
	Score    float64        `json:"score"`    // 0.0 - 1.0 confidence
	Labels   []string       `json:"labels"`   // e.g. ["nsfw", "violence"]
	Reason   string         `json:"reason"`   // human-readable explanation
	Provider string         `json:"provider"` // which engine made the decision
	Extra    map[string]any `json:"extra"`    // provider-specific metadata
}

// Moderator is the interface for content moderation engines.
// Implementations can be:
//   - NoopModerator:   no moderation (immediate approval)
//   - ManualModerator: marks images as pending for human review
//   - AIModerator:     calls external AI service (future)
//   - HybridModerator: local pre-filter + AI + human escalation (future)
type Moderator interface {
	// Name returns the identifier of this moderator.
	Name() string

	// Moderate analyzes an image and returns a moderation result.
	// For synchronous moderators (Noop, Manual) this is fast.
	// For async moderators (AI) this may enqueue a background job.
	Moderate(ctx context.Context, imageID int64, imageKey string, fileData []byte) (*Result, error)

	// SupportsAsync returns true if this moderator operates asynchronously.
	// When true, the upload pipeline will save the image as pending and
	// the moderator will update the status later via a callback/queue.
	SupportsAsync() bool
}

// NoopModerator immediately approves all images.
type NoopModerator struct{}

func NewNoopModerator() *NoopModerator { return &NoopModerator{} }

func (m *NoopModerator) Name() string { return "noop" }

func (m *NoopModerator) Moderate(ctx context.Context, imageID int64, imageKey string, fileData []byte) (*Result, error) {
	return &Result{
		Status:   StatusApproved,
		Score:    0,
		Provider: "noop",
	}, nil
}

func (m *NoopModerator) SupportsAsync() bool { return false }

// ManualModerator marks every image as pending for human review.
type ManualModerator struct {
	DB *sqlc.Queries
}

func NewManualModerator(db *sqlc.Queries) *ManualModerator {
	return &ManualModerator{DB: db}
}

func (m *ManualModerator) Name() string { return "manual" }

func (m *ManualModerator) Moderate(ctx context.Context, imageID int64, imageKey string, fileData []byte) (*Result, error) {
	// Record a moderation entry so it shows up in the review queue
	_, err := m.DB.CreateImageModeration(ctx, sqlc.CreateImageModerationParams{
		ImageID:  imageID,
		Status:   string(StatusPending),
		Provider: "manual",
	})
	if err != nil {
		slog.Warn("failed to create moderation record", "image_id", imageID, "error", err)
	}

	return &Result{
		Status:   StatusPending,
		Score:    0,
		Provider: "manual",
		Reason:   "awaiting human review",
	}, nil
}

func (m *ManualModerator) SupportsAsync() bool { return false }

// ModConstructor creates a Moderator. db may be nil for modes that don't need it.
type ModConstructor func(db *sqlc.Queries) (Moderator, error)

var modRegistry = map[Mode]ModConstructor{}
var modeAliases = map[string]Mode{}

// RegisterMode registers a moderation mode with its constructor and aliases.
// Call from init() in each moderator implementation.
func RegisterMode(mode Mode, aliases []string, ctor ModConstructor) {
	modRegistry[mode] = ctor
	for _, a := range aliases {
		modeAliases[a] = mode
	}
	modeAliases[string(mode)] = mode
}

// New creates a Moderator based on the configured mode.
func New(mode Mode, db *sqlc.Queries) (Moderator, error) {
	ctor, ok := modRegistry[mode]
	if !ok {
		return nil, fmt.Errorf("unknown moderation mode: %s", mode)
	}
	return ctor(db)
}

// ParseMode converts a string to a Mode with validation.
func ParseMode(s string) (Mode, error) {
	if mode, ok := modeAliases[s]; ok {
		return mode, nil
	}
	return "", fmt.Errorf("invalid moderation mode: %s", s)
}

func init() {
	RegisterMode(ModeDisabled, []string{"off", "none"}, func(db *sqlc.Queries) (Moderator, error) {
		return NewNoopModerator(), nil
	})
	RegisterMode(ModeManual, []string{"human"}, func(db *sqlc.Queries) (Moderator, error) {
		if db == nil {
			return nil, fmt.Errorf("manual moderator requires database")
		}
		return NewManualModerator(db), nil
	})
}

// ModerationContext carries moderation state through the request context.
type contextKey string

const moderatorKey contextKey = "moderator"

// WithModerator attaches a moderator to the context.
func WithModerator(ctx context.Context, mod Moderator) context.Context {
	return context.WithValue(ctx, moderatorKey, mod)
}

// FromContext retrieves the moderator from context, or returns nil.
func FromContext(ctx context.Context) Moderator {
	v := ctx.Value(moderatorKey)
	if v == nil {
		return nil
	}
	return v.(Moderator)
}

// UpdateImageModeration updates the moderation status of an image and records the audit log.
func UpdateImageModeration(ctx context.Context, db *sqlc.Queries, imageID int64, status Status, moderatorID int64, reason string) error {
	// Update image status
	_, err := db.UpdateImageModerationStatus(ctx, sqlc.UpdateImageModerationStatusParams{
		ID:               imageID,
		ModerationStatus: string(status),
	})
	if err != nil {
		return fmt.Errorf("update image status: %w", err)
	}

	// Update moderation record
	_, err = db.UpdateImageModeration(ctx, sqlc.UpdateImageModerationParams{
		ImageID:     imageID,
		Status:      string(status),
		ModeratorID: domain.PgInt8(moderatorID),
		Reason:      reason,
	})
	if err != nil {
		return fmt.Errorf("update moderation record: %w", err)
	}

	return nil
}
