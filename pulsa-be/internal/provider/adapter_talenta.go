package provider

import (
	"context"

	"pulsa2/talenta"
)

type TalentaAdapter struct{ C *talenta.Client }

func (a *TalentaAdapter) Name() string { return "talentapay" }
func (a *TalentaAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
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
