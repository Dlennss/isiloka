package repository

import "time"

type ProviderMerchantIDRow struct {
	ID         int64     `json:"id"`
	Provider   string    `json:"provider"`
	MerchantID string    `json:"merchant_id"`
	Label      string    `json:"label"`
	Catatan    string    `json:"catatan"`
	Aktif      bool      `json:"aktif"`
	DibuatPada time.Time `json:"dibuat_pada"`
	DiubahPada time.Time `json:"diubah_pada"`
}

type ProviderMerchantIDUpsertInput struct {
	Provider   string `json:"provider"`
	MerchantID string `json:"merchant_id"`
	Label      string `json:"label"`
	Catatan    string `json:"catatan"`
	Aktif      *bool  `json:"aktif"`
}
