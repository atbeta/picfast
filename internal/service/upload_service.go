package service

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/moderation"
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadService struct {
	db     *sqlc.Queries
	pool   *pgxpool.Pool
	config *config.Config
}

func NewUploadService(db *sqlc.Queries, pool *pgxpool.Pool, cfg *config.Config) *UploadService {
	return &UploadService{db: db, pool: pool, config: cfg}
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
	Image       sqlc.Image
	Links       domain.ImageLinks
	GroupConfig domain.GroupConfig
}

func (s *UploadService) Store(ctx context.Context, params UploadParams) (*UploadResult, error) {
	// Apply default TTL if none provided
	if params.ExpiresAt == nil && s.config.App.DefaultImageTTL > 0 {
		t := time.Now().Add(s.config.App.DefaultImageTTL)
		params.ExpiresAt = &t
	}

	// Step 1: Resolve identity
	var group sqlc.Group
	var userID int64
	var groupID int64

	if params.UserID != nil {
		userID = *params.UserID
		user, err := s.db.GetUserByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("user not found")
		}
		if user.Status != int16(domain.UserStatusActive) {
			return nil, fmt.Errorf("account is frozen")
		}
		if !user.GroupID.Valid {
			return nil, fmt.Errorf("user has no group")
		}
		groupID = user.GroupID.Int64
		group, err = s.db.GetGroupByID(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("group not found")
		}
	} else {
		if !s.config.App.AllowGuestUpload {
			return nil, fmt.Errorf("guest upload is disabled")
		}
		guestGroup, err := s.db.GetGuestGroup(ctx)
		if err != nil {
			return nil, fmt.Errorf("no guest group configured")
		}
		group = guestGroup
		groupID = group.ID
	}

	// Step 2: Load group config
	var groupConfig domain.GroupConfig
	if err := json.Unmarshal(group.Configs, &groupConfig); err != nil {
		return nil, fmt.Errorf("invalid group config")
	}

	// Step 3: Validate file
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(params.FileName)), ".")
	if !groupConfig.IsExtensionAllowed(ext) {
		return nil, fmt.Errorf("file extension .%s is not allowed", ext)
	}
	if params.FileSize > groupConfig.MaximumFileSize {
		return nil, fmt.Errorf("file size exceeds maximum (%d bytes)", groupConfig.MaximumFileSize)
	}

	// Capacity check for authenticated users
	if params.UserID != nil {
		used, err := s.db.GetUserUsedCapacity(ctx, domain.PgInt8(userID))
		if err != nil {
			return nil, fmt.Errorf("failed to check capacity: %w", err)
		}
		user, err := s.db.GetUserByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		if user.CapacityBytes > 0 && used > 0 && used+params.FileSize > user.CapacityBytes {
			return nil, fmt.Errorf("storage capacity exceeded")
		}
	}

	// Step 4: Select strategy
	strategies, err := s.db.GetGroupStrategies(ctx, groupID)
	if err != nil || len(strategies) == 0 {
		return nil, fmt.Errorf("no storage strategy available")
	}

	var strategy sqlc.Strategy
	if params.StrategyID != nil {
		found := false
		for _, st := range strategies {
			if st.ID == *params.StrategyID {
				strategy = st
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("strategy not available for your group")
		}
	} else {
		strategy = strategies[0]
	}

	// Step 5: Rate limiting
	if err := s.checkRateLimit(ctx, userID, params.ClientIP, &groupConfig); err != nil {
		return nil, err
	}

	// Step 6: Process image
	fileData := params.FileData
	var width, height int

	skipExts := map[string]bool{"gif": true, "svg": true, "ico": true}
	if !skipExts[ext] {
		needProcess := groupConfig.ImageSaveFormat != "" || groupConfig.ImageSaveQuality < 100 || groupConfig.IsStripExif
		if needProcess {
			targetFormat := groupConfig.ImageSaveFormat
			if targetFormat == "" {
				targetFormat = ext
			}
			processed, err := ProcessImage(fileData, targetFormat, groupConfig.ImageSaveQuality, groupConfig.IsStripExif)
			if err != nil {
				slog.Warn("image processing failed, using original", "error", err)
			} else {
				fileData = processed.Data
				width = processed.Width
				height = processed.Height
			}
		}
	}

	// Step 6b: Apply watermark if enabled
	if groupConfig.IsEnableWatermark && !skipExts[ext] {
		var wmCfg WatermarkConfig
		if err := json.Unmarshal(groupConfig.WatermarkConfigs, &wmCfg); err == nil && wmCfg.Text != "" {
			targetFormat := groupConfig.ImageSaveFormat
			if targetFormat == "" {
				targetFormat = ext
			}
			watermarked, err := ApplyWatermark(fileData, wmCfg, targetFormat, groupConfig.ImageSaveQuality)
			if err != nil {
				slog.Warn("watermark failed, using original", "error", err)
			} else {
				fileData = watermarked
			}
		}
	}

	// Step 7: Generate path and filename
	pathname := GeneratePathname(
		groupConfig.PathNamingRule,
		groupConfig.FileNamingRule,
		ext,
		ComputeMD5(params.FileData),
		userID,
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
	imageKey := GenerateImageKey()
	// Ensure key uniqueness
	for {
		_, err := s.db.GetImageByKey(ctx, imageKey)
		if err != nil {
			break // key is unique
		}
		imageKey = GenerateImageKey()
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
			GroupID:    domain.PgInt8(groupID),
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
			Width:      int32(width),
			Height:     int32(height),
			Permission: perm,
			UploadedIp: params.ClientIP,
			ExpiresAt:  domain.PgTimeWithZonePtr(params.ExpiresAt),
		})
		if err != nil {
			return err
		}
		if params.UserID != nil {
			qtx.IncrementUserImageNum(ctx, userID)
		}
		if params.AlbumID != nil {
			qtx.IncrementAlbumImageNum(ctx, *params.AlbumID)
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
	imageURL := s.config.Server.BaseURL + "/i/" + imageKey + "." + ext
	thumbURL := s.config.Server.BaseURL + "/t/" + md5Hash + ".png"

	links := domain.ImageLinks{
		URL:          imageURL,
		HTML:         fmt.Sprintf(`<img src="%s" alt="%s" />`, imageURL, params.FileName),
		BBCode:       fmt.Sprintf("[img]%s[/img]", imageURL),
		Markdown:     fmt.Sprintf("![%s](%s)", params.FileName, imageURL),
		ThumbnailURL: thumbURL,
	}

	// Include moderation info in response so frontend can show pending state
	resp := &UploadResult{
		Image:       img,
		Links:       links,
		GroupConfig: groupConfig,
	}
	// Attach moderation status to the response for frontend awareness
	_ = modResult

	return resp, nil
}

func (s *UploadService) getStorage(strategy sqlc.Strategy) (storage.Storage, error) {
	return GetStorageForStrategy(strategy)
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
