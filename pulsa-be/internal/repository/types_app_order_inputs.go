package repository

import "time"

type AppBillingCheckCreateInput struct {
	ID                 int64
	RefID              string
	MemberID           *int64
	BuyerRole          string
	GuestNama          string
	GuestEmail         string
	GuestPhone         string
	ProdukID           int64
	ProdukSKUSnapshot  string
	ProdukNamaSnapshot string
	Dest               string
	BuyerType          string
	Provider           string
	HargaProvider      int64
	Status             string
	KodeRespon         string
	Pesan              string
	SN                 string
	RawRequest         string
	RawCallback        string
}

type AppBillingCheckUpdateInput struct {
	ID            int64
	HargaProvider *int64
	Status        string
	KodeRespon    *string
	Pesan         *string
	SN            *string
	RawCallback   string
}

type AppOrderCreateInput struct {
	ID                     int64
	InvoiceID              string
	MemberID               *int64
	BuyerRole              string
	GuestNama              string
	GuestEmail             string
	GuestPhone             string
	SourceCheckInvoiceID   string
	SourceCheckRefID       string
	ProdukID               int64
	ProdukSKUSnapshot      string
	ProdukNamaSnapshot     string
	Dest                   string
	Qty                    int64
	Nominal                int64
	BuyerType              string
	HargaDasar             int64
	Fee                    int64
	HargaFinal             int64
	FeeUserSnapshot        int64
	FeeAgentSnapshot       int64
	FeeMasterSnapshot      int64
	RetailAgentIDSnapshot  *int64
	RetailMasterIDSnapshot *int64
	Status                 string
	Catatan                string
	AlasanGagal            string
}

type AppOrderPaymentCreateInput struct {
	ID                int64
	AppOrderID        int64
	OrderID           string
	TransactionID     string
	GrossAmount       int64
	PaymentType       string
	TransactionStatus string
	FraudStatus       string
	Acquirer          string
	QRURL             string
	RawRequest        string
	RawCallback       string
	PaidAt            *time.Time
	ExpiredAt         *time.Time
	SettlementTime    *time.Time
}

type AppOrderPaymentUpdateInput struct {
	OrderID           string
	TransactionID     string
	PaymentType       string
	TransactionStatus string
	FraudStatus       string
	Acquirer          string
	QRURL             string
	RawCallback       string
	PaidAt            *time.Time
	ExpiredAt         *time.Time
	SettlementTime    *time.Time
}

type AppOrderProviderTrxCreateInput struct {
	ID            int64
	AppOrderID    int64
	Provider      string
	RefID         string
	HargaProvider int64
	Status        string
	KodeRespon    string
	Pesan         string
	SN            string
	RawRequest    string
	RawCallback   string
}

type AppOrderProviderTrxUpdateInput struct {
	ID            int64
	HargaProvider *int64
	Status        string
	KodeRespon    string
	Pesan         string
	SN            string
	RawCallback   string
}
