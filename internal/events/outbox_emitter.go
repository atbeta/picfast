package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type OutboxEmitter struct {
	db *sqlc.Queries
}

func NewOutboxEmitter(db *sqlc.Queries) *OutboxEmitter {
	return &OutboxEmitter{db: db}
}

func (e *OutboxEmitter) Emit(ctx context.Context, event Envelope) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	ownerUserID := extractOwnerUserID(event)

	_, err = e.db.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		Type:           event.Type,
		Version:        event.Version,
		IdempotencyKey: event.IdempotencyKey,
		Payload:        payload,
		OwnerUserID:    ownerUserID,
	})
	if err != nil {
		slog.Warn("outbox insert failed", "idempotency_key", event.IdempotencyKey, "error", err)
		return nil
	}
	return nil
}

func extractOwnerUserID(event Envelope) pgtype.Int8 {
	switch d := event.Data.(type) {
	case ImageUploadedData:
		if d.UserID != nil {
			return pgtype.Int8{Int64: *d.UserID, Valid: true}
		}
	case ImageProcessedData:
		if d.UserID != nil {
			return pgtype.Int8{Int64: *d.UserID, Valid: true}
		}
	case ImageDeletedData:
		if d.UserID != nil {
			return pgtype.Int8{Int64: *d.UserID, Valid: true}
		}
	case ModerationReviewedData:
		if d.OwnerID != nil {
			return pgtype.Int8{Int64: *d.OwnerID, Valid: true}
		}
	case UserRegisteredData:
		return pgtype.Int8{Int64: d.UserID, Valid: true}
	}
	return pgtype.Int8{}
}
