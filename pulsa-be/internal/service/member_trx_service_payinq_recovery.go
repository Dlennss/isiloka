package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pulsa2/db"
	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/model"
)

func restoreProviderAttemptMeta(raw any) (specialCode string, mode string) {
	m, ok := raw.(map[string]any)
	if !ok || m == nil {
		return "", ""
	}

	mode = strings.ToUpper(strings.TrimSpace(helper.TrxToString(m["mode"])))
	specialCode = strings.ToUpper(strings.TrimSpace(helper.TrxToString(m["special_code"])))
	if specialCode != "" {
		return specialCode, mode
	}

	productSent := strings.TrimSpace(helper.TrxToString(m["product_sent"]))
	if idx := strings.Index(productSent, ":"); idx > 0 {
		specialCode = strings.ToUpper(strings.TrimSpace(productSent[:idx]))
	}
	return specialCode, mode
}

func normalizeResendProviderKodeProduk(providerName, kodeProduk, specialCode string) string {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	kodeProduk = strings.TrimSpace(kodeProduk)
	specialCode = strings.ToUpper(strings.TrimSpace(specialCode))
	if kodeProduk == "" {
		return ""
	}
	if specialCode == "" {
		return kodeProduk
	}
	prefix := specialCode + ":"
	upperKode := strings.ToUpper(kodeProduk)
	if strings.HasPrefix(upperKode, prefix) {
		return strings.TrimSpace(kodeProduk[len(prefix):])
	}
	if providerName == "loketbayar" {
		parts := strings.Split(strings.TrimSpace(kodeProduk), ":")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[len(parts)-1])
		}
	}
	return kodeProduk
}

func resolveProviderAttemptDest(providerName, trxDest, rowDest string) string {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	rowDest = strings.TrimSpace(rowDest)
	trxDest = strings.TrimSpace(trxDest)
	if providerName == "loketbayar" && rowDest != "" {
		return rowDest
	}
	if rowDest != "" {
		return rowDest
	}
	return trxDest
}

func (h *MemberTrxService) RecoverPendingRef(ctx context.Context, memberID int64, refID string) (string, int64, error) {
	if h == nil || h.MemberRepo == nil || h.JPRepo == nil {
		return "", 0, fmt.Errorf("service belum siap")
	}

	trxMap, err := h.MemberRepo.GetTransaksiMemberByRef(ctx, memberID, refID)
	if err != nil {
		return "", 0, err
	}
	if trxMap == nil {
		return "", 0, fmt.Errorf("transaksi tidak ditemukan")
	}
	if !strings.EqualFold(strings.TrimSpace(helper.TrxToString(trxMap["status"])), "pending") {
		return "", 0, fmt.Errorf("transaksi bukan pending")
	}
	trxID, _ := trxMap["id"].(int64)
	if trxID <= 0 {
		return "", 0, fmt.Errorf("trx id tidak valid")
	}
	trx, err := h.MemberRepo.GetTransaksiMemberByID(ctx, trxID)
	if err != nil {
		return "", 0, err
	}
	if trx == nil {
		return "", 0, fmt.Errorf("transaksi penuh tidak ditemukan")
	}

	triedAttempts, err := h.JPRepo.ListByRefID(ctx, refID)
	if err != nil {
		return "", 0, err
	}
	triedProviders := providerAttemptTriedKeys(triedAttempts)

	// Cek apakah ada provider yang sudah sukses — kalau ada, settle member jika masih pending
	for _, att := range triedAttempts {
		if att == nil {
			continue
		}
		attRow := &model.JavapayTrxRow{
			ID: att.ID, KodeRespon: att.KodeRespon, Pesan: att.Pesan,
			NoReferensi: att.NoReferensi, Harga: att.Harga,
		}
		provider := strings.ToLower(strings.TrimSpace(att.Provider))
		if ok, _, msg, price, noreff := providerRowSuccessState(provider, attRow); ok {
			// Settle member jika masih pending
			if strings.EqualFold(strings.TrimSpace(trx.Status), "pending") {
				ketDB := strings.TrimSpace(noreff)
				if ketDB == "" {
					ket, info := helper.SafeMemberKeterangan("success", msg)
					ketDB = strings.TrimSpace(info.SN)
					if ketDB == "" {
						ketDB = ket
					}
				}
				h.settleLikeCallback(ctx, trx, "success", ketDB, noreff, price)
				_, _, _ = h.sendFinalWebhook(ctx, trx, "success", "Transaksi berhasil", noreff, noreff, price)
				h.logf("RECOVER settle existing success refid=%s provider=%s price=%d", refID, provider, price)
			}
			return provider, att.ID, nil
		}
	}

	// Recheck: untuk provider pending >5 menit, cek status:
	// - Javapay: pakai Status API
	// - Provider lain (kecuali SMB): PAY ulang dengan refid sama → respon idempoten "sdh pernah...Sukses/Gagal"
	for _, att := range triedAttempts {
		if att == nil {
			continue
		}
		provider := strings.ToLower(strings.TrimSpace(att.Provider))
		if provider == "smb" {
			continue // SMB punya mekanisme sendiri (DispatchPayOnly)
		}
		pending, _, _ := providerRowPendingState(provider, att)
		if !pending {
			continue
		}
		if time.Since(att.DibuatPada) < 5*time.Minute {
			continue
		}

		// Javapay: pakai Status API khusus
		if provider == "javapay" && h.JPClient != nil {
			statusCtx, statusCancel := context.WithTimeout(ctx, 15*time.Second)
			jpResp, _, _, statusErr := h.JPClient.Status(statusCtx, trx.RefID)
			statusCancel()
			if statusErr == nil && jpResp != nil {
				if data, ok := jpResp["data"].(map[string]any); ok {
					jpRC := helper.TrxToString(data["rc"])
					jpMsg := helper.TrxToString(data["message"])
					jpNoRef := helper.TrxToString(data["noreff"])
					jpPrice := int64(0)
					if v, ok2 := helper.TrxToInt64(data["price"]); ok2 {
						jpPrice = v
					}
					finalStatus := helper.ClassifyJavapayResponseStatus(jpRC, jpMsg)
					if finalStatus == helper.ProviderResponseSuccess {
						h.logf("RECHECK javapay STATUS sukses refid=%s rc=%s price=%d", refID, jpRC, jpPrice)
						// Update provider row supaya konsisten dengan member
						hs200 := 200
						_ = h.JPRepo.UpdateResult(ctx, att.ID, db.UpdateResult{
							HTTPStatus: &hs200, KodeRespon: helper.PtrString(jpRC),
							Pesan: helper.PtrString(jpMsg), Harga: helper.PtrI64(jpPrice),
							NoReferensi: helper.PtrString(jpNoRef),
						})
						ketDB := strings.TrimSpace(jpNoRef)
						if ketDB == "" {
							ketDB = strings.TrimSpace(jpMsg)
						}
						h.settleLikeCallback(ctx, trx, "success", ketDB, jpNoRef, jpPrice)
						_, _, _ = h.sendFinalWebhook(ctx, trx, "success", "Transaksi berhasil", jpNoRef, jpNoRef, jpPrice)
						return "javapay", att.ID, nil
					}
					if finalStatus == helper.ProviderResponseFailed {
						h.logf("RECHECK javapay STATUS gagal refid=%s rc=%s msg=%s", refID, jpRC, jpMsg)
						hs200 := 200
						_ = h.JPRepo.UpdateResult(ctx, att.ID, db.UpdateResult{
							HTTPStatus: &hs200, KodeRespon: helper.PtrString(jpRC),
							Pesan: helper.PtrString(jpMsg),
						})
						break // lanjut ke fallback
					}
					h.logf("RECHECK javapay STATUS pending refid=%s rc=%s msg=%s", refID, jpRC, jpMsg)
				}
			}
			break
		}

		specialCode, mode := restoreProviderAttemptMeta(att.RequestMentah)
		recheckIn := trxmemberdto.TrxRequest{
			Commands: "PAY", Product: trx.KodeProduk, Dest: resolveProviderAttemptDest(provider, trx.Tujuan, att.Tujuan),
			Qty: trx.QtyProvider, RefID: trx.RefID,
		}
		if recheckIn.Qty <= 0 {
			recheckIn.Qty = trx.Qty
		}
		attempt := providerRouteAttempt{
			Name: provider, Src: "recheck_pending_same_refid",
			ProdukSKUSnapshot: att.ProdukSKUSnapshot, ProdukProviderMapID: att.ProdukProviderMapID,
			KodeProduk:  normalizeResendProviderKodeProduk(provider, att.KodeProduk, specialCode),
			SpecialCode: specialCode,
			Mode:        mode,
		}

		recheckCtx, recheckCancel := context.WithTimeout(ctx, 20*time.Second)
		usedP, newRowID, callErr := h.callProviderOnly(recheckCtx, trx.ID, recheckIn, attempt, "recheck_pending_same_refid")
		recheckCancel()

		// Cek respon row baru dulu — mungkin "sdh pernah...Sukses" meski callErr == nil
		if newRowID > 0 {
			newRow, rErr := h.JPRepo.GetByID(ctx, newRowID)
			if rErr == nil && newRow != nil {
				if ok, _, msg, price, noreff := providerRowSuccessState(provider, newRow); ok {
					h.logf("RECHECK sukses provider=%s refid=%s price=%d msg=%s", provider, refID, price, msg)
					ketDB := strings.TrimSpace(noreff)
					if ketDB == "" {
						ket, info := helper.SafeMemberKeterangan("success", msg)
						ketDB = strings.TrimSpace(info.SN)
						if ketDB == "" {
							ketDB = ket
						}
					}
					h.settleLikeCallback(ctx, trx, "success", ketDB, noreff, price)
					_, _, _ = h.sendFinalWebhook(ctx, trx, "success", "Transaksi berhasil", noreff, noreff, price)
					return provider, newRow.ID, nil
				}
			}
		}

		if callErr == nil {
			h.logf("RECHECK pending provider=%s refid=%s accepted, row_id=%d", provider, refID, newRowID)
			ketPending, _ := helper.SafeMemberKeterangan("pending", "recheck same refid accepted")
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
			return usedP, newRowID, nil
		}
		h.logf("RECHECK pending provider=%s refid=%s result=failed err=%v", provider, refID, callErr)
		break // cukup recheck 1 provider pending terlama
	}

	usedProvider, providerRowID, ferr := h.tryStatusPayFallback(ctx, trx, triedProviders)
	if ferr != nil {
		errMsg := strings.ToLower(ferr.Error())
		// Hanya settle failed kalau benar-benar tidak ada provider tersisa
		// JANGAN settle kalau masih ada provider pending/sukses (guard block)
		noProviderLeft := strings.Contains(errMsg, "tidak ada provider fallback tersisa") ||
			strings.Contains(errMsg, "tidak ada provider eligible")
		if noProviderLeft && strings.EqualFold(strings.TrimSpace(trx.Status), "pending") {
			if ok, reason := h.allExistingRouteAttemptsFinalFailed(ctx, trx.ID, trx.RefID); ok {
				ketFailed, _ := helper.SafeMemberKeterangan("failed", "semua provider gagal")
				h.settleLikeCallback(ctx, trx, "failed", ketFailed, "", 0)
				_, _, _ = h.sendFinalWebhook(ctx, trx, "failed", ketFailed, "", ketFailed, 0)
				h.logf("RECOVER settle failed (semua mapping gagal) refid=%s reason=%s err=%v", refID, reason, ferr)
			} else {
				ketPending, _ := helper.SafeMemberKeterangan("pending", reason)
				_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
				h.logf("RECOVER tetap pending refid=%s reason=%s err=%v", refID, reason, ferr)
			}
		}
		return usedProvider, providerRowID, ferr
	}

	ketPending, _ := helper.SafeMemberKeterangan("pending", "manual recovery fallback started")
	_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trxID, "pending", ketPending, 0)
	h.logf("RECOVER pending refid=%s member_id=%d fallback_provider=%s provider_row_id=%d", refID, memberID, usedProvider, providerRowID)
	return usedProvider, providerRowID, nil
}

func (h *MemberTrxService) ResendPendingProviderRow(ctx context.Context, providerTrxID int64) (string, int64, error) {
	if h == nil || h.MemberRepo == nil || h.JPRepo == nil {
		return "", 0, fmt.Errorf("service belum siap")
	}
	if providerTrxID <= 0 {
		return "", 0, fmt.Errorf("provider_trx_id required")
	}

	row, err := h.JPRepo.GetByID(ctx, providerTrxID)
	if err != nil {
		return "", 0, err
	}
	if row == nil {
		return "", 0, fmt.Errorf("transaksi provider tidak ditemukan")
	}
	if !strings.EqualFold(strings.TrimSpace(row.Perintah), "PAY") {
		return "", 0, fmt.Errorf("kirim ulang hanya untuk perintah PAY")
	}

	if !strings.EqualFold(strings.TrimSpace(row.Status), "pending") {
		return "", 0, fmt.Errorf("transaksi provider bukan pending")
	}

	trx, err := h.MemberRepo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil {
		return "", 0, err
	}
	if trx == nil {
		return "", 0, fmt.Errorf("transaksi member tidak ditemukan")
	}
	memberStatus := strings.ToLower(strings.TrimSpace(trx.Status))
	if memberStatus == "success" {
		return "", 0, fmt.Errorf("transaksi member sudah success, resend dibatalkan")
	}
	if memberStatus == "failed" {
		// Reopen member ke pending — admin sengaja kirim ulang
		ketPending, _ := helper.SafeMemberKeterangan("pending", "admin resend: reopen dari failed")
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		h.logf("ADMIN RESEND reopen failed→pending refid=%s trx_id=%d", trx.RefID, trx.ID)
		// Refresh trx
		trx, err = h.MemberRepo.GetTransaksiMemberByID(ctx, trx.ID)
		if err != nil {
			return "", 0, err
		}
	}

	// Cek apakah ada provider yang sudah sukses/masih proses untuk refid ini
	allAttempts, listErr := h.JPRepo.ListByRefID(ctx, trx.RefID)
	if listErr == nil {
		for _, att := range allAttempts {
			if att == nil || att.ID == providerTrxID {
				continue // skip row yang sedang di-resend
			}
			attRow := &model.JavapayTrxRow{
				ID: att.ID, KodeRespon: att.KodeRespon, Pesan: att.Pesan,
				NoReferensi: att.NoReferensi, Harga: att.Harga,
			}
			provider := strings.ToLower(strings.TrimSpace(att.Provider))
			if ok, _, _, _, _ := providerRowSuccessState(provider, attRow); ok {
				return "", 0, fmt.Errorf("provider %s sudah sukses untuk refid ini, resend dibatalkan", provider)
			}
			if ok, _, _ := providerRowPendingState(provider, attRow); ok {
				return "", 0, fmt.Errorf("provider %s masih pending untuk refid ini, resend dibatalkan", provider)
			}
		}
	}

	in := trxmemberdto.TrxRequest{
		Commands: strings.TrimSpace(trx.Perintah),
		Product:  strings.TrimSpace(trx.KodeProduk),
		Dest:     resolveProviderAttemptDest(strings.TrimSpace(row.Provider), trx.Tujuan, row.Tujuan),
		Qty:      trx.QtyProvider,
		RefID:    strings.TrimSpace(trx.RefID),
	}
	if in.Qty <= 0 {
		in.Qty = trx.Qty
	}

	specialCode, mode := restoreProviderAttemptMeta(row.RequestMentah)
	attempt := providerRouteAttempt{
		Name:                strings.ToLower(strings.TrimSpace(row.Provider)),
		Src:                 "manual_resend_same_provider_row",
		ProdukSKUSnapshot:   strings.TrimSpace(row.ProdukSKUSnapshot),
		ProdukProviderMapID: row.ProdukProviderMapID,
		KodeProduk:          normalizeResendProviderKodeProduk(strings.TrimSpace(row.Provider), row.KodeProduk, specialCode),
		SpecialCode:         specialCode,
		Mode:                mode,
	}
	if attempt.Name == "" {
		return "", 0, fmt.Errorf("provider row tidak valid")
	}
	if attempt.KodeProduk == "" {
		return "", 0, fmt.Errorf("kode produk provider tidak valid")
	}

	attemptCtx, cancel := context.WithTimeout(context.Background(), providerCallWindowForName(attempt.Name)+5*time.Second)
	defer cancel()

	usedProvider, newProviderRowID, callErr := h.callProviderOnly(attemptCtx, trx.ID, in, attempt, attempt.Src)
	if callErr != nil {
		return usedProvider, newProviderRowID, callErr
	}

	ketPending, _ := helper.SafeMemberKeterangan("pending", "manual resend same provider started")
	_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
	h.logf("MANUAL resend pending provider_row_id=%d refid=%s provider=%s new_provider_row_id=%d", providerTrxID, trx.RefID, usedProvider, newProviderRowID)
	return usedProvider, newProviderRowID, nil
}

func providerRowRefundLike(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	parts := []string{row.Status}
	if row.Pesan != nil {
		parts = append(parts, *row.Pesan)
	}
	if row.NoReferensi != nil {
		parts = append(parts, *row.NoReferensi)
	}
	msg := strings.ToLower(strings.Join(parts, " "))
	return strings.Contains(msg, "refund") || strings.Contains(msg, "saldo dikembalikan") || strings.Contains(msg, "dikembalikan")
}

func (h *MemberTrxService) ResendRefundProviderRowNoSuccess(ctx context.Context, providerTrxID int64) (string, int64, error) {
	if h == nil || h.MemberRepo == nil || h.JPRepo == nil {
		return "", 0, fmt.Errorf("service belum siap")
	}
	if providerTrxID <= 0 {
		return "", 0, fmt.Errorf("provider_trx_id required")
	}

	row, err := h.JPRepo.GetByID(ctx, providerTrxID)
	if err != nil {
		return "", 0, err
	}
	if row == nil {
		return "", 0, fmt.Errorf("transaksi provider tidak ditemukan")
	}
	if !strings.EqualFold(strings.TrimSpace(row.Perintah), "PAY") {
		return "", 0, fmt.Errorf("proses ulang hanya untuk perintah PAY")
	}
	if strings.EqualFold(strings.TrimSpace(row.Status), "pending") {
		return "", 0, fmt.Errorf("transaksi provider masih pending")
	}
	if !providerRowRefundLike(row) {
		return "", 0, fmt.Errorf("transaksi provider bukan refund/saldo dikembalikan")
	}

	trx, err := h.MemberRepo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil {
		return "", 0, err
	}
	if trx == nil {
		return "", 0, fmt.Errorf("transaksi member tidak ditemukan")
	}
	memberStatus := strings.ToLower(strings.TrimSpace(trx.Status))
	if memberStatus != "success" {
		return "", 0, fmt.Errorf("proses ulang hanya untuk transaksi member success")
	}

	allAttempts, listErr := h.JPRepo.ListByRefID(ctx, trx.RefID)
	if listErr != nil {
		return "", 0, listErr
	}
	for _, att := range allAttempts {
		if att == nil || att.ID == providerTrxID {
			continue
		}
		if att.TransaksiMemberID != trx.ID {
			continue
		}
		providerName := strings.ToLower(strings.TrimSpace(att.Provider))
		if ok, _, _, _, _ := providerRowSuccessState(providerName, att); ok {
			return "", 0, fmt.Errorf("provider %s sudah sukses untuk refid ini, proses ulang dibatalkan", providerName)
		}
		if ok, _, _ := providerRowPendingState(providerName, att); ok {
			return "", 0, fmt.Errorf("provider %s masih pending untuk refid ini, proses ulang dibatalkan", providerName)
		}
	}

	in := trxmemberdto.TrxRequest{
		Commands: strings.TrimSpace(trx.Perintah),
		Product:  strings.TrimSpace(trx.KodeProduk),
		Dest:     resolveProviderAttemptDest(strings.TrimSpace(row.Provider), trx.Tujuan, row.Tujuan),
		Qty:      trx.QtyProvider,
		RefID:    strings.TrimSpace(trx.RefID),
	}
	if in.Qty <= 0 {
		in.Qty = trx.Qty
	}

	specialCode, mode := restoreProviderAttemptMeta(row.RequestMentah)
	attempt := providerRouteAttempt{
		Name:                strings.ToLower(strings.TrimSpace(row.Provider)),
		Src:                 "manual_resend_refund_no_success",
		ProdukSKUSnapshot:   strings.TrimSpace(row.ProdukSKUSnapshot),
		ProdukProviderMapID: row.ProdukProviderMapID,
		KodeProduk:          normalizeResendProviderKodeProduk(strings.TrimSpace(row.Provider), row.KodeProduk, specialCode),
		SpecialCode:         specialCode,
		Mode:                mode,
	}
	if attempt.Name == "" {
		return "", 0, fmt.Errorf("provider row tidak valid")
	}
	if attempt.KodeProduk == "" {
		return "", 0, fmt.Errorf("kode produk provider tidak valid")
	}

	attemptCtx, cancel := context.WithTimeout(context.Background(), providerCallWindowForName(attempt.Name)+5*time.Second)
	defer cancel()

	usedProvider, newProviderRowID, callErr := h.callProviderOnly(attemptCtx, trx.ID, in, attempt, attempt.Src)
	if callErr != nil {
		return usedProvider, newProviderRowID, callErr
	}

	h.logf("MANUAL resend refund provider_row_id=%d refid=%s provider=%s new_provider_row_id=%d member_status_kept=success", providerTrxID, trx.RefID, usedProvider, newProviderRowID)
	return usedProvider, newProviderRowID, nil
}
