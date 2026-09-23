package service

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/repository"
)

func smpaySourceMarkerEnabled() bool {
	return envBool("P24_SMPAY_SOURCE_MARKER_ENABLED")
}

func smpaySourceMarkerPresent(in trxmemberdto.TrxRequest) bool {
	if strings.EqualFold(strings.TrimSpace(in.SourceSystem), "SMPAY") {
		return true
	}
	return in.SMPAYTransactionID > 0 || in.SMPAYWebsiteID > 0 || in.SMPAYDivisionID > 0 || in.SkipH2HCommission != nil
}

func smpaySourceMemberFallbackEnabled() bool {
	return envBool("P24_SMPAY_SOURCE_MEMBER_FALLBACK_ENABLED")
}

func smpaySourceMemberAllowed(memberID int64) bool {
	raw := strings.TrimSpace(os.Getenv("P24_SMPAY_SOURCE_MEMBER_IDS"))
	if raw == "" || memberID <= 0 {
		return false
	}
	for _, part := range strings.Split(raw, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil && id == memberID {
			return true
		}
	}
	return false
}

func smpaySourceMarkerShouldRecord(memberID int64, in trxmemberdto.TrxRequest) bool {
	if smpaySourceMarkerPresent(in) {
		return true
	}
	return smpaySourceMemberFallbackEnabled() && smpaySourceMemberAllowed(memberID)
}

func buildSMPAYSourceMarker(memberID, trxID int64, in trxmemberdto.TrxRequest) repository.SMPAYSourceMarker {
	skip := true
	if in.SkipH2HCommission != nil {
		skip = *in.SkipH2HCommission
	}
	raw, _ := json.Marshal(map[string]any{
		"commands":               strings.TrimSpace(strings.ToUpper(in.Commands)),
		"product":                strings.TrimSpace(in.Product),
		"refid":                  strings.TrimSpace(in.RefID),
		"source_system":          "SMPAY",
		"smpay_transaction_id":   in.SMPAYTransactionID,
		"smpay_website_id":       in.SMPAYWebsiteID,
		"smpay_division_id":      in.SMPAYDivisionID,
		"skip_h2h_commission":    skip,
		"p24_request_has_marker": smpaySourceMarkerPresent(in),
	})
	return repository.SMPAYSourceMarker{
		MemberID:           memberID,
		TransaksiMemberID:  trxID,
		RefID:              in.RefID,
		SMPAYTransactionID: in.SMPAYTransactionID,
		SMPAYWebsiteID:     in.SMPAYWebsiteID,
		SMPAYDivisionID:    in.SMPAYDivisionID,
		SourceSystem:       "SMPAY",
		SkipH2HCommission:  skip,
		RawRequestJSON:     string(raw),
	}
}

func (h *MemberTrxService) recordSMPAYRefSource(ctx context.Context, auth *repository.MemberAuth, in trxmemberdto.TrxRequest) *ServiceError {
	if !smpaySourceMarkerEnabled() || auth == nil || !smpaySourceMarkerShouldRecord(auth.MemberID, in) {
		return nil
	}
	if !smpaySourceMemberAllowed(auth.MemberID) {
		return &ServiceError{Kind: ErrForbidden, Message: "smpay source marker not allowed"}
	}
	if h == nil || h.MemberRepo == nil {
		return &ServiceError{Kind: ErrUpstream, Message: "repo marker belum tersedia"}
	}
	if err := h.MemberRepo.UpsertSMPAYRefSource(ctx, buildSMPAYSourceMarker(auth.MemberID, 0, in)); err != nil {
		h.logf("SMPAY marker ref upsert gagal member_id=%d refid=%s err=%v", auth.MemberID, in.RefID, err)
		return &ServiceError{Kind: ErrUpstream, Message: "gagal simpan marker smpay"}
	}
	return nil
}

func (h *MemberTrxService) attachSMPAYTransactionSource(ctx context.Context, auth *repository.MemberAuth, trxID int64, in trxmemberdto.TrxRequest) *ServiceError {
	if !smpaySourceMarkerEnabled() || auth == nil || !smpaySourceMarkerShouldRecord(auth.MemberID, in) {
		return nil
	}
	if !smpaySourceMemberAllowed(auth.MemberID) {
		return &ServiceError{Kind: ErrForbidden, Message: "smpay source marker not allowed"}
	}
	if h == nil || h.MemberRepo == nil || trxID <= 0 {
		return nil
	}
	if err := h.MemberRepo.AttachSMPAYTransactionSource(ctx, buildSMPAYSourceMarker(auth.MemberID, trxID, in)); err != nil {
		h.logf("SMPAY marker trx attach gagal member_id=%d trx_id=%d refid=%s err=%v", auth.MemberID, trxID, in.RefID, err)
		return &ServiceError{Kind: ErrUpstream, Message: "gagal attach marker smpay"}
	}
	return nil
}

func envBool(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
