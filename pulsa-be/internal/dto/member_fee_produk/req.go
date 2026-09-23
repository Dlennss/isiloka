package memberfeeprodukdto

type UpsertRequest struct {
	MemberID   int64    `json:"member_id"`
	ProdukID   int64    `json:"produk_id"`
	KodeProduk string   `json:"kode_produk"`
	FeePersen  *float64 `json:"fee_persen"`
	FeeRp      *int64   `json:"fee_rp"`
}

type DeleteRequest struct {
	MemberID   int64  `json:"member_id"`
	ProdukID   int64  `json:"produk_id"`
	KodeProduk string `json:"kode_produk"`
}
