package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (h *MemberTrxService) handleStatusPayMissingMemberAsPay(ctx context.Context, auth *repository.MemberAuth, in trxmemberdto.TrxRequest) serviceResponse {
	payReq := in
	payReq.Commands = "PAY"

	var billingNominal int64
	productRuleSource := "legacy"
	chargeReceiverEligible := false

	rule, rErr := h.MemberRepo.GetProdukPricingRuleBySKU(ctx, payReq.Product)
	if rErr != nil {
		h.logf("STATUS-PAY create PAY gagal lookup rule produk=%s refid=%s err=%v", payReq.Product, payReq.RefID, rErr)
		return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "gagal baca master produk"}}
	}

	if rule == nil {
		return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "produk tidak ditemukan"}}
	}

	switch rule.TipeHarga {
	case "FIXED":
		if payReq.Qty != 1 {
			return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "qty untuk produk FIXED harus 1"}}
		}
		if rule.Nominal == nil || *rule.Nominal <= 0 {
			return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "konfigurasi produk FIXED belum valid (nominal kosong)"}}
		}
		billingNominal = *rule.Nominal
		productRuleSource = "status_pay_missing_member_master_fixed"
	case "OPEN_AMOUNT":
		if payReq.Qty <= 0 {
			return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "qty untuk OPEN_AMOUNT harus > 0"}}
		}
		billingNominal = payReq.Qty
		productRuleSource = "status_pay_missing_member_master_open_amount"
		chargeReceiverEligible = isChargeReceiverEligibleProduct(payReq.Product, rule.TipeHarga, rule.KategoriNama, rule.BrandNama)
	default:
		return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "tipe_harga produk tidak didukung"}}
	}

	return h.handlePayInqBranch(ctx, auth, payReq, billingNominal, productRuleSource, chargeReceiverEligible, rule.JamBuka, rule.JamTutup)
}

func (h *MemberTrxService) handleStatusPayBranch(ctx context.Context, auth *repository.MemberAuth, in trxmemberdto.TrxRequest) (*serviceResponse, bool) {
	if in.Commands != "STATUS-PAY" {
		return nil, false
	}

	trxMap, err := h.MemberRepo.GetTransaksiMemberByRef(ctx, auth.MemberID, in.RefID)
	if err != nil {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: err.Error()}}, true
	}
	if trxMap == nil {
		h.logf("STATUS-PAY refid tidak ada di transaksi_member => buat PAY baru refid=%s produk=%s tujuan=%s qty=%d", in.RefID, in.Product, in.Dest, in.Qty)
		out := h.handleStatusPayMissingMemberAsPay(ctx, auth, in)
		return &out, true
	}

	trxID, ok := helper.TrxToInt64(trxMap["id"])
	if !ok || trxID <= 0 {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "invalid trx id"}}, true
	}
	if markerErr := h.attachSMPAYTransactionSource(ctx, auth, trxID, in); markerErr != nil {
		return &serviceResponse{Err: markerErr}, true
	}

	trx, err := h.MemberRepo.GetTransaksiMemberByID(ctx, trxID)
	if err != nil || trx == nil {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "trx not found by id"}}, true
	}

	rows, err := h.loadStatusPayProviderRows(ctx, in.RefID)
	if err != nil {
		return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: err.Error()}}, true
	}

	providerStates := rows.providerState
	hasProviderState := rows.hasAny
	adminFinalStatus, adminFinalKet, adminFinal := detectAdminManualFinalStatus(trx.Status, helper.TrxToString(trxMap["keterangan"]))
	if adminFinal {
		price := int64(0)
		if adminFinalStatus == "success" {
			price = trx.BiayaAktual
			if price <= 0 {
				price = trx.BiayaPerkiraan
			}
		}
		h.logf("STATUS-PAY hormati final admin refid=%s trx_id=%d status=%s ket=%s", trx.RefID, trx.ID, adminFinalStatus, strings.TrimSpace(adminFinalKet))
		st, _, whErr := h.sendFinalWebhook(ctx, trx, adminFinalStatus, adminFinalKet, "", "", price)
		return &serviceResponse{Body: trxmemberdto.MapAlreadyFinalResponse(trx.RefID, adminFinalStatus, hasProviderState, mapCallbackDelivery(st, whErr))}, true
	}

	for _, ps := range providerStates {
		if ps.row == nil {
			continue
		}
		psRC := ""
		psMsg := ""
		if ps.row.KodeRespon != nil {
			psRC = *ps.row.KodeRespon
		}
		if ps.row.Pesan != nil {
			if len(*ps.row.Pesan) > 80 {
				psMsg = (*ps.row.Pesan)[:80]
			} else {
				psMsg = *ps.row.Pesan
			}
		}
		psState := helper.ProviderResponseStateOf(ps.name, psRC, psMsg)
		h.logf("STATUS-PAY providerState refid=%s provider=%s id=%d rc=%s state=%s pesan=%s", trx.RefID, ps.name, ps.row.ID, psRC, psState, psMsg)
	}
	if strings.EqualFold(strings.TrimSpace(trx.Status), "success") || strings.EqualFold(strings.TrimSpace(trx.Status), "failed") {
		h.logf("STATUS-PAY hormati member final lebih dulu refid=%s trx_id=%d status=%s", trx.RefID, trx.ID, trx.Status)
		return h.buildAlreadyFinalStatusPayResponse(ctx, trx, rows.p24, rows.jp, rows.ys, rows.tl, rows.mk, rows.sg, rows.mn, rows.tr, rows.aj, rows.gm, rows.sm, rows.lb), true
	}

	successProvider, successRow, rc, msg, price, noreff := pickLatestProviderSuccess(providerStates)
	h.logf("STATUS-PAY pickSuccess refid=%s found=%v provider=%s", trx.RefID, successRow != nil, successProvider)
	if successRow != nil {
		ket, info := helper.SafeMemberKeterangan("success", msg)
		ketDB := strings.TrimSpace(noreff)
		if ketDB == "" {
			ketDB = strings.TrimSpace(info.SN)
		}
		if ketDB == "" {
			ketDB = ket
		}

		h.logf("STATUS-PAY reconcile provider success refid=%s provider=%s rc=%s", trx.RefID, successProvider, rc)
		if strings.EqualFold(strings.TrimSpace(trx.Status), "failed") {
			biayaAktual := trx.BiayaPerkiraan
			if biayaAktual <= 0 {
				biayaAktual = trx.BiayaAktual
			}
			if biayaAktual <= 0 {
				h.logf("STATUS-PAY reconcile gagal karena biaya invalid refid=%s trx_id=%d provider=%s", trx.RefID, trx.ID, successProvider)
				return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "invalid trx cost for reconcile"}}, true
			}

			hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
			if err := h.MemberRepo.ForceReconcileFailedToSuccess(ctx, trx.ID, ketDB, biayaAktual, price, hargaMember); err != nil {
				if err == repository.ErrInsufficientBalance {
					h.logf("STATUS-PAY reconcile gagal karena saldo member tidak cukup refid=%s trx_id=%d provider=%s biaya=%d", trx.RefID, trx.ID, successProvider, biayaAktual)
					return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "saldo member tidak cukup untuk reconcile transaksi sukses"}}, true
				}
				if err != sql.ErrNoRows {
					h.logf("STATUS-PAY reconcile gagal update failed->success refid=%s trx_id=%d provider=%s err=%v", trx.RefID, trx.ID, successProvider, err)
					return &serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: err.Error()}}, true
				}
			}
			trx.Status = "success"
			trx.BiayaAktual = biayaAktual
			trx.HargaJavapay = price
			trx.HargaMember = effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
		} else {
			h.settleLikeCallback(ctx, trx, "success", ketDB, info.Reff, price)
		}
		whKet, whProviderRef, whSN, whPrice := providerRowWebhookInfo(successProvider, successRow, "success")
		if whKet == "" {
			whKet = ket
		}
		if whProviderRef == "" {
			whProviderRef = strings.TrimSpace(info.Reff)
		}
		if whSN == "" {
			whSN = strings.TrimSpace(info.SN)
		}
		if whSN == "" {
			whSN = strings.TrimSpace(noreff)
		}
		if whPrice <= 0 {
			whPrice = price
		}
		st, _, whErr := h.sendFinalWebhook(ctx, trx, "success", whKet, whProviderRef, whSN, whPrice)

		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "success", rc, msg, mapCallbackDelivery(st, whErr))}, true
	}

	triedProviders := providerAttemptTriedKeys(rows.allAttempts)

	pendingProvider, pendingRow, rc, msg := pickLatestProviderPending(providerStates)
	if pendingProvider != "" {
		if out := h.statusPayRetryPendingProvider(ctx, trx, pendingProvider, pendingRow, triedProviders); out != nil {
			return out, true
		}
		if pendingProvider == "smb" && isStaleSMBCheckOnlyPending(pendingRow, time.Now()) {
			h.logf("STATUS-PAY stale SMB cek pending refid=%s row_id=%d age_ms=%d", trx.RefID, pendingRow.ID, time.Since(pendingRow.DibuatPada).Milliseconds())
			ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
			return &serviceResponse{Body: map[string]any{
				"ok":       true,
				"refid":    trx.RefID,
				"status":   "pending",
				"provider": "smb",
				"reason":   "stale_smb_check_waiting_no_status_pay_dispatch",
			}}, true
		}
		ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		ageMin := 0
		if pendingRow != nil && !pendingRow.DibuatPada.IsZero() {
			ageMin = int(time.Since(pendingRow.DibuatPada).Minutes())
		}
		h.logf("STATUS-PAY tunggu provider pending refid=%s provider=%s rc=%s age_min=%d reason=no_fallback_for_pending", trx.RefID, pendingProvider, rc, ageMin)
		return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", rc, msg, nil)}, true
	}

	failedProvider, failedRow, retryableFailed, failedRC, failedMsg, failedPrice, failedNoRef := pickLatestProviderFailure(providerStates)
	if failedRow != nil {
		if strings.EqualFold(strings.TrimSpace(trx.Status), "failed") || strings.EqualFold(strings.TrimSpace(trx.Status), "success") {
			return h.buildAlreadyFinalStatusPayResponse(ctx, trx, rows.p24, rows.jp, rows.ys, rows.tl, rows.mk, rows.sg, rows.mn, rows.tr, rows.aj, rows.gm, rows.sm, rows.lb), true
		}
		if retryableFailed {
			if ok, reason := h.allExistingRouteAttemptsFinalFailed(ctx, trx.ID, trx.RefID); ok {
				h.logf("STATUS-PAY provider final failed dan semua route gagal refid=%s provider=%s reason=%s", trx.RefID, failedProvider, reason)
				retryableFailed = false
			} else {
				h.logf("STATUS-PAY provider failed retryable tapi belum semua route final refid=%s provider=%s reason=%s", trx.RefID, failedProvider, reason)
				ketPending, _ := helper.SafeMemberKeterangan("pending", reason)
				_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
				return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "pending", failedRC, failedMsg, nil)}, true
			}
		}
		if !retryableFailed {
			ketFailed, infoFailed := helper.SafeMemberKeterangan("failed", failedMsg)
			ketDB := strings.TrimSpace(failedNoRef)
			if ketDB == "" {
				ketDB = ketFailed
			}
			h.settleLikeCallback(ctx, trx, "failed", ketDB, infoFailed.Reff, failedPrice)
			failedSN := strings.TrimSpace(infoFailed.SN)
			if failedSN == "" {
				failedSN = ketFailed
			}
			st, _, whErr := h.sendFinalWebhook(ctx, trx, "failed", ketFailed, strings.TrimSpace(infoFailed.Reff), failedSN, failedPrice)
			return &serviceResponse{Body: trxmemberdto.MapStatusResponse(trx.RefID, "failed", failedRC, failedMsg, mapCallbackDelivery(st, whErr))}, true
		}
	}

	if trx.Status == "success" || trx.Status == "failed" {
		return h.buildAlreadyFinalStatusPayResponse(ctx, trx, rows.p24, rows.jp, rows.ys, rows.tl, rows.mk, rows.sg, rows.mn, rows.tr, rows.aj, rows.gm, rows.sm, rows.lb), true
	}

	if isTalentaProductCode(trx.KodeProduk) {
		if rows.tl != nil {
			return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "provider": "talentapay", "reason": "wait_talentapay_callback"}}, true
		}
		if rows.ys != nil {
			return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "provider": "yuscom", "reason": "wait_yuscom_callback"}}, true
		}
		if rows.mk != nil {
			return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "provider": "multikom", "reason": "wait_multikom_callback"}}, true
		}
		return &serviceResponse{Body: map[string]any{"ok": true, "refid": trx.RefID, "status": "pending", "provider": "provider", "reason": "wait_provider_callback"}}, true
	}

	if rows.jp == nil {
		return h.handleStatusPayWithoutJavapayRow(ctx, trx, rows, triedProviders), true
	}

	return h.handleStatusPayJavapayLookup(ctx, trx, triedProviders), true
}
