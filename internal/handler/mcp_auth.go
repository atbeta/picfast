package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/sqlc"
)

// MCPAuth implements auth.TokenVerifier for MCP connections.
type MCPAuth struct {
	DB *sqlc.Queries
}

func NewMCPAuth(db *sqlc.Queries) *MCPAuth {
	return &MCPAuth{DB: db}
}

func (a *MCPAuth) VerifyToken(ctx context.Context, token string, req *http.Request) (*auth.TokenInfo, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	row, err := a.DB.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	// Check expiration
	if row.ExpiresAt.Valid && row.ExpiresAt.Time.Before(time.Now()) {
		return nil, auth.ErrInvalidToken
	}

	// Check user status
	if row.Status != int16(domain.UserStatusActive) {
		return nil, auth.ErrInvalidToken
	}

	// Update last used (best effort)
	go a.DB.UpdateAPITokenLastUsed(context.Background(), row.ID)

	var scopes []string
	// scopes is stored as JSON array; try to parse
	_ = row.Scopes // ignore parse for now, default to read+write

	return &auth.TokenInfo{
		UserID:     strconv.FormatInt(row.UserID, 10),
		Scopes:     scopes,
		Expiration: row.ExpiresAt.Time,
	}, nil
}
