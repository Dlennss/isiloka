package helper

import (
	"net/http"
	"os"
	"strings"
)

// CORS adds Access-Control headers. Allowed origins from CORS_ORIGINS env (comma-separated).
// If env is empty, defaults to allowing pulsakilat.net and pulsakilat.org.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-Token, X-Internal-Secret, X-Api-Key, X-Kantor24-Key")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	envOrigins := os.Getenv("CORS_ORIGINS")
	if envOrigins == "" {
		envOrigins = "https://pulsakilat.local,https://pulsakilat.org,https://pulsakilat.local,http://localhost:3000,http://localhost:3003"
	}
	origin = strings.TrimRight(strings.ToLower(strings.TrimSpace(origin)), "/")
	for _, allowed := range strings.Split(envOrigins, ",") {
		allowed = strings.TrimRight(strings.ToLower(strings.TrimSpace(allowed)), "/")
		if allowed == origin {
			return true
		}
	}
	return false
}
