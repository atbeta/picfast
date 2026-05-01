package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDefaultMCScopes(t *testing.T) {
	t.Run("defaults to read and write when empty", func(t *testing.T) {
		scopes := defaultMCScopes(nil)
		if len(scopes) != 2 || scopes[0] != mcpScopeRead || scopes[1] != mcpScopeWrite {
			t.Fatalf("unexpected default scopes: %#v", scopes)
		}
	})

	t.Run("normalizes case and duplicates", func(t *testing.T) {
		scopes := defaultMCScopes([]string{" Read ", "write", "READ"})
		if len(scopes) != 2 || scopes[0] != mcpScopeRead || scopes[1] != mcpScopeWrite {
			t.Fatalf("unexpected normalized scopes: %#v", scopes)
		}
	})
}

func TestRequireMCPScopeInfo(t *testing.T) {
	t.Run("allows granted scope", func(t *testing.T) {
		err := requireMCPScopeInfo(authTokenInfo{Scopes: []string{mcpScopeRead}}.toSDK(), mcpScopeRead)
		if err != nil {
			t.Fatalf("requireMCPScopeInfo returned error: %v", err)
		}
	})

	t.Run("rejects missing scope", func(t *testing.T) {
		err := requireMCPScopeInfo(authTokenInfo{Scopes: []string{mcpScopeRead}}.toSDK(), mcpScopeWrite)
		if err == nil || err.Error() != "forbidden: write scope required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestLoadOwnedImageByKey(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	_ = pool

	group := testutil.SeedDefaultGroup(t, db)
	strategy := testutil.SeedStrategy(t, db, group.ID)
	owner := testutil.SeedUser(t, db, group.ID, "owner@example.com", "password123", string(domain.RoleUser))
	other := testutil.SeedUser(t, db, group.ID, "other@example.com", "password123", string(domain.RoleUser))

	img, err := db.CreateImage(t.Context(), sqlc.CreateImageParams{
		UserID:     domain.PgInt8(owner.ID),
		GroupID:    domain.PgInt8(group.ID),
		StrategyID: domain.PgInt8(strategy.ID),
		Key:        "owned-image-key",
		Path:       "2026/04/29",
		Name:       "owned.png",
		OriginName: "owned.png",
		SizeBytes:  123,
		Mimetype:   "image/png",
		Extension:  "png",
		Md5:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Sha1:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Permission: int16(domain.PermissionPublic),
		UploadedIp: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("create image: %v", err)
	}

	t.Run("owner can load image", func(t *testing.T) {
		got, err := loadOwnedImageByKey(t.Context(), db, owner.ID, img.Key)
		if err != nil {
			t.Fatalf("loadOwnedImageByKey returned error: %v", err)
		}
		if got.ID != img.ID {
			t.Fatalf("loaded image id = %d, want %d", got.ID, img.ID)
		}
	})

	t.Run("other user cannot load image", func(t *testing.T) {
		_, err := loadOwnedImageByKey(t.Context(), db, other.ID, img.Key)
		if !errors.Is(err, errMCPImageNotFound) {
			t.Fatalf("expected errMCPImageNotFound, got %v", err)
		}
	})
}

func TestMCPAuthVerifyTokenParsesScopes(t *testing.T) {
	_, db := testutil.SetupDB(t)

	group := testutil.SeedDefaultGroup(t, db)
	_ = testutil.SeedStrategy(t, db, group.ID)
	user := testutil.SeedUser(t, db, group.ID, "token@example.com", "password123", string(domain.RoleUser))

	plain := "img_test_token"
	hash := sha256.Sum256([]byte(plain))
	scopesJSON, err := json.Marshal([]string{mcpScopeRead})
	if err != nil {
		t.Fatalf("marshal scopes: %v", err)
	}

	if _, err := db.CreateAPIToken(t.Context(), sqlc.CreateAPITokenParams{
		UserID:    user.ID,
		Name:      "read-only",
		TokenHash: hex.EncodeToString(hash[:]),
		Scopes:    scopesJSON,
		ExpiresAt: pgtype.Timestamptz{},
	}); err != nil {
		t.Fatalf("create api token: %v", err)
	}

	info, err := NewMCPAuth(db).VerifyToken(context.Background(), plain, nil)
	if err != nil {
		t.Fatalf("VerifyToken returned error: %v", err)
	}
	if info.UserID != strconv.FormatInt(user.ID, 10) {
		t.Fatalf("user id = %q, want %q", info.UserID, strconv.FormatInt(user.ID, 10))
	}
	if len(info.Scopes) != 1 || info.Scopes[0] != mcpScopeRead {
		t.Fatalf("scopes = %#v, want [%q]", info.Scopes, mcpScopeRead)
	}
	if info.Expiration.IsZero() {
		t.Fatalf("expiration should not be zero for non-expiring token")
	}
}

func TestMCPErrorResultFormat(t *testing.T) {
	res := mcpErrorResult(mcpErrorUploadFailed, "upload failed: invalid format")
	if !res.IsError {
		t.Fatalf("expected IsError=true")
	}
	if len(res.Content) != 1 {
		t.Fatalf("expected single content entry, got %d", len(res.Content))
	}

	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content")
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(text.Text), &payload); err != nil {
		t.Fatalf("expected JSON payload, got error: %v", err)
	}
	if payload.Error.Code != mcpErrorUploadFailed {
		t.Fatalf("unexpected error code %q", payload.Error.Code)
	}
	if payload.Error.Message == "" {
		t.Fatalf("error message should not be empty")
	}
}

type authTokenInfo struct {
	Scopes []string
}

func (i authTokenInfo) toSDK() *auth.TokenInfo {
	return &auth.TokenInfo{Scopes: i.Scopes}
}
