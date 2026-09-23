package minions

import "testing"

func TestParseLastBalanceFromMsg(t *testing.T) {
	msg := "status=20&message=HDANA.08123456789 SUKSES. SN: DANA TEST/21000/2026033110121481030100166446394454002 saldo 5,000,924 - 21,055 = 4,979,869 @31/03 19:39 R#REF123 trxid=174337105, harga = 21,055"
	got, ok := ParseLastBalanceFromMsg(msg)
	if !ok || got != 4979869 {
		t.Fatalf("ParseLastBalanceFromMsg() = (%d,%v), want (4979869,true)", got, ok)
	}
}

func TestParsePriceFromMsg(t *testing.T) {
	msg := "status=20&message=HDANA.08123456789 SUKSES. SN: DANA TEST/21000/2026033110121481030100166446394454002 saldo 5,000,924 - 21,055 = 4,979,869 @31/03 19:39 R#REF123 trxid=174337105, harga = 21,055"
	got := ParsePriceFromMsg(msg)
	if got != 21055 {
		t.Fatalf("ParsePriceFromMsg() = %d, want 21055", got)
	}
}
