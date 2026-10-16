package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/testutil"
)

// Production databases are created by running the migrations and nothing else.
// testutil.SetupDB truncates every table afterwards, which masked the fact that
// migration 008 used to seed a local storage strategy: bootstrapCoreData then
// always saw "strategies already exist" and silently ignored
// PICFAST_STORAGE_DEFAULT_TYPE / PICFAST_STORAGE_DEFAULT_CONFIG.
func TestBootstrapSeedsConfiguredStrategyOnMigratedDatabase(t *testing.T) {
	pool, queries := testutil.SetupDBWithoutTruncate(t)
	ctx := context.Background()

	var seeded int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM strategies").Scan(&seeded); err != nil {
		t.Fatalf("count strategies: %v", err)
	}
	if seeded != 0 {
		t.Fatalf("migrations seeded %d storage strategies, want 0: the default strategy belongs to the bootstrap", seeded)
	}

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
				"link_mode": "proxy"
			}`),
		},
	}

	if err := bootstrapCoreData(ctx, queries, cfg); err != nil {
		t.Fatalf("bootstrapCoreData: %v", err)
	}

	strategies, err := queries.ListStrategies(ctx)
	if err != nil {
		t.Fatalf("ListStrategies: %v", err)
	}
	if len(strategies) != 1 {
		t.Fatalf("strategies = %d, want 1", len(strategies))
	}
	if got := strategies[0].StrategyType; got != "s3" {
		t.Fatalf("strategy type = %q, want s3", got)
	}
	if got := strategies[0].Name; got != "Default Storage" {
		t.Fatalf("strategy name = %q, want Default Storage", got)
	}
	var parsed map[string]any
	if err := json.Unmarshal(strategies[0].Configs, &parsed); err != nil {
		t.Fatalf("unmarshal strategy config: %v", err)
	}
	if parsed["bucket"] != "pics" || parsed["link_mode"] != "proxy" {
		t.Fatalf("unexpected strategy config: %s", strategies[0].Configs)
	}

	// Both seeded/created groups must point at the configured strategy.
	var linked int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM group_strategies").Scan(&linked); err != nil {
		t.Fatalf("count group_strategies: %v", err)
	}
	if linked != 2 {
		t.Fatalf("group_strategies = %d, want 2 (default + guest)", linked)
	}
}
