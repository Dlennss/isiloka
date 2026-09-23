package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *BankRepository) Create(ctx context.Context, actorID int64, in BankUpsertInput, refID string) (id int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = tx.QueryRowContext(ctx, `
INSERT INTO public.bank (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
VALUES ($1, $2, $3, $4, $5, $6, now(), now())
RETURNING id
`, strings.TrimSpace(in.Nama), strings.TrimSpace(in.NomorRekening), strings.TrimSpace(in.AtasNama), in.Saldo, in.Aktif, in.AdminStaffOnly).Scan(&id); err != nil {
		return 0, err
	}

	if in.Saldo > 0 {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'BANK_OPENING_BALANCE','Saldo awal pembuatan bank',$4,$5,$6,now())
`, id, refID, in.Saldo, 0, in.Saldo, actorID); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *BankRepository) Update(ctx context.Context, in BankUpsertInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.bank
SET nama = $2,
    nomor_rekening = $3,
    atas_nama = $4,
    aktif = $5,
    admin_staff_only = $6,
    diubah_pada = now()
WHERE id = $1
`, in.ID, strings.TrimSpace(in.Nama), strings.TrimSpace(in.NomorRekening), strings.TrimSpace(in.AtasNama), in.Aktif, in.AdminStaffOnly)
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

func (r *BankRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.bank
SET aktif = false, diubah_pada = now()
WHERE id = $1
`, id)
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

func (r *BankRepository) ToggleActive(ctx context.Context, id int64, aktif bool) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.bank
SET aktif = $2, diubah_pada = now()
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
