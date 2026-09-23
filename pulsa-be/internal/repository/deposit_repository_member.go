package repository

import (
	"context"
	"strings"
)

func (r *DepositRepository) MemberEmailMatches(ctx context.Context, memberID int64, email string) (bool, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if memberID <= 0 || email == "" {
		return false, nil
	}

	var matched bool
	err := r.db.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM public.member
	WHERE id = $1
	  AND lower(trim(COALESCE(email, ''))) = $2
)
`, memberID, email).Scan(&matched)
	return matched, err
}
