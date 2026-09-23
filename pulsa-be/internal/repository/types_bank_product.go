package repository

import "time"

type MasterSimpleRow struct {
	ID         int64     `json:"id"`
	Nama       string    `json:"nama"`
	Aktif      bool      `json:"aktif"`
	Keterangan string    `json:"keterangan,omitempty"`
	DibuatPada time.Time `json:"dibuat_pada"`
	DiubahPada time.Time `json:"diubah_pada"`
}

type BankRow struct {
	ID             int64      `json:"id"`
	Nama           string     `json:"nama"`
	NomorRekening  string     `json:"nomor_rekening"`
	AtasNama       string     `json:"atas_nama"`
	Saldo          int64      `json:"saldo"`
	Aktif          bool       `json:"aktif"`
	AdminStaffOnly bool       `json:"admin_staff_only"`
	DibuatPada     *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada     *time.Time `json:"diubah_pada,omitempty"`
}

type BankUpsertInput struct {
	ID             int64  `json:"id"`
	Nama           string `json:"nama"`
	NomorRekening  string `json:"nomor_rekening"`
	AtasNama       string `json:"atas_nama"`
	Saldo          int64  `json:"saldo"`
	Aktif          bool   `json:"aktif"`
	AdminStaffOnly bool   `json:"admin_staff_only"`
}

type BankMutasiRow struct {
	ID              int64      `json:"id"`
	BankID          int64      `json:"bank_id"`
	BankNama        string     `json:"bank_nama"`
	RefID           string     `json:"ref_id"`
	Arah            string     `json:"arah"`
	Jumlah          int64      `json:"jumlah"`
	Alasan          string     `json:"alasan"`
	Catatan         string     `json:"catatan"`
	SaldoSebelum    int64      `json:"saldo_sebelum"`
	SaldoSesudah    int64      `json:"saldo_sesudah"`
	Provider        *string    `json:"provider,omitempty"`
	TargetRefID     *string    `json:"target_ref_id,omitempty"`
	MemberID        *int64     `json:"member_id,omitempty"`
	MemberNama      *string    `json:"member_nama,omitempty"`
	DiubahOleh      *int64     `json:"diubah_oleh,omitempty"`
	DiubahNama      *string    `json:"diubah_oleh_nama,omitempty"`
	WaktuMutasiBank *time.Time `json:"waktu_mutasi_bank,omitempty"`
	Pengirim        *string    `json:"pengirim,omitempty"`
	Penerima        *string    `json:"penerima,omitempty"`
	DibuatPada      time.Time  `json:"dibuat_pada"`
}

type ProdukRow struct {
	ID              int64      `json:"id"`
	SKU             string     `json:"sku"`
	Nama            string     `json:"nama"`
	GroupName       string     `json:"group_name"`
	KategoriID      int64      `json:"kategori_id"`
	KategoriNama    string     `json:"kategori_nama"`
	BrandID         int64      `json:"brand_id"`
	BrandNama       string     `json:"brand_nama"`
	TipeHarga       string     `json:"tipe_harga"`
	Nominal         *int64     `json:"nominal,omitempty"`
	MaksimalNominal *int64     `json:"maksimal_nominal,omitempty"`
	JamBuka         string     `json:"jam_buka"`
	JamTutup        string     `json:"jam_tutup"`
	Aktif           bool       `json:"aktif"`
	DibuatPada      *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada      *time.Time `json:"diubah_pada,omitempty"`
}

type ProdukUpsertInput struct {
	ID              int64
	SKU             string
	Nama            string
	GroupName       string
	KategoriID      int64
	BrandID         int64
	TipeHarga       string
	Nominal         *int64
	MaksimalNominal *int64
	JamBuka         string
	JamTutup        string
	Aktif           bool
}

type MemberFeeProductRow struct {
	ID         int64    `json:"id"`
	MemberID   int64    `json:"member_id"`
	ProdukID   int64    `json:"produk_id"`
	KodeProduk string   `json:"kode_produk"`
	NamaProduk string   `json:"nama_produk"`
	FeePersen  *float64 `json:"fee_persen,omitempty"`
	FeeRp      *int64   `json:"fee_rp,omitempty"`
}

type H2HInitialFeeSetupInput struct {
	FeeDana    int64
	FeeGopay   int64
	FeeLinkAja int64
	FeeOvo     int64
	FeeShopee  int64
	FeeBank    int64
	FeeLainnya int64
}

type ProdukProviderMapRow struct {
	ID              int64      `json:"id"`
	ProdukID        int64      `json:"produk_id"`
	ProdukSKU       string     `json:"produk_sku"`
	ProdukNama      string     `json:"produk_nama"`
	TipeHarga       string     `json:"tipe_harga"`
	Nominal         *int64     `json:"nominal,omitempty"`
	MinimalNominal  *int64     `json:"minimal_nominal,omitempty"`
	MaksimalNominal *int64     `json:"maksimal_nominal,omitempty"`
	Provider        string     `json:"provider"`
	KodeProvider    string     `json:"kode_provider"`
	SpecialCode     *string    `json:"special_code,omitempty"`
	Mode            *string    `json:"mode,omitempty"`
	FeeRp           int64      `json:"fee_rp"`
	JamBuka         *string    `json:"jam_buka,omitempty"`
	JamTutup        *string    `json:"jam_tutup,omitempty"`
	Aktif           bool       `json:"aktif"`
	DibuatPada      *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada      *time.Time `json:"diubah_pada,omitempty"`
}

type ProdukProviderMapUpsertInput struct {
	ID              int64
	ProdukID        int64
	Provider        string
	KodeProvider    string
	SpecialCode     *string
	Mode            *string
	MinimalNominal  *int64
	MaksimalNominal *int64
	FeeRp           int64
	JamBuka         *string
	JamTutup        *string
	Aktif           bool
}

type ProviderOpenAmountRule struct {
	KodeProvider    string
	SpecialCode     *string
	Mode            *string
	MinimalNominal  *int64
	MaksimalNominal *int64
}

type ProviderRouteCandidate struct {
	ProdukProviderMapID *int64
	ProdukSKUSnapshot   string
	Provider            string
	KodeProvider        string
	SpecialCode         *string
	Mode                *string
	MinimalNominal      *int64
	MaksimalNominal     *int64
}

type KategoriFeeAppRow struct {
	ID           int64      `json:"id"`
	KategoriID   int64      `json:"kategori_id"`
	KategoriNama string     `json:"kategori_nama"`
	FeeMaster    int64      `json:"fee_master"`
	FeeAgent     int64      `json:"fee_agent"`
	FeeUser      int64      `json:"fee_user"`
	FeeNonUser   int64      `json:"fee_non_user"`
	Aktif        bool       `json:"aktif"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type KategoriFeeAppUpsertInput struct {
	ID         int64
	KategoriID int64
	FeeMaster  int64
	FeeAgent   int64
	FeeUser    int64
	FeeNonUser int64
	Aktif      bool
}

type ProdukAppPricingRow struct {
	ID                 int64      `json:"id"`
	ProdukID           int64      `json:"produk_id"`
	ProdukSKU          string     `json:"produk_sku"`
	ProdukNama         string     `json:"produk_nama"`
	KategoriID         int64      `json:"kategori_id"`
	KategoriNama       string     `json:"kategori_nama"`
	Provider           string     `json:"provider"`
	Harga              int64      `json:"harga"`
	Aktif              bool       `json:"aktif"`
	YuscomGroup        string     `json:"yuscom_group"`
	YuscomCategory     string     `json:"yuscom_category"`
	YuscomSubcategory  string     `json:"yuscom_subcategory"`
	YuscomSKU          string     `json:"yuscom_sku"`
	YuscomName         string     `json:"yuscom_name"`
	YuscomStatus       string     `json:"yuscom_status"`
	YuscomDisplayBrand string     `json:"yuscom_display_brand"`
	FetchedAt          *time.Time `json:"fetched_at,omitempty"`
	CreatedAt          *time.Time `json:"created_at,omitempty"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
}

type ProdukAppPricingUpsertInput struct {
	ID        int64
	ProdukID  int64
	Harga     int64
	Aktif     bool
	FetchedAt *time.Time
}

type AppProdukRow struct {
	ID                             int64      `json:"id"`
	SKU                            string     `json:"sku"`
	Nama                           string     `json:"nama"`
	GroupName                      string     `json:"group_name"`
	KategoriID                     int64      `json:"kategori_id"`
	KategoriNama                   string     `json:"kategori_nama"`
	BrandID                        int64      `json:"brand_id"`
	BrandNama                      string     `json:"brand_nama"`
	TipeHarga                      string     `json:"tipe_harga"`
	HargaDasarApp                  int64      `json:"harga_dasar_app"`
	Nominal                        *int64     `json:"nominal,omitempty"`
	MaksimalNominal                *int64     `json:"maksimal_nominal,omitempty"`
	FeeMaster                      int64      `json:"fee_master"`
	FeeAgent                       int64      `json:"fee_agent"`
	FeeUser                        int64      `json:"fee_user"`
	FeeGuest                       int64      `json:"fee_guest"`
	HargaMasterFinal               *int64     `json:"harga_master_final,omitempty"`
	HargaAgentFinal                *int64     `json:"harga_agent_final,omitempty"`
	HargaUserFinal                 *int64     `json:"harga_user_final,omitempty"`
	HargaGuestFinal                *int64     `json:"harga_guest_final,omitempty"`
	PaymentFeeMaster               int64      `json:"payment_fee_master"`
	PaymentFeeAgent                int64      `json:"payment_fee_agent"`
	PaymentFeeUser                 int64      `json:"payment_fee_user"`
	PaymentFeeGuest                int64      `json:"payment_fee_guest"`
	HargaMasterWithPaymentFeeFinal *int64     `json:"harga_master_with_payment_fee_final,omitempty"`
	HargaAgentWithPaymentFeeFinal  *int64     `json:"harga_agent_with_payment_fee_final,omitempty"`
	HargaUserWithPaymentFeeFinal   *int64     `json:"harga_user_with_payment_fee_final,omitempty"`
	HargaGuestWithPaymentFeeFinal  *int64     `json:"harga_guest_with_payment_fee_final,omitempty"`
	Aktif                          bool       `json:"aktif"`
	DibuatPada                     *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada                     *time.Time `json:"diubah_pada,omitempty"`
}

type H2HProdukRow struct {
	ID              int64      `json:"id"`
	SKU             string     `json:"sku"`
	Nama            string     `json:"nama"`
	GroupName       string     `json:"group_name"`
	KategoriID      int64      `json:"-"`
	KategoriNama    string     `json:"kategori_nama"`
	BrandID         int64      `json:"-"`
	BrandNama       string     `json:"brand_nama"`
	TipeHarga       string     `json:"tipe_harga"`
	Harga           *int64     `json:"harga,omitempty"`
	FeeTambahan     *int64     `json:"fee_tambahan,omitempty"`
	Nominal         *int64     `json:"-"`
	MaksimalNominal *int64     `json:"maksimal_nominal,omitempty"`
	Aktif           bool       `json:"-"`
	DibuatPada      *time.Time `json:"dibuat_pada,omitempty"`
	DiubahPada      *time.Time `json:"diubah_pada,omitempty"`
}

type AppAdRow struct {
	ID         int64      `json:"id"`
	Judul      string     `json:"judul"`
	Keterangan string     `json:"keterangan"`
	ImageURL   string     `json:"image_url"`
	LinkURL    string     `json:"link_url"`
	Urutan     int        `json:"urutan"`
	Aktif      bool       `json:"aktif"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type AppAdUpsertInput struct {
	ID         int64
	Judul      string
	Keterangan string
	ImageURL   string
	LinkURL    string
	Urutan     int
	Aktif      bool
}
