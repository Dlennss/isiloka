package repository

import (
	"context"
	"database/sql"
	"strings"
)

type MessagePattern struct {
	Pattern string
	Status  string
}

func LoadMessagePatterns(ctx context.Context, db *sql.DB) (failed, success, pending []string, err error) {
	rows, err := db.QueryContext(ctx, `
SELECT upper(trim(pattern)), lower(trim(status))
FROM public.provider_message_pattern
WHERE aktif = true
ORDER BY id`)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pattern, status string
		if err := rows.Scan(&pattern, &status); err != nil {
			return nil, nil, nil, err
		}
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		switch status {
		case "failed":
			failed = append(failed, pattern)
		case "success":
			success = append(success, pattern)
		case "pending":
			pending = append(pending, pattern)
		}
	}
	return failed, success, pending, rows.Err()
}
