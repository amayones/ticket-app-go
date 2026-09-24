package audit

import (
	"net/http"

	"golang-backend/internal/web"
	"golang-backend/models"
)

// Handler melayani menu Audit Log.
//
// Konvensi menu: raw SQL di variabel `query`, di-run, dipetakan ke response
// (lihat audit_repository.go); error service via web.ServiceError.
type Handler struct {
	Service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	logs, err := h.Service.List(r.Context(), models.AuditFilter{
		Action: q.Get("action"), Entity: q.Get("entity"), Actor: q.Get("actor"),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, logs)
}
