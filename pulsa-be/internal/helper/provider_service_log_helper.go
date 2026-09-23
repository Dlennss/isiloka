package helper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var providerServiceLogMu sync.Mutex

func AppendProviderServiceLog(filename string, format string, args ...any) {
	providerServiceLogMu.Lock()
	defer providerServiceLogMu.Unlock()

	dir := callbackLogDir()
	if err := ensureLogDir(dir); err != nil {
		log.Printf("provider service log: mkdir log dir error: %v", err)
		return
	}

	name := strings.TrimSpace(filename)
	if name == "" {
		name = "provider_service.log"
	}
	file := filepath.Join(dir, name)
	rotateIfNeeded(file)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("provider service log: open log file error: %v", err)
		return
	}
	defer f.Close()

	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [provider_service] %s\n", time.Now().Format(time.RFC3339Nano), msg)
	if _, err := f.WriteString(line); err != nil {
		log.Printf("provider service log: write log file error: %v", err)
	}
}
