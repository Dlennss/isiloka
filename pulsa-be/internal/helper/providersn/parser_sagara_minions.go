package providersn

import "strings"

func ParseSagaraSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}

	if m := reSagaraSN.FindStringSubmatch(msg); len(m) == 2 {
		val := pickSagaraSNValue(strings.TrimSpace(m[1]))
		if val != "" {
			return val, val
		}
	}

	m := reSagaraSNRefLine.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", ""
	}
	seg := strings.TrimSpace(m[1])
	if i := strings.Index(strings.ToUpper(seg), "HARGA"); i >= 0 {
		seg = strings.TrimSpace(seg[:i])
	}
	seg = strings.Trim(seg, " .,:;")
	if seg == "" || seg == "." {
		return "", ""
	}
	if val := pickSagaraSNValue(seg); val != "" {
		return val, val
	}
	return "", ""
}

func ParseMinionsSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}

	if m := reMinionsSN.FindStringSubmatch(msg); len(m) == 2 {
		val := pickSagaraSNValue(strings.TrimSpace(m[1]))
		if val != "" {
			return val, val
		}
	}

	m := reMinionsSNRefLine.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", ""
	}
	seg := strings.TrimSpace(m[1])
	if i := strings.Index(strings.ToUpper(seg), "HARGA"); i >= 0 {
		seg = strings.TrimSpace(seg[:i])
	}
	seg = strings.Trim(seg, " .,:;")
	if seg == "" || seg == "." {
		return "", ""
	}
	if val := pickSagaraSNValue(seg); val != "" {
		return val, val
	}
	return "", ""
}

func pickSagaraSNValue(seg string) string {
	seg = strings.TrimSpace(seg)
	if seg == "" {
		return ""
	}
	if val := normalizeSagaraSNToken(seg); val != "" && !looksLikeSagaraProviderPrefix(val) {
		return val
	}
	parts := strings.Split(seg, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		val := normalizeSagaraSNToken(parts[i])
		if val == "" || looksLikeSagaraProviderPrefix(val) {
			continue
		}
		return val
	}
	return ""
}
