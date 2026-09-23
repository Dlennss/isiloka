package repository

import (
	"strconv"
	"strings"
	"time"
)

type AuthLoginRow struct {
	ID           int64
	Email        string
	Nama         string
	Role         string
	Aktif        bool
	PasswordHash string
}

type AuthGoogleRow struct {
	ID        int64
	Role      string
	Aktif     bool
	GoogleSub string
	AppleSub  string
}

type AuthMeRow struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Nama  string `json:"nama"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
	Aktif bool   `json:"aktif"`
}

type UserRow struct {
	ID                       int64      `json:"id"`
	Email                    string     `json:"email"`
	Nama                     string     `json:"nama"`
	Phone                    string     `json:"phone"`
	Role                     string     `json:"role"`
	Aktif                    bool       `json:"aktif"`
	FeeMemberRp              int64      `json:"fee_member_rp"`
	RetailAgentCommissionRp  int64      `json:"retail_agent_commission_rp"`
	RetailMasterCommissionRp int64      `json:"retail_master_commission_rp"`
	H2HAgentCommissionRp     int64      `json:"h2h_agent_commission_rp"`
	H2HMasterCommissionRp    int64      `json:"h2h_master_commission_rp"`
	Saldo                    int64      `json:"saldo"`
	RetailAgentID            *int64     `json:"retail_agent_id,omitempty"`
	RetailMasterID           *int64     `json:"retail_master_id,omitempty"`
	H2HAgentID               *int64     `json:"h2h_agent_member_id,omitempty"`
	H2HMasterID              *int64     `json:"h2h_master_member_id,omitempty"`
	DibuatPada               *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada               *time.Time `json:"diubah_pada,omitempty"`
}

type UserCreateInput struct {
	Email                    string
	Nama                     string
	Phone                    string
	StoreName                string
	PasswordHash             string
	PinHash                  string
	Role                     string
	APIKey                   string
	Label                    string
	Aktif                    bool
	RetailAgentCommissionRp  int64
	RetailMasterCommissionRp int64
	H2HAgentCommissionRp     int64
	H2HMasterCommissionRp    int64
	RetailAgentID            *int64
	RetailMasterID           *int64
	H2HAgentID               *int64
	H2HMasterID              *int64
	MarketingID              *int64
}

type UserUpdateInput struct {
	ID                       int64
	Email                    string
	Nama                     string
	Phone                    string
	Role                     string
	Aktif                    bool
	FeeMemberRp              int64
	RetailAgentCommissionRp  int64
	RetailMasterCommissionRp int64
	H2HAgentCommissionRp     int64
	H2HMasterCommissionRp    int64
	PasswordHash             *string
	PinHash                  *string
}

type MarketingAccountRow struct {
	ID                     int64      `json:"id"`
	Email                  string     `json:"email"`
	Nama                   string     `json:"nama"`
	Phone                  string     `json:"phone"`
	Aktif                  bool       `json:"aktif"`
	AgentCount             int64      `json:"agent_count"`
	ActiveAgentCount       int64      `json:"active_agent_count"`
	LastAgentTransactionAt *time.Time `json:"last_agent_transaction_at,omitempty"`
	DibuatPada             *time.Time `json:"dibuat_pada,omitempty"`
}

type HierarchyAssignPreviewTarget struct {
	MemberID        int64  `json:"member_id"`
	Email           string `json:"email"`
	Nama            string `json:"nama"`
	Role            string `json:"role"`
	CommissionRp    int64  `json:"commission_rp"`
	CalculatedTotal int64  `json:"calculated_total"`
	CalculatedCount int64  `json:"calculated_count"`
}

type HierarchyAssignPreviewRow struct {
	Scope                  string                        `json:"scope"`
	MemberID               int64                         `json:"member_id"`
	MemberEmail            string                        `json:"member_email"`
	MemberNama             string                        `json:"member_nama"`
	MemberRole             string                        `json:"member_role"`
	CurrentAgentID         *int64                        `json:"current_agent_id,omitempty"`
	CurrentMasterID        *int64                        `json:"current_master_id,omitempty"`
	NextAgentID            *int64                        `json:"next_agent_id,omitempty"`
	NextMasterID           *int64                        `json:"next_master_id,omitempty"`
	DerivedMasterFromAgent bool                          `json:"derived_master_from_agent"`
	TransactionCount       int64                         `json:"transaction_count"`
	Agent                  *HierarchyAssignPreviewTarget `json:"agent,omitempty"`
	Master                 *HierarchyAssignPreviewTarget `json:"master,omitempty"`
}

type HierarchyAssignApplyRow struct {
	Preview            HierarchyAssignPreviewRow `json:"preview"`
	AgentAppliedCount  int64                     `json:"agent_applied_count"`
	AgentAppliedTotal  int64                     `json:"agent_applied_total"`
	MasterAppliedCount int64                     `json:"master_applied_count"`
	MasterAppliedTotal int64                     `json:"master_applied_total"`
}

type UserMonthStatsRow struct {
	Month             time.Time `json:"month"`
	TrxSuccessCount   int64     `json:"trx_success_count"`
	TrxFailedCount    int64     `json:"trx_failed_count"`
	TrxSuccessAmount  int64     `json:"trx_success_amount"`
	TrxFailedAmount   int64     `json:"trx_failed_amount"`
	DepApprovedCount  int64     `json:"dep_approved_count"`
	DepRejectedCount  int64     `json:"dep_rejected_count"`
	DepApprovedAmount int64     `json:"dep_approved_amount"`
	DepRejectedAmount int64     `json:"dep_rejected_amount"`
	WalletAdjustNet   int64     `json:"wallet_adjust_net"`
}

type DepositRequestRow struct {
	ID           int64      `json:"id"`
	MemberID     int64      `json:"member_id"`
	RefID        string     `json:"ref_id"`
	MemberNama   *string    `json:"member_nama,omitempty"`
	BankID       *int64     `json:"bank_id,omitempty"`
	BankNama     string     `json:"bank_nama"`
	BankNomor    string     `json:"bank_nomor_rekening"`
	BankAtasNama string     `json:"bank_atas_nama"`
	Amount       int64      `json:"amount"`
	Requested    int64      `json:"requested_amount,omitempty"`
	UniqueCode   int64      `json:"unique_code,omitempty"`
	Approved     int64      `json:"approved_amount,omitempty"`
	Metode       string     `json:"metode"`
	BuktiURL     string     `json:"bukti_url"`
	Status       string     `json:"status"`
	Note         string     `json:"note"`
	DibuatPada   time.Time  `json:"dibuat_pada"`
	DiprosesPada *time.Time `json:"diproses_pada"`
	DiprosesOleh *int64     `json:"diproses_oleh"`
	DiprosesNama *string    `json:"diproses_nama,omitempty"`
}

func (d *DepositRequestRow) HydrateTicketFields() {
	if d == nil {
		return
	}
	if d.Requested <= 0 {
		d.Requested = d.Amount
	}
	if d.Note == "" {
		return
	}
	for _, part := range strings.Fields(d.Note) {
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		if err != nil {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "requested_amount":
			if n > 0 {
				d.Requested = n
			}
		case "unique_code":
			if n > 0 {
				d.UniqueCode = n
			}
		}
	}
}

type RetailMemberContextRow struct {
	MemberID       int64  `json:"member_id"`
	Email          string `json:"email"`
	Nama           string `json:"nama"`
	Role           string `json:"role"`
	Aktif          bool   `json:"aktif"`
	Saldo          int64  `json:"saldo"`
	RetailAgentID  *int64 `json:"retail_agent_id,omitempty"`
	RetailMasterID *int64 `json:"retail_master_id,omitempty"`
	H2HAgentID     *int64 `json:"h2h_agent_member_id,omitempty"`
	H2HMasterID    *int64 `json:"h2h_master_member_id,omitempty"`
}

type RetailDownlineRow struct {
	ID               int64      `json:"id"`
	Email            string     `json:"email"`
	Nama             string     `json:"nama"`
	Phone            string     `json:"phone"`
	StoreName        string     `json:"store_name"`
	Role             string     `json:"role"`
	Aktif            bool       `json:"aktif"`
	Saldo            int64      `json:"saldo"`
	RetailAgentID    *int64     `json:"retail_agent_id,omitempty"`
	RetailAgentNama  *string    `json:"retail_agent_nama,omitempty"`
	RetailMasterID   *int64     `json:"retail_master_id,omitempty"`
	RetailMasterNama *string    `json:"retail_master_nama,omitempty"`
	MarketingID      *int64     `json:"marketing_id,omitempty"`
	MarketingNama    *string    `json:"marketing_nama,omitempty"`
	MarketingEmail   *string    `json:"marketing_email,omitempty"`
	DibuatPada       *time.Time `json:"dibuat_pada,omitempty"`
}

type H2HDownlineRow struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email"`
	Nama          string     `json:"nama"`
	Role          string     `json:"role"`
	Aktif         bool       `json:"aktif"`
	Saldo         int64      `json:"saldo"`
	H2HAgentID    *int64     `json:"h2h_agent_member_id,omitempty"`
	H2HAgentNama  *string    `json:"h2h_agent_nama,omitempty"`
	H2HMasterID   *int64     `json:"h2h_master_member_id,omitempty"`
	H2HMasterNama *string    `json:"h2h_master_nama,omitempty"`
	DibuatPada    *time.Time `json:"dibuat_pada,omitempty"`
}

type H2HCommissionLedgerRow struct {
	ID                int64      `json:"id"`
	MemberID          int64      `json:"member_id"`
	SourceMemberID    *int64     `json:"source_member_id,omitempty"`
	SourceMemberNama  *string    `json:"source_member_nama,omitempty"`
	SourceMemberRole  *string    `json:"source_member_role,omitempty"`
	SourceTrxMemberID int64      `json:"source_trx_member_id"`
	RefID             string     `json:"ref_id"`
	Level             string     `json:"level"`
	Amount            int64      `json:"amount"`
	Kategori          string     `json:"kategori"`
	Note              string     `json:"note"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
}

type H2HCommissionSummaryRow struct {
	TotalEarned           int64 `json:"total_earned"`
	TotalPendingWithdraw  int64 `json:"total_pending_withdraw"`
	TotalApprovedWithdraw int64 `json:"total_approved_withdraw"`
	TotalRejectedWithdraw int64 `json:"total_rejected_withdraw"`
	AvailableSaldo        int64 `json:"available_saldo"`
}

type H2HWithdrawRequestRow struct {
	ID            int64      `json:"id"`
	MemberID      int64      `json:"member_id"`
	MemberNama    *string    `json:"member_nama,omitempty"`
	MemberEmail   *string    `json:"member_email,omitempty"`
	Amount        int64      `json:"amount"`
	BankName      string     `json:"bank_name"`
	AccountName   string     `json:"account_name"`
	AccountNumber string     `json:"account_number"`
	Status        string     `json:"status"`
	Note          string     `json:"note"`
	RejectReason  string     `json:"reject_reason"`
	RefID         string     `json:"ref_id"`
	ProcessedBy   *int64     `json:"processed_by,omitempty"`
	ProcessedName *string    `json:"processed_name,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type RetailCommissionLedgerRow struct {
	ID               int64      `json:"id"`
	MemberID         int64      `json:"member_id"`
	SourceMemberID   *int64     `json:"source_member_id,omitempty"`
	SourceMemberNama *string    `json:"source_member_nama,omitempty"`
	SourceMemberRole *string    `json:"source_member_role,omitempty"`
	SourceAppOrderID int64      `json:"source_app_order_id"`
	InvoiceID        string     `json:"invoice_id"`
	Level            string     `json:"level"`
	Amount           int64      `json:"amount"`
	Note             string     `json:"note"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
}

type RetailCommissionSummaryRow struct {
	TotalEarned           int64 `json:"total_earned"`
	TotalPendingWithdraw  int64 `json:"total_pending_withdraw"`
	TotalApprovedWithdraw int64 `json:"total_approved_withdraw"`
	TotalRejectedWithdraw int64 `json:"total_rejected_withdraw"`
	AvailableSaldo        int64 `json:"available_saldo"`
}

type RetailWithdrawRequestRow struct {
	ID            int64      `json:"id"`
	MemberID      int64      `json:"member_id"`
	MemberNama    *string    `json:"member_nama,omitempty"`
	MemberEmail   *string    `json:"member_email,omitempty"`
	Amount        int64      `json:"amount"`
	SourceType    string     `json:"source_type"`
	CreditLoanID  *int64     `json:"credit_loan_id,omitempty"`
	ApplicationID *int64     `json:"credit_application_id,omitempty"`
	BankName      string     `json:"bank_name"`
	AccountName   string     `json:"account_name"`
	AccountNumber string     `json:"account_number"`
	Status        string     `json:"status"`
	Note          string     `json:"note"`
	RejectReason  string     `json:"reject_reason"`
	RefID         string     `json:"ref_id"`
	ProcessedBy   *int64     `json:"processed_by,omitempty"`
	ProcessedName *string    `json:"processed_name,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}
