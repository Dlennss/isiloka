package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

func (s *HistoryService) AdminCancelPendingTransaksi(ctx context.Context, adminID, trxID int64, reason string, allowSuccessCancel bool) (*repository.AdminCancelTrxResult, error) {
	if adminID <= 0 {
		return nil, errors.New("admin only")
	}
	if trxID <= 0 {
		return nil, errors.New("trx_id required")
	}
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("reason required")
	}
	item, err := s.repo.AdminCancelPendingTransaksi(ctx, adminID, trxID, reason, allowSuccessCancel)
	if err != nil {
		return nil, err
	}

	callbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
	defer cancel()
	callback, callbackErr := s.AdminSendTransaksiCallback(callbackCtx, adminID, item.TrxID)
	if callbackErr != nil {
		item.CallbackError = callbackErr.Error()
	} else {
		item.Callback = callback
	}

	return item, nil
}

func (s *HistoryService) AdminSendTransaksiCallback(ctx context.Context, adminID, trxID int64) (*repository.AdminSendCallbackResult, error) {
	if adminID <= 0 {
		return nil, errors.New("admin only")
	}
	if trxID <= 0 {
		return nil, errors.New("trx_id required")
	}

	trx, err := s.repo.GetAdminTrxCallbackTarget(ctx, trxID)
	if err != nil {
		return nil, err
	}
	webhookURL, err := s.repo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(webhookURL) == "" {
		return nil, errors.New("member webhook tidak diatur")
	}
	memberBalance, err := s.repo.GetSaldo(ctx, trx.MemberID)
	if err != nil {
		return nil, err
	}
	providerRow, err := s.repo.GetLatestProviderByRefID(ctx, trx.RefID)
	if err != nil {
		return nil, err
	}

	finalStatus := strings.ToLower(strings.TrimSpace(trx.Status))
	if finalStatus == "" {
		finalStatus = "pending"
	}
	ket, providerRef, sn, _, providerName := buildAdminSendCallbackInfo(finalStatus, trx, providerRow)
	biayaAktualOut := int64(0)
	if finalStatus == "success" {
		biayaAktualOut = trx.BiayaAktual
		if biayaAktualOut <= 0 {
			biayaAktualOut = trx.BiayaPerkiraan
		}
	}

	payload := buildAdminSendCallbackPayload(trx, finalStatus, ket, providerRef, sn, memberBalance, biayaAktualOut)
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	httpStatus, body, postErr := helper.PostJSON(cctx, webhookURL, payload)
	if postErr != nil {
		return nil, fmt.Errorf("gagal kirim callback: %w", postErr)
	}

	return &repository.AdminSendCallbackResult{
		TrxID:         trx.ID,
		RefID:         trx.RefID,
		Status:        finalStatus,
		Provider:      providerName,
		ProviderRef:   providerRef,
		SN:            sn,
		CallbackHTTP:  httpStatus,
		CallbackBody:  body,
		WebhookURL:    webhookURL,
		ProcessedByID: adminID,
	}, nil
}

func (s *HistoryService) AdminSendTransaksiCallbackBulk(ctx context.Context, adminID int64, trxIDs []int64) (resolved []*repository.AdminSendCallbackResult, failed []map[string]any, err error) {
	if adminID <= 0 {
		return nil, nil, errors.New("admin only")
	}
	if len(trxIDs) == 0 {
		return nil, nil, errors.New("trx_ids required")
	}
	for _, trxID := range trxIDs {
		item, callErr := s.AdminSendTransaksiCallback(ctx, adminID, trxID)
		if callErr != nil {
			failed = append(failed, map[string]any{
				"trx_id": trxID,
				"error":  callErr.Error(),
			})
			continue
		}
		resolved = append(resolved, item)
	}
	return resolved, failed, nil
}
func buildAdminSendCallbackPayload(trx *repository.AdminTrxCallbackTarget, finalStatus, ket, providerRef, sn string, memberBalance, biayaAktualOut int64) map[string]any {
	qtyProvider := trx.QtyProvider
	if qtyProvider <= 0 {
		qtyProvider = trx.Qty
	}
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	snField := strings.TrimSpace(sn)
	if strings.EqualFold(strings.TrimSpace(finalStatus), "failed") && snField == "" {
		snField = strings.TrimSpace(ket)
	}

	return map[string]any{
		"refid":          trx.RefID,
		"status":         finalStatus,
		"member_balance": memberBalance,
		"trx": map[string]any{
			"id":           trx.ID,
			"commands":     trx.Perintah,
			"product":      trx.KodeProduk,
			"dest":         trx.Tujuan,
			"qty":          trx.Qty,
			"qty_provider": qtyProvider,
			"harga_member": hargaMember,
			"biaya_aktual": biayaAktualOut,
			"message":      ket,
			"provider_ref": providerRef,
			"sn":           snField,
		},
	}
}

func buildAdminSendCallbackInfo(finalStatus string, trx *repository.AdminTrxCallbackTarget, row *model.JavapayTrxRow) (ket string, providerRef string, sn string, price int64, provider string) {
	if row == nil {
		ket, _ = helper.SafeMemberKeterangan(finalStatus, strings.TrimSpace(trx.Keterangan))
		if ket == "" {
			ket = strings.TrimSpace(trx.Keterangan)
		}
		return ket, "", "", 0, ""
	}

	provider = strings.ToLower(strings.TrimSpace(row.Provider))
	msg := ""
	noreff := ""
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	if row.NoReferensi != nil {
		noreff = strings.TrimSpace(*row.NoReferensi)
	}
	if row.Harga != nil {
		price = *row.Harga
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	providerRef = strings.TrimSpace(info.Reff)
	sn = strings.TrimSpace(info.SN)

	switch provider {
	case "talentapay":
		parsedRef, parsedSN := providersn.ParseTalentaSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "multikom":
		parsedRef, parsedSN := providersn.ParseMultikomSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "yuscom":
		parsedRef, parsedSN := providersn.ParseYuscomSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "sagaramobile":
		parsedRef, parsedSN := providersn.ParseSagaraSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "minions":
		parsedRef, parsedSN := providersn.ParseMinionsSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "trionik":
		parsedRef, parsedSN := providersn.ParseYuscomSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "ajs":
		parsedRef, parsedSN := providersn.ParseAJSSNRefFromMsg(msg)
		providerRef, sn = mergeHistoryProviderInfo(providerRef, sn, parsedRef, parsedSN)
	}

	if sn == "" {
		sn = noreff
	}
	if providerRef == "" {
		providerRef = noreff
	}
	if ket == "" {
		ket = strings.TrimSpace(trx.Keterangan)
	}
	return ket, providerRef, sn, price, provider
}

func mergeHistoryProviderInfo(providerRef, sn string, parsedRef, parsedSN string) (string, string) {
	if providerRef == "" || isWeakHistoryProviderRef(providerRef) {
		providerRef = strings.TrimSpace(parsedRef)
	}
	if sn == "" || isWeakHistoryProviderRef(sn) {
		sn = strings.TrimSpace(parsedSN)
	}
	return providerRef, sn
}

func isWeakHistoryProviderRef(v string) bool {
	v = strings.TrimSpace(strings.ToUpper(v))
	switch v {
	case "", "NO", "N/A", "-":
		return true
	default:
		return false
	}
}
