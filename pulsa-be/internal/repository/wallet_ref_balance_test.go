package repository

import "testing"

func TestClampWalletRefundAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		requested   int64
		outstanding int64
		want        int64
	}{
		{name: "zero requested", requested: 0, outstanding: 100, want: 0},
		{name: "zero outstanding", requested: 100, outstanding: 0, want: 0},
		{name: "negative outstanding", requested: 100, outstanding: -50, want: 0},
		{name: "refund all outstanding", requested: 100, outstanding: 100, want: 100},
		{name: "refund clamped to outstanding", requested: 100, outstanding: 40, want: 40},
		{name: "refund below outstanding", requested: 60, outstanding: 100, want: 60},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := clampWalletRefundAmount(tt.requested, tt.outstanding); got != tt.want {
				t.Fatalf("clampWalletRefundAmount(%d, %d) = %d, want %d", tt.requested, tt.outstanding, got, tt.want)
			}
		})
	}
}

func TestComputeWalletHoldDebitAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		target      int64
		outstanding int64
		want        int64
	}{
		{name: "zero target", target: 0, outstanding: 0, want: 0},
		{name: "no outstanding", target: 100, outstanding: 0, want: 100},
		{name: "partial outstanding", target: 100, outstanding: 40, want: 60},
		{name: "enough outstanding", target: 100, outstanding: 100, want: 0},
		{name: "over outstanding", target: 100, outstanding: 140, want: 0},
		{name: "negative outstanding treated as zero", target: 100, outstanding: -20, want: 100},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ComputeWalletHoldDebitAmount(tt.target, tt.outstanding); got != tt.want {
				t.Fatalf("ComputeWalletHoldDebitAmount(%d, %d) = %d, want %d", tt.target, tt.outstanding, got, tt.want)
			}
		})
	}
}
