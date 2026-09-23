package sagaramobile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestParseRefIDFromMsgSupportsAlphaNumeric(t *testing.T) {
	got := ParseRefIDFromMsg("status=1&message=KOMPLAIN MAX H+7 R# SGRTESTQTY1774960752 DANARP.082124307365 akan diproses")
	if got != "SGRTESTQTY1774960752" {
		t.Fatalf("unexpected refid: %q", got)
	}
}

func TestParsePriceFromPendingBalanceDelta(t *testing.T) {
	got := ParsePriceFromMsg("status=1&message=KOMPLAIN MAX H+7 R# SGRTESTQTY1774960752 DANARP.082124307365 akan diproses @19:39. Saldo 5,000,924 - 10,030 = 4,990,894")
	if got != 10030 {
		t.Fatalf("unexpected price: %d", got)
	}
}

func TestParseLastBalanceFromSuccessMessagePrefersSaldoSection(t *testing.T) {
	got, ok := ParseLastBalanceFromMsg("status=20&message=DANARP.082124307365 SUKSES. SN: DNID 082124307365/21000/2026033110121481030100166446394454002 saldo 4,990,894 - 21,030 = 4,969,864 @31/03 19:47:28 R#SGRTEST210001774961228, harga = 21,030")
	if !ok {
		t.Fatalf("expected balance to parse")
	}
	if got != 4969864 {
		t.Fatalf("unexpected balance: %d", got)
	}
}

func TestTrxUsesNoSignAuthWhenPinAndPasswordPresent(t *testing.T) {
	var got url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		_, _ = w.Write([]byte("status=1&message=R# TESTREF akan diproses"))
	}))
	defer ts.Close()

	c := New(ts.URL, "SGR1849", "", "", "050505", "Boju0505@", 0)
	_, hs, _, err := c.Trx(context.Background(), "DANARP", "082124307365", 10000, "TESTREF")
	if err != nil {
		t.Fatalf("trx error: %v", err)
	}
	if hs != 200 {
		t.Fatalf("unexpected status: %d", hs)
	}
	if got.Get("memberID") != "SGR1849" || got.Get("pin") != "050505" || got.Get("password") != "Boju0505@" {
		t.Fatalf("unexpected auth query: %v", got)
	}
	if got.Get("qty") != "10000" {
		t.Fatalf("unexpected qty: %q", got.Get("qty"))
	}
	if got.Get("sign") != "" {
		t.Fatalf("sign should be empty in no-sign mode: %q", got.Get("sign"))
	}
}
