package security

import (
	"net/http"

	"golang-backend/internal/web"
)

// Handler melayani menu Security Center.
//
// Konvensi menu: agregasi via security.Service; error via web.ServiceError.
type Handler struct {
	Service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) SecuritySummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.Service.Summary(r.Context())
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, summary)
}
