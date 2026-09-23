package provider

import (
	"context"
	"errors"
	"strings"

	"pulsa2/chytron"
	"pulsa2/smb"
)

type ChytronAdapter struct{ C *chytron.Client }

func (a *ChytronAdapter) Name() string { return "chytron" }

func (a *ChytronAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	if a == nil || a.C == nil {
		return nil, errors.New("chytron client nil")
	}
	dispatch, callErr := a.C.Pay(ctx, chytron.Request{
		KodeProduk: strings.ToUpper(strings.TrimSpace(req.Product)),
		Tujuan:     strings.TrimSpace(req.Dest),
		Qty:        req.Qty,
		RefID:      strings.TrimSpace(req.RefID),
	})
	if dispatch == nil {
		return &PayResponse{HTTPStatus: 0, Raw: map[string]any{"error": "chytron dispatch nil"}}, callErr
	}

	body := dispatch.Pay.Body
	resp := &PayResponse{
		HTTPStatus:  dispatch.Pay.HTTPStatus,
		Body:        body,
		RC:          smb.ExtractStatusCode(body),
		Message:     body,
		ProviderRef: smb.ParseProviderRef(body),
		Price:       smb.ParsePrice(body),
		Raw:         map[string]any{"dispatch": dispatch},
		RequestRaw:  dispatch.Pay.Request,
	}
	if bal, ok := smb.ParseLastBalance(body); ok && bal > 0 {
		resp.Balance = bal
	}
	return resp, callErr
}
