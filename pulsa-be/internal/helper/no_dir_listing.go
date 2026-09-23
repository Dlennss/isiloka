package helper

import (
	"net/http"
	"strings"
)

// NoDirListing wraps an http.FileServer to prevent directory listing.
// Returns 404 for directory requests (paths ending with /).
func NoDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
