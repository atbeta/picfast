package service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/events"
	picmetrics "github.com/atbeta/picfast/internal/metrics"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeleteService struct {
	db       *sqlc.Queries
	pool     *pgxpool.Pool
	thumbDir string
	emitter  events.Emitter
}

func NewDeleteService(db *sqlc.Queries, pool *pgxpool.Pool, thumbDir string, emitter events.Emitter) *DeleteService {
	return &DeleteService{db: db, pool: pool, thumbDir: thumbDir, emitter: emitter}
}

func (s *DeleteService) DeleteImage(ctx context.Context, imgID int64) error {
	img, err := s.db.GetImageByID(ctx, imgID)
	if err != nil {
		return err
	}
	return s.deleteImageRecord(ctx, img)
}

func (s *DeleteService) CleanExpiredImages(ctx context.Context, batchSize int32) (int, error) {
	start := time.Now()
	expired, err := s.db.GetExpiredImages(ctx, batchSize)
	if err != nil {
		picmetrics.ObserveCleanupRun("error", time.Since(start))
		return 0, err
	}
	if len(expired) == 0 {
		picmetrics.ObserveCleanupRun("success", time.Since(start))
		return 0, nil
	}

	ctx = context.WithValue(ctx, domain.ContextKeyDeletedBy, "system")

	deleted := 0
	var deletedBytes int64
	for _, img := range expired {
		if err := s.deleteImageRecord(ctx, img); err != nil {
			slog.Warn("failed to delete expired image", "image_id", img.ID, "error", err)
			continue
		}
		deleted++
		deletedBytes += img.SizeBytes
	}
	picmetrics.ObserveCleanupRun("success", time.Since(start))
	picmetrics.ObserveCleanupDeleted("expired", deleted, deletedBytes)
	return deleted, nil
}

func (s *DeleteService) deleteImageRecord(ctx context.Context, img sqlc.Image) error {
	// Sibling rows sharing this md5 decide two things: a same-strategy
	// md5+sha1 dup shares the stored object and its thumbnail (keep both),
	// and any md5 sibling keeps the legacy on-disk thumbnail fallback file.
	hasMd5Siblings := false
	hasStorageDup := false
	if siblings, err := s.db.GetImagesByMD5(ctx, img.Md5); err == nil {
		for _, sib := range siblings {
			if sib.ID == img.ID {
				continue
			}
			hasMd5Siblings = true
			if sib.StrategyID == img.StrategyID && sib.Sha1 == img.Sha1 {
				hasStorageDup = true
			}
		}
	}

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
	// FindDuplicateImage has no id exclusion and can return img itself; the
	// sibling scan is the authoritative same-strategy dup check.
	if hasStorageDup {
		shouldDeleteFile = false
	}

	if shouldDeleteFile && img.StrategyID.Valid {
		strategy, err := s.db.GetStrategyByID(ctx, img.StrategyID.Int64)
		if err == nil {
			store, err := GetStorageForStrategy(strategy)
			if err == nil {
				defer store.Close()
				pathname := img.Name
				if img.Path != "" && img.Path != "." {
					pathname = img.Path + "/" + img.Name
				}
				if err := store.Delete(ctx, pathname); err != nil {
					slog.Warn("failed to delete file from storage", "error", err)
				}
				// Same-md5 siblings in other strategies never read this
				// strategy's thumbnail object, and same-strategy dups are
				// covered by shouldDeleteFile=false.
				if err := store.Delete(ctx, ThumbnailObjectPath(img.Md5)); err != nil && !isNotFoundError(err) {
					slog.Warn("failed to delete thumbnail from storage", "md5", img.Md5, "error", err)
				}
			}
		}
	}

	// The on-disk thumbnail dir is a legacy fallback; only remove the file
	// when no other image row shares this md5.
	if s.thumbDir != "" && !hasMd5Siblings {
		_ = os.Remove(filepath.Join(s.thumbDir, img.Md5+".png"))
	}

	userID := img.UserID
	albumID := img.AlbumID

	txErr := sqlc.RunInTx(ctx, s.pool, func(qtx *sqlc.Queries) error {
		if err := qtx.DeleteImage(ctx, img.ID); err != nil {
			return err
		}
		if userID.Valid {
			if err := qtx.DecrementUserImageNum(ctx, userID.Int64); err != nil {
				slog.Warn("decrement user image_num failed", "user_id", userID.Int64, "error", err)
			}
		}
		if albumID.Valid {
			if err := qtx.DecrementAlbumImageNum(ctx, albumID.Int64); err != nil {
				slog.Warn("decrement album image_num failed", "album_id", albumID.Int64, "error", err)
			}
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}

	deletedBy, _ := ctx.Value(domain.ContextKeyDeletedBy).(string)
	if deletedBy == "" {
		deletedBy = "owner"
	}

	e := events.BuildImageDeleted(img, deletedBy)
	switch deletedBy {
	case "admin":
		e.Actor = events.AdminActor(0)
	case "system":
		e.Actor = events.SystemActor()
	default:
		e.Actor = events.UserActor("", "", domain.PgInt8Val(img.UserID))
	}
	events.EmitBestEffort(ctx, s.emitter, e)

	return nil
}
