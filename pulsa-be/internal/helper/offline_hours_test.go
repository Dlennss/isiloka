package helper

import "testing"

func TestH2HProductUsesOfflineHoursOnlyForOVO(t *testing.T) {
	tests := []struct {
		product string
		want    bool
	}{
		{product: "OVO", want: true},
		{product: "ovo50", want: true},
		{product: " SUPERBANK ", want: false},
		{product: "SHOPEE", want: false},
		{product: "DANA", want: false},
		{product: "GOPAY", want: false},
		{product: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.product, func(t *testing.T) {
			if got := H2HProductUsesOfflineHours(tt.product); got != tt.want {
				t.Fatalf("H2HProductUsesOfflineHours(%q) = %v, want %v", tt.product, got, tt.want)
			}
		})
	}
}

func TestIsH2HProductAvailableForNowBypassesNonOVO(t *testing.T) {
	if !IsH2HProductAvailableForNow("SUPERBANK", "", "") {
		t.Fatalf("non-OVO H2H products must bypass offline hours")
	}
	if !IsH2HProductAvailableForNow("OVO", "00:00:00", "23:59:00") {
		t.Fatalf("OVO H2H must still respect its configured online window")
	}
}
