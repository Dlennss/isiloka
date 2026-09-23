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
	"pulsa2/smb"
)

type summaryRow struct {
	Label string
	Value int64
}

type sampleRow struct {
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

func liveBalance(ctx context.Context, cfg config.Config, qty int64) int64 {
	client := smb.New(cfg.SMBBaseURL, cfg.SMBDirectBaseURL, cfg.SMBID, cfg.SMBPIN, cfg.SMBUser, cfg.SMBPassword, cfg.SMBTimeout)
	if err := client.Validate(); err != nil {
		log.Printf("skip live SMB balance check: %v", err)
		return 0
	}
	res, err := client.Dispatch(ctx, smb.ModeWalletPPOB, smb.Request{
		KodeProduk: "DANA",
		Tujuan:     "6282124307365",
		Qty:        qty,
		RefID:      fmt.Sprintf("AS%d", time.Now().Unix()%100000000),
	}, "INQ")
	if err != nil {
		log.Printf("skip live SMB balance check: %v", err)
		return 0
	}
	if res != nil && res.LastBalance != nil {
		return *res.LastBalance
	}
	return 0
}

func main() {
	sampleLimit := flag.Int("sample-limit", 10, "number of sample refs per category")
	liveQty := flag.Int64("live-qty", 1000, "qty for safe live INQ balance check")
	flag.Parse()

	cfg := config.Load()
	dbConn := connectDB(cfg.DatabaseURL)
	defer dbConn.Close()

	ctx := context.Background()

	rows, err := dbConn.QueryContext(ctx, `
WITH wallet AS (
  SELECT saldo
  FROM public.dompet_provider
  WHERE provider = 'smb'
),
latest_snapshot AS (
  SELECT saldo_provider
  FROM public.provider_saldo_snapshot
  WHERE provider = 'smb'
  ORDER BY dibuat_pada DESC
  LIMIT 1
),
pay_success AS (
  SELECT tp.id, tp.ref_id, tp.transaksi_provider_id, tp.transaksi_member_id, COALESCE(tp.harga, 0) AS harga
  FROM (
    SELECT id, ref_id, id AS transaksi_provider_id, transaksi_member_id, harga, pesan
    FROM public.transaksi_provider
    WHERE provider = 'smb'
      AND COALESCE(kode_respon, '') IN ('1', '20')
      AND COALESCE(harga, 0) > 0
      AND UPPER(COALESCE(pesan, '')) NOT LIKE '%INQSUKSES%'
  ) tp
),
pay_success_missing_wallet AS (
  SELECT ps.*
  FROM pay_success ps
  LEFT JOIN public.mutasi_dompet_provider m
    ON m.provider = 'smb'
   AND m.ref_id = ps.ref_id
   AND m.arah = 'debit'
   AND m.alasan = 'TRX_SUCCESS_COST'
   AND COALESCE(m.transaksi_provider_id, 0) = ps.transaksi_provider_id
  WHERE m.id IS NULL
),
check_only_success AS (
  SELECT id, ref_id, transaksi_member_id, COALESCE(harga, 0) AS harga
  FROM public.transaksi_provider
  WHERE provider = 'smb'
    AND COALESCE(kode_respon, '') IN ('1', '20')
    AND COALESCE(harga, 0) > 0
    AND UPPER(COALESCE(pesan, '')) LIKE '%INQSUKSES%'
)
SELECT 'wallet_internal', COALESCE((SELECT saldo FROM wallet), 0)
UNION ALL
SELECT 'snapshot_latest', COALESCE((SELECT saldo_provider FROM latest_snapshot), 0)
UNION ALL
SELECT 'pay_success_count', COALESCE((SELECT count(*) FROM pay_success), 0)
UNION ALL
SELECT 'pay_success_total_harga', COALESCE((SELECT sum(harga) FROM pay_success), 0)
UNION ALL
SELECT 'pay_success_missing_wallet_count', COALESCE((SELECT count(*) FROM pay_success_missing_wallet), 0)
UNION ALL
SELECT 'pay_success_missing_wallet_total', COALESCE((SELECT sum(harga) FROM pay_success_missing_wallet), 0)
UNION ALL
SELECT 'check_only_count', COALESCE((SELECT count(*) FROM check_only_success), 0)
UNION ALL
SELECT 'check_only_total_harga_view', COALESCE((SELECT sum(harga) FROM check_only_success), 0)
`)
	if err != nil {
		log.Fatalf("query summary: %v", err)
	}
	defer rows.Close()

	summary := map[string]int64{}
	for rows.Next() {
		var row summaryRow
		if err := rows.Scan(&row.Label, &row.Value); err != nil {
			log.Fatalf("scan summary: %v", err)
		}
		summary[row.Label] = row.Value
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("iterate summary: %v", err)
	}

	live := liveBalance(ctx, cfg, *liveQty)
	internal := summary["wallet_internal"]
	fmt.Printf("wallet_internal=%d\n", internal)
	fmt.Printf("snapshot_latest=%d\n", summary["snapshot_latest"])
	fmt.Printf("live_balance=%d\n", live)
	if live > 0 {
		fmt.Printf("gap_internal_vs_live=%d\n", internal-live)
	}
	fmt.Printf("pay_success_count=%d\n", summary["pay_success_count"])
	fmt.Printf("pay_success_total_harga=%d\n", summary["pay_success_total_harga"])
	fmt.Printf("pay_success_missing_wallet_count=%d\n", summary["pay_success_missing_wallet_count"])
	fmt.Printf("pay_success_missing_wallet_total=%d\n", summary["pay_success_missing_wallet_total"])
	fmt.Printf("check_only_count=%d\n", summary["check_only_count"])
	fmt.Printf("check_only_total_harga_view=%d\n", summary["check_only_total_harga_view"])

	samples, err := dbConn.QueryContext(ctx, `
WITH pay_success_missing_wallet AS (
  SELECT tp.ref_id, tp.id AS transaksi_provider_id, tp.transaksi_member_id, COALESCE(tp.harga, 0) AS harga, tp.dibuat_pada
  FROM public.transaksi_provider tp
  LEFT JOIN public.mutasi_dompet_provider m
    ON m.provider = 'smb'
   AND m.ref_id = tp.ref_id
   AND m.arah = 'debit'
   AND m.alasan = 'TRX_SUCCESS_COST'
   AND COALESCE(m.transaksi_provider_id, 0) = tp.id
  WHERE tp.provider = 'smb'
    AND COALESCE(tp.kode_respon, '') IN ('1', '20')
    AND COALESCE(tp.harga, 0) > 0
    AND UPPER(COALESCE(tp.pesan, '')) NOT LIKE '%INQSUKSES%'
    AND m.id IS NULL
  ORDER BY tp.dibuat_pada DESC
  LIMIT $1
)
SELECT ref_id, transaksi_provider_id, transaksi_member_id, harga, dibuat_pada
FROM pay_success_missing_wallet
`, *sampleLimit)
	if err != nil {
		log.Fatalf("query missing wallet samples: %v", err)
	}
	defer samples.Close()

	fmt.Println("missing_wallet_samples:")
	for samples.Next() {
		var row sampleRow
		if err := samples.Scan(&row.RefID, &row.TransaksiProviderID, &row.TransaksiMemberID, &row.Harga, &row.DibuatPada); err != nil {
			log.Fatalf("scan missing wallet sample: %v", err)
		}
		fmt.Printf("refid=%s trx_provider_id=%d trx_member_id=%d harga=%d dibuat_pada=%s\n",
			row.RefID, row.TransaksiProviderID, row.TransaksiMemberID, row.Harga, row.DibuatPada.Format(time.RFC3339))
	}
	if err := samples.Err(); err != nil {
		log.Fatalf("iterate missing wallet samples: %v", err)
	}
}
