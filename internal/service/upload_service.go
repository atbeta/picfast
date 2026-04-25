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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pbeta/imgapi/internal/config"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/service/storage"
	"github.com/pbeta/imgapi/internal/sqlc"
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
	FileData     []byte
	FileName     string
	FileSize     int64
	StrategyID   *int64
	AlbumID      *int64
	Permission   *int16
	UserID       *int64
	ClientIP     string
}

type UploadResult struct {
	Image       sqlc.Image
	Links       domain.ImageLinks
	GroupConfig domain.GroupConfig
}

func (s *UploadService) Store(ctx context.Context, params UploadParams) (*UploadResult, error) {
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
		used, _ := s.db.GetUserUsedCapacity(ctx, domain.PgInt8(userID))
		user, _ := s.db.GetUserByID(ctx, userID)
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
		if groupConfig.ImageSaveFormat != "" || groupConfig.ImageSaveQuality < 100 {
			targetFormat := groupConfig.ImageSaveFormat
			if targetFormat == "" {
				targetFormat = ext
			}
			processed, err := ProcessImage(fileData, targetFormat, groupConfig.ImageSaveQuality)
			if err != nil {
				slog.Warn("image processing failed, using original", "error", err)
			} else {
				fileData = processed.Data
				width = processed.Width
				height = processed.Height
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
	existing, _ := s.db.FindDuplicateImage(ctx, sqlc.FindDuplicateImageParams{
		StrategyID: domain.PgInt8(strategy.ID),
		Md5:        md5Hash,
		Sha1:       sha1Hash,
	})
	dedup := existing.ID != 0

	// Step 10: Write to storage
	if !dedup {
		store, err := s.getStorage(strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to init storage: %w", err)
		}
		if err := store.Write(ctx, pathname, fileData); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
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

	img, err := s.db.CreateImage(ctx, sqlc.CreateImageParams{
		UserID:      domain.PgInt8Ptr(params.UserID),
		AlbumID:     domain.PgInt8Ptr(params.AlbumID),
		GroupID:     domain.PgInt8(groupID),
		StrategyID:  domain.PgInt8(strategy.ID),
		Key:         imageKey,
		Path:        filepath.Dir(pathname),
		Name:        filepath.Base(pathname),
		OriginName:  params.FileName,
		SizeBytes:   int64(len(fileData)),
		Mimetype:    mimetypeFromExt(ext),
		Extension:   ext,
		Md5:         md5Hash,
		Sha1:        sha1Hash,
		Width:       int32(width),
		Height:      int32(height),
		Permission:  perm,
		UploadedIp:  params.ClientIP,
	})
	if err != nil {
		// Cleanup file on DB error (only if we wrote it)
		if !dedup {
			store, _ := s.getStorage(strategy)
			if store != nil {
				store.Delete(ctx, pathname)
			}
		}
		return nil, fmt.Errorf("failed to save image record: %w", err)
	}

	// Update counters
	if params.UserID != nil {
		s.db.IncrementUserImageNum(ctx, userID)
	}
	if params.AlbumID != nil {
		s.db.IncrementAlbumImageNum(ctx, *params.AlbumID)
	}

	// Step 12: Generate thumbnail (best effort)
	go func() {
		GenerateThumbnail(fileData, ext, s.config.Storage.ThumbnailDir, md5Hash)
	}()

	// Step 13: Build response
	imageURL := s.config.Server.BaseURL + "/i/" + imageKey + "." + ext
	thumbURL := s.config.Server.BaseURL + "/t/" + md5Hash + ".png"

	links := domain.ImageLinks{
		URL:          imageURL,
		HTML:         fmt.Sprintf(`<img src="%s" alt="%s" />`, imageURL, params.FileName),
		BBCode:       fmt.Sprintf("[img]%s[/img]", imageURL),
		Markdown:     fmt.Sprintf("![%s](%s)", params.FileName, imageURL),
		ThumbnailURL: thumbURL,
	}

	return &UploadResult{
		Image:       img,
		Links:       links,
		GroupConfig: groupConfig,
	}, nil
}

func (s *UploadService) getStorage(strategy sqlc.Strategy) (storage.Storage, error) {
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
		if userID > 0 {
			c, _ := s.db.CountImagesInWindow(ctx, sqlc.CountImagesInWindowParams{
				UserID: domain.PgInt8(userID),
				Column2: lim.secs,
			})
			count = c
		} else {
			c, _ := s.db.CountImagesInWindowByIP(ctx, sqlc.CountImagesInWindowByIPParams{
				UploadedIp: clientIP,
				Column2:    lim.secs,
			})
			count = c
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
