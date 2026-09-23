package providerreportingdto

import "time"

type TransactionRow struct {
	ID                int64     `json:"id"`
	DibuatPada        time.Time `json:"dibuat_pada"`
	Provider          string    `json:"provider"`
	StatusProvider    string    `json:"status_provider"`
	TransaksiMemberID int64     `json:"transaksi_member_id"`
	MemberID          *int64    `json:"member_id,omitempty"`
	MemberNama        *string   `json:"member_nama,omitempty"`
	RefID             string    `json:"ref_id"`
	Perintah          string    `json:"perintah"`
	KodeProduk        string    `json:"kode_produk"`
	Tujuan            string    `json:"tujuan"`
	Qty               int64     `json:"qty"`
	TrxIDJavapay      *string   `json:"trx_id_javapay,omitempty"`
	TrxIDProvider     *string   `json:"trx_id_provider,omitempty"`
	KodeRespon        *string   `json:"kode_respon,omitempty"`
	Pesan             *string   `json:"pesan,omitempty"`
	NoReferensi       *string   `json:"no_referensi,omitempty"`
	Harga             *int64    `json:"harga,omitempty"`
	SaldoTerakhir     *int64    `json:"saldo_terakhir,omitempty"`
	HTTPStatus        *int      `json:"http_status,omitempty"`
	Percobaan         int       `json:"percobaan"`
}

type AnalyticsRow struct {
	Provider       string `json:"provider"`
	Total          int64  `json:"total"`
	Success        int64  `json:"success"`
	Failed         int64  `json:"failed"`
	SumQty         int64  `json:"sum_qty"`
	SumHarga       int64  `json:"sum_harga"`
	SuccessNominal int64  `json:"success_nominal"`
	DepositAmount  int64  `json:"deposit_amount"`
}

type AnalyticsPeriodRow struct {
	Provider       string    `json:"provider"`
	PeriodStart    time.Time `json:"period_start"`
	SuccessCount   int64     `json:"success_count"`
	SuccessNominal int64     `json:"success_nominal"`
	DepositAmount  int64     `json:"deposit_amount"`
}

type DailyProductSuccessRow struct {
	PeriodStart        time.Time `json:"period_start"`
	InternalSKU        string    `json:"internal_sku"`
	ProductName        string    `json:"product_name"`
	GroupName          string    `json:"group_name"`
	SuccessCount       int64     `json:"success_count"`
	TotalQty           int64     `json:"total_qty"`
	TotalQtyProvider   int64     `json:"total_qty_provider"`
	TotalProviderPrice int64     `json:"total_provider_price"`
	TotalMemberPrice   int64     `json:"total_member_price"`
	TotalMargin        int64     `json:"total_margin"`
	ProviderCount      int64     `json:"provider_count"`
	Providers          string    `json:"providers"`
	FirstSuccessAt     time.Time `json:"first_success_at"`
	LastSuccessAt      time.Time `json:"last_success_at"`
}

type DailyProductSuccessSummary struct {
	GroupCount         int64 `json:"group_count"`
	UniqueSKUCount     int64 `json:"unique_sku_count"`
	SuccessCount       int64 `json:"success_count"`
	TotalQty           int64 `json:"total_qty"`
	TotalQtyProvider   int64 `json:"total_qty_provider"`
	TotalProviderPrice int64 `json:"total_provider_price"`
	TotalMemberPrice   int64 `json:"total_member_price"`
	TotalMargin        int64 `json:"total_margin"`
}

type DailyProductSuccessResponse struct {
	Ok      bool                       `json:"ok"`
	Items   []DailyProductSuccessRow   `json:"items"`
	Total   int64                      `json:"total"`
	Summary DailyProductSuccessSummary `json:"summary"`
}

type ListTransactionsResponse struct {
	Ok    bool             `json:"ok"`
	Items []TransactionRow `json:"items"`
	Total int64            `json:"total"`
}

type AnalyticsResponse struct {
	Ok           bool                 `json:"ok"`
	Items        []AnalyticsRow       `json:"items"`
	DailyItems   []AnalyticsPeriodRow `json:"daily_items"`
	MonthlyItems []AnalyticsPeriodRow `json:"monthly_items"`
}

type AnomalyRow struct {
	ID               int64     `json:"id"`
	DibuatPada       time.Time `json:"dibuat_pada"`
	Provider         string    `json:"provider"`
	RefID            *string   `json:"ref_id,omitempty"`
	KodeRespon       *string   `json:"kode_respon,omitempty"`
	Pesan            *string   `json:"pesan,omitempty"`
	Harga            *int64    `json:"harga,omitempty"`
	Tujuan           *string   `json:"tujuan,omitempty"`
	Qty              *int64    `json:"qty,omitempty"`
	PayloadHash      *string   `json:"payload_hash,omitempty"`
	IsDuplicate      bool      `json:"is_duplicate"`
	IsSuspectedFraud bool      `json:"is_suspected_fraud"`
	FraudReason      *string   `json:"fraud_reason,omitempty"`
	RawQuery         *string   `json:"raw_query,omitempty"`
	RawBody          *string   `json:"raw_body,omitempty"`
}

type ListAnomaliesResponse struct {
	Ok    bool         `json:"ok"`
	Items []AnomalyRow `json:"items"`
	Total int64        `json:"total"`
}
