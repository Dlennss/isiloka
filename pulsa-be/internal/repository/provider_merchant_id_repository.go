package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type ProviderMerchantIDRepository struct {
	db *sql.DB
}

func NewProviderMerchantIDRepository(db *sql.DB) *ProviderMerchantIDRepository {
	return &ProviderMerchantIDRepository{db: db}
}

func (r *ProviderMerchantIDRepository) List(ctx context.Context, provider, q string, aktifOnly bool, limit, offset int) ([]ProviderMerchantIDRow, int64, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	q = strings.TrimSpace(q)
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	args := []any{}
	where := []string{"1=1"}
	if provider != "" {
		args = append(args, provider)
		where = append(where, "lower(trim(provider)) = $"+strconv.Itoa(len(args)))
	}
	if q != "" {
		args = append(args, "%"+strings.ToLower(q)+"%")
		textParam := "$" + strconv.Itoa(len(args))
		where = append(where, "(lower(provider) LIKE "+textParam+" OR lower(merchant_id) LIKE "+textParam+" OR lower(label) LIKE "+textParam+" OR lower(catatan) LIKE "+textParam+")")
	}
	if aktifOnly {
		where = append(where, "aktif = true")
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	countArgs := append([]any{}, args...)
	if err := r.db.QueryRowContext(ctx, `SELECT count(*)::bigint FROM public.provider_merchant_id WHERE `+whereSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
SELECT id, provider, merchant_id, label, catatan, aktif, dibuat_pada, diubah_pada
FROM public.provider_merchant_id
WHERE `+whereSQL+`
ORDER BY lower(provider), aktif DESC, lower(label), lower(merchant_id), id
LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderMerchantIDRow, 0, limit)
	for rows.Next() {
		var item ProviderMerchantIDRow
		if err := rows.Scan(&item.ID, &item.Provider, &item.MerchantID, &item.Label, &item.Catatan, &item.Aktif, &item.DibuatPada, &item.DiubahPada); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *ProviderMerchantIDRepository) Get(ctx context.Context, id int64) (*ProviderMerchantIDRow, error) {
	var item ProviderMerchantIDRow
	err := r.db.QueryRowContext(ctx, `
SELECT id, provider, merchant_id, label, catatan, aktif, dibuat_pada, diubah_pada
FROM public.provider_merchant_id
WHERE id = $1
LIMIT 1
`, id).Scan(&item.ID, &item.Provider, &item.MerchantID, &item.Label, &item.Catatan, &item.Aktif, &item.DibuatPada, &item.DiubahPada)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ProviderMerchantIDRepository) Create(ctx context.Context, in ProviderMerchantIDUpsertInput) (int64, error) {
	normalized, err := r.normalizeInput(ctx, &in)
	if err != nil {
		return 0, err
	}
	var id int64
	err = r.db.QueryRowContext(ctx, `
INSERT INTO public.provider_merchant_id
  (provider, merchant_id, label, catatan, aktif, dibuat_pada, diubah_pada)
VALUES
  ($1, $2, $3, $4, $5, now(), now())
ON CONFLICT ((lower(trim(provider))), (lower(trim(merchant_id)))) DO UPDATE
SET label = EXCLUDED.label,
    catatan = EXCLUDED.catatan,
    aktif = EXCLUDED.aktif,
    diubah_pada = now()
RETURNING id
`, normalized, in.MerchantID, in.Label, in.Catatan, activeValue(in.Aktif)).Scan(&id)
	return id, err
}

func (r *ProviderMerchantIDRepository) Update(ctx context.Context, id int64, in ProviderMerchantIDUpsertInput) error {
	normalized, err := r.normalizeInput(ctx, &in)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE public.provider_merchant_id
SET provider = $2,
    merchant_id = $3,
    label = $4,
    catatan = $5,
    aktif = $6,
    diubah_pada = now()
WHERE id = $1
`, id, normalized, in.MerchantID, in.Label, in.Catatan, activeValue(in.Aktif))
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProviderMerchantIDRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.provider_merchant_id WHERE id = $1`, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProviderMerchantIDRepository) RandomActive(ctx context.Context, provider string) (string, bool, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return "", false, nil
	}
	var merchantID string
	err := r.db.QueryRowContext(ctx, `
SELECT merchant_id
FROM public.provider_merchant_id
WHERE lower(trim(provider)) = $1
  AND aktif = true
ORDER BY random(), id DESC
LIMIT 1
`, provider).Scan(&merchantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		if looksLikeMissingProviderMerchantIDTable(err) {
			return "", false, nil
		}
		return "", false, err
	}
	merchantID = strings.TrimSpace(merchantID)
	return merchantID, merchantID != "", nil
}

func (r *ProviderMerchantIDRepository) normalizeInput(ctx context.Context, in *ProviderMerchantIDUpsertInput) (string, error) {
	provider, err := resolveProviderName(ctx, r.db, in.Provider)
	if err != nil {
		return "", err
	}
	in.MerchantID = strings.TrimSpace(in.MerchantID)
	in.Label = strings.TrimSpace(in.Label)
	in.Catatan = strings.TrimSpace(in.Catatan)
	if in.MerchantID == "" {
		return "", errors.New("merchant_id required")
	}
	return provider, nil
}

func looksLikeMissingProviderMerchantIDTable(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "42P01" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "provider_merchant_id")
}
