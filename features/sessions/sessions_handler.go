package sessions

import (
	"net/http"
	"strconv"

	"golang-backend/features/audit"
	"golang-backend/features/roles"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// Handler melayani menu Authentication & Session Management.
//
// Konvensi menu: raw SQL di variabel `query`, di-run, dipetakan ke response
// (lihat sessions_repository.go); error service via web.ServiceError.
type Handler struct {
	Service ServiceInterface
	Perms   roles.ServiceInterface
	Audit   audit.ServiceInterface
}

func NewHandler(service ServiceInterface, perms roles.ServiceInterface, audit audit.ServiceInterface) *Handler {
	return &Handler{Service: service, Perms: perms, Audit: audit}
}

func (h *Handler) audit(r *http.Request, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	code, _ := middleware.GetUserCode(r)
	_ = h.Audit.Log(r.Context(), code, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) ListMySessions(w http.ResponseWriter, r *http.Request) {
	code, _ := middleware.GetUserCode(r)
	sessions, err := h.Service.ListSessions(r.Context(), code)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) ListAllSessions(w http.ResponseWriter, r *http.Request) {
	limit, offset := web.Paginate(r)
	sessions, err := h.Service.ListAllSessions(r.Context(), limit, offset)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	rawID, ok := web.PathCode(r, "id")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		web.WriteError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}
	caller, _ := middleware.GetUserCode(r)
	manageAll := h.Perms.CheckPermission(r.Context(), caller, models.MenuSessions) == nil
	if err := h.Service.RevokeSession(r.Context(), caller, id, manageAll); err != nil {
		if err == services.ErrForbidden {
			web.WriteError(w, http.StatusForbidden, err.Error())
			return
		}
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditSessionRevoke, models.EntitySession, strconv.Itoa(id), "Sesi dicabut")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Session revoked"})
}
