package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/services"
)

// MaxBodyBytes caps JSON bodies (DoS protection).
const MaxBodyBytes = 1 << 20 // 1 MB

// UserHandler depends on the service interface (mockable), not concrete.
type UserHandler struct {
	Service services.UserServiceInterface
}

func NewUserHandler(service services.UserServiceInterface) *UserHandler {
	return &UserHandler{Service: service}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return false
	}
	return true
}

func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInputRequired),
		errors.Is(err, services.ErrUsernameTooShort),
		errors.Is(err, services.ErrInvalidEmail),
		errors.Is(err, services.ErrPasswordTooShort),
		errors.Is(err, services.ErrPasswordTooLong):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrUsernameTaken),
		errors.Is(err, services.ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidLogin),
		errors.Is(err, services.ErrInvalidRefresh):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, services.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, services.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		slog.Error("internal error", "err", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong")
	}
}

func parseID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func requireSelf(w http.ResponseWriter, r *http.Request, id int) bool {
	callerID, ok := middleware.GetUserID(r)
	if !ok || callerID != id {
		writeError(w, http.StatusForbidden, services.ErrForbidden.Error())
		return false
	}
	return true
}

func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginate(r)
	users, err := h.Service.ListUsers(r.Context(), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if users == nil {
		users = []models.UserResponse{}
	}
	writeJSON(w, http.StatusOK, users)
}

func paginate(r *http.Request) (limit, offset int) {
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

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	user, err := h.Service.GetUserByID(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user.ToResponse())
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	id, err := h.Service.CreateUser(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"id":      id,
	})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if !requireSelf(w, r, id) {
		return
	}
	var req models.UpdateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.UpdateUser(r.Context(), id, req); err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if !requireSelf(w, r, id) {
		return
	}
	if err := h.Service.DeleteUser(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	accessToken, refreshToken, err := h.Service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	accessToken, newRefreshToken, err := h.Service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.LogoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.Logout(r.Context(), req.RefreshToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *UserHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if !requireSelf(w, r, id) {
		return
	}
	if err := h.Service.LogoutAll(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out from all devices"})
}
