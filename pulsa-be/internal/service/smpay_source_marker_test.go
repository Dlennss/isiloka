package service

import (
	"testing"

	trxmemberdto "pulsa2/internal/dto/trx_member"
)

func TestSMPAYSourceMarkerPresent(t *testing.T) {
	if smpaySourceMarkerPresent(trxmemberdto.TrxRequest{}) {
		t.Fatalf("empty request should not be treated as SMPAY marker")
	}
	if !smpaySourceMarkerPresent(trxmemberdto.TrxRequest{SourceSystem: "smpay"}) {
		t.Fatalf("source_system SMPAY should be treated as marker")
	}
	if !smpaySourceMarkerPresent(trxmemberdto.TrxRequest{SMPAYTransactionID: 123}) {
		t.Fatalf("smpay transaction id should be treated as marker")
	}
}

func TestSMPAYSourceMemberAllowedRequiresExplicitList(t *testing.T) {
	t.Setenv("P24_SMPAY_SOURCE_MEMBER_IDS", "")
	if smpaySourceMemberAllowed(10) {
		t.Fatalf("empty allowlist must not allow marker writes")
	}

	t.Setenv("P24_SMPAY_SOURCE_MEMBER_IDS", "7, 10, 99")
	if !smpaySourceMemberAllowed(10) {
		t.Fatalf("expected member in allowlist to be allowed")
	}
	if smpaySourceMemberAllowed(11) {
		t.Fatalf("did not expect member outside allowlist to be allowed")
	}
}

func TestSMPAYSourceMarkerShouldRecordFallback(t *testing.T) {
	t.Setenv("P24_SMPAY_SOURCE_MEMBER_IDS", "7, 10, 99")
	t.Setenv("P24_SMPAY_SOURCE_MEMBER_FALLBACK_ENABLED", "")
	if smpaySourceMarkerShouldRecord(10, trxmemberdto.TrxRequest{}) {
		t.Fatalf("fallback disabled should not record empty marker request")
	}

	t.Setenv("P24_SMPAY_SOURCE_MEMBER_FALLBACK_ENABLED", "true")
	if !smpaySourceMarkerShouldRecord(10, trxmemberdto.TrxRequest{}) {
		t.Fatalf("fallback enabled should record allowlisted member")
	}
	if smpaySourceMarkerShouldRecord(11, trxmemberdto.TrxRequest{}) {
		t.Fatalf("fallback must not record non-allowlisted member")
	}
	if !smpaySourceMarkerShouldRecord(11, trxmemberdto.TrxRequest{SourceSystem: "SMPAY"}) {
		t.Fatalf("explicit marker should still be recognized before allowlist validation")
	}
}

func TestBuildSMPAYSourceMarkerDefaultsSkipTrue(t *testing.T) {
	marker := buildSMPAYSourceMarker(7, 88, trxmemberdto.TrxRequest{
		Commands:           "PAY",
		Product:            "DANA",
		RefID:              "ref-1",
		SourceSystem:       "SMPAY",
		SMPAYTransactionID: 99,
		SMPAYWebsiteID:     12,
		SMPAYDivisionID:    3,
	})
	if marker.MemberID != 7 || marker.TransaksiMemberID != 88 || marker.RefID != "ref-1" {
		t.Fatalf("unexpected marker identity: %#v", marker)
	}
	if marker.SourceSystem != "SMPAY" {
		t.Fatalf("source system = %q", marker.SourceSystem)
	}
	if !marker.SkipH2HCommission {
		t.Fatalf("skip_h2h_commission should default true for SMPAY marker")
	}
}
