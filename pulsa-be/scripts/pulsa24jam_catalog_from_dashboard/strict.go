package main

import (
	"fmt"
	"strconv"
	"strings"
)

// The dashboard exports five fields per product. Never scan ahead for a price:
// doing so can silently attach the next product's price to a malformed row.
func parseDashboardStrict(lines []string) ([]productRow, error) {
	first := -1
	for i := range lines {
		if isTableHeader(lines, i) {
			first = i
			break
		}
	}
	if first < 2 {
		return nil, fmt.Errorf("missing category/brand table header")
	}
	expected := -1
	expectedFixed := -1
	for i := 0; i+1 < first; i++ {
		var target *int
		switch lines[i] {
		case "Total Produk Aktif":
			target = &expected
		case "Produk Dengan Harga Tetap":
			target = &expectedFixed
		}
		if target != nil {
			n, err := strconv.Atoi(strings.ReplaceAll(lines[i+1], ".", ""))
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid dashboard count %q", lines[i+1])
			}
			*target = n
		}
	}
	if expected < 1 || expectedFixed < 0 {
		return nil, fmt.Errorf("dashboard totals are required")
	}
	category, brand := lines[first-2], lines[first-1]
	rows := make([]productRow, 0, expected)
	seen := make(map[string]bool)
	fixed := 0
	for i := first + 3; i < len(lines); {
		if isTableHeader(lines, i+1) {
			brand = lines[i]
			i += 4
			continue
		}
		if isTableHeader(lines, i+2) {
			category, brand = lines[i], lines[i+1]
			i += 5
			continue
		}
		if i+4 >= len(lines) {
			return nil, fmt.Errorf("truncated product at line %d", i+1)
		}
		sku, name, group, kind, priceLine := lines[i], lines[i+1], lines[i+2], lines[i+3], lines[i+4]
		if sku == "" || seen[sku] {
			return nil, fmt.Errorf("invalid or duplicate SKU %q at line %d", sku, i+1)
		}
		priceType := kind
		max := ""
		switch {
		case kind == "FIXED":
			fixed++
		case kind == "OPEN_AMOUNT" || strings.HasPrefix(kind, "OPEN_AMOUNT "):
			priceType = "OPEN_AMOUNT"
			max = parseMaximumNominal(kind)
			if strings.Contains(strings.ToUpper(kind), "MAKS") && max == "" {
				return nil, fmt.Errorf("invalid limit for %s", sku)
			}
		default:
			return nil, fmt.Errorf("invalid price type %q for %s at line %d", kind, sku, i+4)
		}
		price, ok := parsePriceLine(priceLine)
		if !ok {
			return nil, fmt.Errorf("invalid price %q for %s", priceLine, sku)
		}
		rows = append(rows, productRow{Category: category, Brand: brand, SKU: sku, Name: name, GroupName: group, PriceType: priceType, Price: price, MaximumNominal: max})
		seen[sku] = true
		i += 5
	}
	if len(rows) != expected || fixed != expectedFixed {
		return nil, fmt.Errorf("catalog incomplete: parsed %d/%d products, %d/%d fixed", len(rows), expected, fixed, expectedFixed)
	}
	return rows, nil
}
