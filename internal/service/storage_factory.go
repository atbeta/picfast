package service

import (
	"encoding/json"
	"fmt"

	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/service/storage"
	"github.com/pbeta/imgapi/internal/sqlc"
)

// GetStorageForStrategy creates a storage.Storage from a strategy configuration.
func GetStorageForStrategy(strategy sqlc.Strategy) (storage.Storage, error) {
	switch strategy.StrategyType {
	case string(domain.StrategyTypeLocal):
		var cfg domain.LocalStrategyConfig
		if err := json.Unmarshal(strategy.Configs, &cfg); err != nil {
			return nil, err
		}
		return storage.NewLocalStorage(cfg.Root, cfg.URL), nil
	case string(domain.StrategyTypeS3):
		var cfg domain.S3StrategyConfig
		if err := json.Unmarshal(strategy.Configs, &cfg); err != nil {
			return nil, err
		}
		return storage.NewS3Storage(cfg)
	}
	return nil, fmt.Errorf("unknown strategy type: %s", strategy.StrategyType)
}
