package helper

import "testing"

func TestExtractProviderInfoSupportsMultipleProviderPatterns(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantReff string
		wantSN   string
	}{
		{
			name:     "talenta dana",
			msg:      "TALENTATRONIK : R#1774188045796675 TDBN329000.081389929865 SUKSES. SN/Ref: TopUp DANA MUHXXXX IKBXX/329000/2026032210121481030100166192990543034.. Saldo Rp.134.786.318 - Hrg Rp.329.033,Saldo akhir Rp. 134.457.285 @21.01 - 2026032210121481030100166192990543034",
			wantReff: "2026032210121481030100166192990543034",
			wantSN:   "2026032210121481030100166192990543034",
		},
		{
			name:     "yuscom gopay",
			msg:      "R#1774190789926414204 T#84006103 GPAY.081911196525 SUKSES. SN/Ref: MURYANTO/150000/8iqs5FHOx8P0NB6nGM. Saldo 833.311.709-150.950=833.160.759 @22/03 21:47:45 C#1 KP#GPAY. - 84006103",
			wantReff: "8iqs5FHOx8P0NB6nGM",
			wantSN:   "8iqs5FHOx8P0NB6nGM",
		},
		{
			name:     "multikom linkaja",
			msg:      "R#1774181897069912 SUKSES LNRP ke 081260607840 SN: TopUp LinkAja - MUHAMMAD KHAIDIR/235000/Reff:000383327037. Stok 54.636.832 - 235700 = 54.401.132. @22/03 19:20",
			wantReff: "000383327037",
			wantSN:   "000383327037",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractProviderInfo(tc.msg)
			if got.Reff != tc.wantReff || got.SN != tc.wantSN {
				t.Fatalf("unexpected provider info: got=%+v wantReff=%q wantSN=%q", got, tc.wantReff, tc.wantSN)
			}
		})
	}
}

func TestSafeMemberKeteranganKeepsUsefulFailureReason(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{name: "rc 52 only", msg: "status=52", want: "Nomor tujuan salah"},
		{name: "nomor tujuan salah", msg: "R#1 GAGAL. Nomor tujuan salah.", want: "Nomor tujuan salah"},
		{name: "qty tidak sesuai", msg: "status=61 Qty tidak sesuai. Allowed QTY is 1000-500000", want: "Qty tidak sesuai"},
		{name: "cutoff", msg: "provider cutoff sementara", want: "Provider cutoff"},
		{name: "fallback generic", msg: "some raw provider failure", want: "Transaksi gagal"},
		{name: "loketbayar raw saldo", msg: "loketbayar reject bisnis: . SALDO: 474534515", want: "Transaksi gagal"},
		{name: "smb sql raw", msg: "smb reject bisnis: {\"success\":false,\"msg\":{\"sqlMessage\":\"Data too long\"}}", want: "Transaksi gagal"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := SafeMemberKeterangan("failed", tc.msg)
			if got != tc.want {
				t.Fatalf("unexpected failure message: got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestShouldBlockProviderFallback(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "rc 52 final invalid destination", msg: "status=52", want: true},
		{name: "invalid number", msg: "R#1 GAGAL. Nomor tujuan salah.", want: true},
		{name: "ajs customer account not found", msg: "Qty#100000 R#1775445461678200 OVO.08954284702 Hrg:100.665 GAGAL. Customer Account Not Found. Stok:113.533.847 @10:57", want: true},
		{name: "talenta invalid id", msg: "GAGAL. Nomor /ID salah/tidak bisa diproses. cek nomor sebelum diulang.", want: true},
		{name: "account limit", msg: "GAGAL. Batas maksimal pembelian/ Akun sdh Limit.Stok dikembalikan.", want: true},
		{name: "qty limit retryable", msg: "status=61 Qty tidak sesuai. Allowed QTY is 1000-500000", want: false},
		{name: "provider timeout", msg: "status=55 STATUS TIMEOUT", want: false},
		{name: "provider saldo kurang", msg: "GAGAL. Saldo provider tidak cukup", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldBlockProviderFallback(tc.msg); got != tc.want {
				t.Fatalf("unexpected block decision: got=%v want=%v msg=%q", got, tc.want, tc.msg)
			}
		})
	}
}

func TestLooksLikeProviderOutOfBalance(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "loketbayar saldo tidak mencukupi", msg: "#ref TRX TOPUP DANAPLUS status GAGAL.Saldo tidak mencukupi.HARGA:0.SALDO:0", want: true},
		{name: "provider saldo kurang", msg: "GAGAL. Saldo provider tidak cukup", want: true},
		{name: "stock not enough", msg: "produk gagal stok tidak cukup", want: true},
		{name: "insufficient balance", msg: "failed: insufficient balance", want: true},
		{name: "invalid destination", msg: "Nomor tujuan salah", want: false},
		{name: "member limit", msg: "Akun sdh limit", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := LooksLikeProviderOutOfBalance(tc.msg); got != tc.want {
				t.Fatalf("unexpected out-of-balance decision: got=%v want=%v msg=%q", got, tc.want, tc.msg)
			}
		})
	}
}

func TestExtractSaldoTerakhirFromMsgSupportsAJSStockLabel(t *testing.T) {
	msg := "Qty#99000 R#1775195853629243 DND.088210501055 Hrg:99.090 Sukses. SN: IPIX/2026040310121481030100166370096141302. Stok:4.619.617 @03/04 12:56:59 RC:00 GLOBAL"
	got, ok := ExtractSaldoTerakhirFromMsg(msg)
	if !ok {
		t.Fatalf("expected AJS stock label format to be parsed")
	}
	if got != 4619617 {
		t.Fatalf("unexpected parsed saldo: got=%d want=%d", got, 4619617)
	}
}

func TestAJSQueueWaitResponseShouldStayPending(t *testing.T) {
	msg := "R#1775410549301713872 Mhn tunggu trx sblmnya selesai: DND.085280973341 @00:38, status Menunggu Jawaban. Stok:8.616.119"
	if !LooksLikeAJSAccepted(msg) {
		t.Fatalf("expected queue-wait AJS response to be accepted/pending")
	}
	if LooksLikeAJSImmediateReject(msg) {
		t.Fatalf("expected queue-wait AJS response to not be treated as immediate reject")
	}
	if ShouldBlockProviderFallback(msg) {
		t.Fatalf("expected queue-wait AJS response to not block fallback categorization")
	}
}
