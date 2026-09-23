package repository

import (
	"database/sql"
	"strings"
)

type AdminBusinessReportRepository struct {
	db *sql.DB
}

func NewAdminBusinessReportRepository(db *sql.DB) *AdminBusinessReportRepository {
	return &AdminBusinessReportRepository{db: db}
}

func normalizeAdminReportScope(scope string) string {
	scope = strings.ToLower(strings.TrimSpace(scope))
	switch scope {
	case "", "all", "retail", "h2h":
		return scope
	default:
		return ""
	}
}
