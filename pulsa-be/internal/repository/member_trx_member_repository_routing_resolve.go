package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *MemberTrxMemberRepository) ResolveProviderProductCode(ctx context.Context, provider string, internalSKU string) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if provider == "" || internalSKU == "" {
		return "", nil
	}

	var mapped sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT ppm.kode_provider
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
JOIN public.provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
WHERE UPPER(TRIM(p.sku)) = $1
  AND LOWER(TRIM(ppm.provider)) = $2
  AND pr.aktif = true
  AND ppm.aktif = true
  AND (
    p.jam_buka IS NULL OR p.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN p.jam_buka AND p.jam_tutup
  )
ORDER BY
  CASE
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN:%' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN2' THEN 1
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN2:%' THEN 1
    ELSE 0
  END ASC,
  ppm.id DESC
LIMIT 1
`, internalSKU, provider).Scan(&mapped)
	if err == nil && mapped.Valid {
		v := strings.ToUpper(strings.TrimSpace(mapped.String))
		if v != "" {
			return v, nil
		}
	}
	if err != nil && err != sql.ErrNoRows {
		if !looksLikeMissingRelation(err) {
			return "", err
		}
	}

	switch provider {
	case "javapay":
		var code sql.NullString
		err = r.db.QueryRowContext(ctx, `
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
		err = r.db.QueryRowContext(ctx, `
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

	return "", nil
}

func (r *MemberTrxMemberRepository) ResolveProviderProductCodeByNominal(ctx context.Context, provider string, internalSKU string, nominal int64) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if provider == "" || internalSKU == "" {
		return "", nil
	}
	if nominal <= 0 {
		return r.ResolveProviderProductCode(ctx, provider, internalSKU)
	}

	var mapped sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT ppm.kode_provider
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
JOIN public.provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
WHERE UPPER(TRIM(p.sku)) = $1
  AND LOWER(TRIM(ppm.provider)) = $2
  AND pr.aktif = true
  AND ppm.aktif = true
  AND (ppm.minimal_nominal IS NULL OR ppm.minimal_nominal <= $3)
  AND (ppm.maksimal_nominal IS NULL OR ppm.maksimal_nominal >= $3)
  AND (
    p.jam_buka IS NULL OR p.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN p.jam_buka AND p.jam_tutup
  )
ORDER BY
  CASE
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN:%' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN2' THEN 1
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN2:%' THEN 1
    ELSE 0
  END ASC,
  COALESCE(ppm.maksimal_nominal, 9223372036854775807) ASC,
  COALESCE(ppm.minimal_nominal, 0) DESC,
  ppm.id DESC
LIMIT 1
`, internalSKU, provider, nominal).Scan(&mapped)
	if err == nil && mapped.Valid {
		v := strings.ToUpper(strings.TrimSpace(mapped.String))
		if v != "" {
			return v, nil
		}
	}
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		if !looksLikeMissingRelation(err) {
			return "", err
		}
	}
	return r.ResolveProviderProductCode(ctx, provider, internalSKU)
}
