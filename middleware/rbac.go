package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

// PermissionChecker diimplementasikan *roles.Service
// (didefinisikan ulang di sini agar middleware tidak mengimpor features/roles).
type PermissionChecker interface {
	CheckPermission(ctx context.Context, userCode, permCode string) error
}

// Dipasang setelah NewAuth agar user_code sudah ada di context.
func RequirePermission(checker PermissionChecker, permCodes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code, ok := GetUserCode(r)
			if !ok {
				writeAuthError(w, "Missing authorization")
				return
			}
			for _, perm := range permCodes {
				if err := checker.CheckPermission(r.Context(), code, perm); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden: missing " + perm})
					return
				}
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
