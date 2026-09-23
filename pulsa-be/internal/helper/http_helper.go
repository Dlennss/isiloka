package helper

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var queryDateLocation = time.FixedZone("Asia/Jakarta", 7*60*60)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func ParseIDFromPath(path string, base string) (int64, bool) {
	prefix := base + "/"
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if rest == "" || strings.Contains(rest, "/") {
		return 0, false
	}
	id, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func QueryString(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

func QueryInt(r *http.Request, key string, def int) int {
	v := QueryString(r, key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func QueryInt64(r *http.Request, key string, def int64) int64 {
	v := QueryString(r, key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func QueryDate(r *http.Request, key string) (time.Time, bool) {
	v := QueryString(r, key)
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", v, queryDateLocation)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
