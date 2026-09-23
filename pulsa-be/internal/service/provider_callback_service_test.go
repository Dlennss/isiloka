package service

import (
	"net/url"
	"testing"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

func TestResolveYuscomFinalStatus(t *testing.T) {
	t.Run("explicit failure beats pending code", func(t *testing.T) {
		got := resolveYuscomFinalStatus(1, "1", "R#123 T#99 GOPAY ke 08123 GAGAL. Nomor tujuan salah.")
		if got != "failed" {
			t.Fatalf("unexpected status: got=%q want=%q", got, "failed")
		}
	})

	t.Run("accepted success remains success", func(t *testing.T) {
		got := resolveYuscomFinalStatus(20, "20", "R#123 SUKSES. SN/Ref: abc")
		if got != "success" {
			t.Fatalf("unexpected status: got=%q want=%q", got, "success")
		}
	})
}

func TestResolveMultikomFinalStatus(t *testing.T) {
	t.Run("explicit failure beats pending code", func(t *testing.T) {
		got := resolveMultikomFinalStatus(1, "1", "R#123 DANA ke 08123 GAGAL. Nomor tujuan salah.")
		if got != "failed" {
			t.Fatalf("unexpected status: got=%q want=%q", got, "failed")
		}
	})

	t.Run("pending text stays pending", func(t *testing.T) {
		got := resolveMultikomFinalStatus(2, "2", "R#123 masih diproses")
		if got != "pending" {
			t.Fatalf("unexpected status: got=%q want=%q", got, "pending")
		}
	})
}

func TestHasProviderFailureSignal(t *testing.T) {
	cases := []string{
		"GAGAL. Nomor tujuan salah.",
		"Qty tidak sesuai. Allowed QTY is 1000-200000.",
		"Tidak diproses karena cutoff",
		"Dibatalkan provider",
	}
	for _, msg := range cases {
		if !hasProviderFailureSignal(msg) {
			t.Fatalf("expected failure signal for %q", msg)
		}
	}

	if hasProviderFailureSignal("AKAN DIPROSES") {
		t.Fatalf("did not expect failure signal for accepted message")
	}
}

func TestBuildDirectMemberWebhookPayloadIncludesExplicitNominals(t *testing.T) {
	trx := &repository.CallbackTrxMemberFull{
		ID:                    7,
		RefID:                 "REF789",
		Perintah:              "PAY",
		KodeProduk:            "DANA",
		Tujuan:                "08123",
		Qty:                   200000,
		QtyProvider:           199400,
		ChargeReceiverApplied: true,
		BiayaPerkiraan:        200000,
		FeeMemberRp:           600,
		HargaMember:           200000,
	}

	got := buildDirectMemberWebhookPayload(trx, "success", "ok", "PRX", "SNX", 1000000, 200000)
	trxOut, ok := got["trx"].(map[string]any)
	if !ok {
		t.Fatalf("trx payload missing or invalid")
	}
	if trxOut["qty_provider"] != int64(199400) {
		t.Fatalf("unexpected qty_provider: %#v", trxOut["qty_provider"])
	}
	if trxOut["harga_member"] != int64(200000) {
		t.Fatalf("unexpected harga_member: %#v", trxOut["harga_member"])
	}
}

func TestPulsa24JamFinalStatusUsesNumericCallbackStatus(t *testing.T) {
	tests := []struct {
		name string
		data pulsa24JamCallbackData
		want string
	}{
		{
			name: "status two success",
			data: pulsa24JamCallbackData{status: "2", rc: "00", msg: "SUKSES"},
			want: "success",
		},
		{
			name: "status three failed",
			data: pulsa24JamCallbackData{status: "3", rc: "55", msg: "Provider timeout"},
			want: "failed",
		},
		{
			name: "status one pending",
			data: pulsa24JamCallbackData{status: "1", msg: "Sedang diproses"},
			want: "pending",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pulsa24JamFinalStatus(tt.data); got != tt.want {
				t.Fatalf("pulsa24JamFinalStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFallbackProviderClassifierUsesTargetProviderRules(t *testing.T) {
	yuscomAccepted := "status=1&message=R#123 T#99 DANA.08123 akan diproses @17:30"
	talentaAccepted := "status=1&message=TALENTATRONIK : trx Hari ini R#123 TDBN.08123 akan diproses @ 23/03 17.30"
	multikomRejected := "status=43&message=R#123 DANARP ke 08123 GAGAL. Stok tidak cukup."
	smbAccepted := `{"success":true,"produk":"DANA","tujuan":"100000@08123","reffid":"BYR123","rc":"68","status":"68","sn":"-","msg":"PENDING DALAM PROSES"}`
	smbSuccess := `{"success":true,"produk":"DANA","tujuan":"100000@08123","reffid":"BYR123","rc":"1","status":1,"sn":"ABC","msg":"PAYSUKSES TRANSAKSI BERHASIL"}`

	if !fallbackAcceptedForProvider("yuscom", yuscomAccepted) {
		t.Fatalf("expected yuscom accepted body to be classified by yuscom rules")
	}
	if fallbackAcceptedForProvider("yuscom", multikomRejected) {
		t.Fatalf("did not expect multikom reject body to be accepted as yuscom")
	}

	if !fallbackAcceptedForProvider("talentapay", talentaAccepted) {
		t.Fatalf("expected talenta accepted body to be classified by talenta rules")
	}
	if fallbackAcceptedForProvider("talentapay", multikomRejected) {
		t.Fatalf("did not expect multikom reject body to be classified as talenta accepted")
	}

	if !fallbackImmediateRejectForProvider("multikom", multikomRejected) {
		t.Fatalf("expected multikom reject body to be classified by multikom rules")
	}
	if fallbackImmediateRejectForProvider("yuscom", talentaAccepted) {
		t.Fatalf("did not expect talenta accepted body to be classified as yuscom reject")
	}
	if !fallbackAcceptedForProvider("smb", smbAccepted) {
		t.Fatalf("expected smb accepted/pending body to be classified by smb rules")
	}
	if !fallbackAcceptedForProvider("smb", smbSuccess) {
		t.Fatalf("expected smb success body to be classified by smb rules")
	}
	if fallbackImmediateRejectForProvider("smb", smbAccepted) {
		t.Fatalf("did not expect smb pending body to be classified as immediate reject")
	}
}

func TestHasOtherProviderAttempt(t *testing.T) {
	t.Run("detects other provider usage", func(t *testing.T) {
		used := map[string]bool{
			"javapay": true,
			"yuscom":  true,
		}
		ok, provider := hasOtherProviderAttempt("javapay", used)
		if !ok {
			t.Fatalf("expected other provider usage to be detected")
		}
		if provider != "yuscom" {
			t.Fatalf("unexpected provider: got=%q want=%q", provider, "yuscom")
		}
	})

	t.Run("ignores excluded provider", func(t *testing.T) {
		used := map[string]bool{
			"javapay": true,
		}
		ok, provider := hasOtherProviderAttempt("javapay", used)
		if ok {
			t.Fatalf("did not expect other provider usage, got provider=%q", provider)
		}
	})
}

func TestHasOtherProviderAttemptForRefIgnoresFinalFailedProvider(t *testing.T) {
	multikomFailed := &model.JavapayTrxRow{
		ID:         10,
		KodeRespon: strPtr("61"),
		Pesan:      strPtr("R#123 DANARP.2710000.08123 GAGAL, Qty tidak sesuai. Allowed QTY is 1000-1000000."),
	}
	if ok, _, _, _, _ := providerRowSuccessState("multikom", multikomFailed); ok {
		t.Fatalf("expected multikom failed row to not be success")
	}
	if ok, _, _ := providerRowPendingState("multikom", multikomFailed); ok {
		t.Fatalf("expected multikom failed row to not be pending")
	}
}

func TestSMBCallbackValueReadsLiveQueryAliases(t *testing.T) {
	q := url.Values{
		"clientid":   []string{"CEK1775119139077414"},
		"statuscode": []string{"1"},
		"msg":        []string{"INQSUKSES TRANSFER KE REKENING:085273037912,BANK:DANA"},
	}

	if got := smbCallbackValue(q, "message", "msg", "pesan"); got != "INQSUKSES TRANSFER KE REKENING:085273037912,BANK:DANA" {
		t.Fatalf("unexpected msg alias result: %q", got)
	}
	if got := smbCallbackValue(q, "status", "statuscode", "rc"); got != "1" {
		t.Fatalf("unexpected status alias result: %q", got)
	}
	if got := smbCallbackValue(q, "refid", "idtrx", "clientid", "reffid"); got != "CEK1775119139077414" {
		t.Fatalf("unexpected refid alias result: %q", got)
	}
	if got := smbCallbackStage(smbCallbackValue(q, "clientid")); got != "check" {
		t.Fatalf("unexpected stage: %q", got)
	}
}

func TestSMBCheckCallbackAlreadyPromoted(t *testing.T) {
	payMsg := "PAYSUKSES TRANSFER KE REKENING"
	payRef := "2026040210121481030100166446395367956/3501"
	payRow := &repository.ProviderTrxRefRow{
		Pesan:       &payMsg,
		NoReferensi: &payRef,
	}
	if !smbCheckCallbackAlreadyPromoted(payRow) {
		t.Fatalf("expected pay-started row to block duplicate check callback")
	}

	checkMsg := "INQSUKSES TRANSFER KE REKENING"
	checkRef := "NAMA:DNID-082124307365"
	checkRow := &repository.ProviderTrxRefRow{
		Pesan:       &checkMsg,
		NoReferensi: &checkRef,
	}
	if smbCheckCallbackAlreadyPromoted(checkRow) {
		t.Fatalf("did not expect plain inquiry row to block check callback")
	}
}

func TestSMBEffectiveWalletPrice(t *testing.T) {
	rowPrice := int64(149065)
	row := &repository.ProviderTrxRefRow{Harga: &rowPrice}

	if got := smbEffectiveWalletPrice(1100, row); got != 1100 {
		t.Fatalf("unexpected callback price precedence: got=%d want=%d", got, 1100)
	}
	if got := smbEffectiveWalletPrice(0, row); got != 149065 {
		t.Fatalf("unexpected row fallback price: got=%d want=%d", got, 149065)
	}
	if got := smbEffectiveWalletPrice(0, nil); got != 0 {
		t.Fatalf("unexpected zero fallback price: got=%d want=%d", got, 0)
	}
}

func TestSMBCheckCallbackAttemptClosed(t *testing.T) {
	failRC := "91"
	failMsg := "SMB GAGAL: callback cek tidak diterima dalam 5 detik"
	closedRow := &repository.ProviderTrxRefRow{
		KodeRespon: &failRC,
		Pesan:      &failMsg,
	}
	if !smbCheckCallbackAttemptClosed(closedRow) {
		t.Fatalf("expected closed SMB attempt to ignore late check callback")
	}

	pendingRC := "1"
	pendingMsg := "INQSUKSES TRANSFER KE REKENING"
	pendingRow := &repository.ProviderTrxRefRow{
		KodeRespon: &pendingRC,
		Pesan:      &pendingMsg,
	}
	if smbCheckCallbackAttemptClosed(pendingRow) {
		t.Fatalf("did not expect active SMB check row to be considered closed")
	}
}

func TestSMBCheckCallbackReadyToPay(t *testing.T) {
	if !smbCheckCallbackReadyToPay("INQSUKSES TRANSFER KE REKENING:085273037912,BANK:DANA") {
		t.Fatalf("expected inquiry success callback to be payable")
	}
	if smbCheckCallbackReadyToPay("PENDING DALAM PROSES") {
		t.Fatalf("did not expect plain pending callback to trigger pay")
	}
	if smbCheckCallbackReadyToPay("") {
		t.Fatalf("did not expect empty callback message to trigger pay")
	}
}

func TestSMBPersistedMessage(t *testing.T) {
	rawCheck := "INQSUKSES TRANSFER KE REKENING:085273037912,BANK:DANA,NAMA:TEST"
	if got := smbPersistedMessage("check", rawCheck); got != rawCheck {
		t.Fatalf("unexpected persisted check message: got=%q want=%q", got, rawCheck)
	}

	rawPay := "PAYSUKSES TRANSFER KE REKENING:085273037912,BANK:DANA"
	if got := smbPersistedMessage("pay", rawPay); got != rawPay {
		t.Fatalf("unexpected persisted pay message: got=%q want=%q", got, rawPay)
	}

	rawDirect := "REFF#1775 ELDN.08123 BERHASIL"
	if got := smbPersistedMessage("", rawDirect); got != rawDirect {
		t.Fatalf("unexpected persisted direct message: got=%q want=%q", got, rawDirect)
	}
}

func TestSMBPayAfterCheckBodyClassifiesFinalState(t *testing.T) {
	successBody := `{"success":true,"rc":"1","msg":"PAYSUKSES TRANSAKSI BERHASIL","sn":"ABC"}`
	if got := helper.ProviderResponseStateOf("smb", "1", successBody); got != helper.ProviderResponseSuccess {
		t.Fatalf("unexpected success state: got=%q want=%q", got, helper.ProviderResponseSuccess)
	}

	failedBody := `{"success":false,"rc":"61","msg":"cek balance dan transaksi ke CS/admin"}`
	if got := helper.ProviderResponseStateOf("smb", "61", failedBody); got != helper.ProviderResponseFailed {
		t.Fatalf("unexpected failed state: got=%q want=%q", got, helper.ProviderResponseFailed)
	}
}

func strPtr(v string) *string { return &v }
