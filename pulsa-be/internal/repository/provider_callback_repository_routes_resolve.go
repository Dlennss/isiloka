package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (r *ProviderCallbackRepository) ResolveProviderProductCode(ctx context.Context, provider string, internalSKU string) (string, error) {
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
		s := strings.ToLower(err.Error())
		if !(strings.Contains(s, "relation") && strings.Contains(s, "does not exist")) {
			return "", err
		}
	}

	// if mapping table and legacy column both failed to yield a code, and
	// caller asked for yuscom we can treat the internal SKU itself as the
	// provider code. this allows app commerce clients to send native
	// yuscom product codes without creating a database mapping entry.
	if provider == "yuscom" {
		raw := strings.ToUpper(strings.TrimSpace(internalSKU))
		if raw != "" {
			return raw, nil
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
	}
	return "", nil
}

func (r *ProviderCallbackRepository) ResolveProviderProductCodeByNominal(ctx context.Context, provider string, internalSKU string, nominal int64) (string, error) {
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
		return r.ResolveProviderProductCode(ctx, provider, internalSKU)
	}
	if err != nil {
		s := strings.ToLower(err.Error())
		if !(strings.Contains(s, "relation") && strings.Contains(s, "does not exist")) {
			return "", err
		}
	}
	return r.ResolveProviderProductCode(ctx, provider, internalSKU)
}

func (r *ProviderCallbackRepository) GetProviderFeeByProduct(ctx context.Context, provider string, kodeProduk string, produkProviderMapID *int64, kodeProvider string) (int64, string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	kodeProduk = strings.ToUpper(strings.TrimSpace(kodeProduk))
	kodeProvider = strings.ToUpper(strings.TrimSpace(kodeProvider))
	if provider == "" || kodeProduk == "" {
		return 0, "", nil
	}

	var feeRP sql.NullInt64
	if produkProviderMapID != nil && *produkProviderMapID > 0 {
		err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(ppm.fee_rp, 0)
FROM public.produk_provider_map ppm
JOIN public.provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
WHERE ppm.id = $1
  AND LOWER(TRIM(ppm.provider)) = $2
  AND pr.aktif = true
  AND ppm.aktif = true
  AND (
    ppm.jam_buka IS NULL OR ppm.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN ppm.jam_buka AND ppm.jam_tutup
  )
LIMIT 1
`, *produkProviderMapID, provider).Scan(&feeRP)
		if err == nil && feeRP.Valid && feeRP.Int64 > 0 {
			return feeRP.Int64, "produk_provider_map_id", nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, "", err
		}
	}

	err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(ppm.fee_rp, 0)
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
ORDER BY CASE WHEN $3 <> '' AND UPPER(TRIM(ppm.kode_provider)) = $3 THEN 0 ELSE 1 END,
  CASE
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN:%' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN2' THEN 1
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN2:%' THEN 1
    ELSE 0
  END ASC, ppm.id DESC
LIMIT 1
`, kodeProduk, provider, kodeProvider).Scan(&feeRP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", nil
		}
		return 0, "", err
	}

	if feeRP.Valid && feeRP.Int64 > 0 {
		return feeRP.Int64, "produk_provider_map", nil
	}

	return 0, "", nil
}

type CallbackProviderWalletTxIn struct {
	Provider              string
	RefID                 string
	Arah                  string
	Jumlah                int64
	AllowNegative         bool
	Alasan                string
	Catatan               string
	TransaksiMemberID     *int64
	TransaksiProviderID   *int64
	AppOrderProviderTrxID *int64
	MetaJSON              []byte
}
