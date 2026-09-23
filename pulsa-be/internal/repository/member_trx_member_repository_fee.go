package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"pulsa2/internal/helper"
)

func (r *MemberTrxMemberRepository) GetMemberFee(ctx context.Context, memberID int64) (int64, error) {
	var fee int64
	err := r.db.QueryRowContext(ctx, `SELECT fee_member_rp FROM public.member WHERE id=$1`, memberID).Scan(&fee)
	if err != nil {
		return 0, err
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

func (r *MemberTrxMemberRepository) GetMemberFeeByH2HCategory(ctx context.Context, memberID int64, kodeProduk string, qty int64) (int64, string, error) {
	kodeProduk = strings.TrimSpace(strings.ToUpper(kodeProduk))
	if memberID <= 0 || kodeProduk == "" {
		return 0, "", nil
	}
	if qty < 0 {
		qty = 0
	}
	categoryName, catErr := r.ClassifyH2HFeeCategoryBySKU(ctx, kodeProduk)
	if catErr != nil {
		return 0, "", catErr
	}

	var feeRp sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
SELECT fee_rp
FROM public.member_h2h_fee mhf
WHERE mhf.member_id = $1
  AND mhf.aktif = true
  AND mhf.fee_code = $2
LIMIT 1
`, memberID, categoryName).Scan(&feeRp)
	if err == nil {
		if feeRp.Valid {
			if feeRp.Int64 < 0 {
				return 0, "category_flat", nil
			}
			return feeRp.Int64, "category_flat", nil
		}
		return 0, "category_flat", nil
	}
	if err != sql.ErrNoRows {
		return 0, "", err
	}
	return 0, "", nil
}

func ClassifyH2HFeeCategoryBySKU(ctx context.Context, db *sql.DB, sku string) (string, error) {
	sku = strings.TrimSpace(strings.ToUpper(sku))
	if sku == "" {
		return "", nil
	}
	if db == nil {
		return helper.ClassifyH2HFeeCategory(sku), nil
	}

	var kategoriNama string
	var brandNama string
	err := db.QueryRowContext(ctx, `
SELECT COALESCE(k.nama, ''), COALESCE(b.nama, '')
FROM public.produk p
LEFT JOIN public.kategori k ON k.id = p.kategori_id
LEFT JOIN public.brand b ON b.id = p.brand_id
WHERE UPPER(TRIM(p.sku)) = $1
  AND p.aktif = true
LIMIT 1
`, sku).Scan(&kategoriNama, &brandNama)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return helper.ClassifyH2HFeeCategory(sku), nil
		}
		return "", err
	}

	categoryName := helper.ClassifyH2HFeeCategoryFromMetadata(sku, kategoriNama, brandNama)
	if categoryName == "" {
		categoryName = helper.ClassifyH2HFeeCategory(sku)
	}
	return categoryName, nil
}

func (r *MemberTrxMemberRepository) ClassifyH2HFeeCategoryBySKU(ctx context.Context, sku string) (string, error) {
	if r == nil {
		return helper.ClassifyH2HFeeCategory(sku), nil
	}
	return ClassifyH2HFeeCategoryBySKU(ctx, r.db, sku)
}

func (r *MemberTrxMemberRepository) GetProviderFeeByProduct(ctx context.Context, provider string, kodeProduk string, produkProviderMapID *int64, kodeProvider string) (int64, string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	kodeProduk = strings.ToUpper(strings.TrimSpace(kodeProduk))
	kodeProvider = strings.ToUpper(strings.TrimSpace(kodeProvider))
	if provider == "" || kodeProduk == "" {
		return 0, "", nil
	}

	var fee sql.NullInt64
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
`, *produkProviderMapID, provider).Scan(&fee)
		if err == nil && fee.Valid && fee.Int64 > 0 {
			return fee.Int64, "produk_provider_map_id", nil
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
ORDER BY CASE WHEN $3 <> '' AND UPPER(TRIM(ppm.kode_provider)) = $3 THEN 0 ELSE 1 END, ppm.id DESC
LIMIT 1
`, kodeProduk, provider, kodeProvider).Scan(&fee)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", nil
		}
		return 0, "", err
	}
	if !fee.Valid || fee.Int64 <= 0 {
		return 0, "", nil
	}
	return fee.Int64, "produk_provider_map", nil
}

func (r *MemberTrxMemberRepository) GetProdukPricingRuleBySKU(ctx context.Context, sku string) (*ProdukPricingRule, error) {
	sku = strings.TrimSpace(strings.ToUpper(sku))
	if sku == "" {
		return nil, nil
	}

	var (
		tipeHarga    string
		nominal      sql.NullInt64
		maxNom       sql.NullInt64
		jamBuka      string
		jamTutup     string
		aktif        bool
		kategoriNama string
		brandNama    string
	)

	err := r.db.QueryRowContext(ctx, `
SELECT p.tipe_harga::text, p.nominal, p.maksimal_nominal, p.jam_buka::text, p.jam_tutup::text, p.aktif,
       COALESCE(k.nama, ''), COALESCE(b.nama, '')
FROM public.produk p
LEFT JOIN public.kategori k ON k.id = p.kategori_id
LEFT JOIN public.brand b ON b.id = p.brand_id
WHERE UPPER(TRIM(p.sku)) = $1
  AND p.aktif = true
LIMIT 1
`, sku).Scan(&tipeHarga, &nominal, &maxNom, &jamBuka, &jamTutup, &aktif, &kategoriNama, &brandNama)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	out := &ProdukPricingRule{
		TipeHarga:    strings.ToUpper(strings.TrimSpace(tipeHarga)),
		JamBuka:      jamBuka,
		JamTutup:     jamTutup,
		Aktif:        aktif,
		KategoriNama: strings.TrimSpace(kategoriNama),
		BrandNama:    strings.TrimSpace(brandNama),
	}
	if nominal.Valid {
		v := nominal.Int64
		out.Nominal = &v
	}
	if maxNom.Valid {
		v := maxNom.Int64
		out.MaksimalNominal = &v
	}
	return out, nil
}
