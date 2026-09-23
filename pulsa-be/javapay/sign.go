package javapay

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// tanggal format 2006-01-02 (Asia/Jakarta)
func Sign(memberID, apiKey, pin, refid, product, dest string, t time.Time) (sign string, tanggal string) {
	tanggal = t.Format("2006-01-02")
	raw := memberID + apiKey + pin + refid + product + dest + tanggal
	return md5Hex(raw), tanggal
}
