package helper

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func resetProviderIPGuardForTest(t *testing.T) {
	t.Helper()
	providerIPMap = nil
	providerIPMapOnce = sync.Once{}
	t.Cleanup(func() {
		providerIPMap = nil
		providerIPMapOnce = sync.Once{}
	})
}

func TestProviderIPGuardLoketBayarUsesCallbackIPWhitelist(t *testing.T) {
	t.Setenv("LOKETBAYAR_CALLBACK_IP", "16.78.237.121")
	resetProviderIPGuardForTest(t)

	hit := false
	handler := ProviderIPGuard("loketbayar", func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusNoContent)
	})

	blockedReq := httptest.NewRequest(http.MethodGet, "/webhook/loketbayar", nil)
	blockedReq.Header.Set("X-Forwarded-For", "192.0.2.10")
	blockedRR := httptest.NewRecorder()
	handler(blockedRR, blockedReq)
	if blockedRR.Code != http.StatusForbidden {
		t.Fatalf("blocked status = %d, want %d", blockedRR.Code, http.StatusForbidden)
	}
	if hit {
		t.Fatal("handler was called for non-whitelisted LoketBayar IP")
	}

	allowedReq := httptest.NewRequest(http.MethodGet, "/webhook/loketbayar", nil)
	allowedReq.Header.Set("X-Forwarded-For", "16.78.237.121")
	allowedRR := httptest.NewRecorder()
	handler(allowedRR, allowedReq)
	if allowedRR.Code != http.StatusNoContent {
		t.Fatalf("allowed status = %d, want %d", allowedRR.Code, http.StatusNoContent)
	}
	if !hit {
		t.Fatal("handler was not called for whitelisted LoketBayar IP")
	}
}
