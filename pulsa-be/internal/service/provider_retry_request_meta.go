package service

import (
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
)

func applyProviderRetryRequestMeta(raw any, req *provider.PayRequest) {
	if req == nil {
		return
	}

	_, mode := restoreProviderAttemptMeta(raw)
	if strings.TrimSpace(req.Mode) == "" {
		req.Mode = mode
	}

	m, ok := raw.(map[string]any)
	if !ok || m == nil {
		return
	}
	if strings.TrimSpace(req.HP) == "" {
		req.HP = strings.TrimSpace(helper.TrxToString(m["hp"]))
	}
	if strings.TrimSpace(req.Berita) == "" {
		req.Berita = strings.TrimSpace(helper.TrxToString(m["berita"]))
	}
	if strings.TrimSpace(req.MerchantID) == "" {
		for _, key := range []string{"id_merchant", "merchant_id", "provider_merchant_id"} {
			if v := strings.TrimSpace(helper.TrxToString(m[key])); v != "" {
				req.MerchantID = v
				break
			}
		}
	}
}
