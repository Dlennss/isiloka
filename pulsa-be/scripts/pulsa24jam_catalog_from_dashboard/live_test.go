package main

import (
	"strings"
	"testing"
)

func TestLiveCatalogStrict(t *testing.T) {
	raw := `{"ok":true,"items":[{"sku":"DB COMMONWEALTH","nama":"Transfer Bank","group_name":"","kategori_nama":"Bank Transfer","brand_nama":"Commonwealth","tipe_harga":"OPEN_AMOUNT","fee_tambahan":1500}]}`
	rows, err := parseLiveCatalog([]byte(raw))
	if err != nil || len(rows) != 1 || rows[0].SKU != "DB COMMONWEALTH" || rows[0].Price != 1500 || rows[0].GroupName != "" {
		t.Fatalf("%+v, %v", rows, err)
	}
	for _, invalid := range []string{
		strings.Replace(raw, `"ok":true`, `"ok":false`, 1),
		strings.Replace(raw, `,"fee_tambahan":1500`, "", 1),
		strings.Replace(raw, `"Commonwealth"`, `""`, 1),
		strings.Replace(raw, `1500`, `15.5`, 1),
		`{"ok":true,"items":[]}`,
	} {
		if _, err := parseLiveCatalog([]byte(invalid)); err == nil {
			t.Fatalf("accepted %s", invalid)
		}
	}
}

func TestPriceGrouping(t *testing.T) {
	for _, valid := range []string{"Rp 0", "Rp 1.500", "+ Rp 1000"} {
		if _, ok := parsePriceLine(valid); !ok {
			t.Fatal(valid)
		}
	}
	for _, invalid := range []string{"Rp 1.50", "Rp 1..500", "Rp -1", "Rp 1,500"} {
		if _, ok := parsePriceLine(invalid); ok {
			t.Fatal(invalid)
		}
	}
}
