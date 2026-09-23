package repository

import "time"

type ProviderRekeningRow struct {
	ID                  int64     `json:"id"`
	Provider            string    `json:"provider"`
	Nama                string    `json:"nama"`
	Bank                string    `json:"bank"`
	NomorRekening       string    `json:"nomor_rekening"`
	NomorRekeningDigits string    `json:"nomor_rekening_digits"`
	Catatan             string    `json:"catatan"`
	Aktif               bool      `json:"aktif"`
	DibuatPada          time.Time `json:"dibuat_pada"`
	DiubahPada          time.Time `json:"diubah_pada"`
}

type ProviderRekeningUpsertInput struct {
	Provider      string `json:"provider"`
	Nama          string `json:"nama"`
	Bank          string `json:"bank"`
	NomorRekening string `json:"nomor_rekening"`
	Catatan       string `json:"catatan"`
	Aktif         *bool  `json:"aktif"`
}
