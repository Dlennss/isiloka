package helper

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// RequireInternalSecret validates X-Internal-Secret header against ADMIN_TOKEN env var.
func RequireInternalSecret(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("ADMIN_TOKEN")
		if secret == "" {
			http.Error(w, `{"ok":false,"error":"internal auth not configured"}`, http.StatusServiceUnavailable)
			return
		}
		provided := r.Header.Get("X-Internal-Secret")
		if provided == "" {
			provided = r.Header.Get("Authorization")
			if len(provided) > 7 && provided[:7] == "Bearer " {
				provided = provided[7:]
			}
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			http.Error(w, `{"ok":false,"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
