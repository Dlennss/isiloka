package service

import (
	"errors"
	"testing"

	"pulsa2/internal/repository"
)

func TestRetryPendingNoFallbackProviderLeft(t *testing.T) {
	if !retryPendingNoFallbackProviderLeft(errors.New("tidak ada provider fallback tersisa")) {
		t.Fatal("expected no fallback provider left")
	}
	if !retryPendingNoFallbackProviderLeft(errors.New("tidak ada provider eligible")) {
		t.Fatal("expected no eligible provider to count as no fallback provider left")
	}
	if retryPendingNoFallbackProviderLeft(errors.New("provider smb masih pending untuk refid ini")) {
		t.Fatal("pending/success guard must not count as no fallback provider left")
	}
	if retryPendingNoFallbackProviderLeft(nil) {
		t.Fatal("nil error must not count as no fallback provider left")
	}
}

func TestMergeRetryPendingRowsDedupesMemberRows(t *testing.T) {
	primary := []*repository.TrxMemberFull{{ID: 10, RefID: "a"}, {ID: 20, RefID: "b"}}
	extra := []*repository.TrxMemberFull{{ID: 20, RefID: "b-duplicate"}, {ID: 30, RefID: "c"}}

	got := mergeRetryPendingRows(primary, extra)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].ID != 10 || got[1].ID != 20 || got[2].ID != 30 {
		t.Fatalf("unexpected order/ids: %#v", got)
	}
}
