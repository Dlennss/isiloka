package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pulsa2/rajabiller"
)

func TestRajabillerShouldSendNominal(t *testing.T) {
	tests := []struct {
		name    string
		product string
		mode    string
		want    bool
	}{
		{name: "pln prabayar", product: "PLNPRAH", want: true},
		{name: "emoney open prefix", product: "EMOVOH", want: true},
		{name: "credit card mode", product: "BCA", mode: "KARTU_KREDIT", want: true},
		{name: "explicit nominal mode", product: "OVO", mode: "OPEN_DENOM", want: true},
		{name: "bank transfer mode", product: "BLTRFAG", mode: "BANK_TRANSFER", want: true},
		{name: "fixed denom omits nominal", product: "OVO50K", want: false},
		{name: "explicit no nominal wins", product: "PLNPRAH", mode: "NO_NOMINAL", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := rajabillerShouldSendNominal(tc.product, tc.mode); got != tc.want {
				t.Fatalf("rajabillerShouldSendNominal() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSplitRajabillerProductBankUsesSpecialCodeAsKodeBank(t *testing.T) {
	parts := splitRajabillerProduct("008:BLTRFAG", "BANK_TRANSFER")
	if parts.Product != "BLTRFAG" || parts.KodeBank != "008" || parts.Server != "" {
		t.Fatalf("unexpected bank parts: %#v", parts)
	}
}

func TestSplitRajabillerProductBankUsesTidyDBLayout(t *testing.T) {
	parts := splitRajabillerProduct("BLTRFAG:008", "BANK_TRANSFER")
	if parts.Product != "BLTRFAG" || parts.KodeBank != "008" || parts.Server != "" {
		t.Fatalf("unexpected bank parts: %#v", parts)
	}
}

func TestSplitRajabillerProductNonBankKeepsSpecialCodeAsServer(t *testing.T) {
	parts := splitRajabillerProduct("SERVER1:EMDANAX", "OPEN_DENOM")
	if parts.Product != "EMDANAX" || parts.Server != "SERVER1" || parts.KodeBank != "" {
		t.Fatalf("unexpected non-bank parts: %#v", parts)
	}
}

func TestRajabillerBankPayRunsInquiryBeforePayment(t *testing.T) {
	var requests []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, got)
		switch got["method"] {
		case "cek":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"rc":          "00",
				"status":      "Sukses",
				"tagihan":     "335000",
				"total_bayar": "341500",
			})
		case "bayar":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"trxid":       "ref-bank-1",
				"rc":          "68",
				"status":      "Sedang diproses",
				"refid":       "319859875",
				"harga":       "335000",
				"saldo_akhir": "1000000",
			})
		default:
			t.Fatalf("unexpected method: %#v", got["method"])
		}
	}))
	defer srv.Close()

	adapter := &RajabillerAdapter{C: rajabiller.New(srv.URL, "uid", "pin", time.Second)}
	resp, err := adapter.Pay(context.Background(), PayRequest{
		Command:    "PAY",
		Product:    "BLTRFAG:008",
		Mode:       "BANK_TRANSFER",
		Dest:       "0710051127",
		Qty:        335000,
		RefID:      "ref-bank-1",
		HP:         "085648889293",
		Berita:     "untuk test",
		MerchantID: "MERCHANT-A",
	})
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d, want 2", len(requests))
	}
	if requests[0]["method"] != "cek" || requests[1]["method"] != "bayar" {
		t.Fatalf("methods = %#v then %#v, want cek then bayar", requests[0]["method"], requests[1]["method"])
	}
	if requests[0]["produk"] != "BLTRFAG" || requests[0]["kodebank"] != "008" || requests[0]["ref1"] != "ref-bank-1" {
		t.Fatalf("unexpected inquiry request: %#v", requests[0])
	}
	if requests[1]["produk"] != "BLTRFAG" || requests[1]["kodebank"] != "008" || requests[1]["ref1"] != "ref-bank-1" {
		t.Fatalf("unexpected payment request: %#v", requests[1])
	}
	if requests[0]["berita"] != "untuk test" {
		t.Fatalf("inquiry berita = %#v, want passed through", requests[0]["berita"])
	}
	if _, ok := requests[1]["berita"]; ok {
		t.Fatalf("payment request must omit berita: %#v", requests[1])
	}
	if requests[0]["id_merchant"] != "MERCHANT-A" || requests[1]["id_merchant"] != "MERCHANT-A" {
		t.Fatalf("merchant id not forwarded to bank cek/pay requests: %#v %#v", requests[0], requests[1])
	}
	if resp.RC != "68" || resp.ProviderRef != "319859875" {
		t.Fatalf("response = rc %q ref %q, want pending provider ref", resp.RC, resp.ProviderRef)
	}
	if resp.RequestRaw["cek_response"] == nil {
		t.Fatalf("RequestRaw missing cek_response: %#v", resp.RequestRaw)
	}
}

func TestRajabillerBankPayStopsWhenInquiryFails(t *testing.T) {
	var requests []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, got)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rc":     "03",
			"status": "ID Pelanggan tidak terdaftar",
		})
	}))
	defer srv.Close()

	adapter := &RajabillerAdapter{C: rajabiller.New(srv.URL, "uid", "pin", time.Second)}
	resp, err := adapter.Pay(context.Background(), PayRequest{
		Command: "PAY",
		Product: "BLTRFAG:014",
		Mode:    "BANK_TRANSFER",
		Dest:    "0710051127",
		Qty:     335000,
		RefID:   "ref-bank-2",
		HP:      "085648889293",
	})
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("request count = %d, want only inquiry", len(requests))
	}
	if requests[0]["method"] != "cek" {
		t.Fatalf("method = %#v, want cek", requests[0]["method"])
	}
	if resp.RC != "03" {
		t.Fatalf("RC = %q, want inquiry failure", resp.RC)
	}
}

func TestRajabillerOpenDenomPayRunsInquiryBeforePayment(t *testing.T) {
	var requests []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, got)
		switch got["method"] {
		case "cek":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"rc":          "00",
				"status":      "Sukses",
				"produk":      "EMDANAX",
				"idpel":       "081234567890",
				"total_bayar": "50000",
			})
		case "bayar":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"trxid":       "ref-ewallet-1",
				"rc":          "00",
				"status":      "Sukses",
				"refid":       "rb-pay-1",
				"harga":       "50000",
				"saldo_akhir": "1000000",
			})
		default:
			t.Fatalf("unexpected method: %#v", got["method"])
		}
	}))
	defer srv.Close()

	adapter := &RajabillerAdapter{C: rajabiller.New(srv.URL, "uid", "pin", time.Second)}
	resp, err := adapter.Pay(context.Background(), PayRequest{
		Command: "PAY",
		Product: "EMDANAX",
		Mode:    "OPEN_DENOM",
		Dest:    "081234567890",
		Qty:     50000,
		RefID:   "ref-ewallet-1",
	})
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d, want 2", len(requests))
	}
	if requests[0]["method"] != "cek" || requests[1]["method"] != "bayar" {
		t.Fatalf("methods = %#v then %#v, want cek then bayar", requests[0]["method"], requests[1]["method"])
	}
	for i, got := range requests {
		if got["produk"] != "EMDANAX" || got["idpel"] != "081234567890" || got["ref1"] != "ref-ewallet-1" {
			t.Fatalf("request %d has unexpected identity fields: %#v", i, got)
		}
		if got["nominal"] != "50000" {
			t.Fatalf("request %d nominal = %#v, want 50000", i, got["nominal"])
		}
		if _, ok := got["kodebank"]; ok {
			t.Fatalf("open denom request must omit kodebank: %#v", got)
		}
		if _, ok := got["hp"]; ok {
			t.Fatalf("open denom request must omit hp: %#v", got)
		}
		if _, ok := got["berita"]; ok {
			t.Fatalf("open denom request must omit berita: %#v", got)
		}
	}
	if resp.RC != "00" || resp.ProviderRef != "rb-pay-1" {
		t.Fatalf("response = rc %q ref %q, want payment success provider ref", resp.RC, resp.ProviderRef)
	}
	if resp.RequestRaw["cek_response"] == nil {
		t.Fatalf("RequestRaw missing cek_response: %#v", resp.RequestRaw)
	}
}

func TestRajabillerOpenDenomPayStopsWhenInquiryFails(t *testing.T) {
	var requests []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, got)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rc":     "16",
			"status": "data inquiry tidak ditemukan",
		})
	}))
	defer srv.Close()

	adapter := &RajabillerAdapter{C: rajabiller.New(srv.URL, "uid", "pin", time.Second)}
	resp, err := adapter.Pay(context.Background(), PayRequest{
		Command: "PAY",
		Product: "EMOVOH",
		Mode:    "OPEN_DENOM",
		Dest:    "081234567890",
		Qty:     25000,
		RefID:   "ref-ewallet-2",
	})
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("request count = %d, want only inquiry", len(requests))
	}
	if requests[0]["method"] != "cek" {
		t.Fatalf("method = %#v, want cek", requests[0]["method"])
	}
	if requests[0]["produk"] != "EMOVOH" {
		t.Fatalf("produk = %#v, want mapped provider sku EMOVOH", requests[0]["produk"])
	}
	if resp.RC != "16" {
		t.Fatalf("RC = %q, want inquiry failure", resp.RC)
	}
}

func TestRajabillerBankPayDefaultsHPAndBeritaForInquiry(t *testing.T) {
	var requests []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, got)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rc":     "16",
			"status": "invalid request",
		})
	}))
	defer srv.Close()

	adapter := &RajabillerAdapter{C: rajabiller.New(srv.URL, "uid", "pin", time.Second)}
	resp, err := adapter.Pay(context.Background(), PayRequest{
		Command: "PAY",
		Product: "BLTRFAG:008",
		Mode:    "BANK_TRANSFER",
		Dest:    "1730014386461",
		Qty:     20000,
		RefID:   "ref-bank-defaults",
	})
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("request count = %d, want only inquiry", len(requests))
	}
	hp, _ := requests[0]["hp"].(string)
	if !strings.HasPrefix(hp, "08") || len(hp) != 12 {
		t.Fatalf("hp = %#v, want generated 08xxxxxxxxxx", requests[0]["hp"])
	}
	if hp == "1730014386461" {
		t.Fatalf("hp must not default to destination account: %#v", requests[0])
	}
	if requests[0]["berita"] != rajabillerBankDefaultBerita {
		t.Fatalf("berita = %#v, want default %q", requests[0]["berita"], rajabillerBankDefaultBerita)
	}
	if resp.RC != "16" {
		t.Fatalf("RC = %q, want inquiry failure", resp.RC)
	}
}
