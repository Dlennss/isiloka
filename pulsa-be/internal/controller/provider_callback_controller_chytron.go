package controller

import (
	"net/http"

	"pulsa2/internal/helper"
)

func (h *ProviderCallbackController) CallbackChytron(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("chytron", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessChytronCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("chytron", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessChytronCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}
