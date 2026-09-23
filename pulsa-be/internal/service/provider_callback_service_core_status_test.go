package service

import "testing"

func TestShouldKeepExistingMemberFinalStatusKeepsFailedOnLateSuccess(t *testing.T) {
	if !shouldKeepExistingMemberFinalStatus("failed", "success") {
		t.Fatalf("failed member transactions must stay final on late success callbacks")
	}
	if !shouldKeepExistingMemberFinalStatus("success", "failed") {
		t.Fatalf("success member transactions must stay final on late failed callbacks")
	}
	if shouldKeepExistingMemberFinalStatus("pending", "success") {
		t.Fatalf("pending member transactions should still accept final callbacks")
	}
}
