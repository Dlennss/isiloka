package repository

import "time"

type AppOrderListFilter struct {
	MemberID  int64
	All       bool
	Q         string
	BuyerType string
	Status    string
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}

type AppOrderProviderTrxListFilter struct {
	Provider  string
	Status    string
	RefID     string
	InvoiceID string
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}
