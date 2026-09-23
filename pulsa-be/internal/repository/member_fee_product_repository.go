package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type MemberFeeProductRepository struct {
	db *sql.DB
}

func NewMemberFeeProductRepository(db *sql.DB) *MemberFeeProductRepository {
	return &MemberFeeProductRepository{db: db}
}

func (r *MemberFeeProductRepository) Upsert(ctx context.Context, memberID int64, produkID int64, kodeProduk string, feePersen *float64, feeRp *int64) error {
	kodeProduk = strings.TrimSpace(kodeProduk)
	res, err := r.db.ExecContext(ctx, `
INSERT INTO public.member_fee_produk
  (member_id, produk_id, fee_persen, fee_rp, dibuat_pada, diubah_pada)
SELECT
  $1, p.id, $4, $5, now(), now()
FROM public.produk p
WHERE ($2 > 0 AND p.id = $2)
   OR ($3 <> '' AND UPPER(p.sku) = UPPER($3))
ON CONFLICT (member_id, produk_id)
DO UPDATE SET
  fee_persen = EXCLUDED.fee_persen,
  fee_rp = EXCLUDED.fee_rp,
  diubah_pada = now()
`, memberID, produkID, kodeProduk, feePersen, feeRp)
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

func (r *MemberFeeProductRepository) Delete(ctx context.Context, memberID int64, produkID int64, kodeProduk string) error {
	kodeProduk = strings.TrimSpace(kodeProduk)
	res, err := r.db.ExecContext(ctx, `
DELETE FROM public.member_fee_produk mfp
USING public.produk p
WHERE mfp.member_id = $1
  AND mfp.produk_id = p.id
  AND (
    ($2 > 0 AND p.id = $2)
    OR ($3 <> '' AND UPPER(p.sku) = UPPER($3))
  )
`, memberID, produkID, kodeProduk)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *MemberFeeProductRepository) List(ctx context.Context, memberID int64) ([]MemberFeeProductRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT mfp.id, mfp.member_id, p.id, p.sku, p.nama, mfp.fee_persen, mfp.fee_rp
FROM public.member_fee_produk mfp
JOIN public.produk p ON p.id = mfp.produk_id
WHERE mfp.member_id = $1
ORDER BY p.sku ASC
`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemberFeeProductRow, 0, 64)
	for rows.Next() {
		var (
			row MemberFeeProductRow
			p   *float64
			rp  *int64
		)
		if err := rows.Scan(&row.ID, &row.MemberID, &row.ProdukID, &row.KodeProduk, &row.NamaProduk, &p, &rp); err != nil {
			return nil, err
		}
		row.FeePersen = p
		row.FeeRp = rp
		out = append(out, row)
	}
	return out, rows.Err()
}
