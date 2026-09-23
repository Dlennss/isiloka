package db

import (
  "context"
  "database/sql"
  "errors"

  "golang.org/x/crypto/bcrypt"
)

func (r *MemberRepo) MemberChangePassword(ctx context.Context, memberID int64, oldPassword string, newPasswordHash string) error {
  if memberID <= 0 || oldPassword == "" || newPasswordHash == "" {
    return errors.New("invalid payload")
  }

  var oldHash sql.NullString
  err := r.DB.QueryRowContext(ctx, `SELECT password_hash FROM public.member WHERE id=$1`, memberID).Scan(&oldHash)
  if err != nil {
    return err
  }
  if !oldHash.Valid || oldHash.String == "" {
    return errors.New("password belum diset")
  }
  if bcrypt.CompareHashAndPassword([]byte(oldHash.String), []byte(oldPassword)) != nil {
    return errors.New("old_password salah")
  }

  _, err = r.DB.ExecContext(ctx, `UPDATE public.member SET password_hash=$1 WHERE id=$2`, newPasswordHash, memberID)
  return err
}

func (r *MemberRepo) MemberChangePIN(ctx context.Context, memberID int64, oldPIN string, newPinHash string) error {
  if memberID <= 0 || oldPIN == "" || newPinHash == "" {
    return errors.New("invalid payload")
  }

  var oldHash sql.NullString
  err := r.DB.QueryRowContext(ctx, `SELECT pin_hash FROM public.member WHERE id=$1`, memberID).Scan(&oldHash)
  if err != nil {
    return err
  }
  if !oldHash.Valid || oldHash.String == "" {
    return errors.New("pin belum diset")
  }
  if bcrypt.CompareHashAndPassword([]byte(oldHash.String), []byte(oldPIN)) != nil {
    return errors.New("old_pin salah")
  }

  _, err = r.DB.ExecContext(ctx, `UPDATE public.member SET pin_hash=$1 WHERE id=$2`, newPinHash, memberID)
  return err
}
