package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var providerCallbackLogMu sync.Mutex

const (
	permanentLogDir = "/home/syarif/app/logs"
	maxLogFileSize  = 100 * 1024 * 1024 // 100MB per file
)

func callbackClientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func callbackLogDir() string {
	if d := os.Getenv("PULSA_LOG_DIR"); d != "" {
		return d
	}
	return permanentLogDir
}

// rotateIfNeeded rotates the log file if it exceeds maxLogFileSize.
// Old file is renamed with timestamp suffix, never deleted.
func rotateIfNeeded(file string) {
	info, err := os.Stat(file)
	if err != nil || info.Size() < maxLogFileSize {
		return
	}
	rotated := file + "." + time.Now().Format("20060102-150405")
	_ = os.Rename(file, rotated)
}

func AppendProviderCallbackLog(provider string, r *http.Request, raw []byte) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		provider = "provider"
	}

	providerCallbackLogMu.Lock()
	defer providerCallbackLogMu.Unlock()

	dir := callbackLogDir()
	if err := ensureLogDir(dir); err != nil {
		log.Printf("%s callback: mkdir log dir error: %v", provider, err)
		return
	}

	file := filepath.Join(dir, fmt.Sprintf("%s_callback.log", provider))
	rotateIfNeeded(file)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("%s callback: open log file error: %v", provider, err)
		return
	}
	defer f.Close()

	entry := map[string]any{
		"provider":         provider,
		"ts":               time.Now().Format(time.RFC3339Nano),
		"method":           r.Method,
		"path":             r.URL.Path,
		"raw_query":        r.URL.RawQuery,
		"remote_addr":      strings.TrimSpace(r.RemoteAddr),
		"client_ip":        callbackClientIP(r),
		"x_forwarded_for":  strings.TrimSpace(r.Header.Get("X-Forwarded-For")),
		"cf_connecting_ip": strings.TrimSpace(r.Header.Get("CF-Connecting-IP")),
		"content_type":     strings.TrimSpace(r.Header.Get("Content-Type")),
		"content_length":   r.ContentLength,
		"raw_body":         string(raw),
	}

	line, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s callback: marshal log entry error: %v", provider, err)
		return
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		log.Printf("%s callback: write log file error: %v", provider, err)
	}
}
