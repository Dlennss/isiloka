package helper

import "testing"

func TestParseAppBillingInquiryPLNSuccess(t *testing.T) {
	msg := "SP343444 INQ Trx ID:84282316CEK Produk:PLNPASCH Idpel:411020097453 Transaksi Berhasil/Status RC:00.Data Cust:Nama:ELSYEMINA MATULESSY SPD/TD:R1 1300/BLTH:MAR 26/St MTR:15919-16122,Data Tag:Rp327.601,Detil(Rp Tag:Rp322.601,Denda:Rp5.000).Adm:Rp4.200,Total Tag:Rp331.801.Waktu Trx:2026-03-27 16:25/"
	got := ParseAppBillingInquiry("yuscom", "success", msg)
	if got == nil {
		t.Fatal("expected parsed inquiry, got nil")
	}
	if got.CustomerName != "ELSYEMINA MATULESSY SPD" {
		t.Fatalf("unexpected customer name: %q", got.CustomerName)
	}
	if got.MeterType != "R1 1300" {
		t.Fatalf("unexpected meter type: %q", got.MeterType)
	}
	if got.PeriodLabel != "MAR 26" {
		t.Fatalf("unexpected period label: %q", got.PeriodLabel)
	}
	if got.MeterRange != "15919-16122" {
		t.Fatalf("unexpected meter range: %q", got.MeterRange)
	}
	if got.UsageLabel != "TD:R1 1300/BLTH:MAR 26/St MTR:15919-16122" {
		t.Fatalf("unexpected usage label: %q", got.UsageLabel)
	}
	if got.BillAmount != 322601 {
		t.Fatalf("unexpected bill amount: %d", got.BillAmount)
	}
	if got.PenaltyAmount != 5000 {
		t.Fatalf("unexpected penalty amount: %d", got.PenaltyAmount)
	}
	if got.AdminAmount != 4200 {
		t.Fatalf("unexpected admin amount: %d", got.AdminAmount)
	}
	if got.TotalAmount != 331801 {
		t.Fatalf("unexpected total amount: %d", got.TotalAmount)
	}
	if !got.CanPay {
		t.Fatal("expected can_pay true")
	}
}

func TestParseAppBillingInquiryPLNAlreadyPaid(t *testing.T) {
	msg := "PLNC.120040257237 sdh pernah jam 16:16, status Gagal. SN/Ref: .Gagal Trx ke-2/hr: PLNC.2.120040257237.pin.Tagihan sudah terbayar.  Saldo 392.295"
	got := ParseAppBillingInquiry("yuscom", "failed", msg)
	if got == nil {
		t.Fatal("expected parsed inquiry, got nil")
	}
	if got.CanPay {
		t.Fatal("expected can_pay false")
	}
	if got.TotalAmount != 0 {
		t.Fatalf("expected zero total amount, got %d", got.TotalAmount)
	}
	if got.DisplayMessage != "Tagihan sudah terbayar" {
		t.Fatalf("unexpected display message: %q", got.DisplayMessage)
	}
}
