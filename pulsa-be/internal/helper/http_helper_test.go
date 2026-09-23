package helper

import (
	"net/http/httptest"
	"testing"
)

func TestQueryDateUsesJakartaTimezone(t *testing.T) {
	r := httptest.NewRequest("GET", "/?from=2026-04-01", nil)

	got, ok := QueryDate(r, "from")
	if !ok {
		t.Fatalf("expected date to parse")
	}
	if got.Location().String() != "Asia/Jakarta" {
		t.Fatalf("unexpected location: got=%q want=%q", got.Location().String(), "Asia/Jakarta")
	}
	if got.Format("2006-01-02 15:04:05 -0700") != "2026-04-01 00:00:00 +0700" {
		t.Fatalf("unexpected parsed date: got=%q", got.Format("2006-01-02 15:04:05 -0700"))
	}
}
