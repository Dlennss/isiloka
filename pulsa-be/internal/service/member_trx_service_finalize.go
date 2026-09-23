package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (h *MemberTrxService) sendFinalWebhook(ctx context.Context, trx *repository.TrxMemberFull, finalStatus string, ket string, providerReff string, snOut string, price int64) (st int, body string, err error) {
	webhookURL, err := h.MemberRepo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil || strings.TrimSpace(webhookURL) == "" {
		return 0, "", nil
	}

	memberSaldo := int64(0)
	if bal, err2 := h.MemberRepo.GetSaldo(ctx, trx.MemberID); err2 == nil {
		memberSaldo = bal
	}

	biayaAktualOut := int64(0)
	if strings.EqualFold(strings.TrimSpace(finalStatus), "success") {
		// Model hold sederhana: member dipotong sekali saat request PAY (qty + fee).
		// Saat final success tidak ada adjust tambahan.
		biayaAktualOut = trx.BiayaPerkiraan
	}

	out := buildFinalWebhookPayload(trx, finalStatus, ket, providerReff, snOut, memberSaldo, biayaAktualOut)
	trxPayload, _ := out["trx"].(map[string]any)
	payloadSN, _ := trxPayload["sn"].(string)
	payloadProviderRef, _ := trxPayload["provider_ref"].(string)

	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	snTrimmed := strings.TrimSpace(payloadSN)
	providerRefTrimmed := strings.TrimSpace(payloadProviderRef)
	h.logf("WEBHOOK kirim refid=%s trx_id=%d status=%s url=%s qty=%d qty_provider=%d biaya_perkiraan=%d biaya_aktual=%d provider_ref=%s provider_ref_present=%t sn_present=%t sn_len=%d harga=%d fee=%d",
		trx.RefID, trx.ID, finalStatus, webhookURL, trx.Qty, trx.QtyProvider, trx.BiayaPerkiraan, biayaAktualOut, providerRefTrimmed, providerRefTrimmed != "", snTrimmed != "", len(snTrimmed), price, trx.FeeMemberRp)

	st, body, err = helper.PostJSON(cctx, webhookURL, out)

	if err != nil {
		h.logf("WEBHOOK gagal refid=%s trx_id=%d err=%v", trx.RefID, trx.ID, err)
	} else {
		h.logf("WEBHOOK selesai refid=%s trx_id=%d http=%d panjang_body=%d", trx.RefID, trx.ID, st, len(body))
	}

	return st, body, err
}

func buildFinalWebhookPayload(trx *repository.TrxMemberFull, finalStatus string, ket string, providerReff string, snOut string, memberSaldo int64, biayaAktualOut int64) map[string]any {
	qtyProvider := trx.QtyProvider
	if qtyProvider <= 0 {
		qtyProvider = trx.Qty
	}
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	storedKet := strings.TrimSpace(trx.Keterangan)
	messageField, providerRefField, snField := normalizeMemberWebhookFields(finalStatus, ket, providerReff, snOut, storedKet)

	return map[string]any{
		"refid":          trx.RefID,
		"status":         finalStatus,
		"member_balance": memberSaldo,
		"trx": map[string]any{
			"id":           trx.ID,
			"commands":     trx.Perintah,
			"product":      trx.KodeProduk,
			"dest":         trx.Tujuan,
			"qty":          trx.Qty,
			"qty_provider": qtyProvider,
			"harga_member": hargaMember,
			"biaya_aktual": biayaAktualOut,
			"message":      messageField,
			"provider_ref": providerRefField,
			"sn":           snField,
		},
	}
}

func mapCallbackDelivery(statusCode int, err error) *trxmemberdto.CallbackDelivery {
	if statusCode == 0 && err == nil {
		return nil
	}
	out := &trxmemberdto.CallbackDelivery{
		Attempted: true,
		Delivered: err == nil,
	}
	if statusCode > 0 {
		out.HTTPStatus = statusCode
	}
	if err != nil {
		out.Error = err.Error()
	}
	return out
}

func (h *MemberTrxService) settleLikeCallback(ctx context.Context, trx *repository.TrxMemberFull, finalStatus string, ket string, infoReff string, price int64) {
	finalStatus = strings.TrimSpace(strings.ToLower(finalStatus))
	cur := strings.TrimSpace(strings.ToLower(trx.Status))

	if cur == "success" || cur == "failed" {
		h.logf("SETTLE lewati karena sudah final refid=%s trx_id=%d saat_ini=%s target=%s", trx.RefID, trx.ID, cur, finalStatus)
		return
	}

	h.logf("SETTLE mulai refid=%s trx_id=%d dari=%s ke=%s harga=%d fee=%d biaya_perkiraan=%d info_reff=%s",
		trx.RefID, trx.ID, cur, finalStatus, price, trx.FeeMemberRp, trx.BiayaPerkiraan, infoReff)

	if finalStatus == "success" {
		biayaAktual := trx.BiayaPerkiraan
		hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
		if err := h.MemberRepo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ket, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			h.logf("SETTLE gagal update success refid=%s trx_id=%d err=%v", trx.RefID, trx.ID, err)
		}
		h.applyH2HCommission(ctx, trx.ID, trx.RefID)
		h.cleanupOtherPendingProviderRows(ctx, trx.RefID, finalStatus)
		h.logf("SETTLE sukses, hold dipertahankan refid=%s trx_id=%d biaya_aktual=%d", trx.RefID, trx.ID, biayaAktual)
		return
	}

	if finalStatus == "failed" {
		hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
		// Update status ke failed → DB trigger otomatis refund saldo
		if err := h.MemberRepo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ket, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			h.logf("SETTLE gagal update failed refid=%s trx_id=%d err=%v", trx.RefID, trx.ID, err)
		}
		h.cleanupOtherPendingProviderRows(ctx, trx.RefID, finalStatus)
		h.logf("SETTLE gagal, refund refid=%s trx_id=%d nominal=%d (db trigger)", trx.RefID, trx.ID, trx.BiayaPerkiraan)
		return
	}

	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	if err := h.MemberRepo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ket, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
		h.logf("SETTLE gagal update pending refid=%s trx_id=%d err=%v", trx.RefID, trx.ID, err)
	}
	h.logf("SETTLE update pending refid=%s trx_id=%d", trx.RefID, trx.ID)
}

// cleanupOtherPendingProviderRows — seharusnya TIDAK PERNAH menemukan rows pending.
// Kalau ada, berarti ada bug dispatch duplikat yang harus diperbaiki.
func (h *MemberTrxService) cleanupOtherPendingProviderRows(ctx context.Context, refID string, memberFinalStatus string) {
	if h.JPRepo == nil || strings.TrimSpace(refID) == "" {
		return
	}
	// Cek dulu berapa yang pending — kalau ada, log sebagai BUG ALERT
	allAttempts, lErr := h.JPRepo.ListByRefID(ctx, refID)
	if lErr != nil {
		return
	}
	var pendingProviders []string
	for _, att := range allAttempts {
		if att == nil {
			continue
		}
		provider := strings.ToLower(strings.TrimSpace(att.Provider))
		if pending, _, _ := providerRowPendingState(provider, att); pending {
			pendingProviders = append(pendingProviders, fmt.Sprintf("%s(id=%d)", provider, att.ID))
		}
	}
	if len(pendingProviders) > 0 {
		h.logf("BUG ALERT: settle %s tapi masih ada %d provider pending refid=%s providers=%v — SEHARUSNYA TIDAK TERJADI, cek code path dispatch",
			memberFinalStatus, len(pendingProviders), refID, pendingProviders)
		helper.AppendProviderServiceLog("provider_callback_error.log",
			"BUG ALERT double dispatch detected: settle=%s refid=%s pending_providers=%v",
			memberFinalStatus, refID, pendingProviders)
		// Tetap cleanup supaya tidak nyangkut, tapi ini indikasi bug
		reason := fmt.Sprintf("BUG CLEANUP: transaksi final %s, provider ini seharusnya tidak pernah di-dispatch", memberFinalStatus)
		_, _ = h.JPRepo.MarkPendingRowsAsFailed(ctx, refID, reason)
	}
}

func (h *MemberTrxService) insertProviderSnapshot(ctx context.Context, provider string, trxID int64, providerTrxID int64, refID string, saldo int64, source string, raw any) {
	if h.ProviderWallet == nil || saldo <= 0 {
		return
	}
	var rawJSON []byte
	if raw != nil {
		rawJSON, _ = json.Marshal(raw)
	}
	tmID := trxID
	tpID := providerTrxID
	if err := h.ProviderWallet.InsertSnapshot(ctx, repository.ProviderSnapshotTxIn{
		Provider:            provider,
		SaldoProvider:       saldo,
		RefID:               refID,
		TransaksiMemberID:   &tmID,
		TransaksiProviderID: &tpID,
		Sumber:              source,
		RawJSON:             rawJSON,
	}); err != nil {
		h.logf("gagal simpan snapshot saldo provider refid=%s provider=%s saldo=%d source=%s err=%v", refID, provider, saldo, source, err)
	}
}

func (h *MemberTrxService) providerAvailableBalance(ctx context.Context, provider string) int64 {
	if h.ProviderWallet == nil {
		return 0
	}
	internalSaldo, err := h.ProviderWallet.GetSaldo(ctx, provider)
	if err != nil {
		return 0
	}
	return internalSaldo
}

func (h *MemberTrxService) applyH2HCommission(ctx context.Context, trxID int64, refID string) {
	if h.H2HRepo == nil || trxID <= 0 {
		return
	}
	if err := h.H2HRepo.ApplyCommissionForSuccessTrx(ctx, trxID); err != nil {
		h.logf("H2H komisi gagal refid=%s trx_id=%d err=%v", refID, trxID, err)
	}
}
