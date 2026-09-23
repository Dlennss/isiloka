package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (r *AuthRepository) GetByAppleSub(ctx context.Context, appleSub string) (*AuthGoogleRow, error) {
	const q = `
SELECT id, role, aktif, COALESCE(apple_sub, '')
FROM public.member
WHERE apple_sub = $1
LIMIT 1`
	var out AuthGoogleRow
	err := r.db.QueryRowContext(ctx, q, strings.TrimSpace(appleSub)).Scan(&out.ID, &out.Role, &out.Aktif, &out.AppleSub)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AuthRepository) LinkAppleSubByMemberID(ctx context.Context, memberID int64, appleSub string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET apple_sub = $2
WHERE id = $1
  AND (apple_sub IS NULL OR apple_sub = '')
`, memberID, strings.TrimSpace(appleSub))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("apple account already linked to another id")
	}
	return nil
}

func (r *AuthRepository) DeactivateByAppleSub(ctx context.Context, appleSub string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE public.member SET aktif = false
WHERE apple_sub = $1 AND apple_sub != ''
`, strings.TrimSpace(appleSub))
	return err
}

func (r *AuthRepository) CreateAppleMember(ctx context.Context, email, nama, appleSub string) (*AuthGoogleRow, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var memberID int64
	var role string
	var aktif bool
	var outAppleSub string

	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (email, nama, apple_sub, role, aktif)
VALUES ($1, NULLIF($2,''), $3, 'user', true)
ON CONFLICT (email) DO UPDATE SET
  apple_sub = CASE WHEN COALESCE(member.apple_sub, '') = '' THEN EXCLUDED.apple_sub ELSE member.apple_sub END
RETURNING id, role, aktif, COALESCE(apple_sub, '')
`, strings.TrimSpace(strings.ToLower(email)), nullIfEmptyStr(strings.TrimSpace(nama)), strings.TrimSpace(appleSub)).Scan(&memberID, &role, &aktif, &outAppleSub)
	if err != nil {
		return nil, fmt.Errorf("create apple member failed: %w", err)
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
		ID:       memberID,
		Role:     role,
		Aktif:    aktif,
		AppleSub: outAppleSub,
	}, nil
}
