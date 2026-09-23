package repository

import "testing"

func TestProviderUpdateResultStatusHTTPZeroStaysPending(t *testing.T) {
	httpStatus := 0

	got := providerUpdateResultStatus("trionik", &httpStatus, nil, nil)
	if got != "pending" {
		t.Fatalf("status = %q, want pending", got)
	}
}

func TestProviderUpdateResultStatusPositiveNon200Fails(t *testing.T) {
	httpStatus := 502

	got := providerUpdateResultStatus("trionik", &httpStatus, nil, nil)
	if got != "failed" {
		t.Fatalf("status = %q, want failed", got)
	}
}

func TestKeepLoketBayarHTTP200OnFailedCheck(t *testing.T) {
	transportErr := "loketbayar error http=0 setelah 12x retry"
	if !keepLoketBayarHTTP200OnFailedCheck("loketbayar", 200, "failed", nil, &transportErr) {
		t.Fatal("loketbayar row with existing HTTP 200 must not be downgraded by a transport/non-200 check")
	}
	if keepLoketBayarHTTP200OnFailedCheck("loketbayar", 0, "failed", nil, &transportErr) {
		t.Fatal("loketbayar row without existing HTTP 200 may still be failed by a failed check")
	}
	if keepLoketBayarHTTP200OnFailedCheck("smb", 200, "failed", nil, &transportErr) {
		t.Fatal("rule must only apply to loketbayar")
	}
	if keepLoketBayarHTTP200OnFailedCheck("loketbayar", 200, "success", nil, &transportErr) {
		t.Fatal("rule must only apply to failed check results")
	}
	rc := "GAGAL"
	msg := "#ref TRX TOPUP TRFBANK ke 4517360962002 status GAGAL. TRANSAKSI GAGAL DI PROVIDER."
	if keepLoketBayarHTTP200OnFailedCheck("loketbayar", 200, "failed", &rc, &msg) {
		t.Fatal("loketbayar explicit status GAGAL must become failed")
	}
}
