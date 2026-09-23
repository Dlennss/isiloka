package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type H2HRepository struct {
	db *sql.DB
}

func NewH2HRepository(db *sql.DB) *H2HRepository {
	return &H2HRepository{db: db}
}

func (r *H2HRepository) GetMemberContext(ctx context.Context, memberID int64) (*RetailMemberContextRow, error) {
	if memberID <= 0 {
		return nil, sql.ErrNoRows
	}

	var (
		row            RetailMemberContextRow
		retailAgentID  sql.NullInt64
		retailMasterID sql.NullInt64
		h2hAgentID     sql.NullInt64
		h2hMasterID    sql.NullInt64
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  m.id, COALESCE(m.email, ''), COALESCE(m.nama, ''), COALESCE(m.role, ''), m.aktif, COALESCE(d.saldo, 0),
  m.retail_agent_id, m.retail_master_id, m.h2h_agent_member_id, m.h2h_master_member_id
FROM public.member m
LEFT JOIN public.dompet_member d ON d.member_id = m.id
WHERE m.id = $1
LIMIT 1
`, memberID).Scan(
		&row.MemberID, &row.Email, &row.Nama, &row.Role, &row.Aktif, &row.Saldo,
		&retailAgentID, &retailMasterID, &h2hAgentID, &h2hMasterID,
	)
	if err != nil {
		return nil, err
	}
	if retailAgentID.Valid {
		v := retailAgentID.Int64
		row.RetailAgentID = &v
	}
	if retailMasterID.Valid {
		v := retailMasterID.Int64
		row.RetailMasterID = &v
	}
	if h2hAgentID.Valid {
		v := h2hAgentID.Int64
		row.H2HAgentID = &v
	}
	if h2hMasterID.Valid {
		v := h2hMasterID.Int64
		row.H2HMasterID = &v
	}
	return &row, nil
}

func (r *H2HRepository) ListDownlines(ctx context.Context, actor *RetailMemberContextRow) ([]H2HDownlineRow, error) {
	if actor == nil || actor.MemberID <= 0 {
		return nil, sql.ErrNoRows
	}

	where := "false"
	switch strings.TrimSpace(strings.ToLower(actor.Role)) {
	case "master_member":
		where = "(m.h2h_master_member_id = $1 OR m.h2h_agent_member_id = $1)"
	case "agent_member":
		where = "m.h2h_agent_member_id = $1"
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  m.id, COALESCE(m.email, ''), COALESCE(m.nama, ''), COALESCE(m.role, ''), m.aktif, COALESCE(d.saldo, 0),
  m.h2h_agent_member_id, COALESCE(ha.nama, ''), m.h2h_master_member_id, COALESCE(hm.nama, ''), m.dibuat_pada
FROM public.member m
LEFT JOIN public.dompet_member d ON d.member_id = m.id
LEFT JOIN public.member ha ON ha.id = m.h2h_agent_member_id
LEFT JOIN public.member hm ON hm.id = m.h2h_master_member_id
WHERE %s
  AND lower(COALESCE(m.role, '')) IN ('member', 'agent_member')
ORDER BY CASE lower(COALESCE(m.role, '')) WHEN 'agent_member' THEN 0 ELSE 1 END, m.id DESC
`, where), actor.MemberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]H2HDownlineRow, 0, 32)
	for rows.Next() {
		var (
			item          H2HDownlineRow
			h2hAgentID    sql.NullInt64
			h2hMasterID   sql.NullInt64
			h2hAgentNama  sql.NullString
			h2hMasterNama sql.NullString
			dibuat        sql.NullTime
		)
		if err := rows.Scan(
			&item.ID, &item.Email, &item.Nama, &item.Role, &item.Aktif, &item.Saldo,
			&h2hAgentID, &h2hAgentNama, &h2hMasterID, &h2hMasterNama, &dibuat,
		); err != nil {
			return nil, err
		}
		if h2hAgentID.Valid {
			v := h2hAgentID.Int64
			item.H2HAgentID = &v
		}
		if h2hMasterID.Valid {
			v := h2hMasterID.Int64
			item.H2HMasterID = &v
		}
		if h2hAgentNama.Valid && strings.TrimSpace(h2hAgentNama.String) != "" {
			v := h2hAgentNama.String
			item.H2HAgentNama = &v
		}
		if h2hMasterNama.Valid && strings.TrimSpace(h2hMasterNama.String) != "" {
			v := h2hMasterNama.String
			item.H2HMasterNama = &v
		}
		if dibuat.Valid {
			v := dibuat.Time
			item.DibuatPada = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
