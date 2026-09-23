package repository

import (
	"context"
	"strings"

	"github.com/lib/pq"
)

func (r *UserRepository) SumSaldo(ctx context.Context, q, role, scope string) (int64, error) {
	search := "%" + strings.TrimSpace(strings.ToLower(q)) + "%"
	role = strings.TrimSpace(strings.ToLower(role))
	scopeRoles := rolesForScope(scope)

	var sqlQ strings.Builder
	sqlQ.WriteString(`
SELECT COALESCE(SUM(COALESCE(d.saldo,0)), 0)
FROM public.member m
LEFT JOIN public.dompet_member d ON d.member_id = m.id
WHERE ($1 = '%%' OR lower(m.email) LIKE $1 OR lower(COALESCE(m.nama,'')) LIKE $1)
  AND ($2 = '' OR lower(m.role) = $2)`)

	args := []any{search, role}
	if len(scopeRoles) > 0 {
		sqlQ.WriteString(" AND lower(m.role) = ANY($3)")
		args = append(args, pq.Array(scopeRoles))
	}

	var totalSaldo int64
	if err := r.db.QueryRowContext(ctx, sqlQ.String(), args...).Scan(&totalSaldo); err != nil {
		return 0, err
	}
	return totalSaldo, nil
}

func (r *UserRepository) Count(ctx context.Context, q, role, scope string) (int64, error) {
	search := "%" + strings.TrimSpace(strings.ToLower(q)) + "%"
	role = strings.TrimSpace(strings.ToLower(role))
	scopeRoles := rolesForScope(scope)

	var sqlQ strings.Builder
	sqlQ.WriteString(`
SELECT COUNT(*)
FROM public.member m
WHERE ($1 = '%%' OR lower(m.email) LIKE $1 OR lower(COALESCE(m.nama,'')) LIKE $1)
  AND ($2 = '' OR lower(m.role) = $2)`)

	args := []any{search, role}
	if len(scopeRoles) > 0 {
		sqlQ.WriteString(" AND lower(m.role) = ANY($3)")
		args = append(args, pq.Array(scopeRoles))
	}

	var totalCount int64
	if err := r.db.QueryRowContext(ctx, sqlQ.String(), args...).Scan(&totalCount); err != nil {
		return 0, err
	}
	return totalCount, nil
}
