package service

import (
	"strings"
	"testing"
)

func TestValidateMidtransServerKeySandbox(t *testing.T) {
	if err := validateMidtransServerKey("SB-Mid-server-real-sandbox-key", false); err != nil {
		t.Fatalf("sandbox key should be valid in sandbox mode: %v", err)
	}

	if err := validateMidtransServerKey("Mid-server-key-from-sandbox-dashboard", false); err != nil {
		t.Fatalf("sandbox dashboard key without SB prefix should be valid in sandbox mode: %v", err)
	}
}

func TestValidateMidtransServerKeyProduction(t *testing.T) {
	if err := validateMidtransServerKey("Mid-server-real-production-key", true); err != nil {
		t.Fatalf("production key should be valid in production mode: %v", err)
	}

	err := validateMidtransServerKey("SB-Mid-server-real-sandbox-key", true)
	if err == nil || !strings.Contains(err.Error(), "sandbox") {
		t.Fatalf("sandbox key in production mode error = %v, want sandbox mismatch", err)
	}
}

func TestValidateMidtransServerKeyRejectsPlaceholder(t *testing.T) {
	err := validateMidtransServerKey("SB-Mid-server-GANTI_DENGAN_SERVER_KEY", false)
	if err == nil || !strings.Contains(err.Error(), "placeholder") {
		t.Fatalf("placeholder key error = %v, want placeholder rejection", err)
	}
}

func TestValidateMidtransServerKeyRejectsUnknownFormat(t *testing.T) {
	err := validateMidtransServerKey("server-key-without-midtrans-prefix", false)
	if err == nil || !strings.Contains(err.Error(), "Settings > Access Keys") {
		t.Fatalf("unknown sandbox key error = %v, want access key format rejection", err)
	}
}
