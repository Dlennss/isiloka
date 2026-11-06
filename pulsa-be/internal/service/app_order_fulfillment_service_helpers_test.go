package service

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

func TestAppOrderProviderImmediateRejectPulsa24Jam(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "insufficient provider balance",
			body: `{"message":"saldo tidak cukup","ok":true,"refid":"INV-1","status":3}`,
			want: true,
		},
		{
			name: "numeric rejected status",
			body: `{"message":"produk tidak ditemukan","ok":true,"refid":"INV-2","status":3}`,
			want: true,
		},
		{
			name: "pending accepted",
			body: `{"message":"Transaksi sedang diproses","ok":true,"refid":"INV-3","status":"pending"}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appOrderProviderImmediateReject("pulsa24jam", tt.body); got != tt.want {
				t.Fatalf("appOrderProviderImmediateReject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppOrderProviderTimeoutPulsa24JamStaysPending(t *testing.T) {
	body := `{"ok":true,"refid":"PKA2B","status":1,"message":"Provider timeout"}`
	if appOrderProviderLooksLikeSystemIssue("pulsa24jam", body) {
		t.Fatal("pulsa24jam timeout must not be treated as system failure for app orders")
	}
	if !appOrderProviderLooksLikePending("pulsa24jam", body) {
		t.Fatal("pulsa24jam timeout should stay pending for status-pay/callback reconciliation")
	}
	if !appOrderProviderLooksLikeAccepted("pulsa24jam", body) {
		t.Fatal("pulsa24jam timeout should be accepted as pending")
	}
}

func TestAppOrderProviderProductUnavailable(t *testing.T) {
	if !appOrderProviderProductUnavailable("pulsa24jam", `{"message":"Produk kehabisan stok","status":3}`) {
		t.Fatal("out-of-stock response should quarantine the product")
	}
	if appOrderProviderProductUnavailable("pulsa24jam", `{"message":"Nomor tujuan salah","status":3}`) {
		t.Fatal("business rejection unrelated to stock must not quarantine the product")
	}
	if appOrderProviderProductUnavailable("yuscom", `{"message":"Produk kehabisan stok","status":3}`) {
		t.Fatal("only Pulsa24Jam products should be quarantined")
	}
}

func TestResolvePulsa24JamAppRequest(t *testing.T) {
	tests := []struct {
		name        string
		order       repository.AppOrderRow
		wantProduct string
		wantQty     int64
	}{
		{
			name:        "fixed dana keeps selected H2HR SKU",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "UDDND10", ProdukNamaSnapshot: "Dana 10.000", Qty: 1, HargaDasar: 11055},
			wantProduct: "UDDND10", wantQty: 1,
		},
		{
			name:        "fixed gopay keeps selected H2HR SKU",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "UDGP10", ProdukNamaSnapshot: "Gopay 10.000", Qty: 1, HargaDasar: 11650},
			wantProduct: "UDGP10", wantQty: 1,
		},
		{
			name:        "gopay driver keeps dedicated route",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "UDGD15", ProdukNamaSnapshot: "Gopay Driver 15.000", Qty: 1, HargaDasar: 16450},
			wantProduct: "UDGD15", wantQty: 1,
		},
		{
			name:        "open amount remains unchanged",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "DANA", ProdukNamaSnapshot: "Dana Bebas Nominal", Qty: 25000, HargaDasar: 26000},
			wantProduct: "DANA", wantQty: 25000,
		},
		{
			name:        "fixed transfer pulsa keeps unit quantity",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "ATF15", ProdukNamaSnapshot: "AXIS TRANSFER PULSA 15.000", Qty: 1, HargaDasar: 16568},
			wantProduct: "ATF15", wantQty: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, qty := resolvePulsa24JamAppRequest(tt.order.ProdukSKUSnapshot, &tt.order)
			if product != tt.wantProduct || qty != tt.wantQty {
				t.Fatalf("got product=%s qty=%d, want product=%s qty=%d", product, qty, tt.wantProduct, tt.wantQty)
			}
		})
	}
}

func TestPulsa24JamAppOrderRefIDFitsH2HRLimit(t *testing.T) {
	valid := regexp.MustCompile(`^PKA[A-Z0-9]{0,14}$`)
	for _, order := range []*repository.AppOrderRow{
		{ID: 1, InvoiceID: "INV-20261106173421-9AF1F343"},
		{ID: 81, InvoiceID: "INV-20260921170504-F95D2B50"},
		{InvoiceID: "ABCDEFGHIJKLMN"},
		{InvoiceID: "ABCDEFGHIJKLMNO"},
		{ID: math.MaxInt64},
		nil,
	} {
		got := pulsa24JamAppOrderRefID(order)
		if len(got) > 17 || !valid.MatchString(got) {
			t.Fatalf("invalid H2HR refid %q (length %d)", got, len(got))
		}
		if again := pulsa24JamAppOrderRefID(order); again != got {
			t.Fatalf("refid must be stable: %q != %q", got, again)
		}
	}
}

func TestPulsa24JamAppOrderRefIDUsesInvoiceBeforeID(t *testing.T) {
	got := pulsa24JamAppOrderRefID(&repository.AppOrderRow{
		ID:        10,
		InvoiceID: "INV-20261006170913-D083278E",
	})
	want := pulsa24JamAppOrderRefID(&repository.AppOrderRow{ID: 99, InvoiceID: "INV-20261006170913-D083278E"})
	if got != want {
		t.Fatalf("same invoice must retain its refid: %q != %q", got, want)
	}
	otherDay := pulsa24JamAppOrderRefID(&repository.AppOrderRow{ID: 10, InvoiceID: "INV-20261007170913-D083278E"})
	if got == otherDay {
		t.Fatal("different invoice dates must not collapse to the same refid")
	}
}

func TestPulsa24JamAppOrderRefIDFallbackFitsH2HRLimit(t *testing.T) {
	got := pulsa24JamAppOrderRefID(&repository.AppOrderRow{InvoiceID: "INV-20260921170504-F95D2B50"})
	if len(got) > 17 {
		t.Fatalf("refid length = %d, want <= 17: %s", len(got), got)
	}
	if got[:3] != "PKA" {
		t.Fatalf("refid = %q, want PKA prefix", got)
	}
}

func TestPulsa24JamAppOrderDispatchSendsSMBCompatibleReference(t *testing.T) {
	order := &repository.AppOrderRow{ID: 1, InvoiceID: "INV-20261106173421-9AF1F343", Dest: "TEST-DEST", Qty: 20000}
	var received struct {
		Command string `json:"commands"`
		Product string `json:"product"`
		Dest    string `json:"dest"`
		RefID   string `json:"refid"`
		Qty     int64  `json:"qty"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if len(received.RefID) > 17 {
			_, _ = w.Write([]byte(`{"ok":true,"status":3,"message":"refid terlalu panjang untuk H2HR SMB, maksimal 17 karakter"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"status":1,"message":"Transaksi sedang diproses"}`))
	}))
	defer server.Close()
	adapter := provider.NewPulsa24JamAdapter(provider.Pulsa24JamConfig{
		BaseURL: server.URL + "/v2/trx", APIKey: "test-key", PIN: "123456",
	})
	svc := NewAppOrderFulfillmentService(nil, nil, nil, nil, nil, nil, adapter)
	status, body, _, _, err := svc.callAppOrderProvider(context.Background(), "pulsa24jam", "GPCAD1000", order.Qty, order)
	if err != nil || status != http.StatusOK || !appOrderProviderLooksLikePending("pulsa24jam", body) {
		t.Fatalf("dispatch rejected: status=%d body=%s err=%v", status, body, err)
	}
	if received.RefID != pulsa24JamAppOrderRefID(order) || received.Command != "PAY" || received.Product != "GPCAD1000" || received.Qty != order.Qty || received.Dest != order.Dest {
		t.Fatalf("unexpected provider payload: %+v", received)
	}
}
