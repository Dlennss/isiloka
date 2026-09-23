package provider

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/smb"
)

func normalizeSMBDirectDest(rawProduct string, code string, bankPrefix string, dest string) string {
	dest = strings.TrimSpace(dest)
	rawProduct = strings.ToUpper(strings.TrimSpace(rawProduct))
	code = strings.ToUpper(strings.TrimSpace(code))
	bankPrefix = strings.TrimSpace(bankPrefix)

	stripLead := func(prefix string) {
		prefix = strings.ToUpper(strings.TrimSpace(prefix))
		if prefix == "" {
			return
		}
		if strings.HasPrefix(strings.ToUpper(dest), prefix) {
			dest = strings.TrimSpace(dest[len(prefix):])
			dest = strings.TrimLeft(dest, ":.- ")
		}
	}

	stripLead(rawProduct)
	stripLead(code + ":" + bankPrefix)
	stripLead(code + ".")
	stripLead(code + ":")
	stripLead(code)

	if bankPrefix != "" && !strings.HasPrefix(dest, bankPrefix) {
		dest = bankPrefix + dest
	}
	return dest
}

type SMBAdapter struct{ C *smb.Client }

func (a *SMBAdapter) Name() string { return "smb" }
func (a *SMBAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	mode, code, bankPrefix, err := smb.ParseMappedCodeTargetWithMode(req.Mode, req.Product)
	if err != nil {
		return nil, fmt.Errorf("smb parse code: %w", err)
	}
	dest := strings.TrimSpace(req.Dest)
	if code == "BIFASTOPEN" || code == "BIFASTOPEN2" {
		dest = normalizeSMBDirectDest(req.Product, code, bankPrefix, dest)
	}

	dispatch, callErr := a.C.Dispatch(ctx, mode, smb.Request{
		KodeProduk: code,
		Tujuan:     dest,
		Qty:        req.Qty,
		RefID:      req.RefID,
	}, "PAY")
	if dispatch == nil {
		return &PayResponse{HTTPStatus: 0, Raw: map[string]any{"error": "smb dispatch nil"}}, callErr
	}

	body := dispatch.Final.Body
	resp := &PayResponse{
		HTTPStatus:  dispatch.Final.HTTPStatus,
		Body:        body,
		RC:          smb.ExtractStatusCode(body),
		Message:     body,
		ProviderRef: smb.ParseProviderRef(body),
		Price:       smb.ParsePrice(body),
		Raw:         map[string]any{"dispatch": dispatch},
	}
	if bal, ok := smb.ParseLastBalance(body); ok && bal > 0 {
		resp.Balance = bal
	} else if dispatch.LastBalance != nil && *dispatch.LastBalance > 0 {
		resp.Balance = *dispatch.LastBalance
	}
	return resp, callErr
}
