package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/smb"
)

func (s *ProviderCallbackService) ProcessSMBCallback(ctx context.Context, rawQuery string, q url.Values) (int, map[string]any) {
	data := extractSMBCallbackData(q)
	if data.refid == "" {
		return s.handleSMBMissingRef(ctx, rawQuery, data)
	}

	row, err := s.getSMBCallbackRow(ctx, data.refid, data.routeCodeHint)
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		return s.handleSMBMissingRow(ctx, rawQuery, data)
	}

	if locked, _ := s.repo.AcquireCallbackLock(ctx, row.TransaksiMemberID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, row.TransaksiMemberID)
	}

	row, err = s.getSMBCallbackRow(ctx, data.refid, data.routeCodeHint)
	if err != nil || row == nil {
		return 200, map[string]any{"ok": true, "refid": data.refid}
	}

	trx, trxErr := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	isBank := false
	routeMode, routeCode, _, _ := smb.ParseMappedCodeTargetWithMode(ptrString(row.RequestMode), row.KodeProduk)
	if trxErr == nil && trx != nil {
		isBank = s.isBankH2HProduct(ctx, trx.KodeProduk)
	}

	if smbPayAlreadyProcessed(row) && !smb.LooksLikeSuccess(data.rawMsg) {
		prevSN := ""
		if row.NoReferensi != nil {
			prevSN = strings.TrimSpace(*row.NoReferensi)
		}
		prevPesan := ""
		if row.Pesan != nil {
			prevPesan = strings.TrimSpace(*row.Pesan)
		}

		helper.AppendProviderServiceLog("provider_callback_service.log",
			"SMB callback GAGAL setelah SUCCESS — refid=%s rowID=%d prev_sn=%s incoming_msg=%s",
			data.refid, row.ID, prevSN, data.rawMsg)

		data.msg = fmt.Sprintf("[SEMPAT SUKSES SN:%s] %s | sebelumnya: %s", prevSN, data.rawMsg, prevPesan)
	}

	if strings.EqualFold(strings.TrimSpace(row.Status), "failed") && data.stage == "check" && !isBank {
		helper.AppendProviderServiceLog("provider_callback_service.log",
			"SMB CHECK callback blocked — row already failed refid=%s rowID=%d", data.refid, row.ID)
		return 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "late_check_row_already_failed"}
	}

	if strings.EqualFold(strings.TrimSpace(row.Status), "success") && data.stage == "check" {
		helper.AppendProviderServiceLog("provider_callback_service.log",
			"SMB CHECK callback blocked — row already success refid=%s rowID=%d", data.refid, row.ID)
		return 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "late_check_row_already_success"}
	}

	if strings.EqualFold(strings.TrimSpace(row.Status), "success") && !smb.LooksLikeSuccess(data.rawMsg) {
		allowDowngrade := smbFailureAllowsDowngradeOrFallback(isBank, routeMode, routeCode, data.msg)
		if !allowDowngrade {
			helper.AppendProviderServiceLog("provider_callback_service.log",
				"SMB callback blocked — row already success refid=%s rowID=%d incoming=%s", data.refid, row.ID, data.stage)
			return 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "row_already_success"}
		}
	}

	if err := s.updateSMBProviderResult(ctx, row, data); err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "smb callback update result failed refid=%s rowID=%d stage=%s err=%v", data.refid, row.ID, data.stage, err)
		return 502, map[string]any{"ok": false, "error": err.Error(), "refid": data.refid}
	}

	if trxErr != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": data.refid}
	}

	if handled, status, out := s.processSMBCheckStage(ctx, row, trx, &data); handled {
		return status, out
	}
	return s.finalizeSMBCallback(ctx, row, trx, data)
}
