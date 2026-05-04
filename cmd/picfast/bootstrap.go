package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

func bootstrapCoreData(ctx context.Context, queries *sqlc.Queries, cfg *config.Config) error {
	defaultGroup, err := ensureDefaultGroup(ctx, queries)
	if err != nil {
		return fmt.Errorf("ensure default group: %w", err)
	}

	guestGroup, err := ensureGuestGroup(ctx, queries)
	if err != nil {
		return fmt.Errorf("ensure guest group: %w", err)
	}

	strategy, err := ensureLocalStrategy(ctx, queries, cfg)
	if err != nil {
		return fmt.Errorf("ensure local strategy: %w", err)
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

func ensureLocalStrategy(ctx context.Context, queries *sqlc.Queries, cfg *config.Config) (sqlc.Strategy, error) {
	strategies, err := queries.ListStrategies(ctx)
	if err == nil && len(strategies) > 0 {
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
		Password:      string(hash),
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
		MaximumFileSize:            10 << 20,
		AcceptedExtensions:         []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "ico"},
		LimitPerMinute:             60,
		LimitPerHour:               500,
		LimitPerDay:                2000,
		LimitPerMonth:              50000,
		PathNamingRule:             "{Y}/{m}/{d}",
		FileNamingRule:             "{uniqid}",
		ImageSaveQuality:           85,
		ImageSaveFormat:            "",
		IsEnableWatermark:          false,
		WatermarkConfigs:    json.RawMessage(`{}`),
		IsStripExif:         true,
	}
	b, _ := json.Marshal(cfg)
	return b
}

func localBootstrapStrategyConfig(root, url string) []byte {
	cfg := domain.LocalStrategyConfig{Root: root, URL: url}
	b, _ := json.Marshal(cfg)
	return b
}
