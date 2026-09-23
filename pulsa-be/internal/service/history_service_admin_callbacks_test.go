package service

import (
	"testing"

	"pulsa2/internal/repository"
)

func TestBuildAdminSendCallbackPayloadFailed(t *testing.T) {
	trx := &repository.AdminTrxCallbackTarget{
		ID:             123,
		RefID:          "junitestcancel1",
		Perintah:       "PAY",
		KodeProduk:     "DANA",
		Tujuan:         "081234567890",
		Qty:            300000,
		QtyProvider:    300000,
		BiayaPerkiraan: 301000,
		HargaMember:    301000,
	}

	payload := buildAdminSendCallbackPayload(trx, "failed", "dibatalkan admin: REFOUND", "", "", 99000, 0)

	if got := payload["status"]; got != "failed" {
		t.Fatalf("status = %v, want failed", got)
	}
	if got := payload["member_balance"]; got != int64(99000) {
		t.Fatalf("member_balance = %v, want 99000", got)
	}

	trxPayload, ok := payload["trx"].(map[string]any)
	if !ok {
		t.Fatalf("trx payload type = %T, want map[string]any", payload["trx"])
	}
	if got := trxPayload["biaya_aktual"]; got != int64(0) {
		t.Fatalf("biaya_aktual = %v, want 0", got)
	}
	if got := trxPayload["harga_member"]; got != int64(301000) {
		t.Fatalf("harga_member = %v, want 301000", got)
	}
	if got := trxPayload["sn"]; got != "dibatalkan admin: REFOUND" {
		t.Fatalf("sn = %v, want failure note", got)
	}
}
