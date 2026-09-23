package repository

import "database/sql"

type AppOrderProviderTrxRepository struct {
	db *sql.DB
}

func NewAppOrderProviderTrxRepository(db *sql.DB) *AppOrderProviderTrxRepository {
	return &AppOrderProviderTrxRepository{db: db}
}
