package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubProvider struct {
	id           string
	displayName  string
	clientID     string
	clientSecret string
	redirectURI  string
	oauthCfg     *oauth2.Config
	httpClient   *http.Client
}

type GitHubProviderConfig struct {
	ID           string
	DisplayName  string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func NewGitHubProvider(ctx context.Context, cfg GitHubProviderConfig) (*GitHubProvider, error) {
	if cfg.ID == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
		return nil, fmt.Errorf("github: missing required provider config fields")
	}
	return &GitHubProvider{
		id:           cfg.ID,
		displayName:  cfg.DisplayName,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURI:  cfg.RedirectURI,
		oauthCfg: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  cfg.RedirectURI,
			Scopes:       []string{"read:user", "user:email"},
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *GitHubProvider) ID() string          { return p.id }
func (p *GitHubProvider) DisplayName() string { return p.displayName }

func (p *GitHubProvider) StartURL(state string, pkce *PKCEConfig) string {
	return p.oauthCfg.AuthCodeURL(state)
}

func (p *GitHubProvider) Exchange(ctx context.Context, code string, pkce *PKCEConfig) (Identity, error) {
	oauth2Token, err := p.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return Identity{}, fmt.Errorf("github: code exchange: %w", err)
	}

	id := Identity{
		Email:         "",
		EmailVerified: false,
		Raw:           make(map[string]any),
	}

	ghUser, err := p.fetchUser(ctx, oauth2Token.AccessToken)
	if err != nil {
		return Identity{}, fmt.Errorf("github: fetch user: %w", err)
	}

	id.Subject = fmt.Sprintf("%d", ghUser.ID)
	id.Name = ghUser.Name
	if id.Name == "" {
		id.Name = ghUser.Login
	}
	id.Avatar = ghUser.AvatarURL

	email, verified, err := p.fetchPrimaryEmail(ctx, oauth2Token.AccessToken)
	if err != nil {
		slog.Warn("github: failed to fetch emails, using login", "login", ghUser.Login, "error", err)
		id.Email = ghUser.Email
	} else {
		id.Email = email
		id.EmailVerified = verified
	}
	if id.Email == "" {
		id.Email = ghUser.Login + "@github"
	}

	id.Raw["github_id"] = ghUser.ID
	id.Raw["github_login"] = ghUser.Login
	id.Raw["github_avatar"] = ghUser.AvatarURL

	return id, nil
}

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *GitHubProvider) fetchUser(ctx context.Context, token string) (*githubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: user api returned %d: %s", resp.StatusCode, string(body))
	}

	var u githubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *GitHubProvider) fetchPrimaryEmail(ctx context.Context, token string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("github: emails api returned %d", resp.StatusCode)
	}

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, e.Verified, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, true, nil
		}
	}
	return "", false, fmt.Errorf("github: no primary or verified email found")
}
