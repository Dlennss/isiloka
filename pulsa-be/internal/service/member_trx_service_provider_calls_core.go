package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/db"
	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/model"
)

func buildProviderProduct(code string, specialCode string) string {
	code = strings.TrimSpace(code)
	specialCode = strings.TrimSpace(specialCode)
	if specialCode == "" {
		return code
	}
	return strings.ToUpper(specialCode) + ":" + code
}

func isRajabillerProvider(provider string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), "rajabiller")
}

func rajabillerHTTPFinalFailure(httpStatus int) bool {
	return httpStatus == 401 || httpStatus == 403 || httpStatus == 429
}

func (h *MemberTrxService) callProviderOnly(
	ctx context.Context,
	trxMemberID int64,
	in trxmemberdto.TrxRequest,
	attempt providerRouteAttempt,
	routeReason string,
) (usedProvider string, providerRowID int64, err error) {
	return h.callProviderOnlyWithDest(ctx, trxMemberID, in, attempt, routeReason, in.Dest, "")
}

func (h *MemberTrxService) callProviderOnlyWithDest(
	ctx context.Context,
	trxMemberID int64,
	in trxmemberdto.TrxRequest,
	attempt providerRouteAttempt,
	routeReason string,
	destOverride string,
	fallbackFrom string,
) (usedProvider string, providerRowID int64, err error) {
	reason := strings.TrimSpace(routeReason)
	if reason == "" {
		reason = "random_routing"
	}
	if !strings.EqualFold(strings.TrimSpace(attempt.Name), provider.Pulsa24JamProviderName) {
		return attempt.Name, 0, fmt.Errorf("provider H2H wajib pulsa24jam, got %s", attempt.Name)
	}
	requestDest := strings.TrimSpace(destOverride)
	if requestDest == "" {
		requestDest = strings.TrimSpace(in.Dest)
	}
	if strings.EqualFold(strings.TrimSpace(attempt.Name), "loketbayar") &&
		h.isBankH2HProduct(ctx, in.Product) &&
		isLoketBayarBankTransferProduct(attempt.SpecialCode) {
		requestDest = buildLoketBankDest(attempt.KodeProduk, requestDest)
	}

	client := h.Clients[attempt.Name]
	if client == nil {
		return attempt.Name, 0, fmt.Errorf("provider tidak didukung: %s", attempt.Name)
	}

	providerProduct := buildProviderProduct(attempt.KodeProduk, attempt.SpecialCode)
	merchantID := h.rajabillerMerchantIDForAttempt(ctx, in.Product, attempt, in.MerchantID)

	reqPayload := map[string]any{
		"commands":      in.Commands,
		"provider":      attempt.Name,
		"product_in":    in.Product,
		"product_sent":  providerProduct,
		"special_code":  attempt.SpecialCode,
		"mode":          attempt.Mode,
		"dest":          requestDest,
		"qty":           in.Qty,
		"refid":         in.RefID,
		"route_reason":  reason,
		"fallback_from": strings.TrimSpace(fallbackFrom),
	}
	if hp := strings.TrimSpace(in.HP); hp != "" {
		reqPayload["hp"] = hp
	}
	if berita := strings.TrimSpace(in.Berita); berita != "" {
		reqPayload["berita"] = berita
	}
	if merchantID != "" {
		reqPayload["id_merchant"] = merchantID
	}

	row, cErr := h.JPRepo.Create(ctx, model.JavapayTrxCreateIn{
		Provider:            attempt.Name,
		TransaksiMemberID:   trxMemberID,
		RefID:               in.RefID,
		Perintah:            in.Commands,
		ProdukSKUSnapshot:   attempt.ProdukSKUSnapshot,
		ProdukProviderMapID: attempt.ProdukProviderMapID,
		KodeProduk:          providerProduct,
		Tujuan:              requestDest,
		Qty:                 in.Qty,
	}, reqPayload)
	if cErr != nil {
		return attempt.Name, 0, cErr
	}
	if row != nil {
		providerRowID = row.ID
	}

	h.logf("REQUEST provider=%s refid=%s payload=%v", attempt.Name, in.RefID, reqPayload)

	payReq := provider.PayRequest{
		Command:    in.Commands,
		Product:    providerProduct,
		Mode:       attempt.Mode,
		Dest:       requestDest,
		Qty:        in.Qty,
		RefID:      in.RefID,
		HP:         in.HP,
		Berita:     in.Berita,
		MerchantID: merchantID,
	}

	retryLimit := providerRetryLimitForName(attempt.Name)
	resp, callErr := runProviderPayWithImmediateRetries(attempt.Name, ctx, func(callCtx context.Context) (*provider.PayResponse, error) {
		return client.Pay(callCtx, payReq)
	}, func(nextAttempt int, limit int, hs int, retryErr error) {
		h.logf("RETRY provider=%s refid=%s attempt=%d/%d http=%d err=%v",
			attempt.Name, in.RefID, nextAttempt, limit, hs, retryErr)
	})

	if resp == nil {
		hs := 0
		failMsg := fmt.Sprintf("provider tidak merespon setelah %dx retry", retryLimit)
		definitelyNotDispatched := providerTransportDefinitelyNotDispatched(callErr)
		if definitelyNotDispatched {
			failMsg = fmt.Sprintf("provider transport gagal sebelum request terkirim setelah %dx retry: %v", retryLimit, callErr)
		} else if callErr != nil {
			failMsg = callErr.Error()
		}
		_ = h.JPRepo.UpdateResult(ctx, providerRowID, db.UpdateResult{
			HTTPStatus:   &hs,
			Pesan:        &failMsg,
			ResponMentah: map[string]any{"error": failMsg, "transport_error": fmt.Sprint(callErr)},
		})
		if providerRetriesUntilCallback(attempt.Name) {
			h.logf("ROUTING hold pending retry-until-callback provider=%s refid=%s row_id=%d reason=no_response err=%v",
				attempt.Name, in.RefID, providerRowID, callErr)
			return attempt.Name, providerRowID, nil
		}
		if definitelyNotDispatched {
			if providerRowID > 0 {
				_ = h.JPRepo.ForceFailIfPending(ctx, providerRowID, hs, failMsg, map[string]any{"error": failMsg, "transport_error": fmt.Sprint(callErr), "provider": attempt.Name, "route_reason": reason})
			}
			return attempt.Name, providerRowID, fmt.Errorf("%s transport gagal sebelum request terkirim setelah %dx retry: %v", attempt.Name, retryLimit, callErr)
		}
		if providerRowID > 0 {
			_ = h.JPRepo.ForceFailIfPending(ctx, providerRowID, hs, failMsg, map[string]any{"error": failMsg, "provider": attempt.Name, "route_reason": reason})
		}
		return attempt.Name, providerRowID, fmt.Errorf("%s gagal setelah %dx retry: %v", attempt.Name, retryLimit, callErr)
	}

	if resp.RequestRaw != nil {
		_ = h.JPRepo.UpdateRequestMentah(ctx, providerRowID, resp.RequestRaw)
	}

	upd := db.UpdateResult{
		HTTPStatus:   &resp.HTTPStatus,
		Pesan:        helper.PtrString(resp.Message),
		Harga:        helper.PtrI64(resp.Price),
		NoReferensi:  helper.PtrString(resp.ProviderRef),
		ResponMentah: resp.Raw,
	}
	if resp.RC != "" {
		upd.KodeRespon = &resp.RC
	}
	definitelyNotDispatched := callErr != nil && resp.HTTPStatus == 0 && providerTransportDefinitelyNotDispatched(callErr)
	hasProviderReply := providerPayResponseHasProviderReply(resp)
	if (resp.HTTPStatus != 200 || callErr != nil) && resp.Message == "" {
		failMsg := fmt.Sprintf("%s gagal http=%d setelah %dx retry", attempt.Name, resp.HTTPStatus, retryLimit)
		if definitelyNotDispatched {
			failMsg = fmt.Sprintf("provider transport gagal sebelum request terkirim setelah %dx retry: %v", retryLimit, callErr)
		} else if callErr != nil && resp.HTTPStatus == 0 && providerMaySendLateCallback(attempt.Name) && hasProviderReply {
			failMsg = providerLateCallbackPendingMessage()
		}
		upd.Pesan = &failMsg
	}
	if resp.Balance > 0 {
		upd.SaldoTerakhir = &resp.Balance
		h.insertProviderSnapshot(ctx, attempt.Name, trxMemberID, providerRowID, in.RefID, resp.Balance, "member_trx_"+attempt.Name, resp.Raw)
	}
	_ = h.JPRepo.UpdateResult(ctx, providerRowID, upd)

	if callErr != nil || resp.HTTPStatus != 200 {
		failMsg := strings.TrimSpace(resp.Message)
		if failMsg == "" {
			failMsg = strings.TrimSpace(resp.Body)
		}
		if failMsg == "" {
			failMsg = fmt.Sprintf("%s error http=%d err=%v", attempt.Name, resp.HTTPStatus, callErr)
		}
		if callErr != nil && resp.HTTPStatus == 0 && providerMaySendLateCallback(attempt.Name) && hasProviderReply {
			if definitelyNotDispatched {
				failMsg = fmt.Sprintf("provider transport gagal sebelum request terkirim setelah %dx retry: %v", retryLimit, callErr)
				if providerRowID > 0 {
					_ = h.JPRepo.ForceFailIfPending(ctx, providerRowID, resp.HTTPStatus, failMsg, resp.Raw)
				}
				return attempt.Name, providerRowID, fmt.Errorf("%s transport gagal sebelum request terkirim setelah %dx retry: %v", attempt.Name, retryLimit, callErr)
			}
			h.logf("ROUTING hold pending after uncertain provider transport provider=%s refid=%s row_id=%d err=%v",
				attempt.Name, in.RefID, providerRowID, callErr)
			return attempt.Name, providerRowID, nil
		}
		if providerRetriesUntilCallback(attempt.Name) {
			h.logf("ROUTING hold pending retry-until-callback provider=%s refid=%s row_id=%d http=%d err=%v",
				attempt.Name, in.RefID, providerRowID, resp.HTTPStatus, callErr)
			return attempt.Name, providerRowID, nil
		}
		if isRajabillerProvider(attempt.Name) && !rajabillerHTTPFinalFailure(resp.HTTPStatus) {
			return attempt.Name, providerRowID, nil
		}
		if providerRowID > 0 {
			_ = h.JPRepo.ForceFailIfPending(ctx, providerRowID, resp.HTTPStatus, failMsg, resp.Raw)
		}
		h.disableProviderForOutOfBalance(ctx, attempt.Name, failMsg)
		return attempt.Name, providerRowID, fmt.Errorf("%s error http=%d err=%v body=%s", attempt.Name, resp.HTTPStatus, callErr, strings.TrimSpace(resp.Body))
	}

	h.logf("%s menerima transaksi refid=%s ticket=%s harga=%d produk_input=%s produk_kirim=%s body=%s",
		strings.ToUpper(attempt.Name), in.RefID, resp.ProviderRef, resp.Price, in.Product, providerProduct, resp.Body)

	if helper.ProviderResponseAccepted(attempt.Name, resp.Body) {
		return attempt.Name, providerRowID, nil
	}
	if helper.ProviderResponseImmediateReject(attempt.Name, resp.Body) {
		failMsg := strings.TrimSpace(resp.Message)
		if failMsg == "" {
			failMsg = strings.TrimSpace(resp.Body)
		}
		if failMsg == "" {
			failMsg = fmt.Sprintf("%s reject bisnis", attempt.Name)
		}
		h.disableProviderForOutOfBalance(ctx, attempt.Name, failMsg)
		if providerRowID > 0 {
			_ = h.JPRepo.ForceFailIfPending(ctx, providerRowID, resp.HTTPStatus, failMsg, resp.Raw)
		}
		return attempt.Name, providerRowID, fmt.Errorf("%s reject bisnis: %s", attempt.Name, failMsg)
	}
	return attempt.Name, providerRowID, fmt.Errorf("respons %s tidak dikenali: %s", attempt.Name, strings.TrimSpace(resp.Body))
}

func (h *MemberTrxService) routeProviderForPayInqSkipping(
	ctx context.Context,
	trxMemberID int64,
	in trxmemberdto.TrxRequest,
	billingNominal int64,
	skipCandidates map[string]bool,
) (used string, rowID int64, err error) {
	attempts := h.buildProviderAttempts(ctx, in, billingNominal, in.Commands == "PAY", skipCandidates)
	isBank := h.isBankH2HProduct(ctx, in.Product)
	if len(attempts) == 0 {
		return "", 0, fmt.Errorf("tidak ada provider eligible")
	}

	h.logf("ROUTING kandidat refid=%s produk=%s cmd=%s kandidat=%v",
		in.RefID, in.Product, in.Commands, attempts)

	var lastUsed string
	var lastErr error
	var lastRowID int64
	for i, a := range attempts {
		attemptCtx, cancel := context.WithTimeout(ctx, providerCallWindowForName(a.Name))
		usedProvider, providerRowID, callErr := h.callProviderOnly(attemptCtx, trxMemberID, in, a, a.Src)
		cancel()

		if callErr == nil {
			return usedProvider, providerRowID, nil
		}

		lastUsed = usedProvider
		lastErr = callErr
		lastRowID = providerRowID

		h.logf("ROUTING SKU gagal refid=%s produk=%s cmd=%s sku=%d/%d provider=%s kode=%s err=%v",
			in.RefID, in.Product, in.Commands, i+1, len(attempts), a.Name, a.KodeProduk, callErr)

		if isBank && (strings.EqualFold(strings.TrimSpace(a.Name), "smb") || strings.EqualFold(strings.TrimSpace(a.Name), "rajabiller")) {
			continue
		}

		if !isProviderSystemError(callErr) {
			return lastUsed, lastRowID, lastErr
		}
	}

	return lastUsed, lastRowID, lastErr
}
