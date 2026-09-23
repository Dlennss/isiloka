package trxmemberdto

type TrxRequest struct {
	Commands           string `json:"commands"`
	Product            string `json:"product"` // kode produk INTERNAL (contoh: DNID). Akan di-map ke provider.
	Dest               string `json:"dest"`
	Qty                int64  `json:"qty"`   // PAY = nominal, INQ = 1, STATUS-PAY = sama seperti PAY (format identik)
	RefID              string `json:"refid"` // WAJIB: refid dari member, dipakai juga ke provider
	PIN                string `json:"pin"`   // wajib
	HP                 string `json:"hp,omitempty"`
	Berita             string `json:"berita,omitempty"`
	MerchantID         string `json:"id_merchant,omitempty"`
	SourceSystem       string `json:"source_system,omitempty"`
	SMPAYTransactionID int64  `json:"smpay_transaction_id,omitempty"`
	SMPAYWebsiteID     int64  `json:"smpay_website_id,omitempty"`
	SMPAYDivisionID    int64  `json:"smpay_division_id,omitempty"`
	SkipH2HCommission  *bool  `json:"skip_h2h_commission,omitempty"`
}
