package service

import (
	"net/url"
	"testing"
)

func TestParseChytronCallbackRealFields(t *testing.T) {
	t.Parallel()

	raw := "serverid=39517374&clientid=ctref123&statuscode=1&kp=BDANAP&msisdn=082124307365&sn=DNID-TEST&msg=REFF%23ctref123+BDANAP+ke+082124307365+BERHASIL%2C+SN%3A+DNID-TEST+Saldo%3A+10.000.893+-+10.070+%3D+10.000.892"
	q, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}

	payload, refid, rc, msg, providerRef, price, saldo := parseChytronCallbackPayload(raw, q)
	if refid != "ctref123" {
		t.Fatalf("refid = %q", refid)
	}
	if rc != "1" {
		t.Fatalf("rc = %q", rc)
	}
	if providerRef != "DNID-TEST" {
		t.Fatalf("providerRef = %q", providerRef)
	}
	if price != 10070 {
		t.Fatalf("price = %d", price)
	}
	if saldo != nil {
		t.Fatalf("saldo = %v, want nil without saldo field", *saldo)
	}
	if payload["tujuan"] != "082124307365" {
		t.Fatalf("payload tujuan = %#v", payload["tujuan"])
	}
	if payload["kode_produk"] != "BDANAP" {
		t.Fatalf("payload kode_produk = %#v", payload["kode_produk"])
	}
	if got := resolveChytronFinalStatus(rc, msg); got != "success" {
		t.Fatalf("final status = %q", got)
	}
}
