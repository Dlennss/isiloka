package providersn

import "strings"

func trimTalentaSNRefSegment(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	cutMarkers := []string{" saldo ", ",saldo", " saldo rp", " hrg ", ",hrg", " @"}
	lower := strings.ToLower(s)
	cut := len(s)
	for _, marker := range cutMarkers {
		if idx := strings.Index(lower, marker); idx >= 0 && idx < cut {
			cut = idx
		}
	}
	s = strings.TrimSpace(s[:cut])
	return strings.TrimRight(s, " .,;")
}

func normalizeTalentaSNToken(s string) string {
	s = strings.TrimSpace(s)
	up := strings.ToUpper(strings.TrimSpace(s))
	if s == "" || isTalentaNoiseToken(s) {
		return ""
	}
	switch {
	case strings.HasPrefix(up, "NAMA:"),
		strings.HasPrefix(up, "NOMINAL:"),
		up == "NAMA",
		up == "NOMINAL",
		up == "TOPUP",
		up == "DANA",
		up == "GOPAY",
		up == "OVO",
		up == "LINKAJA",
		up == "SHOPEE":
		return ""
	}
	return normalizeCommonSNToken(s)
}

func normalizeSagaraSNToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, " .,:;")
	if s == "" || s == "." {
		return ""
	}
	up := strings.ToUpper(s)
	if strings.HasPrefix(up, "SN:") {
		if i := strings.Index(s, ":"); i >= 0 && i+1 < len(s) {
			s = strings.TrimSpace(s[i+1:])
		}
	}
	up = strings.ToUpper(strings.TrimSpace(s))
	switch {
	case up == "TRANSFER BERHASIL",
		strings.HasPrefix(up, "NAMA NASABAH:"),
		strings.HasPrefix(up, "NOMINAL:"),
		up == "-",
		up == "N/A",
		up == "NO":
		return ""
	}
	val := normalizeCommonSNToken(s)
	if val == "" {
		return ""
	}
	if !strings.ContainsAny(val, "0123456789") && len(val) < 10 {
		return ""
	}
	return val
}

func looksLikeSagaraProviderPrefix(s string) bool {
	up := strings.ToUpper(strings.TrimSpace(s))
	switch up {
	case "DNID", "OVO", "DANA", "GOPAY", "LINKAJA", "SHOPEE":
		return true
	default:
		return false
	}
}

func isTalentaNoiseToken(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return true
	}
	switch s {
	case "pengirim", "berita":
		return true
	}
	return strings.HasPrefix(s, "berita:")
}

func normalizeYuscomSNToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, " .,:;")
	if s == "" {
		return ""
	}
	up := strings.ToUpper(s)
	switch {
	case strings.HasPrefix(up, "NO REFERENSI:"),
		strings.HasPrefix(up, "REFF:"),
		strings.HasPrefix(up, "REF:"):
		if i := strings.Index(s, ":"); i >= 0 && i+1 < len(s) {
			s = strings.TrimSpace(s[i+1:])
		}
	}
	return normalizeCommonSNToken(s)
}

func normalizeCommonSNToken(s string) string {
	s = strings.TrimSpace(s)
	up := strings.ToUpper(s)
	if strings.HasPrefix(up, "NO REFERENSI:") || strings.HasPrefix(up, "IDT:") || strings.HasPrefix(up, "REFF:") || strings.HasPrefix(up, "REF:") || strings.HasPrefix(up, "SERIAL:") {
		if i := strings.Index(s, ":"); i >= 0 && i+1 < len(s) {
			s = strings.TrimSpace(s[i+1:])
		}
	}
	s = strings.Trim(s, " .,:;")
	if s == "" {
		return ""
	}
	i := 0
	for i < len(s) {
		c := s[i]
		isNum := c >= '0' && c <= '9'
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'
		isDash := c == '-'
		if !isNum && !isUpper && !isLower && !isDash {
			break
		}
		i++
	}
	return strings.TrimSpace(s[:i])
}

func normalizeGemilangSNToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, " .,:;")
	if s == "" || s == "." {
		return ""
	}
	up := strings.ToUpper(s)
	switch {
	case strings.HasPrefix(up, "NAMA:"),
		strings.HasPrefix(up, "NOMINAL:"),
		up == "NAMA",
		up == "NOMINAL",
		up == "TOPUP",
		up == "DANA",
		up == "GOPAY",
		up == "OVO",
		up == "LINKAJA",
		up == "SHOPEE",
		up == "DNID",
		up == "IDT",
		up == "TRANSFER",
		up == "BERHASIL":
		return ""
	}
	val := normalizeCommonSNToken(s)
	if val == "" {
		return ""
	}
	if !strings.ContainsAny(val, "0123456789") && len(val) < 10 {
		return ""
	}
	return val
}
