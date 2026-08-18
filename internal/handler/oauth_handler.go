package handler

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/oauth"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	oauthStateCookie    = "oauth_state"
	oauthLinkUserCookie = "oauth_link_user"
	oauthPKCECookie     = "oauth_pkce"
	oauthRefreshCookie  = "picfast_refresh"
	oauthStateTTL       = 15 * time.Minute
	oauthRedirectBase   = "/console"
)

var errAccountDisabled = errors.New("user account is not active")
var errOAuthRegistrationDisabled = errors.New("oauth registration is disabled")

type OAuthHandler struct {
	db        *sqlc.Queries
	pool      *pgxpool.Pool
	jwt       *JWTService
	config    *config.Config
	providers sync.Map // map[string]oauth.Provider
}

func NewOAuthHandler(db *sqlc.Queries, pool *pgxpool.Pool, jwtSvc *JWTService, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		db:     db,
		pool:   pool,
		jwt:    jwtSvc,
		config: cfg,
	}
}

func (h *OAuthHandler) loadProvider(ctx context.Context, providerID string) (oauth.Provider, error) {
	if v, ok := h.providers.Load(providerID); ok {
		return v.(oauth.Provider), nil
	}

	pc, ok := h.config.OAuthProviderByID(providerID)
	if !ok {
		return nil, fmt.Errorf("unknown oauth provider: %s", providerID)
	}

	server := h.config.ServerSnapshot()
	redirectURI := strings.TrimRight(server.BaseURL, "/") + "/api/v1/auth/oauth/" + providerID + "/callback"

	var p oauth.Provider
	var err error

	switch pc.Type {
	case "oidc", "":
		p, err = oauth.NewOIDCProvider(ctx, oauth.OIDCProviderConfig{
			ID:           pc.ID,
			DisplayName:  pc.DisplayName,
			ClientID:     pc.ClientID,
			ClientSecret: pc.ClientSecret,
			Issuer:       pc.Issuer,
			AuthURL:      pc.AuthURL,
			TokenURL:     pc.TokenURL,
			UserInfoURL:  pc.UserInfoURL,
			Scopes:       pc.Scopes,
			RedirectURI:  redirectURI,
			JWKSURL:      pc.JWKSURL,
		})
	case "github":
		p, err = oauth.NewGitHubProvider(ctx, oauth.GitHubProviderConfig{
			ID:           pc.ID,
			DisplayName:  pc.DisplayName,
			ClientID:     pc.ClientID,
			ClientSecret: pc.ClientSecret,
			RedirectURI:  redirectURI,
		})
	default:
		return nil, fmt.Errorf("unsupported oauth provider type: %s", pc.Type)
	}

	if err != nil {
		return nil, err
	}
	actual, _ := h.providers.LoadOrStore(providerID, p)
	return actual.(oauth.Provider), nil
}

func (h *OAuthHandler) stateSecret() string {
	raw := h.config.JWT.Secret
	d := sha256.Sum256([]byte("picfast-oauth-state:" + raw))
	return hex.EncodeToString(d[:])
}

func (h *OAuthHandler) Providers(w http.ResponseWriter, r *http.Request) {
	Success(w, h.config.OAuthProviderList())
}

func (h *OAuthHandler) Start(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	provider, err := h.loadProvider(r.Context(), providerID)
	if err != nil {
		slog.Warn("oauth start: load provider", "provider", providerID, "error", err)
		Fail(w, http.StatusNotFound, "oauth provider not available")
		return
	}

	state, err := generateState()
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate pkce verifier")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    signState(state, h.stateSecret()),
		Path:     "/api/v1/auth/oauth",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     oauthPKCECookie,
		Value:    codeVerifier,
		Path:     "/api/v1/auth/oauth",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, provider.StartURL(state, &oauth.PKCEConfig{CodeVerifier: codeVerifier}), http.StatusFound)
}

func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")

	stateCookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		Fail(w, http.StatusBadRequest, "missing state cookie")
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		Fail(w, http.StatusBadRequest, "missing state parameter")
		return
	}

	if !verifyState(state, stateCookie.Value, h.stateSecret()) {
		Fail(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	isLink := false
	var linkUserID int64
	if lc, err := r.Cookie(oauthLinkUserCookie); err == nil {
		if uid, ok := verifySignedLinkUserID(lc.Value, h.stateSecret()); ok {
			linkUserID = uid
			isLink = true
		} else {
			slog.Warn("oauth callback: invalid link cookie signature")
		}
	}

	pkce, err := r.Cookie(oauthPKCECookie)
	if err != nil {
		Fail(w, http.StatusBadRequest, "missing pkce verifier")
		return
	}
	codeVerifier := pkce.Value
	if codeVerifier == "" {
		Fail(w, http.StatusBadRequest, "empty pkce verifier")
		return
	}

	clearOAuthState(w, r)

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		slog.Warn("oauth callback: provider returned error", "provider", providerID, "error", errParam, "description", errDesc)
		if isLink {
			http.Redirect(w, r, h.settingsErrorURL(r, "auth_failed"), http.StatusFound)
			return
		}
		http.Redirect(w, r, h.loginErrorURL(r, "auth_failed"), http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		if isLink {
			http.Redirect(w, r, h.settingsErrorURL(r, "no_code"), http.StatusFound)
			return
		}
		http.Redirect(w, r, h.loginErrorURL(r, "no_code"), http.StatusFound)
		return
	}

	provider, err := h.loadProvider(r.Context(), providerID)
	if err != nil {
		slog.Warn("oauth callback: load provider", "provider", providerID, "error", err)
		if isLink {
			http.Redirect(w, r, h.settingsErrorURL(r, "provider_unavailable"), http.StatusFound)
			return
		}
		http.Redirect(w, r, h.loginErrorURL(r, "provider_unavailable"), http.StatusFound)
		return
	}

	identity, err := provider.Exchange(r.Context(), code, &oauth.PKCEConfig{CodeVerifier: codeVerifier})
	if err != nil {
		slog.Warn("oauth callback: exchange failed", "provider", providerID, "error", err)
		if isLink {
			http.Redirect(w, r, h.settingsErrorURL(r, "exchange_failed"), http.StatusFound)
			return
		}
		http.Redirect(w, r, h.loginErrorURL(r, "exchange_failed"), http.StatusFound)
		return
	}

	if isLink {
		h.handleLink(w, r, linkUserID, providerID, identity)
		return
	}

	user, err := h.findOrCreateUser(r, providerID, identity)
	if err != nil {
		slog.Warn("oauth callback: find or create user", "provider", providerID, "error", err)
		if errors.Is(err, errAccountDisabled) {
			http.Redirect(w, r, h.loginErrorURL(r, "account_disabled"), http.StatusFound)
			return
		}
		if errors.Is(err, errOAuthRegistrationDisabled) {
			http.Redirect(w, r, h.loginErrorURL(r, "registration_disabled"), http.StatusFound)
			return
		}
		http.Redirect(w, r, h.loginErrorURL(r, "user_lookup_failed"), http.StatusFound)
		return
	}

	tokens, err := h.generateTokens(r.Context(), user.ID, domain.UserRole(user.Role))
	if err != nil {
		slog.Warn("oauth callback: generate tokens", "user_id", user.ID, "error", err)
		http.Redirect(w, r, h.loginErrorURL(r, "token_failed"), http.StatusFound)
		return
	}

	setAccessTokenCookie(w, r, tokens.AccessToken, tokens.ExpiresIn)
	setRefreshTokenCookie(w, r, tokens.RefreshToken, h.config.JWT.RefreshTTL)

	writeAuditLogWithActor(h.db, r, user.ID, "oauth.login", "user", fmt.Sprintf("%d", user.ID), user.Name, map[string]any{
		"provider": providerID,
		"subject":  identity.Subject,
		"email":    identity.Email,
	})

	// successURL goes to the frontend (user-facing), using WebBaseURL.
	http.Redirect(w, r, h.successURL(r), http.StatusFound)
}

func (h *OAuthHandler) handleLink(w http.ResponseWriter, r *http.Request, userID int64, providerID string, identity oauth.Identity) {
	_, err := h.db.CreateUserIdentity(r.Context(), sqlc.CreateUserIdentityParams{
		UserID:          userID,
		Provider:        providerID,
		ProviderSubject: identity.Subject,
		Email:           identity.Email,
	})
	if err != nil {
		slog.Warn("oauth link: create identity failed", "user_id", userID, "provider", providerID, "error", err)
		http.Redirect(w, r, h.settingsErrorURL(r, "link_exists"), http.StatusFound)
		return
	}

	writeAuditLog(h.db, r, "oauth.link", "user_identity", fmt.Sprintf("%d", userID), providerID, map[string]any{
		"provider": providerID,
		"subject":  identity.Subject,
		"email":    identity.Email,
		"action":   "complete",
	})

	http.Redirect(w, r, h.settingsURL(r), http.StatusFound)
}

func (h *OAuthHandler) Link(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "authentication required")
		return
	}

	provider, err := h.loadProvider(r.Context(), providerID)
	if err != nil {
		slog.Warn("oauth link: load provider", "provider", providerID, "error", err)
		Fail(w, http.StatusNotFound, "oauth provider not available")
		return
	}

	state, err := generateState()
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate pkce verifier")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    signState(state, h.stateSecret()),
		Path:     "/api/v1/auth/oauth",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     oauthLinkUserCookie,
		Value:    signSignedLinkUserID(userID, h.stateSecret()),
		Path:     "/api/v1/auth/oauth",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     oauthPKCECookie,
		Value:    codeVerifier,
		Path:     "/api/v1/auth/oauth",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})

	writeAuditLog(h.db, r, "oauth.link", "user_identity", fmt.Sprintf("%d", userID), providerID, map[string]any{
		"provider": providerID,
		"action":   "initiate",
	})

	http.Redirect(w, r, provider.StartURL(state, &oauth.PKCEConfig{CodeVerifier: codeVerifier}), http.StatusFound)
}

func (h *OAuthHandler) Unlink(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "authentication required")
		return
	}

	u, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to read user")
		return
	}
	if !u.Password.Valid {
		remaining, _ := h.db.ListUserIdentitiesByUser(r.Context(), userID)
		willLock := len(remaining) == 0
		if !willLock && len(remaining) == 1 && remaining[0].Provider == providerID {
			willLock = true
		}
		if willLock {
			Fail(w, http.StatusBadRequest, "unlink would lock out the account; set a password first")
			return
		}
	}

	err = h.db.DeleteUserIdentity(r.Context(), sqlc.DeleteUserIdentityParams{
		UserID:   userID,
		Provider: providerID,
	})
	if err != nil {
		slog.Warn("oauth unlink: delete failed", "user_id", userID, "provider", providerID, "error", err)
		Fail(w, http.StatusInternalServerError, "failed to unlink provider")
		return
	}

	writeAuditLog(h.db, r, "oauth.unlink", "user_identity", fmt.Sprintf("%d", userID), providerID, map[string]any{
		"provider": providerID,
	})

	SuccessMessage(w, "provider unlinked")
}

func (h *OAuthHandler) Identities(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "authentication required")
		return
	}

	identities, err := h.db.ListUserIdentitiesByUser(r.Context(), userID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to read identities")
		return
	}

	list := make([]map[string]any, 0, len(identities))
	for _, id := range identities {
		list = append(list, map[string]any{
			"provider":         id.Provider,
			"email":            id.Email,
			"linked_at":        id.LinkedAt,
		})
	}
	Success(w, list)
}

func (h *OAuthHandler) findOrCreateUser(r *http.Request, providerID string, identity oauth.Identity) (sqlc.User, error) {
	ctx := r.Context()

	existing, err := h.db.GetUserIdentityByProviderSubject(ctx, sqlc.GetUserIdentityByProviderSubjectParams{
		Provider:        providerID,
		ProviderSubject: identity.Subject,
	})
	if err == nil && existing.UserID > 0 {
		user, err := h.db.GetUserByID(ctx, existing.UserID)
		if err != nil {
			return sqlc.User{}, fmt.Errorf("get existing user: %w", err)
		}
		if user.Status != int16(domain.UserStatusActive) {
			return sqlc.User{}, errAccountDisabled
		}
		if identity.EmailVerified && !user.EmailVerified {
			if err := h.db.UpdateUserEmailVerified(ctx, user.ID); err != nil {
				slog.Warn("oauth: failed to update email_verified for existing identity user", "user_id", user.ID, "error", err)
			}
		}
		return user, nil
	}

	// Auto-link by email only when IdP verified the email.
	if identity.Email != "" && identity.EmailVerified {
		emailUser, err := h.db.GetUserByEmail(ctx, identity.Email)
		if err == nil && emailUser.ID > 0 {
			if emailUser.Status != int16(domain.UserStatusActive) {
				return sqlc.User{}, errAccountDisabled
			}
			if _, err := h.db.CreateUserIdentity(ctx, sqlc.CreateUserIdentityParams{
				UserID:          emailUser.ID,
				Provider:        providerID,
				ProviderSubject: identity.Subject,
				Email:           identity.Email,
			}); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					existing, lookupErr := h.db.GetUserIdentityByProviderSubject(ctx, sqlc.GetUserIdentityByProviderSubjectParams{
						Provider:        providerID,
						ProviderSubject: identity.Subject,
					})
					if lookupErr == nil && existing.UserID > 0 {
						found, ferr := h.db.GetUserByID(ctx, existing.UserID)
						if ferr == nil {
							if found.Status != int16(domain.UserStatusActive) {
								return sqlc.User{}, errAccountDisabled
							}
							return found, nil
						}
					}
					return sqlc.User{}, fmt.Errorf("identity already linked to another user")
				}
				return sqlc.User{}, fmt.Errorf("create identity: %w", err)
			}
			slog.Info("oauth: auto-linked existing user by email", "user_id", emailUser.ID, "email", identity.Email, "provider", providerID)
			if !emailUser.EmailVerified {
				if err := h.db.UpdateUserEmailVerified(ctx, emailUser.ID); err != nil {
					slog.Warn("oauth: failed to update email_verified for auto-linked user", "user_id", emailUser.ID, "error", err)
				}
			}
			return emailUser, nil
		}
	}

	app := h.config.AppSnapshot()
	if !app.AllowOauthRegistration {
		return sqlc.User{}, fmt.Errorf("%w", errOAuthRegistrationDisabled)
	}

	group, err := h.db.GetDefaultGroup(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("no default group: %w", err)
	}

	name := identity.Name
	if name == "" {
		name = strings.Split(identity.Email, "@")[0]
	}
	if name == "" {
		name = providerID + "_user"
	}

	settings, _ := json.Marshal(domain.UserSettings{})

	emailVerified := identity.EmailVerified

	var user sqlc.User

	// Check for email collision before entering the transaction — only reuse
	// existing account when the IdP has verified the email.
	emailExists := false
	if identity.Email != "" && identity.EmailVerified {
		existingByEmail, emailErr := h.db.GetUserByEmail(ctx, identity.Email)
		if emailErr == nil && existingByEmail.ID > 0 {
			emailExists = true
			user = existingByEmail
		}
	}
	if !emailExists && identity.Email != "" && !identity.EmailVerified {
		if existingByEmail, emailErr := h.db.GetUserByEmail(ctx, identity.Email); emailErr == nil && existingByEmail.ID > 0 {
			return sqlc.User{}, fmt.Errorf("email already registered; please verify your email first")
		}
	}

	if emailExists {
		_ = sqlc.RunInTx(ctx, h.pool, func(qtx *sqlc.Queries) error {
			_, err = qtx.CreateUserIdentity(ctx, sqlc.CreateUserIdentityParams{
				UserID:          user.ID,
				Provider:        providerID,
				ProviderSubject: identity.Subject,
				Email:           identity.Email,
			})
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					if existing, lookupErr := h.db.GetUserIdentityByProviderSubject(ctx, sqlc.GetUserIdentityByProviderSubjectParams{
						Provider:        providerID,
						ProviderSubject: identity.Subject,
					}); lookupErr == nil && existing.UserID > 0 {
						foundUser, ferr := h.db.GetUserByID(ctx, existing.UserID)
						if ferr == nil {
							user = foundUser
							err = nil
						}
					}
				}
			}
			return err
		})
		slog.Info("oauth: reused existing user by email", "user_id", user.ID, "email", identity.Email, "provider", providerID)
		return user, nil
	}

	err = sqlc.RunInTx(ctx, h.pool, func(qtx *sqlc.Queries) error {
		user, err = qtx.CreateUser(ctx, sqlc.CreateUserParams{
			GroupID:       domain.PgInt8(group.ID),
			Email:         identity.Email,
			Password:      pgtype.Text{},
			Name:          name,
			Role:          string(domain.RoleUser),
			CapacityBytes: resolveUserCapacity(group, app),
			Settings:      settings,
			Status:        int16(domain.UserStatusActive),
			EmailVerified: emailVerified,
			RegisteredIp:  clientip.FromRequest(r),
		})
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		_, err = qtx.CreateUserIdentity(ctx, sqlc.CreateUserIdentityParams{
			UserID:          user.ID,
			Provider:        providerID,
			ProviderSubject: identity.Subject,
			Email:           identity.Email,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				// Another request created this identity first — lookup the user.
				if existing, lookupErr := h.db.GetUserIdentityByProviderSubject(ctx, sqlc.GetUserIdentityByProviderSubjectParams{
					Provider:        providerID,
					ProviderSubject: identity.Subject,
				}); lookupErr == nil && existing.UserID > 0 {
					foundUser, foundErr := h.db.GetUserByID(ctx, existing.UserID)
					if foundErr == nil {
						user = foundUser
						err = nil
					}
				}
			}
		}
		return err
	})
	if err != nil {
		return sqlc.User{}, err
	}

	slog.Info("oauth: created new user", "user_id", user.ID, "email", identity.Email, "provider", providerID)
	return user, nil
}

func (h *OAuthHandler) generateTokens(ctx context.Context, userID int64, role domain.UserRole) (*domain.AuthTokens, error) {
	groupID := int64(0)
	user, err := h.db.GetUserByID(ctx, userID)
	if err == nil {
		if user.GroupID.Valid {
			groupID = user.GroupID.Int64
		}
	}

	accessToken, expiresIn, err := h.jwt.GenerateAccessToken(userID, role, groupID)
	if err != nil {
		return nil, err
	}

	plain, hash, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	_, err = h.db.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(h.config.JWT.RefreshTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: plain,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// successURL returns the URL to redirect after OAuth login.
// Prefers app.web_base_url; falls back to server.base_url.
// Never derives from r.Host (header-injectable).
func (h *OAuthHandler) successURL(r *http.Request) string {
	return h.webBaseURLOrRequest(r) + oauthRedirectBase
}

func (h *OAuthHandler) webBaseURLOrRequest(r *http.Request) string {
	app := h.config.AppSnapshot()
	if app.WebBaseURL != "" {
		return strings.TrimRight(app.WebBaseURL, "/")
	}
	server := h.config.ServerSnapshot()
	return strings.TrimRight(server.BaseURL, "/")
}

func (h *OAuthHandler) loginErrorURL(r *http.Request, reason string) string {
	return h.webBaseURLOrRequest(r) + "/login?oauth_error=" + url.QueryEscape(reason)
}

func (h *OAuthHandler) settingsURL(r *http.Request) string {
	return h.webBaseURLOrRequest(r) + "/console/settings"
}

func (h *OAuthHandler) settingsErrorURL(r *http.Request, reason string) string {
	return h.settingsURL(r) + "?oauth_error=" + url.QueryEscape(reason)
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func signState(state, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(state))
	return state + "." + hex.EncodeToString(mac.Sum(nil))
}

func verifyState(state, signedState, secret string) bool {
	parts := strings.SplitN(signedState, ".", 2)
	if len(parts) != 2 {
		return false
	}
	expected := signState(parts[0], secret)
	return hmac.Equal([]byte(signedState), []byte(expected)) && parts[0] == state
}

func signSignedLinkUserID(userID int64, secret string) string {
	payload := fmt.Sprintf("link:%d", userID)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil))
}

func verifySignedLinkUserID(signedValue, secret string) (int64, bool) {
	parts := strings.SplitN(signedValue, ".", 2)
	if len(parts) != 2 {
		return 0, false
	}
	payload := parts[0]
	sig := parts[1]

	if !strings.HasPrefix(payload, "link:") {
		return 0, false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return 0, false
	}

	uid, err := strconv.ParseInt(strings.TrimPrefix(payload, "link:"), 10, 64)
	if err != nil {
		return 0, false
	}
	return uid, true
}

func setRefreshTokenCookie(w http.ResponseWriter, r *http.Request, token string, maxAge time.Duration) {
	if token == "" {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthRefreshCookie,
		Value:    token,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthRefreshCookie,
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOAuthState(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/api/v1/auth/oauth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     oauthLinkUserCookie,
		Value:    "",
		Path:     "/api/v1/auth/oauth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     oauthPKCECookie,
		Value:    "",
		Path:     "/api/v1/auth/oauth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}
