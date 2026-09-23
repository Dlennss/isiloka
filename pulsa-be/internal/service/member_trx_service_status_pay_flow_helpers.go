package service

import (
	"context"
	"strings"

	"pulsa2/db"
	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

type statusPayProviderRows struct {
	jp            *model.JavapayTrxRow
	ys            *model.JavapayTrxRow
	tl            *model.JavapayTrxRow
	mk            *model.JavapayTrxRow
	sg            *model.JavapayTrxRow
	mn            *model.JavapayTrxRow
	tr            *model.JavapayTrxRow
	aj            *model.JavapayTrxRow
	gm            *model.JavapayTrxRow
	sm            *model.JavapayTrxRow
	lb            *model.JavapayTrxRow
	p24           *model.JavapayTrxRow
	allAttempts   []*model.JavapayTrxRow
	providerState []providerState
	hasAny        bool
}

func (h *MemberTrxService) loadStatusPayProviderRows(ctx context.Context, refID string) (*statusPayProviderRows, error) {
	jp, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "javapay")
	if err != nil {
		return nil, err
	}
	ys, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "yuscom")
	if err != nil {
		return nil, err
	}
	tl, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "talentapay")
	if err != nil {
		return nil, err
	}
	mk, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "multikom")
	if err != nil {
		return nil, err
	}
	sg, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "sagaramobile")
	if err != nil {
		return nil, err
	}
	mn, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "minions")
	if err != nil {
		return nil, err
	}
	tr, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "trionik")
	if err != nil {
		return nil, err
	}
	aj, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "ajs")
	if err != nil {
		return nil, err
	}
	gm, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "gemilang")
	if err != nil {
		return nil, err
	}
	sm, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "smb")
	if err != nil {
		return nil, err
	}
	lb, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "loketbayar")
	if err != nil {
		return nil, err
	}
	p24, err := h.JPRepo.GetLatestByRefIDProvider(ctx, refID, "pulsa24jam")
	if err != nil {
		return nil, err
	}
	allAttempts, err := h.JPRepo.ListByRefID(ctx, refID)
	if err != nil {
		return nil, err
	}

	rows := &statusPayProviderRows{
		jp:          jp,
		ys:          ys,
		tl:          tl,
		mk:          mk,
		sg:          sg,
		mn:          mn,
		tr:          tr,
		aj:          aj,
		gm:          gm,
		sm:          sm,
		lb:          lb,
		p24:         p24,
		allAttempts: allAttempts,
		hasAny:      jp != nil || ys != nil || tl != nil || mk != nil || sg != nil || mn != nil || tr != nil || aj != nil || gm != nil || sm != nil || lb != nil || p24 != nil,
	}
	rows.providerState = []providerState{
		{name: "pulsa24jam", row: p24},
		{name: "javapay", row: jp},
		{name: "yuscom", row: ys},
		{name: "talentapay", row: tl},
		{name: "multikom", row: mk},
		{name: "sagaramobile", row: sg},
		{name: "minions", row: mn},
		{name: "trionik", row: tr},
		{name: "ajs", row: aj},
		{name: "gemilang", row: gm},
		{name: "smb", row: sm},
		{name: "loketbayar", row: lb},
	}
	return rows, nil
}

func (h *MemberTrxService) handleStatusPayWithoutJavapayRow(ctx context.Context, trx *repository.TrxMemberFull, rows *statusPayProviderRows, _ map[string]bool) *serviceResponse {
	if rows.p24 != nil || rows.ys != nil || rows.tl != nil || rows.mk != nil || rows.sg != nil || rows.mn != nil || rows.tr != nil || rows.aj != nil || rows.gm != nil || rows.sm != nil || rows.lb != nil {
		waitProvider := "yuscom"
		waitReason := "wait_yuscom_callback"
		if rows.p24 != nil {
			waitProvider = "pulsa24jam"
			waitReason = "wait_pulsa24jam_callback"
		} else if rows.tl != nil {
			waitProvider = "talentapay"
			waitReason = "wait_talentapay_callback"
		} else if rows.mk != nil {
			waitProvider = "multikom"
			waitReason = "wait_multikom_callback"
		} else if rows.sg != nil {
			waitProvider = "sagaramobile"
			waitReason = "wait_sagaramobile_callback"
		} else if rows.mn != nil {
			waitProvider = "minions"
			waitReason = "wait_minions_callback"
		} else if rows.tr != nil {
			waitProvider = "trionik"
			waitReason = "wait_trionik_callback"
		} else if rows.aj != nil {
			waitProvider = "ajs"
			waitReason = "wait_ajs_callback"
		} else if rows.gm != nil {
			waitProvider = "gemilang"
			waitReason = "wait_gemilang_callback"
		} else if rows.sm != nil {
			waitProvider = "smb"
			waitReason = "wait_smb_callback"
		} else if rows.lb != nil {
			waitProvider = "loketbayar"
			waitReason = "wait_loketbayar_callback"
		}
		return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "provider": waitProvider, "reason": waitReason}}
	}

	hasAnyProviderRow, exErr := h.JPRepo.ExistsByTransaksiMemberID(ctx, trx.ID)
	if exErr != nil {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: exErr.Error()}}
	}
	if hasAnyProviderRow {
		h.logf("STATUS-PAY tunggu callback karena trx_member sudah punya transaksi_provider refid=%s trx_member_id=%d provider=provider reason=wait_provider_callback", trx.RefID, trx.ID)
		return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "provider": "provider", "reason": "wait_provider_callback"}}
	}

	h.logf("STATUS-PAY tidak menemukan transaksi_provider refid=%s trx_member_id=%d => final failed", trx.RefID, trx.ID)
	failMsg := "transaksi provider belum ada"
	ketFailed, _ := helper.SafeMemberKeterangan("failed", failMsg)
	h.settleLikeCallback(ctx, trx, "failed", ketFailed, "", 0)
	st, _, whErr := h.sendFinalWebhook(ctx, trx, "failed", ketFailed, "", ketFailed, 0)
	return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "failed", "", failMsg, mapCallbackDelivery(st, whErr))}
}

func (h *MemberTrxService) statusPayRetryPendingProvider(ctx context.Context, trx *repository.TrxMemberFull, providerName string, row *model.JavapayTrxRow, triedProviders map[string]bool) *serviceResponse {
	if h == nil || trx == nil || row == nil {
		return nil
	}
	providerName = strings.TrimSpace(strings.ToLower(providerName))
	if providerName == "" {
		return nil
	}
	client := h.Clients[providerName]
	if client == nil {
		return nil
	}
	payQty := row.Qty
	if payQty <= 0 {
		payQty = trx.QtyProvider
	}
	if payQty <= 0 {
		payQty = trx.Qty
	}
	requestDest := strings.TrimSpace(row.Tujuan)
	if requestDest == "" {
		requestDest = strings.TrimSpace(trx.Tujuan)
	}
	providerCode := strings.TrimSpace(row.KodeProduk)
	if providerCode == "" {
		return nil
	}

	payReq := provider.PayRequest{
		Command: "PAY",
		Product: providerCode,
		Dest:    requestDest,
		Qty:     payQty,
		RefID:   trx.RefID,
	}
	applyProviderRetryRequestMeta(row.RequestMentah, &payReq)

	h.logf("STATUS-PAY %s PAY ulang refid=%s kode=%s mode=%s row_id=%d", providerName, trx.RefID, providerCode, payReq.Mode, row.ID)
	callCtx, cancel := context.WithTimeout(ctx, providerCallWindowForName(providerName))
	resp, callErr := client.Pay(callCtx, payReq)
	cancel()

	if resp == nil || !providerPayResponseHasProviderReply(resp) {
		msg := providerName + " status-pay PAY ulang menunggu balasan/callback provider"
		errText := ""
		if callErr != nil {
			errText = callErr.Error()
		}
		hs := 0
		body := ""
		if resp != nil {
			hs = resp.HTTPStatus
			body = resp.Body
		}
		if providerRowHasProviderReply(row) {
			ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
			h.logf("STATUS-PAY %s PAY ulang timeout setelah provider pernah balas; tahan pending refid=%s row_id=%d err=%v", providerName, trx.RefID, row.ID, callErr)
			return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", "", msg, nil)}
		}
		_ = h.JPRepo.UpdateResult(ctx, row.ID, db.UpdateResult{
			Pesan: &msg,
			ResponMentah: map[string]any{
				"status_pay_retry": true,
				"provider":         providerName,
				"http_status":      hs,
				"body":             body,
				"error":            errText,
			},
		})
		ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		h.logf("STATUS-PAY %s PAY ulang timeout refid=%s row_id=%d err=%v", providerName, trx.RefID, row.ID, callErr)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", "", msg, nil)}
	}

	if resp.RequestRaw != nil {
		_ = h.JPRepo.UpdateRequestMentah(ctx, row.ID, resp.RequestRaw)
	}
	if resp.HTTPStatus != 200 {
		msg := providerName + " status-pay PAY ulang menunggu balasan/callback provider"
		callErrText := ""
		if callErr != nil {
			callErrText = callErr.Error()
		}
		_ = h.JPRepo.UpdateResult(ctx, row.ID, db.UpdateResult{
			Pesan: &msg,
			ResponMentah: map[string]any{
				"status_pay_retry": true,
				"provider":         providerName,
				"http_status":      resp.HTTPStatus,
				"body":             resp.Body,
				"message":          resp.Message,
				"error":            callErrText,
			},
		})
		ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		h.logf("STATUS-PAY %s PAY ulang http error/no-response refid=%s row_id=%d http=%d err=%v tetap pending",
			providerName, trx.RefID, row.ID, resp.HTTPStatus, callErr)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", "", msg, nil)}
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
	if resp.Balance > 0 {
		upd.SaldoTerakhir = &resp.Balance
	}
	_ = h.JPRepo.UpdateResult(ctx, row.ID, upd)

	status := helper.ProviderResponseStatusString(providerName, upd.KodeRespon, upd.Pesan)
	h.logf("STATUS-PAY %s PAY ulang respon refid=%s http=%d status=%s body=%s", providerName, trx.RefID, resp.HTTPStatus, status, resp.Body)
	switch status {
	case "success":
		ket := strings.TrimSpace(resp.ProviderRef)
		if ket == "" {
			ket = "Transaksi berhasil (" + providerName + " status-pay retry)"
		}
		h.settleLikeCallback(ctx, trx, "success", ket, resp.ProviderRef, resp.Price)
		st, _, whErr := h.sendFinalWebhook(ctx, trx, "success", ket, resp.ProviderRef, resp.ProviderRef, resp.Price)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "success", derefStatusPayString(upd.KodeRespon), derefStatusPayString(upd.Pesan), mapCallbackDelivery(st, whErr))}
	case "failed":
		usedProvider, providerRowID, ferr := h.tryStatusPayFallback(ctx, trx, triedProviders)
		if ferr == nil && strings.TrimSpace(usedProvider) != "" {
			ketPending, _ := helper.SafeMemberKeterangan("pending", "wait_provider_status")
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
			return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "fallback_provider": usedProvider, "provider_row_id": providerRowID}}
		}
		if ferr != nil && !retryPendingNoFallbackProviderLeft(ferr) {
			ketPending, _ := helper.SafeMemberKeterangan("pending", ferr.Error())
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
			return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", derefStatusPayString(upd.KodeRespon), ferr.Error(), nil)}
		}
		ketFailed, _ := helper.SafeMemberKeterangan("failed", derefStatusPayString(upd.Pesan))
		h.settleLikeCallback(ctx, trx, "failed", ketFailed, strings.TrimSpace(resp.ProviderRef), 0)
		st, _, whErr := h.sendFinalWebhook(ctx, trx, "failed", ketFailed, strings.TrimSpace(resp.ProviderRef), ketFailed, 0)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "failed", derefStatusPayString(upd.KodeRespon), derefStatusPayString(upd.Pesan), mapCallbackDelivery(st, whErr))}
	default:
		ketPending, _ := helper.SafeMemberKeterangan("pending", derefStatusPayString(upd.Pesan))
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", derefStatusPayString(upd.KodeRespon), derefStatusPayString(upd.Pesan), nil)}
	}
}

func derefStatusPayString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func (h *MemberTrxService) handleStatusPayJavapayLookup(ctx context.Context, trx *repository.TrxMemberFull, _ map[string]bool) *serviceResponse {
	if h.JPClient == nil {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "javapay client nil"}}
	}

	prodJP, ok := h.resolveMappedProductForProvider(ctx, "javapay", trx.KodeProduk)
	if !ok {
		return &serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "mapping javapay tidak ditemukan untuk produk ini"}}
	}
	statusQty := trx.QtyProvider
	if statusQty <= 0 {
		statusQty = trx.Qty
	}
	statusRow, cErr := h.JPRepo.Create(ctx, model.JavapayTrxCreateIn{
		Provider:          "javapay",
		TransaksiMemberID: trx.ID,
		RefID:             trx.RefID,
		Perintah:          "STATUS-PAY",
		KodeProduk:        prodJP,
		Tujuan:            trx.Tujuan,
		Qty:               statusQty,
	}, map[string]any{"from": "member_status_pay"})
	if cErr != nil || statusRow == nil {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "failed create status-pay row"}}
	}

	jpResp, httpStatus, reqMentah, callErr := h.JPClient.Status(ctx, trx.RefID)
	if callErr != nil {
		hs := 0
		_ = h.JPRepo.UpdateResult(ctx, statusRow.ID, db.UpdateResult{
			HTTPStatus:   &hs,
			ResponMentah: map[string]any{"status": false, "error": callErr.Error()},
		})
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: callErr.Error()}}
	}
	_ = h.JPRepo.UpdateRequestMentah(ctx, statusRow.ID, reqMentah)

	var jpStatusRC string
	var jpStatusMsg string
	var jpStatusNoRef string
	var jpStatusPrice int64
	var jpStatusTrxID string
	if data, ok := jpResp["data"].(map[string]any); ok {
		jpStatusRC = helper.TrxToString(data["rc"])
		jpStatusMsg = helper.TrxToString(data["message"])
		jpStatusNoRef = helper.TrxToString(data["noreff"])
		jpStatusTrxID = helper.TrxToString(data["trxid"])
		if v, ok := data["price"]; ok {
			if p, ok2 := helper.TrxToInt64(v); ok2 {
				jpStatusPrice = p
			}
		}
	}

	upd := db.UpdateResult{HTTPStatus: &httpStatus, ResponMentah: helper.SummarizeJPResp(jpResp)}
	if jpStatusTrxID != "" {
		upd.TrxIDJavapay = &jpStatusTrxID
	}
	if jpStatusRC != "" {
		upd.KodeRespon = &jpStatusRC
	}
	if jpStatusMsg != "" {
		upd.Pesan = &jpStatusMsg
	}
	if jpStatusNoRef != "" {
		upd.NoReferensi = &jpStatusNoRef
	}
	if jpStatusPrice > 0 {
		upd.Harga = &jpStatusPrice
	}
	if upd.SaldoTerakhir == nil {
		if bal, ok := helper.ExtractSaldoTerakhirFromMsg(jpStatusMsg); ok {
			upd.SaldoTerakhir = &bal
			h.insertProviderSnapshot(ctx, "javapay", trx.ID, statusRow.ID, trx.RefID, bal, "member_status_pay_javapay", helper.SummarizeJPResp(jpResp))
		}
	}
	_ = h.JPRepo.UpdateResult(ctx, statusRow.ID, upd)

	if isJavapayStatusNotFoundMessage(jpStatusMsg) {
		h.logf("STATUS-PAY transaksi Javapay tidak ditemukan => treat failed and fallback/finalize refid=%s pesan=%s", trx.RefID, jpStatusMsg)
	}

	finalStatus := classifyJavapayResponseStatus(jpStatusRC, jpStatusMsg)
	if finalStatus == "unknown" {
		h.logf("STATUS-PAY hasil Javapay unknown => pending tanpa retry PAY refid=%s rc=%s msg=%s", trx.RefID, jpStatusRC, strings.TrimSpace(jpStatusMsg))
		ketPending, _ := helper.SafeMemberKeterangan("pending", "respons javapay tidak dikenali")
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", jpStatusRC, jpStatusMsg, nil)}
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, jpStatusMsg)
	ketDB := strings.TrimSpace(jpStatusNoRef)
	if ketDB == "" {
		ketDB = strings.TrimSpace(info.SN)
	}
	if ketDB == "" {
		ketDB = strings.TrimSpace(jpStatusTrxID)
	}
	if ketDB == "" {
		ketDB = ket
	}

	if finalStatus == "success" || finalStatus == "failed" {
		h.settleLikeCallback(ctx, trx, finalStatus, ketDB, info.Reff, jpStatusPrice)
		snOut := strings.TrimSpace(info.SN)
		if snOut == "" {
			snOut = strings.TrimSpace(jpStatusNoRef)
		}
		if snOut == "" {
			snOut = strings.TrimSpace(jpStatusTrxID)
		}
		if finalStatus == "failed" && snOut == "" {
			snOut = ket
		}
		st, _, whErr := h.sendFinalWebhook(ctx, trx, finalStatus, ket, info.Reff, snOut, jpStatusPrice)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, finalStatus, jpStatusRC, jpStatusMsg, mapCallbackDelivery(st, whErr))}
	}

	_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketDB, 0)
	return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, finalStatus, jpStatusRC, jpStatusMsg, nil)}
}
