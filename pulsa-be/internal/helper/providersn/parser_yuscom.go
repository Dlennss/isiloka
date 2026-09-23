package providersn

import "strings"

func ParseYuscomSNRefFromMsg(msg string) (providerRef string, sn string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ""
	}
	up := strings.ToUpper(msg)
	idx := strings.Index(up, "SN/REF:")
	if idx < 0 {
		return "", ""
	}
	seg := strings.TrimSpace(msg[idx+len("SN/REF:"):])
	upSeg := strings.ToUpper(seg)
	if j := strings.Index(upSeg, "SALDO"); j >= 0 {
		seg = strings.TrimSpace(seg[:j])
	}
	seg = strings.Trim(seg, " .,")
	if seg == "" {
		return "", ""
	}
	if m := reYuscomPLNToken.FindStringSubmatch(seg); len(m) == 2 {
		val := normalizeCommonSNToken(m[1])
		if val != "" {
			return val, val
		}
	}
	if m := reYuscomNoResi.FindStringSubmatch(seg); len(m) == 2 {
		val := normalizeCommonSNToken(m[1])
		if val != "" {
			return val, val
		}
	}
	if m := reYuscomExplicitSN.FindStringSubmatch(seg); len(m) == 2 {
		val := normalizeCommonSNToken(m[1])
		if val != "" {
			return val, val
		}
	}
	parts := strings.Split(seg, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		val := normalizeYuscomSNToken(parts[i])
		if val == "" {
			continue
		}
		return val, val
	}
	return "", ""
}
