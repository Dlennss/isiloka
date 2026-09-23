package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/config"
)

const providerStatusExpr = `lower(trim(coalesce(tp.status, 'unknown')))`

type providerCount struct {
	Provider string
	Total    int64
}

type sampleRow struct {
	ID         int64
	Provider   string
	RefID      string
	KodeRespon string
	Pesan      string
}

func main() {
	log.SetFlags(0)

	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	loc := time.FixedZone("Asia/Jakarta", 7*60*60)
	now := time.Now().In(loc)
	cutoff := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	beforeCount, err := countPendingOld(ctx, db, cutoff)
	if err != nil {
		log.Fatal(err)
	}
	beforeByProvider, err := countPendingOldByProvider(ctx, db, cutoff)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("cutoff=%s\n", cutoff.Format(time.RFC3339))
	fmt.Printf("pending_old_before=%d\n", beforeCount)
	for _, row := range beforeByProvider {
		fmt.Printf("before %s=%d\n", row.Provider, row.Total)
	}

	res, err := db.ExecContext(ctx, `
UPDATE public.transaksi_provider tp
SET
  status = 'failed',
  kode_respon = COALESCE(NULLIF(TRIM(kode_respon),''), '91'),
  pesan = 'Provider stale sebelum hari ini; ditandai gagal cleanup admin'
WHERE tp.dibuat_pada < $1
  AND `+providerStatusExpr+` = 'pending'
`, cutoff)
	if err != nil {
		log.Fatal(err)
	}

	affected, _ := res.RowsAffected()
	fmt.Printf("affected=%d\n", affected)

	afterCount, err := countPendingOld(ctx, db, cutoff)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("pending_old_after=%d\n", afterCount)

	samples, err := sampleUpdated(ctx, db, cutoff)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range samples {
		fmt.Printf("sample id=%d provider=%s refid=%s rc=%s msg=%s\n", s.ID, s.Provider, s.RefID, s.KodeRespon, trimMsg(s.Pesan))
	}
}

func countPendingOld(ctx context.Context, db *sql.DB, cutoff time.Time) (int64, error) {
	var total int64
	err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM public.transaksi_provider tp WHERE tp.dibuat_pada < $1 AND `+providerStatusExpr+` = 'pending'`,
		cutoff,
	).Scan(&total)
	return total, err
}

func countPendingOldByProvider(ctx context.Context, db *sql.DB, cutoff time.Time) ([]providerCount, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT tp.provider, count(*) FROM public.transaksi_provider tp WHERE tp.dibuat_pada < $1 AND `+providerStatusExpr+` = 'pending' GROUP BY tp.provider ORDER BY count(*) DESC, tp.provider ASC`,
		cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]providerCount, 0, 16)
	for rows.Next() {
		var row providerCount
		if err := rows.Scan(&row.Provider, &row.Total); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func sampleUpdated(ctx context.Context, db *sql.DB, cutoff time.Time) ([]sampleRow, error) {
	rows, err := db.QueryContext(ctx, `
SELECT
  tp.id,
  tp.provider,
  tp.ref_id,
  COALESCE(tp.kode_respon, ''),
  COALESCE(tp.pesan, '')
FROM public.transaksi_provider tp
WHERE tp.dibuat_pada < $1
  AND tp.kode_respon = '91'
  AND UPPER(COALESCE(tp.pesan, '')) LIKE 'PROVIDER STALE SEBELUM HARI INI; DITANDAI GAGAL CLEANUP ADMIN%'
ORDER BY tp.dibuat_pada DESC, tp.id DESC
LIMIT 10
`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]sampleRow, 0, 10)
	for rows.Next() {
		var row sampleRow
		if err := rows.Scan(&row.ID, &row.Provider, &row.RefID, &row.KodeRespon, &row.Pesan); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func trimMsg(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 140 {
		return v[:140] + "..."
	}
	return v
}
