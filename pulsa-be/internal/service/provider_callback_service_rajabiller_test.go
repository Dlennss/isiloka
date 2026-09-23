package service

import (
	"net/url"
	"testing"
)

func TestParseRajabillerCallbackPayloadUsesTokenAsProviderRef(t *testing.T) {
	raw := `{"trxid":"ref-json-1","rc":"00","status":"Sukses","token":"1234-5678-9012-3456-7890","harga":"83645","saldo_akhir":"14827574"}`

	_, refid, rc, msg, providerRef, price, saldo := parseRajabillerCallbackPayload(raw, url.Values{})

	if refid != "ref-json-1" {
		t.Fatalf("refid = %q", refid)
	}
	if rc != "00" || msg != "Sukses" {
		t.Fatalf("rc/msg = %q/%q", rc, msg)
	}
	if providerRef != "1234-5678-9012-3456-7890" {
		t.Fatalf("providerRef = %q", providerRef)
	}
	if price != 83645 {
		t.Fatalf("price = %d", price)
	}
	if saldo == nil || *saldo != 14827574 {
		t.Fatalf("saldo = %v", saldo)
	}
}
