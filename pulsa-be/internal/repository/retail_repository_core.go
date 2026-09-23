package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type RetailRepository struct {
	db *sql.DB
}

func NewRetailRepository(db *sql.DB) *RetailRepository {
	return &RetailRepository{db: db}
}

// EnsureWithdrawSchema keeps production databases upgraded before a credit
// withdrawal is accepted. Older installations did not have the funding-source
// columns, which made PostgreSQL errors appear as a generic client error.
func (r *RetailRepository) EnsureWithdrawSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("retail repository tidak tersedia")
	}
	_, err := r.db.ExecContext(ctx, `
ALTER TABLE public.retail_withdraw_request
  ADD COLUMN IF NOT EXISTS bank_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS account_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS account_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS note TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS reject_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS ref_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS processed_by BIGINT REFERENCES public.member(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS processed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS source_type TEXT NOT NULL DEFAULT 'main_balance',
  ADD COLUMN IF NOT EXISTS credit_loan_id BIGINT REFERENCES public.agent_credit_loan(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS credit_application_id BIGINT REFERENCES public.agent_credit_application(id) ON DELETE SET NULL;

UPDATE public.retail_withdraw_request
SET source_type = 'main_balance'
WHERE source_type IS NULL OR source_type NOT IN ('main_balance', 'credit');

ALTER TABLE public.retail_withdraw_request
  DROP CONSTRAINT IF EXISTS retail_withdraw_request_source_type_check;

ALTER TABLE public.retail_withdraw_request
  ADD CONSTRAINT retail_withdraw_request_source_type_check
  CHECK (source_type IN ('main_balance', 'credit'));

CREATE INDEX IF NOT EXISTS retail_withdraw_request_source_status_idx
  ON public.retail_withdraw_request(source_type, status, created_at DESC);
`)
	return err
}

func (r *RetailRepository) GetMemberContext(ctx context.Context, memberID int64) (*RetailMemberContextRow, error) {
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
  m.id,
  COALESCE(m.email, ''),
  COALESCE(m.nama, ''),
  COALESCE(m.role, ''),
  m.aktif,
  COALESCE(d.saldo, 0),
  m.retail_agent_id,
  m.retail_master_id,
  m.h2h_agent_member_id,
  m.h2h_master_member_id
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

func (r *RetailRepository) CreateRetailChild(ctx context.Context, in UserCreateInput) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
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
		return 0, fmt.Errorf("email sudah terdaftar")
	}
	if dupErr != sql.ErrNoRows {
		return 0, dupErr
	}

	var memberID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.member (
  email, nama, phone, store_name, password_hash, role, aktif,
  retail_agent_commission_rp, retail_master_commission_rp,
  h2h_agent_commission_rp, h2h_master_commission_rp,
  retail_agent_id, retail_master_id, h2h_agent_member_id, h2h_master_member_id, marketing_id
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
RETURNING id
`,
		in.Email, retailNullString(in.Nama), retailNullString(in.Phone), retailNullString(in.StoreName), in.PasswordHash, in.Role, in.Aktif,
		in.RetailAgentCommissionRp, in.RetailMasterCommissionRp,
		in.H2HAgentCommissionRp, in.H2HMasterCommissionRp,
		retailNullableInt64(in.RetailAgentID), retailNullableInt64(in.RetailMasterID),
		retailNullableInt64(in.H2HAgentID), retailNullableInt64(in.H2HMasterID),
		retailNullableInt64(in.MarketingID),
	).Scan(&memberID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "member_email_key") {
			return 0, fmt.Errorf("email sudah terdaftar")
		}
		return 0, fmt.Errorf("create retail child failed: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID); err != nil {
		return 0, fmt.Errorf("create dompet failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return memberID, nil
}

func (r *RetailRepository) ListDownlines(ctx context.Context, actor *RetailMemberContextRow) ([]RetailDownlineRow, error) {
	if actor == nil || actor.MemberID <= 0 {
		return nil, sql.ErrNoRows
	}

	where := "false"
	switch strings.TrimSpace(strings.ToLower(actor.Role)) {
	case "master":
		where = "(m.retail_master_id = $1 OR m.retail_agent_id = $1)"
	case "agent":
		where = "m.retail_agent_id = $1"
	case "marketing":
		where = "m.marketing_id = $1"
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  m.id, COALESCE(m.email, ''), COALESCE(m.nama, ''), COALESCE(m.phone, ''), COALESCE(m.store_name, ''), COALESCE(m.role, ''), m.aktif, COALESCE(d.saldo, 0),
  m.retail_agent_id, COALESCE(ra.nama, ''), m.retail_master_id, COALESCE(rm.nama, ''),
  m.marketing_id, COALESCE(marketing.nama, ''), COALESCE(marketing.email, ''), m.dibuat_pada
FROM public.member m
LEFT JOIN public.dompet_member d ON d.member_id = m.id
LEFT JOIN public.member ra ON ra.id = m.retail_agent_id
LEFT JOIN public.member rm ON rm.id = m.retail_master_id
LEFT JOIN public.member marketing ON marketing.id = m.marketing_id
WHERE %s
  AND lower(COALESCE(m.role, '')) IN ('user', 'agent')
ORDER BY CASE lower(COALESCE(m.role, '')) WHEN 'agent' THEN 0 ELSE 1 END, m.id DESC
`, where), actor.MemberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RetailDownlineRow, 0, 32)
	for rows.Next() {
		var (
			item             RetailDownlineRow
			retailAgentID    sql.NullInt64
			retailMasterID   sql.NullInt64
			retailAgentNama  sql.NullString
			retailMasterNama sql.NullString
			marketingID      sql.NullInt64
			marketingNama    sql.NullString
			marketingEmail   sql.NullString
			dibuat           sql.NullTime
		)
		if err := rows.Scan(
			&item.ID, &item.Email, &item.Nama, &item.Phone, &item.StoreName, &item.Role, &item.Aktif, &item.Saldo,
			&retailAgentID, &retailAgentNama, &retailMasterID, &retailMasterNama,
			&marketingID, &marketingNama, &marketingEmail, &dibuat,
		); err != nil {
			return nil, err
		}
		if retailAgentID.Valid {
			v := retailAgentID.Int64
			item.RetailAgentID = &v
		}
		if retailMasterID.Valid {
			v := retailMasterID.Int64
			item.RetailMasterID = &v
		}
		if retailAgentNama.Valid && strings.TrimSpace(retailAgentNama.String) != "" {
			v := retailAgentNama.String
			item.RetailAgentNama = &v
		}
		if retailMasterNama.Valid && strings.TrimSpace(retailMasterNama.String) != "" {
			v := retailMasterNama.String
			item.RetailMasterNama = &v
		}
		if marketingID.Valid {
			v := marketingID.Int64
			item.MarketingID = &v
		}
		if marketingNama.Valid && strings.TrimSpace(marketingNama.String) != "" {
			v := marketingNama.String
			item.MarketingNama = &v
		}
		if marketingEmail.Valid && strings.TrimSpace(marketingEmail.String) != "" {
			v := marketingEmail.String
			item.MarketingEmail = &v
		}
		if dibuat.Valid {
			v := dibuat.Time
			item.DibuatPada = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func retailNullableInt64(v *int64) any {
	if v == nil || *v <= 0 {
		return nil
	}
	return *v
}

func retailNullString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.TrimSpace(v)
}
