package web

// Shared kernel HTTP: tulis JSON, batasi body, paginasi, baca path param.
// Dipakai semua handler fitur (features/<menu>) agar bentuknya seragam.
import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"golang-backend/services"
)

// MaxBodyBytes caps JSON bodies (DoS protection).
const MaxBodyBytes = 1 << 20 // 1 MB

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return false
	}
	return true
}

// Paginate reads limit/offset query params with safe defaults.
func Paginate(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

// PathCode reads a non-empty chi URL param (kode publik, bukan ID numerik).
func PathCode(r *http.Request, name string) (string, bool) {
	code := strings.TrimSpace(chi.URLParam(r, name))
	if code == "" {
		return "", false
	}
	return code, true
}

// PathInt reads a positive integer chi URL param (mis. session ID numerik).
func PathInt(r *http.Request, name string) (int, bool) {
	raw, ok := PathCode(r, name)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// ServiceError maps domain errors to HTTP statuses (single source for all
// menus). Returns the status so callers can add behavior (e.g. syslog on 500).
func ServiceError(w http.ResponseWriter, err error) int {
	switch {
	case errors.Is(err, services.ErrInputRequired),
		errors.Is(err, services.ErrUsernameTooShort),
		errors.Is(err, services.ErrInvalidEmail),
		errors.Is(err, services.ErrPasswordTooShort),
		errors.Is(err, services.ErrPasswordTooLong),
		errors.Is(err, services.ErrInvalidRole),
		errors.Is(err, services.ErrInvalidTemplate),
		errors.Is(err, services.ErrInvalidChannel),
		errors.Is(err, services.ErrInvalidMenu):
		WriteError(w, http.StatusBadRequest, err.Error())
		return http.StatusBadRequest
	case errors.Is(err, services.ErrUsernameTaken),
		errors.Is(err, services.ErrEmailTaken),
		errors.Is(err, services.ErrRoleExists),
		errors.Is(err, services.ErrRoleInUse),
		errors.Is(err, services.ErrPermissionExists),
		errors.Is(err, services.ErrMenuInUse),
		errors.Is(err, services.ErrMenuHasChildren):
		WriteError(w, http.StatusConflict, err.Error())
		return http.StatusConflict
	case errors.Is(err, services.ErrInvalidLogin),
		errors.Is(err, services.ErrInvalidRefresh):
		WriteError(w, http.StatusUnauthorized, err.Error())
		return http.StatusUnauthorized
	case errors.Is(err, services.ErrForbidden),
		errors.Is(err, services.ErrRoleProtected):
		WriteError(w, http.StatusForbidden, err.Error())
		return http.StatusForbidden
	case errors.Is(err, services.ErrUserNotFound),
		errors.Is(err, services.ErrRoleNotFound),
		errors.Is(err, services.ErrTemplateNotFound),
		errors.Is(err, services.ErrSessionNotFound),
		errors.Is(err, services.ErrMenuNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
		return http.StatusNotFound
	default:
		slog.Error("internal error", "err", err)
		WriteError(w, http.StatusInternalServerError, "Something went wrong")
		return http.StatusInternalServerError
	}
}

// ServiceErrorWithSyslog maps the error like ServiceError and forwards
// unexpected 500s to syslog (best-effort). Pakai helper ini di semua handler
// agar perilaku log 500 konsisten antar menu.
func ServiceErrorWithSyslog(w http.ResponseWriter, err error, log func(msg string)) int {
	status := ServiceError(w, err)
	if status == http.StatusInternalServerError && log != nil {
		log(err.Error())
	}
	return status
}
