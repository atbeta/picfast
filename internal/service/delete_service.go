package service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
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
	return s.deleteImageRecord(ctx, img)
}

func (s *DeleteService) CleanExpiredImages(ctx context.Context, batchSize int32) (int, error) {
	expired, err := s.db.GetExpiredImages(ctx, batchSize)
	if err != nil {
		return 0, err
	}
	if len(expired) == 0 {
		return 0, nil
	}

	deleted := 0
	for _, img := range expired {
		if err := s.deleteImageRecord(ctx, img); err != nil {
			slog.Warn("failed to delete expired image", "image_id", img.ID, "error", err)
			continue
		}
		deleted++
	}
	return deleted, nil
}

func (s *DeleteService) deleteImageRecord(ctx context.Context, img sqlc.Image) error {
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
			store, err := GetStorageForStrategy(strategy)
			if err == nil {
				pathname := img.Name
				if img.Path != "" && img.Path != "." {
					pathname = img.Path + "/" + img.Name
				}
				if err := store.Delete(ctx, pathname); err != nil {
					slog.Warn("failed to delete file from storage", "error", err)
				}
			}
		}
	}

	thumbPath := filepath.Join(s.thumbDir, img.Md5+".png")
	os.Remove(thumbPath)

	userID := img.UserID
	albumID := img.AlbumID

	return sqlc.RunInTx(ctx, s.pool, func(qtx *sqlc.Queries) error {
		if err := qtx.DeleteImage(ctx, img.ID); err != nil {
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
}


