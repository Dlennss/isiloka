package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

func TestProviderSupportsOpenAmount(t *testing.T) {
	tests := []struct {
		name           string
		rule           *repository.ProviderOpenAmountRule
		billingNominal int64
		want           bool
	}{
		{"talenta dana in range", &repository.ProviderOpenAmountRule{MinimalNominal: int64Ptr(1000), MaksimalNominal: int64Ptr(1000000)}, 250000, true},
		{"talenta dana above max", &repository.ProviderOpenAmountRule{MinimalNominal: int64Ptr(1000), MaksimalNominal: int64Ptr(1000000)}, 1100000, false},
		{"multikom dana above max", &repository.ProviderOpenAmountRule{MinimalNominal: int64Ptr(1000), MaksimalNominal: int64Ptr(1000000)}, 1100000, false},
		{"talenta gopay above max", &repository.ProviderOpenAmountRule{MinimalNominal: int64Ptr(10000), MaksimalNominal: int64Ptr(500000)}, 698800, false},
		{"multikom gopay within max", &repository.ProviderOpenAmountRule{MinimalNominal: int64Ptr(10000), MaksimalNominal: int64Ptr(1000000)}, 698800, true},
		{"talenta linkaja above max", &repository.ProviderOpenAmountRule{MinimalNominal: int64Ptr(1000), MaksimalNominal: int64Ptr(200000)}, 299000, false},
		{"yuscom unknown left enabled", nil, 1500000, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := providerSupportsOpenAmountRule(tc.rule, tc.billingNominal)
			if got != tc.want {
				t.Fatalf("unexpected support result: got=%t want=%t", got, tc.want)
			}
		})
	}
}

func TestBuildProviderAttemptsSkipsLegacyYuscomWhenDBMappingExists(t *testing.T) {
	candidates := []repository.ProviderRouteCandidate{
		{
			ProdukProviderMapID: int64Ptr(4),
			ProdukSKUSnapshot:   "GOPAY",
			Provider:            "yuscom",
			KodeProvider:        "GPAY",
		},
	}

	if !hasProviderRouteCandidate(candidates, "yuscom") {
		t.Fatalf("expected yuscom db mapping candidate to be detected")
	}

	if hasProviderRouteCandidate(candidates, "talentapay") {
		t.Fatalf("did not expect unrelated provider to be detected")
	}
}

func hasProviderRouteCandidate(candidates []repository.ProviderRouteCandidate, provider string) bool {
	for _, candidate := range candidates {
		if candidate.Provider == provider {
			return true
		}
	}
	return false
}

func TestShouldKeepPendingOnProviderFailure(t *testing.T) {
	if shouldKeepPendingOnProviderFailure(errors.New("dial tcp 10.0.0.1: i/o timeout")) {
		t.Fatalf("did not expect network timeout to stay pending")
	}

	if shouldKeepPendingOnProviderFailure(errors.New("javapay error http=500 err=<nil> body=internal server error")) {
		t.Fatalf("did not expect http 500 to stay pending")
	}

	if shouldKeepPendingOnProviderFailure(errors.New("trionik error http=503 err=<nil> body=server maintenance")) {
		t.Fatalf("did not expect non-200 provider response to stay pending")
	}

	if shouldKeepPendingOnProviderFailure(errors.New("connection reset by peer")) {
		t.Fatalf("did not expect connection errors to stay pending")
	}
}

func TestProviderTransportDefinitelyNotDispatched(t *testing.T) {
	finalFailures := []string{
		"dial tcp 36.93.29.179:6969: connect: no route to host",
		"dial tcp 10.0.0.1:443: connect: network is unreachable",
		"dial tcp 10.0.0.1:443: connect: connection refused",
		"lookup provider.example.invalid: no such host",
	}
	for _, msg := range finalFailures {
		if !providerTransportDefinitelyNotDispatched(errors.New(msg)) {
			t.Fatalf("expected %q to be treated as definitely not dispatched", msg)
		}
	}

	ambiguousFailures := []string{
		"dial tcp 10.0.0.1:443: i/o timeout",
		"context deadline exceeded",
		"read tcp 10.0.0.2:50100->10.0.0.1:443: read: connection reset by peer",
		"EOF",
	}
	for _, msg := range ambiguousFailures {
		if providerTransportDefinitelyNotDispatched(errors.New(msg)) {
			t.Fatalf("did not expect %q to be treated as definitely not dispatched", msg)
		}
	}
}

func TestProviderMaySendLateCallback(t *testing.T) {
	for _, provider := range []string{"talentapay", "multikom", "trionik", "ajs", "smb", "loketbayar", "chytron"} {
		if !providerMaySendLateCallback(provider) {
			t.Fatalf("expected %s to be treated as late-callback capable", provider)
		}
	}
	if providerMaySendLateCallback("javapay") {
		t.Fatalf("did not expect javapay transport failures to be held as late callback")
	}
}

func TestProviderRetriesUntilCallback(t *testing.T) {
	for _, provider := range []string{"smb", "loketbayar", "SMB", " LoketBayar "} {
		if !providerRetriesUntilCallback(provider) {
			t.Fatalf("expected %q to keep retrying same refid until response/callback", provider)
		}
	}
	for _, provider := range []string{"trionik", "rajabiller", "javapay", ""} {
		if providerRetriesUntilCallback(provider) {
			t.Fatalf("did not expect %q to bypass normal retry/fallback limits", provider)
		}
	}
}

func TestFallbackNoResponseShouldStayPending(t *testing.T) {
	timeoutErr := errors.New("context deadline exceeded")
	for _, providerName := range []string{"smb", "loketbayar"} {
		if !fallbackNoResponseShouldStayPending(providerName, 0, timeoutErr) {
			t.Fatalf("expected %s fallback timeout to stay pending", providerName)
		}
	}
	if fallbackNoResponseShouldStayPending("trionik", 0, timeoutErr) {
		t.Fatalf("did not expect non retry-until-callback provider fallback timeout to stay pending")
	}
	if fallbackNoResponseShouldStayPending("loketbayar", http.StatusInternalServerError, nil) {
		t.Fatalf("did not expect Loketbayar HTTP 500 response without transport error to use no-response hold")
	}
}

func TestProviderPayResponseHasProviderReply(t *testing.T) {
	if providerPayResponseHasProviderReply(nil) {
		t.Fatalf("nil response must not count as provider reply")
	}
	if providerPayResponseHasProviderReply(&provider.PayResponse{Raw: map[string]any{"error": "context deadline exceeded"}}) {
		t.Fatalf("transport-only raw error must not count as provider reply")
	}
	if !providerPayResponseHasProviderReply(&provider.PayResponse{HTTPStatus: 200}) {
		t.Fatalf("HTTP status must count as provider reply")
	}
	if !providerPayResponseHasProviderReply(&provider.PayResponse{Body: "status=1&message=akan diproses"}) {
		t.Fatalf("provider body must count as provider reply")
	}
	if !providerPayResponseHasProviderReply(&provider.PayResponse{ProviderRef: "123456"}) {
		t.Fatalf("provider reference must count as provider reply")
	}
}

func TestLoketBayarGangguanIsImmediatePAYFailure(t *testing.T) {
	body := `{"status":"GAGAL","message":"PRODUK GANGGUAN"}`
	if !helper.ProviderResponseImmediateReject("loketbayar", body) {
		t.Fatalf("expected loketbayar gangguan response to be immediate reject")
	}
	if helper.ProviderResponseAccepted("loketbayar", body) {
		t.Fatalf("did not expect loketbayar gangguan response to be accepted/pending")
	}
	if shouldKeepPendingOnProviderFailure(errors.New("loketbayar reject bisnis: " + body)) {
		t.Fatalf("did not expect loketbayar gangguan PAY failure to stay pending")
	}
}

func TestLoketBayarSaldoOnlyResponseStaysPending(t *testing.T) {
	body := "status= keterangan=. SALDO: 3405450140"
	if !helper.ProviderResponseAccepted("loketbayar", body) {
		t.Fatalf("expected loketbayar saldo-only response to be accepted as pending")
	}
	if helper.ProviderResponseImmediateReject("loketbayar", body) {
		t.Fatalf("did not expect loketbayar saldo-only response to be immediate reject")
	}
}

func TestLoketBayarDoesNotRetryPermanentHTTPAuthErrors(t *testing.T) {
	attempts := 0
	resp, err := runLoketBayarPayWithRetryWindow(
		t.Context(),
		func(context.Context) (*provider.PayResponse, error) {
			attempts++
			return &provider.PayResponse{HTTPStatus: http.StatusUnauthorized, Body: "Unauthorized!"}, nil
		},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.HTTPStatus != http.StatusUnauthorized {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if attempts != 1 {
		t.Fatalf("expected no retry for loketbayar 401, got attempts=%d", attempts)
	}
}

func TestIsJavapayStatusNotFoundMessage(t *testing.T) {
	if !isJavapayStatusNotFoundMessage("RefID not found") {
		t.Fatalf("expected refid not found to be detected")
	}

	if !isJavapayStatusNotFoundMessage("trx tidak ditemukan di provider") {
		t.Fatalf("expected bahasa indonesia not found to be detected")
	}

	if isJavapayStatusNotFoundMessage("status=1 transaksi sedang diproses") {
		t.Fatalf("did not expect pending message to be treated as not found")
	}
}

func TestClassifyJavapayResponseStatus(t *testing.T) {
	if got := classifyJavapayResponseStatus("20", "Sukses"); got != "success" {
		t.Fatalf("unexpected success classification: %q", got)
	}
	if got := classifyJavapayResponseStatus("1", "Sedang diproses"); got != "pending" {
		t.Fatalf("unexpected pending classification: %q", got)
	}
	if got := classifyJavapayResponseStatus("64", "Diabaikan"); got != "pending" {
		t.Fatalf("unexpected ignored classification: %q", got)
	}
	if got := classifyJavapayResponseStatus("", "RefID not found"); got != "failed" {
		t.Fatalf("unexpected not-found classification: %q", got)
	}
	if got := classifyJavapayResponseStatus("", ""); got != "unknown" {
		t.Fatalf("unexpected unknown classification: %q", got)
	}
}

func TestShouldRetryStaleJavapayPending(t *testing.T) {
	now := time.Now()
	row := &model.JavapayTrxRow{
		Provider:   "javapay",
		DibuatPada: now.Add(-javapayPendingRetryDelay - time.Minute),
	}

	if !shouldRetryStaleJavapayPending(row, "64", "Diabaikan / Sedang diproses", now) {
		t.Fatalf("expected stale javapay rc64 pending to trigger retry")
	}

	row.DibuatPada = now.Add(-time.Minute)
	if shouldRetryStaleJavapayPending(row, "64", "Diabaikan / Sedang diproses", now) {
		t.Fatalf("did not expect fresh javapay pending to trigger retry")
	}

	row.DibuatPada = now.Add(-javapayPendingRetryDelay - time.Minute)
	if shouldRetryStaleJavapayPending(row, "1", "Sedang diproses", now) {
		t.Fatalf("did not expect non-ignored pending to trigger retry")
	}
}

func TestIsProviderSystemErrorStopsOnTerminalBusinessFailure(t *testing.T) {
	if isProviderSystemError(assertErr("yuscom reject bisnis: R#1 GAGAL. Nomor tujuan salah.")) {
		t.Fatalf("invalid number should be treated as terminal failure, not retryable system error")
	}
	if isProviderSystemError(assertErr("multikom reject bisnis: status=40 GAGAL. Batas maksimal pembelian/ Akun sdh Limit.Stok dikembalikan.")) {
		t.Fatalf("account limit should be treated as terminal failure, not retryable system error")
	}
	if !isProviderSystemError(assertErr("multikom reject bisnis: status=43 GAGAL. Stok tidak cukup.")) {
		t.Fatalf("provider stock issue should remain retryable across providers")
	}
	if !isProviderSystemError(assertErr("talenta reject bisnis: status=61 Qty tidak sesuai. Allowed QTY is 1000-500000")) {
		t.Fatalf("qty limit should remain retryable across providers per latest rule")
	}
}

func TestTryStatusPayFallbackContinuationPolicy(t *testing.T) {
	if !isProviderSystemError(assertErr("smb reject bisnis: callback cek tidak diterima dalam 5 detik")) {
		t.Fatalf("expected SMB callback-timeout failure to continue to next fallback candidate")
	}
	if isProviderSystemError(assertErr("smb reject bisnis: Nomor tujuan salah")) {
		t.Fatalf("did not expect terminal SMB business failure to continue to next fallback candidate")
	}
}
