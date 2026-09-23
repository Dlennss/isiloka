package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

type ProviderRekeningRepository struct {
	db *sql.DB
}

func NewProviderRekeningRepository(db *sql.DB) *ProviderRekeningRepository {
	return &ProviderRekeningRepository{db: db}
}

func normalizeRekeningDigits(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (r *ProviderRekeningRepository) List(ctx context.Context, provider, q string, aktifOnly bool, limit, offset int) ([]ProviderRekeningRow, int64, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	q = strings.TrimSpace(q)
	qDigits := normalizeRekeningDigits(q)
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
		parts := []string{
			"lower(provider) LIKE " + textParam,
			"lower(nama) LIKE " + textParam,
			"lower(bank) LIKE " + textParam,
			"lower(nomor_rekening) LIKE " + textParam,
			"lower(catatan) LIKE " + textParam,
		}
		if qDigits != "" {
			args = append(args, "%"+qDigits+"%")
			parts = append(parts, "nomor_rekening_digits LIKE $"+strconv.Itoa(len(args)))
		}
		where = append(where, "("+strings.Join(parts, " OR ")+")")
	}
	if aktifOnly {
		where = append(where, "aktif = true")
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	countArgs := append([]any{}, args...)
	if err := r.db.QueryRowContext(ctx, `SELECT count(*)::bigint FROM public.provider_rekening WHERE `+whereSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
SELECT id, provider, nama, bank, nomor_rekening, nomor_rekening_digits, catatan, aktif, dibuat_pada, diubah_pada
FROM public.provider_rekening
WHERE `+whereSQL+`
ORDER BY lower(provider), lower(bank), lower(nama), nomor_rekening_digits, id
LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderRekeningRow, 0, limit)
	for rows.Next() {
		var item ProviderRekeningRow
		if err := rows.Scan(&item.ID, &item.Provider, &item.Nama, &item.Bank, &item.NomorRekening, &item.NomorRekeningDigits, &item.Catatan, &item.Aktif, &item.DibuatPada, &item.DiubahPada); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *ProviderRekeningRepository) Get(ctx context.Context, id int64) (*ProviderRekeningRow, error) {
	var item ProviderRekeningRow
	err := r.db.QueryRowContext(ctx, `
SELECT id, provider, nama, bank, nomor_rekening, nomor_rekening_digits, catatan, aktif, dibuat_pada, diubah_pada
FROM public.provider_rekening
WHERE id = $1
LIMIT 1
`, id).Scan(&item.ID, &item.Provider, &item.Nama, &item.Bank, &item.NomorRekening, &item.NomorRekeningDigits, &item.Catatan, &item.Aktif, &item.DibuatPada, &item.DiubahPada)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ProviderRekeningRepository) Create(ctx context.Context, in ProviderRekeningUpsertInput) (int64, error) {
	normalized, digits, err := r.normalizeInput(ctx, &in)
	if err != nil {
		return 0, err
	}
	var id int64
	err = r.db.QueryRowContext(ctx, `
INSERT INTO public.provider_rekening
  (provider, nama, bank, nomor_rekening, nomor_rekening_digits, catatan, aktif, dibuat_pada, diubah_pada)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT ((lower(trim(provider))), nomor_rekening_digits) DO UPDATE
SET nama = EXCLUDED.nama,
    bank = EXCLUDED.bank,
    nomor_rekening = EXCLUDED.nomor_rekening,
    catatan = EXCLUDED.catatan,
    aktif = EXCLUDED.aktif,
    diubah_pada = now()
RETURNING id
`, normalized, in.Nama, in.Bank, in.NomorRekening, digits, in.Catatan, activeValue(in.Aktif)).Scan(&id)
	return id, err
}

func (r *ProviderRekeningRepository) Update(ctx context.Context, id int64, in ProviderRekeningUpsertInput) error {
	normalized, digits, err := r.normalizeInput(ctx, &in)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE public.provider_rekening
SET provider = $2,
    nama = $3,
    bank = $4,
    nomor_rekening = $5,
    nomor_rekening_digits = $6,
    catatan = $7,
    aktif = $8,
    diubah_pada = now()
WHERE id = $1
`, id, normalized, in.Nama, in.Bank, in.NomorRekening, digits, in.Catatan, activeValue(in.Aktif))
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

func (r *ProviderRekeningRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.provider_rekening WHERE id = $1`, id)
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

func (r *ProviderRekeningRepository) normalizeInput(ctx context.Context, in *ProviderRekeningUpsertInput) (string, string, error) {
	provider, err := resolveProviderName(ctx, r.db, in.Provider)
	if err != nil {
		return "", "", err
	}
	in.Nama = strings.TrimSpace(in.Nama)
	in.Bank = strings.TrimSpace(in.Bank)
	in.NomorRekening = strings.TrimSpace(in.NomorRekening)
	in.Catatan = strings.TrimSpace(in.Catatan)
	digits := normalizeRekeningDigits(in.NomorRekening)
	if in.Nama == "" {
		return "", "", errors.New("nama required")
	}
	if digits == "" {
		return "", "", errors.New("nomor rekening required")
	}
	return provider, digits, nil
}

func activeValue(v *bool) bool {
	if v == nil {
		return true
	}
	return *v
}
