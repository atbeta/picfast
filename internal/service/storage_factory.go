package service

import (
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
)

func GetStorageForStrategy(strategy sqlc.Strategy) (storage.Storage, error) {
	s, err := storage.New(strategy.StrategyType, strategy.Configs)
	if err != nil {
		return nil, err
	}
	return storage.WithMetrics(strategy.StrategyType, s), nil
}
