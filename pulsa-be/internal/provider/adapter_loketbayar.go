package provider

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/loketbayar"
)

type LoketBayarAdapter struct{ C *loketbayar.Client }

func (a *LoketBayarAdapter) Name() string { return "loketbayar" }

func normalizeLoketProduct(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if head, _, ok := strings.Cut(raw, ":"); ok {
		return strings.TrimSpace(head)
	}
	return raw
}

func (a *LoketBayarAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	if a == nil || a.C == nil {
		return nil, fmt.Errorf("loketbayar client nil")
	}
	resp, hs, reqRaw, callErr := a.C.Topup(ctx, loketbayar.TopupRequest{
		ProductCode: normalizeLoketProduct(req.Product),
		Dest:        req.Dest,
		RefID:       req.RefID,
		Nominal:     req.Qty,
	})
	rc := strings.TrimSpace(resp.Status)
	msg := strings.TrimSpace(resp.Keterangan)
	providerRef := strings.TrimSpace(resp.Reff)
	if providerRef == "" {
		providerRef = strings.TrimSpace(resp.SN)
	}
	if providerRef == "" {
		providerRef = strings.TrimSpace(resp.TrxID)
	}
	body := strings.TrimSpace(fmt.Sprintf("status=%s keterangan=%s", rc, msg))
	return &PayResponse{
		HTTPStatus:  hs,
		Body:        body,
		RC:          rc,
		Message:     msg,
		ProviderRef: providerRef,
		Price:       resp.Price,
		Balance:     resp.Saldo,
		Raw:         resp.Raw,
		RequestRaw:  reqRaw,
	}, callErr
}
