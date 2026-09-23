package service

import (
	"strings"
	"testing"
)

func validAgentCreditInput() AgentCreditSubmitInput {
	return AgentCreditSubmitInput{
		ApplicantData: map[string]any{
			"agent_name":      "Budi Santoso",
			"store_name":      "Konter Maju Jaya",
			"nik":             "3273011408950001",
			"whatsapp":        "081234567890",
			"email":           "budi@example.com",
			"home_address":    "Jalan Melati nomor 10",
			"store_address":   "Jalan Mawar nomor 20",
			"family_name":     "Siti Aminah",
			"family_relation": "Ibu kandung",
			"family_whatsapp": "081298765432",
		},
		AgentSignature: "data:image/png;base64," + strings.Repeat("YWJj", 40),
	}
}

func TestValidateAgentCreditSubmissionPassesAndStampsResult(t *testing.T) {
	in := validAgentCreditInput()
	if err := validateAgentCreditSubmission(&in, false); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if in.ApplicantData["system_validation_status"] != "passed" {
		t.Fatalf("validation status = %v", in.ApplicantData["system_validation_status"])
	}
}

func TestValidateAgentCreditSubmissionRejectsCarelessData(t *testing.T) {
	tests := []struct {
		name   string
		change func(*AgentCreditSubmitInput)
	}{
		{"placeholder name", func(in *AgentCreditSubmitInput) { in.ApplicantData["agent_name"] = "test" }},
		{"invalid nik", func(in *AgentCreditSubmitInput) { in.ApplicantData["nik"] = "1111111111111111" }},
		{"same family phone", func(in *AgentCreditSubmitInput) { in.ApplicantData["family_whatsapp"] = "081234567890" }},
		{"short address", func(in *AgentCreditSubmitInput) { in.ApplicantData["home_address"] = "jalan" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validAgentCreditInput()
			tt.change(&in)
			if err := validateAgentCreditSubmission(&in, false); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateAgentCreditDocumentsRejectsDuplicatePhotos(t *testing.T) {
	image := "data:image/jpeg;base64," + strings.Repeat("YWJj", 700)
	in := validAgentCreditInput()
	in.DocumentData = map[string]any{
		"ktp":              map[string]any{"data_url": image},
		"store":            map[string]any{"data_url": image},
		"selfie_ktp":       map[string]any{"data_url": image},
		"selfie_marketing": map[string]any{"data_url": image},
	}
	if err := validateAgentCreditSubmission(&in, true); err == nil {
		t.Fatal("expected duplicate photo validation error")
	}
}
