package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pulsa2/db"
	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

func pickLatestProviderSuccess(states []providerState) (string, *model.JavapayTrxRow, string, string, int64, string) {
	bestName := ""
	var bestRow *model.JavapayTrxRow
	bestRC := ""
	bestMsg := ""
	bestPrice := int64(0)
	bestNoRef := ""
	bestID := int64(0)

	for _, st := range states {
		ok, rc, msg, price, noreff := providerRowSuccessState(st.name, st.row)
		if !ok || st.row == nil {
			continue
		}
		if st.row.ID > bestID {
			bestID = st.row.ID
			bestName = st.name
			bestRow = st.row
			bestRC = rc
			bestMsg = msg
			bestPrice = price
			bestNoRef = noreff
		}
	}
	if bestRow == nil {
		return "", nil, "", "", 0, ""
	}
	return bestName, bestRow, bestRC, bestMsg, bestPrice, bestNoRef
}

func pickLatestProviderPending(states []providerState) (string, *model.JavapayTrxRow, string, string) {
	bestName := ""
	var bestRow *model.JavapayTrxRow
	bestRC := ""
	bestMsg := ""
	bestID := int64(0)

	for _, st := range states {
		pending, rc, msg := providerRowPendingState(st.name, st.row)
		if !pending || st.row == nil {
			continue
		}
		if st.row.ID > bestID {
			bestID = st.row.ID
			bestName = st.name
			bestRow = st.row
			bestRC = rc
			bestMsg = msg
		}
	}
	if bestRow == nil {
		return "", nil, "", ""
	}
	return bestName, bestRow, bestRC, bestMsg
}

func pickLatestProviderFailure(states []providerState) (string, *model.JavapayTrxRow, bool, string, string, int64, string) {
	bestName := ""
	var bestRow *model.JavapayTrxRow
	bestRetryable := false
	bestRC := ""
	bestMsg := ""
	bestPrice := int64(0)
	bestNoRef := ""
	bestID := int64(0)

	for _, st := range states {
		failed, retryable, rc, msg, price, noreff := providerRowFailureState(st.name, st.row)
		if !failed || st.row == nil {
			continue
		}
		if st.row.ID > bestID {
			bestID = st.row.ID
			bestName = st.name
			bestRow = st.row
			bestRetryable = retryable
			bestRC = rc
			bestMsg = msg
			bestPrice = price
			bestNoRef = noreff
		}
	}
	if bestRow == nil {
		return "", nil, false, "", "", 0, ""
	}
	return bestName, bestRow, bestRetryable, bestRC, bestMsg, bestPrice, bestNoRef
}

func (h *MemberTrxService) tryStatusPayFallback(ctx context.Context, trx *repository.TrxMemberFull, tried map[string]bool) (string, int64, error) {
	if trx == nil {
		return "", 0, fmt.Errorf("trx nil")
	}

	// Cek apakah ada provider lain yang sudah sukses/pending/masih proses
	// untuk refid ini — kalau ada, jangan fallback
	if h.JPRepo != nil {
		allAttempts, lErr := h.JPRepo.ListByRefID(ctx, trx.RefID)
		if lErr == nil {
			h.logf("GUARD tryStatusPayFallback refid=%s total_attempts=%d", trx.RefID, len(allAttempts))
			for _, att := range allAttempts {
				if att == nil {
					continue
				}
				rc := ""
				msg := ""
				if att.KodeRespon != nil {
					rc = *att.KodeRespon
				}
				if att.Pesan != nil {
					if len(*att.Pesan) > 80 {
						msg = (*att.Pesan)[:80]
					} else {
						msg = *att.Pesan
					}
				}
				provider := strings.ToLower(strings.TrimSpace(att.Provider))
				state := helper.ProviderResponseStateOf(provider, rc, msg)
				h.logf("GUARD tryStatusPayFallback refid=%s provider=%s id=%d rc=%s pesan=%s state=%s", trx.RefID, provider, att.ID, rc, msg, state)
				if ok, _, _, _, _ := providerRowSuccessState(provider, att); ok {
					h.logf("STATUS-PAY fallback skip refid=%s alasan=provider_sudah_sukses provider=%s", trx.RefID, provider)
					return "", 0, fmt.Errorf("provider %s sudah sukses untuk refid ini", provider)
				}
				if providerRowDefinitelyFailed(provider, att) {
					continue
				}
				if ok, _, _ := providerRowPendingState(provider, att); ok {
					h.logf("STATUS-PAY fallback skip refid=%s alasan=provider_masih_pending provider=%s age_min=%d", trx.RefID, provider, int(time.Since(att.DibuatPada).Minutes()))
					return "", 0, fmt.Errorf("provider %s masih pending untuk refid ini", provider)
				}
				// Bukan success, bukan definitelyFailed, bukan pending → status tidak jelas, block
				h.logf("STATUS-PAY fallback skip refid=%s alasan=provider_belum_final provider=%s", trx.RefID, provider)
				return "", 0, fmt.Errorf("provider %s belum final untuk refid ini", provider)
			}
			h.logf("GUARD tryStatusPayFallback refid=%s PASSED — all attempts definitely failed, proceeding with fallback", trx.RefID)
		} else {
			h.logf("GUARD tryStatusPayFallback refid=%s ListByRefID error=%v — BLOCKING fallback (safety)", trx.RefID, lErr)
			return "", 0, fmt.Errorf("gagal cek attempt existing refid=%s: %w", trx.RefID, lErr)
		}
	} else {
		h.logf("GUARD tryStatusPayFallback refid=%s JPRepo=nil — BLOCKING fallback (safety)", trx.RefID)
		return "", 0, fmt.Errorf("JPRepo nil, tidak bisa cek attempt existing")
	}

	qtyProvider := trx.QtyProvider
	if qtyProvider <= 0 {
		qtyProvider = trx.Qty
	}
	in := trxmemberdto.TrxRequest{
		Commands: "PAY",
		Product:  trx.KodeProduk,
		Dest:     trx.Tujuan,
		Qty:      qtyProvider,
		RefID:    trx.RefID,
	}

	// Fallback mode: sagaramobile boleh dipakai
	if tried == nil {
		tried = map[string]bool{}
	}
	tried["__fallback_mode__"] = true
	isBank := h.isBankH2HProduct(ctx, trx.KodeProduk)
	attempts := h.buildProviderAttempts(ctx, in, qtyProvider, true, tried)
	var lastUsed string
	var lastRowID int64
	var lastErr error
	for _, a := range attempts {
		if tried[routeAttemptKey(a.Name, a.ProdukProviderMapID, a.KodeProduk)] {
			continue
		}
		used, rowID, err := h.callProviderOnly(ctx, trx.ID, in, a, "status_pay_provider_failed_fallback")
		if err == nil {
			return used, rowID, nil
		}
		lastUsed = used
		lastRowID = rowID
		lastErr = err
		if isBank && (strings.EqualFold(strings.TrimSpace(a.Name), "smb") || strings.EqualFold(strings.TrimSpace(a.Name), "rajabiller")) {
			continue
		}
		if !isProviderSystemError(err) {
			return used, rowID, err
		}
	}
	if lastErr != nil {
		return lastUsed, lastRowID, lastErr
	}
	return "", 0, fmt.Errorf("tidak ada provider fallback tersisa")
}

type StaleSMBPendingRepairResult struct {
	TrxID            int64
	RefID            string
	Status           string
	FallbackProvider string
	ProviderRowID    int64
	Message          string
}

func (h *MemberTrxService) handleStaleSMBCheckPending(ctx context.Context, trx *repository.TrxMemberFull, row *model.JavapayTrxRow, tried map[string]bool) (*StaleSMBPendingRepairResult, error) {
	if trx == nil {
		return nil, fmt.Errorf("trx nil")
	}
	if row == nil {
		return nil, fmt.Errorf("smb row nil")
	}

	failMsg := buildStaleSMBCheckFailureMessage(row)
	originalMsg := ""
	if row.Pesan != nil {
		originalMsg = strings.TrimSpace(*row.Pesan)
	}
	failAt := time.Now()
	upd := db.UpdateResult{
		Pesan: &failMsg,
		ResponMentah: map[string]any{
			"status":                 false,
			"reason":                 "smb_check_callback_timeout",
			"stale_timeout_ms":       int(smbCheckCallbackTimeout / time.Millisecond),
			"failed_at":              failAt.Format(time.RFC3339Nano),
			"original_provider_body": originalMsg,
		},
	}
	if err := h.JPRepo.UpdateResult(ctx, row.ID, upd); err != nil {
		return nil, err
	}

	// SMB tidak punya recheck aman. Jangan dispatch PAY dari STATUS-PAY/recovery;
	// tunggu callback BYR atau fallback jika memang row CEK sudah final timeout.

	if provider, providerRowID, ferr := h.tryStatusPayFallback(ctx, trx, tried); ferr == nil {
		ketPending, _ := helper.SafeMemberKeterangan("pending", failMsg)
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		return &StaleSMBPendingRepairResult{
			TrxID:            trx.ID,
			RefID:            trx.RefID,
			Status:           "pending",
			FallbackProvider: provider,
			ProviderRowID:    providerRowID,
			Message:          failMsg,
		}, nil
	}

	if ok, reason := h.allExistingRouteAttemptsFinalFailed(ctx, trx.ID, trx.RefID); !ok {
		ketPending, _ := helper.SafeMemberKeterangan("pending", reason)
		_ = h.MemberRepo.UpdateTransaksiMemberStatus(ctx, trx.ID, "pending", ketPending, 0)
		h.logf("STATUS-PAY stale SMB tetap pending refid=%s reason=%s", trx.RefID, reason)
		return &StaleSMBPendingRepairResult{
			TrxID:   trx.ID,
			RefID:   trx.RefID,
			Status:  "pending",
			Message: reason,
		}, nil
	}

	ketFailed, _ := helper.SafeMemberKeterangan("failed", failMsg)
	h.settleLikeCallback(ctx, trx, "failed", ketFailed, "", 0)
	st, _, whErr := h.sendFinalWebhook(ctx, trx, "failed", ketFailed, "", ketFailed, 0)
	if whErr != nil {
		h.logf("STATUS-PAY stale SMB gagal kirim webhook refid=%s err=%v callback_status=%d", trx.RefID, whErr, st)
	}
	return &StaleSMBPendingRepairResult{
		TrxID:   trx.ID,
		RefID:   trx.RefID,
		Status:  "failed",
		Message: failMsg,
	}, nil
}

func (h *MemberTrxService) RepairStaleSMBPendingByTrxID(ctx context.Context, trxID int64) (*StaleSMBPendingRepairResult, error) {
	if trxID <= 0 {
		return nil, fmt.Errorf("invalid trx id")
	}

	trx, err := h.MemberRepo.GetTransaksiMemberByID(ctx, trxID)
	if err != nil {
		return nil, err
	}
	if trx == nil {
		return nil, fmt.Errorf("trx not found")
	}

	row, err := h.JPRepo.GetLatestByRefIDProvider(ctx, trx.RefID, "smb")
	if err != nil {
		return nil, err
	}
	if !isStaleSMBCheckOnlyPending(row, time.Now()) {
		return &StaleSMBPendingRepairResult{
			TrxID:   trx.ID,
			RefID:   trx.RefID,
			Status:  "skipped",
			Message: "not stale smb check pending",
		}, nil
	}

	allAttempts, listErr := h.JPRepo.ListByRefID(ctx, trx.RefID)
	if listErr != nil {
		return nil, listErr
	}
	tried := providerAttemptTriedKeys(allAttempts)
	return h.handleStaleSMBCheckPending(ctx, trx, row, tried)
}

func canRetryFailedPaySameRefID(existing map[string]any, now time.Time) bool {
	if existing == nil {
		return false
	}

	cmd := strings.TrimSpace(strings.ToUpper(helper.TrxToString(existing["perintah"])))
	status := strings.TrimSpace(strings.ToLower(helper.TrxToString(existing["status"])))
	ket := strings.TrimSpace(strings.ToLower(helper.TrxToString(existing["keterangan"])))

	if cmd != "PAY" || status != "failed" {
		return false
	}
	if strings.Contains(ket, "dibatalkan admin") {
		return false
	}

	if updatedAt, ok := existing["diperbarui_pada"].(time.Time); ok && !updatedAt.IsZero() {
		return now.Sub(updatedAt) >= retrySameRefCooldown
	}

	return true
}
