package repository

import "time"

type AdminCommissionBySourceArgs struct {
	Scope   string
	Limit   int
	Offset  int
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
}

type AdminCommissionBySourceRow struct {
	Scope            string     `json:"scope"`
	UplineMemberID   int64      `json:"upline_member_id"`
	UplineEmail      string     `json:"upline_email"`
	UplineNama       string     `json:"upline_nama"`
	UplineRole       string     `json:"upline_role"`
	Level            string     `json:"level"`
	SourceMemberID   int64      `json:"source_member_id"`
	SourceEmail      string     `json:"source_email"`
	SourceNama       string     `json:"source_nama"`
	SourceRole       string     `json:"source_role"`
	TransactionCount int64      `json:"transaction_count"`
	TotalCommission  int64      `json:"total_commission"`
	LastCreatedAt    *time.Time `json:"last_created_at,omitempty"`
}

type AdminDailyBusinessArgs struct {
	Scope   string
	Months  int
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
}

type AdminDailyBusinessRow struct {
	Scope                 string    `json:"scope"`
	Day                   time.Time `json:"day"`
	MonthKey              string    `json:"month_key"`
	TransactionCount      int64     `json:"transaction_count"`
	TransactionAmount     int64     `json:"transaction_amount"`
	ProviderPaymentAmount int64     `json:"provider_payment_amount"`
	MarginAmount          int64     `json:"margin_amount"`
	CommissionAmount      int64     `json:"commission_amount"`
	TransactionExpense    int64     `json:"transaction_expense_amount"`
	MemberDepositAmount   int64     `json:"member_deposit_amount"`
	ProviderDepositAmount int64     `json:"provider_deposit_amount"`
	DepositGapAmount      int64     `json:"deposit_gap_amount"`
	ProfitAmount          int64     `json:"profit_amount"`
}

type AdminDailyBusinessCacheRefreshResult struct {
	Days          int       `json:"days"`
	RefreshedRows int64     `json:"refreshed_rows"`
	RefreshedAt   time.Time `json:"refreshed_at"`
}
