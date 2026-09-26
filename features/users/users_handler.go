package users

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"golang-backend/features/audit"
	"golang-backend/features/roles"
	"golang-backend/features/syslog"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// Handler depends on abstractions (mockable), not concretes.
// Audit + Syslog are best-effort (nil-safe) so failures never break requests.
type Handler struct {
	Service ServiceInterface
	Perms   roles.ServiceInterface
	Audit   audit.ServiceInterface
	Sys     syslog.ServiceInterface
}

func NewHandler(service ServiceInterface, perms roles.ServiceInterface, audit audit.ServiceInterface, sys syslog.ServiceInterface) *Handler {
	return &Handler{Service: service, Perms: perms, Audit: audit, Sys: sys}
}

func (h *Handler) audit(r *http.Request, actorCode, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	_ = h.Audit.Log(r.Context(), actorCode, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	if web.ServiceError(w, err) == http.StatusInternalServerError && h.Sys != nil {
		_ = h.Sys.Error(context.Background(), "users_handler", err.Error())
	}
}

// requireMenuOrSelf allows the owner, or anyone holding the menu permission
// (e.g. admin with MENU_USERS can edit other users).
func requireMenuOrSelf(h *Handler, w http.ResponseWriter, r *http.Request, code, menuPermission string) bool {
	callerCode, ok := middleware.GetUserCode(r)
	if !ok {
		web.WriteError(w, http.StatusForbidden, services.ErrForbidden.Error())
		return false
	}
	if callerCode == code {
		return true
	}
	if err := h.Perms.CheckPermission(r.Context(), callerCode, menuPermission); err != nil {
		web.WriteError(w, http.StatusForbidden, services.ErrForbidden.Error())
		return false
	}
	return true
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	web.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := web.Paginate(r)
	users, err := h.Service.ListUsers(r.Context(), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if users == nil {
		users = []models.UserResponse{}
	}
	web.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) GetUserByCode(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid user code")
		return
	}
	user, err := h.Service.GetUserByCode(r.Context(), code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, user.ToResponse())
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	// Hanya pemegang MenuUsers boleh menentukan role non-default.
	roleCode := models.DefaultRoleCode
	if strings.TrimSpace(req.RoleCode) != "" &&
		!strings.EqualFold(strings.TrimSpace(req.RoleCode), models.DefaultRoleCode) {
		caller, _ := middleware.GetUserCode(r)
		if err := h.Perms.CheckPermission(r.Context(), caller, models.MenuUsers); err != nil {
			web.WriteError(w, http.StatusForbidden, services.ErrForbidden.Error())
			return
		}
		roleCode = strings.ToUpper(strings.TrimSpace(req.RoleCode))
	}
	code, err := h.Service.CreateUser(r.Context(), req.Username, req.Email, req.Password, roleCode)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	caller, _ := middleware.GetUserCode(r)
	h.audit(r, caller, models.AuditRegister, models.EntityUser, code, "Akun "+req.Username+" dibuat oleh "+caller)
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"code":    code,
	})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid user code")
		return
	}
	if !requireMenuOrSelf(h, w, r, code, models.MenuUsers) {
		return
	}
	var req models.UpdateUserRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.UpdateUser(r.Context(), code, req); err != nil {
		h.handleServiceError(w, err)
		return
	}
	caller, _ := middleware.GetUserCode(r)
	h.audit(r, caller, models.AuditUpdateUser, models.EntityUser, code, "Profil diperbarui")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid user code")
		return
	}
	if !requireMenuOrSelf(h, w, r, code, models.MenuUsers) {
		return
	}
	if err := h.Service.DeleteUser(r.Context(), code); err != nil {
		h.handleServiceError(w, err)
		return
	}
	caller, _ := middleware.GetUserCode(r)
	h.audit(r, caller, models.AuditDeleteUser, models.EntityUser, code, "Akun dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	accessToken, refreshToken, err := h.Service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		h.audit(r, "", models.AuditLoginFailed, models.EntityAuth, req.Username, "Login gagal: "+req.Username)
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, "", models.AuditLogin, models.EntityAuth, req.Username, "Login berhasil: "+req.Username)
	w.Header().Set("Cache-Control", "no-store")
	web.WriteJSON(w, http.StatusOK, map[string]string{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	accessToken, newRefreshToken, err := h.Service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	web.WriteJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.LogoutRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.Logout(r.Context(), req.RefreshToken); err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, "", models.AuditLogout, models.EntitySession, "", "Sesi dicabut via logout")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid user code")
		return
	}
	if !requireMenuOrSelf(h, w, r, code, models.MenuUsers) {
		return
	}
	if err := h.Service.LogoutAll(r.Context(), code); err != nil {
		h.handleServiceError(w, err)
		return
	}
	caller, _ := middleware.GetUserCode(r)
	h.audit(r, caller, models.AuditLogoutAll, models.EntitySession, code, "Semua sesi dicabut")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Logged out from all devices"})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	code, ok := middleware.GetUserCode(r)
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, "Missing user code")
		return
	}
	roleCode := middleware.GetUserRole(r)
	resp := models.MeResponse{
		Code:     code,
		Role:     roleCode,
		RoleCode: roleCode,
	}
	user, err := h.Service.GetUserByCode(r.Context(), code)
	if err != nil && !errors.Is(err, services.ErrUserNotFound) {
		h.handleServiceError(w, err)
		return
	}
	if user != nil {
		resp.Username = user.Username
		resp.Email = user.Email
		resp.Role = user.RoleCode
		resp.RoleCode = user.RoleCode
		resp.RoleName = user.RoleName
	}
	perms, err := h.Perms.GetRolePermissions(r.Context(), resp.RoleCode)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if perms == nil {
		perms = []string{}
	}
	resp.Permissions = perms
	// Entri menu milik user (sidebar + placeholder 404). Best-effort:
	// kegagalan di sini tidak boleh menggagalkan /me.
	if h.Perms != nil {
		if entries, merr := h.Perms.MyMenus(r.Context(), code); merr == nil {
			if entries == nil {
				entries = []models.MenuEntry{}
			}
			resp.Menus = entries
		}
	}
	if resp.Menus == nil {
		resp.Menus = []models.MenuEntry{}
	}
	web.WriteJSON(w, http.StatusOK, resp)
}
