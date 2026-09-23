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

func isDuplicatePayWhileProcessing(in trxmemberdto.TrxRequest, existing map[string]any) bool {
	if strings.TrimSpace(strings.ToUpper(in.Commands)) != "PAY" || existing == nil {
		return false
	}
	status := strings.TrimSpace(strings.ToLower(helper.TrxToString(existing["status"])))
	switch status {
	case "pending", "process", "proses":
		return true
	default:
		return false
	}
}

func (h *MemberTrxService) handlePayInqBranch(
	ctx context.Context,
	auth *repository.MemberAuth,
	in trxmemberdto.TrxRequest,
	billingNominal int64,
	productRuleSource string,
	chargeReceiverEligible bool,
	productJamBuka string,
	productJamTutup string,
) serviceResponse {
	retrySameRefReopened := false
	retrySameRefTrxID := int64(0)
	retrySameRefSkipProviders := map[string]bool{}
	existing, err := h.MemberRepo.GetTransaksiMemberByRef(ctx, auth.MemberID, in.RefID)
	if err != nil {
		h.logf("PAY/INQ gagal ambil transaksi berdasarkan ref cmd=%s refid=%s err=%v", in.Commands, in.RefID, err)
		return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: err.Error()}}
	}
	if existing != nil {
		existingID, _ := existing["id"].(int64)
		if markerErr := h.attachSMPAYTransactionSource(ctx, auth, existingID, in); markerErr != nil {
			return serviceResponse{Err: markerErr}
		}
		if strings.TrimSpace(strings.ToUpper(in.Commands)) == "PAY" && canRetryFailedPaySameRefID(existing, time.Now()) {
			usedAttempts, exErr := h.JPRepo.ListByRefID(ctx, in.RefID)
			if exErr != nil {
				h.logf("PAY retry same-ref cek provider terdahulu gagal refid=%s trx_member_id=%d err=%v", in.RefID, existingID, exErr)
				return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: exErr.Error()}}
			}
			ketRetry, _ := helper.SafeMemberKeterangan("pending", "retry pay same ref after failed")
			if existingID > 0 {
				if upErr := h.MemberRepo.ForceReopenFailedToPending(ctx, existingID, ketRetry); upErr != nil && upErr != sql.ErrNoRows {
					h.logf("PAY retry same-ref gagal reopen pending refid=%s trx_member_id=%d err=%v", in.RefID, existingID, upErr)
					return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: upErr.Error()}}
				}
				retrySameRefReopened = true
				retrySameRefTrxID = existingID
				// Bank hanya SMB — jangan skip SMB saat retry, biar bisa kirim ulang
				if h.isBankH2HProduct(ctx, in.Product) {
					retrySameRefSkipProviders = map[string]bool{}
				} else {
					retrySameRefSkipProviders = providerAttemptTriedKeys(usedAttempts)
				}
			}
			h.logf("PAY retry same-ref diizinkan refid=%s candidate_terdahulu=%v", in.RefID, retrySameRefSkipProviders)
		} else {
			if strings.TrimSpace(strings.ToUpper(in.Commands)) == "PAY" {
				if existingID > 0 {
					if updatedAt, ok := existing["diperbarui_pada"].(time.Time); ok && !updatedAt.IsZero() {
						remaining := retrySameRefCooldown - time.Since(updatedAt)
						if remaining > 0 && strings.TrimSpace(strings.ToLower(helper.TrxToString(existing["status"]))) == "failed" {
							h.logf("PAY retry same-ref ditolak refid=%s trx_member_id=%d alasan=cooldown_belum_lewat sisa_ms=%d", in.RefID, existingID, remaining.Milliseconds())
						}
					}
				}
			}
			if isDuplicatePayWhileProcessing(in, existing) {
				h.logf("PAY duplicate same-ref pakai status existing karena transaksi pertama masih jalan refid=%s trx_member_id=%d", in.RefID, existingID)
				return serviceResponse{Body: trxmemberdto.MapExistingResponse(existing)}
			}
			h.logf("PAY/INQ transaksi existing ditemukan cmd=%s refid=%s", in.Commands, in.RefID)
			return serviceResponse{Body: trxmemberdto.MapExistingResponse(existing)}
		}
	}

	if strings.TrimSpace(strings.ToUpper(in.Commands)) == "PAY" && !helper.IsH2HProductAvailableForNow(in.Product, productJamBuka, productJamTutup) {
		h.logf("PAY produk sedang offline refid=%s produk=%s jam_buka=%s jam_tutup=%s", in.RefID, in.Product, productJamBuka, productJamTutup)
		return serviceResponse{Body: trxmemberdto.MapBusinessStatusResponse(in.RefID, 3, "produk ini tidak tersedia pada jam ini, silakan coba lagi nanti")}
	}

	fee := int64(0)
	feeSource := ""
	if feeByCategory, src, fErr := h.MemberRepo.GetMemberFeeByH2HCategory(ctx, auth.MemberID, in.Product, billingNominal); fErr == nil && src != "" {
		fee = feeByCategory
		feeSource = src
	} else {
		if fErr != nil {
			h.logf("PAY/INQ gagal lookup fee kategori member_id=%d produk=%s err=%v", auth.MemberID, in.Product, fErr)
			return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "gagal membaca fee kategori member"}}
		}
		categoryName, catErr := h.MemberRepo.ClassifyH2HFeeCategoryBySKU(ctx, in.Product)
		if catErr != nil || strings.TrimSpace(categoryName) == "" {
			categoryName = helper.ClassifyH2HFeeCategory(in.Product)
		}
		h.logf("PAY/INQ fee kategori belum diset member_id=%d produk=%s kategori=%s", auth.MemberID, in.Product, categoryName)
		return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "fee kategori member belum diset"}}
	}

	var biayaPerkiraan int64
	providerOrderNominal := billingNominal
	chargeReceiverApplied := false
	if in.Commands == "PAY" {
		biayaPerkiraan = billingNominal + fee
		if auth.ChargeReceiver && chargeReceiverEligible {
			chargeReceiverApplied = true
			biayaPerkiraan = billingNominal
			providerOrderNominal = billingNominal - fee
			if providerOrderNominal <= 0 {
				return serviceResponse{Err: &ServiceError{Kind: ErrBadRequest, Message: "nominal setelah potong fee harus > 0"}}
			}
		}
	}

	trxMemberID := retrySameRefTrxID
	if !retrySameRefReopened {
		created := false
		trxMemberID, created, err = h.MemberRepo.CreateTransaksiMember(ctx, auth.MemberID, in.RefID, in.Commands, in.Product, in.Dest, in.Qty, providerOrderNominal, chargeReceiverApplied, biayaPerkiraan)
		if err != nil {
			h.logf("CREATE trx_member gagal cmd=%s refid=%s err=%v", in.Commands, in.RefID, err)
			if strings.Contains(strings.ToLower(err.Error()), "saldo tidak cukup") {
				return serviceResponse{Body: map[string]any{
					"ok":      true,
					"refid":   in.RefID,
					"status":  3,
					"message": "saldo tidak cukup",
				}}
			}
			return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: err.Error()}}
		}
		if !created {
			existingAfterCreate, exErr := h.MemberRepo.GetTransaksiMemberByRef(ctx, auth.MemberID, in.RefID)
			if exErr != nil {
				h.logf("PAY/INQ gagal ambil existing setelah conflict create cmd=%s refid=%s trx_member_id=%d err=%v", in.Commands, in.RefID, trxMemberID, exErr)
				return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: exErr.Error()}}
			}
			if existingAfterCreate != nil {
				if existingID, _ := existingAfterCreate["id"].(int64); existingID > 0 {
					if markerErr := h.attachSMPAYTransactionSource(ctx, auth, existingID, in); markerErr != nil {
						return serviceResponse{Err: markerErr}
					}
				}
				if isDuplicatePayWhileProcessing(in, existingAfterCreate) {
					h.logf("PAY dedupe same-ref pakai status existing karena transaksi pertama masih jalan refid=%s trx_member_id=%d", in.RefID, trxMemberID)
					return serviceResponse{Body: trxmemberdto.MapExistingResponse(existingAfterCreate)}
				}
				h.logf("PAY/INQ dedupe same-ref, pakai transaksi existing cmd=%s refid=%s trx_member_id=%d", in.Commands, in.RefID, trxMemberID)
				return serviceResponse{Body: trxmemberdto.MapExistingResponse(existingAfterCreate)}
			}
			h.logf("PAY/INQ conflict create tapi existing tidak ditemukan cmd=%s refid=%s trx_member_id=%d", in.Commands, in.RefID, trxMemberID)
			return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: "gagal membaca transaksi existing"}}
		}
	} else {
		h.logf("PAY retry same-ref lanjut pakai trx_member existing cmd=%s refid=%s trx_member_id=%d", in.Commands, in.RefID, trxMemberID)
	}
	if markerErr := h.attachSMPAYTransactionSource(ctx, auth, trxMemberID, in); markerErr != nil {
		return serviceResponse{Err: markerErr}
	}
	_ = h.MemberRepo.UpdateTransaksiMemberFee(ctx, trxMemberID, fee)

	h.logf("CREATE trx_member id=%d cmd=%s refid=%s biaya_perkiraan=%d fee=%d sumber_fee=%s produk=%s qty=%d billing_nominal=%d rule_produk=%s",
		trxMemberID, in.Commands, in.RefID, biayaPerkiraan, fee, feeSource, in.Product, in.Qty, billingNominal, productRuleSource)
	if in.Commands == "PAY" && auth.ChargeReceiver && chargeReceiverEligible {
		h.logf("CHARGE_RECEIVER aktif refid=%s produk=%s billing_nominal=%d fee=%d hold_member=%d nominal_provider=%d", in.RefID, in.Product, billingNominal, fee, biayaPerkiraan, providerOrderNominal)
	}

	// Hold saldo di-handle oleh DB trigger (trg_enforce_saldo_on_status).
	// Trigger otomatis DEBIT saat INSERT transaksi_member dengan status pending.
	// Kalau saldo tidak cukup, trigger raise exception → INSERT gagal.
	h.logf("HOLD cmd=PAY refid=%s nominal=%d (db trigger)", in.RefID, biayaPerkiraan)

	providerReq := in
	if in.Commands == "PAY" && auth.ChargeReceiver && chargeReceiverEligible {
		providerReq.Qty = providerOrderNominal
	}

	used, _, pErr := h.routeProviderForPayInqSkipping(ctx, trxMemberID, providerReq, providerOrderNominal, retrySameRefSkipProviders)

	if pErr != nil {
		h.logf("provider gagal cmd=%s refid=%s provider_terpakai=%s err=%v", in.Commands, in.RefID, used, pErr)

		if shouldKeepPendingOnProviderFailure(pErr) {
			waitReason := "wait_provider_callback"
			switch used {
			case "yuscom":
				waitReason = "wait_yuscom_callback"
			case "talentapay":
				waitReason = "wait_talentapay_callback"
			case "multikom":
				waitReason = "wait_multikom_callback"
			case "javapay":
				waitReason = "wait_javapay_reconcile"
			case "sagaramobile":
				waitReason = "wait_sagaramobile_callback"
			case "minions":
				waitReason = "wait_minions_callback"
			}
			ket, _ := helper.SafeMemberKeterangan("pending", waitReason)
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trxMemberID, "pending", ket, 0)
			h.logf("provider timeout/system issue => tetap pending refid=%s provider=%s reason=%s", in.RefID, used, waitReason)
			h.logf("AUDIT pending_timeout trx_member_id=%d refid=%s cmd=%s provider=%s reason=%s billing_nominal=%d nominal_provider=%d biaya_perkiraan=%d fee=%d err=%v", trxMemberID, in.RefID, in.Commands, used, waitReason, billingNominal, providerOrderNominal, biayaPerkiraan, fee, pErr)
			return serviceResponse{Body: trxmemberdto.MapTransaksiMemberResponse(trxMemberID, in.RefID, "pending", providerOrderNominal, biayaPerkiraan, fee, chargeReceiverApplied, ket, used)}
		}

		// Refund di-handle oleh DB trigger saat status di-update ke 'failed'.
		errMsg := strings.TrimSpace(pErr.Error())
		if strings.Contains(strings.ToLower(strings.TrimSpace(pErr.Error())), "tidak ada provider eligible") {
			errMsg = "produk kehabisan stok, coba lagi nanti"
		}
		ket, _ := helper.SafeMemberKeterangan("failed", errMsg)
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trxMemberID, "failed", ket, 0)
		if in.Commands == "PAY" {
			if trxFull, tErr := h.MemberRepo.GetTransaksiMemberByID(ctx, trxMemberID); tErr == nil && trxFull != nil {
				_, _, _ = h.sendFinalWebhook(ctx, trxFull, "failed", ket, "", "", 0)
			}
		}
		return serviceResponse{Err: &ServiceError{Kind: ErrUpstream, Message: ket}}
	}

	if trxFinal, tErr := h.MemberRepo.GetTransaksiMemberByID(ctx, trxMemberID); tErr == nil && trxFinal != nil {
		if strings.EqualFold(strings.TrimSpace(trxFinal.Status), "success") || strings.EqualFold(strings.TrimSpace(trxFinal.Status), "failed") {
			ketFinal := ""
			if existingNow, exErr := h.MemberRepo.GetTransaksiMemberByRef(ctx, auth.MemberID, in.RefID); exErr == nil && existingNow != nil {
				ketFinal = helper.TrxToString(existingNow["keterangan"])
			}
			return serviceResponse{Body: trxmemberdto.MapTransaksiMemberResponse(trxFinal.ID, trxFinal.RefID, trxFinal.Status, trxFinal.QtyProvider, trxFinal.BiayaPerkiraan, trxFinal.FeeMemberRp, trxFinal.ChargeReceiverApplied, ketFinal, used)}
		}
		if strings.EqualFold(strings.TrimSpace(in.Commands), "PAY") && strings.TrimSpace(used) != "" {
			if providerRow, rowErr := h.JPRepo.GetLatestByRefIDProvider(ctx, in.RefID, used); rowErr != nil {
				h.logf("PAY immediate success lookup provider row gagal refid=%s provider=%s err=%v", in.RefID, used, rowErr)
			} else if ok, _, msg, price, noreff := providerRowSuccessState(used, providerRow); ok {
				ket, info := helper.SafeMemberKeterangan("success", msg)
				ketDB := strings.TrimSpace(noreff)
				if ketDB == "" {
					ketDB = strings.TrimSpace(info.SN)
				}
				if ketDB == "" {
					ketDB = ket
				}

				h.logf("PAY immediate provider success settle refid=%s provider=%s provider_row_id=%d price=%d", in.RefID, used, providerRow.ID, price)
				h.settleLikeCallback(ctx, trxFinal, "success", ketDB, info.Reff, price)

				whKet, whProviderRef, whSN, whPrice := providerRowWebhookInfo(used, providerRow, "success")
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
				st, _, whErr := h.sendFinalWebhook(ctx, trxFinal, "success", whKet, whProviderRef, whSN, whPrice)
				h.logf("PAY immediate provider success webhook refid=%s provider=%s http=%d err=%v", in.RefID, used, st, whErr)

				return serviceResponse{Body: trxmemberdto.MapTransaksiMemberResponse(trxFinal.ID, trxFinal.RefID, "success", trxFinal.QtyProvider, trxFinal.BiayaPerkiraan, trxFinal.FeeMemberRp, trxFinal.ChargeReceiverApplied, ketDB, used)}
			}
		}
	}

	ket, _ := helper.SafeMemberKeterangan("pending", "")
	_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trxMemberID, "pending", ket, 0)
	return serviceResponse{Body: trxmemberdto.MapTransaksiMemberResponse(trxMemberID, in.RefID, "pending", providerOrderNominal, biayaPerkiraan, fee, chargeReceiverApplied, ket, used)}
}
