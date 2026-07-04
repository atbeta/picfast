package events

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/google/uuid"
)

type ImageUploadedData struct {
	Key          string    `json:"key"`
	OriginName   string    `json:"origin_name"`
	SizeBytes    int64     `json:"size_bytes"`
	Mimetype     string    `json:"mimetype"`
	Extension    string    `json:"extension"`
	Width        int32     `json:"width"`
	Height       int32     `json:"height"`
	Md5          string    `json:"md5"`
	Sha1         string    `json:"sha1"`
	Permission   int16     `json:"permission"`
	AlbumID      *int64    `json:"album_id"`
	StrategyID   int64     `json:"strategy_id"`
	UserID       *int64    `json:"user_id"`
	UploadedIP   string    `json:"uploaded_ip"`
	ExpiresAt    *time.Time `json:"expires_at"`
	Links        ImageLinks `json:"links"`
	Exif         json.RawMessage `json:"exif,omitempty"`
	Phash        uint64          `json:"phash,omitempty"`
}

type ImageLinks struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type ImageProcessedData struct {
	Key              string `json:"key"`
	Processed        bool   `json:"processed"`
	OriginalSizeBytes int64  `json:"original_size_bytes"`
	StoredSizeBytes   int64  `json:"stored_size_bytes"`
	Thumbnail        string `json:"thumbnail"`
	ModerationStatus string `json:"moderation_status"`
	ModerationProvider string `json:"moderation_provider"`
	UserID            *int64 `json:"user_id"`
}

type ImageDeletedData struct {
	Key        string `json:"key"`
	OriginName string `json:"origin_name"`
	UserID     *int64 `json:"user_id"`
	DeletedBy  string `json:"deleted_by"`
	Reason     string `json:"reason"`
}

type ModerationReviewedData struct {
	ImageID    int64  `json:"image_id"`
	Key        string `json:"key"`
	Status     string `json:"status"`
	ReviewerID int64  `json:"reviewer_id"`
	Reason     string `json:"reason"`
	Provider   string `json:"provider"`
	OwnerID    *int64 `json:"owner_id"`
}

type UserRegisteredData struct {
	UserID                   int64  `json:"user_id"`
	Email                    string `json:"email"`
	Name                     string `json:"name"`
	GroupID                  int64  `json:"group_id"`
	EmailVerificationRequired bool   `json:"email_verification_required"`
}

type WebhookTestData struct {
	Message string `json:"message"`
}

func BuildImageUploaded(img sqlc.Image, links ImageLinks) Envelope {
	now := time.Now().UTC()
	var expiresAt *time.Time
	if img.ExpiresAt.Valid {
		expiresAt = &img.ExpiresAt.Time
	}
	return Envelope{
		ID:             uuid.NewString(),
		Type:           TypeImageUploaded,
		Version:        Version20260601,
		OccurredAt:     now,
		IdempotencyKey: ImageUploadedKey(img.ID),
		Subject: &Subject{
			Kind: "image",
			ID:   formatInt64(img.ID),
			Key:  img.Key,
		},
		Data: ImageUploadedData{
			Key:        img.Key,
			OriginName: img.OriginName,
			SizeBytes:  img.SizeBytes,
			Mimetype:   img.Mimetype,
			Extension:  img.Extension,
			Width:      img.Width,
			Height:     img.Height,
			Md5:        img.Md5,
			Sha1:       img.Sha1,
			Permission: img.Permission,
			AlbumID:    domain.PgInt8PtrVal(img.AlbumID),
			StrategyID: domain.PgInt8Val(img.StrategyID),
			UserID:     domain.PgInt8PtrVal(img.UserID),
			UploadedIP: img.UploadedIp,
			ExpiresAt:  expiresAt,
			Links:      links,
		},
	}
}

func BuildImageUploadedWithActor(img sqlc.Image, links ImageLinks, actor *Actor) Envelope {
	e := BuildImageUploaded(img, links)
	e.Actor = actor
	return e
}

func BuildImageProcessed(img sqlc.Image, originalSize, storedSize int64, processed bool, thumbnailStatus, moderationStatus, moderationProvider string) Envelope {
	now := time.Now().UTC()
	return Envelope{
		ID:             uuid.NewString(),
		Type:           TypeImageProcessed,
		Version:        Version20260601,
		OccurredAt:     now,
		IdempotencyKey: ImageProcessedKey(img.ID),
		Subject: &Subject{
			Kind: "image",
			ID:   formatInt64(img.ID),
			Key:  img.Key,
		},
		Data: ImageProcessedData{
			Key:               img.Key,
			Processed:         processed,
			OriginalSizeBytes: originalSize,
			StoredSizeBytes:   storedSize,
			Thumbnail:         thumbnailStatus,
			ModerationStatus:  moderationStatus,
			ModerationProvider: moderationProvider,
			UserID:            domain.PgInt8PtrVal(img.UserID),
		},
	}
}

func BuildImageDeleted(img sqlc.Image, deletedBy string) Envelope {
	now := time.Now().UTC()
	return Envelope{
		ID:             uuid.NewString(),
		Type:           TypeImageDeleted,
		Version:        Version20260601,
		OccurredAt:     now,
		IdempotencyKey: ImageDeletedKey(img.ID),
		Subject: &Subject{
			Kind: "image",
			ID:   formatInt64(img.ID),
			Key:  img.Key,
		},
		Data: ImageDeletedData{
			Key:        img.Key,
			OriginName: img.OriginName,
			UserID:     domain.PgInt8PtrVal(img.UserID),
			DeletedBy:  deletedBy,
			Reason:     "",
		},
	}
}

func BuildModerationReviewed(imageID int64, key, status string, reviewerID int64, reason string, ownerID *int64) Envelope {
	now := time.Now().UTC()
	return Envelope{
		ID:             uuid.NewString(),
		Type:           TypeModerationReviewed,
		Version:        Version20260601,
		OccurredAt:     now,
		IdempotencyKey: ModerationReviewedKey(imageID, status),
		Subject: &Subject{
			Kind: "image",
			ID:   formatInt64(imageID),
			Key:  key,
		},
		Data: ModerationReviewedData{
			ImageID:    imageID,
			Key:        key,
			Status:     status,
			ReviewerID: reviewerID,
			Reason:     reason,
			Provider:   "manual",
			OwnerID:    ownerID,
		},
	}
}

func BuildUserRegistered(userID int64, email, name string, groupID int64, emailVerificationRequired bool) Envelope {
	now := time.Now().UTC()
	return Envelope{
		ID:             uuid.NewString(),
		Type:           TypeUserRegistered,
		Version:        Version20260601,
		OccurredAt:     now,
		IdempotencyKey: UserRegisteredKey(userID),
		Subject: &Subject{
			Kind: "user",
			ID:   formatInt64(userID),
		},
		Data: UserRegisteredData{
			UserID:                   userID,
			Email:                    email,
			Name:                     name,
			GroupID:                  groupID,
			EmailVerificationRequired: emailVerificationRequired,
		},
	}
}

func UserActor(email, name string, userID int64) *Actor {
	return &Actor{
		Kind:  "user",
		ID:    formatInt64(userID),
		Email: email,
		Name:  name,
	}
}

func AdminActor(userID int64) *Actor {
	return &Actor{
		Kind: "admin",
		ID:   formatInt64(userID),
	}
}

func SystemActor() *Actor {
	return &Actor{
		Kind: "system",
		ID:   "0",
	}
}

func GuestActor() *Actor {
	return &Actor{
		Kind: "guest",
		ID:   "",
	}
}

func BuildWebhookTest() Envelope {
	now := time.Now().UTC()
	return Envelope{
		ID:             uuid.NewString(),
		Type:           TypeWebhookTest,
		Version:        Version20260601,
		OccurredAt:     now,
		IdempotencyKey: "webhook.test:" + uuid.NewString(),
		Data: WebhookTestData{
			Message: "This is a test event from PicFast webhook integration.",
		},
	}
}

func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}
