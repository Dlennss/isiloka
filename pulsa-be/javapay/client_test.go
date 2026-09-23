package javapay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStatusUsesLatestPayload(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trx" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":true,"data":{"rc":"0","message":"pending"}}`))
	}))
	defer srv.Close()

	client := New(srv.URL, "MID", "APIKEY", "123456", 5*time.Second)
	_, _, req, err := client.Status(context.Background(), "REF-001")
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if req["commands"] != "STATUS" {
		t.Fatalf("unexpected commands: %#v", req["commands"])
	}
	if _, ok := req["product"]; ok {
		t.Fatalf("product should be omitted on status request")
	}
	if _, ok := req["dest"]; ok {
		t.Fatalf("dest should be omitted on status request")
	}
	if got["commands"] != "STATUS" {
		t.Fatalf("unexpected upstream commands: %#v", got["commands"])
	}
	if _, ok := got["product"]; ok {
		t.Fatalf("upstream product should be omitted on status request")
	}
}

func TestTrxDelegatesStatusAlias(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":true,"data":{"rc":"0","message":"pending"}}`))
	}))
	defer srv.Close()

	client := New(srv.URL, "MID", "APIKEY", "123456", 5*time.Second)
	_, _, _, err := client.Trx(context.Background(), "STATUS-PAY", "DANA", "08123", 1000, "REF-002")
	if err != nil {
		t.Fatalf("trx error: %v", err)
	}
	if got["commands"] != "STATUS" {
		t.Fatalf("expected STATUS command, got %#v", got["commands"])
	}
}

func TestVerifyCallbackSignature(t *testing.T) {
	client := New("https://api.javapay.id/api", "MID", "secret", "123456", 5*time.Second)
	body := []byte(`{"status":true}`)

	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	if !client.VerifyCallbackSignature(body, signature) {
		t.Fatal("expected signature to verify")
	}
	if client.VerifyCallbackSignature(body, "bad-signature") {
		t.Fatal("expected invalid signature to fail")
	}
}

func TestHTTPClientForContextUsesSoonerDeadline(t *testing.T) {
	c := &Client{
		httpc: &http.Client{Timeout: 20 * time.Second},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got := c.httpClientForContext(ctx)
	if got == c.httpc {
		t.Fatalf("expected a cloned client when context deadline is sooner")
	}
	if got.Timeout > 6*time.Second || got.Timeout < 4*time.Second {
		t.Fatalf("unexpected timeout: got=%s", got.Timeout)
	}
}

func TestHTTPClientForContextKeepsBaseClientWhenDeadlineIsLonger(t *testing.T) {
	base := &http.Client{Timeout: 5 * time.Second}
	c := &Client{httpc: base}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	got := c.httpClientForContext(ctx)
	if got != base {
		t.Fatalf("expected base client when context deadline is not stricter")
	}
}
