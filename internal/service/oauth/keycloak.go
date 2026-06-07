package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	id           string
	displayName  string
	clientID     string
	clientSecret string
	scopes       []string
	redirectURI  string
	issuer       string
	oauthCfg     *oauth2.Config
	keySet       oidc.KeySet
	relyingParty rp.RelyingParty
	userInfoURL  string
	httpClient   *http.Client
}

type OIDCProviderConfig struct {
	ID           string
	DisplayName  string
	ClientID     string
	ClientSecret string
	Issuer       string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	RedirectURI  string
	JWKSURL      string
}

func NewOIDCProvider(ctx context.Context, cfg OIDCProviderConfig) (*OIDCProvider, error) {
	if cfg.ID == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
		return nil, fmt.Errorf("oauth: missing required OIDC provider config fields")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("oauth: issuer is required for OIDC providers")
	}
	if cfg.AuthURL != "" || cfg.TokenURL != "" {
		if cfg.AuthURL == "" || cfg.TokenURL == "" || cfg.JWKSURL == "" {
			return nil, fmt.Errorf("oauth: manual mode requires auth_url, token_url, and jwks_url together")
		}
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail}
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	var (
		relyingParty rp.RelyingParty
		keySet       oidc.KeySet
		issuer       string
		err          error
	)

	if cfg.AuthURL != "" {
		oaCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
			RedirectURL: cfg.RedirectURI,
			Scopes:      scopes,
		}
		relyingParty, err = rp.NewRelyingPartyOAuth(oaCfg, rp.WithHTTPClient(httpClient))
		if err != nil {
			return nil, fmt.Errorf("oauth: create relying party for %s: %w", cfg.ID, err)
		}
		issuer = cfg.Issuer
		keySet = rp.NewRemoteKeySet(httpClient, cfg.JWKSURL)
	} else {
		relyingParty, err = rp.NewRelyingPartyOIDC(
			ctx,
			cfg.Issuer,
			cfg.ClientID,
			cfg.ClientSecret,
			cfg.RedirectURI,
			scopes,
			rp.WithHTTPClient(httpClient),
		)
		if err != nil {
			return nil, fmt.Errorf("oauth: create relying party for %s: %w", cfg.ID, err)
		}
		issuer = relyingParty.Issuer()
		if cfg.JWKSURL != "" {
			keySet = rp.NewRemoteKeySet(httpClient, cfg.JWKSURL)
		} else if rpVerifier := relyingParty.IDTokenVerifier(); rpVerifier != nil {
			keySet = (*oidc.Verifier)(rpVerifier).KeySet
		}
	}

	return &OIDCProvider{
		id:           cfg.ID,
		displayName:  cfg.DisplayName,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		scopes:       scopes,
		redirectURI:  cfg.RedirectURI,
		issuer:       issuer,
		oauthCfg:     relyingParty.OAuthConfig(),
		keySet:       keySet,
		relyingParty: relyingParty,
		userInfoURL:  cfg.UserInfoURL,
		httpClient:   httpClient,
	}, nil
}

func (p *OIDCProvider) ID() string                  { return p.id }
func (p *OIDCProvider) DisplayName() string         { return p.displayName }
func (p *OIDCProvider) OAuthConfig() *oauth2.Config { return p.oauthCfg }

func (p *OIDCProvider) StartURL(state string, pkce *PKCEConfig) string {
	opts := []oauth2.AuthCodeOption{}
	if pkce != nil && pkce.CodeVerifier != "" {
		opts = append(opts, oauth2.S256ChallengeOption(pkce.CodeVerifier))
	}
	return p.oauthCfg.AuthCodeURL(state, opts...)
}

func (p *OIDCProvider) Exchange(ctx context.Context, code string, pkce *PKCEConfig) (Identity, error) {
	opts := []oauth2.AuthCodeOption{}
	if pkce != nil && pkce.CodeVerifier != "" {
		opts = append(opts, oauth2.VerifierOption(pkce.CodeVerifier))
	}
	oauth2Token, err := p.oauthCfg.Exchange(ctx, code, opts...)
	if err != nil {
		return Identity{}, fmt.Errorf("oauth: code exchange: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return Identity{}, fmt.Errorf("oauth: no id_token in token response")
	}

	verifier := rp.NewIDTokenVerifier(p.issuer, p.clientID, p.keySet)

	claims, err := rp.VerifyIDToken[*oidc.IDTokenClaims](ctx, rawIDToken, verifier)
	if err != nil {
		return Identity{}, fmt.Errorf("oauth: verify id_token: %w", err)
	}

	id := Identity{
		Subject:       claims.GetSubject(),
		Email:         claims.Email,
		EmailVerified: bool(claims.EmailVerified),
		Name:          claims.Name,
		Avatar:        claims.Picture,
		Raw:           rawMap(claims),
	}

	if id.Name == "" {
		id.Name = claims.PreferredUsername
	}
	if id.Email == "" {
		id.Email = claims.PreferredUsername
	}

	if p.userInfoURL != "" {
		userInfo, err := p.fetchUserInfo(ctx, oauth2Token.AccessToken, p.userInfoURL)
		if err != nil {
			slog.Warn("oauth: userinfo request failed, falling back to id_token claims", "provider", p.id, "error", err)
		} else if userInfo.Subject != "" && userInfo.Subject != claims.GetSubject() {
			slog.Warn("oauth: userinfo sub mismatch with id_token, ignoring userinfo", "provider", p.id, "expected", claims.GetSubject(), "got", userInfo.Subject)
		} else {
			applyUserInfo(&id, userInfo)
		}
	} else if p.relyingParty.UserinfoEndpoint() != "" {
		userInfo, err := rp.Userinfo[*oidc.UserInfo](ctx, oauth2Token.AccessToken, oauth2Token.TokenType, claims.GetSubject(), p.relyingParty)
		if err != nil {
			slog.Warn("oauth: userinfo request failed, falling back to id_token claims", "provider", p.id, "error", err)
		} else {
			applyUserInfo(&id, userInfo)
		}
	}

	return id, nil
}

func applyUserInfo(id *Identity, userInfo *oidc.UserInfo) {
	if userInfo.Email != "" {
		id.Email = userInfo.Email
	}
	if userInfo.Name != "" {
		id.Name = userInfo.Name
	}
	if userInfo.Picture != "" {
		id.Avatar = userInfo.Picture
	}
	if userInfo.EmailVerified {
		id.EmailVerified = true
	}
}

func (p *OIDCProvider) fetchUserInfo(ctx context.Context, accessToken, url string) (*oidc.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth: userinfo returned status %d", resp.StatusCode)
	}

	var info oidc.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("oauth: decode userinfo: %w", err)
	}
	return &info, nil
}

func rawMap(v any) map[string]any {
	raw := make(map[string]any)
	b, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	_ = json.Unmarshal(b, &raw)
	return raw
}
