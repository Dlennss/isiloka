package produkprovidermapdto

type Request struct {
	ProdukID        int64   `json:"produk_id"`
	Provider        string  `json:"provider"`
	KodeProvider    string  `json:"kode_provider"`
	SpecialCode     *string `json:"special_code"`
	Mode            *string `json:"mode"`
	MinimalNominal  *int64  `json:"minimal_nominal"`
	MaksimalNominal *int64  `json:"maksimal_nominal"`
	FeeRp           int64   `json:"fee_rp"`
	JamBuka         *string `json:"jam_buka"`
	JamTutup        *string `json:"jam_tutup"`
	Aktif           *bool   `json:"aktif"`
}

type ToggleAktifRequest struct {
	Aktif bool `json:"aktif"`
}
