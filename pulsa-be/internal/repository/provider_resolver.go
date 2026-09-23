package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func resolveProviderName(ctx context.Context, db *sql.DB, raw string) (string, error) {
	p := strings.TrimSpace(strings.ToLower(raw))
	if p == "" {
		return "", errors.New("unknown provider")
	}

	var normalized string
	err := db.QueryRowContext(ctx, `
SELECT lower(trim(nama))
FROM public.provider
WHERE lower(trim(nama)) = $1
LIMIT 1
`, p).Scan(&normalized)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("unknown provider: %s", p)
		}
		return "", err
	}
	return normalized, nil
}
