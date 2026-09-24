package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang-backend/utils"
)

// Private context key type prevents collisions with other packages.
type ctxKey string

const (
	userIDKey   ctxKey = "user_code"
	userRoleKey ctxKey = "user_role"
)

var (
	ErrMissingClaim = errors.New("missing user_code claim")
	ErrInvalidClaim = errors.New("invalid user_code claim")
)

// NewAuth returns auth middleware bound to the configured JWT secret.
// Secret is injected (no global env read) so routes/tests control it.
func NewAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, "Missing authorization header")
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
				writeAuthError(w, "Invalid authorization header format")
				return
			}
			claims, err := utils.ValidateToken(strings.TrimSpace(parts[1]), jwtSecret)
			if err != nil {
				writeAuthError(w, "Invalid or expired token")
				return
			}
		userCode, err := parseUserCodeClaim(claims)
		if err != nil {
			writeAuthError(w, "Invalid token claims")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userCode)
		if role, ok := claims["role"].(string); ok {
			ctx = context.WithValue(ctx, userRoleKey, role)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// parseUserCodeClaim accepts the current string code and the legacy
// numeric user_id (tokens issued before the CODE migration).
func parseUserCodeClaim(claims map[string]interface{}) (string, error) {
	if raw, ok := claims["user_code"]; ok {
		if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s), nil
		}
		return "", ErrInvalidClaim
	}
	if raw, ok := claims["user_id"]; ok {
		switch v := raw.(type) {
		case float64:
			if v <= 0 || v != float64(int(v)) {
				return "", ErrInvalidClaim
			}
			return fmt.Sprintf("LEGACY-%d", int(v)), nil
		case json.Number:
			n, err := v.Int64()
			if err != nil || n <= 0 {
				return "", ErrInvalidClaim
			}
			return fmt.Sprintf("LEGACY-%d", n), nil
		case string:
			if strings.TrimSpace(v) == "" {
				return "", ErrInvalidClaim
			}
			return strings.TrimSpace(v), nil
		}
	}
	return "", ErrMissingClaim
}

// GetUserCode extracts the authenticated user CODE from context.
func GetUserCode(r *http.Request) (string, bool) {
	v := r.Context().Value(userIDKey)
	if v == nil {
		return "", false
	}
	code, ok := v.(string)
	return code, ok && code != ""
}

// GetUserRole extracts the role claim (empty when absent).
func GetUserRole(r *http.Request) string {
	role, _ := r.Context().Value(userRoleKey).(string)
	return role
}
