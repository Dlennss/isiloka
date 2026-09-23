package repository

import "time"

type AppOrderRow struct {
	ID                     int64                  `json:"id"`
	InvoiceID              string                 `json:"invoice_id"`
	MemberID               *int64                 `json:"member_id,omitempty"`
	MemberNama             *string                `json:"member_nama,omitempty"`
	GuestNama              *string                `json:"guest_nama,omitempty"`
	GuestEmail             *string                `json:"guest_email,omitempty"`
	GuestPhone             *string                `json:"guest_phone,omitempty"`
	ProdukID               int64                  `json:"produk_id"`
	ProdukSKUSnapshot      string                 `json:"produk_sku_snapshot"`
	ProdukNamaSnapshot     string                 `json:"produk_nama_snapshot"`
	Dest                   string                 `json:"dest"`
	Qty                    int64                  `json:"qty"`
	Nominal                int64                  `json:"nominal"`
	BuyerType              string                 `json:"buyer_type"`
	BuyerRole              string                 `json:"buyer_role"`
	HargaDasar             int64                  `json:"harga_dasar"`
	Fee                    int64                  `json:"fee"`
	HargaFinal             int64                  `json:"harga_final"`
	FeeUserSnapshot        int64                  `json:"fee_user_snapshot"`
	FeeAgentSnapshot       int64                  `json:"fee_agent_snapshot"`
	FeeMasterSnapshot      int64                  `json:"fee_master_snapshot"`
	RetailAgentIDSnapshot  *int64                 `json:"retail_agent_id_snapshot,omitempty"`
	RetailMasterIDSnapshot *int64                 `json:"retail_master_id_snapshot,omitempty"`
	Status                 string                 `json:"status"`
	Catatan                *string                `json:"catatan,omitempty"`
	AlasanGagal            *string                `json:"alasan_gagal,omitempty"`
	SN                     *string                `json:"sn,omitempty"`
	BillingInquiry         *AppBillingInquiryInfo `json:"billing_inquiry,omitempty"`
	DibuatPada             *time.Time             `json:"dibuat_pada,omitempty"`
	DiubahPada             *time.Time             `json:"diubah_pada,omitempty"`
}

type AppBillingInquiryInfo struct {
	Provider        string  `json:"provider"`
	ProviderStatus  string  `json:"provider_status"`
	ProviderMessage string  `json:"provider_message"`
	DisplayMessage  string  `json:"display_message"`
	CustomerName    *string `json:"customer_name,omitempty"`
	UsageLabel      *string `json:"usage_label,omitempty"`
	MeterType       *string `json:"meter_type,omitempty"`
	PeriodLabel     *string `json:"period_label,omitempty"`
	MeterRange      *string `json:"meter_range,omitempty"`
	BillAmount      int64   `json:"bill_amount"`
	PenaltyAmount   int64   `json:"penalty_amount"`
	AdminAmount     int64   `json:"admin_amount"`
	TotalAmount     int64   `json:"total_amount"`
	TransactionTime *string `json:"transaction_time,omitempty"`
	CanPay          bool    `json:"can_pay"`
}

type AppBillingCheckRow struct {
	ID                 int64                  `json:"id"`
	RefID              string                 `json:"ref_id"`
	MemberID           *int64                 `json:"member_id,omitempty"`
	MemberNama         *string                `json:"member_nama,omitempty"`
	GuestNama          *string                `json:"guest_nama,omitempty"`
	GuestEmail         *string                `json:"guest_email,omitempty"`
	GuestPhone         *string                `json:"guest_phone,omitempty"`
	ProdukID           int64                  `json:"produk_id"`
	ProdukSKUSnapshot  string                 `json:"produk_sku_snapshot"`
	ProdukNamaSnapshot string                 `json:"produk_nama_snapshot"`
	Dest               string                 `json:"dest"`
	BuyerType          string                 `json:"buyer_type"`
	BuyerRole          string                 `json:"buyer_role"`
	Provider           string                 `json:"provider"`
	HargaProvider      int64                  `json:"harga_provider"`
	Status             string                 `json:"status"`
	KodeRespon         *string                `json:"kode_respon,omitempty"`
	Pesan              *string                `json:"pesan,omitempty"`
	SN                 *string                `json:"sn,omitempty"`
	BillingInquiry     *AppBillingInquiryInfo `json:"billing_inquiry,omitempty"`
	DibuatPada         *time.Time             `json:"dibuat_pada,omitempty"`
	DiubahPada         *time.Time             `json:"diubah_pada,omitempty"`
}

type AppOrderPaymentRow struct {
	ID                int64      `json:"id"`
	AppOrderID        int64      `json:"app_order_id"`
	OrderID           string     `json:"order_id"`
	TransactionID     *string    `json:"transaction_id,omitempty"`
	GrossAmount       int64      `json:"gross_amount"`
	PaymentType       *string    `json:"payment_type,omitempty"`
	TransactionStatus *string    `json:"transaction_status,omitempty"`
	FraudStatus       *string    `json:"fraud_status,omitempty"`
	Acquirer          *string    `json:"acquirer,omitempty"`
	QRURL             *string    `json:"qr_url,omitempty"`
	RawRequest        *string    `json:"raw_request,omitempty"`
	RawCallback       *string    `json:"raw_callback,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	ExpiredAt         *time.Time `json:"expired_at,omitempty"`
	SettlementTime    *time.Time `json:"settlement_time,omitempty"`
	DibuatPada        *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada        *time.Time `json:"diubah_pada,omitempty"`
}

type AppOrderProviderTrxRow struct {
	ID            int64      `json:"id"`
	AppOrderID    int64      `json:"app_order_id"`
	Provider      string     `json:"provider"`
	RefID         string     `json:"ref_id"`
	HargaProvider int64      `json:"harga_provider"`
	Status        string     `json:"status"`
	KodeRespon    *string    `json:"kode_respon,omitempty"`
	Pesan         *string    `json:"pesan,omitempty"`
	SN            *string    `json:"sn,omitempty"`
	RawRequest    *string    `json:"raw_request,omitempty"`
	RawCallback   *string    `json:"raw_callback,omitempty"`
	DibuatPada    *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada    *time.Time `json:"diubah_pada,omitempty"`
}

type AppOrderProviderTrxAdminRow struct {
	ID                 int64      `json:"id"`
	AppOrderID         int64      `json:"app_order_id"`
	InvoiceID          string     `json:"invoice_id"`
	Provider           string     `json:"provider"`
	RefID              string     `json:"ref_id"`
	KodeProvider       string     `json:"kode_provider"`
	ProdukNamaSnapshot string     `json:"produk_nama_snapshot"`
	Dest               string     `json:"dest"`
	HargaProvider      int64      `json:"harga_provider"`
	Status             string     `json:"status"`
	KodeRespon         *string    `json:"kode_respon,omitempty"`
	Pesan              *string    `json:"pesan,omitempty"`
	SN                 *string    `json:"sn,omitempty"`
	DibuatPada         *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada         *time.Time `json:"diubah_pada,omitempty"`
}
