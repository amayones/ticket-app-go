package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/services"
)

// AdminHandler melayani menu admin: role, permission, sesi, audit,
// security center, system log, dan notifikasi.
type AdminHandler struct {
	Service services.UserServiceInterface
	Audit   AuditReader
	Sys     SyslogAdmin
	Notif   Notifier
}

func NewAdminHandler(svc services.UserServiceInterface, audit AuditReader, sys SyslogAdmin, notif Notifier) *AdminHandler {
	return &AdminHandler{Service: svc, Audit: audit, Sys: sys, Notif: notif}
}

func (h *AdminHandler) actor(r *http.Request) string {
	code, _ := middleware.GetUserCode(r)
	return code
}

func (h *AdminHandler) ip(r *http.Request) string {
	return services.ClientIP(r.RemoteAddr)
}

func (h *AdminHandler) audit(r *http.Request, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	_ = h.Audit.Log(r.Context(), h.actor(r), action, entity, entityCode, detail, h.ip(r))
}

// --- Role & Permission (RBAC) ---

func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	role, err := h.Service.CreateRole(r.Context(), req.Code, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditRoleCreate, models.EntityRole, role.Code, "Role "+role.Code+" dibuat")
	writeJSON(w, http.StatusCreated, role)
}

func (h *AdminHandler) GetRoleDetail(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	detail, err := h.Service.GetRoleDetail(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *AdminHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if err := h.Service.DeleteRole(r.Context(), code); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditRoleDelete, models.EntityRole, code, "Role "+code+" dihapus")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

func (h *AdminHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.Service.ListPermissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, perms)
}

func (h *AdminHandler) SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	var req struct {
		Permissions []string `json:"permissions"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Permissions == nil {
		req.Permissions = []string{}
	}
	if err := h.Service.SetRolePermissions(r.Context(), code, req.Permissions); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditPermissionAssign, models.EntityRole, code,
		"Permission role "+code+" diatur ulang ("+strconv.Itoa(len(req.Permissions))+" item)")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Role permissions updated"})
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	code, ok := parseCode(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Invalid user code")
		return
	}
	var req struct {
		RoleCode string `json:"role_code"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.UpdateUserRole(r.Context(), code, req.RoleCode); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditRoleAssign, models.EntityUser, code, "Role diubah ke "+strings.ToUpper(strings.TrimSpace(req.RoleCode)))
	writeJSON(w, http.StatusOK, map[string]string{"message": "User role updated"})
}

// --- Authentication & Session Management ---

func (h *AdminHandler) ListMySessions(w http.ResponseWriter, r *http.Request) {
	code, _ := middleware.GetUserCode(r)
	sessions, err := h.Service.ListSessions(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *AdminHandler) ListAllSessions(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginate(r)
	sessions, err := h.Service.ListAllSessions(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *AdminHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}
	caller, _ := middleware.GetUserCode(r)
	manageAll := h.Service.CheckPermission(r.Context(), caller, models.PermSessionManage) == nil
	if err := h.Service.RevokeSession(r.Context(), caller, id, manageAll); err != nil {
		status := http.StatusBadRequest
		if err == services.ErrForbidden {
			status = http.StatusForbidden
		}
		writeError(w, status, err.Error())
		return
	}
	h.audit(r, models.AuditSessionRevoke, models.EntitySession, strconv.Itoa(id), "Sesi dicabut")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Session revoked"})
}

// --- Audit Log ---

func (h *AdminHandler) ListAudit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := paginate(r)
	logs, err := h.Audit.List(r.Context(), models.AuditFilter{
		Action: q.Get("action"), Entity: q.Get("entity"), Actor: q.Get("actor"),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// --- Security Center ---

func (h *AdminHandler) SecuritySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, _ := h.Service.CountUsers(ctx)
	roles, _ := h.Service.CountRoles(ctx)
	sessions, _ := h.Service.CountActiveSessions(ctx)
	audit24, _ := h.Audit.CountSince(ctx, 24)
	errors24, _ := h.Sys.CountSince(ctx, 24, models.SyslogError)
	writeJSON(w, http.StatusOK, map[string]int{
		"total_users": users, "total_roles": roles, "active_sessions": sessions,
		"audit_last_24h": audit24, "errors_last_24h": errors24,
	})
}

// --- Error / System Log ---

func (h *AdminHandler) ListSyslog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := paginate(r)
	logs, err := h.Sys.List(r.Context(), models.SyslogFilter{
		Level: strings.ToUpper(q.Get("level")), Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *AdminHandler) PruneSyslog(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	n, err := h.Sys.Prune(r.Context(), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	h.audit(r, models.AuditSyslogPrune, models.EntitySystem, "",
		"System log lebih tua dari "+strconv.Itoa(days)+" hari dihapus ("+strconv.FormatInt(n, 10)+" baris)")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Old system logs pruned",
		"deleted": n,
	})
}

// --- Notification Template & Log ---

func (h *AdminHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "1"
	tmpls, err := h.Notif.ListTemplates(r.Context(), activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, tmpls)
}

func (h *AdminHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Active  *bool  `json:"is_active"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	tmpl, err := h.Notif.CreateTemplate(r.Context(), req.Name, req.Channel, req.Subject, req.Body, active)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditTemplateCreate, models.EntityTemplate, tmpl.Code, "Template "+tmpl.Name+" dibuat")
	writeJSON(w, http.StatusCreated, tmpl)
}

func (h *AdminHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	var req struct {
		Name    string `json:"name"`
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Active  *bool  `json:"is_active"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	if err := h.Notif.UpdateTemplate(r.Context(), code, req.Name, req.Channel, req.Subject, req.Body, active); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditTemplateUpdate, models.EntityTemplate, code, "Template diperbarui")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Template updated"})
}

func (h *AdminHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if err := h.Notif.DeleteTemplate(r.Context(), code); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditTemplateDelete, models.EntityTemplate, code, "Template dihapus")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Template deleted"})
}

func (h *AdminHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req models.NotifSendRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	entry, err := h.Notif.Send(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.audit(r, models.AuditNotifSend, models.EntityNotification, entry.Code,
		"Notifikasi "+entry.Channel+" ke "+entry.Recipient)
	writeJSON(w, http.StatusCreated, entry)
}

func (h *AdminHandler) ListNotifLogs(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginate(r)
	logs, err := h.Notif.ListLogs(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}
