package provider

import (
	"context"

	"pulsa2/ajs"
)

type AJSAdapter struct{ C *ajs.Client }

func (a *AJSAdapter) Name() string { return "ajs" }
func (a *AJSAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	acc, hs, body, err := a.C.TrxNoSign(ctx, req.Product, req.Qty, req.Dest, req.RefID)
	if err != nil {
		return &PayResponse{HTTPStatus: hs, Body: body, Raw: map[string]any{"error": err.Error()}}, err
	}
	return &PayResponse{
		HTTPStatus:  hs,
		Body:        body,
		Message:     body,
		ProviderRef: acc.Ticket,
		Price:       acc.Price,
		Raw:         map[string]any{"raw": body},
	}, nil
}
