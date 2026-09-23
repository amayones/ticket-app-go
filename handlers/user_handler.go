package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"golang-backend/middleware"
	"golang-backend/services"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrUsernameRequired),
		errors.Is(err, services.ErrUsernameTooShort),
		errors.Is(err, services.ErrInvalidEmail),
		errors.Is(err, services.ErrPasswordTooShort):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, services.ErrUsernameTaken),
		errors.Is(err, services.ErrEmailTaken):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, services.ErrInvalidLogin),
		errors.Is(err, services.ErrInvalidRefresh):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, services.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, services.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		log.Println("Internal error:", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.GetAllUsers()
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := h.Service.GetUserByID(id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	err := h.Service.CreateUser(request.Username, request.Email, request.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "User created successfully",
	})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	callerID, ok := middleware.GetUserID(r)
	if !ok || callerID != id {
		http.Error(w, services.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	var request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateUser(id, request.Username, request.Email, request.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "User updated successfully",
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	callerID, ok := middleware.GetUserID(r)
	if !ok || callerID != id {
		http.Error(w, services.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if err := h.Service.DeleteUser(id); err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	accessToken, refreshToken, err := h.Service.Login(request.Username, request.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	accessToken, newRefreshToken, err := h.Service.RefreshAccessToken(request.RefreshToken)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.Service.Logout(request.RefreshToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
