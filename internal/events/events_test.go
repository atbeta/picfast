package events

import (
	"context"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func testImage() sqlc.Image {
	return sqlc.Image{
		ID:         42,
		UserID:     pgtype.Int8{Int64: 7, Valid: true},
		AlbumID:    pgtype.Int8{Int64: 3, Valid: true},
		StrategyID: pgtype.Int8{Int64: 1, Valid: true},
		Key:        "aBcDeFgH",
		OriginName: "photo.jpg",
		SizeBytes:  102400,
		Mimetype:   "image/jpeg",
		Extension:  "jpg",
		Md5:        "abc123",
		Sha1:       "def456",
		Width:      1920,
		Height:     1080,
		Permission: 1,
		UploadedIp: "203.0.113.1",
		ExpiresAt:  pgtype.Timestamptz{Valid: false},
	}
}

func testImageWithExpiry() sqlc.Image {
	img := testImage()
	img.ExpiresAt = pgtype.Timestamptz{Time: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	return img
}

func TestBuildImageUploaded(t *testing.T) {
	img := testImage()
	links := ImageLinks{
		URL:          "https://example.com/i/aBcDeFgH.jpg",
		ThumbnailURL: "https://example.com/t/abc123.png",
	}

	e := BuildImageUploaded(img, links)

	if e.ID == "" {
		t.Error("expected non-empty id")
	}
	if e.Type != TypeImageUploaded {
		t.Errorf("expected type %s, got %s", TypeImageUploaded, e.Type)
	}
	if e.Version != Version20260601 {
		t.Errorf("expected version %s, got %s", Version20260601, e.Version)
	}
	if e.OccurredAt.IsZero() {
		t.Error("expected non-zero occurred_at")
	}
	if e.IdempotencyKey != ImageUploadedKey(42) {
		t.Errorf("expected key %s, got %s", ImageUploadedKey(42), e.IdempotencyKey)
	}
	if e.Subject == nil || e.Subject.Kind != "image" || e.Subject.ID != "42" || e.Subject.Key != "aBcDeFgH" {
		t.Error("incorrect subject")
	}
	if e.Actor != nil {
		t.Error("expected nil actor in base BuildImageUploaded")
	}

	data, ok := e.Data.(ImageUploadedData)
	if !ok {
		t.Fatal("data is not ImageUploadedData")
	}
	if data.Key != "aBcDeFgH" {
		t.Errorf("expected key aBcDeFgH, got %s", data.Key)
	}
	if data.OriginName != "photo.jpg" {
		t.Errorf("expected origin_name photo.jpg, got %s", data.OriginName)
	}
	if data.SizeBytes != 102400 {
		t.Errorf("expected size_bytes 102400, got %d", data.SizeBytes)
	}
	if data.Width != 1920 || data.Height != 1080 {
		t.Errorf("expected 1920x1080, got %dx%d", data.Width, data.Height)
	}
	if data.Links.URL != links.URL {
		t.Errorf("expected url %s, got %s", links.URL, data.Links.URL)
	}
	if data.Links.ThumbnailURL != links.ThumbnailURL {
		t.Errorf("expected thumbnail_url %s, got %s", links.ThumbnailURL, data.Links.ThumbnailURL)
	}
}

func TestBuildImageUploadedWithActor(t *testing.T) {
	img := testImage()
	links := ImageLinks{URL: "https://example.com/i/aBcDeFgH.jpg", ThumbnailURL: "https://example.com/t/abc123.png"}
	actor := UserActor("alice@example.com", "Alice", 7)

	e := BuildImageUploadedWithActor(img, links, actor)

	if e.Actor == nil {
		t.Fatal("expected actor to be set")
	}
	if e.Actor.Kind != "user" {
		t.Errorf("expected actor kind user, got %s", e.Actor.Kind)
	}
	if e.Actor.Email != "alice@example.com" {
		t.Errorf("expected actor email alice@example.com, got %s", e.Actor.Email)
	}
}

func TestBuildImageUploadedWithExpiry(t *testing.T) {
	img := testImageWithExpiry()
	links := ImageLinks{URL: "https://example.com/i/aBcDeFgH.jpg", ThumbnailURL: "https://example.com/t/abc123.png"}

	e := BuildImageUploaded(img, links)

	data, ok := e.Data.(ImageUploadedData)
	if !ok {
		t.Fatal("data is not ImageUploadedData")
	}
	if data.ExpiresAt == nil {
		t.Error("expected expires_at to be set")
	}
}

func TestBuildImageProcessed(t *testing.T) {
	img := testImage()

	e := BuildImageProcessed(img, 204800, 102400, true, "completed", "approved", "noop")

	if e.Type != TypeImageProcessed {
		t.Errorf("expected type %s, got %s", TypeImageProcessed, e.Type)
	}

	data, ok := e.Data.(ImageProcessedData)
	if !ok {
		t.Fatal("data is not ImageProcessedData")
	}
	if !data.Processed {
		t.Error("expected processed=true")
	}
	if data.OriginalSizeBytes != 204800 {
		t.Errorf("expected original_size_bytes 204800, got %d", data.OriginalSizeBytes)
	}
	if data.ModerationStatus != "approved" {
		t.Errorf("expected moderation_status approved, got %s", data.ModerationStatus)
	}
	if data.Thumbnail != "completed" {
		t.Errorf("expected thumbnail completed, got %s", data.Thumbnail)
	}
}

func TestBuildImageDeleted(t *testing.T) {
	img := testImage()

	e := BuildImageDeleted(img, "owner")

	if e.Type != TypeImageDeleted {
		t.Errorf("expected type %s, got %s", TypeImageDeleted, e.Type)
	}

	data, ok := e.Data.(ImageDeletedData)
	if !ok {
		t.Fatal("data is not ImageDeletedData")
	}
	if data.DeletedBy != "owner" {
		t.Errorf("expected deleted_by owner, got %s", data.DeletedBy)
	}
	if data.Key != "aBcDeFgH" {
		t.Errorf("expected key aBcDeFgH, got %s", data.Key)
	}

	eAdmin := BuildImageDeleted(img, "admin")
	adminData := eAdmin.Data.(ImageDeletedData)
	if adminData.DeletedBy != "admin" {
		t.Errorf("expected deleted_by admin, got %s", adminData.DeletedBy)
	}

	eSystem := BuildImageDeleted(img, "system")
	sysData := eSystem.Data.(ImageDeletedData)
	if sysData.DeletedBy != "system" {
		t.Errorf("expected deleted_by system, got %s", sysData.DeletedBy)
	}
}

func TestBuildModerationReviewed(t *testing.T) {
	e := BuildModerationReviewed(42, "aBcDeFgH", "approved", 1, "looks good", nil)

	if e.Type != TypeModerationReviewed {
		t.Errorf("expected type %s, got %s", TypeModerationReviewed, e.Type)
	}

	data, ok := e.Data.(ModerationReviewedData)
	if !ok {
		t.Fatal("data is not ModerationReviewedData")
	}
	if data.Status != "approved" {
		t.Errorf("expected status approved, got %s", data.Status)
	}
	if data.ReviewerID != 1 {
		t.Errorf("expected reviewer_id 1, got %d", data.ReviewerID)
	}
	if data.Provider != "manual" {
		t.Errorf("expected provider manual, got %s", data.Provider)
	}

	eReject := BuildModerationReviewed(42, "aBcDeFgH", "rejected", 2, "nsfw", nil)
	rejectData := eReject.Data.(ModerationReviewedData)
	if rejectData.Status != "rejected" {
		t.Errorf("expected status rejected, got %s", rejectData.Status)
	}
}

func TestBuildUserRegistered(t *testing.T) {
	e := BuildUserRegistered(7, "alice@example.com", "Alice", 2, true)

	if e.Type != TypeUserRegistered {
		t.Errorf("expected type %s, got %s", TypeUserRegistered, e.Type)
	}

	data, ok := e.Data.(UserRegisteredData)
	if !ok {
		t.Fatal("data is not UserRegisteredData")
	}
	if data.UserID != 7 {
		t.Errorf("expected user_id 7, got %d", data.UserID)
	}
	if data.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", data.Email)
	}
	if !data.EmailVerificationRequired {
		t.Error("expected email_verification_required=true")
	}
}

func TestIdempotencyKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"ImageUploadedKey", ImageUploadedKey(42), "image.uploaded:42"},
		{"ImageProcessedKey", ImageProcessedKey(42), "image.processed:42"},
		{"ImageDeletedKey", ImageDeletedKey(99), "image.deleted:99"},
		{"ModerationReviewedKey", ModerationReviewedKey(42, "approved"), "moderation.reviewed:42:approved"},
		{"UserRegisteredKey", UserRegisteredKey(7), "user.registered:7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != tt.want {
				t.Errorf("expected %s, got %s", tt.want, tt.key)
			}
		})
	}
}

func TestActorConstructors(t *testing.T) {
	guest := GuestActor()
	if guest.Kind != "guest" || guest.ID != "" {
		t.Error("incorrect guest actor")
	}

	user := UserActor("a@b.com", "Bob", 5)
	if user.Kind != "user" || user.ID != "5" || user.Email != "a@b.com" || user.Name != "Bob" {
		t.Error("incorrect user actor")
	}

	admin := AdminActor(1)
	if admin.Kind != "admin" || admin.ID != "1" {
		t.Error("incorrect admin actor")
	}

	system := SystemActor()
	if system.Kind != "system" || system.ID != "0" {
		t.Error("incorrect system actor")
	}
}

func TestCollectEmitter(t *testing.T) {
	emitter := NewCollectEmitter()
	ctx := context.Background()

	e1 := BuildImageUploaded(testImage(), ImageLinks{URL: "https://example.com", ThumbnailURL: "https://example.com/thumb"})
	e2 := BuildImageDeleted(testImage(), "owner")

	if err := emitter.Emit(ctx, e1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := emitter.Emit(ctx, e2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(emitter.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(emitter.Events))
	}

	emitter.Reset()
	if len(emitter.Events) != 0 {
		t.Errorf("expected 0 events after reset, got %d", len(emitter.Events))
	}
}

func TestLogEmitter(t *testing.T) {
	emitter := NewLogEmitter()
	ctx := context.Background()

	e := BuildImageUploaded(testImage(), ImageLinks{URL: "https://example.com", ThumbnailURL: "https://example.com/thumb"})

	if err := emitter.Emit(ctx, e); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEmitAsyncDoesNotPanicOnFailedEmit(t *testing.T) {
	// Verify that EmitAsync with a failing emitter does not panic.
	// This simulates the contract: emit failure must not affect the caller.
	e := BuildImageUploaded(testImage(), ImageLinks{URL: "https://example.com", ThumbnailURL: "https://example.com/thumb"})
	EmitAsync(FailEmitter{}, e)
	// Give the goroutine time to run.
	time.Sleep(10 * time.Millisecond)
	// If we reached here without panicking, the test passes.
}

func TestEmitAsyncSequential(t *testing.T) {
	emitter := NewCollectEmitter()
	e1 := BuildImageUploaded(testImage(), ImageLinks{URL: "https://example.com", ThumbnailURL: "https://example.com/thumb"})
	e2 := BuildImageProcessed(testImage(), 100, 50, true, "completed", "approved", "noop")

	EmitAsync(emitter, e1, e2)
	time.Sleep(10 * time.Millisecond)

	if len(emitter.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(emitter.Events))
	}
	if emitter.Events[0].Type != TypeImageUploaded {
		t.Errorf("expected first event to be %s, got %s", TypeImageUploaded, emitter.Events[0].Type)
	}
	if emitter.Events[1].Type != TypeImageProcessed {
		t.Errorf("expected second event to be %s, got %s", TypeImageProcessed, emitter.Events[1].Type)
	}
}
