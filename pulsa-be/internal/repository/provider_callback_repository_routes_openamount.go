package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *ProviderCallbackRepository) GetProviderOpenAmountRuleBySKU(ctx context.Context, provider string, internalSKU string) (*ProviderOpenAmountRule, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if provider == "" || internalSKU == "" {
		return nil, nil
	}

	var (
		kodeProvider sql.NullString
		special      sql.NullString
		mode         sql.NullString
		minNom       sql.NullInt64
		maxNom       sql.NullInt64
	)
	err := r.db.QueryRowContext(ctx, `
SELECT ppm.kode_provider, ppm.special_code, ppm.mode, ppm.minimal_nominal, ppm.maksimal_nominal
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
JOIN public.provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
WHERE UPPER(TRIM(p.sku)) = $1
  AND LOWER(TRIM(ppm.provider)) = $2
  AND pr.aktif = true
  AND ppm.aktif = true
  AND (
    ppm.jam_buka IS NULL OR ppm.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN ppm.jam_buka AND ppm.jam_tutup
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
`, internalSKU, provider).Scan(&kodeProvider, &special, &mode, &minNom, &maxNom)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if looksLikeMissingRelation(err) || looksLikeMissingColumn(err) {
			return nil, nil
		}
		return nil, err
	}

	out := &ProviderOpenAmountRule{}
	if kodeProvider.Valid {
		out.KodeProvider = strings.ToUpper(strings.TrimSpace(kodeProvider.String))
	}
	if special.Valid {
		v := strings.ToUpper(strings.TrimSpace(special.String))
		if v != "" {
			out.SpecialCode = &v
		}
	}
	if mode.Valid {
		v := strings.ToUpper(strings.TrimSpace(mode.String))
		if v != "" {
			out.Mode = &v
		}
	}
	if minNom.Valid {
		v := minNom.Int64
		out.MinimalNominal = &v
	}
	if maxNom.Valid {
		v := maxNom.Int64
		out.MaksimalNominal = &v
	}
	return out, nil
}

func (r *ProviderCallbackRepository) GetProviderOpenAmountRuleBySKUNominal(ctx context.Context, provider string, internalSKU string, nominal int64) (*ProviderOpenAmountRule, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if provider == "" || internalSKU == "" {
		return nil, nil
	}
	if nominal <= 0 {
		return r.GetProviderOpenAmountRuleBySKU(ctx, provider, internalSKU)
	}

	var (
		kodeProvider sql.NullString
		special      sql.NullString
		mode         sql.NullString
		minNom       sql.NullInt64
		maxNom       sql.NullInt64
	)
	err := r.db.QueryRowContext(ctx, `
SELECT ppm.kode_provider, ppm.special_code, ppm.mode, ppm.minimal_nominal, ppm.maksimal_nominal
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
    ppm.jam_buka IS NULL OR ppm.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN ppm.jam_buka AND ppm.jam_tutup
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
`, internalSKU, provider, nominal).Scan(&kodeProvider, &special, &mode, &minNom, &maxNom)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if looksLikeMissingRelation(err) || looksLikeMissingColumn(err) {
			return nil, nil
		}
		return nil, err
	}

	out := &ProviderOpenAmountRule{}
	if kodeProvider.Valid {
		out.KodeProvider = strings.ToUpper(strings.TrimSpace(kodeProvider.String))
	}
	if special.Valid {
		v := strings.ToUpper(strings.TrimSpace(special.String))
		if v != "" {
			out.SpecialCode = &v
		}
	}
	if mode.Valid {
		v := strings.ToUpper(strings.TrimSpace(mode.String))
		if v != "" {
			out.Mode = &v
		}
	}
	if minNom.Valid {
		v := minNom.Int64
		out.MinimalNominal = &v
	}
	if maxNom.Valid {
		v := maxNom.Int64
		out.MaksimalNominal = &v
	}
	return out, nil
}
