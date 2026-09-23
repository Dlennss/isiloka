package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (r *AuthRepository) GetForLogin(ctx context.Context, email string) (*AuthLoginRow, error) {
	const q = `
SELECT id, email, COALESCE(nama,''), role, aktif, COALESCE(password_hash,'')
FROM public.member
WHERE email = $1
LIMIT 1`
	var out AuthLoginRow
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&out.ID, &out.Email, &out.Nama, &out.Role, &out.Aktif, &out.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AuthRepository) GetMe(ctx context.Context, memberID int64) (*AuthMeRow, error) {
	const q = `
SELECT id, email, COALESCE(nama,''), COALESCE(phone,''), role, aktif
FROM public.member
WHERE id = $1
LIMIT 1`
	var out AuthMeRow
	err := r.db.QueryRowContext(ctx, q, memberID).Scan(
		&out.ID, &out.Email, &out.Nama, &out.Phone, &out.Role, &out.Aktif,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AuthRepository) GetByEmail(ctx context.Context, email string) (*AuthGoogleRow, error) {
	const q = `
SELECT id, role, aktif, COALESCE(google_sub, ''), COALESCE(apple_sub, '')
FROM public.member
WHERE LOWER(email) = LOWER($1)
LIMIT 1`
	var out AuthGoogleRow
	err := r.db.QueryRowContext(ctx, q, strings.TrimSpace(email)).Scan(
		&out.ID,
		&out.Role,
		&out.Aktif,
		&out.GoogleSub,
		&out.AppleSub,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AuthRepository) GetByGoogleSub(ctx context.Context, googleSub string) (*AuthGoogleRow, error) {
	const q = `
SELECT id, role, aktif, COALESCE(google_sub, '')
FROM public.member
WHERE google_sub = $1
LIMIT 1`
	var out AuthGoogleRow
	err := r.db.QueryRowContext(ctx, q, strings.TrimSpace(googleSub)).Scan(&out.ID, &out.Role, &out.Aktif, &out.GoogleSub)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AuthRepository) LinkGoogleSubByMemberID(ctx context.Context, memberID int64, googleSub string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET google_sub = $2
WHERE id = $1
  AND (google_sub IS NULL OR google_sub = '')
`, memberID, strings.TrimSpace(googleSub))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("google account already linked to another id")
	}
	return nil
}
