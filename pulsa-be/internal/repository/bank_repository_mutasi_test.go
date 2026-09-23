package repository

import (
	"strings"
	"testing"
)

func TestBankMutasiListOrderByDefaultsToChronological(t *testing.T) {
	orderBy := bankMutasiListOrderBy(false, "admin_fee_expr")
	trimmed := strings.TrimSpace(orderBy)
	if !strings.HasPrefix(trimmed, "COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) DESC") {
		t.Fatalf("expected chronological order first, got %q", trimmed)
	}
	if strings.Contains(trimmed, "CASE") {
		t.Fatalf("default history order must not prioritize unassigned rows: %q", trimmed)
	}
}

func TestBankMutasiListOrderByCanPrioritizeUnassignedRows(t *testing.T) {
	orderBy := bankMutasiListOrderBy(true, "admin_fee_expr")
	if !strings.Contains(orderBy, "CASE") {
		t.Fatalf("expected prioritized order to include CASE expression: %q", orderBy)
	}
	if !strings.Contains(orderBy, "admin_fee_expr") {
		t.Fatalf("expected prioritized order to preserve admin fee exclusion: %q", orderBy)
	}
	if !strings.Contains(orderBy, "COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) DESC") {
		t.Fatalf("expected prioritized order to fall back to chronological sorting: %q", orderBy)
	}
}

func TestBankMutasiAdminFeeExprExcludesFixedBankFees(t *testing.T) {
	expr := bankMutasiAdminFeeExpr("mb")
	if !strings.Contains(expr, "mb.jumlah IN (2500, 6500)") {
		t.Fatalf("expected fixed 2500/6500 bank fees to be excluded: %q", expr)
	}
	if !strings.Contains(expr, "biaya potongan") || !strings.Contains(expr, "potongan rekening") {
		t.Fatalf("expected account deduction fees to be excluded: %q", expr)
	}
	if !strings.Contains(expr, "biaya adm") || !strings.Contains(expr, "adm bank") {
		t.Fatalf("expected abbreviated admin fee text to be excluded: %q", expr)
	}
}

func TestBankMutasiUnpairedDebitExcludedReasonExprSkipsNonTransferDebits(t *testing.T) {
	expr := bankMutasiUnpairedDebitExcludedReasonExpr("mb")
	for _, want := range []string{"BANK_RECONCILE_ADJUSTMENT", "INTERNAL_FINANCE_DEBIT", "BANK_ADJUST_DEBIT"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("expected unpaired debit exclusion to include %q: %q", want, expr)
		}
	}
}

func TestBankMutasiProviderPairExistsExprIncludesManualTopupAndProviderNamePairing(t *testing.T) {
	expr := bankMutasiProviderPairExistsExpr("mb")
	for _, want := range []string{"manual_provider_topup_ref", "matched_k24_ref", "provider_rekening", "nomor_rekening_digits", "regexp_replace(pr.nama"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("expected provider pair expression to include %q: %q", want, expr)
		}
	}
}

func TestBankMutasiProviderRefundPairExistsExprDetectsRefundedProviderDebits(t *testing.T) {
	expr := bankMutasiProviderRefundPairExistsExpr("mb")
	for _, want := range []string{"BANK_TRANSFER_PROVIDER_REFUND", "refund_of_bank_ref", "refund_of_bank_mutasi_id", "provider_credit_refunded_by_bank_ref"} {
		if !strings.Contains(expr, want) {
			t.Fatalf("expected provider refund pair expression to include %q: %q", want, expr)
		}
	}
}
