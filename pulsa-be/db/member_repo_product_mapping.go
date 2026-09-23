package db

import (
	"context"
	"database/sql"
	"strings"
)

// ResolveProviderProductCode mencoba map SKU internal -> kode produk provider.
// Urutan:
// 1) tabel relasi public.produk_provider_map (jika tersedia)
// 2) kolom legacy di public.produk (kode_javapay / kode_yuscom) bila ada
// 3) return empty => caller pakai fallback default
func (r *MemberRepo) ResolveProviderProductCode(ctx context.Context, provider string, internalSKU string) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if provider == "" || internalSKU == "" {
		return "", nil
	}

	// 1) mapping table (opsional; aman jika belum ada)
	var mapped sql.NullString
	err := r.DB.QueryRowContext(ctx, `
SELECT ppm.kode_provider
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
WHERE UPPER(p.sku) = $1
  AND LOWER(ppm.provider) = $2
  AND ppm.aktif = true
ORDER BY ppm.id DESC
LIMIT 1
`, internalSKU, provider).Scan(&mapped)
	if err == nil && mapped.Valid {
		v := strings.ToUpper(strings.TrimSpace(mapped.String))
		if v != "" {
			return v, nil
		}
	}
	if err != nil && err != sql.ErrNoRows {
		// jika tabel belum ada, lanjut fallback tanpa memutus transaksi
		if !looksLikeMissingRelation(err) {
			return "", err
		}
	}

	// 2) kolom legacy di produk (opsional)
	switch provider {
	case "javapay":
		var code sql.NullString
		err = r.DB.QueryRowContext(ctx, `
SELECT NULLIF(TRIM(kode_javapay), '')
FROM public.produk
WHERE UPPER(sku) = $1
  AND aktif = true
LIMIT 1
`, internalSKU).Scan(&code)
		if err == nil && code.Valid {
			v := strings.ToUpper(strings.TrimSpace(code.String))
			if v != "" {
				return v, nil
			}
		}
		if err != nil && err != sql.ErrNoRows {
			if !looksLikeMissingColumn(err) {
				return "", err
			}
		}
	case "yuscom":
		var code sql.NullString
		err = r.DB.QueryRowContext(ctx, `
SELECT NULLIF(TRIM(kode_yuscom), '')
FROM public.produk
WHERE UPPER(sku) = $1
  AND aktif = true
LIMIT 1
`, internalSKU).Scan(&code)
		if err == nil && code.Valid {
			v := strings.ToUpper(strings.TrimSpace(code.String))
			if v != "" {
				return v, nil
			}
		}
		if err != nil && err != sql.ErrNoRows {
			if !looksLikeMissingColumn(err) {
				return "", err
			}
		}
	}

	// fallback: if we're resolving yuscom and still have no mapping, just
	// assume the SKU itself is the provider code. this mirrors the behaviour
	// in other repos and keeps app-commerce use‑case simple.
	if strings.ToLower(provider) == "yuscom" {
		raw := strings.ToUpper(strings.TrimSpace(internalSKU))
		if raw != "" {
			return raw, nil
		}
	}

	return "", nil
}

func looksLikeMissingRelation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "relation") && strings.Contains(s, "does not exist")
}

func looksLikeMissingColumn(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "column") && strings.Contains(s, "does not exist")
}
