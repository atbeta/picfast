package moderation

import (
	"context"
	"testing"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestNoopModerator(t *testing.T) {
	mod := NewNoopModerator()

	if mod.Name() != "noop" {
		t.Errorf("Name() = %q, want %q", mod.Name(), "noop")
	}
	if mod.SupportsAsync() {
		t.Error("SupportsAsync() = true, want false")
	}

	ctx := context.Background()
	result, err := mod.Moderate(ctx, 1, "test-key", []byte("fake-data"))
	if err != nil {
		t.Fatalf("Moderate error = %v, want nil", err)
	}
	if result.Status != StatusApproved {
		t.Errorf("Status = %v, want %v", result.Status, StatusApproved)
	}
	if result.Provider != "noop" {
		t.Errorf("Provider = %q, want %q", result.Provider, "noop")
	}
}

func TestManualModerator(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	mod := NewManualModerator(db)
	if mod.Name() != "manual" {
		t.Errorf("Name() = %q, want manual", mod.Name())
	}

	ctx := context.Background()
	result, err := mod.Moderate(ctx, 1, "test-key", []byte("fake-data"))
	if err != nil {
		t.Fatalf("Moderate error = %v", err)
	}
	if result.Status != StatusPending {
		t.Errorf("Status = %v, want %v", result.Status, StatusPending)
	}
	if result.Provider != "manual" {
		t.Errorf("Provider = %q, want manual", result.Provider)
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		input    string
		wantMode Mode
		wantErr  bool
	}{
		{"disabled", ModeDisabled, false},
		{"off", ModeDisabled, false},
		{"none", ModeDisabled, false},
		{"manual", ModeManual, false},
		{"human", ModeManual, false},
		{"auto", "", true},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mode, err := ParseMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMode(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && mode != tt.wantMode {
				t.Errorf("ParseMode(%q) = %v, want %v", tt.input, mode, tt.wantMode)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("disabled mode", func(t *testing.T) {
		mod, err := New(ModeDisabled, nil)
		if err != nil {
			t.Fatalf("New error = %v, want nil", err)
		}
		if mod.Name() != "noop" {
			t.Errorf("Name() = %q, want %q", mod.Name(), "noop")
		}
	})

	t.Run("manual mode with nil db", func(t *testing.T) {
		_, err := New(ModeManual, nil)
		if err == nil {
			t.Error("New error = nil, want error for nil DB")
		}
	})

	t.Run("unknown mode", func(t *testing.T) {
		_, err := New("unknown", nil)
		if err == nil {
			t.Error("New error = nil, want error for unknown mode")
		}
	})
}

func TestModeratorContext(t *testing.T) {
	mod := NewNoopModerator()
	ctx := context.Background()

	modCtx := WithModerator(ctx, mod)
	if got := FromContext(modCtx); got == nil {
		t.Error("FromContext returned nil, want moderator")
	}
	if got := FromContext(ctx); got != nil {
		t.Error("FromContext returned non-nil for plain context, want nil")
	}
}

func TestUpdateImageModeration(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	group := testutil.SeedDefaultGroup(t, db)
	strategy := testutil.SeedStrategy(t, db, group.ID)
	user := testutil.SeedUser(t, db, group.ID, "mod-test@example.com", "password", "user")

	ctx := context.Background()

	image, err := db.CreateImage(ctx, sqlc.CreateImageParams{
		UserID:     pgtype.Int8{Int64: user.ID, Valid: true},
		AlbumID:    pgtype.Int8{Int64: 0, Valid: false},
		GroupID:    pgtype.Int8{Int64: group.ID, Valid: true},
		StrategyID: pgtype.Int8{Int64: strategy.ID, Valid: true},
		Key:        "test-moderation-img",
		Path:       "/test/path",
		Name:       "test.jpg",
		OriginName: "test.jpg",
		SizeBytes:  1024,
		Mimetype:   "image/jpeg",
		Extension:  ".jpg",
		Md5:        "d41d8cd98f00b204e9800998ecf8427e",
		Sha1:       "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Width:      100,
		Height:     100,
		Permission: 1,
		UploadedIp: "127.0.0.1",
		ExpiresAt:  pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		t.Fatalf("CreateImage error = %v", err)
	}

	_, err = db.CreateImageModeration(ctx, sqlc.CreateImageModerationParams{
		ImageID:  image.ID,
		Status:   string(StatusPending),
		Provider: "manual",
	})
	if err != nil {
		t.Fatalf("CreateImageModeration error = %v", err)
	}

	err = UpdateImageModeration(ctx, db, image.ID, StatusApproved, user.ID, "approved by test")
	if err != nil {
		t.Fatalf("UpdateImageModeration error = %v", err)
	}

	updated, err := db.GetImageByID(ctx, image.ID)
	if err != nil {
		t.Fatalf("GetImageByID error = %v", err)
	}
	if updated.ModerationStatus != string(StatusApproved) {
		t.Errorf("ModerationStatus = %q, want %q", updated.ModerationStatus, StatusApproved)
	}
}
