package smb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseMappedCode(t *testing.T) {
	mode, code, err := ParseMappedCode("ELDN")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if mode != ModeDirect || code != "ELDN" {
		t.Fatalf("unexpected parse result: mode=%s code=%s", mode, code)
	}
}

func TestParseMappedCodeLegacyPrefixStillWorks(t *testing.T) {
	mode, code, err := ParseMappedCode("wallet_ppob:dana")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if mode != ModeWalletPPOB || code != "DANA" {
		t.Fatalf("unexpected parse result: mode=%s code=%s", mode, code)
	}
}

func TestLooksLikeBodyClassification(t *testing.T) {
	if !LooksLikeSuccess("status=20 transaksi sukses harga=1000") {
		t.Fatal("expected success")
	}
	if !LooksLikePending("status=2 sedang diproses") {
		t.Fatal("expected pending")
	}
	if !LooksLikeImmediateReject("status=55 gagal karena timeout") {
		t.Fatal("expected reject")
	}
}

func TestParseRefIDStripsCheckAndPayPrefix(t *testing.T) {
	got := ParseRefID("idtrx=BYR177123 status=20 sukses")
	if got != "177123" {
		t.Fatalf("unexpected refid: %s", got)
	}
}

func TestParsePriceFromJSONMessage(t *testing.T) {
	body := `{"rc":"1","status":1,"success":true,"msg":"TRANSAKSI BERHASIL HARGA:1.100 SALDO:45.015.905"}`
	got := ParsePrice(body)
	if got != 1100 {
		t.Fatalf("unexpected price: %d", got)
	}
}

func TestLooksLikeSuccessFromJSONSuccessFlag(t *testing.T) {
	body := `{"rc":"1","status":1,"success":true,"msg":"TRANSAKSI BERHASIL"}`
	if !LooksLikeSuccess(body) {
		t.Fatal("expected success from json success flag")
	}
	if !LooksLikeAccepted(body) {
		t.Fatal("expected accepted body")
	}
}

func TestLooksLikePendingFromJSONPendingFlag(t *testing.T) {
	body := `{"rc":"68","status":"68","success":true,"msg":"PENDING DALAM PROSES"}`
	if !LooksLikePending(body) {
		t.Fatal("expected pending from json pending status")
	}
	if LooksLikeSuccess(body) {
		t.Fatal("did not expect pending body to be treated as success")
	}
	if !LooksLikeAccepted(body) {
		t.Fatal("expected pending body to remain accepted")
	}
}

func TestLooksLikePendingFromJSONZeroPaddedStatus(t *testing.T) {
	body := `{"success":true,"produk":"ELDN","tujuan":"082235804266","reffid":"1775158471455740","rc":"0068","harga":299065,"msg":"Trx ELDN 082235804266 Under proses...","saldo":"106.204.167"}`
	if got := ExtractStatusCode(body); got != "68" {
		t.Fatalf("unexpected normalized rc: %q", got)
	}
	if !LooksLikePending(body) {
		t.Fatal("expected zero-padded pending status to stay pending")
	}
	if LooksLikeImmediateReject(body) {
		t.Fatal("did not expect zero-padded pending status to be treated as reject")
	}
}

func TestLooksLikePendingFromJSONQueuedStatus(t *testing.T) {
	body := `{"success":true,"produk":"DANA","tujuan":"50000@085260432115","reffid":"BYR1775212328264805","rc":"0027","msg":"[0027] transaksi sedang dalam antrian"}`
	if got := ExtractStatusCode(body); got != "27" {
		t.Fatalf("unexpected normalized rc: %q", got)
	}
	if !LooksLikePending(body) {
		t.Fatal("expected queued callback to stay pending")
	}
	if LooksLikeImmediateReject(body) {
		t.Fatal("did not expect queued callback to be treated as reject")
	}
}

func TestLooksLikeImmediateRejectFromJSONFailure(t *testing.T) {
	body := `{"rc":"14","status":14,"success":false,"msg":"invalid Credential"}`
	if !LooksLikeImmediateReject(body) {
		t.Fatal("expected immediate reject")
	}
}

func TestLooksLikeImmediateRejectFromBusinessMessage(t *testing.T) {
	tests := []string{
		`{"success":false,"rc":"0061","msg":"[0061] cek balance dan transaksi ke CS/admin"}`,
		`{"success":true,"rc":"0061","msg":"[0061] cek balance dan transaksi ke CS/admin"}`,
		`{"success":false,"produk":"GOPAY","tujuan":"150000@089533010870","reffid":"CEK1775280515060398","rc":"2","status":2,"sn":"","msg":"REFF#CEK1775280515060398 GOPAY 150000@089533010870 GAGAL, KET: AN ERROR OCCURRED WHEN DOING ACCOUNT VALIDATION, SISA SALDO: 18.505.568  @WAKTU:04/04/2026 12:31:27."}`,
		`SMB CHECK saldo provider kurang: saldo=22103 jumlah=76000`,
	}

	for _, body := range tests {
		if !LooksLikeImmediateReject(body) {
			t.Fatalf("expected immediate reject for body=%s", body)
		}
	}
}

func TestLooksLikePendingDoesNotMaskExplicitFailureSignal(t *testing.T) {
	body := `{"success":false,"produk":"GOPAY","tujuan":"150000@089533010870","reffid":"CEK1775280515060398","rc":"2","status":2,"sn":"","msg":"REFF#CEK1775280515060398 GOPAY 150000@089533010870 GAGAL, KET: AN ERROR OCCURRED WHEN DOING ACCOUNT VALIDATION, SISA SALDO: 18.505.568  @WAKTU:04/04/2026 12:31:27."}`
	if LooksLikePending(body) {
		t.Fatal("expected explicit SMB failure signal to avoid pending classification")
	}
}

func TestDispatchUsesCheckBalanceWhenPayBodyHasNoBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idtrx := r.URL.Query().Get("idtrx")
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(idtrx, "CEK"):
			_, _ = w.Write([]byte(`{"success":true,"rc":"1","msg":"INQSUKSES HARGA:0,SISASALDO:22.103 - 0 = 22.103"}`))
		case strings.HasPrefix(idtrx, "BYR"):
			_, _ = w.Write([]byte(`{"success":false,"rc":"0061","msg":"[0061] cek balance dan transaksi ke CS/admin"}`))
		default:
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, "id", "pin", "user", "pass", 2*time.Second)
	out, err := c.Dispatch(context.Background(), ModeWalletPPOB, Request{
		KodeProduk: "DANA",
		Tujuan:     "08123",
		Qty:        76000,
		RefID:      "R1",
	}, "PAY")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.LastBalance == nil || *out.LastBalance != 22103 {
		t.Fatalf("unexpected last balance: %+v", out.LastBalance)
	}
}

func TestParseLastBalanceUsesFinalResultAfterEquals(t *testing.T) {
	body := `{"success":true,"rc":"1","msg":"REFF#SMBT2002 ELDN.6282124307365 BERHASIL. HARGA: 2.065, SISA SALDO: 37.030.443 - 2.065 = 37.028.378"}`
	got, ok := ParseLastBalance(body)
	if !ok {
		t.Fatal("expected last balance to be parsed")
	}
	if got != 37028378 {
		t.Fatalf("unexpected last balance: %d", got)
	}
}

func TestParseMappedCodeTargetWithModeOverride(t *testing.T) {
	mode, code, bankPrefix, err := ParseMappedCodeTargetWithMode("direct", "DANAOPEN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ModeDirect || code != "DANAOPEN" || bankPrefix != "" {
		t.Fatalf("unexpected override: mode=%s code=%s bankPrefix=%s", mode, code, bankPrefix)
	}
}

func TestParseMappedCodeBIFastOpenBankPrefix(t *testing.T) {
	mode, code, bankPrefix, err := ParseMappedCodeTarget("BIFASTOPEN:008")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ModeDirect || code != "BIFASTOPEN" || bankPrefix != "008" {
		t.Fatalf("unexpected mapping: mode=%s code=%s bankPrefix=%s", mode, code, bankPrefix)
	}
}

func TestParseMappedCodeBIFastOpen2BankPrefix(t *testing.T) {
	mode, code, bankPrefix, err := ParseMappedCodeTarget("BIFASTOPEN2:008")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ModeDirect || code != "BIFASTOPEN2" || bankPrefix != "008" {
		t.Fatalf("unexpected mapping: mode=%s code=%s bankPrefix=%s", mode, code, bankPrefix)
	}
}
