package oauth

import "context"

type Identity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Avatar        string
	Raw           map[string]any
}

type PKCEConfig struct {
	CodeVerifier string
}

type Provider interface {
	ID() string
	DisplayName() string
	StartURL(state string, pkce *PKCEConfig) string
	Exchange(ctx context.Context, code string, pkce *PKCEConfig) (Identity, error)
}
