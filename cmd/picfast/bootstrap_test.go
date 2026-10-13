package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBootstrapCoreData(t *testing.T) {
	pool, queries := testutil.SetupDB(t)

	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "https://picfast.example.com",
		},
		Storage: config.StorageConfig{
			LocalRoot:    "/app/data/uploads",
			ThumbnailDir: "/app/data/thumbnails",
		},
		App: config.AppConfig{
			AdminEmail:    "admin@example.com",
			AdminPassword: "secret-pass",
		},
	}

	ctx := context.Background()
	if err := bootstrapCoreData(ctx, queries, cfg); err != nil {
		t.Fatalf("bootstrapCoreData returned error: %v", err)
	}
	if err := bootstrapCoreData(ctx, queries, cfg); err != nil {
		t.Fatalf("bootstrapCoreData second run returned error: %v", err)
	}

	defaultGroup, err := queries.GetDefaultGroup(ctx)
	if err != nil {
		t.Fatalf("GetDefaultGroup: %v", err)
	}
	if defaultGroup.Name != "Default" {
		t.Fatalf("default group name = %q, want Default", defaultGroup.Name)
	}

	guestGroup, err := queries.GetGuestGroup(ctx)
	if err != nil {
		t.Fatalf("GetGuestGroup: %v", err)
	}
	if guestGroup.Name != "Guest" {
		t.Fatalf("guest group name = %q, want Guest", guestGroup.Name)
	}

	strategies, err := queries.ListStrategies(ctx)
	if err != nil {
		t.Fatalf("ListStrategies: %v", err)
	}
	if len(strategies) != 1 {
		t.Fatalf("strategies = %d, want 1", len(strategies))
	}

	var strategyCfg map[string]string
	if err := json.Unmarshal(strategies[0].Configs, &strategyCfg); err != nil {
		t.Fatalf("unmarshal strategy config: %v", err)
	}
	if strategyCfg["root"] != "/app/data/uploads" {
		t.Fatalf("strategy root = %q, want /app/data/uploads", strategyCfg["root"])
	}
	if strategyCfg["url"] != "https://picfast.example.com/i" {
		t.Fatalf("strategy url = %q, want https://picfast.example.com/i", strategyCfg["url"])
	}

	admin, err := queries.GetUserByEmail(ctx, "admin@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail admin: %v", err)
	}
	if admin.Role != "admin" {
		t.Fatalf("admin role = %q, want admin", admin.Role)
	}

	assertCount(t, pool, "SELECT COUNT(*) FROM groups", 2)
	assertCount(t, pool, "SELECT COUNT(*) FROM strategies", 1)
	assertCount(t, pool, "SELECT COUNT(*) FROM users", 1)
	assertCount(t, pool, "SELECT COUNT(*) FROM group_strategies", 2)
}

func TestBootstrapCoreDataNormalizesLegacyCombinedGroup(t *testing.T) {
	_, queries := testutil.SetupDB(t)

	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "https://picfast.example.com",
		},
		Storage: config.StorageConfig{
			LocalRoot: "/app/data/uploads",
		},
		App: config.AppConfig{
			AdminEmail:    "admin@example.com",
			AdminPassword: "secret-pass",
		},
	}

	legacyGroup, err := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		Name:      "Default",
		IsDefault: true,
		IsGuest:   true,
		Configs:   defaultBootstrapGroupConfig(),
	})
	if err != nil {
		t.Fatalf("create legacy group: %v", err)
	}
	_, err = queries.CreateStrategy(context.Background(), sqlc.CreateStrategyParams{
		Name:         "Default Local",
		StrategyType: string(domain.StrategyTypeLocal),
		Configs:      localBootstrapStrategyConfig("./data/uploads", "/i"),
	})
	if err != nil {
		t.Fatalf("create legacy strategy: %v", err)
	}

	if err := bootstrapCoreData(context.Background(), queries, cfg); err != nil {
		t.Fatalf("bootstrapCoreData returned error: %v", err)
	}

	defaultGroup, err := queries.GetDefaultGroup(context.Background())
	if err != nil {
		t.Fatalf("GetDefaultGroup: %v", err)
	}
	if defaultGroup.ID != legacyGroup.ID {
		t.Fatalf("default group id = %d, want %d", defaultGroup.ID, legacyGroup.ID)
	}
	if defaultGroup.IsGuest {
		t.Fatal("default group should no longer be marked as guest")
	}

	guestGroup, err := queries.GetGuestGroup(context.Background())
	if err != nil {
		t.Fatalf("GetGuestGroup: %v", err)
	}
	if guestGroup.ID == legacyGroup.ID {
		t.Fatal("guest group should be split from legacy default group")
	}
	if guestGroup.IsDefault {
		t.Fatal("guest group should not be marked as default")
	}
}

func TestBootstrapCreatesS3StrategyFromDefaultConfig(t *testing.T) {
	pool, queries := testutil.SetupDB(t)

	cfg := &config.Config{
		Server: config.ServerConfig{BaseURL: "https://picfast.example.com"},
		Storage: config.StorageConfig{
			LocalRoot:   "/app/data/uploads",
			DefaultType: "s3",
			DefaultConfig: json.RawMessage(`{
				"endpoint": "https://s3.example.com",
				"region": "auto",
				"bucket": "pics",
				"access_key": "AKID",
				"secret_key": "SECRET",
				"url": "https://cdn.example.com",
				"use_path_style": true
			}`),
		},
	}

	ctx := context.Background()
	if err := bootstrapCoreData(ctx, queries, cfg); err != nil {
		t.Fatalf("bootstrapCoreData returned error: %v", err)
	}

	strategies, err := queries.ListStrategies(ctx)
	if err != nil {
		t.Fatalf("ListStrategies: %v", err)
	}
	if len(strategies) != 1 {
		t.Fatalf("strategies = %d, want 1", len(strategies))
	}
	strategy := strategies[0]
	if strategy.StrategyType != "s3" {
		t.Fatalf("strategy type = %q, want s3", strategy.StrategyType)
	}
	if strategy.Name != "Default Storage" {
		t.Fatalf("strategy name = %q, want Default Storage", strategy.Name)
	}
	var parsed map[string]any
	if err := json.Unmarshal(strategy.Configs, &parsed); err != nil {
		t.Fatalf("unmarshal strategy config: %v", err)
	}
	if parsed["endpoint"] != "https://s3.example.com" || parsed["bucket"] != "pics" {
		t.Fatalf("unexpected strategy config: %s", strategy.Configs)
	}

	// linked to both default and guest groups
	assertCount(t, pool, "SELECT COUNT(*) FROM group_strategies", 2)
}

func TestBootstrapDefaultStorageInvalidConfig(t *testing.T) {
	t.Run("unknown type", func(t *testing.T) {
		_, queries := testutil.SetupDB(t)
		cfg := &config.Config{
			Storage: config.StorageConfig{
				DefaultType:   "not-a-backend",
				DefaultConfig: json.RawMessage(`{}`),
			},
		}
		err := bootstrapCoreData(context.Background(), queries, cfg)
		if err == nil || !strings.Contains(err.Error(), "unknown storage default_type") {
			t.Fatalf("err = %v, want unknown default_type error", err)
		}
	})

	t.Run("missing config", func(t *testing.T) {
		_, queries := testutil.SetupDB(t)
		cfg := &config.Config{
			Storage: config.StorageConfig{DefaultType: "s3"},
		}
		err := bootstrapCoreData(context.Background(), queries, cfg)
		if err == nil || !strings.Contains(err.Error(), "invalid storage default_config") {
			t.Fatalf("err = %v, want invalid default_config error", err)
		}
	})

	t.Run("incomplete s3 config", func(t *testing.T) {
		_, queries := testutil.SetupDB(t)
		cfg := &config.Config{
			Storage: config.StorageConfig{
				DefaultType:   "s3",
				DefaultConfig: json.RawMessage(`{"endpoint":"https://s3.example.com"}`),
			},
		}
		err := bootstrapCoreData(context.Background(), queries, cfg)
		if err == nil || !strings.Contains(err.Error(), "invalid storage default_config") {
			t.Fatalf("err = %v, want invalid default_config error", err)
		}
	})
}

func TestBootstrapIgnoresDefaultTypeWhenStrategiesExist(t *testing.T) {
	pool, queries := testutil.SetupDB(t)

	_, err := queries.CreateStrategy(context.Background(), sqlc.CreateStrategyParams{
		Name:         "Existing Local",
		StrategyType: string(domain.StrategyTypeLocal),
		Configs:      localBootstrapStrategyConfig("./data/uploads", "/i"),
	})
	if err != nil {
		t.Fatalf("create strategy: %v", err)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{BaseURL: "https://picfast.example.com"},
		Storage: config.StorageConfig{
			LocalRoot:   "/app/data/uploads",
			DefaultType: "s3",
			DefaultConfig: json.RawMessage(`{
				"endpoint": "https://s3.example.com",
				"bucket": "pics",
				"access_key": "AKID",
				"secret_key": "SECRET"
			}`),
		},
	}

	if err := bootstrapCoreData(context.Background(), queries, cfg); err != nil {
		t.Fatalf("bootstrapCoreData returned error: %v", err)
	}

	strategies, err := queries.ListStrategies(context.Background())
	if err != nil {
		t.Fatalf("ListStrategies: %v", err)
	}
	if len(strategies) != 1 {
		t.Fatalf("strategies = %d, want 1 (env default ignored)", len(strategies))
	}
	if strategies[0].StrategyType != string(domain.StrategyTypeLocal) {
		t.Fatalf("strategy type = %q, want local", strategies[0].StrategyType)
	}
	assertCount(t, pool, "SELECT COUNT(*) FROM strategies", 1)
}

func assertCount(t *testing.T, pool *pgxpool.Pool, query string, want int) {
	t.Helper()

	var got int
	if err := pool.QueryRow(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != want {
		t.Fatalf("%s => %d, want %d", query, got, want)
	}
}
