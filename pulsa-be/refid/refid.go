package refid

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// Generate membuat refid alfanumerik yang konsisten lintas provider.
// Format: <PREFIX><yyyyMMddHHmmss><RANDOM_HEX_8>
func Generate(prefix string) string {
	p := strings.ToUpper(strings.TrimSpace(prefix))
	if p == "" {
		p = "TRX"
	}

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	rnd := strings.ToUpper(hex.EncodeToString(b))

	return p + time.Now().Format("20060102150405") + rnd
}

