package db

import (
	"context"
	"database/sql"
	"errors"
)

// Reset API key: nonaktifkan semua key aktif lalu insert key baru aktif.
func (r *MemberRepo) MemberResetAPIKey(ctx context.Context, memberID int64, apiKey string, label string) error {
  if memberID <= 0 || apiKey == "" {
    return errors.New("invalid payload")
  }

  tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
  if err != nil {
    return err
  }
  defer func() { _ = tx.Rollback() }()

  // nonaktifkan key lama
  _, err = tx.ExecContext(ctx, `UPDATE public.member_api_key SET aktif=false WHERE member_id=$1`, memberID)
  if err != nil {
    return err
  }

  _, err = tx.ExecContext(ctx, `
INSERT INTO public.member_api_key (member_id, api_key, label, aktif)
VALUES ($1,$2,NULLIF($3,''),true)
`, memberID, apiKey, label)
  if err != nil {
    return err
  }

  return tx.Commit()
}
