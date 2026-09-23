package memberfeekategoridto

type UpsertRequest struct {
	MemberID   int64  `json:"member_id"`
	FeeCode    string `json:"fee_code"`
	KategoriID int64  `json:"kategori_id"`
	FeeRp      int64  `json:"fee_rp"`
	Aktif      *bool  `json:"aktif"`
}

type DeleteRequest struct {
	MemberID   int64  `json:"member_id"`
	FeeCode    string `json:"fee_code"`
	KategoriID int64  `json:"kategori_id"`
}
