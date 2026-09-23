package service

import (
	"testing"

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
			name:        "fixed dana uses open amount route",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "UDDND10", ProdukNamaSnapshot: "Dana 10.000", Qty: 1, HargaDasar: 11055},
			wantProduct: "DANA", wantQty: 10000,
		},
		{
			name:        "fixed gopay uses open amount route",
			order:       repository.AppOrderRow{ProdukSKUSnapshot: "UDGP10", ProdukNamaSnapshot: "Gopay 10.000", Qty: 1, HargaDasar: 11650},
			wantProduct: "GOPAY", wantQty: 10000,
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
	got := pulsa24JamAppOrderRefID(&repository.AppOrderRow{
		ID:        81,
		InvoiceID: "INV-20260921170504-F95D2B50",
	})
	if got != "PKA29" {
		t.Fatalf("refid = %q, want %q", got, "PKA29")
	}
	if len(got) > 20 {
		t.Fatalf("refid length = %d, want <= 20", len(got))
	}
}

func TestPulsa24JamAppOrderRefIDFallbackFitsH2HRLimit(t *testing.T) {
	got := pulsa24JamAppOrderRefID(&repository.AppOrderRow{InvoiceID: "INV-20260921170504-F95D2B50"})
	if len(got) > 20 {
		t.Fatalf("refid length = %d, want <= 20: %s", len(got), got)
	}
	if got[:3] != "PKA" {
		t.Fatalf("refid = %q, want PKA prefix", got)
	}
}
