package notifications

import (
	"net/http"

	"golang-backend/features/audit"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/utils"
)

// Handler melayani menu Notification Template & Log.
//
// Konvensi menu: raw SQL di variabel `query`, di-run, dipetakan ke response
// (lihat notifications_repository.go); error service via web.ServiceError.
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

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "1"
	tmpls, err := h.Service.ListTemplates(r.Context(), activeOnly)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, tmpls)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Active  *bool  `json:"is_active"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	tmpl, err := h.Service.CreateTemplate(r.Context(), req.Name, req.Channel, req.Subject, req.Body, active)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditTemplateCreate, models.EntityTemplate, tmpl.Code, "Template "+tmpl.Name+" dibuat")
	web.WriteJSON(w, http.StatusCreated, tmpl)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid template code")
		return
	}
	var req struct {
		Name    string `json:"name"`
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Active  *bool  `json:"is_active"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	if err := h.Service.UpdateTemplate(r.Context(), code, req.Name, req.Channel, req.Subject, req.Body, active); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditTemplateUpdate, models.EntityTemplate, code, "Template diperbarui")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Template updated"})
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid template code")
		return
	}
	if err := h.Service.DeleteTemplate(r.Context(), code); err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditTemplateDelete, models.EntityTemplate, code, "Template dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Template deleted"})
}

func (h *Handler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req models.NotifSendRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	entry, err := h.Service.Send(r.Context(), req)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditNotifSend, models.EntityNotification, entry.Code,
		"Notifikasi "+entry.Channel+" ke "+entry.Recipient)
	web.WriteJSON(w, http.StatusCreated, entry)
}

func (h *Handler) ListNotifLogs(w http.ResponseWriter, r *http.Request) {
	limit, offset := web.Paginate(r)
	logs, err := h.Service.ListLogs(r.Context(), limit, offset)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, logs)
}
