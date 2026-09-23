package helper

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

func RequireKantor24APIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := strings.TrimSpace(os.Getenv("KANTOR24_API_KEY"))
		if secret == "" {
			WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "kantor24 api key not configured"})
			return
		}
		provided := strings.TrimSpace(r.Header.Get("X-Kantor24-Key"))
		if provided == "" {
			provided = strings.TrimSpace(r.Header.Get("X-Api-Key"))
		}
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		next(w, r)
	}
}
