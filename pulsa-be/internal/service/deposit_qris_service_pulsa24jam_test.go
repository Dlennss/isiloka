package service

import "testing"

func TestNormalizePulsa24JamDepositStatus(t *testing.T) {
	cases := map[string]string{
		"settlement": "approved",
		"paid":       "approved",
		"expired":    "rejected",
		"pending":    "pending",
	}
	for input, want := range cases {
		if got := normalizePulsa24JamDepositStatus(input); got != want {
			t.Fatalf("normalize(%q) = %q, want %q", input, got, want)
		}
	}
}
