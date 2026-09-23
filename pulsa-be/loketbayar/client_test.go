package loketbayar

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTopupUsesOtomaxGETTrxQuery(t *testing.T) {
	const username = "loket-user"
	const password = "secret-pass"

	var gotMethod string
	var gotPath string
	var gotAuth string
	var gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		gotQuery = map[string]string{
			"username": r.URL.Query().Get("username"),
			"password": r.URL.Query().Get("password"),
			"product":  r.URL.Query().Get("product"),
			"dest":     r.URL.Query().Get("dest"),
			"refID":    r.URL.Query().Get("refID"),
			"qty":      r.URL.Query().Get("qty"),
		}
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		_, _ = w.Write([]byte("#5535389-73822 TRX PLN3000PASC ke 530000000001 status SUKSES. SN/REF: reff:1TTY21G506658843727. HARGA: 20000. SALDO: 516074"))
	}))
	defer srv.Close()

	c := New(srv.URL, username, password, 0)
	resp, hs, reqRaw, err := c.Topup(context.Background(), TopupRequest{
		ProductCode: "TRFBANK",
		Dest:        "014901585829270",
		RefID:       "test-ref",
		Nominal:     113500,
	})
	if err != nil {
		t.Fatalf("unexpected topup error: %v", err)
	}
	if hs != http.StatusOK {
		t.Fatalf("unexpected http status: %d", hs)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: got=%q want=%q", gotMethod, http.MethodGet)
	}
	if gotPath != "/trx" {
		t.Fatalf("unexpected path: got=%q want=%q", gotPath, "/trx")
	}
	if gotAuth != "" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}
	if gotBody != "" {
		t.Fatalf("unexpected request body: %q", gotBody)
	}
	if gotQuery["username"] != username || gotQuery["password"] != password ||
		gotQuery["product"] != "TRFBANK" || gotQuery["dest"] != "014901585829270" ||
		gotQuery["refID"] != "test-ref" {
		t.Fatalf("unexpected query: %#v", gotQuery)
	}
	if gotQuery["qty"] != "113500" {
		t.Fatalf("unexpected qty: %q", gotQuery["qty"])
	}
	if reqRaw["password"] != "<redacted>" {
		t.Fatalf("request raw should redact password: %#v", reqRaw)
	}
	if reqRaw["method"] != http.MethodGet || reqRaw["endpoint"] != "/trx" {
		t.Fatalf("unexpected req raw: %#v", reqRaw)
	}
	if reqRaw["qty"] != int64(113500) {
		t.Fatalf("unexpected qty in req raw: %#v", reqRaw)
	}
	if resp.Status != "SUKSES" || resp.ProductCode != "PLN3000PASC" || resp.Dest != "530000000001" ||
		resp.TrxID != "5535389-73822" || resp.Price != 20000 || resp.Saldo != 516074 {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if resp.Reff != "reff:1TTY21G506658843727" || resp.SN != resp.Reff {
		t.Fatalf("unexpected sn/reff: %#v", resp)
	}
}

func TestTopupUsesProductSpecificBaseURL(t *testing.T) {
	defaultHit := false
	defaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultHit = true
		http.Error(w, "default endpoint should not be used", http.StatusTeapot)
	}))
	defer defaultSrv.Close()

	var gotPath string
	var gotProduct string
	v2Srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotProduct = r.URL.Query().Get("product")
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		_, _ = w.Write([]byte("#REF123 TRX TOPUP DANAPLUS ke 082124307365 status SUKSES. SN/REF: DNID TEST. HARGA:10100. SALDO:309321487"))
	}))
	defer v2Srv.Close()

	c := New(defaultSrv.URL, "loket-user", "secret-pass", 0)
	c.SetProductBaseURL("DANAPLUS", v2Srv.URL)
	resp, hs, reqRaw, err := c.Topup(context.Background(), TopupRequest{
		ProductCode: "DANAPLUS",
		Dest:        "082124307365",
		RefID:       "REF123",
		Nominal:     10000,
	})
	if err != nil {
		t.Fatalf("unexpected topup error: %v", err)
	}
	if defaultHit {
		t.Fatal("default endpoint was used for product override")
	}
	if hs != http.StatusOK {
		t.Fatalf("unexpected http status: %d", hs)
	}
	if gotPath != "/trx" || gotProduct != "DANAPLUS" {
		t.Fatalf("unexpected v2 request path/product: path=%q product=%q", gotPath, gotProduct)
	}
	if reqRaw["product_endpoint"] != "DANAPLUS" {
		t.Fatalf("expected product endpoint marker in req raw: %#v", reqRaw)
	}
	if resp.Status != "SUKSES" || resp.Price != 10100 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestAdviceUsesSameOtomaxGETTrxEndpoint(t *testing.T) {
	var gotPath string
	var gotRefID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRefID = r.URL.Query().Get("refID")
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		_, _ = w.Write([]byte("#4316400-97601 TRX PLN2000PASC ke 530000000001 status PENDING. SN/REF: Menunggu jawaban provider. SALDO: 536074"))
	}))
	defer srv.Close()

	c := New(srv.URL, "loket-user", "secret-pass", 0)
	resp, hs, reqRaw, err := c.Advice(context.Background(), TopupRequest{
		ProductCode: "TRFBANK",
		Dest:        "014901585829270",
		RefID:       "test-ref",
		Nominal:     113500,
	})
	if err != nil {
		t.Fatalf("unexpected advice error: %v", err)
	}
	if hs != http.StatusOK {
		t.Fatalf("unexpected http status: %d", hs)
	}
	if gotPath != "/trx" {
		t.Fatalf("unexpected path: got=%q want=%q", gotPath, "/trx")
	}
	if gotRefID != "test-ref" {
		t.Fatalf("unexpected refID: got=%q", gotRefID)
	}
	if reqRaw["advice"] != true {
		t.Fatalf("expected advice marker in req raw: %#v", reqRaw)
	}
	if resp.Status != "PENDING" || resp.Saldo != 536074 {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if resp.Raw["advice"] != true || resp.Raw["endpoint"] != "/trx" {
		t.Fatalf("unexpected raw markers: %#v", resp.Raw)
	}
}

func TestCancelDepositTicketUsesLoketBayarTicketCancelQuery(t *testing.T) {
	const username = "loket-user"

	var gotMethod string
	var gotPath string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = map[string]string{
			"username":   r.URL.Query().Get("username"),
			"kode_tiket": r.URL.Query().Get("kode_tiket"),
		}
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		_, _ = w.Write([]byte("Tiket 10000000 berhasil dibatalkan"))
	}))
	defer srv.Close()

	c := New(srv.URL, username, "secret-pass", 0)
	resp, hs, reqRaw, err := c.CancelDepositTicket(context.Background(), DepositTicketCancelRequest{
		TicketID: "10000000",
	})
	if err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}
	if hs != http.StatusOK {
		t.Fatalf("unexpected http status: %d", hs)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: got=%q want=%q", gotMethod, http.MethodGet)
	}
	if gotPath != "/tiket/cancel" {
		t.Fatalf("unexpected path: got=%q want=%q", gotPath, "/tiket/cancel")
	}
	if gotQuery["username"] != username || gotQuery["kode_tiket"] != "10000000" {
		t.Fatalf("unexpected query: %#v", gotQuery)
	}
	if reqRaw["method"] != http.MethodGet || reqRaw["endpoint"] != "/tiket/cancel" {
		t.Fatalf("unexpected req raw: %#v", reqRaw)
	}
	if reqRaw["kode_tiket"] != "10000000" {
		t.Fatalf("unexpected ticket in req raw: %#v", reqRaw)
	}
	if resp.Status != "SUKSES" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestCancelDepositTicketGagalNotPendingBeatsPendingText(t *testing.T) {
	resp := parseDepositTicketCancelResponse("GAGAL.Tiket tidak ditemukan atau sudah tidak dalam status pending")
	if resp.Status != "GAGAL" {
		t.Fatalf("status = %q, want GAGAL", resp.Status)
	}
}

func TestParseOtomaxFailedResponse(t *testing.T) {
	resp := parseOtomaxResponse("#100-200 TRX ABC ke 08123456789 status GAGAL. SN/REF: nomor tujuan salah. SALDO: 123456")
	if resp.Status != "GAGAL" {
		t.Fatalf("status = %q, want GAGAL", resp.Status)
	}
	if resp.ProductCode != "ABC" || resp.Dest != "08123456789" {
		t.Fatalf("unexpected product/dest: %#v", resp)
	}
	if resp.Reff != "nomor tujuan salah" {
		t.Fatalf("unexpected reff: %q", resp.Reff)
	}
	if resp.Saldo != 123456 {
		t.Fatalf("saldo = %d, want 123456", resp.Saldo)
	}
}

func TestParseOtomaxSNLabelResponse(t *testing.T) {
	resp := parseOtomaxResponse("#REF123 TRX TOPUP DANAPLUS ke 082124307365 status SUKSES.SN:DNID 082124307365/082124307365/10000/2026062710121481030100166432841869275.HARGA:10100.SALDO:309311387")
	if resp.Status != "SUKSES" {
		t.Fatalf("status = %q, want SUKSES", resp.Status)
	}
	if resp.ProductCode != "DANAPLUS" || resp.Dest != "082124307365" {
		t.Fatalf("unexpected product/dest: %#v", resp)
	}
	wantSN := "DNID 082124307365/082124307365/10000/2026062710121481030100166432841869275"
	if resp.Reff != wantSN || resp.SN != wantSN {
		t.Fatalf("unexpected sn/reff: reff=%q sn=%q", resp.Reff, resp.SN)
	}
	if resp.Price != 10100 || resp.Saldo != 309311387 {
		t.Fatalf("unexpected price/saldo: price=%d saldo=%d", resp.Price, resp.Saldo)
	}
}
