package testutil

import (
	"encoding/json"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"golang.org/x/crypto/bcrypt"
)

// SeedDefaultGroup creates the default+guest group with standard config.
func SeedDefaultGroup(t *testing.T, q *sqlc.Queries) sqlc.Group {
	t.Helper()
	cfg := domain.GroupConfig{
		MaximumFileSize:            5242880,
		AcceptedExtensions:         []string{"jpeg", "jpg", "png", "gif", "webp"},
		LimitPerMinute:             20,
		LimitPerHour:               100,
		LimitPerDay:                300,
		LimitPerMonth:              999,
		PathNamingRule:             "{Y}/{m}/{d}",
		FileNamingRule:             "{uniqid}",
		ImageSaveQuality:           75,
		ImageSaveFormat:            "",
		IsEnableWatermark:          false,
		WatermarkConfigs:  json.RawMessage(`{}`),
	}
	configs, _ := json.Marshal(cfg)

	group, err := q.CreateGroup(t.Context(), sqlc.CreateGroupParams{
		Name:      "Default",
		IsDefault: true,
		IsGuest:   true,
		Configs:   configs,
	})
	if err != nil {
		t.Fatalf("seed default group: %v", err)
	}
	return group
}

// SeedStrategy creates a local strategy and links it to the group.
func SeedStrategy(t *testing.T, q *sqlc.Queries, groupID int64) sqlc.Strategy {
	t.Helper()
	cfg := domain.LocalStrategyConfig{Root: "/tmp/test-uploads", URL: "/i"}
	configs, _ := json.Marshal(cfg)

	strategy, err := q.CreateStrategy(t.Context(), sqlc.CreateStrategyParams{
		Name:         "Test Local",
		StrategyType: "local",
		Configs:      configs,
	})
	if err != nil {
		t.Fatalf("seed strategy: %v", err)
	}
	if err := q.AddGroupStrategy(t.Context(), sqlc.AddGroupStrategyParams{
		GroupID:    groupID,
		StrategyID: strategy.ID,
	}); err != nil {
		t.Fatalf("seed group_strategy: %v", err)
	}
	return strategy
}

// SeedUser creates a user with bcrypt-hashed password.
func SeedUser(t *testing.T, q *sqlc.Queries, groupID int64, email, password, role string) sqlc.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	settings, _ := json.Marshal(domain.UserSettings{})
	user, err := q.CreateUser(t.Context(), sqlc.CreateUserParams{
		GroupID:       domain.PgInt8(groupID),
		Email:         email,
		Password:      string(hash),
		Name:          email,
		Role:          role,
		CapacityBytes: 524288000,
		Settings:      settings,
		Status:        int16(domain.UserStatusActive),
		EmailVerified: false,
		RegisteredIp:  "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return user
}

// SeedAlbum creates an album for the given user.
func SeedAlbum(t *testing.T, q *sqlc.Queries, userID int64, name string) sqlc.Album {
	t.Helper()
	album, err := q.CreateAlbum(t.Context(), sqlc.CreateAlbumParams{
		UserID: userID,
		Name:   name,
		Intro:  "test album",
	})
	if err != nil {
		t.Fatalf("seed album: %v", err)
	}
	return album
}
