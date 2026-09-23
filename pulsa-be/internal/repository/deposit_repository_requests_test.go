package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

type fakeDepositRowScanner struct {
	values []any
}

func (s fakeDepositRowScanner) Scan(dest ...any) error {
	if len(dest) != len(s.values) {
		return fmt.Errorf("scan dest count mismatch: got %d want %d", len(dest), len(s.values))
	}
	for i, d := range dest {
		v := s.values[i]
		switch out := d.(type) {
		case *int64:
			n, ok := v.(int64)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into int64", i, v)
			}
			*out = n
		case *string:
			if v == nil {
				return fmt.Errorf("column %d: converting NULL to string is unsupported", i)
			}
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into string", i, v)
			}
			*out = s
		case *sql.NullString:
			if v == nil {
				*out = sql.NullString{}
				continue
			}
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into NullString", i, v)
			}
			*out = sql.NullString{String: s, Valid: true}
		case *sql.NullInt64:
			if v == nil {
				*out = sql.NullInt64{}
				continue
			}
			n, ok := v.(int64)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into NullInt64", i, v)
			}
			*out = sql.NullInt64{Int64: n, Valid: true}
		case *time.Time:
			t, ok := v.(time.Time)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into time.Time", i, v)
			}
			*out = t
		case **time.Time:
			if v == nil {
				*out = nil
				continue
			}
			t, ok := v.(time.Time)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into *time.Time", i, v)
			}
			*out = &t
		case **int64:
			if v == nil {
				*out = nil
				continue
			}
			n, ok := v.(int64)
			if !ok {
				return fmt.Errorf("column %d: cannot scan %T into *int64", i, v)
			}
			*out = &n
		default:
			return fmt.Errorf("column %d: unsupported scan dest %T", i, d)
		}
	}
	return nil
}

func TestScanDepositRowAllowsNullableOptionalStrings(t *testing.T) {
	createdAt := time.Date(2026, 7, 14, 13, 57, 59, 0, time.FixedZone("WIB", 7*60*60))
	row, err := scanDepositRow(fakeDepositRowScanner{values: []any{
		int64(100799),
		int64(48),
		"K24-MANDIRI-3-test",
		"Murahslt HP",
		int64(3),
		"MANDIRI",
		"1560027075276",
		"PT PULSA MITRA NASIONAL",
		int64(25058568),
		int64(25000000),
		int64(58568),
		int64(25058568),
		"transfer",
		nil,
		"approved",
		"manual correction",
		createdAt,
		nil,
		nil,
		nil,
	}})
	if err != nil {
		t.Fatalf("scanDepositRow returned error: %v", err)
	}
	if row.BuktiURL != "" {
		t.Fatalf("BuktiURL = %q, want empty string", row.BuktiURL)
	}
	if row.BankNama != "MANDIRI" || row.Status != "approved" || row.Requested != 25000000 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestRandomDepositUniqueCodeUsesLastThreeDigits(t *testing.T) {
	for i := 0; i < 200; i++ {
		code, err := randomDepositUniqueCode()
		if err != nil {
			t.Fatal(err)
		}
		if code < 1 || code > 999 {
			t.Fatalf("kode unik = %d, want 1..999", code)
		}
	}
}
