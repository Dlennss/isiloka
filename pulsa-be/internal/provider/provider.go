package provider

import "context"

// PayRequest — request seragam untuk semua provider
type PayRequest struct {
	Command    string // PAY/INQ dari request member
	Product    string // kode provider (GPYOPEN, DND, DANARP, dll)
	Mode       string // mode eksekusi provider dari mapping DB (opsional)
	Dest       string // tujuan (nomor HP, nomor rekening)
	Qty        int64  // nominal
	RefID      string // referensi transaksi
	HP         string // nomor HP end user bila provider mewajibkan
	Berita     string // catatan/berita transfer bila provider mendukung
	MerchantID string // id merchant mitra bila provider mendukung
}

// PayResponse — response seragam dari semua provider
type PayResponse struct {
	HTTPStatus  int
	Body        string         // raw response body
	RC          string         // kode respon
	Message     string         // pesan dari provider
	ProviderRef string         // nomor referensi provider
	Price       int64          // harga dari provider
	Balance     int64          // saldo terakhir provider (0 jika tidak ada)
	Raw         map[string]any // response mentah untuk logging
	RequestRaw  map[string]any // request final yang benar-benar dikirim
}

// Client — interface yang harus diimplementasi setiap provider
type Client interface {
	// Name — nama provider (yuscom, javapay, smb, dll)
	Name() string
	// Pay — kirim transaksi ke provider
	Pay(ctx context.Context, req PayRequest) (*PayResponse, error)
}
