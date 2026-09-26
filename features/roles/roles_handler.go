package roles

import (
	"fmt"
	"net/http"

	"golang-backend/features/audit"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
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
	if roles == nil {
		roles = []models.Role{}
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
	deletedUsers, err := h.Service.DeleteRole(r.Context(), code)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	detail := "Role " + code + " dihapus"
	if deletedUsers > 0 {
		detail += fmt.Sprintf(" (+%d user ikut terhapus)", deletedUsers)
	}
	h.audit(r, models.AuditRoleDelete, models.EntityRole, code, detail)
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"message":       "Role deleted successfully",
		"deleted_users": deletedUsers,
	})
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.Service.ListPermissions(r.Context())
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	if perms == nil {
		perms = []models.Permission{}
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
	if len(req.Permissions) > 100 {
		web.WriteError(w, http.StatusBadRequest, "Too many permissions (max 100)")
		return
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

// --- Registry menu (CPMENU + CPMATRIX) ----------------------------------------

func (h *Handler) ListMenus(w http.ResponseWriter, r *http.Request) {
	menus, err := h.Service.ListMenus(r.Context())
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	if menus == nil {
		menus = []models.Menu{}
	}
	web.WriteJSON(w, http.StatusOK, menus)
}

func (h *Handler) GetMatrix(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	rows, err := h.Service.GetMatrix(r.Context(), role)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.MatrixRow{}
	}
	web.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) CreateMenu(w http.ResponseWriter, r *http.Request) {
	var req models.MenuInput
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	menu, err := h.Service.CreateMenu(r.Context(), req)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditMenuCreate, models.EntitySystem, menu.Code,
		"Menu "+menu.Code+" dibuat di modul "+menu.MControl+" (tanpa auto-grant role)")
	web.WriteJSON(w, http.StatusCreated, menu)
}

// UpdateMenu mengubah label/urutan/modul/parent menu (kode permission
// tidak bisa diubah - buat menu baru bila perlu).
func (h *Handler) UpdateMenu(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid menu code")
		return
	}
	var req models.MenuUpdateInput
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	menu, err := h.Service.UpdateMenu(r.Context(), code, req)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditMenuUpdate, models.EntitySystem, menu.Code,
		"Menu "+menu.Code+" diperbarui (label/urutan/modul/parent)")
	web.WriteJSON(w, http.StatusOK, menu)
}

func (h *Handler) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid menu code")
		return
	}
	if err := h.Service.DeleteMenu(r.Context(), code); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditMenuDelete, models.EntitySystem, code, "Menu "+code+" dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Menu deleted successfully"})
}

// MyMenus mengembalikan menu milik user login (sidebar + placeholder 404).
func (h *Handler) MyMenus(w http.ResponseWriter, r *http.Request) {
	code, ok := middleware.GetUserCode(r)
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, "Missing user code")
		return
	}
	entries, err := h.Service.MyMenus(r.Context(), code)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	if entries == nil {
		entries = []models.MenuEntry{}
	}
	web.WriteJSON(w, http.StatusOK, entries)
}

// --- Master modul (CPMATRIX tabel) ---------------------------------------------

func (h *Handler) ListModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.Service.ListModules(r.Context())
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	if modules == nil {
		modules = []models.Module{}
	}
	web.WriteJSON(w, http.StatusOK, modules)
}

func (h *Handler) CreateModule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code      string `json:"code"`
		Label     string `json:"label"`
		SortOrder int    `json:"sort_order"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	m, err := h.Service.CreateModule(r.Context(), req.Code, req.Label, req.SortOrder)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditModuleCreate, models.EntitySystem, m.Code, "Modul "+m.Code+" dibuat")
	web.WriteJSON(w, http.StatusCreated, m)
}

func (h *Handler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid module code")
		return
	}
	if err := h.Service.DeleteModule(r.Context(), code); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditModuleDelete, models.EntitySystem, code, "Modul "+code+" dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Module deleted successfully"})
}
