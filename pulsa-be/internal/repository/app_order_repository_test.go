package repository

import "testing"

func TestShouldApplyAppOrderStatusTransition(t *testing.T) {
	tests := []struct {
		current string
		next    string
		want    bool
	}{
		{current: "pending_payment", next: "paid", want: true},
		{current: "pending_payment", next: "expired", want: true},
		{current: "paid", next: "paid", want: true},
		{current: "paid", next: "failed", want: true},
		{current: "paid", next: "processing_provider", want: true},
		{current: "processing_provider", next: "paid", want: false},
		{current: "processing_provider", next: "failed", want: true},
		{current: "processing_provider", next: "success", want: true},
		{current: "success", next: "paid", want: false},
		{current: "success", next: "cancelled", want: false},
		{current: "success", next: "refunded", want: true},
		{current: "refunded", next: "paid", want: false},
		{current: "failed", next: "paid", want: false},
		{current: "failed", next: "refunded", want: true},
	}

	for _, tc := range tests {
		if got := shouldApplyAppOrderStatusTransition(tc.current, tc.next); got != tc.want {
			t.Fatalf("unexpected transition current=%q next=%q got=%t want=%t", tc.current, tc.next, got, tc.want)
		}
	}
}
