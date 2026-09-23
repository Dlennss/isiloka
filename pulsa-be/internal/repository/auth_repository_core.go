package repository

import (
	"database/sql"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}
func nullIfEmptyStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableInt64Value(v *int64) any {
	if v == nil || *v <= 0 {
		return nil
	}
	return *v
}
