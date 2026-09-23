package controller

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"pulsa2/internal/helper"
)

func (h *ProviderCallbackController) CallbackPulsa24Jam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	if !validPulsa24JamCallbackToken(r) {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"status": false, "message": "Invalid callback token"})
		return
	}

	raw := []byte(r.URL.RawQuery)
	if r.Method == http.MethodPost {
		raw = readProviderCallbackBody(r)
	}
	appendProviderLog("pulsa24jam", r, raw)

	payload := map[string]any{}
	if len(raw) > 0 && strings.TrimSpace(string(raw)) != "" {
		_ = json.Unmarshal(raw, &payload)
	}
	status, out := h.svc.ProcessPulsa24JamCallback(r.Context(), string(raw), r.URL.Query(), payload)
	helper.WriteJSON(w, status, out)
}

func validPulsa24JamCallbackToken(r *http.Request) bool {
	expected := strings.TrimSpace(os.Getenv("PULSA24JAM_CALLBACK_TOKEN"))
	if expected == "" {
		return true
	}

	got := strings.TrimSpace(r.Header.Get("X-Callback-Token"))
	if got == "" {
		got = strings.TrimSpace(r.Header.Get("X-Pulsa24Jam-Token"))
	}
	if got == "" {
		got = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}
