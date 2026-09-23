package model

import "time"

type JavapayTrxCreateIn struct {
	Provider            string `json:"provider"` // "yuscom" | "javapay"
	TransaksiMemberID   int64  `json:"transaksi_member_id"`
	RefID               string `json:"ref_id"`
	Perintah            string `json:"perintah"`
	ProdukSKUSnapshot   string `json:"produk_sku_snapshot"`
	ProdukProviderMapID *int64 `json:"produk_provider_map_id,omitempty"`
	KodeProduk          string `json:"kode_produk"`
	Tujuan              string `json:"tujuan"`
	Qty                 int64  `json:"qty"`
}

type JavapayTrxRow struct {
	ID                  int64     `json:"id"`
	TransaksiMemberID   int64     `json:"transaksi_member_id"`
	Provider            string    `json:"provider"`
	RefID               string    `json:"ref_id"`
	Perintah            string    `json:"perintah"`
	ProdukSKUSnapshot   string    `json:"produk_sku_snapshot"`
	ProdukProviderMapID *int64    `json:"produk_provider_map_id,omitempty"`
	KodeProduk          string    `json:"kode_produk"`
	Tujuan              string    `json:"tujuan"`
	Qty                 int64     `json:"qty"`
	TrxIDJavapay        *string   `json:"trx_id_javapay,omitempty"`
	KodeRespon          *string   `json:"kode_respon,omitempty"`
	Pesan               *string   `json:"pesan,omitempty"`
	NoReferensi         *string   `json:"no_referensi,omitempty"`
	Harga               *int64    `json:"harga,omitempty"`
	SaldoTerakhir       *int64    `json:"saldo_terakhir,omitempty"`
	HTTPStatus          *int      `json:"http_status,omitempty"`
	Percobaan           int       `json:"percobaan"`
	DibuatPada          time.Time `json:"dibuat_pada"`
	RequestMentah       any       `json:"request_mentah,omitempty"`
	ResponMentah        any       `json:"respon_mentah,omitempty"`
	Status              string    `json:"status"`
}

type JavapayCallResult struct {
	HTTPStatus int            `json:"http_status"`
	Body       map[string]any `json:"body"`
}
