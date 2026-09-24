package roles

import (
	"net/http"

	"golang-backend/features/audit"
	"golang-backend/internal/web"
	"golang-backend/models"
	"golang-backend/middleware"
	"golang-backend/utils"
)

// Handler melayani menu Role & Permission (RBAC).
//
// Konvensi menu: raw SQL di variabel `query`, di-run, dipetakan ke response
// (lihat roles_repository.go); error service via web.ServiceError.
type Handler struct {
	Service ServiceInterface
	Audit   audit.ServiceInterface
}

func NewHandler(service ServiceInterface, audit audit.ServiceInterface) *Handler {
	return &Handler{Service: service, Audit: audit}
}

func (h *Handler) audit(r *http.Request, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	code, _ := middleware.GetUserCode(r)
	_ = h.Audit.Log(r.Context(), code, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.Service.ListRoles(r.Context())
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	role, err := h.Service.CreateRole(r.Context(), req.Code, req.Name)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditRoleCreate, models.EntityRole, role.Code, "Role "+role.Code+" dibuat")
	web.WriteJSON(w, http.StatusCreated, role)
}

func (h *Handler) GetRoleDetail(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid role code")
		return
	}
	detail, err := h.Service.GetRoleDetail(r.Context(), code)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, detail)
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid role code")
		return
	}
	if err := h.Service.DeleteRole(r.Context(), code); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditRoleDelete, models.EntityRole, code, "Role "+code+" dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.Service.ListPermissions(r.Context())
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, perms)
}

func (h *Handler) SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid role code")
		return
	}
	var req struct {
		Permissions []string `json:"permissions"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if req.Permissions == nil {
		req.Permissions = []string{}
	}
	if err := h.Service.SetRolePermissions(r.Context(), code, req.Permissions); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditPermissionAssign, models.EntityRole, code, "Permission role diatur ulang")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Role permissions updated"})
}

func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid user code")
		return
	}
	var req struct {
		RoleCode string `json:"role_code"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.UpdateUserRole(r.Context(), code, req.RoleCode); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditRoleAssign, models.EntityUser, code, "Role diubah ke "+req.RoleCode)
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "User role updated"})
}
