package chytron

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPayBuildsNewH2HRequestAndRedaction(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotQuery map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = map[string]string{}
		for key, vals := range r.URL.Query() {
			if len(vals) > 0 {
				gotQuery[key] = vals[0]
			}
		}
		_, _ = w.Write([]byte("status=0068&message=Under proses"))
	}))
	defer ts.Close()

	c := New(ts.URL, "idrs", "pin123", "user1", "pass123", 2*time.Second)
	res, err := c.Pay(context.Background(), Request{KodeProduk: "bdanap", Tujuan: "081234567890", Qty: 10000, RefID: "ctref123"})
	if err != nil {
		t.Fatalf("Pay returned error: %v", err)
	}
	if res == nil {
		t.Fatal("Pay returned nil result")
	}
	if gotPath != "/api/h2h" {
		t.Fatalf("path = %q, want /api/h2h", gotPath)
	}
	want := map[string]string{
		"id":         "idrs",
		"pin":        "pin123",
		"user":       "user1",
		"pass":       "pass123",
		"kodeproduk": "BDANAP",
		"tujuan":     "081234567890",
		"idtrx":      "ctref123",
		"counter":    "1",
		"amount":     "10000",
		"jenis":      "1",
	}
	for key, wantVal := range want {
		if gotQuery[key] != wantVal {
			t.Fatalf("query[%s] = %q, want %q", key, gotQuery[key], wantVal)
		}
	}
	if strings.Contains(gotQuery["tujuan"], "@") {
		t.Fatalf("tujuan must not contain amount suffix: %q", gotQuery["tujuan"])
	}
	if res.Pay.Request["pin"] != "***" || res.Pay.Request["pass"] != "***" {
		t.Fatalf("request redaction failed: %#v", res.Pay.Request)
	}
	if res.Pay.Request["amount"] != "10000" || res.Pay.Request["jenis"] != "1" {
		t.Fatalf("request audit fields wrong: %#v", res.Pay.Request)
	}
	if strings.Contains(res.Pay.URL, "pin123") || strings.Contains(res.Pay.URL, "pass123") {
		t.Fatalf("url contains secret: %s", res.Pay.URL)
	}
}
