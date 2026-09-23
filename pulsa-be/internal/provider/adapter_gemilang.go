package provider

import (
	"context"

	"pulsa2/gemilang"
	"pulsa2/internal/helper"
)

type GemilangAdapter struct{ C *gemilang.Client }

func (a *GemilangAdapter) Name() string { return "gemilang" }
func (a *GemilangAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	acc, hs, body, err := a.C.TrxNoSign(ctx, req.Product, req.Qty, req.Dest, req.RefID)
	if err != nil {
		return &PayResponse{HTTPStatus: hs, Body: body, Raw: map[string]any{"error": err.Error()}}, err
	}
	resp := &PayResponse{
		HTTPStatus:  hs,
		Body:        body,
		RC:          helper.ExtractProviderStatusCode(body),
		Message:     body,
		ProviderRef: acc.Ticket,
		Price:       acc.Price,
		Raw:         map[string]any{"raw": body},
	}
	if bal, ok := helper.ParseSaldoAfterFromStockMsg(body); ok {
		resp.Balance = bal
	}
	return resp, nil
}
