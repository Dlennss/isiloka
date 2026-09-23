package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	creditDigitsOnly = regexp.MustCompile(`\D`)
	creditEmail      = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	creditLetters    = regexp.MustCompile(`[A-Za-z]`)
)

var invalidCreditText = map[string]struct{}{
	"-": {}, "0": {}, "123": {}, "aaa": {}, "asal": {}, "coba": {},
	"dummy": {}, "qwerty": {}, "test": {}, "testing": {}, "tidak ada": {},
}

func validateAgentCreditSubmission(in *AgentCreditSubmitInput, validateDocuments bool) error {
	if in == nil {
		return errors.New("data pengajuan kredit tidak valid")
	}
	if in.ApplicantData == nil {
		in.ApplicantData = map[string]any{}
	}

	requiredText := []struct {
		key     string
		label   string
		minLen  int
		letters int
	}{
		{"agent_name", "nama agent", 3, 2},
		{"store_name", "nama toko", 3, 2},
		{"home_address", "alamat rumah", 8, 4},
		{"store_address", "alamat toko", 8, 4},
		{"family_name", "nama kontak keluarga", 3, 2},
		{"family_relation", "hubungan keluarga", 3, 2},
	}
	for _, field := range requiredText {
		value := cleanCreditText(in.ApplicantData[field.key])
		if err := validateCreditText(value, field.label, field.minLen, field.letters); err != nil {
			return err
		}
		in.ApplicantData[field.key] = value
	}

	nik := creditDigitsOnly.ReplaceAllString(cleanCreditText(in.ApplicantData["nik"]), "")
	if len(nik) != 16 || repeatedCreditValue(nik) || sequentialCreditDigits(nik) {
		return errors.New("NIK tidak valid; masukkan 16 angka sesuai KTP")
	}
	in.ApplicantData["nik"] = nik

	whatsapp, err := normalizeCreditPhone(cleanCreditText(in.ApplicantData["whatsapp"]))
	if err != nil {
		return fmt.Errorf("nomor WhatsApp agent %w", err)
	}
	familyWhatsapp, err := normalizeCreditPhone(cleanCreditText(in.ApplicantData["family_whatsapp"]))
	if err != nil {
		return fmt.Errorf("nomor WhatsApp keluarga %w", err)
	}
	if whatsapp == familyWhatsapp {
		return errors.New("nomor WhatsApp keluarga harus berbeda dari nomor agent")
	}
	in.ApplicantData["whatsapp"] = whatsapp
	in.ApplicantData["family_whatsapp"] = familyWhatsapp

	if email := strings.ToLower(cleanCreditText(in.ApplicantData["email"])); email != "" {
		if !creditEmail.MatchString(email) {
			return errors.New("format email tidak valid")
		}
		in.ApplicantData["email"] = email
	}

	if strings.TrimSpace(in.AgentSignature) != "" {
		if err := validateCreditImageData(in.AgentSignature, "tanda tangan", 100); err != nil {
			return err
		}
	}
	if validateDocuments {
		if err := validateAgentCreditDocuments(in.DocumentData); err != nil {
			return err
		}
		stampAgentCreditDocumentValidation(in.ApplicantData)
	}

	in.ApplicantData["system_validation_status"] = "passed"
	in.ApplicantData["system_validation_checked_at"] = time.Now().UTC().Format(time.RFC3339)
	in.ApplicantData["system_validation_checks"] = []string{
		"identity_format", "contact_format", "address_completeness", "signature_format",
	}
	return nil
}

func stampAgentCreditDocumentValidation(applicantData map[string]any) {
	applicantData["system_document_validation_status"] = "passed"
	applicantData["system_document_validation_checked_at"] = time.Now().UTC().Format(time.RFC3339)
}

func validateAgentCreditDocuments(documents map[string]any) error {
	required := []struct{ key, label string }{
		{"ktp", "foto KTP"},
		{"store", "foto toko"},
		{"selfie_ktp", "dokumen formulir"},
		{"selfie_marketing", "selfie bersama marketing"},
	}
	seen := make(map[string]string, len(required))
	for _, field := range required {
		raw, ok := documents[field.key].(map[string]any)
		if !ok {
			return fmt.Errorf("%s wajib diunggah oleh agent", field.label)
		}
		dataURL := cleanCreditText(raw["data_url"])
		if err := validateCreditImageData(dataURL, field.label, 2048); err != nil {
			return err
		}
		fingerprint := dataURL
		if previous, exists := seen[fingerprint]; exists {
			return fmt.Errorf("%s tidak boleh memakai foto yang sama dengan %s", field.label, previous)
		}
		seen[fingerprint] = field.label
	}
	return nil
}

func validateCreditImageData(value, label string, minimumBytes int) error {
	value = strings.TrimSpace(value)
	comma := strings.IndexByte(value, ',')
	if !strings.HasPrefix(strings.ToLower(value), "data:image/") || comma < 0 {
		return fmt.Errorf("%s harus berupa gambar yang valid", label)
	}
	decoded, err := base64.StdEncoding.DecodeString(value[comma+1:])
	if err != nil || len(decoded) < minimumBytes {
		return fmt.Errorf("%s kosong atau kualitasnya terlalu rendah", label)
	}
	return nil
}

func validateCreditText(value, label string, minimumLength, minimumLetters int) error {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if len([]rune(normalized)) < minimumLength {
		return fmt.Errorf("%s terlalu pendek", label)
	}
	if _, invalid := invalidCreditText[normalized]; invalid || repeatedCreditValue(normalized) {
		return fmt.Errorf("%s terlihat tidak valid; isi sesuai data sebenarnya", label)
	}
	if len(creditLetters.FindAllString(normalized, -1)) < minimumLetters {
		return fmt.Errorf("%s harus berisi data yang jelas", label)
	}
	return nil
}

func cleanCreditText(value any) string {
	text, _ := value.(string)
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func normalizeCreditPhone(value string) (string, error) {
	digits := creditDigitsOnly.ReplaceAllString(value, "")
	if strings.HasPrefix(digits, "62") {
		digits = "0" + strings.TrimPrefix(digits, "62")
	}
	if len(digits) < 10 || len(digits) > 14 || !strings.HasPrefix(digits, "08") || repeatedCreditValue(digits) {
		return "", errors.New("tidak valid")
	}
	return digits, nil
}

func repeatedCreditValue(value string) bool {
	runes := []rune(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	if len(runes) < 3 {
		return false
	}
	for _, size := range []int{1, 2, 3} {
		if len(runes)%size != 0 {
			continue
		}
		pattern := string(runes[:size])
		if strings.Repeat(pattern, len(runes)/size) == string(runes) {
			return true
		}
	}
	return false
}

func sequentialCreditDigits(value string) bool {
	return strings.Contains("01234567890123456789", value) || strings.Contains("98765432109876543210", value)
}
