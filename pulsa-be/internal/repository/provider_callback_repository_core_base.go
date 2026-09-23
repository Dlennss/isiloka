package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ProviderCallbackRepository struct {
	db *sql.DB
}

func NewProviderCallbackRepository(db *sql.DB) *ProviderCallbackRepository {
	return &ProviderCallbackRepository{db: db}
}

func (r *ProviderCallbackRepository) DB() *sql.DB {
	return r.db
}

func (r *ProviderCallbackRepository) normalizeProviderForCallback(_ context.Context, provider string) (string, error) {
	p := strings.TrimSpace(strings.ToLower(provider))
	switch p {
	case "talenta":
		p = "talentapay"
	case "sagara":
		p = "sagaramobile"
	}
	switch p {
	case "yuscom", "talentapay", "multikom", "javapay", "sagaramobile", "minions", "trionik", "ajs", "gemilang", "smb", "loketbayar", "chytron", "pulsa24jam":
		return p, nil
	default:
		if p == "" {
			return "", fmt.Errorf("provider required")
		}
		return "", fmt.Errorf("provider tidak valid")
	}
}

func nowWIB() time.Time {
	return time.Now().In(time.FixedZone("WIB", 7*60*60))
}
