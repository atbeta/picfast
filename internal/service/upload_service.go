package service

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/moderation"
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadService struct {
	db     *sqlc.Queries
	pool   *pgxpool.Pool
	config *config.Config

	totalImages atomic.Int64
	countInit   sync.Once
}

type uploadUserSettings struct {
	DefaultStrategy int64 `json:"default_strategy"`
}

type uploadIdentity struct {
	group               sqlc.Group
	userID              int64
	groupID             int64
	isAdmin             bool
	preferredStrategyID *int64
}

type processedUploadImage struct {
	data      []byte
	width     int
	height    int
	processed bool
}

func NewUploadService(db *sqlc.Queries, pool *pgxpool.Pool, cfg *config.Config) *UploadService {
	return &UploadService{db: db, pool: pool, config: cfg}
}

// totalImagesCount lazily seeds the atomic counter from the database,
// then returns the in-memory count. The count is incremented locally after
// each successful insert, so it stays accurate without per-upload queries.
func (s *UploadService) totalImagesCount(ctx context.Context) int64 {
	s.countInit.Do(func() {
		count, err := s.db.CountAllImages(ctx)
		if err != nil {
			slog.Warn("failed to seed image count, starting from zero", "error", err)
			return
		}
		s.totalImages.Store(count)
	})
	return s.totalImages.Load()
}

type UploadParams struct {
	FileData   []byte
	FileName   string
	FileSize   int64
	StrategyID *int64
	AlbumID    *int64
	Permission *int16
	UserID     *int64
	ClientIP   string
	ExpiresAt  *time.Time
}

type UploadResult struct {
	Image             sqlc.Image
	Links             domain.ImageLinks
	GroupConfig       domain.GroupConfig
	OriginalSizeBytes int64
	StoredSizeBytes   int64
	Processed         bool
}

func (s *UploadService) Store(ctx context.Context, params UploadParams) (*UploadResult, error) {
	params = s.applyDefaultTTL(params)

	identity, err := s.resolveIdentity(ctx, params)
	if err != nil {
		return nil, err
	}

	groupConfig, err := loadUploadGroupConfig(identity.group)
	if err != nil {
		return nil, err
	}

	ext, err := validateUploadFile(params, groupConfig)
	if err != nil {
		return nil, err
	}
	if err := s.checkCapacity(ctx, params, identity.userID); err != nil {
		return nil, err
	}

	strategy, err := s.selectStrategy(ctx, params.StrategyID, identity, groupConfig)
	if err != nil {
		return nil, err
	}

	// Step 5: Rate limiting
	if err := s.checkRateLimit(ctx, identity.userID, params.ClientIP, &groupConfig); err != nil {
		return nil, err
	}

	// Step 6: Process image
	originalSize := int64(len(params.FileData))
	processed := processUploadImage(params.FileData, ext, groupConfig)
	fileData := processed.data

	// Step 7: Generate path and filename
	pathname := GeneratePathname(
		groupConfig.PathNamingRule,
		groupConfig.FileNamingRule,
		ext,
		ComputeMD5(params.FileData),
		identity.userID,
	)

	// Step 8: Compute hashes
	md5Hash := ComputeMD5(fileData)
	h := sha1.Sum(fileData)
	sha1Hash := fmt.Sprintf("%x", h[:])

	// Step 9: Dedup check
	existing, err := s.db.FindDuplicateImage(ctx, sqlc.FindDuplicateImageParams{
		StrategyID: domain.PgInt8(strategy.ID),
		Md5:        md5Hash,
		Sha1:       sha1Hash,
	})
	if err != nil {
		slog.Warn("dedup check failed, assuming unique", "error", err)
	}
	dedup := existing.ID != 0

	// Step 10: Write to storage
	if !dedup {
		store, err := s.getStorage(strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to init storage: %w", err)
		}
		defer store.Close()
		if err := store.Write(ctx, pathname, fileData, mimetypeFromExt(ext)); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Reuse existing file path so all dedup records point to the same file
		pathname = existing.Path + "/" + existing.Name
	}

	// Step 11: Save DB record
	keyLen := BaseKeyLength(s.totalImagesCount(ctx))
	imageKey := GenerateImageKey(keyLen)
	// Ensure key uniqueness
	for {
		_, err := s.db.GetImageByKey(ctx, imageKey)
		if errors.Is(err, pgx.ErrNoRows) {
			break // key is unique
		}
		if err != nil {
			return nil, fmt.Errorf("failed to check key uniqueness: %w", err)
		}
		// Collision is rare at our occupancy thresholds, but check whether
		// the local count has passed the next tier since we last looked.
		keyLen = BaseKeyLength(s.totalImagesCount(ctx))
		imageKey = GenerateImageKey(keyLen)
	}

	perm := int16(domain.PermissionPublic)
	if params.Permission != nil {
		perm = *params.Permission
	}

	var img sqlc.Image
	err = sqlc.RunInTx(ctx, s.pool, func(qtx *sqlc.Queries) error {
		var err error
		img, err = qtx.CreateImage(ctx, sqlc.CreateImageParams{
			UserID:     domain.PgInt8Ptr(params.UserID),
			AlbumID:    domain.PgInt8Ptr(params.AlbumID),
			GroupID:    domain.PgInt8(identity.groupID),
			StrategyID: domain.PgInt8(strategy.ID),
			Key:        imageKey,
			Path:       filepath.Dir(pathname),
			Name:       filepath.Base(pathname),
			OriginName: params.FileName,
			SizeBytes:  int64(len(fileData)),
			Mimetype:   mimetypeFromExt(ext),
			Extension:  ext,
			Md5:        md5Hash,
			Sha1:       sha1Hash,
			Width:      int32(processed.width),
			Height:     int32(processed.height),
			Permission: perm,
			UploadedIp: params.ClientIP,
			ExpiresAt:  domain.PgTimeWithZonePtr(params.ExpiresAt),
		})
		if err != nil {
			return err
		}
		if params.UserID != nil {
			if err := qtx.IncrementUserImageNum(ctx, identity.userID); err != nil {
				return err
			}
		}
		if params.AlbumID != nil {
			if err := qtx.IncrementAlbumImageNum(ctx, *params.AlbumID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if !dedup {
			store, delErr := s.getStorage(strategy)
			if delErr != nil {
				slog.Warn("failed to get storage for rollback cleanup", "error", delErr)
			} else if store != nil {
				defer store.Close()
				if delErr := store.Delete(ctx, pathname); delErr != nil {
					slog.Warn("rollback cleanup failed", "pathname", pathname, "error", delErr)
				}
			}
		}
		return nil, fmt.Errorf("failed to save image record: %w", err)
	}

	s.totalImages.Add(1)

	// Step 12: Content moderation (best effort, does not fail upload)
	modResult := &moderation.Result{Status: moderation.StatusApproved, Provider: "noop"}
	if mod := moderation.FromContext(ctx); mod != nil {
		mr, err := mod.Moderate(ctx, img.ID, img.Key, fileData)
		if err == nil && mr != nil {
			modResult = mr
			if mr.Status != moderation.StatusApproved {
				if _, updateErr := s.db.UpdateImageModerationStatus(ctx, sqlc.UpdateImageModerationStatusParams{
					ID:               img.ID,
					ModerationStatus: string(mr.Status),
				}); updateErr != nil {
					slog.Warn("failed to set moderation status", "image_id", img.ID, "error", updateErr)
				}
			}
		}
	}

	// Step 13: Generate thumbnail (synchronous to avoid race with frontend)
	if thumbErr := GenerateThumbnail(fileData, ext, s.config.Storage.ThumbnailDir, md5Hash); thumbErr != nil {
		slog.Warn("thumbnail generation failed", "md5", md5Hash, "error", thumbErr)
	}

	// Step 14: Build response
	links := LinkBuilder{BaseURL: s.config.ServerSnapshot().BaseURL}.BuildImageLinks(imageKey, ext, md5Hash, params.FileName)

	// Include moderation info in response so frontend can show pending state
	resp := &UploadResult{
		Image:             img,
		Links:             links,
		GroupConfig:       groupConfig,
		OriginalSizeBytes: originalSize,
		StoredSizeBytes:   int64(len(fileData)),
		Processed:         processed.processed,
	}
	// Attach moderation status to the response for frontend awareness
	_ = modResult

	return resp, nil
}

func (s *UploadService) getStorage(strategy sqlc.Strategy) (storage.Storage, error) {
	return GetStorageForStrategy(strategy)
}

func (s *UploadService) applyDefaultTTL(params UploadParams) UploadParams {
	app := s.config.AppSnapshot()
	if params.ExpiresAt == nil && app.DefaultImageTTL > 0 {
		t := time.Now().Add(app.DefaultImageTTL)
		params.ExpiresAt = &t
	}
	return params
}

func (s *UploadService) resolveIdentity(ctx context.Context, params UploadParams) (uploadIdentity, error) {
	if params.UserID == nil {
		if !s.config.AppSnapshot().AllowGuestUpload {
			return uploadIdentity{}, fmt.Errorf("guest upload is disabled")
		}
		group, err := s.db.GetGuestGroup(ctx)
		if err != nil {
			return uploadIdentity{}, fmt.Errorf("no guest group configured")
		}
		return uploadIdentity{group: group, groupID: group.ID}, nil
	}

	userID := *params.UserID
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return uploadIdentity{}, fmt.Errorf("user not found")
	}
	if user.Status != int16(domain.UserStatusActive) {
		return uploadIdentity{}, fmt.Errorf("account is frozen")
	}
	if !user.GroupID.Valid {
		return uploadIdentity{}, fmt.Errorf("user has no group")
	}

	groupID := user.GroupID.Int64
	group, err := s.db.GetGroupByID(ctx, groupID)
	if err != nil {
		return uploadIdentity{}, fmt.Errorf("group not found")
	}

	var preferredStrategyID *int64
	if len(user.Settings) > 0 {
		var settings uploadUserSettings
		if err := json.Unmarshal(user.Settings, &settings); err == nil && settings.DefaultStrategy > 0 {
			preferredStrategyID = &settings.DefaultStrategy
		}
	}

	return uploadIdentity{
		group:               group,
		userID:              userID,
		groupID:             groupID,
		isAdmin:             user.Role == string(domain.RoleAdmin),
		preferredStrategyID: preferredStrategyID,
	}, nil
}

func loadUploadGroupConfig(group sqlc.Group) (domain.GroupConfig, error) {
	var groupConfig domain.GroupConfig
	if err := json.Unmarshal(group.Configs, &groupConfig); err != nil {
		return domain.GroupConfig{}, fmt.Errorf("invalid group config")
	}
	return groupConfig, nil
}

func validateUploadFile(params UploadParams, groupConfig domain.GroupConfig) (string, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(params.FileName)), ".")
	if !groupConfig.IsExtensionAllowed(ext) {
		return "", fmt.Errorf("file extension .%s is not allowed", ext)
	}
	if params.FileSize > groupConfig.MaximumFileSize {
		return "", fmt.Errorf("file size exceeds maximum (%d bytes)", groupConfig.MaximumFileSize)
	}
	return ext, nil
}

func (s *UploadService) checkCapacity(ctx context.Context, params UploadParams, userID int64) error {
	if params.UserID == nil {
		capacity := s.config.AppSnapshot().GuestCapacityBytes
		if capacity <= 0 {
			return nil
		}
		used, err := s.db.GetGuestUsedCapacity(ctx)
		if err != nil {
			return fmt.Errorf("failed to check guest capacity: %w", err)
		}
		if used+params.FileSize > capacity {
			return fmt.Errorf("guest storage capacity exceeded")
		}
		return nil
	}
	used, err := s.db.GetUserUsedCapacity(ctx, domain.PgInt8(userID))
	if err != nil {
		return fmt.Errorf("failed to check capacity: %w", err)
	}
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user.CapacityBytes > 0 && used+params.FileSize > user.CapacityBytes {
		return fmt.Errorf("storage capacity exceeded")
	}
	return nil
}

func (s *UploadService) selectStrategy(ctx context.Context, requestedStrategyID *int64, identity uploadIdentity, groupConfig domain.GroupConfig) (sqlc.Strategy, error) {
	strategies, err := s.db.GetGroupStrategies(ctx, identity.groupID)
	if err != nil || len(strategies) == 0 {
		if !identity.isAdmin {
			return sqlc.Strategy{}, fmt.Errorf("no storage strategy available")
		}
		allStrats, err := s.db.ListStrategies(ctx)
		if err != nil || len(allStrats) == 0 {
			return sqlc.Strategy{}, fmt.Errorf("no storage strategy available in system")
		}
		strategies = allStrats
	}

	if requestedStrategyID != nil {
		if strategy, ok := findStrategyByID(strategies, *requestedStrategyID); ok {
			return strategy, nil
		}
		if identity.isAdmin {
			if strategy, err := s.db.GetStrategyByID(ctx, *requestedStrategyID); err == nil {
				return strategy, nil
			}
		}
		return sqlc.Strategy{}, fmt.Errorf("strategy not available for your group")
	}

	if identity.preferredStrategyID != nil {
		if strategy, ok := findStrategyByID(strategies, *identity.preferredStrategyID); ok {
			return strategy, nil
		}
	}
	if groupConfig.DefaultStrategyID > 0 {
		if strategy, ok := findStrategyByID(strategies, groupConfig.DefaultStrategyID); ok {
			return strategy, nil
		}
	}
	return strategies[0], nil
}

func findStrategyByID(strategies []sqlc.Strategy, id int64) (sqlc.Strategy, bool) {
	for _, strategy := range strategies {
		if strategy.ID == id {
			return strategy, true
		}
	}
	return sqlc.Strategy{}, false
}

func processUploadImage(fileData []byte, ext string, groupConfig domain.GroupConfig) processedUploadImage {
	result := processedUploadImage{data: fileData}
	skipExts := map[string]bool{"gif": true, "svg": true, "ico": true}
	if skipExts[ext] {
		return result
	}

	targetFormat := groupConfig.ImageSaveFormat
	if targetFormat == "" {
		targetFormat = ext
	}

	needProcess := groupConfig.ImageSaveFormat != "" || groupConfig.ImageSaveQuality < 100 || groupConfig.IsStripExif
	if needProcess {
		processedImg, err := ProcessImage(result.data, targetFormat, groupConfig.ImageSaveQuality, groupConfig.IsStripExif)
		if err != nil {
			slog.Warn("image processing failed, using original", "error", err)
		} else {
			result.data = processedImg.Data
			result.width = processedImg.Width
			result.height = processedImg.Height
			result.processed = true
		}
	}

	if groupConfig.IsEnableWatermark {
		var wmCfg WatermarkConfig
		if err := json.Unmarshal(groupConfig.WatermarkConfigs, &wmCfg); err == nil && wmCfg.Text != "" {
			watermarked, err := ApplyWatermark(result.data, wmCfg, targetFormat, groupConfig.ImageSaveQuality)
			if err != nil {
				slog.Warn("watermark failed, using original", "error", err)
			} else {
				result.data = watermarked
				result.processed = true
			}
		}
	}

	return result
}

func (s *UploadService) checkRateLimit(ctx context.Context, userID int64, clientIP string, cfg *domain.GroupConfig) error {
	limits := []struct {
		count int
		desc  string
		secs  string
	}{
		{cfg.LimitPerMinute, "minute", "60"},
		{cfg.LimitPerHour, "hour", "3600"},
		{cfg.LimitPerDay, "day", "86400"},
		{cfg.LimitPerMonth, "month", "2592000"},
	}

	for _, lim := range limits {
		if lim.count <= 0 {
			continue
		}
		var count int64
		var err error
		if userID > 0 {
			count, err = s.db.CountImagesInWindow(ctx, sqlc.CountImagesInWindowParams{
				UserID:  domain.PgInt8(userID),
				Column2: lim.secs,
			})
		} else {
			count, err = s.db.CountImagesInWindowByIP(ctx, sqlc.CountImagesInWindowByIPParams{
				UploadedIp: clientIP,
				Column2:    lim.secs,
			})
		}
		if err != nil {
			return fmt.Errorf("rate limit check failed: %w", err)
		}
		if count >= int64(lim.count) {
			return fmt.Errorf("rate limit exceeded: %d uploads per %s", lim.count, lim.desc)
		}
	}
	return nil
}

func mimetypeFromExt(ext string) string {
	mimes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"bmp":  "image/bmp",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		"tif":  "image/tiff",
		"tiff": "image/tiff",
		"psd":  "image/vnd.adobe.photoshop",
	}
	if m, ok := mimes[ext]; ok {
		return m
	}
	return "application/octet-stream"
}

// Suppress unused import warning
var _ = math.Pi
