package helper

import (
	"pulsa2/internal/helper/providersn"
	"regexp"
	"strconv"
	"strings"
)

type ProviderInfo struct {
	Reff string
	SN   string
}

var (
	reReff       = regexp.MustCompile(`(?i)\breff[:=]\s*([0-9]{10,})\b`)
	reSN         = regexp.MustCompile(`(?i)\bSN\s*=\s*([^\.\r\n]+)`)
	reSaldo      = regexp.MustCompile(`(?i)\bsaldo\b[^=]*=\s*([0-9][0-9\.\,]{1,})\b`)
	reSaldoAkhir = regexp.MustCompile(`(?i)\bsaldo\s+akhir\b[^0-9]*([0-9][0-9\.\,]{1,})`)
	reSaldoLabel = regexp.MustCompile(`(?i)\b(?:saldo|stok(?:\s+pulsa)?)\s*[:=]\s*([0-9][0-9\.\,]{1,})\b`)
	rePoin       = regexp.MustCompile(`(?i)\bpoin\s+anda\b[^0-9]*([0-9][0-9\.\,]{1,})`)
)

func extractProviderInfo(msg string) ProviderInfo {
	out := ProviderInfo{}
	if m := reReff.FindStringSubmatch(msg); len(m) == 2 {
		out.Reff = strings.TrimSpace(m[1])
	}
	if m := reSN.FindStringSubmatch(msg); len(m) == 2 {
		out.SN = strings.TrimRight(strings.TrimSpace(m[1]), " ")
	}
	if out.Reff == "" || out.SN == "" {
		for _, parser := range []func(string) (string, string){
			providersn.ParseTalentaSNRefFromMsg,
			providersn.ParseYuscomSNRefFromMsg,
			providersn.ParseMultikomSNRefFromMsg,
			providersn.ParseSagaraSNRefFromMsg,
			providersn.ParseMinionsSNRefFromMsg,
		} {
			pr, sn := parser(msg)
			if out.Reff == "" {
				out.Reff = pr
			}
			if out.SN == "" {
				out.SN = sn
			}
			if out.Reff != "" && out.SN != "" {
				break
			}
		}
	}
	return out
}

func ExtractSaldoTerakhirFromMsg(msg string) (int64, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return 0, false
	}

	if m := reSaldoAkhir.FindStringSubmatch(msg); len(m) == 2 {
		s := strings.TrimSpace(m[1])
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", "")
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil && n > 0 {
			return n, true
		}
	}

	if m := rePoin.FindStringSubmatch(msg); len(m) == 2 {
		s := strings.TrimSpace(m[1])
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", "")
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil && n > 0 {
			return n, true
		}
	}

	if m := reSaldoLabel.FindStringSubmatch(msg); len(m) == 2 {
		s := strings.TrimSpace(m[1])
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", "")
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil && n > 0 {
			return n, true
		}
	}

	m := reSaldo.FindStringSubmatch(msg)
	if len(m) != 2 {
		return 0, false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(m[1]), 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
