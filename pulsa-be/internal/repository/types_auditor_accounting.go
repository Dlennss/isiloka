package repository

import "time"

type InternalFinanceCreateInput struct {
	EntryType    string
	Category     string
	BankID       int64
	Amount       int64
	Fee          int64
	Counterparty string
	Note         string
	OccurredAt   time.Time
	MetaJSON     []byte
}

type InternalFinanceEntryRow struct {
	ID           int64     `json:"id"`
	RefID        string    `json:"ref_id"`
	EntryType    string    `json:"entry_type"`
	Category     string    `json:"category"`
	Direction    string    `json:"direction"`
	BankID       int64     `json:"bank_id"`
	BankNama     string    `json:"bank_nama"`
	Amount       int64     `json:"amount"`
	Fee          int64     `json:"fee"`
	TotalAmount  int64     `json:"total_amount"`
	Counterparty string    `json:"counterparty"`
	Note         string    `json:"note"`
	OccurredAt   time.Time `json:"occurred_at"`
	CreatedBy    *int64    `json:"created_by,omitempty"`
	CreatedNama  *string   `json:"created_nama,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type OpeningBalanceUpsertInput struct {
	PeriodMonth  time.Time `json:"period_month"`
	AccountCode  string    `json:"account_code"`
	AccountName  string    `json:"account_name"`
	AccountGroup string    `json:"account_group"`
	Amount       int64     `json:"amount"`
	Note         string    `json:"note"`
}

type OpeningBalanceRow struct {
	ID           int64     `json:"id"`
	PeriodMonth  time.Time `json:"period_month"`
	AccountCode  string    `json:"account_code"`
	AccountName  string    `json:"account_name"`
	AccountGroup string    `json:"account_group"`
	Amount       int64     `json:"amount"`
	Note         string    `json:"note"`
	CreatedBy    *int64    `json:"created_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuditorTradingSummaryArgs struct {
	Scope   string
	Period  string
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
}

type AuditorTradingSummaryRow struct {
	Scope            string `json:"scope"`
	PeriodKey        string `json:"period_key"`
	TransactionCount int64  `json:"transaction_count"`
	SalesAmount      int64  `json:"sales_amount"`
	ProviderAmount   int64  `json:"provider_amount"`
	CommissionAmount int64  `json:"commission_amount"`
	MarginAmount     int64  `json:"margin_amount"`
}

type AuditorTradingDetailArgs struct {
	Scope   string
	RefID   string
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
	Limit   int
	Offset  int
}

type AuditorTradingDetailRow struct {
	Scope            string    `json:"scope"`
	RefID            string    `json:"ref_id"`
	OccurredAt       time.Time `json:"occurred_at"`
	Status           string    `json:"status"`
	MemberID         *int64    `json:"member_id,omitempty"`
	MemberNama       *string   `json:"member_nama,omitempty"`
	MemberEmail      *string   `json:"member_email,omitempty"`
	ProductCode      string    `json:"product_code"`
	ProductName      string    `json:"product_name"`
	Destination      string    `json:"destination"`
	Provider         *string   `json:"provider,omitempty"`
	HargaBeli        int64     `json:"harga_beli"`
	HargaJual        int64     `json:"harga_jual"`
	Komisi           int64     `json:"komisi"`
	Margin           int64     `json:"margin"`
	ProviderRef      *string   `json:"provider_ref,omitempty"`
	ManualStatusNote *string   `json:"status_note,omitempty"`
}

type AuditorFinanceArgs struct {
	BankID  int64
	Type    string
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
	Limit   int
	Offset  int
}

type AuditorFinanceRow struct {
	ID           int64     `json:"id"`
	BankID       int64     `json:"bank_id"`
	BankNama     string    `json:"bank_nama"`
	RefID        string    `json:"ref_id"`
	Type         string    `json:"type"`
	Arah         string    `json:"arah"`
	Amount       int64     `json:"amount"`
	Fee          int64     `json:"fee"`
	TotalAmount  int64     `json:"total_amount"`
	SaldoBank    int64     `json:"saldo_bank"`
	Reason       string    `json:"reason"`
	Counterparty *string   `json:"counterparty,omitempty"`
	Note         string    `json:"note"`
	MemberID     *int64    `json:"member_id,omitempty"`
	MemberNama   *string   `json:"member_nama,omitempty"`
	Provider     *string   `json:"provider,omitempty"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type AuditorProfitLossArgs struct {
	Month time.Time
}

type AuditorProfitLossLine struct {
	Code   string `json:"code"`
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

type AuditorBalanceLine struct {
	Code   string `json:"code"`
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

type AuditorProfitLossReport struct {
	Month                string                  `json:"month"`
	GeneratedAtWIB       time.Time               `json:"generated_at_wib"`
	RevenueLines         []AuditorProfitLossLine `json:"revenue_lines"`
	CostLines            []AuditorProfitLossLine `json:"cost_lines"`
	ExpenseLines         []AuditorProfitLossLine `json:"expense_lines"`
	OtherIncomeLines     []AuditorProfitLossLine `json:"other_income_lines"`
	OtherExpenseLines    []AuditorProfitLossLine `json:"other_expense_lines"`
	NetProfit            int64                   `json:"net_profit"`
	CurrentYearProfit    int64                   `json:"current_year_profit"`
	AssetLines           []AuditorBalanceLine    `json:"asset_lines"`
	LiabilityLines       []AuditorBalanceLine    `json:"liability_lines"`
	EquityLines          []AuditorBalanceLine    `json:"equity_lines"`
	TotalAsset           int64                   `json:"total_asset"`
	TotalLiability       int64                   `json:"total_liability"`
	TotalEquity          int64                   `json:"total_equity"`
	TotalLiabilityEquity int64                   `json:"total_liability_equity"`
	OpeningBalances      []OpeningBalanceRow     `json:"opening_balances"`
}
