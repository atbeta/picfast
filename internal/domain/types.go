package domain

import "encoding/json"

type UserStatus int16

const (
	UserStatusFrozen UserStatus = 0
	UserStatusActive UserStatus = 1
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type Permission int16

const (
	PermissionPrivate Permission = 0
	PermissionPublic  Permission = 1
)

type StrategyType string

const (
	StrategyTypeLocal StrategyType = "local"
	StrategyTypeS3    StrategyType = "s3"
)

type GroupConfig struct {
	MaximumFileSize          int64           `json:"maximum_file_size"`
	AcceptedExtensions       []string        `json:"accepted_extensions"`
	LimitPerMinute           int             `json:"limit_per_minute"`
	LimitPerHour             int             `json:"limit_per_hour"`
	LimitPerDay              int             `json:"limit_per_day"`
	LimitPerMonth            int             `json:"limit_per_month"`
	PathNamingRule           string          `json:"path_naming_rule"`
	FileNamingRule            string          `json:"file_naming_rule"`
	ImageSaveQuality         int             `json:"image_save_quality"`
	ImageSaveFormat          string          `json:"image_save_format"`
	IsEnableWatermark        bool            `json:"is_enable_watermark"`
	WatermarkConfigs         json.RawMessage `json:"watermark_configs"`
	IsEnableOriginalProtection bool           `json:"is_enable_original_protection"`
	IsStripExif                bool            `json:"is_strip_exif"`
}

func (c *GroupConfig) IsExtensionAllowed(ext string) bool {
	for _, allowed := range c.AcceptedExtensions {
		if allowed == ext {
			return true
		}
	}
	return false
}

type UserSettings struct {
	DefaultAlbum      int64 `json:"default_album"`
	DefaultStrategy   int64 `json:"default_strategy"`
	DefaultPermission int16 `json:"default_permission"`
}

type LocalStrategyConfig struct {
	Root string `json:"root"`
	URL  string `json:"url"`
}

type S3StrategyConfig struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key"`
	SecretAccessKey string `json:"secret_key"`
	URL             string `json:"url"`
}

type WatermarkConfig struct {
	Text     string  `json:"text"`
	Position string  `json:"position"` // bottom-right, bottom-left, top-right, top-left, center
	FontSize float64 `json:"font_size"`
	Color    string  `json:"color"`    // #RRGGBB or #RRGGBBAA
	Opacity  float64 `json:"opacity"`  // 0.0 - 1.0
}

type ImageLinks struct {
	URL          string `json:"url"`
	HTML         string `json:"html"`
	BBCode       string `json:"bbcode"`
	Markdown     string `json:"markdown"`
	ThumbnailURL string `json:"thumbnail_url"`
}
