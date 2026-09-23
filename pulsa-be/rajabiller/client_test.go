package rajabiller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseTransactionResponseReadsTokenAndSaldoTerpotong(t *testing.T) {
	resp := parseTransactionResponse(map[string]any{
		"trxid":           "ref1",
		"rc":              "00",
		"status":          "Sukses",
		"token":           "1234-5678-9012-3456-7890",
		"saldo_terpotong": "83645",
		"saldo_akhir":     "14827574",
	})

	if resp.Token != "1234-5678-9012-3456-7890" {
		t.Fatalf("Token = %q", resp.Token)
	}
	if resp.Price != 83645 {
		t.Fatalf("Price = %d", resp.Price)
	}
	if resp.Balance != 14827574 {
		t.Fatalf("Balance = %d", resp.Balance)
	}
}

func TestTransactionSendsBankFieldsAndRedactsPIN(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rc":          "00",
			"status":      "SUCCESS",
			"harga":       "335000",
			"saldo_akhir": "1000000",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "uid", "secret-pin", time.Second)
	_, hs, reqLog, err := c.Transaction(context.Background(), TransactionRequest{
		Method:      "bayar",
		Product:     "BLTRFAG",
		Dest:        "0710051127",
		RefID:       "ref-bank-1",
		Nominal:     335000,
		SendNominal: true,
		KodeBank:    "008",
		HP:          "085648889293",
		Berita:      "untuk test",
		MerchantID:  "KMC112201",
	})
	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}
	if hs != http.StatusOK {
		t.Fatalf("HTTP status = %d", hs)
	}

	want := map[string]string{
		"method":      "bayar",
		"uid":         "uid",
		"pin":         "secret-pin",
		"produk":      "BLTRFAG",
		"idpel":       "0710051127",
		"ref1":        "ref-bank-1",
		"nominal":     "335000",
		"kodebank":    "008",
		"hp":          "085648889293",
		"berita":      "untuk test",
		"id_merchant": "KMC112201",
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("request[%s] = %#v, want %#v", k, got[k], v)
		}
	}
	if reqLog["pin"] != "***" {
		t.Fatalf("request log pin = %#v, want redacted", reqLog["pin"])
	}
}

func TestTransactionUsesDefaultMerchantID(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rc":     "00",
			"status": "SUCCESS",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "uid", "secret-pin", time.Second)
	c.MerchantID = "PROD-MERCHANT"
	_, _, _, err := c.Transaction(context.Background(), TransactionRequest{
		Method:      "cek",
		Product:     "BLTRFAG",
		Dest:        "0710051127",
		RefID:       "ref-bank-2",
		Nominal:     335000,
		SendNominal: true,
		KodeBank:    "014",
		HP:          "085648889293",
	})
	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	if got["id_merchant"] != "PROD-MERCHANT" {
		t.Fatalf("id_merchant = %#v, want default merchant", got["id_merchant"])
	}
}
