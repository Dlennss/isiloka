package apporderrefunddto

type ClaimGuestRefundRequest struct {
	InvoiceID  string `json:"invoice_id"`
	GuestEmail string `json:"guest_email"`
	GuestPhone string `json:"guest_phone"`
}
