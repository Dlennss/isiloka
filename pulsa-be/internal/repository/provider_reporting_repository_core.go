package repository

import (
	"database/sql"
)

type ProviderReportingRepository struct {
	db *sql.DB
}

func NewProviderReportingRepository(db *sql.DB) *ProviderReportingRepository {
	return &ProviderReportingRepository{db: db}
}
