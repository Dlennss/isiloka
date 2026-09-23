package db

import (
  "context"
  "database/sql"
)

type MemberProfileRow struct {
  ID    int64  `json:"id"`
  Email string `json:"email"`
  Nama  string `json:"nama"`
  Role  string `json:"role"`
  Aktif bool   `json:"aktif"`
}

func (r *MemberRepo) MemberProfile(ctx context.Context, memberID int64) (*MemberProfileRow, error) {
  var out MemberProfileRow
  var nama sql.NullString
  err := r.DB.QueryRowContext(ctx, `
SELECT id, email, nama, role, aktif
FROM public.member
WHERE id=$1
`, memberID).Scan(&out.ID, &out.Email, &nama, &out.Role, &out.Aktif)
  if err != nil {
    return nil, err
  }
  if nama.Valid {
    out.Nama = nama.String
  }
  return &out, nil
}
