package oauth

import (
	"context"
	"strings"
	"testing"
)

func TestNewOIDCProvider_ValidationErrors(t *testing.T) {
	base := OIDCProviderConfig{
		ID:           "test",
		DisplayName:  "Test",
		ClientID:     "cid",
		ClientSecret: "csec",
		Issuer:       "https://issuer.example.com",
		RedirectURI:  "https://app.example.com/cb",
	}

	tests := []struct {
		name    string
		mutate  func(*OIDCProviderConfig)
		wantErr string
	}{
		{
			name:    "missing id",
			mutate:  func(c *OIDCProviderConfig) { c.ID = "" },
			wantErr: "missing required",
		},
		{
			name:    "missing client id",
			mutate:  func(c *OIDCProviderConfig) { c.ClientID = "" },
			wantErr: "missing required",
		},
		{
			name:    "missing client secret",
			mutate:  func(c *OIDCProviderConfig) { c.ClientSecret = "" },
			wantErr: "missing required",
		},
		{
			name:    "missing redirect uri",
			mutate:  func(c *OIDCProviderConfig) { c.RedirectURI = "" },
			wantErr: "missing required",
		},
		{
			name:    "missing issuer",
			mutate:  func(c *OIDCProviderConfig) { c.Issuer = "" },
			wantErr: "issuer is required",
		},
		{
			name: "manual mode - only auth_url",
			mutate: func(c *OIDCProviderConfig) {
				c.AuthURL = "https://issuer.example.com/auth"
			},
			wantErr: "manual mode requires",
		},
		{
			name: "manual mode - only token_url",
			mutate: func(c *OIDCProviderConfig) {
				c.TokenURL = "https://issuer.example.com/token"
			},
			wantErr: "manual mode requires",
		},
		{
			name: "manual mode - auth and token without jwks",
			mutate: func(c *OIDCProviderConfig) {
				c.AuthURL = "https://issuer.example.com/auth"
				c.TokenURL = "https://issuer.example.com/token"
			},
			wantErr: "manual mode requires",
		},
		{
			name: "manual mode - auth/token plus userinfo (no jwks)",
			mutate: func(c *OIDCProviderConfig) {
				c.AuthURL = "https://issuer.example.com/auth"
				c.TokenURL = "https://issuer.example.com/token"
				c.UserInfoURL = "https://issuer.example.com/userinfo"
			},
			wantErr: "manual mode requires",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			tt.mutate(&cfg)
			_, err := NewOIDCProvider(context.Background(), cfg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestNewOIDCProvider_ManualModeFullSet(t *testing.T) {
	cfg := OIDCProviderConfig{
		ID:           "test",
		DisplayName:  "Test",
		ClientID:     "cid",
		ClientSecret: "csec",
		Issuer:       "https://issuer.example.com",
		AuthURL:      "https://issuer.example.com/auth",
		TokenURL:     "https://issuer.example.com/token",
		JWKSURL:      "https://issuer.example.com/jwks",
		UserInfoURL:  "https://issuer.example.com/userinfo",
		RedirectURI:  "https://app.example.com/cb",
	}

	p, err := NewOIDCProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != "test" {
		t.Fatalf("expected id %q, got %q", "test", p.ID())
	}
	if p.DisplayName() != "Test" {
		t.Fatalf("expected display name %q, got %q", "Test", p.DisplayName())
	}
	if p.userInfoURL != cfg.UserInfoURL {
		t.Fatalf("expected userInfoURL %q, got %q", cfg.UserInfoURL, p.userInfoURL)
	}
	if p.issuer != cfg.Issuer {
		t.Fatalf("expected issuer %q, got %q", cfg.Issuer, p.issuer)
	}
	if p.keySet == nil {
		t.Fatal("expected keySet to be set in manual mode")
	}
	if p.relyingParty == nil {
		t.Fatal("expected relyingParty to be set")
	}
}
