package db

import (
	"context"
	"database/sql"
	"errors"
)

type MemberProfile struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	Nama       string `json:"nama"`
	Aktif      bool   `json:"aktif"`
	Saldo      int64  `json:"saldo"`
	DibuatPada string `json:"dibuat_pada"`
}

type MemberApiKey struct {
	ID         int64  `json:"id"`
	MemberID   int64  `json:"member_id"`
	ApiKey     string `json:"api_key"`
	Aktif      bool   `json:"aktif"`
	DibuatPada string `json:"dibuat_pada"`
}

func (r *MemberRepo) GetMemberProfile(ctx context.Context, memberID int64) (MemberProfile, error) {
	if memberID <= 0 {
		return MemberProfile{}, errors.New("invalid member_id")
	}

	const q = `
SELECT
  m.id,
  m.email,
  m.nama,
  m.aktif,
  COALESCE(d.saldo, 0) AS saldo,
  m.dibuat_pada
FROM member m
LEFT JOIN dompet_member d ON d.member_id = m.id
WHERE m.id = $1
LIMIT 1
`
	var p MemberProfile
	err := r.DB.QueryRowContext(ctx, q, memberID).Scan(
		&p.ID,
		&p.Email,
		&p.Nama,
		&p.Aktif,
		&p.Saldo,
		&p.DibuatPada,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return MemberProfile{}, errors.New("member not found")
		}
		return MemberProfile{}, err
	}
	return p, nil
}

func (r *MemberRepo) ListApiKeys(ctx context.Context, memberID int64) ([]MemberApiKey, error) {
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}

	const q = `
SELECT id, member_id, api_key, aktif, dibuat_pada
FROM member_api_key
WHERE member_id = $1
ORDER BY dibuat_pada DESC
`
	rows, err := r.DB.QueryContext(ctx, q, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemberApiKey, 0, 8)
	for rows.Next() {
		var k MemberApiKey
		if err := rows.Scan(&k.ID, &k.MemberID, &k.ApiKey, &k.Aktif, &k.DibuatPada); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}
