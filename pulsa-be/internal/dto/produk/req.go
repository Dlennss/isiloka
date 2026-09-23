package produkdto

type Request struct {
	SKU             string `json:"sku"`
	Nama            string `json:"nama"`
	GroupName       string `json:"group_name"`
	KategoriID      int64  `json:"kategori_id"`
	BrandID         int64  `json:"brand_id"`
	TipeHarga       string `json:"tipe_harga"`
	Nominal         *int64 `json:"nominal"`
	MaksimalNominal *int64 `json:"maksimal_nominal"`
	Aktif           *bool  `json:"aktif"`
}
