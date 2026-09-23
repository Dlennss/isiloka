package repository

import (
	"testing"
	"time"
)

func TestBankMutationSemanticKeyIncludesCounterparty(t *testing.T) {
	noteA := "Pengirim: A | auto browser scrape | date: 20 Mei 2026 17:46:21 | desc: BFST00001 A | amount: 30.043.419 | direction: CR | balance: 100"
	noteB := "Pengirim: B | auto browser scrape | date: 20 Mei 2026 17:46:21 | desc: TRF99999 B | amount: 30.043.419 | direction: CR | balance: 200"

	keyA, legacyA, okA := bankMutationSemanticKeys(27, 30043419, "credit", noteA, nil, "", "")
	keyB, legacyB, okB := bankMutationSemanticKeys(27, 30043419, "credit", noteB, nil, "", "")
	if !okA || !okB {
		t.Fatalf("expected semantic keys to parse")
	}
	if keyA == keyB {
		t.Fatalf("different counterparty should have different semantic keys: %q", keyA)
	}
	if legacyA != legacyB {
		t.Fatalf("legacy semantic key should still ignore counterparty for fallback: %q != %q", legacyA, legacyB)
	}
}

func TestBankMutationSemanticKeySeparatesDifferentTimes(t *testing.T) {
	noteA := "date: 20 Mei 2026 17:46:21 | desc: A"
	noteB := "date: 20 Mei 2026 17:46:22 | desc: A"

	keyA, okA := bankMutationSemanticKey(27, 30043419, "credit", noteA)
	keyB, okB := bankMutationSemanticKey(27, 30043419, "credit", noteB)
	if !okA || !okB {
		t.Fatalf("expected semantic keys to parse")
	}
	if keyA == keyB {
		t.Fatalf("different transaction seconds should have different keys: %q", keyA)
	}
}

func TestBankMutationLegacyDuplicateAllowsSameScrapedBalance(t *testing.T) {
	if !bankMutationLegacyDuplicateAllowed(true, true, 1695939, 1695939) {
		t.Fatalf("same scraped running balance should allow legacy duplicate match even when party text changed")
	}
}

func TestBankMutationLegacyDuplicateRejectsDifferentPartyWithoutBalanceMatch(t *testing.T) {
	if bankMutationLegacyDuplicateAllowed(true, true, 1695939, 1695940) {
		t.Fatalf("different party text with different running balance must not fallback to legacy duplicate")
	}
	if bankMutationLegacyDuplicateAllowed(true, true, 1695939, 0) {
		t.Fatalf("different party text without scraped running balance must not fallback to legacy duplicate")
	}
}

func TestBankMutationLegacyDuplicateAllowsMissingPartyFallback(t *testing.T) {
	if !bankMutationLegacyDuplicateAllowed(false, true, 1, 0) {
		t.Fatalf("legacy fallback should still work when candidate lacks party text")
	}
	if !bankMutationLegacyDuplicateAllowed(true, false, 1, 0) {
		t.Fatalf("legacy fallback should still work when input lacks party text")
	}
}

func TestBankScrapedBalanceBeforeAfterUsesBankBalance(t *testing.T) {
	before, after, ok := bankScrapedBalanceBeforeAfter("credit", 25000, 125000)
	if !ok || before != 100000 || after != 125000 {
		t.Fatalf("credit scraped balance = before %d after %d ok %v", before, after, ok)
	}

	before, after, ok = bankScrapedBalanceBeforeAfter("debit", 25000, 75000)
	if !ok || before != 100000 || after != 75000 {
		t.Fatalf("debit scraped balance = before %d after %d ok %v", before, after, ok)
	}

	if _, _, ok = bankScrapedBalanceBeforeAfter("credit", 25000, 0); ok {
		t.Fatalf("zero actual balance must not be adopted")
	}
}

func TestBankLedgerApplyMutationRecomputesForwardBalance(t *testing.T) {
	after, ok := bankLedgerApplyMutation(1000, "CREDIT", 9000)
	if !ok || after != 10000 {
		t.Fatalf("credit forward recompute = after %d ok %v", after, ok)
	}
	after, ok = bankLedgerApplyMutation(after, "DEBIT", 1100)
	if !ok || after != 8900 {
		t.Fatalf("debit forward recompute = after %d ok %v", after, ok)
	}
}

func TestBankLedgerRecomputeBeforeAfterPreservesScrapedBalance(t *testing.T) {
	row := bankLedgerRecomputeRow{
		arah:                "CREDIT",
		jumlah:              25022554,
		scrapedBalanceAfter: 265594854,
	}
	before, after, ok := bankLedgerRecomputeBeforeAfter(310646730, row)
	if !ok {
		t.Fatalf("expected scraped row to recompute")
	}
	if before != 240572300 || after != 265594854 {
		t.Fatalf("scraped row = before %d after %d, want before 240572300 after 265594854", before, after)
	}
}

func TestBankLedgerRecomputeOrdersSameTimestampByScrapedBalance(t *testing.T) {
	mutationTime := time.Date(2026, 6, 14, 11, 34, 24, 0, time.FixedZone("WIB", 7*60*60))
	rows := []bankLedgerRecomputeRow{
		{
			id:                  97683,
			arah:                "DEBIT",
			jumlah:              99910501,
			mutationTime:        mutationTime,
			scrapedBalanceAfter: 1335232969,
		},
		{
			id:                  97566,
			arah:                "DEBIT",
			jumlah:              99810511,
			mutationTime:        mutationTime,
			scrapedBalanceAfter: 543178079,
		},
		{
			id:                  97689,
			arah:                "DEBIT",
			jumlah:              99610531,
			mutationTime:        mutationTime,
			scrapedBalanceAfter: 742399141,
		},
		{
			id:                  97690,
			arah:                "DEBIT",
			jumlah:              99410551,
			mutationTime:        mutationTime,
			scrapedBalanceAfter: 642988590,
		},
	}

	ordered := orderBankLedgerRecomputeRows(842009672, rows)
	got := make([]int64, 0, len(ordered))
	for _, row := range ordered {
		got = append(got, row.id)
	}
	want := []int64{97689, 97690, 97566}
	if len(got) != len(want) {
		t.Fatalf("ordered ids = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ordered ids = %v, want %v", got, want)
		}
	}
}

func TestBankAdminFeeMutationDetectsBCAProviderFee(t *testing.T) {
	note := "Pengirim: SWITCHING DB BIAYA TXN KE 009 SIMBA PRADANA SOLU KBB | auto browser scrape | date: 21/05/2026 | desc: SWITCHING DB BIAYA TXN KE 009 SIMBA PRADANA SOLU KBB | amount: 6,500.00 | direction: DB | rekening_provider: 017201002852561 | provider: smb"
	if !isBankAdminFeeMutation("BANK_MANUAL_OUT", note) {
		t.Fatalf("expected BCA BIAYA TXN note to be detected as bank admin fee")
	}
	if !isScrapedProviderTransferAdminFee(6500, "debit", "BANK_MANUAL_OUT", note, "", "") {
		t.Fatalf("expected scraped BCA 6500 debit to be treated as provider transfer admin fee")
	}
}

func TestScrapedProviderTransferAdminFeeRequiresDebit6500(t *testing.T) {
	note := "Pengirim: SWITCHING DB BIAYA TXN KE 009 SIMBA PRADANA SOLU KBB | rekening_provider: 017201002852561"
	if isScrapedProviderTransferAdminFee(6501, "debit", "BANK_MANUAL_OUT", note, "", "") {
		t.Fatalf("non-6500 amount must not be treated as provider transfer admin fee")
	}
	if isScrapedProviderTransferAdminFee(6500, "credit", "BANK_MANUAL_IN", note, "", "") {
		t.Fatalf("credit mutation must not be treated as provider transfer admin fee")
	}
}

func TestScrapedProviderTransferAdminFeeDetectsBRIProviderFee(t *testing.T) {
	note := "Pengirim: ATMSTRPRM 08888 000063518 4294123999 | auto browser scrape | date: 22 Mei 2026 05:07:36 | desc: ATMSTRPRM 08888 000063518 4294123999 | amount: - Rp6.500 | direction: debit"
	if !isScrapedProviderTransferAdminFee(6500, "debit", "BANK_MANUAL_OUT", note, "", "") {
		t.Fatalf("expected scraped BRI ATMSTRPRM 6500 debit to be treated as provider transfer admin fee")
	}
}

func TestScrapedProviderTransferAdminFeeDetectsBIFast2500Fee(t *testing.T) {
	note := "Pengirim: 20260601BMRIIDJA010O9938671293 BIFAST Out CS-GL | auto browser scrape"
	if !isScrapedProviderTransferAdminFee(2500, "debit", "BANK_MANUAL_OUT", note, "", "") {
		t.Fatalf("expected scraped BI-Fast 2500 debit to be treated as bank admin fee")
	}
}

func TestBankMutationSemanticKeySeparatesDateOnlyFeeByBalance(t *testing.T) {
	noteA := "auto browser scrape | date: 05/06/2026 | desc: SWITCHING DB BIAYA TXN KE 009 SIMBA PRADANA SOLU KBB | amount: 6,500.00 | direction: DB | balance: 2,909,937,001.00"
	noteB := "auto browser scrape | date: 05/06/2026 | desc: SWITCHING DB BIAYA TXN KE 009 SIMBA PRADANA SOLU KBB | amount: 6,500.00 | direction: DB | balance: 2,862,929,505.00"
	keyA, legacyA, okA := bankMutationSemanticKeys(1, 6500, "debit", noteA, nil, "", "")
	keyB, legacyB, okB := bankMutationSemanticKeys(1, 6500, "debit", noteB, nil, "", "")
	if !okA || !okB {
		t.Fatalf("expected semantic keys for date-only fee rows")
	}
	if keyA == keyB {
		t.Fatalf("date-only fee rows with different balances must not share semantic key")
	}
	if legacyA != legacyB {
		t.Fatalf("legacy semantic key should stay available for exact older fallback")
	}
}

func TestScrapedBCAOperationalTransferDetected(t *testing.T) {
	note := "Pengirim: BFST3432738881 IBIZ:CENAIDJA | auto browser scrape | date: 14 Mei 2026 15:54:37 | desc: BFST3432738881 IBIZ:CENAIDJA | amount: - Rp93.000.000 | direction: debit"
	if !isScrapedBCAOperationalTransfer(93000000, "debit", note, "", "") {
		t.Fatalf("expected transfer to BCA OPERASIONAL 3432738881 to be detected")
	}
	if isScrapedBCAOperationalTransfer(2500, "debit", note, "", "") {
		t.Fatalf("BI-Fast fee must not be credited to BCA OPERASIONAL")
	}
	if !isScrapedBCAOperationalTransferAdminFee(2500, "debit", note, "", "") {
		t.Fatalf("expected BI-Fast 2500 row to be detected as transfer admin fee")
	}
}

func TestScrapedBCAOperationalTransferDetectsBCANameOnlyScrape(t *testing.T) {
	note := "Pengirim: TRSF E-BANKING DB 2305/FTSCY/WS95051 95000000.00 PULSA MITRA NASION | auto browser scrape | date: 23/05/2026 | desc: TRSF E-BANKING DB 2305/FTSCY/WS95051 95000000.00 PULSA MITRA NASION | amount: 95,000,000.00 | direction: DB"
	if !isScrapedBCAOperationalTransfer(95000000, "debit", note, "", "") {
		t.Fatalf("expected BCA e-banking transfer to PULSA MITRA NASION to be detected as BCA OPERASIONAL")
	}
}

func TestScrapedBCAOperationalTransferDetectsMandiriBIFastBCAName(t *testing.T) {
	note := "Pengirim: 202606061953642064 BIFAST Out CS-GL CENAIDJA/PULSA MITRA NASIONAL PT ref: 202606061953642064 | auto browser scrape | date: 06 Jun 2026 19:55:44 | desc: 202606061953642064 BIFAST Out CS-GL CENAIDJA/PULSA MITRA NASIONAL PT ref: 202606061953642064 | amount: 85,000,000.00 | direction: debit"
	if !isScrapedBCAOperationalTransfer(85000000, "debit", note, "", "") {
		t.Fatalf("expected Mandiri BI-Fast CENAIDJA/PULSA MITRA NASIONAL PT to be detected as BCA OPERASIONAL")
	}
}

func TestScrapedBCAOperationalTransferRequiresExactAccount(t *testing.T) {
	note := "Pengirim: PINDAH BUKA Transfer RTGS PT PULSA MITRA NASIONAL PT. BANK RAKYAT INDONESIA"
	if isScrapedBCAOperationalTransfer(1559000000, "debit", note, "", "") {
		t.Fatalf("PT PULSA MITRA NASIONAL without 3432738881 must not be treated as BCA OPERASIONAL")
	}
}

func TestScrapedInternalBankTransferAdminFee(t *testing.T) {
	destination := &internalBankDestination{account: "8761518267"}
	note := "Pengirim: BI-FAST DB KE 8761518267 LISA OKTARIA | amount: 2,500.00 | direction: DB"
	if !isScrapedInternalBankTransferAdminFee(2500, "debit", destination, note, "", "") {
		t.Fatalf("expected BI-Fast fee to internal account to be classified as admin fee")
	}
	if isScrapedInternalBankTransferAdminFee(1000000, "debit", destination, note, "", "") {
		t.Fatalf("main transfer amount must not be classified as admin fee")
	}
	if isScrapedInternalBankTransferAdminFee(2500, "credit", destination, note, "", "") {
		t.Fatalf("credit row must not be classified as internal transfer admin fee")
	}
}

func TestInternalBankExactReceiverAllowsOneWordOwner(t *testing.T) {
	note := "Pengirim: 202606250836561920 BIFAST Out CS-GL CENAIDJA/MARWAN ref: 202606250836561920 | auto browser scrape | date: 25 Jun 2026 08:40:07 | desc: 202606250836561920 BIFAST Out CS-GL CENAIDJA/MARWAN ref: 202606250836561920 | amount: 15,000,000.00 | direction: debit"
	exactReceiver := bankInternalExactReceiverNameFromScrapedDebit(note, "", "")
	if exactReceiver != "MARWAN" {
		t.Fatalf("exact receiver = %q, want MARWAN", exactReceiver)
	}
	if bankInternalBankOwnerExactReceiverMatchScore("Marwan", exactReceiver) == 0 {
		t.Fatalf("expected exact BI-Fast receiver to match one-word internal bank owner")
	}
}

func TestInternalBankOneWordOwnerDoesNotMatchFreeText(t *testing.T) {
	text := bankNormalizeProviderName("Pengirim: CATATAN MANUAL MARWAN TANPA RECEIVER BANK")
	if bankInternalBankOwnerNameMatchScore("Marwan", text) != 0 {
		t.Fatalf("one-word owner must not match generic free text")
	}
	if bankInternalBankOwnerExactReceiverMatchScore("Marwan", "") != 0 {
		t.Fatalf("one-word owner must require exact receiver evidence")
	}
}

func TestScrapedProviderRefundCreditDetectsBNIKOR(t *testing.T) {
	note := "TRF/KOR/PEMINDAHAN DARI 0469000089 O0136595041 PT MULYO TRONIK INDONESIA"
	if !isScrapedProviderRefundCredit(98500391, "credit", note, "", "") {
		t.Fatalf("expected BNI KOR credit to be detected as provider transfer refund")
	}
	if isScrapedProviderRefundCredit(98500391, "debit", note, "", "") {
		t.Fatalf("debit row must not be detected as provider refund credit")
	}
}

func TestBankReferenceTokensRequireSharedTransactionToken(t *testing.T) {
	refund := bankReferenceTokens("TRF/KOR/PEMINDAHAN DARI 0469000089 O0136595041")
	debit := bankReferenceTokens("PEMINDAHAN KE 0469000089 O0136595041")
	other := bankReferenceTokens("PEMINDAHAN KE 0469000089 O9999999999")
	if !bankReferenceTokensIntersect(refund, debit) {
		t.Fatalf("expected shared O0136595041 token to match")
	}
	if bankReferenceTokensIntersect(refund, other) {
		t.Fatalf("different reference token must not match only because account appears")
	}
}

func TestProviderNameMatchDetectsChytronBRINameOnly(t *testing.T) {
	text := bankNormalizeProviderName("Pengirim: IBIZ PULSA MITRA NA TO AGUSTINUS SITUMOR | auto browser scrape")
	if bankProviderNameMatchScore("AGUSTINUS SITUMOR", text) == 0 {
		t.Fatalf("expected exact provider account name to match name-only BRI mutation")
	}
}

func TestProviderNameMatchAllowsBankTruncatedLongerName(t *testing.T) {
	text := bankNormalizeProviderName("Pengirim: IBIZ PULSA MITRA NA TO AGUSTINUS SITUMOR")
	if bankProviderNameMatchScore("AGUSTINUS SITUMORANG", text) == 0 {
		t.Fatalf("expected truncated bank text to match longer provider account name")
	}
}

func TestProviderNameMatchAllowsLegalPrefixAndLastTokenTruncation(t *testing.T) {
	text := bankNormalizeProviderName("TRSF E-BANKING DB 0906/FTSCY/WS95051 99400773.00 AGEN RETAIL DIGITA")
	if bankProviderNameMatchScore("PT AGEN RETAIL DIGITAL", text) == 0 {
		t.Fatalf("expected bank text without PT and truncated final token to match provider account name")
	}
}

func TestProviderNameMatchAllowsFiveCharFinalTokenTruncation(t *testing.T) {
	text := bankNormalizeProviderName("TRSF E-BANKING DB 0707/FTSCY/WS95051 99004512.00 LOKET BILLER TECHN")
	if bankProviderNameMatchScore("LOKET BILLER TECHNOLOGY", text) == 0 {
		t.Fatalf("expected five-character final token truncation to match provider account name")
	}
}

func TestProviderNameMatchAllowsMissingLongLegalSuffix(t *testing.T) {
	text := bankNormalizeProviderName("IBIZ PULSA MITRA NA TO PT CHYKA DIGITAL ESB:IBIZ:0001500F:108097438352")
	if bankProviderNameMatchScore("PT CHYKA DIGITAL NUSANTARA", text) == 0 {
		t.Fatalf("expected bank text missing long trailing legal suffix to match provider account name")
	}
}

func TestProviderNameMatchAllowsSpaceAndPunctuationDifferences(t *testing.T) {
	text := bankNormalizeProviderName("MCM InhouseTrf KE TALENTA PAY Transfer Fee Rp0")
	if bankProviderNameMatchScore("CV.TALENTAPAY", text) == 0 {
		t.Fatalf("expected TALENTA PAY bank text to match CV.TALENTAPAY provider account name")
	}
}

func TestProviderNameMatchRejectsTooShortTruncation(t *testing.T) {
	text := bankNormalizeProviderName("Pengirim: IBIZ PULSA MITRA NA TO AGUSTINUS SITU")
	if bankProviderNameMatchScore("AGUSTINUS SITUMORANG", text) != 0 {
		t.Fatalf("very short final-token prefix must not match provider account name")
	}
}

func TestInternalBankOwnerNameMatchDetectsUniqueOwnerName(t *testing.T) {
	text := bankNormalizeProviderName("TRF/PAY/TOP-UP ECHANNEL | KARTU 6010047890641625 LISA OKTARIA JAKARTA JK | BNI DIRECT 575441")
	if bankInternalBankOwnerNameMatchScore("LISA OKTARIA", text) == 0 {
		t.Fatalf("expected internal bank owner name to match scraped mutation text")
	}
}

func TestInternalBankOwnerNameMatchRejectsGenericNames(t *testing.T) {
	text := bankNormalizeProviderName("TRSF E-BANKING DB PULSA MITRA NASIONAL")
	if bankInternalBankOwnerNameMatchScore("PT PULSA MITRA NASIONAL", text) != 0 {
		t.Fatalf("generic internal bank owner name must not auto-match")
	}
}
