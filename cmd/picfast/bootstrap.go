package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/pglock"
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

// bootstrapLockTimeout bounds how long an instance waits for another instance
// to finish seeding a fresh database, and also bounds the wait for a pooled
// connection to take the lock. A var so tests can shorten it. The seeding pass
// itself takes milliseconds, so this stays well inside PaaS boot-probe windows.
var bootstrapLockTimeout = 10 * time.Second

// bootstrapLockRetry is the polling interval while waiting for the lock.
const bootstrapLockRetry = 200 * time.Millisecond

// bootstrapCoreData seeds the rows the app cannot start without: the default
// and guest groups, a default storage strategy, and the admin user (when
// configured). Instances sharing a fresh database would otherwise race and seed
// duplicates, so the whole pass runs under an advisory lock; the losers wait for
// the winner and then find the rows already in place.
func bootstrapCoreData(ctx context.Context, pool *pgxpool.Pool, queries *sqlc.Queries, cfg *config.Config) error {
	release, err := acquireBootstrapLock(ctx, pool)
	if err != nil {
		return err
	}
	defer release()

	defaultGroup, err := ensureDefaultGroup(ctx, queries)
	if err != nil {
		return fmt.Errorf("ensure default group: %w", err)
	}

	guestGroup, err := ensureGuestGroup(ctx, queries)
	if err != nil {
		return fmt.Errorf("ensure guest group: %w", err)
	}

	strategy, err := ensureDefaultStrategy(ctx, queries, cfg)
	if err != nil {
		return fmt.Errorf("ensure default strategy: %w", err)
	}

	_ = queries.AddGroupStrategy(ctx, sqlc.AddGroupStrategyParams{
		GroupID:    defaultGroup.ID,
		StrategyID: strategy.ID,
	})
	_ = queries.AddGroupStrategy(ctx, sqlc.AddGroupStrategyParams{
		GroupID:    guestGroup.ID,
		StrategyID: strategy.ID,
	})

	if err := ensureAdminUser(ctx, queries, cfg, defaultGroup.ID); err != nil {
		return fmt.Errorf("ensure admin user: %w", err)
	}

	return nil
}

func ensureDefaultGroup(ctx context.Context, queries *sqlc.Queries) (sqlc.Group, error) {
	group, err := queries.GetDefaultGroup(ctx)
	if err == nil && group.ID != 0 {
		if group.IsGuest {
			updated, updateErr := queries.UpdateGroup(ctx, sqlc.UpdateGroupParams{
				ID:        group.ID,
				Name:      group.Name,
				IsDefault: true,
				IsGuest:   false,
				Configs:   group.Configs,
			})
			if updateErr != nil {
				return sqlc.Group{}, updateErr
			}
			group = updated
			slog.Info("normalized default group", "id", group.ID)
		}
		return group, nil
	}

	created, err := queries.CreateGroup(ctx, sqlc.CreateGroupParams{
		Name:      "Default",
		IsDefault: true,
		IsGuest:   false,
		Configs:   defaultBootstrapGroupConfig(),
	})
	if err == nil {
		slog.Info("created default group", "id", created.ID)
	}
	return created, err
}

func ensureGuestGroup(ctx context.Context, queries *sqlc.Queries) (sqlc.Group, error) {
	group, err := queries.GetGuestGroup(ctx)
	if err == nil && group.ID != 0 {
		if group.IsDefault {
			group = sqlc.Group{}
		} else {
			return group, nil
		}
	}

	if group.ID == 0 {
		created, err := queries.CreateGroup(ctx, sqlc.CreateGroupParams{
			Name:      "Guest",
			IsDefault: false,
			IsGuest:   true,
			Configs:   defaultBootstrapGroupConfig(),
		})
		if err == nil {
			slog.Info("created guest group", "id", created.ID)
		}
		return created, err
	}

	updated, err := queries.UpdateGroup(ctx, sqlc.UpdateGroupParams{
		ID:        group.ID,
		Name:      group.Name,
		IsDefault: false,
		IsGuest:   true,
		Configs:   group.Configs,
	})
	if err == nil {
		slog.Info("normalized guest group", "id", updated.ID)
	}
	return updated, err
}

func ensureDefaultStrategy(ctx context.Context, queries *sqlc.Queries, cfg *config.Config) (sqlc.Strategy, error) {
	strategies, err := queries.ListStrategies(ctx)
	if err == nil && len(strategies) > 0 {
		if !isLocalDefaultType(cfg.Storage.DefaultType) {
			slog.Warn("ignoring storage.default_type because storage strategies already exist",
				"type", cfg.Storage.DefaultType,
				"strategies", len(strategies),
				"hint", "add the strategy in the admin console, or remove the existing strategies so the environment default can seed one")
		}
		if len(cfg.Storage.DefaultConfig) > 0 && isLocalDefaultType(cfg.Storage.DefaultType) {
			slog.Warn("ignoring storage.default_config because storage.default_type is local",
				"hint", "set PICFAST_STORAGE_DEFAULT_TYPE (for example s3) as well")
		}
		if strategies[0].StrategyType == string(domain.StrategyTypeLocal) {
			updated, updateErr := queries.UpdateStrategy(ctx, sqlc.UpdateStrategyParams{
				ID:           strategies[0].ID,
				Name:         "Local Storage",
				StrategyType: string(domain.StrategyTypeLocal),
				Configs:      localBootstrapStrategyConfig(cfg.Storage.LocalRoot, cfg.Server.BaseURL+"/i"),
			})
			if updateErr == nil {
				return updated, nil
			}
		}
		return strategies[0], nil
	}

	if isLocalDefaultType(cfg.Storage.DefaultType) {
		created, err := queries.CreateStrategy(ctx, sqlc.CreateStrategyParams{
			Name:         "Local Storage",
			StrategyType: string(domain.StrategyTypeLocal),
			Configs:      localBootstrapStrategyConfig(cfg.Storage.LocalRoot, cfg.Server.BaseURL+"/i"),
		})
		if err == nil {
			slog.Info("created local storage strategy", "id", created.ID)
		}
		return created, err
	}

	defaultType := strings.TrimSpace(cfg.Storage.DefaultType)
	if !storage.IsKnownType(defaultType) {
		return sqlc.Strategy{}, fmt.Errorf("unknown storage default_type %q", defaultType)
	}
	if err := storage.ValidateConfig(defaultType, cfg.Storage.DefaultConfig); err != nil {
		return sqlc.Strategy{}, fmt.Errorf("invalid storage default_config for type %q: %w", defaultType, err)
	}
	created, err := queries.CreateStrategy(ctx, sqlc.CreateStrategyParams{
		Name:         "Default Storage",
		StrategyType: defaultType,
		Configs:      cfg.Storage.DefaultConfig,
	})
	if err == nil {
		// Never log the config: it contains storage credentials.
		slog.Info("created default storage strategy", "id", created.ID, "type", defaultType)
	}
	return created, err
}

func isLocalDefaultType(typ string) bool {
	t := strings.TrimSpace(typ)
	return t == "" || t == string(domain.StrategyTypeLocal)
}

func ensureAdminUser(ctx context.Context, queries *sqlc.Queries, cfg *config.Config, groupID int64) error {
	if cfg.App.AdminEmail == "" || cfg.App.AdminPassword == "" {
		return nil
	}

	existing, err := queries.GetUserByEmail(ctx, cfg.App.AdminEmail)
	if err == nil && existing.ID != 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.App.AdminPassword), bcryptCost)
	if err != nil {
		return err
	}

	_, err = queries.CreateAdminUser(ctx, sqlc.CreateAdminUserParams{
		GroupID:       domain.PgInt8(groupID),
		Email:         cfg.App.AdminEmail,
		Password:      pgtype.Text{String: string(hash), Valid: true},
		Name:          "Admin",
		Role:          string(domain.RoleAdmin),
		CapacityBytes: 0,
		Settings:      []byte(`{}`),
		Status:        int16(domain.UserStatusActive),
		EmailVerified: true,
	})
	if err == nil {
		slog.Info("seeded admin user", "email", cfg.App.AdminEmail)
	}
	return err
}

func defaultBootstrapGroupConfig() []byte {
	cfg := domain.GroupConfig{
		MaximumFileSize:    10 << 20,
		AcceptedExtensions: []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "ico"},
		LimitPerMinute:     60,
		LimitPerHour:       500,
		LimitPerDay:        2000,
		LimitPerMonth:      50000,
		UserCapacityBytes:  0,
		PathNamingRule:     "{Y}/{m}/{d}",
		FileNamingRule:     "{uniqid}",
		ImageSaveQuality:   85,
		ImageSaveFormat:    "",
		IsEnableWatermark:  false,
		WatermarkConfigs:   json.RawMessage(`{}`),
		IsStripExif:        true,
	}
	b, _ := json.Marshal(cfg)
	return b
}

func localBootstrapStrategyConfig(root, url string) []byte {
	cfg := domain.LocalStrategyConfig{Root: root, URL: url}
	b, _ := json.Marshal(cfg)
	return b
}

// acquireBootstrapLock takes the first-boot advisory lock, waiting up to
// bootstrapLockTimeout for another instance that is seeding the database right
// now. If the lock cannot be taken in time the pass continues without it: the
// seeding helpers are idempotent, and refusing to start (or blocking forever on
// a saturated connection pool) would be worse than a rare duplicate the console
// can delete.
func acquireBootstrapLock(ctx context.Context, pool *pgxpool.Pool) (func(), error) {
	deadline := time.Now().Add(bootstrapLockTimeout)
	for {
		// Bound the pool wait too: Acquire blocks until a pooled connection is
		// free, and startup must not hang when the pool is saturated.
		acquireCtx, cancel := context.WithTimeout(ctx, bootstrapLockTimeout)
		release, ok, err := pglock.Acquire(acquireCtx, pool, pglock.KeyBootstrapDefaults)
		cancel()
		if err != nil {
			slog.Warn("bootstrap lock unavailable; continuing without it", "error", err)
			return func() {}, nil
		}
		if ok {
			return release, nil
		}
		if time.Now().After(deadline) {
			slog.Warn("continuing without the bootstrap lock; another instance is still seeding",
				"waited", bootstrapLockTimeout)
			return func() {}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(bootstrapLockRetry):
		}
	}
}
