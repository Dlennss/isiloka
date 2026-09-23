package repository

import (
	"database/sql"
)

type MemberSelfRepository struct {
	db *sql.DB
}

func NewMemberSelfRepository(db *sql.DB) *MemberSelfRepository {
	return &MemberSelfRepository{db: db}
}

type MemberProfile struct {
	ID             int64  `json:"id"`
	Email          string `json:"email"`
	Nama           string `json:"nama"`
	Phone          string `json:"phone"`
	StoreName      string `json:"store_name"`
	ProfilePhoto   string `json:"profile_photo_url"`
	Role           string `json:"role"`
	Aktif          bool   `json:"aktif"`
	ChargeReceiver bool   `json:"charge_receiver"`
	Saldo          int64  `json:"saldo"`
	DibuatPada     string `json:"dibuat_pada"`
}

type MemberAPIKey struct {
	ID         int64  `json:"id"`
	MemberID   int64  `json:"member_id"`
	ApiKey     string `json:"api_key"`
	Aktif      bool   `json:"aktif"`
	DibuatPada string `json:"dibuat_pada"`
}

type MemberIPWhitelist struct {
	ID         int64   `json:"id"`
	IP         string  `json:"ip"`
	Label      *string `json:"label,omitempty"`
	WebhookURL *string `json:"webhook_url,omitempty"`
	Aktif      bool    `json:"aktif"`
}

type MemberMonthStat struct {
	Year         int   `json:"year"`
	Month        int   `json:"month"`
	DepositCount int64 `json:"deposit_count"`
	DepositSum   int64 `json:"deposit_sum"`
	TrxCount     int64 `json:"trx_count"`
	TrxSum       int64 `json:"trx_sum"`
	SuccessCount int64 `json:"success_count"`
	SuccessSum   int64 `json:"success_sum"`
	FailedCount  int64 `json:"failed_count"`
	FailedSum    int64 `json:"failed_sum"`
}

type MemberOverallStat struct {
	DepositCount           int64                     `json:"deposit_count"`
	DepositSum             int64                     `json:"deposit_sum"`
	SuccessCount           int64                     `json:"success_count"`
	SuccessSum             int64                     `json:"success_sum"`
	FailedCount            int64                     `json:"failed_count"`
	FailedSum              int64                     `json:"failed_sum"`
	OtherMutationNet       int64                     `json:"other_mutation_net"`
	OtherMutationBreakdown []MemberOtherMutationStat `json:"other_mutation_breakdown"`
	LedgerBalance          int64                     `json:"ledger_balance"`
	SaldoReconciled        bool                      `json:"saldo_reconciled"`
}

type MemberOtherMutationStat struct {
	Alasan     string `json:"alasan"`
	EntryCount int64  `json:"entry_count"`
	NetAmount  int64  `json:"net_amount"`
}
