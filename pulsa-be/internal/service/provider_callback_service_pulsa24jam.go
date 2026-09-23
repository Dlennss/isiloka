package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type pulsa24JamCallbackData struct {
	refid       string
	status      string
	rc          string
	msg         string
	sn          string
	providerRef string
	price       int64
	balance     int64
	payload     map[string]any
}

func (s *ProviderCallbackService) ProcessPulsa24JamCallback(ctx context.Context, raw string, q url.Values, payload map[string]any) (int, map[string]any) {
	data := parsePulsa24JamCallback(raw, q, payload)
	if data.refid == "" {
		helper.AppendProviderServiceLog("provider_anomali.log", "pulsa24jam callback missing refid raw=%s", raw)
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, data.refid, "pulsa24jam")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		appRow, appErr := s.appProviderRepo.GetByRefID(ctx, data.refid, "pulsa24jam")
		if appErr == nil && appRow != nil {
			return s.processPulsa24JamAppCallback(ctx, data, appRow)
		}
		if appErr != nil && appErr != sql.ErrNoRows {
			return 502, map[string]any{"ok": false, "error": appErr.Error()}
		}
		billingRow, billingErr := s.billingCheckRepo.GetByRefID(ctx, data.refid)
		if billingErr == nil && billingRow != nil && strings.EqualFold(strings.TrimSpace(billingRow.Provider), "pulsa24jam") {
			return s.processPulsa24JamBillingCheckCallback(ctx, data, billingRow, raw)
		}
		if billingErr != nil && billingErr != sql.ErrNoRows {
			return 502, map[string]any{"ok": false, "error": billingErr.Error()}
		}
		if s.retailRepo != nil {
			withdrawRow, withdrawErr := s.retailRepo.GetWithdrawRequestByRefID(ctx, data.refid)
			if withdrawErr == nil && withdrawRow != nil && strings.EqualFold(strings.TrimSpace(withdrawRow.SourceType), "credit") {
				return s.processPulsa24JamRetailWithdrawCallback(ctx, data, withdrawRow)
			}
			if withdrawErr != nil && withdrawErr != sql.ErrNoRows {
				return 502, map[string]any{"ok": false, "error": withdrawErr.Error()}
			}
		}
		helper.AppendProviderServiceLog("provider_anomali.log", "pulsa24jam callback unmatched refid=%s raw=%s", data.refid, raw)
		return 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid}
	}

	httpStatus := 200
	noref := firstText(data.providerRef, data.sn)
	var norefPtr *string
	if noref != "" {
		norefPtr = &noref
	}
	var balancePtr *int64
	if data.balance > 0 {
		balancePtr = &data.balance
	}
	upd := repository.UpdateResult{
		HTTPStatus:    &httpStatus,
		KodeRespon:    &data.rc,
		Pesan:         &data.msg,
		Harga:         &data.price,
		NoReferensi:   norefPtr,
		SaldoTerakhir: balancePtr,
		ResponMentah:  data.payload,
	}
	_ = s.repo.UpdateResult(ctx, row.ID, upd)
	if data.balance > 0 {
		rawJSON, _ := json.Marshal(data.payload)
		tmID := row.TransaksiMemberID
		tpID := row.ID
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:            "pulsa24jam",
			SaldoProvider:       data.balance,
			RefID:               data.refid,
			TransaksiMemberID:   &tmID,
			TransaksiProviderID: &tpID,
			Sumber:              "callback_pulsa24jam",
			RawJSON:             rawJSON,
		})
	}

	trx, err := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": data.refid}
	}
	finalStatus := pulsa24JamFinalStatus(data)
	if locked, _ := s.repo.AcquireCallbackLock(ctx, trx.ID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, trx.ID)
		if fresh, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID); fresh != nil {
			trx = fresh
		}
	}
	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=pulsa24jam refid=%s trx_member_id=%d status=%s callback_status=%s", data.refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid}
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, data.msg)
	ketDB := firstText(data.sn, data.providerRef, strings.TrimSpace(info.SN), strings.TrimSpace(info.Reff), ket)
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)

	switch finalStatus {
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "pulsa24jam", trx); err != nil {
			return 200, map[string]any{"ok": false, "refid": data.refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=pulsa24jam refid=%s trx_member_id=%d status=success err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, data.refid)
	case "failed":
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketDB, 0, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=pulsa24jam refid=%s trx_member_id=%d status=failed err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
	default:
		_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketDB, 0, data.price, hargaMember)
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending"}
	}

	webhookURL, err := s.repo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil || strings.TrimSpace(webhookURL) == "" {
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": finalStatus}
	}
	memberSaldo := int64(0)
	if bal, err := s.repo.GetSaldo(ctx, trx.MemberID); err == nil {
		memberSaldo = bal
	}
	biayaAktualOut := int64(0)
	if finalStatus == "success" {
		biayaAktualOut = trx.BiayaPerkiraan
	}
	return s.sendDirectMemberWebhook(ctx, "pulsa24jam", trx, webhookURL, data.refid, finalStatus, ket, firstText(data.providerRef, data.sn), firstText(data.sn, data.providerRef), memberSaldo, biayaAktualOut, data.price)
}

func (s *ProviderCallbackService) processPulsa24JamBillingCheckCallback(ctx context.Context, data pulsa24JamCallbackData, row *repository.AppBillingCheckRow, raw string) (int, map[string]any) {
	status := "processing_provider"
	switch pulsa24JamFinalStatus(data) {
	case "success":
		status = "success"
	case "failed":
		status = "failed"
	}
	rc := strings.TrimSpace(data.rc)
	msg := strings.TrimSpace(data.msg)
	price := data.price
	if err := s.billingCheckRepo.UpdateResult(ctx, repository.AppBillingCheckUpdateInput{
		ID:            row.ID,
		HargaProvider: &price,
		Status:        status,
		KodeRespon:    &rc,
		Pesan:         &msg,
		RawCallback:   raw,
	}); err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	return 200, map[string]any{"ok": true, "refid": data.refid, "status": status}
}

func (s *ProviderCallbackService) processPulsa24JamRetailWithdrawCallback(ctx context.Context, data pulsa24JamCallbackData, row *repository.RetailWithdrawRequestRow) (int, map[string]any) {
	if row == nil {
		return 200, map[string]any{"ok": true, "refid": data.refid, "ignored": true}
	}
	finalStatus := pulsa24JamFinalStatus(data)
	switch finalStatus {
	case "success":
		note := strings.TrimSpace(firstText(data.sn, data.providerRef, data.msg))
		if note != "" {
			note = "Pulsa24Jam berhasil: " + note
		} else {
			note = "Pulsa24Jam berhasil"
		}
		if err := s.retailRepo.UpdateWithdrawRequestProviderStatus(ctx, data.refid, "approved", note); err != nil {
			return 502, map[string]any{"ok": false, "error": err.Error(), "refid": data.refid}
		}
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "approved"}
	case "failed":
		reason := strings.TrimSpace(firstText(data.msg, data.status, "penarikan Pulsa24Jam gagal"))
		if err := s.retailRepo.RejectWithdrawRequest(ctx, row.ID, row.MemberID, reason); err != nil {
			return 502, map[string]any{"ok": false, "error": err.Error(), "refid": data.refid}
		}
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "rejected"}
	default:
		note := strings.TrimSpace(firstText(data.msg, data.status, "menunggu callback final Pulsa24Jam"))
		if err := s.retailRepo.UpdateWithdrawRequestProviderStatus(ctx, data.refid, "processing_provider", note); err != nil {
			return 502, map[string]any{"ok": false, "error": err.Error(), "refid": data.refid}
		}
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "processing_provider"}
	}
}

func (s *ProviderCallbackService) processPulsa24JamAppCallback(ctx context.Context, data pulsa24JamCallbackData, row *repository.AppOrderProviderTrxRow) (int, map[string]any) {
	order, err := s.appOrderRepo.GetByID(ctx, row.AppOrderID)
	if err != nil || order == nil {
		return 200, map[string]any{"ok": true, "refid": data.refid, "ignored": true}
	}
	finalStatus := pulsa24JamFinalStatus(data)
	rawJSON, _ := json.Marshal(data.payload)
	updateIn := repository.AppOrderProviderTrxUpdateInput{
		ID:          row.ID,
		Status:      finalStatus,
		KodeRespon:  data.rc,
		Pesan:       data.msg,
		SN:          firstText(data.sn, data.providerRef),
		RawCallback: string(rawJSON),
	}
	if data.price > 0 {
		updateIn.HargaProvider = &data.price
	}
	if err := s.appProviderRepo.UpdateResult(ctx, updateIn); err != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log", "pulsa24jam app callback update failed refid=%s app_provider_id=%d err=%v", data.refid, row.ID, err)
	}

	if order.Status == "success" || order.Status == "failed" || order.Status == "refunded" {
		helper.AppendAlreadyFinalLog("callback already_final provider=pulsa24jam refid=%s app_order_id=%d status=%s callback_status=%s", data.refid, order.ID, order.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid}
	}
	if finalStatus == "pending" {
		_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "processing_provider")
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending"}
	}
	if finalStatus == "success" {
		if data.price > 0 {
			appProviderID := row.ID
			if _, _, err := s.repo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
				Provider:              "pulsa24jam",
				RefID:                 data.refid,
				Arah:                  "debit",
				Jumlah:                data.price,
				Alasan:                "APP_TRX_SUCCESS_COST",
				Catatan:               "auto debit by callback (app success)",
				AppOrderProviderTrxID: &appProviderID,
			}); err != nil {
				helper.AppendProviderServiceLog("provider_wallet.log", "provider wallet debit app failed provider=pulsa24jam refid=%s app_provider_id=%d err=%v", data.refid, row.ID, err)
			}
		}
		_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "success")
		if s.retailRepo != nil {
			if err := s.retailRepo.ApplyCommissionForOrder(ctx, order.ID); err != nil {
				helper.AppendProviderServiceLog("provider_wallet.log", "retail commission apply failed invoice=%s order_id=%d err=%v", data.refid, order.ID, err)
			}
		}
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "success"}
	}
	if appOrderProviderProductUnavailable("pulsa24jam", data.msg) && s.appPricingRepo != nil {
		if markErr := s.appPricingRepo.MarkProviderProductUnavailable(ctx, order.ProdukID, "pulsa24jam"); markErr != nil {
			helper.AppendProviderServiceLog("provider_callback_service.log", "mark product unavailable from callback failed provider=pulsa24jam product_id=%d sku=%s err=%v", order.ProdukID, order.ProdukSKUSnapshot, markErr)
		} else {
			helper.AppendProviderServiceLog("provider_callback_service.log", "product unavailable from callback provider=pulsa24jam product_id=%d sku=%s until=verified", order.ProdukID, order.ProdukSKUSnapshot)
		}
	}

	if order.BuyerType == "user" && order.MemberID != nil && *order.MemberID > 0 && order.HargaFinal > 0 {
		reason := "refund saldo otomatis karena transaksi Pulsa24Jam gagal"
		if strings.TrimSpace(data.msg) != "" {
			reason = "refund saldo otomatis: " + strings.TrimSpace(data.msg)
		}
		if err := s.repo.RefundAppOrderFunding(ctx, *order.MemberID, order.InvoiceID, reason); err != nil {
			_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "failed")
			return 200, map[string]any{"ok": true, "refid": data.refid, "status": "failed", "refund_error": err.Error()}
		}
		_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "refunded")
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "refunded"}
	}
	if strings.EqualFold(strings.TrimSpace(order.BuyerType), "guest") && order.HargaFinal > 0 {
		_ = s.appOrderRepo.UpsertGuestRefundTicket(ctx, order, "refund guest pending claim: "+strings.TrimSpace(data.msg))
	}
	_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "failed")
	return 200, map[string]any{"ok": true, "refid": data.refid, "status": "failed"}
}

func parsePulsa24JamCallback(raw string, q url.Values, payload map[string]any) pulsa24JamCallbackData {
	if payload == nil {
		payload = map[string]any{}
	}
	get := func(keys ...string) string {
		for _, key := range keys {
			if v := strings.TrimSpace(q.Get(key)); v != "" {
				return v
			}
			for k, rawValue := range payload {
				if strings.EqualFold(strings.TrimSpace(k), key) {
					if v := strings.TrimSpace(fmt.Sprint(rawValue)); v != "" && v != "<nil>" {
						return v
					}
				}
			}
		}
		return ""
	}
	if len(payload) == 0 && strings.Contains(raw, "=") {
		if parsed, err := url.ParseQuery(raw); err == nil {
			for k, v := range parsed {
				if len(v) > 0 {
					payload[k] = v[0]
				}
			}
		}
	}
	price := parsePulsa24JamInt(get("price", "harga", "amount", "nominal"))
	balance := parsePulsa24JamInt(get("balance", "saldo", "saldo_terakhir"))
	status := strings.ToLower(strings.TrimSpace(get("status")))
	msg := firstText(get("msg"), get("message"), get("keterangan"), status)
	rc := firstText(get("rc"), get("code"), status)
	return pulsa24JamCallbackData{
		refid:       firstText(get("refid"), get("ref_id"), get("reffid")),
		status:      status,
		rc:          rc,
		msg:         msg,
		sn:          get("sn"),
		providerRef: firstText(get("provider_ref"), get("noref"), get("no_referensi")),
		price:       price,
		balance:     balance,
		payload:     payload,
	}
}

func pulsa24JamFinalStatus(data pulsa24JamCallbackData) string {
	switch strings.ToLower(strings.TrimSpace(data.status)) {
	case "2", "20", "00", "success", "sukses":
		return "success"
	case "3", "52", "55", "failed", "fail", "gagal", "error":
		return "failed"
	case "1", "68", "0068", "pending", "process", "processing":
		return "pending"
	}
	state := helper.ProviderResponseStateOf("pulsa24jam", data.rc, firstText(data.msg, data.status))
	switch state {
	case helper.ProviderResponseSuccess:
		return "success"
	case helper.ProviderResponseFailed:
		return "failed"
	default:
		return "pending"
	}
}

func parsePulsa24JamInt(v string) int64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	v = strings.ReplaceAll(v, ".", "")
	v = strings.ReplaceAll(v, ",", "")
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

func firstText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
