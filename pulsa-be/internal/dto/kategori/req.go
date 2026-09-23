package kategoridto

type Request struct {
	Nama  string `json:"nama"`
	Aktif *bool  `json:"aktif"`
}
