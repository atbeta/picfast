package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	cfg *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) signingMethod() jwt.SigningMethod {
	switch s.cfg.SigningMethod {
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	default:
		return jwt.SigningMethodHS256
	}
}

func (s *JWTService) GenerateAccessToken(userID int64, role domain.UserRole, groupID int64) (string, int64, error) {
	expiresAt := time.Now().Add(s.cfg.AccessTTL)
	claims := domain.TokenClaims{
		UserID:  userID,
		Role:    role,
		GroupID: groupID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(s.signingMethod(), claims)
	signed, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", 0, err
	}
	return signed, int64(s.cfg.AccessTTL.Seconds()), nil
}

func (s *JWTService) ValidateAccessToken(tokenStr string) (*domain.TokenClaims, error) {
	expectedMethod := s.signingMethod()
	token, err := jwt.ParseWithClaims(tokenStr, &domain.TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != expectedMethod.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*domain.TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func GenerateRefreshToken() (plain string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(h[:])
	return plain, hash, nil
}

func HashRefreshToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(h[:])
}
