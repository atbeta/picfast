package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	ctx := context.Background()

	// 1. Seed default group
	var defaultGroup sqlc.Group
	existing, _ := queries.GetDefaultGroup(ctx)
	if existing.ID != 0 {
		defaultGroup = existing
		slog.Info("default group already exists", "id", defaultGroup.ID)
	} else {
		group, err := queries.CreateGroup(ctx, sqlc.CreateGroupParams{
			Name:      "Default",
			IsDefault: true,
			IsGuest:   false,
			Configs:   defaultGroupConfig(),
		})
		if err != nil {
			slog.Error("failed to create default group", "error", err)
			os.Exit(1)
		}
		defaultGroup = group
		slog.Info("created default group", "id", defaultGroup.ID)
	}

	// 2. Seed guest group
	var guestGroup sqlc.Group
	existingGuest, _ := queries.GetGuestGroup(ctx)
	if existingGuest.ID != 0 {
		guestGroup = existingGuest
		slog.Info("guest group already exists", "id", guestGroup.ID)
	} else {
		group, err := queries.CreateGroup(ctx, sqlc.CreateGroupParams{
			Name:      "Guest",
			IsDefault: false,
			IsGuest:   true,
			Configs:   defaultGroupConfig(),
		})
		if err != nil {
			slog.Error("failed to create guest group", "error", err)
			os.Exit(1)
		}
		guestGroup = group
		slog.Info("created guest group", "id", guestGroup.ID)
	}

	// 3. Seed local storage strategy
	var strategy sqlc.Strategy
	strategies, _ := queries.ListStrategies(ctx)
	if len(strategies) > 0 {
		strategy = strategies[0]
		slog.Info("strategy already exists", "id", strategy.ID)
	} else {
		st, err := queries.CreateStrategy(ctx, sqlc.CreateStrategyParams{
			Name:         "Local Storage",
			StrategyType: string(domain.StrategyTypeLocal),
			Configs: localStrategyConfig(
				cfg.Storage.LocalRoot,
				cfg.Server.BaseURL+"/i",
			),
		})
		if err != nil {
			slog.Error("failed to create strategy", "error", err)
			os.Exit(1)
		}
		strategy = st
		slog.Info("created local storage strategy", "id", strategy.ID)
	}

	// 4. Link strategy to groups
	_ = queries.AddGroupStrategy(ctx, sqlc.AddGroupStrategyParams{
		GroupID:    defaultGroup.ID,
		StrategyID: strategy.ID,
	})
	_ = queries.AddGroupStrategy(ctx, sqlc.AddGroupStrategyParams{
		GroupID:    guestGroup.ID,
		StrategyID: strategy.ID,
	})

	// 5. Seed test user
	testEmail := "test@example.com"
	existingUser, _ := queries.GetUserByEmail(ctx, testEmail)
	if existingUser.ID != 0 {
		slog.Info("test user already exists", "id", existingUser.ID)
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("failed to hash test user password", "error", err)
			os.Exit(1)
		}
		user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
			GroupID:       domain.PgInt8(defaultGroup.ID),
			Email:         testEmail,
			Password:      string(hash),
			Name:          "Test User",
			Role:          string(domain.RoleUser),
			CapacityBytes: 524288000,
			Settings:      []byte(`{}`),
			Status:        int16(domain.UserStatusActive),
			EmailVerified: true,
			RegisteredIp:  "127.0.0.1",
		})
		if err != nil {
			slog.Error("failed to create test user", "error", err)
		} else {
			slog.Info("created test user", "id", user.ID, "email", testEmail)
		}
	}

	// 6. Seed admin user
	adminEmail := "admin@example.com"
	existingAdmin, _ := queries.GetUserByEmail(ctx, adminEmail)
	if existingAdmin.ID != 0 {
		slog.Info("admin user already exists", "id", existingAdmin.ID)
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("failed to hash admin password", "error", err)
			os.Exit(1)
		}
		admin, err := queries.CreateAdminUser(ctx, sqlc.CreateAdminUserParams{
			GroupID:       domain.PgInt8(defaultGroup.ID),
			Email:         adminEmail,
			Password:      string(hash),
			Name:          "Admin",
			Role:          string(domain.RoleAdmin),
			CapacityBytes: 0,
			Settings:      []byte(`{}`),
			Status:        int16(domain.UserStatusActive),
			EmailVerified: true,
		})
		if err != nil {
			slog.Error("failed to create admin user", "error", err)
		} else {
			slog.Info("created admin user", "id", admin.ID, "email", adminEmail)
		}
	}

	fmt.Println("\n✅ Seed completed!")
	fmt.Println("   Test user:  test@example.com / password123")
	fmt.Println("   Admin user: admin@example.com / admin123")
}

func defaultGroupConfig() []byte {
	cfg := domain.GroupConfig{
		MaximumFileSize:    10 << 20, // 10MB
		AcceptedExtensions: []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "ico"},
		LimitPerMinute:     60,
		LimitPerHour:       500,
		LimitPerDay:        2000,
		LimitPerMonth:      50000,
		PathNamingRule:     "{Y}/{m}/{d}",
		FileNamingRule:     "{uniqid}",
		ImageSaveQuality:   100,
		ImageSaveFormat:    "",
	}
	b, _ := json.Marshal(cfg)
	return b
}

func localStrategyConfig(root, url string) []byte {
	cfg := domain.LocalStrategyConfig{Root: root, URL: url}
	b, _ := json.Marshal(cfg)
	return b
}

