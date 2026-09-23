package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPulsa24JamCreateDepositQRIS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Api-Key"); got != "api-key" {
			t.Fatalf("X-Api-Key = %q", got)
		}
		var payload pulsa24JamPayRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Commands != "DEPOSIT-QRIS" || payload.RefID != "TOPUP-001" || payload.Qty != 50000 || payload.PIN != "123456" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"command":"DEPOSIT-QRIS","refid":"TOPUP-001","provider_refid":"DQR-H2H-ABC","amount":50000,"gross_amount":50000,"status":"pending","qr_url":"https://qr.example/1","balance":1000000}`))
	}))
	defer server.Close()

	adapter := NewPulsa24JamAdapter(Pulsa24JamConfig{
		BaseURL:  server.URL,
		APIKey:   "api-key",
		PIN:      "123456",
		Password: "password",
	})
	result, err := adapter.CreateDepositQRIS(context.Background(), "TOPUP-001", 50000)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProviderRefID != "DQR-H2H-ABC" || result.QRURL != "https://qr.example/1" || result.Amount != 50000 || result.Status != "pending" {
		t.Fatalf("unexpected response: %+v", result)
	}
}
