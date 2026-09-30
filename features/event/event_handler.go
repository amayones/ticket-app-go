package event

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang-backend/features/audit"
	"golang-backend/features/syslog"
	"golang-backend/internal/export"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/utils"
)

// Handler depends on abstractions (mockable), not concretes.
// Audit + Syslog are best-effort (nil-safe) so failures never break requests.
type Handler struct {
	Service ServiceInterface
	Audit   audit.ServiceInterface
	Sys     syslog.ServiceInterface
}

func NewHandler(service ServiceInterface, audit audit.ServiceInterface, sys syslog.ServiceInterface) *Handler {
	return &Handler{Service: service, Audit: audit, Sys: sys}
}

func (h *Handler) audit(r *http.Request, actorCode, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	_ = h.Audit.Log(r.Context(), actorCode, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	if web.ServiceError(w, err) == http.StatusInternalServerError && h.Sys != nil {
		_ = h.Sys.Error(context.Background(), "event_handler", err.Error())
	}
}

func callerCode(w http.ResponseWriter, r *http.Request) (string, bool) {
	code, ok := middleware.GetUserCode(r)
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, "Missing user code")
		return "", false
	}
	return code, true
}

func (h *Handler) MineList(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	items, err := h.Service.MineList(r.Context(), caller,
		q.Get("status"), q.Get("category"), q.Get("city"),
		strings.TrimSpace(q.Get("from")), strings.TrimSpace(q.Get("to")),
		q.Get("q"), q.Get("sort"), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if items == nil {
		items = []models.MyEventItem{}
	}
	web.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	var in models.EventInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	code, err := h.Service.CreateEvent(r.Context(), caller, in)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventCreate, models.EntityEvent, code, "Event draft dibuat")
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Event created as draft",
		"code":    code,
	})
}

func (h *Handler) EventDetail(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	detail, err := h.Service.EventDetail(r.Context(), caller, code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, detail)
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	var in models.EventInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	notify, err := h.Service.UpdateEvent(r.Context(), caller, code, in)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventUpdate, models.EntityEvent, code, "Event diperbarui")
	resp := map[string]interface{}{"message": "Event updated successfully"}
	if notify {
		resp["warning"] = "Schedule or venue changed with tickets sold: buyers must be notified (other module)"
	}
	web.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	missing, err := h.Service.PublishEvent(r.Context(), caller, code)
	if err != nil {
		// Not-ready carries the checklist for the UI (409 + missing[]).
		if len(missing) > 0 {
			web.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"error":   err.Error(),
				"missing": missing,
			})
			return
		}
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventPublish, models.EntityEvent, code, "Event dipublish")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Event published"})
}

func (h *Handler) UnpublishEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	if err := h.Service.UnpublishEvent(r.Context(), caller, code); err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventUnpublish, models.EntityEvent, code, "Event kembali ke draft")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Event unpublished"})
}

func (h *Handler) CancelEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	var req models.CancelRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	affected, err := h.Service.CancelEvent(r.Context(), caller, code, req.Reason)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventCancel, models.EntityEvent, code,
		fmt.Sprintf("Event dibatalkan (%d pembeli terdampak)", affected))
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "Event cancelled",
		"affected_buyers": affected,
	})
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	if err := h.Service.DeleteEvent(r.Context(), caller, code); err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventDelete, models.EntityEvent, code, "Event draft dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Event deleted"})
}

func (h *Handler) DuplicateEvent(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	newCode, err := h.Service.DuplicateEvent(r.Context(), caller, code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditEventCreate, models.EntityEvent, newCode, "Event diduplikat dari "+code)
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Event duplicated as draft",
		"code":    newCode,
	})
}

func (h *Handler) TypeList(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	types, sum, err := h.Service.TypeList(r.Context(), caller, code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if types == nil {
		types = []models.TicketTypeRow{}
	}
	if sum.Missing == nil {
		sum.Missing = []string{}
	}
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{"types": types, "summary": sum})
}

func (h *Handler) CreateType(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	var in models.TicketTypeInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	typeCode, err := h.Service.CreateType(r.Context(), caller, code, in)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditTicketTypeCreate, models.EntityTicketType, typeCode, "Tipe tiket dibuat")
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Ticket type created",
		"code":    typeCode,
	})
}

func (h *Handler) UpdateType(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	typeCode, ok := web.PathCode(r, "tcode")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ticket type code")
		return
	}
	var in models.TicketTypeInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.Service.UpdateType(r.Context(), caller, code, typeCode, in); err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditTicketTypeUpdate, models.EntityTicketType, typeCode, "Tipe tiket diperbarui")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Ticket type updated"})
}

func (h *Handler) SetTypeActive(w http.ResponseWriter, r *http.Request, active bool) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	typeCode, ok := web.PathCode(r, "tcode")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ticket type code")
		return
	}
	if err := h.Service.SetTypeActive(r.Context(), caller, code, typeCode, active); err != nil {
		h.handleServiceError(w, err)
		return
	}
	action := models.AuditTicketTypeUpdate
	state := "diaktifkan"
	if !active {
		state = "dinonaktifkan"
	}
	h.audit(r, caller, action, models.EntityTicketType, typeCode, "Tipe tiket "+state)
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Ticket type " + state})
}

func (h *Handler) DeleteType(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	typeCode, ok := web.PathCode(r, "tcode")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ticket type code")
		return
	}
	if err := h.Service.DeleteType(r.Context(), caller, code, typeCode); err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditTicketTypeDelete, models.EntityTicketType, typeCode, "Tipe tiket dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Ticket type deleted"})
}

func (h *Handler) ReorderTypes(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	var req models.TicketOrderRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.ReorderTypes(r.Context(), caller, code, req.Order); err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Ticket order updated"})
}

func (h *Handler) Attendees(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	rows, sum, err := h.Service.Attendees(r.Context(), caller, code,
		q.Get("type"), q.Get("attendance"), q.Get("payment"), q.Get("q"), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.AttendeeRow{}
	}
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{"attendees": rows, "summary": sum})
}

func (h *Handler) AttendeeExport(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	q := r.URL.Query()
	if format := strings.ToLower(strings.TrimSpace(q.Get("format"))); format != "" && format != "csv" {
		web.WriteError(w, http.StatusBadRequest, "format must be csv")
		return
	}
	tables, err := h.Service.AttendeeExport(r.Context(), caller, code,
		q.Get("type"), q.Get("attendance"), q.Get("payment"), q.Get("q"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	data, err := export.CSV(tables)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditExport, models.EntityAttendee, code, "Export daftar peserta (csv)")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="attendees-`+code+`.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// UploadPoster menerima multipart poster/banner (maks 2 MB, jpeg/png/webp/gif
// dicek magic bytes, bukan sekadar ekstensi). File disimpan di ./uploads/posters
// dengan nama aman dan disajikan di /uploads/posters/* (lihat main.go wiring).
// Ini endpoint upload pertama di project (belum ada pola upload sebelumnya).
func (h *Handler) UploadPoster(w http.ResponseWriter, r *http.Request) {
	if _, ok := callerCode(w, r); !ok {
		return
	}
	const maxPosterBytes = 2 << 20 // 2 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxPosterBytes+1024)
	if err := r.ParseMultipartForm(maxPosterBytes); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Poster too large (max 2 MB)")
		return
	}
	file, _, err := r.FormFile("poster")
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, "Field poster is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxPosterBytes+1))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if len(data) > maxPosterBytes {
		web.WriteError(w, http.StatusBadRequest, "Poster too large (max 2 MB)")
		return
	}
	ext, ok := posterExt(data)
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Poster must be JPEG, PNG, WebP, or GIF")
		return
	}
	name, err := utils.GenerateCode("PST-")
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	dir := filepath.Join("uploads", "posters")
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.handleServiceError(w, err)
		return
	}
	filename := name + ext
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0644); err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, map[string]string{
		"message": "Poster uploaded",
		"url":     "/uploads/posters/" + filename,
	})
}

// posterExt sniffs magic bytes (never trusts the client filename).
func posterExt(data []byte) (string, bool) {
	if len(data) < 12 {
		return "", false
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return ".jpg", true
	case data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G':
		return ".png", true
	case data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P':
		return ".webp", true
	case data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
		return ".gif", true
	default:
		return "", false
	}
}
