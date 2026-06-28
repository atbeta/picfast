package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atbeta/picfast/internal/sqlc"
)

type Service struct {
	db      *sqlc.Queries
	maxPerUser int
	encKey  []byte
}

func NewService(db *sqlc.Queries, maxPerUser int, encKey []byte) *Service {
	return &Service{db: db, maxPerUser: maxPerUser, encKey: encKey}
}

type CreateParams struct {
	UserID int64    `json:"-"`
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type CreateResult struct {
	sqlc.Webhook
	Secret string `json:"secret"`
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*CreateResult, error) {
	count, err := s.db.CountWebhooksByUser(ctx, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("count webhooks: %w", err)
	}
	if int(count) >= s.maxPerUser {
		return nil, fmt.Errorf("webhook limit reached (%d)", s.maxPerUser)
	}

	plain, hash, ciphertext, err := GenerateSecret(s.encKey)
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}

	eventsJSON, _ := json.Marshal(params.Events)

	wh, err := s.db.CreateWebhook(ctx, sqlc.CreateWebhookParams{
		UserID:           params.UserID,
		Name:             params.Name,
		Url:              params.URL,
		SecretHash:       hash,
		SecretCiphertext: ciphertext,
		Events:           eventsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create webhook: %w", err)
	}

	return &CreateResult{Webhook: wh, Secret: plain}, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID int64) (sqlc.Webhook, error) {
	wh, err := s.db.GetWebhookByID(ctx, id)
	if err != nil {
		return sqlc.Webhook{}, err
	}
	if wh.UserID != userID {
		return sqlc.Webhook{}, fmt.Errorf("webhook not found")
	}
	return wh, nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]sqlc.Webhook, error) {
	return s.db.ListWebhooksByUser(ctx, userID)
}

type UpdateParams struct {
	Name    *string  `json:"name,omitempty"`
	URL     *string  `json:"url,omitempty"`
	Events  []string `json:"events,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
}

func (s *Service) Update(ctx context.Context, id, userID int64, params UpdateParams) (sqlc.Webhook, error) {
	existing, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return sqlc.Webhook{}, err
	}

	name := existing.Name
	if params.Name != nil && *params.Name != "" {
		name = *params.Name
	}
	url := existing.Url
	if params.URL != nil && *params.URL != "" {
		url = *params.URL
	}
	var eventsJSON json.RawMessage = existing.Events
	if params.Events != nil {
		eventsJSON, _ = json.Marshal(params.Events)
	}
	enabled := existing.Enabled
	if params.Enabled != nil {
		enabled = *params.Enabled
	}

	wh, err := s.db.UpdateWebhook(ctx, sqlc.UpdateWebhookParams{
		ID:      id,
		UserID:  userID,
		Name:    name,
		Url:     url,
		Events:  eventsJSON,
		Enabled: enabled,
	})
	if err != nil {
		return sqlc.Webhook{}, fmt.Errorf("update webhook: %w", err)
	}
	return wh, nil
}

func (s *Service) Delete(ctx context.Context, id, userID int64) error {
	return s.db.DeleteWebhook(ctx, sqlc.DeleteWebhookParams{
		ID:     id,
		UserID: userID,
	})
}

func (s *Service) RotateSecret(ctx context.Context, id, userID int64) (string, error) {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return "", err
	}

	plain, hash, ciphertext, err := GenerateSecret(s.encKey)
	if err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
	}

	if err := s.db.UpdateWebhookSecret(ctx, sqlc.UpdateWebhookSecretParams{
		ID:               id,
		SecretHash:       hash,
		SecretCiphertext: ciphertext,
	}); err != nil {
		return "", fmt.Errorf("update secret: %w", err)
	}

	return plain, nil
}
