package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ProviderWalletRepo menyimpan saldo provider versi internal (expected) + mutasi ledger.
// Provider yang didukung: yuscom, javapay.
type ProviderWalletRepo struct {
	DB *sql.DB
}

func NewProviderWalletRepo(db *sql.DB) *ProviderWalletRepo { return &ProviderWalletRepo{DB: db} }

func normalizeProvider(p string) (string, error) {
	p = strings.TrimSpace(strings.ToLower(p))
	switch p {
	case "yuscom", "javapay":
		return p, nil
	default:
		return "", fmt.Errorf("unknown provider: %s", p)
	}
}

// Ensure membuat row dompet_provider jika belum ada.
func (r *ProviderWalletRepo) Ensure(ctx context.Context, provider string) error {
	p, err := normalizeProvider(provider)
	if err != nil {
		return err
	}
	const q = `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING`
	_, err = r.DB.ExecContext(ctx, q, p)
	return err
}

// GetSaldo membaca saldo internal. Row akan di-ensure bila belum ada.
func (r *ProviderWalletRepo) GetSaldo(ctx context.Context, provider string) (int64, error) {
	p, err := normalizeProvider(provider)
	if err != nil {
		return 0, err
	}
	if err := r.Ensure(ctx, p); err != nil {
		return 0, err
	}
	const q = `SELECT saldo FROM public.dompet_provider WHERE provider = $1 LIMIT 1`
	var saldo int64
	if err := r.DB.QueryRowContext(ctx, q, p).Scan(&saldo); err != nil {
		return 0, err
	}
	return saldo, nil
}

type ProviderWalletTxIn struct {
	Provider            string
	RefID               string
	Arah                string // credit|debit
	Jumlah              int64
	AllowNegative       bool
	Alasan              string
	Catatan             string
	TransaksiMemberID   *int64
	TransaksiProviderID *int64
	MetaJSON            []byte // optional json (sudah marshalled)
}

// ApplyTx melakukan mutasi saldo internal secara atomic (FOR UPDATE) dan mencatat ledger.
func (r *ProviderWalletRepo) ApplyTx(ctx context.Context, in ProviderWalletTxIn) (before int64, after int64, err error) {
	p, err := normalizeProvider(in.Provider)
	if err != nil {
		return 0, 0, err
	}
	if strings.TrimSpace(in.RefID) == "" {
		return 0, 0, fmt.Errorf("ref_id required")
	}
	in.Arah = strings.TrimSpace(strings.ToLower(in.Arah))
	if in.Arah != "credit" && in.Arah != "debit" {
		return 0, 0, fmt.Errorf("arah must be credit|debit")
	}
	if in.Jumlah <= 0 {
		return 0, 0, fmt.Errorf("jumlah must be > 0")
	}
	if strings.TrimSpace(in.Alasan) == "" {
		return 0, 0, fmt.Errorf("alasan required")
	}

	if err := r.Ensure(ctx, p); err != nil {
		return 0, 0, err
	}

	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const qLock = `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`
	if err = tx.QueryRowContext(ctx, qLock, p).Scan(&before); err != nil {
		return 0, 0, err
	}

	after = before
	if in.Arah == "credit" {
		after = before + in.Jumlah
	} else {
		if before < in.Jumlah && !in.AllowNegative {
			return 0, 0, errors.New("provider saldo tidak cukup")
		}
		after = before - in.Jumlah
	}

	const qUpd = `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`
	if _, err = tx.ExecContext(ctx, qUpd, after, p); err != nil {
		return 0, 0, err
	}

	const qIns = `
INSERT INTO public.mutasi_dompet_provider
  (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah,
   transaksi_member_id, transaksi_provider_id, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,COALESCE($12,'{}'::jsonb))`

	var tm any = nil
	var tp any = nil
	if in.TransaksiMemberID != nil {
		tm = *in.TransaksiMemberID
	}
	if in.TransaksiProviderID != nil {
		tp = *in.TransaksiProviderID
	}

	createdAt := time.Now().UTC()
	var meta any = nil
	if len(in.MetaJSON) > 0 {
		meta = string(in.MetaJSON)
	}

	if _, err = tx.ExecContext(ctx, qIns,
		p, in.RefID, in.Arah, in.Jumlah, in.Alasan, in.Catatan, before, after, tm, tp, createdAt, meta,
	); err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}

type ProviderWalletMutasiRow struct {
	ID                  int64     `json:"id"`
	Provider            string    `json:"provider"`
	RefID               string    `json:"ref_id"`
	Arah                string    `json:"arah"`
	Jumlah              int64     `json:"jumlah"`
	Alasan              string    `json:"alasan"`
	Catatan             string    `json:"catatan"`
	SaldoSebelum        int64     `json:"saldo_sebelum"`
	SaldoSesudah        int64     `json:"saldo_sesudah"`
	TransaksiMemberID   *int64    `json:"transaksi_member_id,omitempty"`
	TransaksiProviderID *int64    `json:"transaksi_provider_id,omitempty"`
	DibuatPada          time.Time `json:"dibuat_pada"`
}

func (r *ProviderWalletRepo) ListMutasi(ctx context.Context, provider, arah string, limit, offset int) ([]ProviderWalletMutasiRow, int64, error) {
	if r == nil || r.DB == nil {
		return nil, 0, errors.New("db not initialized")
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	p := strings.TrimSpace(strings.ToLower(provider))
	if p != "" {
		np, err := normalizeProvider(p)
		if err != nil {
			return nil, 0, err
		}
		p = np
	}
	a := strings.TrimSpace(strings.ToLower(arah))
	if a != "" && a != "credit" && a != "debit" {
		return nil, 0, fmt.Errorf("arah must be credit|debit")
	}

	const qList = `
SELECT
  id, provider, ref_id, arah, jumlah, alasan, coalesce(catatan,''), saldo_sebelum, saldo_sesudah,
  transaksi_member_id, transaksi_provider_id, dibuat_pada
FROM public.mutasi_dompet_provider
WHERE ($1 = '' OR provider = $1)
  AND ($2 = '' OR arah = $2)
ORDER BY dibuat_pada DESC, id DESC
LIMIT $3 OFFSET $4`

	rows, err := r.DB.QueryContext(ctx, qList, p, a, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderWalletMutasiRow, 0, limit)
	for rows.Next() {
		var (
			x  ProviderWalletMutasiRow
			tm sql.NullInt64
			tp sql.NullInt64
		)
		if err := rows.Scan(
			&x.ID,
			&x.Provider,
			&x.RefID,
			&x.Arah,
			&x.Jumlah,
			&x.Alasan,
			&x.Catatan,
			&x.SaldoSebelum,
			&x.SaldoSesudah,
			&tm,
			&tp,
			&x.DibuatPada,
		); err != nil {
			return nil, 0, err
		}
		if tm.Valid {
			v := tm.Int64
			x.TransaksiMemberID = &v
		}
		if tp.Valid {
			v := tp.Int64
			x.TransaksiProviderID = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	const qCount = `
SELECT count(*)
FROM public.mutasi_dompet_provider
WHERE ($1 = '' OR provider = $1)
  AND ($2 = '' OR arah = $2)`
	var total int64
	if err := r.DB.QueryRowContext(ctx, qCount, p, a).Scan(&total); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}
