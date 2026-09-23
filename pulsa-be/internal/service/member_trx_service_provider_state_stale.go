package service

import (
	"strings"
	"time"

	"pulsa2/model"
	"pulsa2/smb"
)

func isStaleSMBCheckOnlyPending(row *model.JavapayTrxRow, now time.Time) bool {
	if row == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(row.Provider), "smb") {
		return false
	}
	if smbProviderRowLooksUnanswered(row) {
		if row.DibuatPada.IsZero() {
			return false
		}
		return now.Sub(row.DibuatPada) >= smbCheckCallbackTimeout
	}
	pending, _, msg := providerRowPendingState("smb", row)
	if !pending && !smbProviderRowHasCheckCallback(row) {
		return false
	}
	if !strings.Contains(strings.ToUpper(smbLastCheckMessage(row, msg)), "INQSUKSES") {
		return false
	}
	if row.DibuatPada.IsZero() {
		return false
	}
	return now.Sub(row.DibuatPada) >= smbCheckCallbackTimeout
}

func smbProviderRowLooksUnanswered(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	if providerRowHasFinalStatus(row) {
		return false
	}
	if smbProviderRowHasAnyCallback(row) {
		return false
	}
	if row.KodeRespon != nil && strings.TrimSpace(*row.KodeRespon) != "" {
		return false
	}
	if row.Pesan != nil && strings.TrimSpace(*row.Pesan) != "" {
		return false
	}
	if row.NoReferensi != nil && strings.TrimSpace(*row.NoReferensi) != "" {
		return false
	}
	if row.HTTPStatus != nil && *row.HTTPStatus > 0 {
		return false
	}
	if row.Harga != nil && *row.Harga > 0 {
		return false
	}
	return true
}

func yuscomProviderRowLooksUnanswered(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	if providerRowHasFinalStatus(row) {
		return false
	}
	if row.KodeRespon != nil && strings.TrimSpace(*row.KodeRespon) != "" {
		return false
	}
	if row.Pesan != nil && strings.TrimSpace(*row.Pesan) != "" {
		return false
	}
	if row.NoReferensi != nil && strings.TrimSpace(*row.NoReferensi) != "" {
		return false
	}
	if row.HTTPStatus != nil && *row.HTTPStatus > 0 {
		return false
	}
	if row.Harga != nil && *row.Harga > 0 {
		return false
	}
	return true
}

func providerRowHasFinalStatus(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(row.Status)) {
	case "success", "failed":
		return true
	default:
		return false
	}
}

func providerRowLooksUnanswered(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	if providerRowHasFinalStatus(row) {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(row.Provider), "smb") {
		return smbProviderRowLooksUnanswered(row)
	}
	return yuscomProviderRowLooksUnanswered(row)
}

func providerRowHasProviderReply(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	if row.KodeRespon != nil && strings.TrimSpace(*row.KodeRespon) != "" {
		return true
	}
	if row.Pesan != nil && strings.TrimSpace(*row.Pesan) != "" {
		return true
	}
	if row.NoReferensi != nil && strings.TrimSpace(*row.NoReferensi) != "" {
		return true
	}
	if row.HTTPStatus != nil && *row.HTTPStatus > 0 {
		return true
	}
	if row.Harga != nil && *row.Harga > 0 {
		return true
	}
	return false
}

func isStaleGenericProviderNoResponse(row *model.JavapayTrxRow, now time.Time) bool {
	if row == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(row.Provider), "yuscom") {
		return false
	}
	if !providerRowLooksUnanswered(row) {
		return false
	}
	if row.DibuatPada.IsZero() {
		return false
	}
	return now.Sub(row.DibuatPada) >= providerCallTimeout
}

func pickLatestStaleNoResponseProvider(states []providerState, now time.Time) (string, *model.JavapayTrxRow) {
	var latest *model.JavapayTrxRow
	name := ""
	for _, st := range states {
		if !isStaleGenericProviderNoResponse(st.row, now) {
			continue
		}
		if latest == nil || st.row.DibuatPada.After(latest.DibuatPada) {
			latest = st.row
			name = st.name
		}
	}
	return name, latest
}

func isStaleYuscomNoResponse(row *model.JavapayTrxRow, now time.Time) bool {
	if row == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(row.Provider), "yuscom") {
		return false
	}
	if !yuscomProviderRowLooksUnanswered(row) {
		return false
	}
	if row.DibuatPada.IsZero() {
		return false
	}
	return now.Sub(row.DibuatPada) >= providerCallTimeout
}

func buildStaleSMBCheckFailureMessage(row *model.JavapayTrxRow) string {
	base := "SMB GAGAL: callback bayar tidak diterima setelah cek sukses"
	if row == nil {
		return base
	}
	msg := strings.Join(strings.Fields(smbLastCheckMessage(row, "")), " ")
	if msg == "" {
		return base
	}
	return base + "; last_check=" + msg
}

func smbLastCheckMessage(row *model.JavapayTrxRow, fallback string) string {
	if row == nil {
		return strings.TrimSpace(fallback)
	}
	payload, ok := row.ResponMentah.(map[string]any)
	if ok && payload != nil {
		if strings.EqualFold(strings.TrimSpace(smbPayloadString(payload["stage"])), "check") {
			if msg := strings.TrimSpace(smbPayloadString(payload["message"])); msg != "" {
				return msg
			}
			if msg := strings.TrimSpace(smbPayloadString(payload["msg"])); msg != "" {
				return msg
			}
		}
	}
	if row.Pesan != nil && strings.TrimSpace(*row.Pesan) != "" {
		return strings.TrimSpace(*row.Pesan)
	}
	return strings.TrimSpace(fallback)
}

func smbPayloadString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func smbProviderRowCallbackStage(row *model.JavapayTrxRow) string {
	if row == nil {
		return ""
	}
	payload, ok := row.ResponMentah.(map[string]any)
	if !ok || payload == nil {
		return ""
	}
	if stage, ok := payload["stage"].(string); ok {
		return strings.TrimSpace(strings.ToLower(stage))
	}
	return ""
}

func smbProviderRowHasCheckCallback(row *model.JavapayTrxRow) bool {
	return smbProviderRowCallbackStage(row) == "check"
}

func smbProviderRowHasAnyCallback(row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	if smbProviderRowCallbackStage(row) != "" {
		return true
	}
	payload, ok := row.ResponMentah.(map[string]any)
	if !ok || payload == nil {
		return false
	}
	for _, key := range []string{"clientid", "serverid", "statuscode", "message", "msg"} {
		if _, exists := payload[key]; exists {
			return true
		}
	}
	return false
}

func smbDirectShouldWaitFinalCallback(httpStatus int, body string) bool {
	if httpStatus != 200 {
		return false
	}
	if smb.LooksLikeSystemIssue(body) {
		return false
	}
	if smb.LooksLikeImmediateReject(body) {
		return false
	}
	return smb.LooksLikePending(body)
}
