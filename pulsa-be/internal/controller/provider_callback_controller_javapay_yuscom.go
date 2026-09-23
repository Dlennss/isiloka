package controller

import (
	"net/http"

	"pulsa2/internal/helper"
)

func (h *ProviderCallbackController) CallbackJavapay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	raw := readProviderCallbackBody(r)
	appendProviderLog("javapay", r, raw)
	if !h.svc.VerifyJavapayCallbackSignature(raw, r.Header.Get("X-Signature")) {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"status": false, "message": "Invalid signature"})
		return
	}

	payload, err := decodeJSONMap(raw)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": false, "message": "Invalid JSON"})
		return
	}
	status, out := h.svc.ProcessJavapayCallback(r.Context(), r.URL.RawQuery, payload)
	if status >= 200 && status < 300 {
		helper.WriteJSON(w, http.StatusOK, map[string]any{"status": true, "message": "Successful"})
		return
	}
	msg, _ := out["error"].(string)
	if msg == "" {
		msg = "Failed"
	}
	helper.WriteJSON(w, status, map[string]any{"status": false, "message": msg})
}

func (h *ProviderCallbackController) CallbackYuscom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	appendProviderLog("yuscom", r, []byte(r.URL.RawQuery))
	status, out := h.svc.ProcessYuscomCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
	helper.WriteJSON(w, status, out)
}
