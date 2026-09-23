package controller

import (
	"net/http"

	"pulsa2/internal/helper"
)

func (h *ProviderCallbackController) CallbackMultikom(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("multikom", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessMultikomCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("multikom", r, raw)
		status, out := h.svc.ProcessMultikomCallback(r.Context(), string(raw), r.URL.Query())
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackTalenta(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.CallbackTalentaGet(w, r)
	case http.MethodPost:
		h.CallbackTalentaPost(w, r)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackTalentaGet(w http.ResponseWriter, r *http.Request) {
	appendProviderLog("talentapay", r, []byte(r.URL.RawQuery))

	q := r.URL.Query()
	payload := make(map[string]any, len(q))
	for k, vals := range q {
		if len(vals) == 0 {
			continue
		}
		payload[k] = vals[0]
	}

	status, out := h.svc.ProcessTalentaCallback(r.Context(), r.URL.RawQuery, payload)
	helper.WriteJSON(w, status, out)
}

func (h *ProviderCallbackController) CallbackTalentaPost(w http.ResponseWriter, r *http.Request) {
	raw := readProviderCallbackBody(r)
	appendProviderLog("talentapay", r, raw)

	payload, err := decodeJSONMap(raw)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}

	status, out := h.svc.ProcessTalentaCallback(r.Context(), string(raw), payload)
	helper.WriteJSON(w, status, out)
}
