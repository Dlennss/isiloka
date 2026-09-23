package helper

import "testing"

func TestProviderResponseStateOfSMB(t *testing.T) {
	t.Run("success json stays success", func(t *testing.T) {
		body := `{"success":true,"produk":"DANA","tujuan":"100000@08123","reffid":"BYR123","rc":"1","status":1,"sn":"ABC","msg":"PAYSUKSES TRANSAKSI BERHASIL"}`
		if got := ProviderResponseStateOf("smb", "1", body); got != ProviderResponseSuccess {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseSuccess)
		}
	})

	t.Run("pending json stays pending", func(t *testing.T) {
		body := `{"success":true,"produk":"DANA","tujuan":"100000@08123","reffid":"BYR123","rc":"68","status":"68","sn":"-","msg":"PENDING DALAM PROSES"}`
		if got := ProviderResponseStateOf("smb", "68", body); got != ProviderResponsePending {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
		}
	})

	t.Run("direct payload under proses stays pending", func(t *testing.T) {
		body := `{"success":true,"produk":"ELDN","tujuan":"082235804266","reffid":"1775158471455740","rc":"0068","harga":299065,"msg":"Trx ELDN 082235804266 Under proses...","saldo":"106.204.167"}`
		if got := ProviderResponseStateOf("smb", "0068", body); got != ProviderResponsePending {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
		}
	})

	t.Run("check payload inqsukses stays pending", func(t *testing.T) {
		body := `INQSUKSES TRANSFER KE REKENING:082235804266,BANK:DANA,NAMA:ASNXXX,NOMINAL:299000,ADMIN:DANA/2500,JUMLAH:301500,HARGA:299.100,SISASALDO:106.204.167 - 0 = 106.204.167, @WAKTU:2026/04/03 02:35:00.`
		if got := ProviderResponseStateOf("smb", "1", body); got != ProviderResponsePending {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
		}
	})

	t.Run("smb failed after prior success stays failed", func(t *testing.T) {
		body := `[SEMPAT SUKSES SN:-] REFF#smpaycb50fd5023 BIFASTOPEN 0141390326293 GAGAL, KET: REFUND/SALDO DIKEMBALIKAN, SISASALDO: 1.109.825.374 @WAKTU:02/05/2026. | sebelumnya: REFF#smpaycb50fd5023 BIFASTOPEN.0141390326293 BERHASIL`
		if got := ProviderResponseStateOf("smb", "2", body); got != ProviderResponseFailed {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
		}
	})

	t.Run("gopay account validation failure stays failed", func(t *testing.T) {
		body := `{"success":false,"produk":"GOPAY","tujuan":"150000@089533010870","reffid":"CEK1775280515060398","rc":"2","status":2,"sn":"","msg":"REFF#CEK1775280515060398 GOPAY 150000@089533010870 GAGAL, KET: AN ERROR OCCURRED WHEN DOING ACCOUNT VALIDATION, SISA SALDO: 18.505.568  @WAKTU:04/04/2026 12:31:27."}`
		if got := ProviderResponseStateOf("smb", "2", body); got != ProviderResponseFailed {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
		}
	})

	t.Run("gemilang pending stays pending", func(t *testing.T) {
		body := `status=1&message=Trx IR5.085733686642 akan diproses @13.41. Stok 27.310.535 - 6.438 = 27.304.097 * TRX Normal`
		if got := ProviderResponseStateOf("gemilang", "1", body); got != ProviderResponsePending {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
		}
	})

	t.Run("gemilang sukses stays success", func(t *testing.T) {
		body := `Trx IR5.085733686642 SUKSES. SN/Ref: 03961200033376299772. Stok 27.310.535-6.438=27.304.097 @05/11 13.41.05.`
		if got := ProviderResponseStateOf("gemilang", "20", body); got != ProviderResponseSuccess {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseSuccess)
		}
	})

	t.Run("gemilang failure stays failed", func(t *testing.T) {
		body := `IN5.082206775242 GAGAL. Nomor tujuan salah. Sal 38.665 @19:06 * TRX Normal`
		if got := ProviderResponseStateOf("gemilang", "44", body); got != ProviderResponseFailed {
			t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
		}
	})
}

func TestProviderResponseStateOfLoketBayar(t *testing.T) {
	if got := ProviderResponseStateOf("loketbayar", "9", "SEDANG DIPROSES DI PROVIDER"); got != ProviderResponsePending {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
	}
	if got := ProviderResponseStateOf("loketbayar", "", ". SALDO: 3405450140"); got != ProviderResponsePending {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
	}
	if got := ProviderResponseStateFromBody("loketbayar", "status= keterangan=. SALDO: 3405450140"); got != ProviderResponsePending {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponsePending)
	}
	if got := ProviderResponseStateOf("loketbayar", "28", "Conflict reference id"); got != ProviderResponseFailed {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
	}
	if got := ProviderResponseStateOf("loketbayar", "96", "Transaksi Gagal"); got != ProviderResponseFailed {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
	}
	if got := ProviderResponseStateOf("loketbayar", "", "#REF TRX TOPUP DANAPLUS status GAGAL.Saldo tidak mencukupi.HARGA:0.SALDO:0"); got != ProviderResponseFailed {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
	}
	if got := ProviderResponseStateOf("loketbayar", "PENDING", "#REF TRX TOPUP TRFBANK status GAGAL. SEDANG DIPROSES|. SALDO: 7081320990"); got != ProviderResponseFailed {
		t.Fatalf("unexpected state: got=%q want=%q", got, ProviderResponseFailed)
	}
}

func TestProviderResponseStateOfRajabiller(t *testing.T) {
	tests := []struct {
		name string
		rc   string
		msg  string
		want ProviderResponseState
	}{
		{name: "rc 00 wins over process text", rc: "00", msg: "SEDANG DIPROSES", want: ProviderResponseSuccess},
		{name: "rc 68 wins over success text", rc: "68", msg: "Success", want: ProviderResponsePending},
		{name: "known failed rc wins over process text", rc: "16", msg: "SEDANG DIPROSES", want: ProviderResponseFailed},
		{name: "unknown rc stays pending", rc: "2007", msg: "Purchase Request Failed", want: ProviderResponsePending},
		{name: "empty rc can still use failure text", rc: "", msg: "Gagal - nomor salah", want: ProviderResponseFailed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ProviderResponseStateOf("rajabiller", tc.rc, tc.msg); got != tc.want {
				t.Fatalf("unexpected state: got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestProviderResponseStateOfPulsa24Jam(t *testing.T) {
	tests := []struct {
		name string
		rc   string
		msg  string
		want ProviderResponseState
	}{
		{
			name: "ok true accepted is pending",
			msg:  `{"ok":true,"refid":"PKA29","status":1,"message":"Transaksi sedang diproses"}`,
			want: ProviderResponsePending,
		},
		{
			name: "ok true without final status is pending",
			msg:  `{"ok":true,"refid":"PKA29","message":"request diterima"}`,
			want: ProviderResponsePending,
		},
		{
			name: "status two is success",
			msg:  `{"ok":true,"refid":"PKA29","status":2,"sn":"1234567890","message":"SUKSES"}`,
			want: ProviderResponseSuccess,
		},
		{
			name: "status three is failed",
			msg:  `{"ok":true,"refid":"PKA29","status":3,"message":"Nomor tujuan salah"}`,
			want: ProviderResponseFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProviderResponseStateOf("pulsa24jam", tt.rc, tt.msg); got != tt.want {
				t.Fatalf("unexpected state: got=%q want=%q", got, tt.want)
			}
		})
	}
}
