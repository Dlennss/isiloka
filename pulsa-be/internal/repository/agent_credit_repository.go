package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AgentCreditRepository struct {
	db *sql.DB
}

func NewAgentCreditRepository(db *sql.DB) *AgentCreditRepository {
	return &AgentCreditRepository{db: db}
}

// EnsureFlexibleLimitSchema keeps older production databases compatible with
// the credit list queries. Without these columns the submit response can look
// successful, while every reload of the credit page fails during the SELECT.
func (r *AgentCreditRepository) EnsureFlexibleLimitSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("agent credit repository tidak tersedia")
	}
	_, err := r.db.ExecContext(ctx, `
ALTER TABLE public.agent_credit_rank_history
  ADD COLUMN IF NOT EXISTS custom_limit_amount BIGINT,
  ADD COLUMN IF NOT EXISTS custom_limit_name TEXT NOT NULL DEFAULT '';

ALTER TABLE public.agent_credit_rank_history
  DROP CONSTRAINT IF EXISTS agent_credit_rank_history_custom_limit_check;

ALTER TABLE public.agent_credit_rank_history
  ADD CONSTRAINT agent_credit_rank_history_custom_limit_check
  CHECK (custom_limit_amount IS NULL OR custom_limit_amount > 0);
`)
	return err
}

type AgentCreditApplicationInput struct {
	ID              int64
	MemberID        int64
	MarketingID     int64
	RequestedAmount int64
	ApplicantData   map[string]any
	DocumentData    map[string]any
	AgentSignature  string
}

type AgentCreditPaymentInput struct {
	ApplicationID int64
	MemberID      int64
	RefID         string
	Amount        int64
	Note          string
	PaymentMethod string
	PaymentProof  map[string]any
}

type AgentCreditDecisionInput struct {
	ID             int64
	MarketingID    int64
	AnalystID      int64
	MasterID       int64
	Status         string
	ApprovedAmount int64
	MarketingNote  string
	AnalystNote    string
	Recommendation string
	ActorRole      string
}

type AgentCreditLoanStatusResult struct {
	ApplicationID         int64  `json:"application_id"`
	MemberID              int64  `json:"member_id"`
	Status                string `json:"status"`
	CreditAvailableAmount int64  `json:"credit_available_amount"`
	OutstandingAmount     int64  `json:"outstanding_amount"`
}

type AgentCreditTeamActivity struct {
	ID         int64     `json:"id"`
	ActorID    int64     `json:"actor_id"`
	ActorName  string    `json:"actor_name"`
	ActorEmail string    `json:"actor_email"`
	ActorRole  string    `json:"actor_role"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

type AgentCreditInactiveAgent struct {
	MemberID          int64      `json:"member_id"`
	MemberName        string     `json:"member_name"`
	MemberEmail       string     `json:"member_email"`
	MemberPhone       string     `json:"member_phone"`
	MarketingID       int64      `json:"marketing_id"`
	MarketingName     string     `json:"marketing_name"`
	MarketingEmail    string     `json:"marketing_email"`
	LastTransactionAt *time.Time `json:"last_transaction_at,omitempty"`
	LastProduct       string     `json:"last_product"`
	LastStatus        string     `json:"last_status"`
	LastAmount        int64      `json:"last_amount"`
	InactiveDays      int        `json:"inactive_days"`
}

func (r *AgentCreditRepository) ListInactiveAgents(ctx context.Context, days int, search string, limit int) ([]AgentCreditInactiveAgent, error) {
	if days < 1 {
		days = 3
	}
	if days > 365 {
		days = 365
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	search = strings.TrimSpace(search)

	rows, err := r.db.QueryContext(ctx, `
SELECT
  m.id,
  COALESCE(m.nama, ''),
  COALESCE(m.email, ''),
  COALESCE(m.phone, ''),
  COALESCE(m.marketing_id, 0),
  COALESCE(mark.nama, ''),
  COALESCE(mark.email, ''),
  last_trx.dibuat_pada,
  COALESCE(last_trx.kode_produk, ''),
  COALESCE(last_trx.status, ''),
  COALESCE(last_trx.nominal, 0),
  CASE
    WHEN COALESCE(last_trx.dibuat_pada, m.dibuat_pada) IS NULL THEN 0
    ELSE GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (NOW() - COALESCE(last_trx.dibuat_pada, m.dibuat_pada))) / 86400)::int)
  END AS inactive_days
FROM public.member m
LEFT JOIN public.member mark ON mark.id = m.marketing_id
LEFT JOIN LATERAL (
  SELECT trx.dibuat_pada, trx.kode_produk, trx.status, trx.nominal
  FROM (
    SELECT ao.dibuat_pada, ao.produk_sku_snapshot AS kode_produk, ao.status,
           COALESCE(ao.harga_final, 0) AS nominal, ao.id
    FROM public.app_order ao
    WHERE ao.member_id = m.id AND LOWER(COALESCE(ao.buyer_type, '')) = 'user'
    UNION ALL
    SELECT tm.dibuat_pada, tm.kode_produk, tm.status,
           COALESCE(tm.biaya_aktual, tm.biaya_perkiraan, 0) AS nominal, tm.id
    FROM public.transaksi_member tm
    WHERE tm.member_id = m.id
  ) trx
  ORDER BY trx.dibuat_pada DESC, trx.id DESC
  LIMIT 1
) last_trx ON true
WHERE LOWER(COALESCE(m.role, '')) = 'agent'
  AND EXISTS (
    SELECT 1
    FROM public.agent_credit_loan active_loan
    WHERE active_loan.member_id = m.id
      AND active_loan.status IN ('active', 'due', 'overdue', 'suspended')
  )
  AND COALESCE(last_trx.dibuat_pada, m.dibuat_pada) < NOW() - ($1 * INTERVAL '1 day')
  AND ($2 = '' OR m.nama ILIKE '%' || $2 || '%' OR m.email ILIKE '%' || $2 || '%' OR m.phone ILIKE '%' || $2 || '%' OR mark.nama ILIKE '%' || $2 || '%')
ORDER BY last_trx.dibuat_pada ASC NULLS FIRST, m.id DESC
LIMIT $3
`, days, search, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentCreditInactiveAgent, 0)
	for rows.Next() {
		var item AgentCreditInactiveAgent
		if err := rows.Scan(&item.MemberID, &item.MemberName, &item.MemberEmail, &item.MemberPhone, &item.MarketingID, &item.MarketingName, &item.MarketingEmail, &item.LastTransactionAt, &item.LastProduct, &item.LastStatus, &item.LastAmount, &item.InactiveDays); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AgentCreditRepository) ListTeamActivity(ctx context.Context, actorRole string, limit int) ([]AgentCreditTeamActivity, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT
  a.id,
  COALESCE(a.actor_id, 0),
  COALESCE(m.nama, ''),
  COALESCE(m.email, ''),
  a.actor_role,
  a.action,
  a.entity_type,
  a.entity_id,
  a.reason,
  a.created_at
FROM public.audit_log a
LEFT JOIN public.member m ON m.id = a.actor_id
WHERE a.actor_role IN ('marketing', 'master', 'operator_credit', 'analis', 'super_admin', 'admin')
  AND (
    $1 = ''
    OR ($1 = 'marketing' AND a.actor_role IN ('marketing', 'master'))
    OR ($1 = 'operator_credit' AND a.actor_role IN ('operator_credit', 'analis'))
    OR ($1 = 'super_admin' AND a.actor_role IN ('super_admin', 'admin'))
  )
ORDER BY a.created_at DESC, a.id DESC
LIMIT $2
`, actorRole, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentCreditTeamActivity, 0)
	for rows.Next() {
		var item AgentCreditTeamActivity
		if err := rows.Scan(&item.ID, &item.ActorID, &item.ActorName, &item.ActorEmail, &item.ActorRole, &item.Action, &item.EntityType, &item.EntityID, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type AgentCreditProfile struct {
	LevelCode          string `json:"credit_level_code"`
	LevelName          string `json:"credit_level_name"`
	LimitAmount        int64  `json:"credit_limit_amount"`
	NeedsRepair        bool   `json:"credit_needs_repair"`
	QualifiedPaidTotal int64  `json:"qualified_paid_total"`
	QualifiedPaidCount int64  `json:"qualified_paid_count"`
}

type AgentCreditManualAgent struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	StoreName      string `json:"store_name"`
	MarketingID    int64  `json:"marketing_id"`
	Marketing      string `json:"marketing_name"`
	MarketingEmail string `json:"marketing_email"`
}

func (r *AgentCreditRepository) ListManualEntryAgents(ctx context.Context, search string, limit int) ([]AgentCreditManualAgent, error) {
	search = strings.TrimSpace(search)
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT m.id, COALESCE(m.nama, ''), COALESCE(m.email, ''), COALESCE(m.phone, ''), COALESCE(m.store_name, ''),
       COALESCE(m.marketing_id, 0), COALESCE(marketing.nama, ''), COALESCE(marketing.email, '')
FROM public.member m
LEFT JOIN public.member marketing ON marketing.id = m.marketing_id
WHERE LOWER(COALESCE(m.role, '')) = 'agent'
  AND COALESCE(m.aktif, TRUE) = TRUE
  AND ($1 = '' OR m.nama ILIKE '%' || $1 || '%' OR m.email ILIKE '%' || $1 || '%' OR m.phone ILIKE '%' || $1 || '%')
ORDER BY m.nama ASC, m.id DESC
LIMIT $2
`, search, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AgentCreditManualAgent, 0)
	for rows.Next() {
		var item AgentCreditManualAgent
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Phone, &item.StoreName, &item.MarketingID, &item.Marketing, &item.MarketingEmail); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AgentCreditRepository) UpdateManualApplication(ctx context.Context, applicationID, memberID, requestedAmount int64, applicantData, documentData map[string]any) error {
	applicantJSON, err := json.Marshal(applicantData)
	if err != nil {
		return err
	}
	documentJSON, err := json.Marshal(documentData)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
UPDATE public.agent_credit_application
SET requested_amount = $3,
    applicant_data = COALESCE(applicant_data, '{}'::jsonb) || $4::jsonb,
    document_data = COALESCE(document_data, '{}'::jsonb) || $5::jsonb,
    updated_at = now()
WHERE id = $1 AND member_id = $2
`, applicationID, memberID, requestedAmount, string(applicantJSON), string(documentJSON))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type AgentCreditApplication struct {
	ID                       int64                `json:"id"`
	MemberID                 int64                `json:"member_id"`
	MemberName               string               `json:"member_name"`
	MemberEmail              string               `json:"member_email"`
	MemberPhone              string               `json:"member_phone"`
	MarketingID              int64                `json:"marketing_id"`
	AnalystID                int64                `json:"analyst_id"`
	RequestedAmount          int64                `json:"requested_amount"`
	ApprovedAmount           int64                `json:"approved_amount"`
	Status                   string               `json:"status"`
	ApplicantData            map[string]any       `json:"applicant_data"`
	DocumentData             map[string]any       `json:"document_data"`
	HasAgentSignature        bool                 `json:"has_agent_signature"`
	AgentSignatureData       string               `json:"agent_signature_data,omitempty"`
	AgentSignatureAt         *time.Time           `json:"agent_signature_at,omitempty"`
	MarketingNote            string               `json:"marketing_note"`
	AnalystNote              string               `json:"analyst_note"`
	AnalystRecommendation    string               `json:"analyst_recommendation"`
	AnalystRecommendedAmount int64                `json:"analyst_recommended_amount"`
	LoanStatus               string               `json:"loan_status"`
	OutstandingAmount        int64                `json:"outstanding_amount"`
	CreditAvailableAmount    int64                `json:"credit_available_amount"`
	PaidAmount               int64                `json:"paid_amount"`
	PaymentCount             int64                `json:"payment_count"`
	Payments                 []AgentCreditPayment `json:"payments,omitempty"`
	CreditLevelCode          string               `json:"credit_level_code"`
	CreditLevelName          string               `json:"credit_level_name"`
	CreditNeedsRepair        bool                 `json:"credit_needs_repair"`
	QualifiedPaidTotal       int64                `json:"qualified_paid_total"`
	CreditLimitAmount        int64                `json:"credit_limit_amount"`
	LoanApprovedAt           *time.Time           `json:"loan_approved_at,omitempty"`
	LoanDueDate              *time.Time           `json:"loan_due_date,omitempty"`
	LastTransactionAt        *time.Time           `json:"last_transaction_at,omitempty"`
	CreatedAt                time.Time            `json:"created_at"`
	UpdatedAt                time.Time            `json:"updated_at"`
}

type AgentCreditPayment struct {
	ID            int64          `json:"id"`
	LoanID        int64          `json:"loan_id"`
	ApplicationID int64          `json:"application_id"`
	MemberID      int64          `json:"member_id"`
	Amount        int64          `json:"amount"`
	DueDate       time.Time      `json:"due_date"`
	PaidAt        time.Time      `json:"paid_at"`
	DaysLate      int64          `json:"days_late"`
	Status        string         `json:"status"`
	Note          string         `json:"note"`
	PaymentMethod string         `json:"payment_method"`
	PaymentProof  map[string]any `json:"payment_proof,omitempty"`
}

type AgentCreditTransferResult struct {
	ApplicationID         int64  `json:"application_id"`
	Amount                int64  `json:"amount"`
	CreditAvailableAmount int64  `json:"credit_available_amount"`
	OutstandingAmount     int64  `json:"outstanding_amount"`
	MainBalance           int64  `json:"main_balance"`
	RefID                 string `json:"ref_id"`
}

type AgentCreditRank struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LimitAmount int64  `json:"limit_amount"`
	SortOrder   int64  `json:"sort_order"`
}

type AgentCreditReviewState struct {
	Status                   string
	AnalystRecommendation    string
	AnalystRecommendedAmount int64
	HasAgentSignature        bool
	TermsAccepted            bool
	HasKTPDocument           bool
	HasStoreDocument         bool
	HasSelfieKTPDocument     bool
	HasSelfieMarketing       bool
	MarketingVerified        bool
	MasterVerified           bool
	AnalystVerified          bool
	RevisionResolved         bool
}

func (r *AgentCreditRepository) CreateApplication(ctx context.Context, in AgentCreditApplicationInput) (*AgentCreditApplication, error) {
	applicantJSON, err := json.Marshal(in.ApplicantData)
	if err != nil {
		return nil, err
	}
	documentJSON, err := json.Marshal(in.DocumentData)
	if err != nil {
		return nil, err
	}

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = r.db.QueryRowContext(ctx, `
INSERT INTO public.agent_credit_application
  (member_id, marketing_user_id, requested_amount, status, applicant_data, document_data, agent_signature_data, agent_signature_at)
VALUES
  ($1, NULLIF($2, 0), $3, 'submitted', $4::jsonb, $5::jsonb, NULLIF($6, ''), CASE WHEN NULLIF($6, '') IS NULL THEN NULL ELSE now() END)
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note, created_at, updated_at
`, in.MemberID, in.MarketingID, in.RequestedAmount, string(applicantJSON), string(documentJSON), in.AgentSignature).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

// CreateManualApplication serializes migrations per agent and refuses to
// create another credit record when the agent already has one. The advisory
// lock closes the double-click/concurrent-request window before disbursement.
func (r *AgentCreditRepository) CreateManualApplication(ctx context.Context, in AgentCreditApplicationInput) (*AgentCreditApplication, error) {
	applicantJSON, err := json.Marshal(in.ApplicantData)
	if err != nil {
		return nil, err
	}
	documentJSON, err := json.Marshal(in.DocumentData)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	const migrationLockNamespace int64 = 824010000000000
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLockNamespace+in.MemberID); err != nil {
		return nil, err
	}
	var existingID int64
	err = tx.QueryRowContext(ctx, `
SELECT id
FROM public.agent_credit_application
WHERE member_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 1
`, in.MemberID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if existingID > 0 {
		return nil, fmt.Errorf("agent sudah memiliki data kredit KRD-%08d; gunakan Edit Migrasi", existingID)
	}

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.agent_credit_application
  (member_id, marketing_user_id, requested_amount, status, applicant_data, document_data, agent_signature_data, agent_signature_at)
VALUES
  ($1, NULLIF($2, 0), $3, 'submitted', $4::jsonb, $5::jsonb, NULLIF($6, ''), CASE WHEN NULLIF($6, '') IS NULL THEN NULL ELSE now() END)
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note, created_at, updated_at
`, in.MemberID, in.MarketingID, in.RequestedAmount, string(applicantJSON), string(documentJSON), in.AgentSignature).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) ListActiveRanks(ctx context.Context) ([]AgentCreditRank, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, code, name, COALESCE(description, ''), limit_amount, sort_order
FROM public.agent_credit_rank
WHERE active = TRUE
ORDER BY sort_order ASC, limit_amount ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentCreditRank, 0)
	for rows.Next() {
		var item AgentCreditRank
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.LimitAmount, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AgentCreditRepository) ChangeMemberCreditRank(ctx context.Context, memberID, newRankID, customLimit, actorID int64, reason string) (*AgentCreditRank, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var newRank AgentCreditRank
	if customLimit > 0 {
		newRank = AgentCreditRank{Code: "custom", Name: "Limit Custom", LimitAmount: customLimit}
	} else if err := tx.QueryRowContext(ctx, `
SELECT id, code, name, COALESCE(description, ''), limit_amount, sort_order
FROM public.agent_credit_rank
WHERE id = $1 AND active = TRUE
`, newRankID).Scan(&newRank.ID, &newRank.Code, &newRank.Name, &newRank.Description, &newRank.LimitAmount, &newRank.SortOrder); err != nil {
		return nil, err
	}

	var memberRole string
	var outstanding int64
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(m.role, ''), COALESCE((
  SELECT MAX(l.outstanding_amount)
  FROM public.agent_credit_loan l
  WHERE l.member_id = m.id AND l.status IN ('active', 'due', 'overdue', 'suspended')
), 0)
FROM public.member m
WHERE m.id = $1
`, memberID).Scan(&memberRole, &outstanding); err != nil {
		return nil, fmt.Errorf("agent tidak ditemukan")
	}
	if !strings.EqualFold(memberRole, "agent") {
		return nil, fmt.Errorf("limit hanya dapat diubah untuk akun agent")
	}
	if outstanding > newRank.LimitAmount {
		return nil, fmt.Errorf("limit baru tidak boleh lebih kecil dari kewajiban berjalan Rp%d", outstanding)
	}

	var loanID, applicationID, currentPrincipal int64
	loanErr := tx.QueryRowContext(ctx, `
SELECT id, application_id, principal_amount
FROM public.agent_credit_loan
WHERE member_id = $1 AND status IN ('active', 'due', 'overdue', 'suspended')
ORDER BY approved_at DESC, id DESC
LIMIT 1
FOR UPDATE
`, memberID).Scan(&loanID, &applicationID, &currentPrincipal)
	if loanErr != nil && !errors.Is(loanErr, sql.ErrNoRows) {
		return nil, loanErr
	}

	if loanErr == nil && newRank.LimitAmount > currentPrincipal {
		increaseAmount := newRank.LimitAmount - currentPrincipal
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID); err != nil {
			return nil, err
		}

		var balanceBefore int64
		if err := tx.QueryRowContext(ctx, `
SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE
`, memberID).Scan(&balanceBefore); err != nil {
			return nil, err
		}
		balanceAfter := balanceBefore + increaseAmount
		if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, balanceAfter); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE public.agent_credit_loan
SET
  principal_amount = $2,
  outstanding_amount = outstanding_amount + $3,
  available_amount = 0,
  updated_at = now()
WHERE id = $1
`, loanID, newRank.LimitAmount, increaseAmount); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE public.agent_credit_application
SET
  approved_amount = $2,
  analyst_recommended_amount = $2,
  updated_at = now()
WHERE id = $1
`, applicationID, newRank.LimitAmount); err != nil {
			return nil, err
		}

		refID := fmt.Sprintf("KREDIT-LIMIT-%d-%d", applicationID, time.Now().UnixNano())
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1, $2, 'CREDIT', $3, 'AGENT_CREDIT_LIMIT_INCREASE', $4, $5, $6, now())
`, memberID, refID, increaseAmount, reason, balanceBefore, balanceAfter); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.agent_credit_mutation
  (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1, $2, $3, $4, 'DEBIT', $5, 'CREDIT_LIMIT_INCREASE_TO_MAIN_BALANCE', $6, $7, $8)
`, loanID, applicationID, memberID, refID, increaseAmount, reason, currentPrincipal, newRank.LimitAmount); err != nil {
			return nil, err
		}
	}

	var oldRankID sql.NullInt64
	_ = tx.QueryRowContext(ctx, `
SELECT h.new_rank_id
FROM public.agent_credit_rank_history h
WHERE h.member_id = $1
ORDER BY h.created_at DESC, h.id DESC
LIMIT 1
`, memberID).Scan(&oldRankID)
	if !oldRankID.Valid {
		_ = tx.QueryRowContext(ctx, `SELECT id FROM public.agent_credit_rank WHERE active = TRUE ORDER BY sort_order ASC LIMIT 1`).Scan(&oldRankID)
	}

	var onTimeCount, lateCount int64
	_ = tx.QueryRowContext(ctx, `
SELECT
  COUNT(*) FILTER (WHERE COALESCE(days_late, 0) <= 0),
  COUNT(*) FILTER (WHERE COALESCE(days_late, 0) > 0)
FROM public.agent_credit_payment
WHERE member_id = $1
`, memberID).Scan(&onTimeCount, &lateCount)

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.agent_credit_rank_history
  (member_id, old_rank_id, new_rank_id, custom_limit_amount, custom_limit_name, reason, on_time_payment_count, late_payment_count, created_by)
VALUES ($1, $2, NULLIF($3, 0), NULLIF($4, 0), $5, $6, $7, $8, NULLIF($9, 0))
`, memberID, oldRankID, newRank.ID, customLimit, newRank.Name, reason, onTimeCount, lateCount, actorID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newRank, nil
}

func (r *AgentCreditRepository) SetLoanOperationalStatus(ctx context.Context, applicationID, actorID int64, suspend bool, reason string) (*AgentCreditLoanStatusResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var result AgentCreditLoanStatusResult
	var loanID int64
	var previousStatus string
	var dueDate time.Time
	err = tx.QueryRowContext(ctx, `
SELECT id, member_id, status, available_amount, outstanding_amount, due_date
FROM public.agent_credit_loan
WHERE application_id = $1
FOR UPDATE
`, applicationID).Scan(
		&loanID,
		&result.MemberID,
		&previousStatus,
		&result.CreditAvailableAmount,
		&result.OutstandingAmount,
		&dueDate,
	)
	if err != nil {
		return nil, err
	}
	if result.OutstandingAmount <= 0 || previousStatus == "paid" || previousStatus == "cancelled" || previousStatus == "defaulted" {
		return nil, errors.New("kredit yang sudah selesai tidak dapat diubah")
	}

	nextStatus := "suspended"
	if !suspend {
		nextStatus = "active"
		if time.Now().After(dueDate) {
			nextStatus = "overdue"
		}
	}
	if previousStatus == nextStatus {
		result.ApplicationID = applicationID
		result.Status = nextStatus
		return &result, nil
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE public.agent_credit_loan
SET status = $2, updated_at = now()
WHERE id = $1
`, loanID, nextStatus); err != nil {
		return nil, err
	}
	if err := insertAuditLogTx(ctx, tx, actorID, "super_admin", "agent_credit_operational_status_changed", "agent_credit_loan", loanID, map[string]any{
		"status": previousStatus,
	}, map[string]any{
		"status":         nextStatus,
		"application_id": applicationID,
	}, reason); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	result.ApplicationID = applicationID
	result.Status = nextStatus
	return &result, nil
}

func (r *AgentCreditRepository) FindOpenEarlyApplicationID(ctx context.Context, memberID int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
SELECT id
FROM public.agent_credit_application
WHERE member_id = $1
  AND status IN ('draft', 'submitted', 'marketing_review')
ORDER BY created_at DESC, id DESC
LIMIT 1
`, memberID).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

// EnsureAgentCanSubmitNextApplication enforces the single-wallet credit cycle.
// An agent may submit another request only after the previously approved
// capital has been fully used from the main wallet.
func (r *AgentCreditRepository) EnsureAgentCanSubmitNextApplication(ctx context.Context, memberID int64) error {
	var mainBalance int64
	var activeLoan bool
	if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT saldo FROM public.dompet_member WHERE member_id = $1), 0),
  EXISTS (
    SELECT 1
    FROM public.agent_credit_loan
    WHERE member_id = $1
      AND status IN ('active', 'overdue')
  )
`, memberID).Scan(&mainBalance, &activeLoan); err != nil {
		return err
	}
	if activeLoan && mainBalance > 0 {
		return fmt.Errorf("saldo utama masih tersedia Rp%d; gunakan saldo tersebut terlebih dahulu sebelum mengajukan kredit lagi", mainBalance)
	}
	return nil
}

func (r *AgentCreditRepository) GetApplicationReviewState(ctx context.Context, applicationID int64) (*AgentCreditReviewState, error) {
	var state AgentCreditReviewState
	err := r.db.QueryRowContext(ctx, `
SELECT
  status,
  COALESCE(analyst_recommendation, ''),
  COALESCE(analyst_recommended_amount, 0),
  COALESCE(agent_signature_data, '') <> '',
  COALESCE((applicant_data->>'terms_accepted')::boolean, false),
  COALESCE(document_data->'ktp'->>'data_url', '') <> '',
  COALESCE(document_data->'store'->>'data_url', '') <> '',
  COALESCE(document_data->'selfie_ktp'->>'data_url', '') <> '',
  COALESCE(document_data->'selfie_marketing'->>'data_url', document_data->'selfie'->>'data_url', '') <> '',
  COALESCE(applicant_data->>'marketing_signature_data', '') <> '',
  COALESCE(applicant_data->>'master_signature_data', applicant_data->>'marketing_signature_data', '') <> '',
  COALESCE(analyst_recommendation, '') = 'approved',
  COALESCE(applicant_data->>'operator_revision_resolved_at', '') <> ''
FROM public.agent_credit_application
WHERE id = $1
`, applicationID).Scan(
		&state.Status,
		&state.AnalystRecommendation,
		&state.AnalystRecommendedAmount,
		&state.HasAgentSignature,
		&state.TermsAccepted,
		&state.HasKTPDocument,
		&state.HasStoreDocument,
		&state.HasSelfieKTPDocument,
		&state.HasSelfieMarketing,
		&state.MarketingVerified,
		&state.MasterVerified,
		&state.AnalystVerified,
		&state.RevisionResolved,
	)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *AgentCreditRepository) GetApplicationRevisionStateForMember(ctx context.Context, applicationID, memberID int64) (map[string]any, bool, error) {
	var raw []byte
	var hasAgentSignature bool
	if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE(document_data, '{}'::jsonb),
  COALESCE(agent_signature_data, '') <> ''
FROM public.agent_credit_application
WHERE id = $1 AND member_id = $2
`, applicationID, memberID).Scan(&raw, &hasAgentSignature); err != nil {
		return nil, false, err
	}

	documents := map[string]any{}
	if err := json.Unmarshal(raw, &documents); err != nil {
		return nil, false, err
	}
	return documents, hasAgentSignature, nil
}

func (r *AgentCreditRepository) CompleteApplicationConsent(ctx context.Context, in AgentCreditApplicationInput) (*AgentCreditApplication, error) {
	applicantJSON, err := json.Marshal(in.ApplicantData)
	if err != nil {
		return nil, err
	}
	documentJSON, err := json.Marshal(in.DocumentData)
	if err != nil {
		return nil, err
	}

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = r.db.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  requested_amount = CASE WHEN $3 > 0 THEN $3 ELSE requested_amount END,
  applicant_data = (COALESCE(applicant_data, '{}'::jsonb) || $4::jsonb)
    - 'operator_revision_documents'
    - 'operator_revision_required'
    - 'operator_revision_requested_at',
  document_data = COALESCE(document_data, '{}'::jsonb) || $5::jsonb,
  marketing_user_id = COALESCE(marketing_user_id, NULLIF($7, 0)),
  agent_signature_data = CASE WHEN NULLIF($6, '') IS NULL THEN agent_signature_data ELSE $6 END,
  agent_signature_at = CASE WHEN NULLIF($6, '') IS NULL THEN agent_signature_at ELSE now() END,
  status = CASE
    WHEN status IN ('draft', 'marketing_review') THEN 'submitted'
    ELSE status
  END,
  updated_at = now()
WHERE id = $1
  AND member_id = $2
  AND status IN ('draft', 'submitted', 'marketing_review', 'analysis_review', 'master_review')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note, created_at, updated_at
`, in.ID, in.MemberID, in.RequestedAmount, string(applicantJSON), string(documentJSON), in.AgentSignature, in.MarketingID).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func creditLevelFromProgress(qualifiedPaidCount int64, needsRepair bool) (string, string, int64) {
	if needsRepair {
		return "start", "Kilat Start", 500000
	}
	switch {
	case qualifiedPaidCount >= 5:
		return "elite", "Kilat Elite", 2000000
	case qualifiedPaidCount >= 3:
		return "plus", "Kilat Plus", 1000000
	default:
		return "start", "Kilat Start", 500000
	}
}

func (r *AgentCreditRepository) GetMemberCreditProfile(ctx context.Context, memberID int64) (*AgentCreditProfile, error) {
	var qualifiedPaidTotal int64
	var qualifiedPaidCount int64
	var needsRepair bool
	err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE(SUM(pl.principal_amount) FILTER (
    WHERE pl.status = 'paid'
      AND NOT EXISTS (
        SELECT 1
        FROM public.agent_credit_payment pp
        WHERE pp.loan_id = pl.id
          AND COALESCE(pp.days_late, 0) > 0
      )
  ), 0)::bigint AS qualified_paid_total,
  COUNT(*) FILTER (
    WHERE pl.status = 'paid'
      AND NOT EXISTS (
        SELECT 1
        FROM public.agent_credit_payment pp
        WHERE pp.loan_id = pl.id
          AND COALESCE(pp.days_late, 0) > 0
      )
  )::bigint AS qualified_paid_count,
  EXISTS (
    SELECT 1
    FROM public.agent_credit_payment rp
    WHERE rp.member_id = $1
      AND COALESCE(rp.days_late, 0) > 3
  ) AS needs_repair
FROM public.agent_credit_loan pl
WHERE pl.member_id = $1
`, memberID).Scan(&qualifiedPaidTotal, &qualifiedPaidCount, &needsRepair)
	if err != nil {
		return nil, err
	}
	code, name, limit := creditLevelFromProgress(qualifiedPaidCount, needsRepair)
	var manualCode, manualName string
	var manualLimit int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(r.code, 'custom'), COALESCE(NULLIF(h.custom_limit_name, ''), r.name, 'Limit Custom'), COALESCE(h.custom_limit_amount, r.limit_amount)
FROM public.agent_credit_rank_history h
LEFT JOIN public.agent_credit_rank r ON r.id = h.new_rank_id
WHERE h.member_id = $1 AND (r.active = TRUE OR h.custom_limit_amount IS NOT NULL)
ORDER BY h.created_at DESC, h.id DESC
LIMIT 1
`, memberID).Scan(&manualCode, &manualName, &manualLimit); err == nil {
		code, name, limit = manualCode, manualName, manualLimit
	}
	return &AgentCreditProfile{
		LevelCode:          code,
		LevelName:          name,
		LimitAmount:        limit,
		NeedsRepair:        needsRepair,
		QualifiedPaidTotal: qualifiedPaidTotal,
		QualifiedPaidCount: qualifiedPaidCount,
	}, nil
}

func (r *AgentCreditRepository) GetApplicationCreditLimit(ctx context.Context, applicationID int64) (int64, error) {
	var memberID int64
	if err := r.db.QueryRowContext(ctx, `SELECT member_id FROM public.agent_credit_application WHERE id = $1`, applicationID).Scan(&memberID); err != nil {
		return 0, err
	}
	profile, err := r.GetMemberCreditProfile(ctx, memberID)
	if err != nil {
		return 0, err
	}
	return profile.LimitAmount, nil
}

func (r *AgentCreditRepository) ListApplications(ctx context.Context, limit int) ([]AgentCreditApplication, error) {
	if err := r.EnsureFlexibleLimitSchema(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  a.id,
  a.member_id,
  COALESCE(m.nama, '') AS member_name,
  COALESCE(m.email, '') AS member_email,
  COALESCE(m.phone, '') AS member_phone,
  COALESCE(a.marketing_user_id, m.marketing_id, 0) AS marketing_id,
  COALESCE(a.analyst_user_id, 0) AS analyst_id,
  a.requested_amount,
  a.approved_amount,
  a.status,
  a.applicant_data,
  a.document_data,
  COALESCE(a.agent_signature_data, '') AS agent_signature_data,
  a.agent_signature_at,
  a.marketing_note,
  COALESCE(a.analyst_note, '') AS analyst_note,
  COALESCE(a.analyst_recommendation, '') AS analyst_recommendation,
  COALESCE(a.analyst_recommended_amount, 0) AS analyst_recommended_amount,
  COALESCE(l.status, '') AS loan_status,
  COALESCE(l.outstanding_amount, 0) AS outstanding_amount,
  0::bigint AS credit_available_amount,
  COALESCE(pay.paid_amount, 0) AS paid_amount,
  COALESCE(pay.payment_count, 0) AS payment_count,
  CASE
    WHEN manual_rank.code IS NOT NULL THEN manual_rank.code
    WHEN COALESCE(credit.needs_repair, false) THEN 'start'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 5 THEN 'elite'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 3 THEN 'plus'
    ELSE 'start'
  END AS credit_level_code,
  CASE
    WHEN manual_rank.code IS NOT NULL THEN manual_rank.name
    WHEN COALESCE(credit.needs_repair, false) THEN 'Kilat Start'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 5 THEN 'Kilat Elite'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 3 THEN 'Kilat Plus'
    ELSE 'Kilat Start'
  END AS credit_level_name,
  COALESCE(credit.needs_repair, false) AS credit_needs_repair,
  COALESCE(credit.qualified_paid_total, 0) AS qualified_paid_total,
  CASE
    WHEN manual_rank.code IS NOT NULL THEN manual_rank.limit_amount
    WHEN COALESCE(credit.needs_repair, false) THEN 500000
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 5 THEN 2000000
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 3 THEN 1000000
    ELSE 500000
  END AS credit_limit_amount,
  l.approved_at AS loan_approved_at,
  l.due_date AS loan_due_date,
  GREATEST(
    (SELECT MAX(COALESCE(o.diubah_pada, o.dibuat_pada)) FROM public.app_order o WHERE o.member_id = a.member_id AND lower(COALESCE(o.status, '')) = 'success'),
    (SELECT MAX(t.diperbarui_pada) FROM public.transaksi_member t WHERE t.member_id = a.member_id AND lower(COALESCE(t.status, '')) = 'success')
  ) AS last_transaction_at,
  a.created_at,
  a.updated_at
FROM public.agent_credit_application a
JOIN public.member m ON m.id = a.member_id
LEFT JOIN public.agent_credit_loan l ON l.application_id = a.id
LEFT JOIN LATERAL (
  SELECT COALESCE(SUM(p.amount), 0)::bigint AS paid_amount, COUNT(*)::bigint AS payment_count
  FROM public.agent_credit_payment p
  WHERE p.loan_id = l.id
) pay ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COALESCE(SUM(pl.principal_amount) FILTER (
      WHERE pl.status = 'paid'
        AND NOT EXISTS (
          SELECT 1
          FROM public.agent_credit_payment pp
          WHERE pp.loan_id = pl.id
            AND COALESCE(pp.days_late, 0) > 0
        )
    ), 0)::bigint AS qualified_paid_total,
    COUNT(*) FILTER (
      WHERE pl.status = 'paid'
        AND NOT EXISTS (
          SELECT 1
          FROM public.agent_credit_payment pp
          WHERE pp.loan_id = pl.id
            AND COALESCE(pp.days_late, 0) > 0
        )
    )::bigint AS qualified_paid_count,
    EXISTS (
      SELECT 1
      FROM public.agent_credit_payment rp
      WHERE rp.member_id = a.member_id
        AND COALESCE(rp.days_late, 0) > 3
    ) AS needs_repair
  FROM public.agent_credit_loan pl
  WHERE pl.member_id = a.member_id
) credit ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COALESCE(r.code, 'custom') AS code,
    COALESCE(NULLIF(h.custom_limit_name, ''), r.name, 'Limit Custom') AS name,
    COALESCE(h.custom_limit_amount, r.limit_amount) AS limit_amount
  FROM public.agent_credit_rank_history h
  LEFT JOIN public.agent_credit_rank r ON r.id = h.new_rank_id
  WHERE h.member_id = a.member_id AND (r.active = TRUE OR h.custom_limit_amount IS NOT NULL)
  ORDER BY h.created_at DESC, h.id DESC
  LIMIT 1
) manual_rank ON TRUE
ORDER BY a.created_at DESC, a.id DESC
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentCreditApplication, 0)
	for rows.Next() {
		var item AgentCreditApplication
		var applicantRaw, documentRaw []byte
		var loanApprovedAt, loanDueDate, lastTransactionAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.MemberID,
			&item.MemberName,
			&item.MemberEmail,
			&item.MemberPhone,
			&item.MarketingID,
			&item.AnalystID,
			&item.RequestedAmount,
			&item.ApprovedAmount,
			&item.Status,
			&applicantRaw,
			&documentRaw,
			&item.AgentSignatureData,
			&item.AgentSignatureAt,
			&item.MarketingNote,
			&item.AnalystNote,
			&item.AnalystRecommendation,
			&item.AnalystRecommendedAmount,
			&item.LoanStatus,
			&item.OutstandingAmount,
			&item.CreditAvailableAmount,
			&item.PaidAmount,
			&item.PaymentCount,
			&item.CreditLevelCode,
			&item.CreditLevelName,
			&item.CreditNeedsRepair,
			&item.QualifiedPaidTotal,
			&item.CreditLimitAmount,
			&loanApprovedAt,
			&loanDueDate,
			&lastTransactionAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
		_ = json.Unmarshal(documentRaw, &item.DocumentData)
		item.HasAgentSignature = item.AgentSignatureData != ""
		if loanApprovedAt.Valid {
			item.LoanApprovedAt = &loanApprovedAt.Time
		}
		if loanDueDate.Valid {
			item.LoanDueDate = &loanDueDate.Time
		}
		if lastTransactionAt.Valid {
			item.LastTransactionAt = &lastTransactionAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachPayments(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AgentCreditRepository) ListMemberApplications(ctx context.Context, memberID int64, limit int) ([]AgentCreditApplication, error) {
	if err := r.EnsureFlexibleLimitSchema(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  a.id,
  a.member_id,
  COALESCE(m.nama, '') AS member_name,
  COALESCE(m.email, '') AS member_email,
  COALESCE(m.phone, '') AS member_phone,
  a.requested_amount,
  a.approved_amount,
  a.status,
  a.applicant_data,
  '{}'::jsonb AS document_data,
  COALESCE(a.agent_signature_data, '') AS agent_signature_data,
  a.agent_signature_at,
  a.marketing_note,
  COALESCE(a.analyst_note, '') AS analyst_note,
  COALESCE(a.analyst_recommendation, '') AS analyst_recommendation,
  COALESCE(a.analyst_recommended_amount, 0) AS analyst_recommended_amount,
  COALESCE(l.status, '') AS loan_status,
  COALESCE(l.outstanding_amount, 0) AS outstanding_amount,
  0::bigint AS credit_available_amount,
  COALESCE(pay.paid_amount, 0) AS paid_amount,
  COALESCE(pay.payment_count, 0) AS payment_count,
  CASE
    WHEN manual_rank.code IS NOT NULL THEN manual_rank.code
    WHEN COALESCE(credit.needs_repair, false) THEN 'start'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 5 THEN 'elite'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 3 THEN 'plus'
    ELSE 'start'
  END AS credit_level_code,
  CASE
    WHEN manual_rank.code IS NOT NULL THEN manual_rank.name
    WHEN COALESCE(credit.needs_repair, false) THEN 'Kilat Start'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 5 THEN 'Kilat Elite'
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 3 THEN 'Kilat Plus'
    ELSE 'Kilat Start'
  END AS credit_level_name,
  COALESCE(credit.needs_repair, false) AS credit_needs_repair,
  COALESCE(credit.qualified_paid_total, 0) AS qualified_paid_total,
  CASE
    WHEN manual_rank.code IS NOT NULL THEN manual_rank.limit_amount
    WHEN COALESCE(credit.needs_repair, false) THEN 500000
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 5 THEN 2000000
    WHEN COALESCE(credit.qualified_paid_count, 0) >= 3 THEN 1000000
    ELSE 500000
  END AS credit_limit_amount,
  l.approved_at AS loan_approved_at,
  l.due_date AS loan_due_date,
  GREATEST(
    (SELECT MAX(COALESCE(o.diubah_pada, o.dibuat_pada)) FROM public.app_order o WHERE o.member_id = a.member_id AND lower(COALESCE(o.status, '')) = 'success'),
    (SELECT MAX(t.diperbarui_pada) FROM public.transaksi_member t WHERE t.member_id = a.member_id AND lower(COALESCE(t.status, '')) = 'success')
  ) AS last_transaction_at,
  a.created_at,
  a.updated_at
FROM public.agent_credit_application a
JOIN public.member m ON m.id = a.member_id
LEFT JOIN public.agent_credit_loan l ON l.application_id = a.id
LEFT JOIN LATERAL (
  SELECT COALESCE(SUM(p.amount), 0)::bigint AS paid_amount, COUNT(*)::bigint AS payment_count
  FROM public.agent_credit_payment p
  WHERE p.loan_id = l.id
) pay ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COALESCE(SUM(pl.principal_amount) FILTER (
      WHERE pl.status = 'paid'
        AND NOT EXISTS (
          SELECT 1
          FROM public.agent_credit_payment pp
          WHERE pp.loan_id = pl.id
            AND COALESCE(pp.days_late, 0) > 0
        )
    ), 0)::bigint AS qualified_paid_total,
    COUNT(*) FILTER (
      WHERE pl.status = 'paid'
        AND NOT EXISTS (
          SELECT 1
          FROM public.agent_credit_payment pp
          WHERE pp.loan_id = pl.id
            AND COALESCE(pp.days_late, 0) > 0
        )
    )::bigint AS qualified_paid_count,
    EXISTS (
      SELECT 1
      FROM public.agent_credit_payment rp
      WHERE rp.member_id = a.member_id
        AND COALESCE(rp.days_late, 0) > 3
    ) AS needs_repair
  FROM public.agent_credit_loan pl
  WHERE pl.member_id = a.member_id
) credit ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COALESCE(r.code, 'custom') AS code,
    COALESCE(NULLIF(h.custom_limit_name, ''), r.name, 'Limit Custom') AS name,
    COALESCE(h.custom_limit_amount, r.limit_amount) AS limit_amount
  FROM public.agent_credit_rank_history h
  LEFT JOIN public.agent_credit_rank r ON r.id = h.new_rank_id
  WHERE h.member_id = a.member_id AND (r.active = TRUE OR h.custom_limit_amount IS NOT NULL)
  ORDER BY h.created_at DESC, h.id DESC
  LIMIT 1
) manual_rank ON TRUE
WHERE a.member_id = $1
ORDER BY a.created_at DESC, a.id DESC
LIMIT $2
`, memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentCreditApplication, 0)
	for rows.Next() {
		var item AgentCreditApplication
		var applicantRaw, documentRaw []byte
		var loanApprovedAt, loanDueDate, lastTransactionAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.MemberID,
			&item.MemberName,
			&item.MemberEmail,
			&item.MemberPhone,
			&item.RequestedAmount,
			&item.ApprovedAmount,
			&item.Status,
			&applicantRaw,
			&documentRaw,
			&item.AgentSignatureData,
			&item.AgentSignatureAt,
			&item.MarketingNote,
			&item.AnalystNote,
			&item.AnalystRecommendation,
			&item.AnalystRecommendedAmount,
			&item.LoanStatus,
			&item.OutstandingAmount,
			&item.CreditAvailableAmount,
			&item.PaidAmount,
			&item.PaymentCount,
			&item.CreditLevelCode,
			&item.CreditLevelName,
			&item.CreditNeedsRepair,
			&item.QualifiedPaidTotal,
			&item.CreditLimitAmount,
			&loanApprovedAt,
			&loanDueDate,
			&lastTransactionAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
		_ = json.Unmarshal(documentRaw, &item.DocumentData)
		item.HasAgentSignature = item.AgentSignatureData != ""
		if loanApprovedAt.Valid {
			item.LoanApprovedAt = &loanApprovedAt.Time
		}
		if loanDueDate.Valid {
			item.LoanDueDate = &loanDueDate.Time
		}
		if lastTransactionAt.Valid {
			item.LastTransactionAt = &lastTransactionAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachPayments(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AgentCreditRepository) DeleteRejectedApplication(ctx context.Context, applicationID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	if err := tx.QueryRowContext(ctx, `
SELECT status
FROM public.agent_credit_application
WHERE id = $1
FOR UPDATE
`, applicationID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "rejected" && status != "analysis_rejected" && status != "master_rejected" {
		return errors.New("hanya pengajuan yang ditolak yang dapat dihapus")
	}

	// Hapus data turunan yang mungkin tercatat saat uji coba sebelum menghapus pengajuan.
	if _, err := tx.ExecContext(ctx, `DELETE FROM public.agent_credit_mutation WHERE application_id = $1`, applicationID); err != nil {
		return err
	}
	// Pembayaran tidak menyimpan application_id secara langsung; pembayaran
	// akan ikut terhapus melalui ON DELETE CASCADE saat loan dihapus.
	if _, err := tx.ExecContext(ctx, `DELETE FROM public.agent_credit_loan WHERE application_id = $1`, applicationID); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM public.agent_credit_application WHERE id = $1`, applicationID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func parseAgentCreditPaymentNote(raw string) (string, string, map[string]any) {
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw, "", nil
	}
	note := raw
	if value, ok := payload["note"].(string); ok {
		note = value
	}
	paymentMethod, _ := payload["payment_method"].(string)
	proof, _ := payload["payment_proof"].(map[string]any)
	return note, paymentMethod, proof
}

func insertAuditLogTx(ctx context.Context, tx *sql.Tx, actorID int64, actorRole, action, entityType string, entityID any, beforeData, afterData map[string]any, reason string) error {
	if tx == nil {
		return nil
	}
	beforeJSON, err := json.Marshal(beforeData)
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(afterData)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO public.audit_log
  (actor_id, actor_role, action, entity_type, entity_id, before_data, after_data, reason)
VALUES
  (NULLIF($1, 0), $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8)
`, actorID, actorRole, action, entityType, fmt.Sprint(entityID), string(beforeJSON), string(afterJSON), reason)
	return err
}

func (r *AgentCreditRepository) attachPayments(ctx context.Context, items []AgentCreditApplication) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(items))
	indexByID := make(map[int64]int, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
		indexByID[items[i].ID] = i
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT
  p.id,
  p.loan_id,
  COALESCE(l.application_id, 0) AS application_id,
  p.member_id,
  p.amount,
  p.due_date,
  p.paid_at,
  p.days_late,
  p.status,
  p.note
FROM public.agent_credit_payment p
JOIN public.agent_credit_loan l ON l.id = p.loan_id
WHERE l.application_id IN (
  SELECT value::bigint FROM jsonb_array_elements_text($1::jsonb)
)
ORDER BY p.paid_at DESC, p.id DESC
`, string(idsJSON))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var payment AgentCreditPayment
		var rawNote string
		if err := rows.Scan(
			&payment.ID,
			&payment.LoanID,
			&payment.ApplicationID,
			&payment.MemberID,
			&payment.Amount,
			&payment.DueDate,
			&payment.PaidAt,
			&payment.DaysLate,
			&payment.Status,
			&rawNote,
		); err != nil {
			return err
		}
		payment.Note, payment.PaymentMethod, payment.PaymentProof = parseAgentCreditPaymentNote(rawNote)
		if index, ok := indexByID[payment.ApplicationID]; ok {
			items[index].Payments = append(items[index].Payments, payment)
		}
	}
	return rows.Err()
}

func (r *AgentCreditRepository) MarketingReviewApplication(ctx context.Context, in AgentCreditDecisionInput, signatureData string) (*AgentCreditApplication, error) {
	meta := map[string]any{
		"marketing_signature_data": signatureData,
		"marketing_signature_at":   time.Now().Format(time.RFC3339),
		"marketing_verified":       true,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var oldStatus string
	_ = tx.QueryRowContext(ctx, `SELECT status FROM public.agent_credit_application WHERE id = $1 FOR UPDATE`, in.ID).Scan(&oldStatus)

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = tx.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = $2,
  marketing_user_id = $3,
  marketing_reviewed_at = now(),
  marketing_note = $4,
  applicant_data = COALESCE(applicant_data, '{}'::jsonb) || $5::jsonb,
  updated_at = now()
WHERE id = $1 AND status IN ('submitted', 'marketing_review')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note, created_at, updated_at
`, in.ID, in.Status, in.MarketingID, in.MarketingNote, string(metaJSON)).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	actorRole := in.ActorRole
	if actorRole == "" {
		actorRole = "marketing"
	}
	if err := insertAuditLogTx(ctx, tx, in.MarketingID, actorRole, "agent_credit_marketing_review", "agent_credit_application", in.ID, map[string]any{
		"status": oldStatus,
	}, map[string]any{
		"status": in.Status,
		"note":   in.MarketingNote,
	}, in.MarketingNote); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) MasterReviewApplication(ctx context.Context, in AgentCreditDecisionInput, signatureData string) (*AgentCreditApplication, error) {
	meta := map[string]any{
		"master_signature_data": signatureData,
		"master_signature_at":   time.Now().Format(time.RFC3339),
		"master_verified":       true,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = r.db.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = $2,
  marketing_user_id = $3,
  marketing_reviewed_at = now(),
  marketing_note = $4,
  applicant_data = COALESCE(applicant_data, '{}'::jsonb) || $5::jsonb,
  updated_at = now()
WHERE id = $1 AND status IN ('submitted', 'marketing_review')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note, created_at, updated_at
`, in.ID, in.Status, in.MasterID, in.MarketingNote, string(metaJSON)).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) AnalystFinalDecision(ctx context.Context, in AgentCreditDecisionInput, signatureData, riskLevel string, riskScore int64) (*AgentCreditApplication, error) {
	meta := map[string]any{
		"analyst_decision_at": time.Now().Format(time.RFC3339),
		"risk_level":          riskLevel,
		"risk_score":          riskScore,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var oldStatus string
	_ = tx.QueryRowContext(ctx, `SELECT status FROM public.agent_credit_application WHERE id = $1 FOR UPDATE`, in.ID).Scan(&oldStatus)

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = tx.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = $2,
  approved_amount = CASE WHEN $2 IN ('approved', 'ready_to_disburse') THEN $3 ELSE 0 END,
  analyst_user_id = $4,
  analyst_reviewed_at = now(),
  analyst_note = $5,
  analyst_recommendation = $6,
  analyst_recommended_amount = CASE WHEN $2 IN ('approved', 'ready_to_disburse') THEN $3 ELSE 0 END,
  applicant_data = COALESCE(applicant_data, '{}'::jsonb) || $7::jsonb,
  updated_at = now()
WHERE id = $1
  AND status IN ('submitted', 'marketing_review', 'analysis_review', 'master_review', 'ready_to_disburse')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note,
  COALESCE(analyst_note, ''), COALESCE(analyst_recommendation, ''), COALESCE(analyst_recommended_amount, 0),
  created_at, updated_at
`, in.ID, in.Status, in.ApprovedAmount, in.AnalystID, in.AnalystNote, in.Recommendation, string(metaJSON)).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.AnalystNote,
		&item.AnalystRecommendation,
		&item.AnalystRecommendedAmount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""

	if in.Status == "approved" {
		var loanID int64
		err = tx.QueryRowContext(ctx, `
INSERT INTO public.agent_credit_loan
  (application_id, member_id, principal_amount, outstanding_amount, available_amount, status, approved_at, due_date)
VALUES
	  ($1, $2, $3, $3, 0, 'active', now(), (now() + interval '1 month')::date)
ON CONFLICT (application_id) DO NOTHING
RETURNING id
`, item.ID, item.MemberID, in.ApprovedAmount).Scan(&loanID)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, err
			}
		} else {
			if _, err = tx.ExecContext(ctx, `INSERT INTO public.dompet_member (member_id, saldo) VALUES ($1, 0) ON CONFLICT (member_id) DO NOTHING`, item.MemberID); err != nil {
				return nil, err
			}
			var balanceBefore int64
			if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, item.MemberID).Scan(&balanceBefore); err != nil {
				return nil, err
			}
			balanceAfter := balanceBefore + in.ApprovedAmount
			if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, item.MemberID, balanceAfter); err != nil {
				return nil, err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO public.mutasi_dompet (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada) VALUES ($1,$2,'CREDIT',$3,'AGENT_CREDIT_APPROVED','Kredit disetujui operator dan masuk ke saldo utama',$4,$5,now())`, item.MemberID, fmt.Sprintf("KREDIT-%d", item.ID), in.ApprovedAmount, balanceBefore, balanceAfter); err != nil {
				return nil, err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO public.agent_credit_mutation (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah) VALUES ($1,$2,$3,$4,'DEBIT',$5,'CREDIT_DISBURSED_TO_MAIN_BALANCE','Kredit dicairkan langsung ke saldo utama',0,0)`, loanID, item.ID, item.MemberID, fmt.Sprintf("KREDIT-%d", item.ID), in.ApprovedAmount); err != nil {
				return nil, err
			}
		}
	} else if in.Status != "ready_to_disburse" {
		_, err = tx.ExecContext(ctx, `
UPDATE public.agent_credit_loan
SET status = 'cancelled', updated_at = now()
WHERE application_id = $1 AND status <> 'paid'
`, item.ID)
		if err != nil {
			return nil, err
		}
	}
	actorRole := in.ActorRole
	if actorRole == "" {
		actorRole = "operator_credit"
	}
	if err := insertAuditLogTx(ctx, tx, in.AnalystID, actorRole, "agent_credit_final_decision", "agent_credit_application", in.ID, map[string]any{
		"status": oldStatus,
	}, map[string]any{
		"status":          in.Status,
		"approved_amount": in.ApprovedAmount,
		"recommendation":  in.Recommendation,
	}, in.AnalystNote); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AgentCreditRepository) AnalystReviewApplication(ctx context.Context, in AgentCreditDecisionInput) (*AgentCreditApplication, error) {
	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err := r.db.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = 'analysis_review',
  analyst_user_id = $2,
  analyst_reviewed_at = now(),
  analyst_note = $3,
  analyst_recommendation = $4,
  analyst_recommended_amount = $5,
  updated_at = now()
WHERE id = $1 AND status IN ('submitted', 'marketing_review', 'analysis_review')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note,
  COALESCE(analyst_note, ''), COALESCE(analyst_recommendation, ''), COALESCE(analyst_recommended_amount, 0),
  created_at, updated_at
`, in.ID, in.AnalystID, in.AnalystNote, in.Recommendation, in.ApprovedAmount).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.AnalystNote,
		&item.AnalystRecommendation,
		&item.AnalystRecommendedAmount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) ReturnApplicationToMarketing(ctx context.Context, in AgentCreditDecisionInput, revisionDocuments []string) (*AgentCreditApplication, error) {
	metaJSON, err := json.Marshal(map[string]any{
		"operator_revision_requested_at": time.Now().UTC().Format(time.RFC3339),
		"operator_revision_required":     true,
		"operator_revision_documents":    revisionDocuments,
		"operator_revision_resolved_at":  nil,
	})
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = tx.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = 'marketing_review',
  analyst_user_id = $2,
  analyst_reviewed_at = now(),
  analyst_note = $3,
	analyst_recommendation = 'revision_required',
  analyst_recommended_amount = 0,
  applicant_data = COALESCE(applicant_data, '{}'::jsonb) || $4::jsonb,
  updated_at = now()
WHERE id = $1 AND status NOT IN ('approved', 'rejected', 'analysis_rejected', 'master_rejected')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note,
  COALESCE(analyst_note, ''), COALESCE(analyst_recommendation, ''), COALESCE(analyst_recommended_amount, 0),
  created_at, updated_at
`, in.ID, in.AnalystID, in.AnalystNote, string(metaJSON)).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.AnalystNote,
		&item.AnalystRecommendation,
		&item.AnalystRecommendedAmount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := insertAuditLogTx(ctx, tx, in.AnalystID, "operator_credit", "agent_credit_returned_to_marketing", "agent_credit_application", in.ID, map[string]any{
		"status": "analysis_review",
	}, map[string]any{
		"status": "marketing_review",
		"note":   in.AnalystNote,
	}, in.AnalystNote); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) AnalystRejectApplication(ctx context.Context, in AgentCreditDecisionInput) (*AgentCreditApplication, error) {
	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err := r.db.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = 'analysis_rejected',
  approved_amount = 0,
  analyst_user_id = $2,
  analyst_reviewed_at = now(),
  analyst_note = $3,
  analyst_recommendation = 'rejected',
  analyst_recommended_amount = 0,
  updated_at = now()
WHERE id = $1 AND status IN ('submitted', 'marketing_review', 'analysis_review')
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note,
  COALESCE(analyst_note, ''), COALESCE(analyst_recommendation, ''), COALESCE(analyst_recommended_amount, 0),
  created_at, updated_at
`, in.ID, in.AnalystID, in.AnalystNote).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.AnalystNote,
		&item.AnalystRecommendation,
		&item.AnalystRecommendedAmount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) DecideApplication(ctx context.Context, in AgentCreditDecisionInput) (*AgentCreditApplication, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var item AgentCreditApplication
	var applicantRaw, documentRaw []byte
	err = tx.QueryRowContext(ctx, `
UPDATE public.agent_credit_application
SET
  status = $2,
  approved_amount = CASE WHEN $2 = 'approved' THEN $3 ELSE 0 END,
  marketing_user_id = $4,
  marketing_reviewed_at = now(),
  marketing_note = $5,
  updated_at = now()
WHERE id = $1 AND (
  ($2 = 'approved' AND status = 'ready_to_disburse')
  OR ($2 <> 'approved' AND status IN ('submitted', 'marketing_review', 'analysis_review', 'master_review', 'ready_to_disburse', 'approved', 'rejected', 'master_rejected'))
)
RETURNING id, member_id, requested_amount, approved_amount, status, applicant_data, document_data,
  COALESCE(agent_signature_data, ''), agent_signature_at, marketing_note, created_at, updated_at
`, in.ID, in.Status, in.ApprovedAmount, in.MasterID, in.MarketingNote).Scan(
		&item.ID,
		&item.MemberID,
		&item.RequestedAmount,
		&item.ApprovedAmount,
		&item.Status,
		&applicantRaw,
		&documentRaw,
		&item.AgentSignatureData,
		&item.AgentSignatureAt,
		&item.MarketingNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(applicantRaw, &item.ApplicantData)
	if in.Status == "approved" {
		var loanID int64
		err = tx.QueryRowContext(ctx, `
INSERT INTO public.agent_credit_loan
  (application_id, member_id, principal_amount, outstanding_amount, available_amount, status, approved_at, due_date)
VALUES
	  ($1, $2, $3, $3, 0, 'active', now(), (now() + interval '1 month')::date)
ON CONFLICT (application_id) DO NOTHING
RETURNING id
`, item.ID, item.MemberID, in.ApprovedAmount).Scan(&loanID)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, err
			}
		} else {
			if _, err = tx.ExecContext(ctx, `INSERT INTO public.dompet_member (member_id, saldo) VALUES ($1, 0) ON CONFLICT (member_id) DO NOTHING`, item.MemberID); err != nil {
				return nil, err
			}
			var balanceBefore int64
			if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, item.MemberID).Scan(&balanceBefore); err != nil {
				return nil, err
			}
			balanceAfter := balanceBefore + in.ApprovedAmount
			if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, item.MemberID, balanceAfter); err != nil {
				return nil, err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO public.mutasi_dompet (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada) VALUES ($1,$2,'CREDIT',$3,'AGENT_CREDIT_APPROVED','Kredit disetujui operator dan masuk ke saldo utama',$4,$5,now())`, item.MemberID, fmt.Sprintf("KREDIT-%d", item.ID), in.ApprovedAmount, balanceBefore, balanceAfter); err != nil {
				return nil, err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO public.agent_credit_mutation (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah) VALUES ($1,$2,$3,$4,'DEBIT',$5,'CREDIT_DISBURSED_TO_MAIN_BALANCE','Kredit dicairkan langsung ke saldo utama',0,0)`, loanID, item.ID, item.MemberID, fmt.Sprintf("KREDIT-%d", item.ID), in.ApprovedAmount); err != nil {
				return nil, err
			}
		}
	} else {
		_, err = tx.ExecContext(ctx, `
UPDATE public.agent_credit_loan
SET status = 'cancelled', updated_at = now()
WHERE application_id = $1 AND status <> 'paid'
`, item.ID)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(documentRaw, &item.DocumentData)
	item.HasAgentSignature = item.AgentSignatureData != ""
	return &item, nil
}

func (r *AgentCreditRepository) PayInstallment(ctx context.Context, in AgentCreditPaymentInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var loanID, memberID, principal, outstanding int64
	var dueDate time.Time
	err = tx.QueryRowContext(ctx, `
SELECT
  l.id,
  l.member_id,
  l.principal_amount,
  l.outstanding_amount,
  l.due_date
FROM public.agent_credit_loan l
WHERE l.application_id = $1 AND l.status IN ('active', 'overdue')
FOR UPDATE
`, in.ApplicationID).Scan(&loanID, &memberID, &principal, &outstanding, &dueDate)
	if err != nil {
		return err
	}
	if in.MemberID > 0 && in.MemberID != memberID {
		return sql.ErrNoRows
	}
	var existingMemberID int64
	err = tx.QueryRowContext(ctx, `
SELECT member_id
FROM public.agent_credit_payment
WHERE ref_id = $1
`, in.RefID).Scan(&existingMemberID)
	if err == nil {
		if existingMemberID != memberID {
			return fmt.Errorf("referensi pembayaran sudah digunakan")
		}
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID); err != nil {
		return err
	}
	var mainBefore int64
	if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&mainBefore); err != nil {
		return err
	}

	var balanceBeforeCredit int64
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE((
  SELECT saldo_sebelum
  FROM public.mutasi_dompet
  WHERE member_id = $1
    AND ref_id = $2
    AND alasan = 'AGENT_CREDIT_APPROVED'
  ORDER BY id ASC
  LIMIT 1
), 0)
`, memberID, fmt.Sprintf("KREDIT-%d", in.ApplicationID)).Scan(&balanceBeforeCredit); err != nil {
		return err
	}
	maximumMainBalance := balanceBeforeCredit + principal
	refillCapacity := maximumMainBalance - mainBefore
	if refillCapacity <= 0 {
		return fmt.Errorf("saldo utama masih penuh, belum ada modal yang perlu dilunasi")
	}
	amount := in.Amount
	if amount > refillCapacity {
		return fmt.Errorf("pelunasan melebihi modal terpakai: maksimal Rp%d", refillCapacity)
	}
	if amount > outstanding {
		return fmt.Errorf("pelunasan melebihi kewajiban berjalan: maksimal Rp%d", outstanding)
	}
	daysLate := 0
	now := time.Now()
	if now.After(dueDate) {
		daysLate = int(now.Sub(dueDate).Hours() / 24)
	}
	paymentStatus := "partial"
	if daysLate > 0 {
		paymentStatus = "late"
	}
	paymentNote := in.Note
	if len(in.PaymentProof) > 0 {
		notePayload, marshalErr := json.Marshal(map[string]any{
			"payment_method": in.PaymentMethod,
			"note":           in.Note,
			"payment_proof":  in.PaymentProof,
		})
		if marshalErr != nil {
			return marshalErr
		}
		paymentNote = string(notePayload)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.agent_credit_payment
  (loan_id, member_id, ref_id, amount, due_date, days_late, status, note)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8)
`, loanID, memberID, in.RefID, amount, dueDate, daysLate, paymentStatus, paymentNote)
	if err != nil {
		return err
	}

	mainAfter := mainBefore + amount
	_, err = tx.ExecContext(ctx, `
UPDATE public.agent_credit_loan
SET
  outstanding_amount = GREATEST(outstanding_amount - $2, 0) + $2,
  available_amount = 0,
  status = 'active',
  paid_at = NULL,
  updated_at = now()
WHERE id = $1
`, loanID, amount)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, mainAfter); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1, $2, 'CREDIT', $3, 'AGENT_CREDIT_REVOLVING_REDRAW',
   'Pelunasan sebagian dicairkan kembali ke saldo utama', $4, $5, now())
`, memberID, in.RefID, amount, mainBefore, mainAfter); err != nil {
		return err
	}
	if err := insertAuditLogTx(ctx, tx, memberID, "agent", "agent_credit_partial_payment_redraw", "agent_credit_loan", loanID, map[string]any{
		"outstanding_amount": outstanding,
		"main_balance":       mainBefore,
	}, map[string]any{
		"paid_amount":        amount,
		"outstanding_amount": outstanding,
		"main_balance":       mainAfter,
		"status":             "active",
		"ref_id":             in.RefID,
	}, in.Note); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AgentCreditRepository) TransferCreditToMainBalance(ctx context.Context, applicationID, memberID, amount int64, refID string) (*AgentCreditTransferResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var loanID, availableBefore, outstanding int64
	err = tx.QueryRowContext(ctx, `
SELECT id, available_amount, outstanding_amount
FROM public.agent_credit_loan
WHERE application_id = $1
  AND member_id = $2
  AND status = 'active'
FOR UPDATE
`, applicationID, memberID).Scan(&loanID, &availableBefore, &outstanding)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("saldo kredit aktif tidak ditemukan")
		}
		return nil, err
	}
	if amount > availableBefore {
		return nil, fmt.Errorf("saldo kredit tidak cukup: tersedia=%d", availableBefore)
	}

	availableAfter := availableBefore - amount
	if _, err := tx.ExecContext(ctx, `
UPDATE public.agent_credit_loan
SET available_amount = $2, updated_at = now()
WHERE id = $1
`, loanID, availableAfter); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.agent_credit_mutation
  (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,$3,$4,'DEBIT',$5,'TRANSFER_TO_MAIN_BALANCE','Mutasi saldo kredit ke saldo utama',$6,$7)
`, loanID, applicationID, memberID, refID, amount, availableBefore, availableAfter); err != nil {
		return nil, err
	}

	var mainBefore int64
	if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&mainBefore); err != nil {
		return nil, err
	}
	mainAfter := mainBefore + amount
	if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, mainAfter); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'AGENT_CREDIT_TRANSFER','Pencairan saldo kredit ke saldo utama',$4,$5,now())
`, memberID, refID, amount, mainBefore, mainAfter); err != nil {
		return nil, err
	}

	if err := insertAuditLogTx(ctx, tx, memberID, "agent", "agent_credit_transfer_to_main_balance", "agent_credit_loan", loanID, map[string]any{
		"credit_available_amount": availableBefore,
		"main_balance":            mainBefore,
	}, map[string]any{
		"credit_available_amount": availableAfter,
		"main_balance":            mainAfter,
		"amount":                  amount,
		"ref_id":                  refID,
	}, "Agent memindahkan saldo kredit ke saldo utama"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &AgentCreditTransferResult{
		ApplicationID:         applicationID,
		Amount:                amount,
		CreditAvailableAmount: availableAfter,
		OutstandingAmount:     outstanding,
		MainBalance:           mainAfter,
		RefID:                 refID,
	}, nil
}
