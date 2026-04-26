package handler

import (
	"encoding/json"
	"time"

	"github.com/atbeta/picfast/internal/domain"
)

// Auth

type AuthResponse = domain.AuthTokens

// Albums

type AlbumResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Intro     string    `json:"intro"`
	ImageNum  int64     `json:"image_num"`
	CreatedAt time.Time `json:"created_at"`
}

type AlbumUpdateResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Intro     string    `json:"intro"`
	ImageNum  int64     `json:"image_num"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Images

type ImageResponse struct {
	ID               int64             `json:"id"`
	Key              string            `json:"key"`
	OriginName       string            `json:"origin_name"`
	SizeBytes        int64             `json:"size_bytes"`
	Mimetype         string            `json:"mimetype"`
	Extension        string            `json:"extension"`
	Width            int32             `json:"width"`
	Height           int32             `json:"height"`
	Md5              string            `json:"md5"`
	Sha1             string            `json:"sha1"`
	Permission       int16             `json:"permission"`
	AlbumID          *int64            `json:"album_id"`
	ModerationStatus string            `json:"moderation_status"`
	Links            domain.ImageLinks `json:"links"`
	CreatedAt        time.Time         `json:"created_at"`
}

type ImageListItem struct {
	ID               int64             `json:"id"`
	Key              string            `json:"key"`
	OriginName       string            `json:"origin_name"`
	SizeBytes        int64             `json:"size_bytes"`
	Mimetype         string            `json:"mimetype"`
	Extension        string            `json:"extension"`
	Width            int32             `json:"width"`
	Height           int32             `json:"height"`
	Permission       int16             `json:"permission"`
	AlbumID          *int64            `json:"album_id"`
	URL              string            `json:"url"`
	ThumbnailURL     string            `json:"thumbnail_url"`
	ModerationStatus string            `json:"moderation_status"`
	StrategyID       *int64            `json:"strategy_id"`
	StrategyName     string            `json:"strategy_name"`
	StrategyType     string            `json:"strategy_type"`
	Links            domain.ImageLinks `json:"links"`
	CreatedAt        time.Time         `json:"created_at"`
}

// Users

type UserProfileResponse struct {
	ID            int64           `json:"id"`
	Email         string          `json:"email"`
	Name          string          `json:"name"`
	Role          string          `json:"role"`
	Status        int16           `json:"status"`
	CapacityBytes int64           `json:"capacity_bytes"`
	UsedBytes     int64           `json:"used_bytes"`
	ImageNum      int64           `json:"image_num"`
	AlbumNum      int64           `json:"album_num"`
	Settings      json.RawMessage `json:"settings"`
	EmailVerified bool            `json:"email_verified"`
	CreatedAt     time.Time       `json:"created_at"`
}

type UserProfileUpdateResponse struct {
	ID        int64           `json:"id"`
	Email     string          `json:"email"`
	Name      string          `json:"name"`
	Settings  json.RawMessage `json:"settings"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Admin - Users

type AdminUserResponse struct {
	ID            int64           `json:"id"`
	Email         string          `json:"email"`
	Name          string          `json:"name"`
	Role          string          `json:"role"`
	GroupID       *int64          `json:"group_id"`
	CapacityBytes int64           `json:"capacity_bytes"`
	ImageNum      int64           `json:"image_num"`
	AlbumNum      int64           `json:"album_num"`
	Status        int16           `json:"status"`
	EmailVerified bool            `json:"email_verified"`
	Settings      json.RawMessage `json:"settings"`
	CreatedAt     time.Time       `json:"created_at"`
}

// Admin - Groups

type AdminGroupResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	IsDefault bool            `json:"is_default"`
	IsGuest   bool            `json:"is_guest"`
	Configs   json.RawMessage `json:"configs"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type AdminGroupDetailResponse struct {
	AdminGroupResponse
	StrategyIDs []int64 `json:"strategy_ids"`
}

// Admin - Strategies

type AdminStrategyResponse struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	StrategyType string          `json:"strategy_type"`
	Configs      json.RawMessage `json:"configs"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Admin - Settings

type SettingsResponse struct {
	AppName              string `json:"app_name"`
	AllowGuestUpload     bool   `json:"allow_guest_upload"`
	AllowRegistration    bool   `json:"allow_registration"`
	UserInitialCapacity  int64  `json:"user_initial_capacity"`
}
