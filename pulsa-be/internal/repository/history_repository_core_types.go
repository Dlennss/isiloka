package repository

import (
	"database/sql"
	"errors"
	"time"
)

type HistoryRepository struct {
	db *sql.DB
}

func NewHistoryRepository(db *sql.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

type MutasiRow struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id,omitempty"`
	RefID          string    `json:"ref_id"`
	Arah           string    `json:"arah"`
	Jumlah         int64     `json:"jumlah"`
	Alasan         string    `json:"alasan"`
	Catatan        *string   `json:"catatan,omitempty"`
	SaldoSebelum   *int64    `json:"saldo_sebelum,omitempty"`
	SaldoSesudah   *int64    `json:"saldo_sesudah,omitempty"`
	DiubahOleh     *int64    `json:"diubah_oleh,omitempty"`
	DiubahOlehNama *string   `json:"diubah_oleh_nama,omitempty"`
	DibuatPada     time.Time `json:"dibuat_pada"`
}

type MutasiFilter struct {
	RefID  string
	Arah   string
	Date   string
	From   string
	To     string
	Limit  int
	Offset int
}

type TrxMemberRow struct {
	ID                    int64     `json:"id"`
	RefID                 string    `json:"ref_id"`
	Perintah              string    `json:"perintah"`
	KodeProduk            string    `json:"kode_produk"`
	Tujuan                string    `json:"tujuan"`
	Qty                   int64     `json:"qty"`
	QtyProvider           int64     `json:"qty_provider"`
	ChargeReceiverApplied bool      `json:"charge_receiver_applied"`
	FeeMemberRp           int64     `json:"fee_member_rp"`
	Status                string    `json:"status"`
	Keterangan            *string   `json:"keterangan,omitempty"`
	BiayaPerkiraan        int64     `json:"biaya_perkiraan"`
	BiayaAktual           int64     `json:"biaya_aktual"`
	DibuatPada            time.Time `json:"dibuat_pada"`
	DiperbaruiPada        time.Time `json:"diperbarui_pada"`
}

type AdminMutasiRow struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	MemberNama     *string   `json:"member_nama,omitempty"`
	RefID          string    `json:"ref_id"`
	Arah           string    `json:"arah"`
	Jumlah         int64     `json:"jumlah"`
	Alasan         string    `json:"alasan"`
	Catatan        *string   `json:"catatan,omitempty"`
	SaldoSebelum   *int64    `json:"saldo_sebelum,omitempty"`
	SaldoSesudah   *int64    `json:"saldo_sesudah,omitempty"`
	DiubahOleh     *int64    `json:"diubah_oleh,omitempty"`
	DiubahOlehNama *string   `json:"diubah_oleh_nama,omitempty"`
	DibuatPada     time.Time `json:"dibuat_pada"`
}

type AdminTrxRow struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	MemberEmail    *string   `json:"member_email,omitempty"`
	RefID          string    `json:"ref_id"`
	Perintah       string    `json:"perintah"`
	KodeProduk     string    `json:"kode_produk"`
	Tujuan         string    `json:"tujuan"`
	Qty            int64     `json:"qty"`
	Status         string    `json:"status"`
	Keterangan     *string   `json:"keterangan,omitempty"`
	HasCallbackURL bool      `json:"has_callback_url"`
	BiayaPerkiraan int64     `json:"biaya_perkiraan"`
	BiayaAktual    int64     `json:"biaya_aktual"`
	DibuatPada     time.Time `json:"dibuat_pada"`
	DiperbaruiPada time.Time `json:"diperbarui_pada"`
}

type AdminCancelTrxResult struct {
	TrxID          int64                    `json:"trx_id"`
	RefID          string                   `json:"ref_id"`
	StatusBefore   string                   `json:"status_before"`
	StatusAfter    string                   `json:"status_after"`
	RefundAmount   int64                    `json:"refund_amount"`
	SaldoSebelum   int64                    `json:"saldo_sebelum"`
	SaldoSesudah   int64                    `json:"saldo_sesudah"`
	AlasanBatal    string                   `json:"alasan_batal_admin"`
	DibatalkanOleh int64                    `json:"dibatalkan_oleh_admin_id"`
	Callback       *AdminSendCallbackResult `json:"callback,omitempty"`
	CallbackError  string                   `json:"callback_error,omitempty"`
}

type AdminCompleteTrxResult struct {
	TrxID          int64  `json:"trx_id"`
	RefID          string `json:"ref_id"`
	StatusBefore   string `json:"status_before"`
	StatusAfter    string `json:"status_after"`
	BiayaAktual    int64  `json:"biaya_aktual"`
	DiselesaikanBy int64  `json:"diselesaikan_oleh_admin_id"`
	Catatan        string `json:"catatan"`
}

type AdminTrxCallbackTarget struct {
	ID                    int64
	MemberID              int64
	RefID                 string
	Perintah              string
	KodeProduk            string
	Tujuan                string
	Qty                   int64
	QtyProvider           int64
	ChargeReceiverApplied bool
	FeeMemberRp           int64
	Status                string
	Keterangan            string
	BiayaPerkiraan        int64
	BiayaAktual           int64
	HargaMember           int64
}

type AdminSendCallbackResult struct {
	TrxID         int64  `json:"trx_id"`
	RefID         string `json:"ref_id"`
	Status        string `json:"status"`
	Provider      string `json:"provider,omitempty"`
	ProviderRef   string `json:"provider_ref,omitempty"`
	SN            string `json:"sn,omitempty"`
	CallbackHTTP  int    `json:"callback_http"`
	CallbackBody  string `json:"callback_body,omitempty"`
	WebhookURL    string `json:"webhook_url,omitempty"`
	ProcessedByID int64  `json:"processed_by_admin_id"`
}

type TrxMemberStatusLogRow struct {
	ID                 int64     `json:"id"`
	TransaksiMemberID  int64     `json:"transaksi_member_id"`
	RefID              string    `json:"ref_id"`
	StatusSebelum      *string   `json:"status_sebelum,omitempty"`
	StatusSesudah      *string   `json:"status_sesudah,omitempty"`
	KeteranganSebelum  *string   `json:"keterangan_sebelum,omitempty"`
	KeteranganSesudah  *string   `json:"keterangan_sesudah,omitempty"`
	BiayaAktualSebelum int64     `json:"biaya_aktual_sebelum"`
	BiayaAktualSesudah int64     `json:"biaya_aktual_sesudah"`
	Aksi               string    `json:"aksi"`
	DiubahOleh         *int64    `json:"diubah_oleh,omitempty"`
	DiubahOlehNama     *string   `json:"diubah_oleh_nama,omitempty"`
	DibuatPada         time.Time `json:"dibuat_pada"`
}

type TrxMemberStatusLogFilter struct {
	TrxID      int64
	RefID      string
	DiubahOleh int64
	ManualOnly bool
	From       string
	To         string
	Limit      int
	Offset     int
}

var ErrTransaksiNotPending = errors.New("transaksi tidak dalam status pending")
