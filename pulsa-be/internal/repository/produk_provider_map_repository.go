package repository

import (
	"context"
	"database/sql"
	"strings"
)

type ProdukProviderMapRepository struct {
	db *sql.DB
}

func NewProdukProviderMapRepository(db *sql.DB) *ProdukProviderMapRepository {
	return &ProdukProviderMapRepository{db: db}
}

func (r *ProdukProviderMapRepository) List(ctx context.Context, q string) ([]ProdukProviderMapRow, error) {
	q = strings.TrimSpace(q)
	rows, err := r.db.QueryContext(ctx, `
SELECT
  m.id,
  m.produk_id,
  p.sku,
  p.nama,
  p.tipe_harga::text,
  p.nominal,
  m.minimal_nominal,
  m.maksimal_nominal,
  m.provider,
  m.kode_provider,
  m.special_code,
  m.mode,
  COALESCE(m.fee_rp, 0),
  m.jam_buka::text,
  m.jam_tutup::text,
  m.aktif,
  m.dibuat_pada,
  m.diubah_pada
FROM public.produk_provider_map m
JOIN public.produk p ON p.id = m.produk_id
WHERE ($1 = '' OR p.sku ILIKE '%'||$1||'%' OR p.nama ILIKE '%'||$1||'%' OR m.provider ILIKE '%'||$1||'%' OR m.kode_provider ILIKE '%'||$1||'%')
ORDER BY m.id DESC
`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProdukProviderMapRow, 0, 128)
	for rows.Next() {
		var (
			row      ProdukProviderMapRow
			nominal  sql.NullInt64
			minNom   sql.NullInt64
			maxNom   sql.NullInt64
			special  sql.NullString
			mode     sql.NullString
			jamBuka  sql.NullString
			jamTutup sql.NullString
			dibuat   sql.NullTime
			diubah   sql.NullTime
		)
		if err := rows.Scan(
			&row.ID,
			&row.ProdukID,
			&row.ProdukSKU,
			&row.ProdukNama,
			&row.TipeHarga,
			&nominal,
			&minNom,
			&maxNom,
			&row.Provider,
			&row.KodeProvider,
			&special,
			&mode,
			&row.FeeRp,
			&jamBuka,
			&jamTutup,
			&row.Aktif,
			&dibuat,
			&diubah,
		); err != nil {
			return nil, err
		}
		if nominal.Valid {
			v := nominal.Int64
			row.Nominal = &v
		}
		if minNom.Valid {
			v := minNom.Int64
			row.MinimalNominal = &v
		}
		if maxNom.Valid {
			v := maxNom.Int64
			row.MaksimalNominal = &v
		}
		if special.Valid {
			v := strings.ToUpper(strings.TrimSpace(special.String))
			if v != "" {
				row.SpecialCode = &v
			}
		}
		if mode.Valid {
			v := strings.ToUpper(strings.TrimSpace(mode.String))
			if v != "" {
				row.Mode = &v
			}
		}
		if jamBuka.Valid {
			v := jamBuka.String
			row.JamBuka = &v
		}
		if jamTutup.Valid {
			v := jamTutup.String
			row.JamTutup = &v
		}
		if dibuat.Valid {
			v := dibuat.Time
			row.DibuatPada = &v
		}
		if diubah.Valid {
			v := diubah.Time
			row.DiubahPada = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *ProdukProviderMapRepository) Get(ctx context.Context, id int64) (*ProdukProviderMapRow, error) {
	var (
		row      ProdukProviderMapRow
		nominal  sql.NullInt64
		minNom   sql.NullInt64
		maxNom   sql.NullInt64
		special  sql.NullString
		mode     sql.NullString
		jamBuka  sql.NullString
		jamTutup sql.NullString
		dibuat   sql.NullTime
		diubah   sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  m.id,
  m.produk_id,
  p.sku,
  p.nama,
  p.tipe_harga::text,
  p.nominal,
  m.minimal_nominal,
  m.maksimal_nominal,
  m.provider,
  m.kode_provider,
  m.special_code,
  m.mode,
  COALESCE(m.fee_rp, 0),
  m.jam_buka::text,
  m.jam_tutup::text,
  m.aktif,
  m.dibuat_pada,
  m.diubah_pada
FROM public.produk_provider_map m
JOIN public.produk p ON p.id = m.produk_id
WHERE m.id = $1
LIMIT 1
`, id).Scan(
		&row.ID,
		&row.ProdukID,
		&row.ProdukSKU,
		&row.ProdukNama,
		&row.TipeHarga,
		&nominal,
		&minNom,
		&maxNom,
		&row.Provider,
		&row.KodeProvider,
		&special,
		&mode,
		&row.FeeRp,
		&jamBuka,
		&jamTutup,
		&row.Aktif,
		&dibuat,
		&diubah,
	)
	if err != nil {
		return nil, err
	}
	if nominal.Valid {
		v := nominal.Int64
		row.Nominal = &v
	}
	if minNom.Valid {
		v := minNom.Int64
		row.MinimalNominal = &v
	}
	if maxNom.Valid {
		v := maxNom.Int64
		row.MaksimalNominal = &v
	}
	if special.Valid {
		v := strings.ToUpper(strings.TrimSpace(special.String))
		if v != "" {
			row.SpecialCode = &v
		}
	}
	if mode.Valid {
		v := strings.ToUpper(strings.TrimSpace(mode.String))
		if v != "" {
			row.Mode = &v
		}
	}
	if jamBuka.Valid {
		v := jamBuka.String
		row.JamBuka = &v
	}
	if jamTutup.Valid {
		v := jamTutup.String
		row.JamTutup = &v
	}
	if dibuat.Valid {
		v := dibuat.Time
		row.DibuatPada = &v
	}
	if diubah.Valid {
		v := diubah.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *ProdukProviderMapRepository) Create(ctx context.Context, in ProdukProviderMapUpsertInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.produk_provider_map
  (produk_id, provider, kode_provider, special_code, mode, minimal_nominal, maksimal_nominal, fee_rp, jam_buka, jam_tutup, aktif, dibuat_pada, diubah_pada)
VALUES
  ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, $7, $8, CASE WHEN $9::text = '' THEN NULL ELSE $9::time END, CASE WHEN $10::text = '' THEN NULL ELSE $10::time END, $11, now(), now())
RETURNING id
`, in.ProdukID, in.Provider, in.KodeProvider, ptrOrEmpty(in.SpecialCode), ptrOrEmpty(in.Mode), in.MinimalNominal, in.MaksimalNominal, in.FeeRp, ptrOrEmpty(in.JamBuka), ptrOrEmpty(in.JamTutup), in.Aktif).Scan(&id)
	return id, err
}

func (r *ProdukProviderMapRepository) Update(ctx context.Context, in ProdukProviderMapUpsertInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.produk_provider_map
SET produk_id = $2,
    provider = $3,
    kode_provider = $4,
    special_code = NULLIF($5,''),
    mode = NULLIF($6,''),
    minimal_nominal = $7,
    maksimal_nominal = $8,
    fee_rp = $9,
    jam_buka = CASE WHEN $10::text = '' THEN NULL ELSE $10::time END,
    jam_tutup = CASE WHEN $11::text = '' THEN NULL ELSE $11::time END,
    aktif = $12,
    diubah_pada = now()
WHERE id = $1
`, in.ID, in.ProdukID, in.Provider, in.KodeProvider, ptrOrEmpty(in.SpecialCode), ptrOrEmpty(in.Mode), in.MinimalNominal, in.MaksimalNominal, in.FeeRp, ptrOrEmpty(in.JamBuka), ptrOrEmpty(in.JamTutup), in.Aktif)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProdukProviderMapRepository) UpdateAktif(ctx context.Context, id int64, aktif bool) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.produk_provider_map
SET aktif = $2,
    diubah_pada = now()
WHERE id = $1
`, id, aktif)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProdukProviderMapRepository) GetOpenAmountRule(ctx context.Context, provider string, internalSKU string) (*ProviderOpenAmountRule, error) {
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
WHERE UPPER(TRIM(p.sku)) = UPPER(TRIM($1))
  AND LOWER(TRIM(ppm.provider)) = LOWER(TRIM($2))
  AND pr.aktif = true
  AND ppm.aktif = true
  AND (
    ppm.jam_buka IS NULL OR ppm.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN ppm.jam_buka AND ppm.jam_tutup
  )
ORDER BY ppm.id DESC
LIMIT 1
`, internalSKU, provider).Scan(&kodeProvider, &special, &mode, &minNom, &maxNom)
	if err != nil {
		if err == sql.ErrNoRows {
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

func ptrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *ProdukProviderMapRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.produk_provider_map WHERE id = $1`, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}
