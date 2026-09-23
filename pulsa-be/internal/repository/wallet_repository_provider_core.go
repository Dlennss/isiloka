package repository

import "time"

type ProviderWalletTxIn struct {
	Provider            string
	RefID               string
	Arah                string
	Jumlah              int64
	Alasan              string
	Catatan             string
	TransaksiMemberID   *int64
	TransaksiProviderID *int64
	DiubahOleh          *int64
	MetaJSON            []byte
}

type ProviderWalletMutasiRow struct {
	ID                  int64     `json:"id"`
	Provider            string    `json:"provider"`
	BankID              *int64    `json:"bank_id,omitempty"`
	BankNama            *string   `json:"bank_nama,omitempty"`
	RefID               string    `json:"ref_id"`
	Arah                string    `json:"arah"`
	Jumlah              int64     `json:"jumlah"`
	Alasan              string    `json:"alasan"`
	Catatan             string    `json:"catatan"`
	SaldoSebelum        int64     `json:"saldo_sebelum"`
	SaldoSesudah        int64     `json:"saldo_sesudah"`
	TransaksiMemberID   *int64    `json:"transaksi_member_id,omitempty"`
	TransaksiProviderID *int64    `json:"transaksi_provider_id,omitempty"`
	DiubahOleh          *int64    `json:"diubah_oleh,omitempty"`
	DiubahOlehNama      *string   `json:"diubah_oleh_nama,omitempty"`
	DibuatPada          time.Time `json:"dibuat_pada"`
}
