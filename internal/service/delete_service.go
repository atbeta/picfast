package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
)

type DeleteService struct {
	db       *sqlc.Queries
	pool     *pgxpool.Pool
	thumbDir string
}

func NewDeleteService(db *sqlc.Queries, pool *pgxpool.Pool, thumbDir string) *DeleteService {
	return &DeleteService{db: db, pool: pool, thumbDir: thumbDir}
}

func (s *DeleteService) DeleteImage(ctx context.Context, imgID int64) error {
	img, err := s.db.GetImageByID(ctx, imgID)
	if err != nil {
		return err
	}

	// Check dedup: only delete physical file if no other images share same md5+sha1+strategy
	shouldDeleteFile := true
	if img.StrategyID.Valid {
		dup, err := s.db.FindDuplicateImage(ctx, sqlc.FindDuplicateImageParams{
			StrategyID: img.StrategyID,
			Md5:        img.Md5,
			Sha1:       img.Sha1,
		})
		if err == nil && dup.ID != 0 && dup.ID != img.ID {
			shouldDeleteFile = false
		}
	}

	if shouldDeleteFile && img.StrategyID.Valid {
		strategy, err := s.db.GetStrategyByID(ctx, img.StrategyID.Int64)
		if err == nil {
			store, err := getStorage(strategy)
			if err == nil {
				pathname := img.Path + "/" + img.Name
				if err := store.Delete(ctx, pathname); err != nil {
					slog.Warn("failed to delete file from storage", "error", err)
				}
			}
		}
	}

	// Delete thumbnail
	thumbPath := filepath.Join(s.thumbDir, img.Md5+".png")
	os.Remove(thumbPath)

	// DB: delete record + decrement counters in one transaction
	userID := img.UserID
	albumID := img.AlbumID

	err = sqlc.RunInTx(ctx, s.pool, func(qtx *sqlc.Queries) error {
		if err := qtx.DeleteImage(ctx, imgID); err != nil {
			return err
		}
		if userID.Valid {
			qtx.DecrementUserImageNum(ctx, userID.Int64)
		}
		if albumID.Valid {
			qtx.DecrementAlbumImageNum(ctx, albumID.Int64)
		}
		return nil
	})
	return err
}

func getStorage(strategy sqlc.Strategy) (storage.Storage, error) {
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
	return nil, nil
}
