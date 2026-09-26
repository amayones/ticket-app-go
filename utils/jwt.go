package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token lifetimes (single source of truth; service re-exports them).
const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
	TokenIssuer     = "go-core"
	MinJWTSecretLen = 32
)

func ValidateSecret(secret string) error {
	if len(strings.TrimSpace(secret)) < MinJWTSecretLen {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

// GenerateAccessToken signs HS256 with iss/aud claims. Identity is the
// public user CODE (never the numeric ID). Secret is injected
// (no utils->config import) so the package stays pure and testable.
// Umur token = AccessTokenTTL default; untuk umur custom (mis. dari Config)
// pakai GenerateAccessTokenWithTTL.
func GenerateAccessToken(secret, userCode, username, roleCode string) (string, error) {
	return GenerateAccessTokenWithTTL(secret, userCode, username, roleCode, AccessTokenTTL)
}

// GenerateAccessTokenWithTTL sama seperti GenerateAccessToken dengan umur
// token eksplisit. ttl <= 0 fallback ke AccessTokenTTL.
func GenerateAccessTokenWithTTL(secret, userCode, username, roleCode string, ttl time.Duration) (string, error) {
	if err := ValidateSecret(secret); err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = AccessTokenTTL
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"user_code": userCode,
		"username":  username,
		"role":      roleCode,
		"iss":       TokenIssuer,
		"aud":       TokenIssuer,
		"iat":       now.Unix(),
		"exp":       now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken pins HS256 explicitly (rejects HS384/512 confusion),
// checks iss/aud and expiry.
func ValidateToken(tokenString, secret string) (jwt.MapClaims, error) {
	if err := ValidateSecret(secret); err != nil {
		return nil, err
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(TokenIssuer), jwt.WithAudience(TokenIssuer))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// GenerateRefreshToken returns a 256-bit hex token for transport (cookie/body).
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken stores only the SHA-256 of the refresh token in DB,
// so a DB leak alone cannot hijack sessions.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
