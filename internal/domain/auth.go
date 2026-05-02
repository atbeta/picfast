package domain

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID  int64    `json:"uid"`
	Role    UserRole `json:"role"`
	GroupID int64    `json:"gid"`
	jwt.RegisteredClaims
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type ContextKey string

const (
	ContextKeyUserID   ContextKey = "user_id"
	ContextKeyRole     ContextKey = "role"
	ContextKeyGroupID  ContextKey = "group_id"
	ContextKeyScopes  ContextKey = "scopes"
)
