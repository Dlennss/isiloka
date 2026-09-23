package provider

import (
	"context"

	"pulsa2/javapay"
)

type JavapayAdapter struct{ C *javapay.Client }

func (a *JavapayAdapter) Name() string { return "javapay" }
func (a *JavapayAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	jpResp, hs, reqMentah, err := a.C.Trx(ctx, "PAY", req.Product, req.Dest, req.Qty, req.RefID)
	if err != nil {
		return &PayResponse{HTTPStatus: hs, Raw: map[string]any{"error": err.Error(), "request": reqMentah}}, err
	}
	rc := ""
	msg := ""
	if data, ok := jpResp["data"].(map[string]any); ok {
		if v, ok := data["rc"]; ok {
			rc, _ = v.(string)
		}
		if v, ok := data["message"]; ok {
			msg, _ = v.(string)
		}
	}
	return &PayResponse{
		HTTPStatus: hs,
		Body:       msg,
		RC:         rc,
		Message:    msg,
		Raw:        map[string]any{"response": jpResp, "request": reqMentah},
	}, nil
}
