package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (r *UserRepository) Create(ctx context.Context, in UserCreateInput) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (
	email, nama, password_hash, pin_hash, role, aktif,
	retail_agent_commission_rp, retail_master_commission_rp,
	h2h_agent_commission_rp, h2h_master_commission_rp
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id
`, in.Email, nullIfEmptyStr(in.Nama), in.PasswordHash, in.PinHash, in.Role, in.Aktif, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp, in.H2HAgentCommissionRp, in.H2HMasterCommissionRp).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create member failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, id)
	if err != nil {
		return 0, fmt.Errorf("create dompet failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.member_api_key (member_id, api_key, label, aktif)
VALUES ($1,$2,$3,true)
`, id, in.APIKey, nullIfEmptyStr(in.Label))
	if err != nil {
		return 0, fmt.Errorf("create api_key failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *UserRepository) Update(ctx context.Context, in UserUpdateInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET email = $2,
	nama = $3,
	phone = $4,
	role = $5,
	aktif = $6,
	fee_member_rp = $7,
	retail_agent_commission_rp = $8,
	retail_master_commission_rp = $9,
	h2h_agent_commission_rp = $10,
	h2h_master_commission_rp = $11,
	password_hash = COALESCE($12, password_hash),
	pin_hash = COALESCE($13, pin_hash)
WHERE id = $1
`, in.ID, in.Email, nullIfEmptyStr(in.Nama), nullIfEmptyStr(in.Phone), in.Role, in.Aktif, in.FeeMemberRp, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp, in.H2HAgentCommissionRp, in.H2HMasterCommissionRp, in.PasswordHash, in.PinHash)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
DELETE FROM public.member
WHERE id = $1
`, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) SetFee(ctx context.Context, userID int64, feeRp int64) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET fee_member_rp = $2
WHERE id = $1
`, userID, feeRp)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) SetPassword(ctx context.Context, userID int64, passwordHash string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET password_hash = $2
WHERE id = $1
`, userID, passwordHash)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) SetPIN(ctx context.Context, userID int64, pinHash string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET pin_hash = $2
WHERE id = $1
`, userID, pinHash)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) SetRetailHierarchy(ctx context.Context, memberID int64, retailAgentID, retailMasterID *int64) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET retail_agent_id = $2,
    retail_master_id = $3
WHERE id = $1
`, memberID, nullableInt64(retailAgentID), nullableInt64(retailMasterID))
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) SetRole(ctx context.Context, memberID int64, role string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET role = $2
WHERE id = $1
`, memberID, role)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) SetH2HHierarchy(ctx context.Context, memberID int64, h2hAgentID, h2hMasterID *int64) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET h2h_agent_member_id = $2,
    h2h_master_member_id = $3
WHERE id = $1
`, memberID, nullableInt64(h2hAgentID), nullableInt64(h2hMasterID))
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}
