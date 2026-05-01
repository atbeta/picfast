package main

import (
	"context"
	"encoding/json"
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
