package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

// ListRouteCandidates — satu fungsi untuk ambil kandidat routing.
// Dipakai oleh PAY dan callback fallback.
//
// Filter dari produk_provider_map:
//   - SKU match
//   - provider aktif dan mapping aktif
//   - Nominal dalam range
//   - Jam online mapping (jam_buka <= sekarang <= jam_tutup)
//   - Exclude map ID dan code key yang sudah pernah dicoba (per refid)
//
// Hasil random supaya load tersebar.
func ListRouteCandidates(ctx context.Context, db *sql.DB, internalSKU string, nominal int64, excludeMapIDs []int64, excludeCodeKeys []string, allowLoketBankFallback bool) ([]ProviderRouteCandidate, error) {
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if internalSKU == "" {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
SELECT
  ppm.id,
  ppm.provider,
  ppm.kode_provider,
  ppm.special_code,
  ppm.mode,
  ppm.minimal_nominal,
  ppm.maksimal_nominal
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
JOIN public.provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
WHERE UPPER(TRIM(p.sku)) = $1
  AND LOWER(TRIM(ppm.provider)) = 'pulsa24jam'
  AND pr.aktif = true
  AND ppm.aktif = true
  AND ($2 <= 0 OR (ppm.minimal_nominal IS NULL OR ppm.minimal_nominal <= $2))
  AND ($2 <= 0 OR (ppm.maksimal_nominal IS NULL OR ppm.maksimal_nominal >= $2))
  AND (
    p.jam_buka IS NULL OR p.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN p.jam_buka AND p.jam_tutup
  )
  AND (COALESCE(array_length($3::bigint[], 1), 0) = 0 OR NOT (ppm.id = ANY($3)))
  AND (
    COALESCE(array_length($4::text[], 1), 0) = 0
    OR NOT ((LOWER(TRIM(ppm.provider)) || '#code:' || UPPER(TRIM(ppm.kode_provider))) = ANY($4))
  )
  AND ($5 = true OR $5 = false)
ORDER BY
  CASE
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN:%' THEN 0
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN2' THEN 1
    WHEN LOWER(TRIM(ppm.provider)) = 'smb' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN2:%' THEN 1
    ELSE 0
  END ASC,
  random(), ppm.id DESC
`, internalSKU, nominal, pq.Array(excludeMapIDs), pq.Array(excludeCodeKeys), allowLoketBankFallback)
	if err != nil {
		if looksLikeMissingRelation(err) || looksLikeMissingColumn(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]ProviderRouteCandidate, 0, 8)
	for rows.Next() {
		var (
			id       sql.NullInt64
			provider sql.NullString
			code     sql.NullString
			special  sql.NullString
			mode     sql.NullString
			minNom   sql.NullInt64
			maxNom   sql.NullInt64
		)
		if err := rows.Scan(&id, &provider, &code, &special, &mode, &minNom, &maxNom); err != nil {
			return nil, err
		}
		candidate := ProviderRouteCandidate{
			ProdukSKUSnapshot: internalSKU,
			Provider:          strings.ToLower(strings.TrimSpace(provider.String)),
			KodeProvider:      strings.ToUpper(strings.TrimSpace(code.String)),
		}
		if id.Valid {
			v := id.Int64
			candidate.ProdukProviderMapID = &v
		}
		if special.Valid {
			v := strings.ToUpper(strings.TrimSpace(special.String))
			if v != "" {
				candidate.SpecialCode = &v
			}
		}
		if mode.Valid {
			v := strings.ToUpper(strings.TrimSpace(mode.String))
			if v != "" {
				candidate.Mode = &v
			}
		}
		if minNom.Valid {
			v := minNom.Int64
			candidate.MinimalNominal = &v
		}
		if maxNom.Valid {
			v := maxNom.Int64
			candidate.MaksimalNominal = &v
		}
		if candidate.Provider != "" && candidate.KodeProvider != "" {
			out = append(out, candidate)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
