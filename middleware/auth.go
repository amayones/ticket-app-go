package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang-backend/utils"
)

// Private context key type prevents collisions with other packages.
type ctxKey struct{}

var userIDKey = ctxKey{}

var (
	ErrMissingClaim = errors.New("missing user_id claim")
	ErrInvalidClaim = errors.New("invalid user_id claim")
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
			userID, err := parseUserIDClaim(claims)
			if err != nil {
				writeAuthError(w, "Invalid token claims")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
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
	case json.Number:
		n, err := v.Int64()
		if err != nil || n <= 0 {
			return 0, ErrInvalidClaim
		}
		return int(n), nil
	case string:
		// Some issuers encode IDs as strings.
		var n json.Number = json.Number(strings.TrimSpace(v))
		i, err := n.Int64()
		if err != nil || i <= 0 {
			return 0, ErrInvalidClaim
		}
		return int(i), nil
	default:
		return 0, ErrInvalidClaim
	}
}

// GetUserID extracts the authenticated user ID from context.
func GetUserID(r *http.Request) (int, bool) {
	v := r.Context().Value(userIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}
