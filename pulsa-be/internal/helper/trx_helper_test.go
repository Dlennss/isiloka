package helper

import "testing"

func TestLooksLikeTalentaAccepted(t *testing.T) {
	if !LooksLikeTalentaAccepted("status=20, message=SUKSES SN/Ref: 12345") {
		t.Fatalf("expected talenta success body to be accepted")
	}
	if !LooksLikeTalentaAccepted("status=2, message=AKAN DIPROSES") {
		t.Fatalf("expected talenta pending body to be accepted")
	}
	if LooksLikeTalentaAccepted("status=61, message=Allowed Qty 100000") {
		t.Fatalf("did not expect talenta reject body to be accepted")
	}
}

func TestLooksLikeTalentaImmediateReject(t *testing.T) {
	if !LooksLikeTalentaImmediateReject("status=61, message=Allowed Qty 100000") {
		t.Fatalf("expected talenta allowed qty reject to be detected")
	}
	if !LooksLikeTalentaImmediateReject("message=Nomor tujuan salah") {
		t.Fatalf("expected talenta explicit failure to be detected")
	}
}

func TestLooksLikeTalentaSuccess(t *testing.T) {
	if !LooksLikeTalentaSuccess("status=20, message=SUKSES SN/Ref: 12345") {
		t.Fatalf("expected talenta success body to be success")
	}
	if LooksLikeTalentaSuccess("status=1, message=AKAN DIPROSES") {
		t.Fatalf("did not expect talenta pending body to be success")
	}
}

func TestLooksLikeMultikomAccepted(t *testing.T) {
	if !LooksLikeMultikomAccepted("status=20, message=STATUS=SUKSES. SN: 12345.") {
		t.Fatalf("expected multikom success body to be accepted")
	}
	if !LooksLikeMultikomAccepted("status=1, message=AKAN DIPROSES") {
		t.Fatalf("expected multikom pending body to be accepted")
	}
	if LooksLikeMultikomAccepted("status=55, message=STATUS TIMEOUT") {
		t.Fatalf("did not expect multikom failed body to be accepted")
	}
	if LooksLikeMultikomAccepted("status=43&message=R#1 DANARP ke 0821 GAGAL. Stok tidak cukup. Hrg 249.050, Stok 4.148, masih proses 0.") {
		t.Fatalf("did not expect multikom failed body with 'masih proses' text to be accepted")
	}
}

func TestLooksLikeMultikomImmediateReject(t *testing.T) {
	if !LooksLikeMultikomImmediateReject("status=55, message=STATUS TIMEOUT") {
		t.Fatalf("expected multikom timeout body to be reject")
	}
	if !LooksLikeMultikomImmediateReject("message=Nomor tujuan salah") {
		t.Fatalf("expected multikom explicit failure to be reject")
	}
}

func TestLooksLikeGemilangAccepted(t *testing.T) {
	if !LooksLikeGemilangAccepted("status=1&message=Trx IR5.085733686642 akan diproses @13.41. Stok - 6.438 = #dfxv592f07ypa61 * TRX Normal") {
		t.Fatalf("expected gemilang pending body to be accepted")
	}
	if LooksLikeGemilangAccepted("status=44&message=INVALID.08123456789 GAGAL, salah kode atau nominal. Stok 5.000.840 @17.03 * TRX Normal") {
		t.Fatalf("did not expect gemilang reject body to be accepted")
	}
}

func TestLooksLikeGemilangImmediateReject(t *testing.T) {
	if !LooksLikeGemilangImmediateReject("IN5.082206775242 GAGAL. Nomor tujuan salah. Sal 38.665 @19:06 * TRX Normal") {
		t.Fatalf("expected gemilang explicit failure to be reject")
	}
}

func TestLooksLikeYuscomSuccess(t *testing.T) {
	if !LooksLikeYuscomSuccess("status=20, message=SUKSES") {
		t.Fatalf("expected yuscom success body to be success")
	}
	if LooksLikeYuscomSuccess("status=1, message=AKAN DIPROSES") {
		t.Fatalf("did not expect yuscom pending body to be success")
	}
}

func TestIsPendingRCDoesNotTreatEmptyAsPending(t *testing.T) {
	if IsPendingRC("") {
		t.Fatalf("empty rc should not be treated as pending")
	}
	if !IsPendingRC("0") {
		t.Fatalf("rc 0 should still be treated as pending")
	}
}
