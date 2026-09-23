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
	"pulsa2/internal/service"
	"pulsa2/javapay"
	"pulsa2/minions"
	"pulsa2/multikom"
	"pulsa2/sagaramobile"
	"pulsa2/smb"
	"pulsa2/talenta"
	"pulsa2/trionik"
	"pulsa2/yuscom"
)

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
	olderFlag := flag.Duration("older-than", 2*time.Minute, "minimum age for SMB check-only pending rows")
	limitFlag := flag.Int("limit", 100, "max stale SMB pending transactions to repair")
	flag.Parse()

	cfg := config.Load()
	dbConn := connectDB(cfg.DatabaseURL)
	defer dbConn.Close()

	repo := repository.NewMemberTrxRepository(dbConn)

	ys := yuscom.New(cfg.YuscomBaseURL, cfg.YuscomMemberID, cfg.YuscomPIN, cfg.YuscomPassword, cfg.YuscomTimeout)
	jp := javapay.New(cfg.JavaPayBaseURL, cfg.JavaPayMemberID, cfg.JavaPayAPIKey, cfg.JavaPayPIN, cfg.JavaPayTimeout)

	var tl *talenta.Client
	if cfg.TalentaBaseURL != "" && cfg.TalentaMemberID != "" && cfg.TalentaPIN != "" && cfg.TalentaPassword != "" {
		tl = talenta.New(cfg.TalentaBaseURL, cfg.TalentaMemberID, cfg.TalentaPIN, cfg.TalentaPassword, cfg.TalentaTimeout)
	}
	var mk *multikom.Client
	if cfg.MultikomBaseURL != "" && cfg.MultikomMemberID != "" && cfg.MultikomPIN != "" && cfg.MultikomPassword != "" {
		mk = multikom.New(cfg.MultikomBaseURL, cfg.MultikomMemberID, cfg.MultikomPIN, cfg.MultikomPassword, cfg.MultikomTimeout)
	}
	var sg *sagaramobile.Client
	if cfg.SagaraBaseURL != "" && cfg.SagaraMemberID != "" && ((cfg.SagaraPIN != "" && cfg.SagaraPassword != "") || cfg.SagaraSign != "" || cfg.SagaraOID != "") {
		sg = sagaramobile.New(cfg.SagaraBaseURL, cfg.SagaraMemberID, cfg.SagaraSign, cfg.SagaraOID, cfg.SagaraPIN, cfg.SagaraPassword, cfg.SagaraTimeout)
	}
	var mn *minions.Client
	if cfg.MinionsBaseURL != "" && cfg.MinionsMemberID != "" && cfg.MinionsPIN != "" && cfg.MinionsPassword != "" {
		mn = minions.New(cfg.MinionsBaseURL, cfg.MinionsMemberID, cfg.MinionsPIN, cfg.MinionsPassword, cfg.MinionsTimeout)
	}
	var tr *trionik.Client
	if cfg.TrionikBaseURL != "" && cfg.TrionikMemberID != "" && cfg.TrionikPIN != "" && cfg.TrionikPassword != "" {
		tr = trionik.New(cfg.TrionikBaseURL, cfg.TrionikMemberID, cfg.TrionikPIN, cfg.TrionikPassword, cfg.TrionikTimeout)
	}
	var sm *smb.Client
	if cfg.SMBBaseURL != "" && cfg.SMBDirectBaseURL != "" && cfg.SMBID != "" && cfg.SMBPIN != "" && cfg.SMBUser != "" && cfg.SMBPassword != "" {
		sm = smb.New(cfg.SMBBaseURL, cfg.SMBDirectBaseURL, cfg.SMBID, cfg.SMBPIN, cfg.SMBUser, cfg.SMBPassword, cfg.SMBTimeout)
	}

	svc := service.NewMemberTrxService(repo, ys, jp, tl, mk, sg, mn, tr, nil, nil, sm, nil, nil, nil)

	cutoff := time.Now().Add(-*olderFlag)
	rows, err := dbConn.QueryContext(context.Background(), `
SELECT tm.id, tm.ref_id
FROM public.transaksi_member tm
JOIN LATERAL (
  SELECT tp.id, tp.pesan, tp.dibuat_pada, tp.respon_mentah
  FROM public.transaksi_provider tp
  WHERE tp.transaksi_member_id = tm.id
    AND tp.provider = 'smb'
  ORDER BY tp.id DESC
  LIMIT 1
) tp ON true
WHERE tm.status = 'pending'
  AND (
    upper(coalesce(tp.pesan, '')) LIKE '%INQSUKSES%'
    OR (
      lower(coalesce(tp.respon_mentah->>'stage', '')) = 'check'
      AND upper(coalesce(tp.respon_mentah->>'message', tp.respon_mentah->>'msg', '')) LIKE '%INQSUKSES%'
    )
  )
  AND tp.dibuat_pada <= $1
ORDER BY tm.id ASC
LIMIT $2
`, cutoff, *limitFlag)
	if err != nil {
		log.Fatalf("query stale smb pending: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var trxID int64
		var refID string
		if err := rows.Scan(&trxID, &refID); err != nil {
			log.Fatalf("scan stale row: %v", err)
		}
		res, err := svc.RepairStaleSMBPendingByTrxID(context.Background(), trxID)
		if err != nil {
			log.Printf("repair trx_id=%d refid=%s err=%v", trxID, refID, err)
			continue
		}
		fmt.Printf("trx_id=%d refid=%s status=%s fallback_provider=%s provider_row_id=%d msg=%s\n",
			res.TrxID, res.RefID, res.Status, res.FallbackProvider, res.ProviderRowID, res.Message)
		count++
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("iterate stale rows: %v", err)
	}
	fmt.Printf("processed=%d\n", count)
}
