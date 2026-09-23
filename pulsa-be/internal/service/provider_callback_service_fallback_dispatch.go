package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

func callbackFallbackDetachedContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

func callbackFallbackDBContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(callbackFallbackDetachedContext(ctx), 15*time.Second)
}

func fallbackNoResponseShouldStayPending(providerName string, httpStatus int, err error) bool {
	return providerRetriesUntilCallback(providerName) && (httpStatus == 0 || err != nil)
}

func (s *ProviderCallbackService) dispatchFallbackCandidate(ctx context.Context, trx *repository.CallbackTrxMemberFull, failedProvider string, candidate callbackFallbackCandidate, fromTag string, providerMsg string) (int64, error) {
	baseCtx := callbackFallbackDetachedContext(ctx)
	reqQty := trx.QtyProvider
	if reqQty <= 0 {
		reqQty = trx.Qty
	}

	client := s.clients[candidate.Provider]
	if client == nil {
		return 0, fmt.Errorf("provider fallback tidak didukung: %s", candidate.Provider)
	}

	providerProduct := buildProviderProduct(candidate.KodeProduk, ptrString(candidate.SpecialCode))
	merchantID := s.rajabillerMerchantIDForFallback(baseCtx, trx.KodeProduk, candidate, "")
	requestDest := strings.TrimSpace(trx.Tujuan)
	if strings.EqualFold(strings.TrimSpace(candidate.Provider), "loketbayar") && s.isBankH2HProduct(baseCtx, trx.KodeProduk) {
		lookupCtx, lookupCancel := callbackFallbackDBContext(baseCtx)
		rows, err := s.repo.ListAttemptsByRefID(lookupCtx, trx.RefID)
		lookupCancel()
		if err != nil {
			return 0, err
		}
		bankCode, err := latestBankCodeFromAttempts(rows)
		if err != nil {
			return 0, err
		}
		requestDest = buildLoketBankDest(bankCode, requestDest)
	}

	createIn := repository.ProviderTrxCreateIn{
		Provider:            candidate.Provider,
		TransaksiMemberID:   trx.ID,
		RefID:               trx.RefID,
		Perintah:            trx.Perintah,
		ProdukSKUSnapshot:   candidate.ProdukSKUSnapshot,
		ProdukProviderMapID: candidate.ProdukProviderMapID,
		KodeProduk:          providerProduct,
		Tujuan:              requestDest,
		Qty:                 reqQty,
	}
	preCtx, preCancel := callbackFallbackDBContext(baseCtx)
	defer preCancel()
	if existing, err := s.repo.FindProviderTrxByRoute(preCtx, createIn); err != nil {
		return 0, err
	} else if existing != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log",
			"fallback duplicate blocked provider=%s refid=%s row_id=%d map_id=%v kode=%s",
			candidate.Provider, trx.RefID, existing.ID, candidate.ProdukProviderMapID, providerProduct)
		return existing.ID, nil
	}

	reqPayload := map[string]any{
		"from":          fromTag,
		"message":       providerMsg,
		"product_in":    trx.KodeProduk,
		"product_sent":  providerProduct,
		"special_code":  ptrString(candidate.SpecialCode),
		"mode":          ptrString(candidate.Mode),
		"fallback_from": strings.TrimSpace(failedProvider),
		"dest":          requestDest,
	}
	if merchantID != "" {
		reqPayload["id_merchant"] = merchantID
	}
	row, err := s.repo.CreateProviderTrx(preCtx, createIn, reqPayload)
	if err != nil {
		return 0, err
	}

	helper.AppendProviderServiceLog("provider_callback_service.log",
		"REQUEST provider=%s refid=%s payload=map[commands:PAY dest:%s product_in:%s product_sent:%s provider:%s qty:%d refid:%s source:%s fallback_from:%s]",
		candidate.Provider, trx.RefID, requestDest, trx.KodeProduk, providerProduct, candidate.Provider, reqQty, trx.RefID, fromTag, failedProvider)

	payReq := provider.PayRequest{
		Command:    "PAY",
		Product:    providerProduct,
		Mode:       strings.ToUpper(strings.TrimSpace(ptrString(candidate.Mode))),
		Dest:       requestDest,
		Qty:        reqQty,
		RefID:      trx.RefID,
		MerchantID: merchantID,
	}

	retryLimit := providerRetryLimitForName(candidate.Provider)
	callCtx, callCancel := context.WithTimeout(baseCtx, providerCallWindowForName(candidate.Provider))
	resp, callErr := runProviderPayWithImmediateRetries(candidate.Provider, callCtx, func(callCtx context.Context) (*provider.PayResponse, error) {
		return client.Pay(callCtx, payReq)
	}, func(nextAttempt int, limit int, hs int, retryErr error) {
		helper.AppendProviderServiceLog("provider_callback_service.log",
			"RETRY fallback provider=%s refid=%s attempt=%d/%d http=%d err=%v",
			candidate.Provider, trx.RefID, nextAttempt, limit, hs, retryErr)
	})
	callCancel()

	postCtx, postCancel := callbackFallbackDBContext(baseCtx)
	defer postCancel()

	if resp == nil {
		hs := 0
		failMsg := fmt.Sprintf("provider fallback tidak merespon setelah %dx retry", retryLimit)
		if callErr != nil {
			failMsg = callErr.Error()
		}
		_ = s.repo.UpdateResult(postCtx, row.ID, repository.UpdateResult{
			HTTPStatus:   &hs,
			Pesan:        &failMsg,
			ResponMentah: map[string]any{"error": failMsg},
		})
		if fallbackNoResponseShouldStayPending(candidate.Provider, hs, callErr) {
			helper.AppendProviderServiceLog("provider_callback_service.log",
				"FALLBACK hold pending retry-until-callback provider=%s refid=%s row_id=%d reason=no_response err=%v",
				candidate.Provider, trx.RefID, row.ID, callErr)
			return row.ID, nil
		}
		if isRajabillerProvider(candidate.Provider) {
			return row.ID, nil
		}
		_ = s.repo.ForceFailIfPending(postCtx, row.ID, hs, failMsg, map[string]any{"error": failMsg, "provider": candidate.Provider, "source": fromTag})
		return row.ID, fmt.Errorf("%s fallback gagal setelah %dx retry: %v", candidate.Provider, retryLimit, callErr)
	}

	if resp.RequestRaw != nil {
		_ = s.repo.UpdateRequestMentah(postCtx, row.ID, resp.RequestRaw)
	}

	upd := repository.UpdateResult{
		HTTPStatus:   &resp.HTTPStatus,
		Pesan:        helper.PtrString(resp.Message),
		Harga:        helper.PtrI64(resp.Price),
		NoReferensi:  helper.PtrString(resp.ProviderRef),
		ResponMentah: resp.Raw,
	}
	if resp.RC != "" {
		upd.KodeRespon = &resp.RC
	}
	if (resp.HTTPStatus != 200 || callErr != nil) && resp.Message == "" {
		failMsg := fmt.Sprintf("%s fallback gagal http=%d setelah %dx retry", candidate.Provider, resp.HTTPStatus, retryLimit)
		upd.Pesan = &failMsg
	}
	if resp.Balance > 0 {
		upd.SaldoTerakhir = &resp.Balance
		_ = s.repo.InsertProviderSnapshot(postCtx, repository.ProviderSnapshotIn{
			Provider:            candidate.Provider,
			SaldoProvider:       resp.Balance,
			RefID:               trx.RefID,
			TransaksiMemberID:   &trx.ID,
			TransaksiProviderID: &row.ID,
			Sumber:              "callback_fallback_dispatch_" + candidate.Provider,
		})
	}
	_ = s.repo.UpdateResult(postCtx, row.ID, upd)

	if callErr != nil || resp.HTTPStatus != 200 {
		failMsg := strings.TrimSpace(resp.Message)
		if failMsg == "" {
			failMsg = strings.TrimSpace(resp.Body)
		}
		if failMsg == "" {
			failMsg = fmt.Sprintf("%s error http=%d err=%v", candidate.Provider, resp.HTTPStatus, callErr)
		}
		if fallbackNoResponseShouldStayPending(candidate.Provider, resp.HTTPStatus, callErr) {
			helper.AppendProviderServiceLog("provider_callback_service.log",
				"FALLBACK hold pending retry-until-callback provider=%s refid=%s row_id=%d http=%d err=%v",
				candidate.Provider, trx.RefID, row.ID, resp.HTTPStatus, callErr)
			return row.ID, nil
		}
		s.disableProviderForOutOfBalance(postCtx, candidate.Provider, failMsg)
		if isRajabillerProvider(candidate.Provider) && !rajabillerHTTPFinalFailure(resp.HTTPStatus) {
			return row.ID, nil
		}
		_ = s.repo.ForceFailIfPending(postCtx, row.ID, resp.HTTPStatus, failMsg, resp.Raw)
		return row.ID, fmt.Errorf("%s error http=%d err=%v body=%s", candidate.Provider, resp.HTTPStatus, callErr, strings.TrimSpace(resp.Body))
	}
	if helper.ProviderResponseAccepted(candidate.Provider, resp.Body) {
		return row.ID, nil
	}
	if helper.ProviderResponseImmediateReject(candidate.Provider, resp.Body) {
		failMsg := strings.TrimSpace(resp.Message)
		if failMsg == "" {
			failMsg = strings.TrimSpace(resp.Body)
		}
		if failMsg == "" {
			failMsg = fmt.Sprintf("%s fallback reject bisnis", candidate.Provider)
		}
		s.disableProviderForOutOfBalance(postCtx, candidate.Provider, failMsg)
		_ = s.repo.ForceFailIfPending(postCtx, row.ID, resp.HTTPStatus, failMsg, resp.Raw)
		return row.ID, fmt.Errorf("%s", failMsg)
	}
	return row.ID, fmt.Errorf("respons %s tidak dikenali: %s", candidate.Provider, strings.TrimSpace(resp.Body))
}

func (s *ProviderCallbackService) tryFallbackFromProvider(ctx context.Context, trx *repository.CallbackTrxMemberFull, failedProvider string, sourceTag string, providerMsg string) (bool, string, int64, error) {
	baseCtx := callbackFallbackDetachedContext(ctx)
	if trx == nil {
		return false, "", 0, nil
	}
	if strings.ToUpper(strings.TrimSpace(trx.Perintah)) != "PAY" {
		return false, "", 0, nil
	}

	unlockLocal := lockProviderCallback(trx.ID)
	defer unlockLocal()

	if used, prov := s.hasOtherProviderAttemptForRef(baseCtx, trx, failedProvider); used {
		helper.AppendProviderServiceLog("provider_callback_service.log", "fallback hold pending from=%s refid=%s alasan=provider_lain_sudah_ada provider=%s", failedProvider, trx.RefID, prov)
		return true, prov, 0, nil
	}
	triedKeys, err := s.listTriedFallbackKeys(baseCtx, trx.RefID)
	if err != nil {
		return false, "", 0, err
	}
	candidates, err := s.buildFallbackCandidates(baseCtx, trx, failedProvider, triedKeys)
	if err != nil {
		return false, "", 0, err
	}
	var lastErr error
	for _, candidate := range candidates {
		key := callbackAttemptKey(candidate.Provider, candidate.ProdukProviderMapID, candidate.KodeProduk)
		if triedKeys[key] {
			continue
		}
		rowID, err := s.dispatchFallbackCandidate(baseCtx, trx, failedProvider, candidate, sourceTag, providerMsg)
		if err == nil {
			return true, candidate.Provider, rowID, nil
		}
		lastErr = err
		triedKeys[key] = true
	}
	if lastErr != nil {
		return false, "", 0, lastErr
	}
	return false, "", 0, nil
}
