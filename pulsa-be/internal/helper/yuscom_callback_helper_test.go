package helper

import "testing"

func TestParseSaldoAfterFromStockMsgSupportsTrionikFormat(t *testing.T) {
	msg := "14:21 25369639 R#1775028094123 DANA.082124307365 HRG : 10.030 SUKSES, SN: DNID 082124307365/10000/2026040110121481030100166446394770662. Stok Pulsa 5.000.599 - 10.030 = 4.990.569 Trx=0 @01/04 14:21:52"
	got, ok := ParseSaldoAfterFromStockMsg(msg)
	if !ok {
		t.Fatalf("expected trionik stock format to be parsed")
	}
	if got != 4990569 {
		t.Fatalf("unexpected parsed saldo: got=%d want=%d", got, 4990569)
	}
}
