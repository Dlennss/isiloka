package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/config"
	"pulsa2/internal/repository"
)

type missingWalletRow struct {
	RefID               string
	TransaksiProviderID int64
	TransaksiMemberID   int64
	Harga               int64
	DibuatPada          time.Time
}

func connectDB(dsn string) *sql.DB {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	return conn
}

func main() {
	applyFlag := flag.Bool("apply", false, "apply missing SMB wallet debits")
	limitFlag := flag.Int("limit", 100, "max missing SMB rows to inspect")
	flag.Parse()

	cfg := config.Load()
	dbConn := connectDB(cfg.DatabaseURL)
	defer dbConn.Close()

	rows, err := dbConn.QueryContext(context.Background(), `
WITH smb_success AS (
  SELECT tp.ref_id, tp.id AS transaksi_provider_id, tp.transaksi_member_id, COALESCE(tp.harga, 0) AS harga, tp.dibuat_pada
  FROM public.transaksi_provider tp
  WHERE tp.provider = 'smb'
    AND COALESCE(tp.kode_respon, '') IN ('1', '20')
    AND COALESCE(tp.harga, 0) > 0
    AND UPPER(COALESCE(tp.pesan, '')) NOT LIKE '%INQSUKSES%'
),
missing AS (
  SELECT s.ref_id, s.transaksi_provider_id, s.transaksi_member_id, s.harga, s.dibuat_pada
  FROM smb_success s
  LEFT JOIN public.mutasi_dompet_provider m
    ON m.provider = 'smb'
   AND m.ref_id = s.ref_id
   AND m.arah = 'debit'
   AND m.alasan = 'TRX_SUCCESS_COST'
   AND COALESCE(m.transaksi_provider_id, 0) = s.transaksi_provider_id
  WHERE m.id IS NULL
)
SELECT ref_id, transaksi_provider_id, transaksi_member_id, harga, dibuat_pada
FROM missing
ORDER BY dibuat_pada ASC
LIMIT $1
`, *limitFlag)
	if err != nil {
		log.Fatalf("query missing smb wallet rows: %v", err)
	}
	defer rows.Close()

	repo := repository.NewProviderCallbackRepository(dbConn)
	total := 0
	totalHarga := int64(0)
	for rows.Next() {
		var row missingWalletRow
		if err := rows.Scan(&row.RefID, &row.TransaksiProviderID, &row.TransaksiMemberID, &row.Harga, &row.DibuatPada); err != nil {
			log.Fatalf("scan missing row: %v", err)
		}
		total++
		totalHarga += row.Harga
		if !*applyFlag {
			fmt.Printf("dry-run refid=%s trx_provider_id=%d trx_member_id=%d harga=%d dibuat_pada=%s\n", row.RefID, row.TransaksiProviderID, row.TransaksiMemberID, row.Harga, row.DibuatPada.Format(time.RFC3339))
			continue
		}
		before, after, err := repo.ApplyProviderWalletTx(context.Background(), repository.CallbackProviderWalletTxIn{
			Provider:            "smb",
			RefID:               row.RefID,
			Arah:                "debit",
			Jumlah:              row.Harga,
			Alasan:              "TRX_SUCCESS_COST",
			Catatan:             "repair missing SMB wallet debit",
			TransaksiMemberID:   &row.TransaksiMemberID,
			TransaksiProviderID: &row.TransaksiProviderID,
		})
		if err != nil {
			log.Printf("repair refid=%s trx_provider_id=%d err=%v", row.RefID, row.TransaksiProviderID, err)
			continue
		}
		fmt.Printf("applied refid=%s trx_provider_id=%d harga=%d saldo=%d->%d\n", row.RefID, row.TransaksiProviderID, row.Harga, before, after)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("iterate missing rows: %v", err)
	}
	fmt.Printf("count=%d total_harga=%d apply=%t\n", total, totalHarga, *applyFlag)
}
