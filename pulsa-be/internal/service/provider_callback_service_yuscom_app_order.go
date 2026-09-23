package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) processYuscomAppCallback(ctx context.Context, refid string, statusNum int, price int64, msg string, q url.Values) (int, map[string]any) {
	row, err := s.appProviderRepo.GetByRefID(ctx, refid, "yuscom")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	order, err := s.appOrderRepo.GetByID(ctx, row.AppOrderID)
	if err != nil || order == nil {
		return 200, map[string]any{"ok": true, "refid": refid, "ignored": true}
	}

	finalStatus := "failed"
	if statusNum == 20 || strings.Contains(strings.ToUpper(msg), "SUKSES") {
		finalStatus = "success"
	} else if statusNum == 0 || statusNum == 1 || statusNum == 2 || statusNum == 3 {
		finalStatus = "pending"
	} else if strings.Contains(strings.ToUpper(msg), "GAGAL") {
		finalStatus = "failed"
	}

	_, snParsed := providersn.ParseYuscomSNRefFromMsg(msg)
	snParsed = strings.TrimSpace(snParsed)
	rcStr := strconv.Itoa(statusNum)
	rawJSON, _ := json.Marshal(map[string]any{
		"t":       q.Get("t"),
		"refid":   refid,
		"status":  statusNum,
		"price":   price,
		"message": msg,
	})
	updateIn := repository.AppOrderProviderTrxUpdateInput{
		ID:          row.ID,
		Status:      finalStatus,
		KodeRespon:  rcStr,
		Pesan:       msg,
		SN:          snParsed,
		RawCallback: string(rawJSON),
	}
	if price > 0 {
		updateIn.HargaProvider = &price
	}
	if finalStatus == "pending" {
		updateIn.Status = "pending"
	}

	if order.Status == "success" || order.Status == "failed" || order.Status == "refunded" {
		if order.Status == "success" && finalStatus == "failed" {
			if err := s.refundLateFailedYuscomAppOrder(ctx, refid, order, row, updateIn, price, msg); err != nil {
				helper.AppendProviderServiceLog("provider_callback_error.log", "yuscom app late failed refund failed refid=%s app_order_id=%d app_provider_id=%d err=%v", refid, order.ID, row.ID, err)
				return 200, map[string]any{"ok": true, "refid": refid, "status": "success", "late_failed": true, "refund_error": err.Error()}
			}
			helper.AppendAlreadyFinalLog("callback late_failed_refunded provider=yuscom refid=%s app_order_id=%d previous_status=success callback_status=failed", refid, order.ID)
			return 200, map[string]any{"ok": true, "refid": refid, "status": "refunded", "late_failed": true}
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=yuscom refid=%s app_order_id=%d status=%s callback_status=%s", refid, order.ID, order.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid}
	}

	if err := s.appProviderRepo.UpdateResult(ctx, updateIn); err != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log", "yuscom app callback update failed refid=%s app_provider_id=%d err=%v", refid, row.ID, err)
	}

	if finalStatus == "pending" {
		_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "processing_provider")
		return 200, map[string]any{"ok": true, "refid": refid, "status": "pending"}
	}

	if finalStatus == "success" {
		if price > 0 {
			appProviderID := row.ID
			if _, _, err := s.repo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
				Provider:              "yuscom",
				RefID:                 refid,
				Arah:                  "debit",
				Jumlah:                price,
				Alasan:                "APP_TRX_SUCCESS_COST",
				Catatan:               "auto debit by callback (app success)",
				AppOrderProviderTrxID: &appProviderID,
			}); err != nil {
				helper.AppendProviderServiceLog("provider_wallet.log", "provider wallet debit app failed provider=yuscom refid=%s app_provider_id=%d err=%v", refid, row.ID, err)
			}
		}
		_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "success")
		if s.retailRepo != nil {
			if err := s.retailRepo.ApplyCommissionForOrder(ctx, order.ID); err != nil {
				helper.AppendProviderServiceLog("provider_wallet.log", "retail commission apply failed invoice=%s order_id=%d err=%v", refid, order.ID, err)
			}
		}
		return 200, map[string]any{"ok": true, "refid": refid, "status": "success"}
	}

	if order.BuyerType == "user" && order.MemberID != nil && *order.MemberID > 0 && order.HargaFinal > 0 {
		catatan := "refund saldo otomatis karena transaksi provider aplikasi gagal"
		if strings.TrimSpace(msg) != "" {
			catatan = "refund saldo otomatis: " + strings.TrimSpace(msg)
		}
		if err := s.repo.CreditDompet(ctx, *order.MemberID, order.InvoiceID, order.HargaFinal, "APP_ORDER_REFUND", catatan); err != nil {
			helper.AppendProviderServiceLog("provider_wallet.log", "app order refund failed member_id=%d refid=%s order_id=%d err=%v", *order.MemberID, refid, order.ID, err)
			_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "failed")
			return 200, map[string]any{"ok": true, "refid": refid, "status": "failed", "refund_error": err.Error()}
		}
		_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "refunded")
		return 200, map[string]any{"ok": true, "refid": refid, "status": "refunded"}
	}

	if strings.TrimSpace(strings.ToLower(order.BuyerType)) == "guest" && order.HargaFinal > 0 {
		reason := "refund guest pending claim karena transaksi provider aplikasi gagal"
		if strings.TrimSpace(msg) != "" {
			reason = "refund guest pending claim: " + strings.TrimSpace(msg)
		}
		if err := s.appOrderRepo.UpsertGuestRefundTicket(ctx, order, reason); err != nil {
			helper.AppendProviderServiceLog("provider_callback_service.log", "guest refund ticket create failed refid=%s order_id=%d err=%v", refid, order.ID, err)
		}
	}

	_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "failed")
	return 200, map[string]any{"ok": true, "refid": refid, "status": "failed"}
}

func (s *ProviderCallbackService) refundLateFailedYuscomAppOrder(ctx context.Context, refid string, order *repository.AppOrderRow, row *repository.AppOrderProviderTrxRow, updateIn repository.AppOrderProviderTrxUpdateInput, price int64, msg string) error {
	if order == nil || row == nil {
		return nil
	}

	if err := s.appProviderRepo.UpdateResult(ctx, updateIn); err != nil {
		return err
	}

	refundProviderAmount := price
	if refundProviderAmount <= 0 {
		refundProviderAmount = row.HargaProvider
	}
	if refundProviderAmount > 0 {
		appProviderID := row.ID
		if _, _, err := s.repo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
			Provider:              "yuscom",
			RefID:                 refid,
			Arah:                  "credit",
			Jumlah:                refundProviderAmount,
			Alasan:                "APP_TRX_SUCCESS_COST_REFUND",
			Catatan:               "auto credit by late failed/refund callback (app order)",
			AppOrderProviderTrxID: &appProviderID,
		}); err != nil {
			return err
		}
	}

	reason := "refund saldo otomatis karena callback gagal/refund Yuscom setelah transaksi retail sukses"
	if strings.TrimSpace(msg) != "" {
		reason = "refund saldo otomatis: " + strings.TrimSpace(msg)
	}

	if order.BuyerType == "user" && order.MemberID != nil && *order.MemberID > 0 && order.HargaFinal > 0 {
		if err := s.repo.CreditDompet(ctx, *order.MemberID, order.InvoiceID, order.HargaFinal, "APP_ORDER_REFUND", reason); err != nil {
			_ = s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "failed")
			return err
		}
		if s.retailRepo != nil {
			if err := s.retailRepo.ReverseCommissionForOrder(ctx, order.ID, "app order refunded by late Yuscom failed/refund callback"); err != nil {
				helper.AppendProviderServiceLog("provider_wallet.log", "retail commission reversal failed invoice=%s order_id=%d err=%v", refid, order.ID, err)
			}
		}
		return s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "refunded")
	}

	if strings.TrimSpace(strings.ToLower(order.BuyerType)) == "guest" && order.HargaFinal > 0 {
		if err := s.appOrderRepo.UpsertGuestRefundTicket(ctx, order, "refund guest pending claim: "+reason); err != nil {
			helper.AppendProviderServiceLog("provider_callback_service.log", "guest refund ticket create failed refid=%s order_id=%d err=%v", refid, order.ID, err)
		}
	}
	return s.appOrderRepo.UpdateStatusByID(ctx, order.ID, "failed")
}
