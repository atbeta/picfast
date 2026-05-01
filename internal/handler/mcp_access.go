package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

const (
	mcpScopeRead  = "read"
	mcpScopeWrite = "write"
)

var errMCPImageNotFound = errors.New("image not found")

func defaultMCScopes(scopes []string) []string {
	normalized := normalizeMCScopes(scopes)
	if len(normalized) == 0 {
		return []string{mcpScopeRead, mcpScopeWrite}
	}
	return normalized
}

func normalizeMCScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(scopes))
	normalized := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	return normalized
}

func hasMCPScope(scopes []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return false
	}
	for _, scope := range defaultMCScopes(scopes) {
		if scope == target {
			return true
		}
	}
	return false
}

func requireMCPScopeInfo(info *auth.TokenInfo, scope string) error {
	if info == nil {
		return fmt.Errorf("unauthorized")
	}
	if !hasMCPScope(info.Scopes, scope) {
		return fmt.Errorf("forbidden: %s scope required", scope)
	}
	return nil
}

func requireMCPScope(ctx context.Context, scope string) error {
	return requireMCPScopeInfo(auth.TokenInfoFromContext(ctx), scope)
}

func imageOwnedByUser(img sqlc.Image, userID int64) bool {
	return img.UserID.Valid && img.UserID.Int64 == userID
}

func loadOwnedImageByKey(ctx context.Context, db *sqlc.Queries, userID int64, key string) (sqlc.Image, error) {
	img, err := db.GetImageByKey(ctx, key)
	if err != nil {
		return sqlc.Image{}, errMCPImageNotFound
	}
	if !imageOwnedByUser(img, userID) {
		return sqlc.Image{}, errMCPImageNotFound
	}
	return img, nil
}
