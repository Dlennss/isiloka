package repository

import (
	"database/sql"
	"strconv"
)

type ProdukAppPricingRepository struct {
	db *sql.DB
}

func NewProdukAppPricingRepository(db *sql.DB) *ProdukAppPricingRepository {
	return &ProdukAppPricingRepository{db: db}
}
func itoa(n int) string {
	return strconv.Itoa(n)
}
