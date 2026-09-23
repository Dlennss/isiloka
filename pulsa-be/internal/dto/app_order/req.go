package apporderdto

type CreateRequest struct {
	ProdukID             int64  `json:"produk_id"`
	Dest                 string `json:"dest"`
	Qty                  int64  `json:"qty"`
	GuestNama            string `json:"guest_nama"`
	GuestEmail           string `json:"guest_email"`
	GuestPhone           string `json:"guest_phone"`
	SourceCheckInvoiceID string `json:"source_check_invoice_id"`
	SourceCheckRefID     string `json:"source_check_ref_id"`
}
