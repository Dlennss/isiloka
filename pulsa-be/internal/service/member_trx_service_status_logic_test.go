package service

import (
	"strings"
	"testing"
	"time"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/repository"
)

func TestCanRetryFailedPaySameRefID(t *testing.T) {
	now := time.Now()

	if canRetryFailedPaySameRefID(nil, now) {
		t.Fatalf("nil existing should not be retryable")
	}

	base := map[string]any{
		"perintah":        "PAY",
		"status":          "failed",
		"keterangan":      "saldo tidak cukup",
		"diperbarui_pada": now.Add(-retrySameRefCooldown - time.Second),
	}
	if !canRetryFailedPaySameRefID(base, now) {
		t.Fatalf("expected retryable when cooldown has passed")
	}

	recent := map[string]any{
		"perintah":        "PAY",
		"status":          "failed",
		"keterangan":      "saldo tidak cukup",
		"diperbarui_pada": now.Add(-5 * time.Second),
	}
	if canRetryFailedPaySameRefID(recent, now) {
		t.Fatalf("did not expect retryable when cooldown has not passed")
	}

	wrongKet := map[string]any{
		"perintah":        "PAY",
		"status":          "failed",
		"keterangan":      "provider error",
		"diperbarui_pada": now.Add(-time.Minute),
	}
	if !canRetryFailedPaySameRefID(wrongKet, now) {
		t.Fatalf("expected retryable for generic failed reasons after cooldown")
	}
}

func TestBuildFinalWebhookPayloadIncludesExplicitNominals(t *testing.T) {
	trx := &repository.TrxMemberFull{
		ID:                    99,
		RefID:                 "REF123",
		Perintah:              "PAY",
		KodeProduk:            "DANA",
		Tujuan:                "08123",
		Qty:                   100000,
		QtyProvider:           99400,
		ChargeReceiverApplied: true,
		BiayaPerkiraan:        100000,
		FeeMemberRp:           600,
		HargaMember:           100000,
	}

	got := buildFinalWebhookPayload(trx, "success", "ok", "PR123", "SN123", 500000, 100000)
	trxOut, ok := got["trx"].(map[string]any)
	if !ok {
		t.Fatalf("trx payload missing or invalid")
	}

	if trxOut["qty"] != int64(100000) {
		t.Fatalf("unexpected qty: %#v", trxOut["qty"])
	}
	if trxOut["qty_provider"] != int64(99400) {
		t.Fatalf("unexpected qty_provider: %#v", trxOut["qty_provider"])
	}
	if trxOut["harga_member"] != int64(100000) {
		t.Fatalf("unexpected harga_member: %#v", trxOut["harga_member"])
	}
	if trxOut["biaya_aktual"] != int64(100000) {
		t.Fatalf("unexpected biaya_aktual: %#v", trxOut["biaya_aktual"])
	}
}

func TestBuildFinalWebhookPayloadFallsBackQtyProviderToQty(t *testing.T) {
	trx := &repository.TrxMemberFull{
		ID:             1,
		RefID:          "REF456",
		Perintah:       "PAY",
		KodeProduk:     "OVO",
		Tujuan:         "08123",
		Qty:            50000,
		BiayaPerkiraan: 50600,
	}

	got := buildFinalWebhookPayload(trx, "failed", "gagal", "", "", 0, 0)
	trxOut := got["trx"].(map[string]any)

	if trxOut["qty_provider"] != int64(50000) {
		t.Fatalf("unexpected fallback qty_provider: %#v", trxOut["qty_provider"])
	}
	if trxOut["harga_member"] != int64(50600) {
		t.Fatalf("unexpected fallback harga_member: %#v", trxOut["harga_member"])
	}
	if trxOut["sn"] != "Transaksi gagal" {
		t.Fatalf("expected failed payload sn to carry failure message, got %#v", trxOut["sn"])
	}
}

func TestBuildFinalWebhookPayloadSanitizesRawProviderFailure(t *testing.T) {
	raw := "loketbayar reject bisnis: . SALDO: 474534515"
	trx := &repository.TrxMemberFull{
		ID:             2,
		RefID:          "REFRAW",
		Perintah:       "PAY",
		KodeProduk:     "BCA",
		Tujuan:         "2981287071",
		Qty:            1000000,
		Status:         "failed",
		Keterangan:     raw,
		BiayaPerkiraan: 1000000,
		HargaMember:    1000000,
	}

	got := buildFinalWebhookPayload(trx, "failed", raw, raw, raw, 0, 0)
	trxOut := got["trx"].(map[string]any)

	for _, key := range []string{"message", "provider_ref", "sn"} {
		value, _ := trxOut[key].(string)
		upper := strings.ToUpper(value)
		if strings.Contains(upper, "LOKETBAYAR") || strings.Contains(upper, "SALDO") || strings.Contains(value, "474534515") {
			t.Fatalf("raw provider text leaked in %s: %q", key, value)
		}
	}
	if trxOut["message"] != "Transaksi gagal" {
		t.Fatalf("unexpected sanitized message: %#v", trxOut["message"])
	}
	if trxOut["provider_ref"] != "" {
		t.Fatalf("expected failed provider_ref to be blank, got %#v", trxOut["provider_ref"])
	}
}

func TestBuildDirectMemberWebhookPayloadSanitizesRawProviderFailure(t *testing.T) {
	raw := "smb reject bisnis: {\"success\":false,\"msg\":{\"sqlMessage\":\"Data too long\"}}"
	trx := &repository.CallbackTrxMemberFull{
		ID:             3,
		RefID:          "REFSMB",
		Perintah:       "PAY",
		KodeProduk:     "BCA",
		Tujuan:         "2981287071",
		Qty:            1000000,
		Status:         "failed",
		Keterangan:     raw,
		BiayaPerkiraan: 1000000,
		HargaMember:    1000000,
	}

	got := buildDirectMemberWebhookPayload(trx, "failed", raw, raw, raw, 0, 0)
	trxOut := got["trx"].(map[string]any)

	for _, key := range []string{"message", "provider_ref", "sn"} {
		value, _ := trxOut[key].(string)
		upper := strings.ToUpper(value)
		if strings.Contains(upper, "SMB") || strings.Contains(upper, "SQLMESSAGE") || strings.Contains(upper, "TRANSAKSIPIPA") {
			t.Fatalf("raw provider text leaked in %s: %q", key, value)
		}
	}
	if trxOut["message"] != "Transaksi gagal" {
		t.Fatalf("unexpected sanitized message: %#v", trxOut["message"])
	}
	if trxOut["provider_ref"] != "" {
		t.Fatalf("expected failed provider_ref to be blank, got %#v", trxOut["provider_ref"])
	}
}

func TestCanRetryFailedPaySameRefIDAllowsGenericFailed(t *testing.T) {
	now := time.Now()
	existing := map[string]any{
		"perintah":        "PAY",
		"status":          "failed",
		"keterangan":      "provider gagal",
		"diperbarui_pada": now.Add(-retrySameRefCooldown - time.Second),
	}
	if !canRetryFailedPaySameRefID(existing, now) {
		t.Fatalf("expected generic failed PAY to be retryable after cooldown")
	}
}

func TestCanRetryFailedPaySameRefIDBlocksAdminCanceled(t *testing.T) {
	now := time.Now()
	existing := map[string]any{
		"perintah":        "PAY",
		"status":          "failed",
		"keterangan":      "dibatalkan admin",
		"diperbarui_pada": now.Add(-retrySameRefCooldown - time.Second),
	}
	if canRetryFailedPaySameRefID(existing, now) {
		t.Fatalf("did not expect admin-canceled failed PAY to be retryable")
	}
}

func TestIsDuplicatePayWhileProcessing(t *testing.T) {
	tests := []struct {
		name     string
		commands string
		status   string
		want     bool
	}{
		{name: "pay pending", commands: "PAY", status: "pending", want: true},
		{name: "pay proses", commands: "PAY", status: "proses", want: true},
		{name: "pay process", commands: "PAY", status: "process", want: true},
		{name: "pay failed", commands: "PAY", status: "failed", want: false},
		{name: "pay success", commands: "PAY", status: "success", want: false},
		{name: "inq pending", commands: "INQ", status: "pending", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			existing := map[string]any{"status": tc.status}
			got := isDuplicatePayWhileProcessing(trxmemberdto.TrxRequest{Commands: tc.commands}, existing)
			if got != tc.want {
				t.Fatalf("unexpected duplicate-processing classification: got=%t want=%t", got, tc.want)
			}
		})
	}

	if isDuplicatePayWhileProcessing(trxmemberdto.TrxRequest{Commands: "PAY"}, nil) {
		t.Fatalf("nil existing should not be treated as duplicate in progress")
	}
}

func TestDuplicatePayGuardDoesNotBlockRetryableFailedSameRef(t *testing.T) {
	now := time.Now()
	existing := map[string]any{
		"perintah":        "PAY",
		"status":          "failed",
		"keterangan":      "provider gagal",
		"diperbarui_pada": now.Add(-retrySameRefCooldown - time.Second),
	}

	if isDuplicatePayWhileProcessing(trxmemberdto.TrxRequest{Commands: "PAY"}, existing) {
		t.Fatalf("failed transaction should not be treated as still processing")
	}
	if !canRetryFailedPaySameRefID(existing, now) {
		t.Fatalf("failed transaction after cooldown should remain retryable with same ref")
	}
}
