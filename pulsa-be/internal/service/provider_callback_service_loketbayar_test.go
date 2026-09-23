package service

import (
	"net/url"
	"testing"
)

func TestParseLoketBayarCallbackPayloadNestedAPI(t *testing.T) {
	raw := `{"saldo_awal":76342373,"saldo_akhir":76241574,"api":{"status":"99","keterangan":"Transaksi Gagal","kodeProduk":"TRFBANK","nomor_rekening":"535901820607964","nama_penerima":"MUHAMMAD ATOUL MAULA","bank":"PT. BANK SEABANK INDONESIA","orderId":"smpay27f3726b13","nominal":100000,"admin":0,"total":100799,"reff":null}}`

	payload, refid, rc, msg, providerRef, price, saldo := parseLoketBayarCallbackPayload(raw, url.Values{})

	if refid != "smpay27f3726b13" {
		t.Fatalf("refid = %q, want smpay27f3726b13", refid)
	}
	if rc != "99" {
		t.Fatalf("rc = %q, want 99", rc)
	}
	if msg != "Transaksi Gagal" {
		t.Fatalf("msg = %q, want Transaksi Gagal", msg)
	}
	if providerRef != "" {
		t.Fatalf("providerRef = %q, want empty", providerRef)
	}
	if price != 100799 {
		t.Fatalf("price = %d, want 100799", price)
	}
	if saldo == nil || *saldo != 76241574 {
		if saldo == nil {
			t.Fatalf("saldo = nil, want 76241574")
		}
		t.Fatalf("saldo = %d, want 76241574", *saldo)
	}
	if got := loketPayloadString(payload, "kodeProduk"); got != "TRFBANK" {
		t.Fatalf("kodeProduk = %q, want TRFBANK", got)
	}
	if got := loketPayloadString(payload, "nomor_rekening"); got != "535901820607964" {
		t.Fatalf("nomor_rekening = %q, want 535901820607964", got)
	}
}

func TestParseLoketBayarCallbackPayloadOtomaxTextQuery(t *testing.T) {
	text := "#juni24053328aa TRX TOPUP TRFBANK ke 0148545383213 status SUKSES. SN/REF: nama:SUGENG RIYADI/nominal:2200000/admin:0/total:2200700/reff:20260605BRINIDJA010O9938754680/bank:BANK BCA/pengirim:TRI USAHA BERKAT/order:260605172937878186ED. HARGA: 2200700. SALDO: 92525248"
	raw := "q=" + url.QueryEscape(text)

	payload, refid, rc, msg, providerRef, price, saldo := parseLoketBayarCallbackPayload(raw, url.Values{})

	if refid != "juni24053328aa" {
		t.Fatalf("refid = %q, want juni24053328aa", refid)
	}
	if rc != "SUKSES" {
		t.Fatalf("rc = %q, want SUKSES", rc)
	}
	if msg != text {
		t.Fatalf("msg = %q, want raw callback text", msg)
	}
	if providerRef != "20260605BRINIDJA010O9938754680" {
		t.Fatalf("providerRef = %q, want 20260605BRINIDJA010O9938754680", providerRef)
	}
	if price != 2200700 {
		t.Fatalf("price = %d, want 2200700", price)
	}
	if saldo == nil || *saldo != 92525248 {
		if saldo == nil {
			t.Fatalf("saldo = nil, want 92525248")
		}
		t.Fatalf("saldo = %d, want 92525248", *saldo)
	}
	if got := loketPayloadString(payload, "kodeProduk"); got != "TRFBANK" {
		t.Fatalf("kodeProduk = %q, want TRFBANK", got)
	}
	if got := loketPayloadString(payload, "nomor_rekening"); got != "0148545383213" {
		t.Fatalf("nomor_rekening = %q, want 0148545383213", got)
	}
	if got := loketPayloadString(payload, "provider_order_id"); got != "260605172937878186ED" {
		t.Fatalf("provider_order_id = %q, want 260605172937878186ED", got)
	}
	if got := loketPayloadString(payload, "orderId"); got != "" {
		t.Fatalf("orderId = %q, want empty because Otomax order is not P24 refid", got)
	}
}

func TestParseLoketBayarCallbackPayloadOtomaxTextDoubleEncodedQuery(t *testing.T) {
	text := "#juli553c396fe6 TRX TOPUP OVOPLUS ke 083853205326 status SUKSES.SN:ERIFAISAL AKBAR/083853205326/52000/2076214068716007424.HARGA:52675.SALDO:3504904015"
	encodedOnce := url.QueryEscape(text)
	raw := "q=" + url.QueryEscape(encodedOnce)
	q, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}

	payload, refid, rc, msg, providerRef, price, saldo := parseLoketBayarCallbackPayload(raw, q)

	if refid != "juli553c396fe6" {
		t.Fatalf("refid = %q, want juli553c396fe6", refid)
	}
	if rc != "SUKSES" {
		t.Fatalf("rc = %q, want SUKSES", rc)
	}
	if msg != text {
		t.Fatalf("msg = %q, want decoded callback text", msg)
	}
	wantRef := "ERIFAISAL AKBAR/083853205326/52000/2076214068716007424"
	if providerRef != wantRef {
		t.Fatalf("providerRef = %q, want %q", providerRef, wantRef)
	}
	if price != 52675 {
		t.Fatalf("price = %d, want 52675", price)
	}
	if saldo == nil || *saldo != 3504904015 {
		if saldo == nil {
			t.Fatalf("saldo = nil, want 3504904015")
		}
		t.Fatalf("saldo = %d, want 3504904015", *saldo)
	}
	if got := loketPayloadString(payload, "kodeProduk"); got != "OVOPLUS" {
		t.Fatalf("kodeProduk = %q, want OVOPLUS", got)
	}
	if got := loketPayloadString(payload, "nomor_rekening"); got != "083853205326" {
		t.Fatalf("nomor_rekening = %q, want 083853205326", got)
	}
}

func TestParseLoketBayarCallbackPayloadOtomaxTextRawBody(t *testing.T) {
	raw := "#juni79177fdcd2 TRX TOPUP TRFBANK ke 0190973005338 status GAGAL. TRANSAKSI GAGAL DI PROVIDER. HARGA: 50700. SALDO: 92069848"

	payload, refid, rc, msg, _, price, saldo := parseLoketBayarCallbackPayload(raw, url.Values{})

	if refid != "juni79177fdcd2" {
		t.Fatalf("refid = %q, want juni79177fdcd2", refid)
	}
	if rc != "GAGAL" {
		t.Fatalf("rc = %q, want GAGAL", rc)
	}
	if msg != raw {
		t.Fatalf("msg = %q, want raw callback text", msg)
	}
	if price != 50700 {
		t.Fatalf("price = %d, want 50700", price)
	}
	if saldo == nil || *saldo != 92069848 {
		if saldo == nil {
			t.Fatalf("saldo = nil, want 92069848")
		}
		t.Fatalf("saldo = %d, want 92069848", *saldo)
	}
	if got := loketPayloadString(payload, "nomor_rekening"); got != "0190973005338" {
		t.Fatalf("nomor_rekening = %q, want 0190973005338", got)
	}
}

func TestParseLoketBayarCallbackPayloadDepositVATicket(t *testing.T) {
	raw := "#10331 TUJUAN 8779600000066862 STATUS SUKSES.SALDO:4187979713"

	payload, refid, rc, msg, providerRef, price, saldo := parseLoketBayarCallbackPayload(raw, url.Values{})

	if refid != "10331" {
		t.Fatalf("refid = %q, want 10331", refid)
	}
	if rc != "SUKSES" {
		t.Fatalf("rc = %q, want SUKSES", rc)
	}
	if msg != raw {
		t.Fatalf("msg = %q, want raw callback text", msg)
	}
	if providerRef != "" {
		t.Fatalf("providerRef = %q, want empty", providerRef)
	}
	if price != 0 {
		t.Fatalf("price = %d, want 0", price)
	}
	if saldo == nil || *saldo != 4187979713 {
		if saldo == nil {
			t.Fatalf("saldo = nil, want 4187979713")
		}
		t.Fatalf("saldo = %d, want 4187979713", *saldo)
	}
	if got := loketPayloadString(payload, "nomor_rekening"); got != "8779600000066862" {
		t.Fatalf("nomor_rekening = %q, want 8779600000066862", got)
	}
}

func TestParseLoketBayarCallbackPayloadOtomaxSNLabel(t *testing.T) {
	raw := "#1JS493CHD TRX TOPUP DANAPLUS ke 083194104413 status SUKSES.SN:DNID 083194104413/083194104413/10000/2026062710121481030100166432841869275.HARGA:10100.SALDO:309311387"

	payload, refid, rc, msg, providerRef, price, saldo := parseLoketBayarCallbackPayload(raw, url.Values{})

	if refid != "1JS493CHD" {
		t.Fatalf("refid = %q, want 1JS493CHD", refid)
	}
	if rc != "SUKSES" {
		t.Fatalf("rc = %q, want SUKSES", rc)
	}
	if msg != raw {
		t.Fatalf("msg = %q, want raw callback text", msg)
	}
	wantRef := "DNID 083194104413/083194104413/10000/2026062710121481030100166432841869275"
	if providerRef != wantRef {
		t.Fatalf("providerRef = %q, want %q", providerRef, wantRef)
	}
	if price != 10100 {
		t.Fatalf("price = %d, want 10100", price)
	}
	if saldo == nil || *saldo != 309311387 {
		if saldo == nil {
			t.Fatalf("saldo = nil, want 309311387")
		}
		t.Fatalf("saldo = %d, want 309311387", *saldo)
	}
	if got := loketPayloadString(payload, "kodeProduk"); got != "DANAPLUS" {
		t.Fatalf("kodeProduk = %q, want DANAPLUS", got)
	}
	if got := loketPayloadString(payload, "nomor_rekening"); got != "083194104413" {
		t.Fatalf("nomor_rekening = %q, want 083194104413", got)
	}
}
