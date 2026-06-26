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
	StrategyTypeLocal  StrategyType = "local"
	StrategyTypeS3     StrategyType = "s3"
	StrategyTypeKodo   StrategyType = "kodo"
	StrategyTypeOSS    StrategyType = "oss"
	StrategyTypeCOS    StrategyType = "cos"
	StrategyTypeWebDAV StrategyType = "webdav"
)

type GroupConfig struct {
	MaximumFileSize            int64           `json:"maximum_file_size"`
	AcceptedExtensions         []string        `json:"accepted_extensions"`
	DefaultStrategyID          int64           `json:"default_strategy_id"`
	LimitPerMinute             int             `json:"limit_per_minute"`
	LimitPerHour               int             `json:"limit_per_hour"`
	LimitPerDay                int             `json:"limit_per_day"`
	LimitPerMonth              int             `json:"limit_per_month"`
	UserCapacityBytes          int64           `json:"user_capacity_bytes"`
	PathNamingRule             string          `json:"path_naming_rule"`
	FileNamingRule             string          `json:"file_naming_rule"`
	ImageSaveQuality           int             `json:"image_save_quality"`
	ImageSaveFormat            string          `json:"image_save_format"`
	IsEnableWatermark          bool            `json:"is_enable_watermark"`
	WatermarkConfigs           json.RawMessage `json:"watermark_configs"`
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

type ThemeOverride struct {
	Preset  string `json:"preset,omitempty"`
	Mode    string `json:"mode,omitempty"`    // light, dark, system
	Density string `json:"density,omitempty"` // compact, comfortable, spacious
	Motion  string `json:"motion,omitempty"`  // none, subtle, playful
}

type UserSettings struct {
	DefaultAlbum      int64                        `json:"default_album,omitempty"`
	DefaultStrategy   int64                        `json:"default_strategy,omitempty"`
	DefaultPermission *int16                       `json:"default_permission,omitempty"`
	ImageProcessing   *UserImageProcessingSettings `json:"image_processing,omitempty"`
	Language          *string                      `json:"language,omitempty"`
}

type UserImageProcessingSettings struct {
	ImageSaveQuality  *int             `json:"image_save_quality,omitempty"`
	ImageSaveFormat   *string          `json:"image_save_format,omitempty"`
	IsStripExif       *bool            `json:"is_strip_exif,omitempty"`
	IsEnableWatermark *bool            `json:"is_enable_watermark,omitempty"`
	WatermarkConfigs  *WatermarkConfig `json:"watermark_configs,omitempty"`
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

type KodoStrategyConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	Domain    string `json:"domain"`
	Zone      string `json:"zone"`
	Private   bool   `json:"private"`
}

type OSSStrategyConfig struct {
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	URL       string `json:"url"`
}

type COSStrategyConfig struct {
	BucketURL string `json:"bucket_url"`
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	URL       string `json:"url"`
}

type WebDAVStrategyConfig struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

type WatermarkConfig struct {
	Text     string  `json:"text"`
	Position string  `json:"position"` // bottom-right, bottom-left, top-right, top-left, center
	FontSize float64 `json:"font_size"`
	Color    string  `json:"color"`   // #RRGGBB or #RRGGBBAA
	Opacity  float64 `json:"opacity"` // 0.0 - 1.0
}

type ImageLinks struct {
	URL          string `json:"url"`
	HTML         string `json:"html"`
	BBCode       string `json:"bbcode"`
	Markdown     string `json:"markdown"`
	ThumbnailURL string `json:"thumbnail_url"`
}
