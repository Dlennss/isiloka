package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/rajabiller"
)

type RajabillerAdapter struct{ C *rajabiller.Client }

const rajabillerBankDefaultBerita = "bayar"

func (a *RajabillerAdapter) Name() string { return "rajabiller" }

func rajabillerMethod(command string, mode string) string {
	mode = strings.ToUpper(strings.TrimSpace(mode))
	switch mode {
	case "CEK", "INQ", "INQUIRY":
		return "cek"
	case "BAYAR", "PAY":
		return "bayar"
	}

	switch strings.ToUpper(strings.TrimSpace(command)) {
	case "INQ":
		return "cek"
	default:
		return "bayar"
	}
}

type rajabillerProductParts struct {
	Product  string
	Server   string
	KodeBank string
}

func rajabillerIsBankTransfer(product string, mode string) bool {
	product = strings.ToUpper(strings.TrimSpace(product))
	mode = strings.ToUpper(strings.TrimSpace(mode))
	return strings.Contains(mode, "BANK") ||
		strings.Contains(mode, "TRANSFER") ||
		strings.Contains(mode, "TRF") ||
		rajabillerLooksLikeBankProductCode(product)
}

func rajabillerLooksLikeBankProductCode(product string) bool {
	product = strings.ToUpper(strings.TrimSpace(product))
	return strings.HasPrefix(product, "BLTRF")
}

func rajabillerIsOpenDenom(product string, mode string) bool {
	product = strings.ToUpper(strings.TrimSpace(product))
	mode = strings.ToUpper(strings.TrimSpace(mode))
	return strings.Contains(mode, "OPEN_DENOM") ||
		strings.Contains(mode, "OPEN-DENOM") ||
		strings.Contains(mode, "OPEN DENOM") ||
		strings.HasPrefix(product, "EM")
}

func rajabillerShouldCheckBeforePay(parts rajabillerProductParts, mode string) bool {
	return parts.KodeBank != "" || rajabillerIsOpenDenom(parts.Product, mode)
}

func splitRajabillerProduct(raw string, mode string) rajabillerProductParts {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return rajabillerProductParts{}
	}
	if left, right, ok := strings.Cut(raw, ":"); ok {
		left = strings.TrimSpace(left)
		right = strings.TrimSpace(right)
		if rajabillerLooksLikeBankProductCode(left) {
			return rajabillerProductParts{Product: left, KodeBank: right}
		}
		if rajabillerLooksLikeBankProductCode(right) || rajabillerIsBankTransfer(right, mode) {
			return rajabillerProductParts{Product: right, KodeBank: left}
		}
		// P24 builds provider product as SPECIAL_CODE:KODE_PROVIDER.
		// For non-bank Rajabiller products, use special_code as optional server.
		return rajabillerProductParts{Product: right, Server: left}
	}
	return rajabillerProductParts{Product: raw}
}

func rajabillerShouldSendNominal(product string, mode string) bool {
	product = strings.ToUpper(strings.TrimSpace(product))
	mode = strings.ToUpper(strings.TrimSpace(mode))
	if product == "" {
		return false
	}

	switch {
	case strings.Contains(mode, "NO_NOMINAL") || strings.Contains(mode, "NONOMINAL"):
		return false
	case strings.Contains(mode, "NOMINAL") ||
		rajabillerIsOpenDenom(product, mode) ||
		strings.Contains(mode, "BANK") ||
		strings.Contains(mode, "TRANSFER") ||
		strings.Contains(mode, "TRF") ||
		strings.Contains(mode, "KARTU_KREDIT") ||
		strings.Contains(mode, "CREDIT_CARD"):
		return true
	case strings.HasPrefix(product, "PLNPRA"):
		return true
	case strings.HasPrefix(product, "EM"):
		return true
	case strings.HasPrefix(product, "BLTRF"):
		return true
	case strings.HasPrefix(product, "CC") || strings.HasPrefix(product, "KK") || strings.Contains(product, "KARTU"):
		return true
	default:
		return false
	}
}

func rajabillerProviderRef(resp rajabiller.TransactionResponse) string {
	providerRef := strings.TrimSpace(resp.SN)
	if providerRef == "" {
		providerRef = strings.TrimSpace(resp.Token)
	}
	if providerRef == "" {
		providerRef = strings.TrimSpace(resp.Ref)
	}
	if providerRef == "" {
		providerRef = strings.TrimSpace(resp.ProviderRefID)
	}
	if providerRef == "" {
		providerRef = strings.TrimSpace(resp.TrxID)
	}
	return providerRef
}

func rajabillerBody(resp rajabiller.TransactionResponse) string {
	bodyBytes, _ := json.Marshal(resp.Raw)
	body := strings.TrimSpace(string(bodyBytes))
	statusPrefix := strings.TrimSpace(fmt.Sprintf("rc=%s status=%s", resp.RC, resp.Status))
	if body == "" || body == "null" {
		body = statusPrefix
	} else if statusPrefix != "rc= status=" {
		body = statusPrefix + " body=" + body
	}
	return body
}

func rajabillerPayResponse(resp rajabiller.TransactionResponse, hs int, reqRaw map[string]any) *PayResponse {
	return &PayResponse{
		HTTPStatus:  hs,
		Body:        rajabillerBody(resp),
		RC:          resp.RC,
		Message:     resp.Status,
		ProviderRef: rajabillerProviderRef(resp),
		Price:       resp.Price,
		Balance:     resp.Balance,
		Raw:         resp.Raw,
		RequestRaw:  reqRaw,
	}
}

func rajabillerBankHP(explicitHP string, refID string) string {
	if hp := strings.TrimSpace(explicitHP); hp != "" {
		return hp
	}
	return rajabillerSyntheticBankHP(refID)
}

func rajabillerSyntheticBankHP(seed string) string {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = "rajabiller"
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(seed))
	n := h.Sum64()

	var b strings.Builder
	b.Grow(12)
	b.WriteString("08")
	for i := 0; i < 10; i++ {
		n = n*2862933555777941757 + 3037000493
		b.WriteByte(byte('0' + n%10))
	}
	return b.String()
}

func rajabillerBankBerita(explicitBerita string) string {
	if berita := strings.TrimSpace(explicitBerita); berita != "" {
		return berita
	}
	return rajabillerBankDefaultBerita
}

func (a *RajabillerAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	if a == nil || a.C == nil {
		return nil, fmt.Errorf("rajabiller client nil")
	}

	parts := splitRajabillerProduct(req.Product, req.Mode)
	method := rajabillerMethod(req.Command, req.Mode)
	hp := strings.TrimSpace(req.HP)
	if hp == "" && parts.KodeBank != "" {
		hp = rajabillerBankHP("", req.RefID)
	}
	berita := strings.TrimSpace(req.Berita)
	if berita == "" && parts.KodeBank != "" {
		berita = rajabillerBankBerita("")
	}
	txReq := rajabiller.TransactionRequest{
		Method:      method,
		Product:     parts.Product,
		Dest:        req.Dest,
		RefID:       req.RefID,
		Nominal:     req.Qty,
		SendNominal: rajabillerShouldSendNominal(parts.Product, req.Mode),
		Server:      parts.Server,
		KodeBank:    parts.KodeBank,
		HP:          hp,
		Berita:      berita,
		MerchantID:  req.MerchantID,
	}

	if rajabillerShouldCheckBeforePay(parts, req.Mode) && method == "bayar" {
		cekReq := txReq
		cekReq.Method = "cek"

		cekResp, cekHTTPStatus, cekReqRaw, cekErr := a.C.Transaction(ctx, cekReq)
		cekPayResp := rajabillerPayResponse(cekResp, cekHTTPStatus, map[string]any{
			"stage":        "cek",
			"cek_request":  cekReqRaw,
			"cek_response": cekResp.Raw,
		})
		if cekErr != nil {
			return cekPayResp, cekErr
		}
		if cekHTTPStatus != 200 || !helper.ProviderResponseAccepted("rajabiller", cekPayResp.Body) {
			return cekPayResp, nil
		}

		bayarReq := txReq
		if parts.KodeBank != "" {
			// Live Rajabiller bank transfer rejects BAYAR when berita is present,
			// while CEK may need it for validation.
			bayarReq.Berita = ""
		}
		resp, hs, reqRaw, callErr := a.C.Transaction(ctx, bayarReq)
		if reqRaw == nil {
			reqRaw = map[string]any{}
		}
		reqRaw["stage"] = "bayar"
		reqRaw["cek_request"] = cekReqRaw
		reqRaw["cek_response"] = cekResp.Raw
		reqRaw["cek_http_status"] = cekHTTPStatus
		return rajabillerPayResponse(resp, hs, reqRaw), callErr
	}

	resp, hs, reqRaw, callErr := a.C.Transaction(ctx, txReq)
	return rajabillerPayResponse(resp, hs, reqRaw), callErr
}
