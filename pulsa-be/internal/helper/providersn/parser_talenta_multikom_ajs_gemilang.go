package providersn

import "strings"

func ParseTalentaSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}
	m := reTalentaSNRefLine.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", ""
	}

	seg := trimTalentaSNRefSegment(strings.TrimSpace(m[1]))
	if seg == "" {
		return "", ""
	}
	if m := reYuscomExplicitSN.FindStringSubmatch(seg); len(m) == 2 {
		val := normalizeCommonSNToken(m[1])
		if val != "" {
			return val, val
		}
	}
	parts := strings.Split(seg, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		val := normalizeTalentaSNToken(parts[i])
		if val == "" {
			continue
		}
		return val, val
	}
	return "", ""
}

func ParseMultikomSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}
	m := reMultikomSN.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", ""
	}
	seg := strings.Trim(strings.TrimSpace(m[1]), " .,:;")
	if seg == "" {
		return "", ""
	}
	parts := strings.Split(seg, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		val := normalizeCommonSNToken(parts[i])
		if val == "" {
			continue
		}
		return val, val
	}
	return "", ""
}

func ParseAJSSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}
	m := reAJSSN.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", ""
	}
	seg := strings.Trim(strings.TrimSpace(m[1]), " .,:;")
	if seg == "" {
		return "", ""
	}
	parts := strings.Split(seg, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		val := normalizeCommonSNToken(parts[i])
		if val == "" {
			continue
		}
		return val, val
	}
	return "", ""
}

func ParseGemilangSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}
	m := reTalentaSNRefLine.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", ""
	}
	seg := strings.Trim(strings.TrimSpace(m[1]), " .,:;")
	if i := strings.Index(strings.ToUpper(seg), "STOK"); i >= 0 {
		seg = strings.TrimSpace(seg[:i])
	}
	if i := strings.Index(seg, "@"); i >= 0 {
		seg = strings.TrimSpace(seg[:i])
	}
	if seg == "" {
		return "", ""
	}
	seg = strings.Trim(seg, " .,:;")
	if seg == "" {
		return "", ""
	}
	if m := reYuscomExplicitSN.FindStringSubmatch(seg); len(m) == 2 {
		val := normalizeCommonSNToken(m[1])
		if val != "" {
			return val, val
		}
	}
	parts := strings.Split(seg, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		val := normalizeGemilangSNToken(parts[i])
		if val == "" {
			continue
		}
		return val, val
	}
	val := normalizeGemilangSNToken(seg)
	return val, val
}
