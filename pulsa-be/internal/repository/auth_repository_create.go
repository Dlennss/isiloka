package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pulsa2/internal/helper"
)

func (r *AuthRepository) CreateGoogleMember(ctx context.Context, email, nama, googleSub string) (*AuthGoogleRow, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var memberID int64
	var role string
	var aktif bool
	var outGoogleSub string
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (email, nama, role, aktif, google_sub)
VALUES ($1, $2, 'user', true, $3)
RETURNING id, role, aktif, COALESCE(google_sub, '')
`, strings.TrimSpace(strings.ToLower(email)), nullIfEmptyStr(strings.TrimSpace(nama)), strings.TrimSpace(googleSub)).Scan(&memberID, &role, &aktif, &outGoogleSub)
	if err != nil {
		return nil, fmt.Errorf("create google member failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return nil, fmt.Errorf("create dompet failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &AuthGoogleRow{
		ID:        memberID,
		Role:      role,
		Aktif:     aktif,
		GoogleSub: outGoogleSub,
	}, nil
}

func (r *AuthRepository) CreateMemberWithKey(ctx context.Context, in UserCreateInput) (int64, string, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = tx.Rollback() }()

	var existingID int64
	dupErr := tx.QueryRowContext(ctx, `
SELECT id
FROM public.member
WHERE LOWER(email) = LOWER($1)
LIMIT 1
`, in.Email).Scan(&existingID)
	if dupErr == nil {
		return 0, "", fmt.Errorf("email sudah terdaftar")
	}
	if dupErr != sql.ErrNoRows {
		return 0, "", dupErr
	}

	var memberID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (
  email, nama, phone, password_hash, pin_hash, role, aktif,
  retail_agent_commission_rp, retail_master_commission_rp,
  h2h_agent_commission_rp, h2h_master_commission_rp,
  h2h_agent_member_id, h2h_master_member_id
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING id
`, in.Email, nullIfEmptyStr(in.Nama), nullIfEmptyStr(in.Phone), in.PasswordHash, in.PinHash, in.Role, in.Aktif, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp, in.H2HAgentCommissionRp, in.H2HMasterCommissionRp, nullableInt64Value(in.H2HAgentID), nullableInt64Value(in.H2HMasterID)).Scan(&memberID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "member_email_key") {
			return 0, "", fmt.Errorf("email sudah terdaftar")
		}
		return 0, "", fmt.Errorf("create member failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return 0, "", fmt.Errorf("create dompet failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.member_api_key (member_id, api_key, label, aktif)
VALUES ($1,$2,$3,true)
`, memberID, in.APIKey, nullIfEmptyStr(in.Label))
	if err != nil {
		return 0, "", fmt.Errorf("create api_key failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return memberID, in.APIKey, nil
}

func (r *AuthRepository) CreateMemberWithKeyAndInitialFees(ctx context.Context, in UserCreateInput, fees H2HInitialFeeSetupInput) (int64, string, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = tx.Rollback() }()

	var existingID int64
	dupErr := tx.QueryRowContext(ctx, `
SELECT id
FROM public.member
WHERE LOWER(email) = LOWER($1)
LIMIT 1
`, in.Email).Scan(&existingID)
	if dupErr == nil {
		return 0, "", fmt.Errorf("email sudah terdaftar")
	}
	if dupErr != sql.ErrNoRows {
		return 0, "", dupErr
	}

	var memberID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (
  email, nama, phone, password_hash, pin_hash, role, aktif,
  retail_agent_commission_rp, retail_master_commission_rp,
  h2h_agent_commission_rp, h2h_master_commission_rp,
  h2h_agent_member_id, h2h_master_member_id
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING id
`, in.Email, nullIfEmptyStr(in.Nama), nullIfEmptyStr(in.Phone), in.PasswordHash, in.PinHash, in.Role, in.Aktif, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp, in.H2HAgentCommissionRp, in.H2HMasterCommissionRp, nullableInt64Value(in.H2HAgentID), nullableInt64Value(in.H2HMasterID)).Scan(&memberID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "member_email_key") {
			return 0, "", fmt.Errorf("email sudah terdaftar")
		}
		return 0, "", fmt.Errorf("create member failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return 0, "", fmt.Errorf("create dompet failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.member_api_key (member_id, api_key, label, aktif)
VALUES ($1,$2,$3,true)
`, memberID, in.APIKey, nullIfEmptyStr(in.Label))
	if err != nil {
		return 0, "", fmt.Errorf("create api_key failed: %w", err)
	}

	seedH2HFee := func(code string, fee int64) error {
		_, execErr := tx.ExecContext(ctx, `
INSERT INTO public.member_h2h_fee
  (member_id, fee_code, fee_rp, aktif, dibuat_pada, diubah_pada)
VALUES
  ($1, $2, $3, true, now(), now())
ON CONFLICT (member_id, fee_code)
DO UPDATE SET
  fee_rp = EXCLUDED.fee_rp,
  aktif = true,
  diubah_pada = now()
`, memberID, code, fee)
		if execErr != nil {
			return execErr
		}
		return nil
	}

	if err := seedH2HFee(helper.H2HFeeCategoryDana, fees.FeeDana); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryDana, err)
	}
	if err := seedH2HFee(helper.H2HFeeCategoryGopay, fees.FeeGopay); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryGopay, err)
	}
	if err := seedH2HFee(helper.H2HFeeCategoryLinkAja, fees.FeeLinkAja); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryLinkAja, err)
	}
	if err := seedH2HFee(helper.H2HFeeCategoryOvo, fees.FeeOvo); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryOvo, err)
	}
	if err := seedH2HFee(helper.H2HFeeCategoryShopeePay, fees.FeeShopee); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryShopeePay, err)
	}
	if err := seedH2HFee(helper.H2HFeeCategoryBank, fees.FeeBank); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryBank, err)
	}
	if err := seedH2HFee(helper.H2HFeeCategoryLainnya, fees.FeeLainnya); err != nil {
		return 0, "", fmt.Errorf("seed member_h2h_fee %s failed: %w", helper.H2HFeeCategoryLainnya, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return memberID, in.APIKey, nil
}

func (r *AuthRepository) CreateInternalUser(ctx context.Context, in UserCreateInput) (int64, string, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = tx.Rollback() }()

	var memberID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (
	email, nama, phone, password_hash, pin_hash, role, aktif,
	retail_agent_commission_rp, retail_master_commission_rp,
	h2h_agent_commission_rp, h2h_master_commission_rp
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
RETURNING id
`, in.Email, nullIfEmptyStr(in.Nama), nullIfEmptyStr(in.Phone), in.PasswordHash, nullIfEmptyStr(in.PinHash), in.Role, in.Aktif, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp, in.H2HAgentCommissionRp, in.H2HMasterCommissionRp).Scan(&memberID)
	if err != nil {
		return 0, "", fmt.Errorf("create internal user failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return 0, "", fmt.Errorf("create dompet failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return memberID, "", nil
}

func (r *AuthRepository) CreatePublicUser(ctx context.Context, email, nama, phone, passwordHash string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var memberID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (
	email, nama, phone, password_hash, role, aktif,
	retail_agent_commission_rp, retail_master_commission_rp,
	h2h_agent_commission_rp, h2h_master_commission_rp
)
VALUES ($1,$2,$3,$4,'user',true,0,0,0,0)
RETURNING id
`, strings.TrimSpace(strings.ToLower(email)), nullIfEmptyStr(strings.TrimSpace(nama)), nullIfEmptyStr(strings.TrimSpace(phone)), strings.TrimSpace(passwordHash)).Scan(&memberID)
	if err != nil {
		return 0, fmt.Errorf("create public user failed: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return 0, fmt.Errorf("create dompet failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return memberID, nil
}
