package service

import (
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
)

func GetStorageForStrategy(strategy sqlc.Strategy) (storage.Storage, error) {
	return storage.New(strategy.StrategyType, strategy.Configs)
}
