package repository

import (
	"context"
	"database/sql"
	"errors"

	"pulsa2/internal/helper"
)

type MemberFeeCategoryRepository struct {
	db *sql.DB
}

type MemberFeeCategoryRow struct {
	ID           int64  `json:"id"`
	MemberID     int64  `json:"member_id"`
	FeeCode      string `json:"fee_code"`
	KategoriNama string `json:"kategori_nama"`
	FeeRp        int64  `json:"fee_rp"`
	Aktif        bool   `json:"aktif"`
}

func NewMemberFeeCategoryRepository(db *sql.DB) *MemberFeeCategoryRepository {
	return &MemberFeeCategoryRepository{db: db}
}

func (r *MemberFeeCategoryRepository) resolveFeeCode(ctx context.Context, feeCode string, kategoriID int64) (string, error) {
	if normalizedCode, ok := helper.NormalizeH2HFeeCategory(feeCode); ok {
		return normalizedCode, nil
	}
	if kategoriID <= 0 {
		return "", errors.New("fee_code invalid")
	}

	var rawName string
	err := r.db.QueryRowContext(ctx, `
SELECT UPPER(TRIM(nama))
FROM public.kategori
WHERE id = $1
LIMIT 1
`, kategoriID).Scan(&rawName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("fee_code invalid")
		}
		return "", err
	}

	normalizedCode, ok := helper.NormalizeH2HFeeCategory(rawName)
	if !ok {
		return "", errors.New("fee_code invalid")
	}
	return normalizedCode, nil
}

func (r *MemberFeeCategoryRepository) Upsert(ctx context.Context, memberID int64, feeCode string, kategoriID, feeRp int64, aktif bool) error {
	normalizedCode, err := r.resolveFeeCode(ctx, feeCode, kategoriID)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `
INSERT INTO public.member_h2h_fee
  (member_id, fee_code, fee_rp, aktif, dibuat_pada, diubah_pada)
VALUES
  ($1, $2, $3, $4, now(), now())
ON CONFLICT (member_id, fee_code)
DO UPDATE SET
  fee_rp = EXCLUDED.fee_rp,
  aktif = EXCLUDED.aktif,
  diubah_pada = now()
`, memberID, normalizedCode, feeRp, aktif)
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

func (r *MemberFeeCategoryRepository) Delete(ctx context.Context, memberID int64, feeCode string, kategoriID int64) error {
	normalizedCode, err := r.resolveFeeCode(ctx, feeCode, kategoriID)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `
DELETE FROM public.member_h2h_fee
WHERE member_id = $1
  AND fee_code = $2
`, memberID, normalizedCode)
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

func (r *MemberFeeCategoryRepository) List(ctx context.Context, memberID int64) ([]MemberFeeCategoryRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
  mf.id,
  mf.member_id,
  mf.fee_code,
  mf.fee_code AS kategori_nama,
  mf.fee_rp,
  mf.aktif
FROM public.member_h2h_fee mf
WHERE mf.member_id = $1
ORDER BY CASE mf.fee_code
  WHEN 'DANA' THEN 1
  WHEN 'GOPAY' THEN 2
  WHEN 'OVO' THEN 3
  WHEN 'LINKAJA' THEN 4
  WHEN 'SHOPEEPAY' THEN 5
  ELSE 6
END ASC, mf.id ASC
`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemberFeeCategoryRow, 0, 32)
	for rows.Next() {
		var row MemberFeeCategoryRow
		if err := rows.Scan(&row.ID, &row.MemberID, &row.FeeCode, &row.KategoriNama, &row.FeeRp, &row.Aktif); err != nil {
			return nil, err
		}
		row.KategoriNama = helper.H2HFeeCategoryLabel(row.FeeCode)
		out = append(out, row)
	}
	return out, rows.Err()
}
