package provider

import (
	"context"

	"pulsa2/sagaramobile"
)

type SagaramobileAdapter struct{ C *sagaramobile.Client }

func (a *SagaramobileAdapter) Name() string { return "sagaramobile" }
func (a *SagaramobileAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	acc, hs, body, err := a.C.Trx(ctx, req.Product, req.Dest, req.Qty, req.RefID)
	if err != nil {
		return &PayResponse{HTTPStatus: hs, Body: body, Raw: map[string]any{"error": err.Error()}}, err
	}
	return &PayResponse{
		HTTPStatus:  hs,
		Body:        body,
		Message:     body,
		ProviderRef: acc.ProviderRef,
		Price:       acc.Price,
		Raw:         map[string]any{"raw": body},
	}, nil
}
