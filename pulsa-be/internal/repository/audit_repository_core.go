package repository

import (
	"context"
	"database/sql"
	"time"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

type AdminStatusMismatchRow struct {
	TransaksiMemberID   int64     `json:"transaksi_member_id"`
	RefID               string    `json:"ref_id"`
	StatusMember        string    `json:"status_member"`
	TransaksiProviderID int64     `json:"transaksi_provider_id"`
	Provider            string    `json:"provider"`
	StatusProvider      *string   `json:"status_provider,omitempty"`
	MemberCreated       time.Time `json:"member_created"`
	ProviderCreated     time.Time `json:"provider_created"`
}

type AdminGuestRefundMissingRow struct {
	AppOrderID   int64     `json:"app_order_id"`
	InvoiceID    string    `json:"invoice_id"`
	BuyerType    string    `json:"buyer_type"`
	Status       string    `json:"status"`
	MemberID     *int64    `json:"member_id,omitempty"`
	GuestNama    *string   `json:"guest_nama,omitempty"`
	GuestEmail   *string   `json:"guest_email,omitempty"`
	GuestPhone   *string   `json:"guest_phone,omitempty"`
	HargaFinal   int64     `json:"harga_final"`
	DibuatPada   time.Time `json:"dibuat_pada"`
	DiubahPada   time.Time `json:"diubah_pada"`
	ProviderTrx  *int64    `json:"provider_trx_id,omitempty"`
	ProviderInfo *string   `json:"provider_info,omitempty"`
}

type AdminProviderEmptyResponseRow struct {
	TransaksiProviderID int64     `json:"transaksi_provider_id"`
	TransaksiMemberID   int64     `json:"transaksi_member_id"`
	MemberID            *int64    `json:"member_id,omitempty"`
	MemberNama          *string   `json:"member_nama,omitempty"`
	Provider            string    `json:"provider"`
	RefID               string    `json:"ref_id"`
	Perintah            string    `json:"perintah"`
	KodeProduk          string    `json:"kode_produk"`
	Tujuan              string    `json:"tujuan"`
	Qty                 int64     `json:"qty"`
	KodeRespon          *string   `json:"kode_respon,omitempty"`
	Pesan               *string   `json:"pesan,omitempty"`
	Harga               *int64    `json:"harga,omitempty"`
	HTTPStatus          *int      `json:"http_status,omitempty"`
	Percobaan           int       `json:"percobaan"`
	DibuatPada          time.Time `json:"dibuat_pada"`
	NoReferensi         *string   `json:"no_referensi,omitempty"`
}

type AdminProviderWalletMissingDebitRow struct {
	TransaksiMemberID   int64     `json:"transaksi_member_id"`
	MemberID            int64     `json:"member_id"`
	StatusMember        string    `json:"status_member"`
	RefIDMember         string    `json:"ref_id_member"`
	ProdukMember        string    `json:"produk_member"`
	TujuanMember        string    `json:"tujuan_member"`
	QtyMember           int64     `json:"qty_member"`
	QtyProvider         *int64    `json:"qty_provider,omitempty"`
	TransaksiProviderID int64     `json:"transaksi_provider_id"`
	Provider            string    `json:"provider"`
	RefIDProvider       string    `json:"ref_id_provider"`
	Perintah            string    `json:"perintah"`
	KodeProduk          string    `json:"kode_produk"`
	Tujuan              string    `json:"tujuan"`
	Qty                 int64     `json:"qty"`
	Harga               int64     `json:"harga"`
	KodeRespon          *string   `json:"kode_respon,omitempty"`
	Pesan               *string   `json:"pesan,omitempty"`
	NoReferensi         *string   `json:"no_referensi,omitempty"`
	ProviderDibuatPada  time.Time `json:"provider_dibuat_pada"`
}

type AdminProviderSuccessSuspiciousMessageRow struct {
	TransaksiMemberID      int64      `json:"transaksi_member_id"`
	MemberID               *int64     `json:"member_id,omitempty"`
	MemberNama             *string    `json:"member_nama,omitempty"`
	StatusMember           *string    `json:"status_member,omitempty"`
	RefIDMember            *string    `json:"ref_id_member,omitempty"`
	ProdukMember           *string    `json:"produk_member,omitempty"`
	TujuanMember           *string    `json:"tujuan_member,omitempty"`
	QtyMember              *int64     `json:"qty_member,omitempty"`
	TransaksiProviderID    int64      `json:"transaksi_provider_id"`
	Provider               string     `json:"provider"`
	StatusProvider         string     `json:"status_provider"`
	RefIDProvider          string     `json:"ref_id_provider"`
	Perintah               string     `json:"perintah"`
	KodeProduk             string     `json:"kode_produk"`
	Tujuan                 string     `json:"tujuan"`
	Qty                    int64      `json:"qty"`
	NominalProviderRequest *int64     `json:"nominal_provider_request,omitempty"`
	Harga                  *int64     `json:"harga,omitempty"`
	KodeRespon             *string    `json:"kode_respon,omitempty"`
	Pesan                  string     `json:"pesan"`
	NoReferensi            *string    `json:"no_referensi,omitempty"`
	ProviderDibuatPada     time.Time  `json:"provider_dibuat_pada"`
	Resolved               bool       `json:"resolved"`
	ResolvedAt             *time.Time `json:"resolved_at,omitempty"`
	ResolvedByUserID       *int64     `json:"resolved_by_user_id,omitempty"`
	ResolvedByName         *string    `json:"resolved_by_name,omitempty"`
	ResolveNote            *string    `json:"resolve_note,omitempty"`
	CanRetryRefund         bool       `json:"can_retry_refund"`
}

type AdminProviderSuccessSuspiciousResendTarget struct {
	TransaksiProviderID int64
	TransaksiMemberID   int64
	RefID               string
	ProdukMember        string
	Tujuan              string
	Qty                 int64
	KodeProdukProvider  string
	Provider            string
	StatusProvider      string
	Pesan               string
	ProdukProviderMapID *int64
}

type MarkingTrxResolveResult struct {
	Resolved bool  `json:"resolved"`
	TrxID    int64 `json:"trx_id"`
	UserID   int64 `json:"user_id"`
}

type ProviderSuspectBankSettlementResult struct {
	Resolved             bool   `json:"resolved"`
	AlreadySettled       bool   `json:"already_settled"`
	TransaksiProviderID  int64  `json:"transaksi_provider_id"`
	TransaksiMemberID    int64  `json:"transaksi_member_id"`
	BankID               int64  `json:"bank_id"`
	BankNama             string `json:"bank_nama"`
	BankNomorRekening    string `json:"bank_nomor_rekening"`
	Nominal              int64  `json:"nominal"`
	Fee                  int64  `json:"fee"`
	TotalAmount          int64  `json:"total_amount"`
	SaldoSebelum         int64  `json:"saldo_sebelum"`
	SaldoSesudah         int64  `json:"saldo_sesudah"`
	MutasiBankID         int64  `json:"mutasi_bank_id"`
	RefID                string `json:"ref_id"`
	ProviderRefID        string `json:"provider_ref_id"`
	TransferAmountSource int64  `json:"transfer_amount_source"`
}

type BulkMarkingTrxResolveResult struct {
	Processed     int              `json:"processed"`
	ResolvedCount int              `json:"resolved_count"`
	FailedCount   int              `json:"failed_count"`
	ProcessedIDs  []int64          `json:"processed_ids,omitempty"`
	Failed        []map[string]any `json:"failed,omitempty"`
}

type ResolveProviderWalletMissingDebitResult struct {
	Resolved            bool                                `json:"resolved"`
	AlreadyResolved     bool                                `json:"already_resolved"`
	SaldoSebelum        int64                               `json:"saldo_sebelum"`
	SaldoSesudah        int64                               `json:"saldo_sesudah"`
	TransaksiProviderID int64                               `json:"transaksi_provider_id"`
	Row                 *AdminProviderWalletMissingDebitRow `json:"row,omitempty"`
}

type BulkResolveProviderWalletMissingDebitResult struct {
	Processed     int     `json:"processed"`
	ResolvedCount int     `json:"resolved_count"`
	AlreadyCount  int     `json:"already_count"`
	ProcessedIDs  []int64 `json:"processed_ids,omitempty"`
}

type BulkIgnoreProviderWalletMissingDebitResult struct {
	Processed    int     `json:"processed"`
	ProcessedIDs []int64 `json:"processed_ids,omitempty"`
}

func (r *AuditRepository) ensureProviderWalletMissingDebitIgnoreTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS public.audit_provider_wallet_missing_debit_ignore (
  id BIGSERIAL PRIMARY KEY,
  transaksi_provider_id BIGINT NOT NULL UNIQUE REFERENCES public.transaksi_provider(id) ON DELETE CASCADE,
  ignored_by_admin_id BIGINT NOT NULL REFERENCES public.member(id),
  note TEXT NOT NULL DEFAULT '',
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
	return err
}

func (r *AuditRepository) DB() *sql.DB {
	return r.db
}
