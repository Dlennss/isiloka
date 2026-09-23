package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/internal/helper/providersn"
)

type row struct {
	TrxID         int64
	ProviderTrxID int64
	Provider      string
	RefID         string
	Keterangan    string
	NoReferensi   string
	Pesan         string
}

func main() {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		log.Fatal("DATABASE_URL kosong")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
WITH latest_success_provider AS (
  SELECT DISTINCT ON (tp.transaksi_member_id)
    tp.id,
    tp.transaksi_member_id,
    LOWER(TRIM(tp.provider)) AS provider,
    COALESCE(tp.no_referensi, '') AS no_referensi,
    COALESCE(tp.pesan, '') AS pesan
  FROM public.transaksi_provider tp
  WHERE LOWER(COALESCE(tp.status, '')) = 'success'
  ORDER BY tp.transaksi_member_id, tp.id DESC
)
SELECT
  tm.id,
  lsp.id,
  lsp.provider,
  tm.ref_id,
  COALESCE(tm.keterangan, '') AS keterangan,
  lsp.no_referensi,
  lsp.pesan
FROM public.transaksi_member tm
JOIN latest_success_provider lsp ON lsp.transaksi_member_id = tm.id
WHERE tm.dibuat_pada >= now() - interval '7 days'
  AND LOWER(COALESCE(tm.status, '')) = 'success'
  AND LENGTH(BTRIM(COALESCE(tm.keterangan, ''))) <= 10
  AND lsp.provider IN ('gemilang', 'talentapay')
ORDER BY tm.id DESC
`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var repaired int
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.TrxID, &r.ProviderTrxID, &r.Provider, &r.RefID, &r.Keterangan, &r.NoReferensi, &r.Pesan); err != nil {
			log.Fatal(err)
		}
		best := strings.TrimSpace(repairRef(r.Provider, r.Pesan, r.NoReferensi))
		if best == "" || best == strings.TrimSpace(r.Keterangan) {
			continue
		}
		if _, err := db.ExecContext(ctx, `
UPDATE public.transaksi_member
SET keterangan = $2
WHERE id = $1
`, r.TrxID, best); err != nil {
			log.Fatalf("update transaksi_member refid=%s err=%v", r.RefID, err)
		}
		if _, err := db.ExecContext(ctx, `
UPDATE public.transaksi_provider
SET no_referensi = $2
WHERE id = $1
`, r.ProviderTrxID, best); err != nil {
			log.Fatalf("update transaksi_provider refid=%s err=%v", r.RefID, err)
		}
		repaired++
		fmt.Printf("repaired refid=%s provider=%s old=%q new=%q\n", r.RefID, r.Provider, r.Keterangan, best)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("total_repaired=%d\n", repaired)
}

func repairRef(provider, msg, noReferensi string) string {
	noReferensi = strings.TrimSpace(noReferensi)
	msg = strings.TrimSpace(msg)
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "gemilang":
		if pr, sn := providersn.ParseGemilangSNRefFromMsg(msg); strings.TrimSpace(sn) != "" {
			return strings.TrimSpace(sn)
		} else if strings.TrimSpace(pr) != "" {
			return strings.TrimSpace(pr)
		}
	case "talentapay":
		if pr, sn := providersn.ParseTalentaSNRefFromMsg(msg); strings.TrimSpace(sn) != "" {
			return strings.TrimSpace(sn)
		} else if strings.TrimSpace(pr) != "" {
			return strings.TrimSpace(pr)
		}
	}
	if isStrongRef(noReferensi) {
		return noReferensi
	}
	return ""
}

func isStrongRef(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	up := strings.ToUpper(v)
	switch up {
	case "NAMA", "OVO", "GOPAY", "DANA", "LINKAJA", "SHOPEE", "SERIAL", "TOPUP":
		return false
	}
	return len(v) >= 8
}
