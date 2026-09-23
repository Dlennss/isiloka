package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pulsa2/db"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/loketbayar"
)

func retryPendingNoFallbackProviderLeft(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "tidak ada provider fallback tersisa") ||
		strings.Contains(errMsg, "tidak ada provider eligible")
}

// RetryPendingTransactions — cari transaksi pending > 5 menit dan recheck status provider.
// Loket Bayar Otomax tidak punya endpoint advice terpisah; recheck memakai GET /trx
// dengan refid yang sama supaya provider menangani idempotensi.
// Provider lain yang punya retry aman tetap mengikuti flow lama.
// Dipanggil dari background goroutine tiap 1 menit.
func (h *MemberTrxService) RetryPendingTransactions(ctx context.Context) int {
	if h == nil || h.MemberRepo == nil || h.JPRepo == nil {
		return 0
	}
	h.retryMu.Lock()
	defer h.retryMu.Unlock()

	rows, err := h.MemberRepo.ListPendingOver(ctx, 5*time.Minute, 20)
	if err != nil {
		helper.AppendMemberTrxLog("RETRY_PENDING gagal list: %v", err)
		return 0
	}
	if loketRows, lErr := h.MemberRepo.ListWithPendingLoketBayarProviderOver(ctx, 5*time.Minute, 20); lErr != nil {
		helper.AppendMemberTrxLog("RETRY_PENDING gagal list loketbayar provider pending: %v", lErr)
	} else {
		rows = mergeRetryPendingRows(rows, loketRows)
	}

	return h.retryPendingRows(ctx, rows)
}

func (h *MemberTrxService) RetryPendingCallbackWaitProviders(ctx context.Context) int {
	if h == nil || h.MemberRepo == nil || h.JPRepo == nil {
		return 0
	}
	h.retryMu.Lock()
	defer h.retryMu.Unlock()

	rows, err := h.MemberRepo.ListWithPendingCallbackWaitProviderOver(ctx, 30*time.Second, 10)
	if err != nil {
		helper.AppendMemberTrxLog("RETRY_PENDING_CALLBACK_WAIT gagal list: %v", err)
		return 0
	}
	return h.retryPendingRows(ctx, rows)
}

func (h *MemberTrxService) retryPendingRows(ctx context.Context, rows []*repository.TrxMemberFull) int {
	retried := 0
	for _, trx := range rows {
		if trx == nil || strings.TrimSpace(trx.RefID) == "" {
			continue
		}

		// Ambil provider rows untuk refid ini
		attempts, lErr := h.JPRepo.ListByRefID(ctx, trx.RefID)
		if lErr != nil || len(attempts) == 0 {
			continue
		}

		// Cari provider row yang masih pending dan sudah > 5 menit
		var pendingProvider string
		var pendingKode string
		var pendingDest string
		var pendingRowID int64
		var pendingQty int64
		var pendingCreatedAt time.Time
		var pendingRequestMentah any
		var pendingHadProviderReply bool
		for _, att := range attempts {
			if att == nil {
				continue
			}
			prov := strings.ToLower(strings.TrimSpace(att.Provider))
			if ok, _, _ := providerRowPendingState(prov, att); ok && time.Since(att.DibuatPada) > 5*time.Minute {
				pendingProvider = prov
				pendingKode = strings.TrimSpace(att.KodeProduk)
				pendingDest = strings.TrimSpace(att.Tujuan)
				pendingRowID = att.ID
				pendingQty = att.Qty
				pendingCreatedAt = att.DibuatPada
				pendingRequestMentah = att.RequestMentah
				pendingHadProviderReply = providerRowHasProviderReply(att)
				break
			}
		}
		if pendingProvider == "" || pendingKode == "" || pendingRowID == 0 {
			continue
		}

		if pendingProvider == "loketbayar" {
			payQty := pendingQty
			if payQty <= 0 {
				payQty = trx.QtyProvider
			}
			if payQty <= 0 {
				payQty = trx.Qty
			}
			requestDest := strings.TrimSpace(pendingDest)
			if requestDest == "" {
				requestDest = strings.TrimSpace(trx.Tujuan)
			}

			if h.LBClient == nil {
				helper.AppendMemberTrxLog("RETRY_PENDING skip loketbayar trx_recheck refid=%s row_id=%d reason=client_nil", trx.RefID, pendingRowID)
				continue
			}

			helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck refid=%s kode=%s row_id=%d age_min=%d",
				trx.RefID, pendingKode, pendingRowID, int(time.Since(pendingCreatedAt).Minutes()))

			callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			resp, hs, reqRaw, callErr := h.LBClient.Advice(callCtx, loketbayar.TopupRequest{
				ProductCode: normalizeLoketRetryProduct(pendingKode),
				Dest:        requestDest,
				RefID:       trx.RefID,
				Nominal:     payQty,
			})
			cancel()

			if reqRaw != nil {
				reqRaw["endpoint"] = "/trx"
				reqRaw["retry_pending"] = true
				_ = h.JPRepo.UpdateRequestMentah(ctx, pendingRowID, reqRaw)
			}

			if callErr != nil {
				failMsg := callErr.Error()
				_ = h.JPRepo.UpdateResult(ctx, pendingRowID, db.UpdateResult{
					Pesan:        &failMsg,
					ResponMentah: map[string]any{"advice": true, "endpoint": "/trx", "error": failMsg},
				})
				helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck error refid=%s row_id=%d err=%v", trx.RefID, pendingRowID, callErr)
				retried++
				continue
			}

			raw := resp.Raw
			if raw == nil {
				raw = map[string]any{}
			}
			raw["advice"] = true
			raw["endpoint"] = "/trx"

			rc := strings.TrimSpace(resp.Status)
			msg := strings.TrimSpace(resp.Keterangan)
			providerRef := strings.TrimSpace(resp.Reff)
			if providerRef == "" {
				providerRef = strings.TrimSpace(resp.SN)
			}
			if providerRef == "" {
				providerRef = strings.TrimSpace(resp.TrxID)
			}
			body := strings.TrimSpace(fmt.Sprintf("status=%s keterangan=%s", rc, msg))

			if hs != 200 {
				waitMsg := fmt.Sprintf("loketbayar trx recheck menunggu balasan/callback provider http=%d", hs)
				raw["http_status"] = hs
				raw["provider_status"] = rc
				raw["provider_message"] = msg
				_ = h.JPRepo.UpdateResult(ctx, pendingRowID, db.UpdateResult{
					Pesan:        &waitMsg,
					ResponMentah: raw,
				})
				ketPending, _ := helper.SafeMemberKeterangan("pending", waitMsg)
				_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
				helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck http error/no-response refid=%s row_id=%d http=%d body=%s tetap pending",
					trx.RefID, pendingRowID, hs, body)
				retried++
				continue
			}

			upd := db.UpdateResult{
				HTTPStatus:   &hs,
				Pesan:        helper.PtrString(msg),
				Harga:        helper.PtrI64(resp.Price),
				NoReferensi:  helper.PtrString(providerRef),
				ResponMentah: raw,
			}
			if rc != "" {
				upd.KodeRespon = &rc
			}
			if resp.Saldo > 0 {
				upd.SaldoTerakhir = &resp.Saldo
			}
			_ = h.JPRepo.UpdateResult(ctx, pendingRowID, upd)

			status := helper.ProviderResponseStatusString("loketbayar", upd.KodeRespon, upd.Pesan)
			helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck respon refid=%s http=%d status=%s body=%s",
				trx.RefID, hs, status, body)

			if shouldKeepExistingMemberFinalStatus(trx.Status, status) {
				helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck provider_only refid=%s row_id=%d member_status=%s provider_status=%s",
					trx.RefID, pendingRowID, trx.Status, status)
				retried++
				continue
			}

			switch status {
			case "success":
				ket := strings.TrimSpace(providerRef)
				if ket == "" {
					ket = "Transaksi berhasil (loketbayar trx recheck)"
				}
				h.settleLikeCallback(ctx, trx, "success", ket, providerRef, resp.Price)
				st, _, _ := h.sendFinalWebhook(ctx, trx, "success", ket, providerRef, providerRef, resp.Price)
				helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck settle success refid=%s webhook=%d", trx.RefID, st)
			case "failed":
				ketFailed, _ := helper.SafeMemberKeterangan("failed", msg)
				h.settleLikeCallback(ctx, trx, "failed", ketFailed, strings.TrimSpace(providerRef), 0)
				st, _, _ := h.sendFinalWebhook(ctx, trx, "failed", ketFailed, strings.TrimSpace(providerRef), ketFailed, 0)
				helper.AppendMemberTrxLog("RETRY_PENDING loketbayar trx_recheck settle failed refid=%s webhook=%d msg=%s", trx.RefID, st, msg)
			}

			retried++
			continue
		}

		client := h.Clients[pendingProvider]
		if client == nil {
			continue
		}
		payQty := pendingQty
		if payQty <= 0 {
			payQty = trx.QtyProvider
		}
		if payQty <= 0 {
			payQty = trx.Qty
		}

		requestDest := strings.TrimSpace(pendingDest)
		if requestDest == "" {
			requestDest = strings.TrimSpace(trx.Tujuan)
		}

		payReq := provider.PayRequest{
			Command: "PAY",
			Product: pendingKode,
			Dest:    requestDest,
			Qty:     payQty,
			RefID:   trx.RefID,
		}
		applyProviderRetryRequestMeta(pendingRequestMentah, &payReq)

		helper.AppendMemberTrxLog("RETRY_PENDING PAY ulang refid=%s provider=%s kode=%s mode=%s row_id=%d age_min=%d",
			trx.RefID, pendingProvider, pendingKode, payReq.Mode, pendingRowID, int(time.Since(pendingCreatedAt).Minutes()))

		callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		resp, callErr := client.Pay(callCtx, payReq)
		cancel()

		// Update provider row dengan respon. Kalau sebelumnya provider sudah
		// pernah balas pending, timeout retry berikutnya tidak boleh menimpa
		// pesan provider menjadi failed internal.
		if resp == nil || !providerPayResponseHasProviderReply(resp) {
			msg := fmt.Sprintf("%s retry pending menunggu balasan/callback provider", pendingProvider)
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
			if pendingHadProviderReply {
				ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
				_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
				helper.AppendMemberTrxLog("RETRY_PENDING no-response setelah provider pernah balas; tahan pending refid=%s provider=%s row_id=%d http=%d err=%v",
					trx.RefID, pendingProvider, pendingRowID, hs, callErr)
				retried++
				continue
			}
			_ = h.JPRepo.UpdateResult(ctx, pendingRowID, db.UpdateResult{
				Pesan: &msg,
				ResponMentah: map[string]any{
					"retry_pending": true,
					"provider":      pendingProvider,
					"http_status":   hs,
					"body":          body,
					"error":         errText,
				},
			})
			helper.AppendMemberTrxLog("RETRY_PENDING nil/no-provider-response refid=%s provider=%s http=%d err=%v", trx.RefID, pendingProvider, hs, callErr)
			retried++
			continue
		}

		if resp.RequestRaw != nil {
			_ = h.JPRepo.UpdateRequestMentah(ctx, pendingRowID, resp.RequestRaw)
		}
		if resp.HTTPStatus != 200 {
			msg := fmt.Sprintf("%s retry pending menunggu balasan/callback provider http=%d", pendingProvider, resp.HTTPStatus)
			callErrText := ""
			if callErr != nil {
				callErrText = callErr.Error()
			}
			_ = h.JPRepo.UpdateResult(ctx, pendingRowID, db.UpdateResult{
				Pesan: &msg,
				ResponMentah: map[string]any{
					"retry_pending": true,
					"provider":      pendingProvider,
					"http_status":   resp.HTTPStatus,
					"body":          resp.Body,
					"message":       resp.Message,
					"error":         callErrText,
				},
			})
			ketPending, _ := helper.SafeMemberKeterangan("pending", msg)
			_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
			helper.AppendMemberTrxLog("RETRY_PENDING http error/no-response refid=%s provider=%s row_id=%d http=%d err=%v tetap pending",
				trx.RefID, pendingProvider, pendingRowID, resp.HTTPStatus, callErr)
			retried++
			continue
		}

		upd := db.UpdateResult{
			HTTPStatus:   &resp.HTTPStatus,
			Pesan:        helper.PtrString(resp.Message),
			Harga:        helper.PtrI64(resp.Price),
			NoReferensi:  helper.PtrString(resp.ProviderRef),
			ResponMentah: resp.Raw,
		}
		ambiguousRetry := retryPendingAmbiguousFailure(pendingProvider, resp.RC, resp.Message, resp.Body)
		if resp.RC != "" {
			if !ambiguousRetry {
				upd.KodeRespon = &resp.RC
			}
		}
		if resp.Balance > 0 {
			upd.SaldoTerakhir = &resp.Balance
		}
		_ = h.JPRepo.UpdateResult(ctx, pendingRowID, upd)

		// Classify: sukses → settle member, gagal → settle member failed
		status := helper.ProviderResponseStatusString(pendingProvider, upd.KodeRespon, upd.Pesan)
		if ambiguousRetry && status == "failed" {
			status = "pending"
		}
		helper.AppendMemberTrxLog("RETRY_PENDING respon refid=%s provider=%s http=%d status=%s body=%s",
			trx.RefID, pendingProvider, resp.HTTPStatus, status, resp.Body)

		switch status {
		case "success":
			ket := strings.TrimSpace(resp.ProviderRef)
			if ket == "" {
				ket = "Transaksi berhasil (retry)"
			}
			h.settleLikeCallback(ctx, trx, "success", ket, resp.ProviderRef, resp.Price)
			st, _, _ := h.sendFinalWebhook(ctx, trx, "success", ket, resp.ProviderRef, resp.ProviderRef, resp.Price)
			helper.AppendMemberTrxLog("RETRY_PENDING settle success refid=%s provider=%s webhook=%d", trx.RefID, pendingProvider, st)
		case "failed":
			ketFailed, _ := helper.SafeMemberKeterangan("failed", resp.Message)
			h.settleLikeCallback(ctx, trx, "failed", ketFailed, strings.TrimSpace(resp.ProviderRef), 0)
			st, _, _ := h.sendFinalWebhook(ctx, trx, "failed", ketFailed, strings.TrimSpace(resp.ProviderRef), ketFailed, 0)
			helper.AppendMemberTrxLog("RETRY_PENDING settle failed refid=%s provider=%s webhook=%d msg=%s", trx.RefID, pendingProvider, st, resp.Message)
		}
		// status == "pending" → biarkan, coba lagi di cycle berikutnya

		retried++
	}
	return retried
}

func normalizeLoketRetryProduct(raw string) string {
	raw = strings.TrimSpace(raw)
	if head, _, ok := strings.Cut(raw, ":"); ok {
		return strings.TrimSpace(head)
	}
	return raw
}

func mergeRetryPendingRows(primary, extra []*repository.TrxMemberFull) []*repository.TrxMemberFull {
	if len(extra) == 0 {
		return primary
	}
	seen := make(map[int64]bool, len(primary)+len(extra))
	out := make([]*repository.TrxMemberFull, 0, len(primary)+len(extra))
	for _, row := range primary {
		if row == nil {
			continue
		}
		seen[row.ID] = true
		out = append(out, row)
	}
	for _, row := range extra {
		if row == nil || seen[row.ID] {
			continue
		}
		seen[row.ID] = true
		out = append(out, row)
	}
	return out
}

func retryPendingAmbiguousFailure(providerName, rc, msg, body string) bool {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName != "smb" {
		return false
	}
	rc = strings.TrimSpace(rc)
	upper := strings.ToUpper(strings.TrimSpace(msg + " " + body))
	if rc == "0021" || rc == "21" {
		if strings.Contains(upper, "CEK TAGIHAN") || strings.Contains(upper, "TERLEBIH DAHULU") {
			return true
		}
	}
	return strings.Contains(upper, "SILAHKAN CEK TAGIHAN TERLEBIH DAHULU")
}
