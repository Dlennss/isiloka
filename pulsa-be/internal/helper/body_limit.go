package helper

import (
	"net/http"
)

// MaxBodySize limits request body to maxBytes for all non-upload endpoints.
func MaxBodySize(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next.ServeHTTP(w, r)
	})
}
