package syslog

import (
	"net/http"
	"strconv"
	"strings"

	"golang-backend/features/audit"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/utils"
)

// Handler melayani menu Error / System Log.
//
// Konvensi menu: raw SQL di variabel `query`, di-run, dipetakan ke response
// (lihat syslog_repository.go); error service via web.ServiceError.
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

func (h *Handler) ListSyslog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	level := strings.ToUpper(strings.TrimSpace(q.Get("level")))
	if level != "" && level != models.SyslogError && level != models.SyslogWarn && level != models.SyslogInfo {
		web.WriteError(w, http.StatusBadRequest, "Invalid level (use ERROR, WARN, INFO)")
		return
	}
	limit, offset := web.Paginate(r)
	logs, err := h.Service.List(r.Context(), models.SyslogFilter{
		Level: level, Limit: limit, Offset: offset,
	})
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	if logs == nil {
		logs = []models.SysLog{}
	}
	web.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) PruneSyslog(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			web.WriteError(w, http.StatusBadRequest, "Invalid days (must be >= 1)")
			return
		}
		days = n
	}
	n, err := h.Service.Prune(r.Context(), days)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditSyslogPrune, models.EntitySystem, "",
		"System log lebih tua dari "+strconv.Itoa(days)+" hari dihapus ("+strconv.FormatInt(n, 10)+" baris)")
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Old system logs pruned",
		"deleted": n,
	})
}
