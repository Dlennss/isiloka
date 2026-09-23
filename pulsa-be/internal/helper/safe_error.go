package helper

import (
	"log"
	"net/http"
	"strings"
)

// SafeErrorResponse writes a generic error to client and logs the real error server-side.
func SafeErrorResponse(w http.ResponseWriter, status int, publicMsg string, internalErr error, context string) {
	if internalErr != nil {
		log.Printf("[ERROR] %s: %v", context, internalErr)
	}
	msg := strings.TrimSpace(publicMsg)
	if msg == "" {
		msg = "internal error"
	}
	WriteJSON(w, status, map[string]any{"ok": false, "error": msg})
}
