package handler

import (
	"strconv"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
)

func TestJWTServiceGenerateAccessTokenIncludesStandardClaims(t *testing.T) {
	svc := NewJWTService(&config.JWTConfig{
		Secret:    "test-secret",
		AccessTTL: time.Hour,
	})

	token, _, err := svc.GenerateAccessToken(42, domain.RoleAdmin, 7)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}

	if claims.Issuer != "picfast" {
		t.Fatalf("Issuer = %q, want picfast", claims.Issuer)
	}
	if claims.Subject != strconv.FormatInt(claims.UserID, 10) {
		t.Fatalf("Subject = %q, want user id", claims.Subject)
	}
	if claims.ID == "" {
		t.Fatal("expected JWT ID to be set")
	}
}
