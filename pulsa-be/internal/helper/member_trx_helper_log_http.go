package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var memberTrxLogMu sync.Mutex
var postJSONClient = &http.Client{Timeout: 12 * time.Second}

func AppendMemberTrxLog(format string, args ...any) {
	appendNamedLog("trxmember.log", format, args...)
}

func AppendAlreadyFinalLog(format string, args ...any) {
	appendNamedLog("already_final.log", format, args...)
}

func appendNamedLog(filename string, format string, args ...any) {
	memberTrxLogMu.Lock()
	defer memberTrxLogMu.Unlock()

	dir := memberTrxLogDir()
	if err := ensureLogDir(dir); err != nil {
		log.Printf("member_trx log: mkdir log dir error: %v", err)
		return
	}

	name := strings.TrimSpace(filename)
	if name == "" {
		name = "trxmember.log"
	}
	file := filepath.Join(dir, name)
	rotateIfNeeded(file)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("member_trx log: open log file error: %v", err)
		return
	}
	defer f.Close()

	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [member_trx] %s\n", time.Now().Format(time.RFC3339Nano), msg)
	if _, err := f.WriteString(line); err != nil {
		log.Printf("member_trx log: write log file error: %v", err)
	}
}

func memberTrxLogDir() string {
	if d := os.Getenv("PULSA_LOG_DIR"); d != "" {
		return d
	}
	return permanentLogDir
}

// ensureLogDir creates dir, handling broken symlinks by removing them first.
func ensureLogDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err == nil {
		return nil
	}
	if fi, lErr := os.Lstat(dir); lErr == nil && fi.Mode()&os.ModeSymlink != 0 {
		_ = os.Remove(dir)
		return os.MkdirAll(dir, 0o755)
	}
	return os.MkdirAll(dir, 0o755)
}

func PostJSON(ctx context.Context, url string, payload any) (int, string, error) {
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := postJSONClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw), nil
}
