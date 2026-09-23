package otomax

import (
	"context"
	"net/http"
	"testing"
	"time"
)

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
