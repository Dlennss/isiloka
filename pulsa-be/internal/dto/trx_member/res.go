package trxmemberdto

import commondto "pulsa2/internal/dto/common"
import "time"

type ErrorResponse = commondto.ErrorResponse

type ExistingResponse struct {
	Ok              bool `json:"ok"`
	Existing        bool `json:"existing"`
	TransaksiMember any  `json:"transaksi_member,omitempty"`
}

type StatusResponse struct {
	Ok           bool              `json:"ok"`
	RefID        string            `json:"refid"`
	Status       string            `json:"status"`
	RC           string            `json:"rc"`
	Msg          string            `json:"msg"`
	ResponseKind string            `json:"response_kind,omitempty"`
	Callback     *CallbackDelivery `json:"callback,omitempty"`
}

type AlreadyFinalResponse struct {
	Ok           bool              `json:"ok"`
	AlreadyFinal bool              `json:"already_final"`
	RefID        string            `json:"refid"`
	Status       string            `json:"status"`
	WebhookRetry bool              `json:"webhook_retry"`
	ProviderJPOK bool              `json:"provider_jp_ok"`
	ResponseKind string            `json:"response_kind,omitempty"`
	Callback     *CallbackDelivery `json:"callback,omitempty"`
}

type CallbackDelivery struct {
	Attempted  bool   `json:"attempted"`
	Delivered  bool   `json:"delivered"`
	HTTPStatus int    `json:"http_status,omitempty"`
	Error      string `json:"error,omitempty"`
}

type RetryResponse struct {
	Ok       bool   `json:"ok"`
	RefID    string `json:"refid"`
	Status   string `json:"status"`
	Retry    string `json:"retry"`
	Provider string `json:"provider,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type TransaksiMemberItem struct {
	ID                    int64  `json:"id"`
	RefID                 string `json:"ref_id"`
	Status                string `json:"status"`
	QtyProvider           int64  `json:"qty_provider"`
	ChargeReceiverApplied bool   `json:"charge_receiver_applied"`
	BiayaPerkiraan        int64  `json:"biaya_perkiraan"`
	FeeMemberRp           int64  `json:"fee_member_rp"`
	Keterangan            string `json:"keterangan"`
	Provider              string `json:"provider"`
}

type TransaksiMemberResponse struct {
	Ok              bool                `json:"ok"`
	TransaksiMember TransaksiMemberItem `json:"transaksi_member"`
}

type ProdukListResponse struct {
	Ok       bool               `json:"ok"`
	Commands string             `json:"commands"`
	Product  string             `json:"product"`
	Items    []H2HProdukItemDTO `json:"items"`
}

type H2HProdukItemDTO struct {
	ID              int64      `json:"id"`
	SKU             string     `json:"sku"`
	Nama            string     `json:"nama"`
	GroupName       string     `json:"group_name"`
	KategoriNama    string     `json:"kategori_nama"`
	BrandNama       string     `json:"brand_nama"`
	TipeHarga       string     `json:"tipe_harga"`
	Harga           *int64     `json:"harga,omitempty"`
	FeeTambahan     *int64     `json:"fee_tambahan,omitempty"`
	MaksimalNominal *int64     `json:"maksimal_nominal,omitempty"`
	DibuatPada      *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada      *time.Time `json:"diubah_pada,omitempty"`
}
