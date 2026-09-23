package service

import (
	"strings"
	"testing"
	"time"

	"pulsa2/model"
)

func TestProviderRowSuccessStateUsesProviderSpecificClassifier(t *testing.T) {
	talentaRejectMsg := "status=61, message=Allowed Qty 100000"
	talentaRow := &model.JavapayTrxRow{
		KodeRespon: &[]string{"61"}[0],
		Pesan:      &talentaRejectMsg,
	}
	if ok, _, _, _, _ := providerRowSuccessState("talentapay", talentaRow); ok {
		t.Fatalf("talenta reject body should not be treated as success")
	}

	multikomPendingMsg := "status=1, message=AKAN DIPROSES"
	multikomRow := &model.JavapayTrxRow{
		KodeRespon: &[]string{"1"}[0],
		Pesan:      &multikomPendingMsg,
	}
	if ok, _, _, _, _ := providerRowSuccessState("multikom", multikomRow); ok {
		t.Fatalf("multikom pending body should not be treated as success-state for reconcile")
	}
}

func TestProviderRowPendingStateDetectsAcceptedButNotFinal(t *testing.T) {
	talentaPendingMsg := "status=1&message=TALENTATRONIK : trx Hari ini 1.350 R#1774261465854774 TDBN.085849123701 akan diproses @ 23/03 17.24.26."
	talentaPendingRow := &model.JavapayTrxRow{
		KodeRespon: &[]string{"1"}[0],
		Pesan:      &talentaPendingMsg,
	}
	if ok, _, _ := providerRowPendingState("talentapay", talentaPendingRow); !ok {
		t.Fatalf("talenta accepted body should be treated as pending-state for reconcile")
	}

	talentaSuccessMsg := "status=20&message=TALENTATRONIK : R#1774261465854774 SUKSES. SN/Ref: 202603..."
	talentaSuccessRow := &model.JavapayTrxRow{
		KodeRespon: &[]string{"20"}[0],
		Pesan:      &talentaSuccessMsg,
	}
	if ok, _, _ := providerRowPendingState("talentapay", talentaSuccessRow); ok {
		t.Fatalf("talenta success body should not be treated as pending-state")
	}

	multikomFailedMsg := "status=43&message=R#1774261465854774 DANARP ke 085849123701 GAGAL. Stok tidak cukup."
	multikomFailedRow := &model.JavapayTrxRow{
		KodeRespon: &[]string{"43"}[0],
		Pesan:      &multikomFailedMsg,
	}
	if ok, _, _ := providerRowPendingState("multikom", multikomFailedRow); ok {
		t.Fatalf("multikom failed body should not be treated as pending-state")
	}

	javapayEmptyRow := &model.JavapayTrxRow{}
	if ok, _, _ := providerRowPendingState("javapay", javapayEmptyRow); ok {
		t.Fatalf("javapay row with empty rc/message should not be treated as pending-state")
	}
}

func TestProviderRowPendingMessageOverridesFailedStatusForStatusPay(t *testing.T) {
	trionikPendingMsg := "status=1&message=Mhn tunggu trx sblmnya selesai: #meie8b68cc9e8 DANA.0882022696559 @21:58, status Menunggu Jawaban. Stok Pulsa 971.491.512"
	row := &model.JavapayTrxRow{
		ID:         5152825,
		Provider:   "trionik",
		Status:     "failed",
		KodeRespon: &[]string{"1"}[0],
		Pesan:      &trionikPendingMsg,
	}

	if ok, rc, msg := providerRowPendingState("trionik", row); !ok || rc != "1" || msg != trionikPendingMsg {
		t.Fatalf("expected Trionik pending message to stay pending despite failed row status, ok=%v rc=%q msg=%q", ok, rc, msg)
	}
	if ok, _, _, _, _, _ := providerRowFailureState("trionik", row); ok {
		t.Fatalf("pending Trionik response must not be treated as provider failure")
	}
	if providerRowDefinitelyFailed("trionik", row) {
		t.Fatalf("pending Trionik response must block final-failed guard")
	}
}

func TestProviderRowHasProviderReply(t *testing.T) {
	if providerRowHasProviderReply(&model.JavapayTrxRow{}) {
		t.Fatalf("blank provider row must not count as provider reply")
	}

	msg := "status=1&message=Mhn tunggu trx sblmnya selesai: #ref DANA.0812, status Sedang Diproses"
	row := &model.JavapayTrxRow{
		Provider: "trionik",
		Status:   "pending",
		Pesan:    &msg,
	}
	if !providerRowHasProviderReply(row) {
		t.Fatalf("pending provider message must count as provider reply")
	}
}

func TestPickLatestProviderSuccessIncludesJavapay(t *testing.T) {
	jpMsg := "Sukses"
	ysMsg := "status=47&message=GAGAL"
	states := []providerState{
		{
			name: "javapay",
			row: &model.JavapayTrxRow{
				ID:         20,
				KodeRespon: &[]string{"20"}[0],
				Pesan:      &jpMsg,
				Harga:      int64Ptr(54075),
			},
		},
		{
			name: "yuscom",
			row: &model.JavapayTrxRow{
				ID:         10,
				KodeRespon: &[]string{"47"}[0],
				Pesan:      &ysMsg,
			},
		},
	}

	provider, row, rc, msg, price, _ := pickLatestProviderSuccess(states)
	if provider != "javapay" || row == nil {
		t.Fatalf("expected javapay success to be selected, got provider=%q row=%#v", provider, row)
	}
	if rc != "20" || msg != "Sukses" || price != 54075 {
		t.Fatalf("unexpected success selection payload rc=%q msg=%q price=%d", rc, msg, price)
	}
}

func TestPickLatestProviderPendingIncludesJavapay(t *testing.T) {
	jpPending := "Transaksi sedang diproses"
	states := []providerState{
		{
			name: "javapay",
			row: &model.JavapayTrxRow{
				ID:         30,
				KodeRespon: &[]string{"1"}[0],
				Pesan:      &jpPending,
			},
		},
	}

	provider, row, rc, msg := pickLatestProviderPending(states)
	if provider != "javapay" || row == nil {
		t.Fatalf("expected javapay pending to be selected, got provider=%q row=%#v", provider, row)
	}
	if rc != "1" || msg != jpPending {
		t.Fatalf("unexpected pending selection rc=%q msg=%q", rc, msg)
	}
}

func TestPickLatestProviderFailurePrefersLatestLoketBayarFallback(t *testing.T) {
	smbMsg := "REFF#smpay-test BIFASTOPEN2 535901260205405 GAGAL, KET: PRODUK GANGGUAN"
	loketMsg := "loketbayar fallback gagal http=0 setelah 12x retry"
	states := []providerState{
		{
			name: "smb",
			row: &model.JavapayTrxRow{
				ID:         1782590,
				Provider:   "smb",
				Status:     "failed",
				KodeRespon: &[]string{"2"}[0],
				Pesan:      &smbMsg,
			},
		},
		{
			name: "loketbayar",
			row: &model.JavapayTrxRow{
				ID:       1782594,
				Provider: "loketbayar",
				Status:   "failed",
				Pesan:    &loketMsg,
			},
		},
	}

	provider, row, retryable, _, msg, _, _ := pickLatestProviderFailure(states)
	if provider != "loketbayar" || row == nil || row.ID != 1782594 {
		t.Fatalf("expected latest loketbayar failure, got provider=%q row=%#v", provider, row)
	}
	if !retryable || msg != loketMsg {
		t.Fatalf("unexpected failure payload retryable=%v msg=%q", retryable, msg)
	}
}

func TestPickLatestProviderPendingTreatsSMBInquiryAsPending(t *testing.T) {
	smbMsg := "INQSUKSES TRANSFER KE REKENING:0851,BANK:DANA,NAMA:TEST,NOMINAL:49000,ADMIN:DANA/2500,JUMLAH:51500,HARGA:49.100,SISASALDO:35.112.764 - 0 = 35.112.764"
	states := []providerState{
		{
			name: "smb",
			row: &model.JavapayTrxRow{
				ID:         40,
				KodeProduk: "WALLET_PPOB:DANA",
				Perintah:   "PAY",
				KodeRespon: &[]string{"1"}[0],
				Pesan:      &smbMsg,
			},
		},
	}

	provider, row, rc, msg := pickLatestProviderPending(states)
	if provider != "smb" || row == nil {
		t.Fatalf("expected smb inquiry row to remain pending, got provider=%q row=%#v", provider, row)
	}
	if rc != "1" || msg != smbMsg {
		t.Fatalf("unexpected pending selection rc=%q msg=%q", rc, msg)
	}

	successProvider, successRow, _, _, _, _ := pickLatestProviderSuccess(states)
	if successProvider != "" || successRow != nil {
		t.Fatalf("did not expect smb inquiry row to be treated as success, got provider=%q row=%#v", successProvider, successRow)
	}
}

func TestIsStaleYuscomNoResponse(t *testing.T) {
	now := time.Now()
	row := &model.JavapayTrxRow{
		Provider:   "yuscom",
		DibuatPada: now.Add(-providerCallTimeout - time.Second),
	}
	if !isStaleYuscomNoResponse(row, now) {
		t.Fatalf("expected blank yuscom row past retry window to be stale")
	}

	recent := &model.JavapayTrxRow{
		Provider:   "yuscom",
		DibuatPada: now.Add(-5 * time.Second),
	}
	if isStaleYuscomNoResponse(recent, now) {
		t.Fatalf("did not expect fresh yuscom row to be stale")
	}

	msg := "akan diproses"
	answered := &model.JavapayTrxRow{
		Provider:   "yuscom",
		Pesan:      &msg,
		DibuatPada: now.Add(-providerCallTimeout - time.Second),
	}
	if isStaleYuscomNoResponse(answered, now) {
		t.Fatalf("did not expect answered yuscom row to be treated as stale no-response")
	}

	finalFailed := &model.JavapayTrxRow{
		Provider:   "yuscom",
		Status:     "failed",
		DibuatPada: now.Add(-providerCallTimeout - time.Second),
	}
	if isStaleYuscomNoResponse(finalFailed, now) {
		t.Fatalf("did not expect final failed yuscom row to be treated as stale no-response")
	}
}

func TestPickLatestStaleNoResponseSkipsFinalRows(t *testing.T) {
	now := time.Now()
	states := []providerState{
		{
			name: "talentapay",
			row: &model.JavapayTrxRow{
				ID:         10,
				Provider:   "talentapay",
				Status:     "failed",
				DibuatPada: now.Add(-providerCallTimeout - time.Second),
			},
		},
		{
			name: "yuscom",
			row: &model.JavapayTrxRow{
				ID:         11,
				Provider:   "yuscom",
				Status:     "success",
				DibuatPada: now.Add(-providerCallTimeout - time.Second),
			},
		},
	}

	provider, row := pickLatestStaleNoResponseProvider(states, now)
	if provider != "" || row != nil {
		t.Fatalf("did not expect final provider rows to be selected as stale, got provider=%q row=%#v", provider, row)
	}
}

func TestPickLatestStaleNoResponseSkipsYuscom(t *testing.T) {
	now := time.Now()
	states := []providerState{
		{
			name: "yuscom",
			row: &model.JavapayTrxRow{
				ID:         12,
				Provider:   "yuscom",
				DibuatPada: now.Add(-providerCallTimeout - time.Second),
			},
		},
	}

	provider, row := pickLatestStaleNoResponseProvider(states, now)
	if provider != "" || row != nil {
		t.Fatalf("yuscom no-response must be handled by its dedicated pending path, got provider=%q row=%#v", provider, row)
	}
}

func TestIsStaleSMBCheckOnlyPending(t *testing.T) {
	now := time.Now()
	msg := "INQSUKSES TRANSFER KE REKENING:0851,BANK:DANA,NAMA:TEST,NOMINAL:49000"
	row := &model.JavapayTrxRow{
		Provider:   "smb",
		DibuatPada: now.Add(-smbCheckCallbackTimeout - time.Second),
		KodeRespon: &[]string{"1"}[0],
		Pesan:      &msg,
	}
	if !isStaleSMBCheckOnlyPending(row, now) {
		t.Fatalf("expected old smb inquiry row to be stale")
	}

	fresh := *row
	fresh.DibuatPada = now.Add(-30 * time.Second)
	if isStaleSMBCheckOnlyPending(&fresh, now) {
		t.Fatalf("did not expect fresh smb inquiry row to be stale")
	}

	nonInquiryMsg := "BERHASIL"
	nonInquiry := *row
	nonInquiry.Pesan = &nonInquiryMsg
	if isStaleSMBCheckOnlyPending(&nonInquiry, now) {
		t.Fatalf("did not expect non-inquiry smb row to be stale")
	}

	unanswered := &model.JavapayTrxRow{
		Provider:   "smb",
		DibuatPada: now.Add(-smbCheckCallbackTimeout - time.Second),
	}
	if !isStaleSMBCheckOnlyPending(unanswered, now) {
		t.Fatalf("expected unanswered smb row to be stale")
	}

	unansweredFresh := *unanswered
	unansweredFresh.DibuatPada = now.Add(-2 * time.Second)
	if isStaleSMBCheckOnlyPending(&unansweredFresh, now) {
		t.Fatalf("did not expect fresh unanswered smb row to be stale")
	}

	checkPersistedAsCEK := &model.JavapayTrxRow{
		Provider:   "smb",
		DibuatPada: now.Add(-smbCheckCallbackTimeout - time.Second),
		KodeRespon: &[]string{"1"}[0],
		Pesan:      &[]string{"CEK"}[0],
		ResponMentah: map[string]any{
			"stage":   "check",
			"message": "INQSUKSES TRANSFER KE REKENING:0851,BANK:DANA,NAMA:TEST,NOMINAL:49000",
		},
	}
	if !isStaleSMBCheckOnlyPending(checkPersistedAsCEK, now) {
		t.Fatalf("expected CEK-persisted SMB callback check row to be stale")
	}
}

func TestSMBDirectShouldWaitFinalCallback(t *testing.T) {
	if !smbDirectShouldWaitFinalCallback(200, `{"success":true,"rc":"0068","msg":"Trx ELDN 082235804266 Under proses..."}`) {
		t.Fatalf("expected smb pending direct response to wait for callback final")
	}
	if smbDirectShouldWaitFinalCallback(200, `{"success":true,"rc":"1","msg":"PAYSUKSES TRANSAKSI BERHASIL"}`) {
		t.Fatalf("did not expect smb success direct response to wait for callback final")
	}
	if smbDirectShouldWaitFinalCallback(200, `{"success":false,"rc":"0061","msg":"cek balance dan transaksi ke CS/admin"}`) {
		t.Fatalf("did not expect smb reject response to wait for callback final")
	}
	if smbDirectShouldWaitFinalCallback(500, "timeout") {
		t.Fatalf("did not expect non-200 smb response to wait for callback final")
	}
	if smbDirectShouldWaitFinalCallback(200, "MAINTENANCE") {
		t.Fatalf("did not expect smb system issue response to be callback-owned")
	}
}

func TestBuildStaleSMBCheckFailureMessage(t *testing.T) {
	msg := "  INQSUKSES   TEST  "
	row := &model.JavapayTrxRow{Pesan: &msg}
	got := buildStaleSMBCheckFailureMessage(row)
	if !strings.Contains(got, "callback bayar tidak diterima") {
		t.Fatalf("unexpected failure message: %q", got)
	}
	if !strings.Contains(got, "INQSUKSES TEST") {
		t.Fatalf("expected normalized original message in failure message: %q", got)
	}

	cekMsg := "CEK"
	rowWithPayload := &model.JavapayTrxRow{
		Pesan: &cekMsg,
		ResponMentah: map[string]any{
			"stage":   "check",
			"message": "INQSUKSES TRANSFER KE REKENING:0851,BANK:DANA,NAMA:TEST",
		},
	}
	got = buildStaleSMBCheckFailureMessage(rowWithPayload)
	if !strings.Contains(got, "INQSUKSES TRANSFER KE REKENING") {
		t.Fatalf("expected failure message to use stored check callback payload: %q", got)
	}
}

func TestSMBProviderRowHasCheckCallback(t *testing.T) {
	row := &model.JavapayTrxRow{
		ResponMentah: map[string]any{
			"stage": "check",
		},
	}
	if !smbProviderRowHasCheckCallback(row) {
		t.Fatalf("expected callback stage check to be detected")
	}

	nonCheck := &model.JavapayTrxRow{
		ResponMentah: map[string]any{
			"stage": "pay_after_check_callback",
		},
	}
	if smbProviderRowHasCheckCallback(nonCheck) {
		t.Fatalf("did not expect pay stage to be treated as check callback")
	}
}

func TestSMBProviderRowHasAnyCallbackDetectsPayAfterCheck(t *testing.T) {
	row := &model.JavapayTrxRow{
		ResponMentah: map[string]any{
			"stage": "pay_after_check_callback",
		},
	}
	if !smbProviderRowHasAnyCallback(row) {
		t.Fatalf("expected pay-after-check stage to count as SMB callback activity")
	}
}
