package middleware

import (
	"context"
	"net/http"
	"strings"

	"golang-backend/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		claims, err := utils.ValidateToken(strings.TrimSpace(parts[1]))
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		userID, err := parseUserIDClaim(claims)
		if err != nil {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseUserIDClaim(claims map[string]interface{}) (int, error) {
	raw, ok := claims["user_id"]
	if !ok {
		return 0, ErrMissingClaim
	}
	switch v := raw.(type) {
	case float64:
		if v <= 0 || v != float64(int(v)) {
			return 0, ErrInvalidClaim
		}
		return int(v), nil
	case float32:
		if v <= 0 || float64(v) != float64(int(v)) {
			return 0, ErrInvalidClaim
		}
		return int(v), nil
	case int:
		if v <= 0 {
			return 0, ErrInvalidClaim
		}
		return v, nil
	case int64:
		if v <= 0 {
			return 0, ErrInvalidClaim
		}
		return int(v), nil
	default:
		return 0, ErrInvalidClaim
	}
}

var (
	ErrMissingClaim = errString("missing user_id claim")
	ErrInvalidClaim = errString("invalid user_id claim")
)

type errString string

func (e errString) Error() string { return string(e) }

func GetUserID(r *http.Request) (int, bool) {
	v := r.Context().Value(UserIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}
