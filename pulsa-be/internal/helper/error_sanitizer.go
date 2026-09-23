package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorSanitizer wraps response writer to intercept JSON error responses
// and replace internal error details with generic messages.
type errorSanitizer struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func SanitizeErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for webhook/callback endpoints (internal)
		path := strings.ToLower(r.URL.Path)
		if strings.Contains(path, "/webhook/") || strings.Contains(path, "/callback") ||
			strings.Contains(path, "/internal/") || strings.Contains(path, "/health") {
			next.ServeHTTP(w, r)
			return
		}

		sw := &errorSanitizer{
			ResponseWriter: w,
			buf:            &bytes.Buffer{},
			statusCode:     200,
		}
		next.ServeHTTP(sw, r)

		body := sw.buf.Bytes()

		// Only sanitize error responses (4xx, 5xx)
		if sw.statusCode >= 400 && len(body) > 0 {
			body = sanitizeErrorBody(body)
		}

		w.WriteHeader(sw.statusCode)
		w.Write(body)
	})
}

func (s *errorSanitizer) WriteHeader(code int) {
	s.statusCode = code
}

func (s *errorSanitizer) Write(b []byte) (int, error) {
	return s.buf.Write(b)
}

func sanitizeErrorBody(body []byte) []byte {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return body // not JSON, pass through
	}

	errField, ok := resp["error"]
	if !ok {
		return body
	}

	errStr, ok := errField.(string)
	if !ok {
		return body
	}

	// Sanitize known internal error patterns
	lower := strings.ToLower(errStr)
	if strings.Contains(lower, "sql:") ||
		strings.Contains(lower, "pq:") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "context canceled") ||
		strings.Contains(lower, "no rows") ||
		strings.Contains(lower, "database") ||
		strings.Contains(lower, "pool") ||
		strings.Contains(lower, "dial tcp") ||
		strings.Contains(lower, "EOF") ||
		strings.Contains(lower, "broken pipe") {
		resp["error"] = "internal error"
		sanitized, _ := json.Marshal(resp)
		return sanitized
	}

	return body
}
