package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"golang-backend/config"
)

func getJWTSecret() ([]byte, error) {
	secret := config.GetEnv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET not configured")
	}
	return []byte(secret), nil
}

func GenerateAccessToken(userID int, username string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      now.Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
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

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
