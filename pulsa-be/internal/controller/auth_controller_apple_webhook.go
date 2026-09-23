package controller

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"pulsa2/internal/helper"
)

func (h *AuthController) AppleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	var payload struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[apple_webhook] invalid json: %s", strings.TrimSpace(string(body)))
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	parts := strings.Split(payload.Payload, ".")
	if len(parts) >= 2 {
		decoded, err := appleBase64URLDecode(parts[1])
		if err == nil {
			var event struct {
				Type    string `json:"type"`
				Sub     string `json:"sub"`
				EventAt int64  `json:"event_time"`
				Email   string `json:"email"`
			}
			if err := json.Unmarshal(decoded, &event); err == nil {
				log.Printf("[apple_webhook] event type=%s sub=%s email=%s", event.Type, event.Sub, event.Email)

				switch event.Type {
				case "consent-revoked", "account-delete":
					if event.Sub != "" {
						_ = h.svc.DeactivateAppleMember(r.Context(), event.Sub)
						log.Printf("[apple_webhook] deactivated apple_sub=%s reason=%s", event.Sub, event.Type)
					}
				}
			}
		}
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func appleBase64URLDecode(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.StdEncoding.DecodeString(s)
}
