package repository

import "time"

type AppOrderGuestRefundRow struct {
	ID              int64      `json:"id"`
	AppOrderID      int64      `json:"app_order_id"`
	InvoiceID       string     `json:"invoice_id"`
	GuestNama       *string    `json:"guest_nama,omitempty"`
	GuestEmail      *string    `json:"guest_email,omitempty"`
	GuestPhone      *string    `json:"guest_phone,omitempty"`
	AmountRefund    int64      `json:"amount_refund"`
	Status          string     `json:"status"`
	Reason          *string    `json:"reason,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	ClaimedMemberID *int64     `json:"claimed_member_id,omitempty"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	DibuatPada      *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada      *time.Time `json:"diubah_pada,omitempty"`
}

type AdminGuestRefundPendingRow struct {
	ID                 int64      `json:"id"`
	AppOrderID         int64      `json:"app_order_id"`
	InvoiceID          string     `json:"invoice_id"`
	GuestNama          *string    `json:"guest_nama,omitempty"`
	GuestEmail         *string    `json:"guest_email,omitempty"`
	GuestPhone         *string    `json:"guest_phone,omitempty"`
	AmountRefund       int64      `json:"amount_refund"`
	Status             string     `json:"status"`
	Reason             *string    `json:"reason,omitempty"`
	ClaimedMemberID    *int64     `json:"claimed_member_id,omitempty"`
	ProdukNamaSnapshot *string    `json:"produk_nama_snapshot,omitempty"`
	Dest               *string    `json:"dest,omitempty"`
	OrderStatus        *string    `json:"order_status,omitempty"`
	DibuatPada         *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada         *time.Time `json:"diubah_pada,omitempty"`
}
