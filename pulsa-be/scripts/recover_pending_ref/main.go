package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/config"
	"pulsa2/gemilang"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
	"pulsa2/javapay"
	"pulsa2/loketbayar"
	"pulsa2/minions"
	"pulsa2/multikom"
	"pulsa2/rajabiller"
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
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./scripts/recover_pending_ref <refid> [<refid>...]")
	}

	cfg := config.Load()
	dbConn := connectDB(cfg.DatabaseURL)
	defer dbConn.Close()

	repo := repository.NewMemberTrxRepository(dbConn)

	var lb *loketbayar.Client
	if cfg.LoketBayarBaseURL != "" && cfg.LoketBayarUsername != "" && cfg.LoketBayarPassword != "" {
		lb = loketbayar.New(
			cfg.LoketBayarBaseURL,
			cfg.LoketBayarUsername,
			cfg.LoketBayarPassword,
			cfg.LoketBayarTimeout,
		)
		lb.SetProductBaseURL("DANAPLUS", cfg.LoketBayarDanaBaseURL)
	}

	var ch *chytron.Client
	if cfg.ChytronBaseURL != "" && cfg.ChytronID != "" && cfg.ChytronPIN != "" && cfg.ChytronUser != "" && cfg.ChytronPassword != "" {
		ch = chytron.New(
			cfg.ChytronBaseURL,
			cfg.ChytronID,
			cfg.ChytronPIN,
			cfg.ChytronUser,
			cfg.ChytronPassword,
			cfg.ChytronTimeout,
		)
	}

	var rj *rajabiller.Client
	if cfg.RajabillerBaseURL != "" && cfg.RajabillerUID != "" && cfg.RajabillerPIN != "" {
		rj = rajabiller.New(
			cfg.RajabillerBaseURL,
			cfg.RajabillerUID,
			cfg.RajabillerPIN,
			cfg.RajabillerTimeout,
		)
		rj.MerchantID = cfg.RajabillerMerchantID
	}

	svc := service.NewMemberTrxService(
		repo,
		yuscom.New(cfg.YuscomBaseURL, cfg.YuscomMemberID, cfg.YuscomPIN, cfg.YuscomPassword, cfg.YuscomTimeout),
		javapay.New(cfg.JavaPayBaseURL, cfg.JavaPayMemberID, cfg.JavaPayAPIKey, cfg.JavaPayPIN, cfg.JavaPayTimeout),
		talenta.New(cfg.TalentaBaseURL, cfg.TalentaMemberID, cfg.TalentaPIN, cfg.TalentaPassword, cfg.TalentaTimeout),
		multikom.New(cfg.MultikomBaseURL, cfg.MultikomMemberID, cfg.MultikomPIN, cfg.MultikomPassword, cfg.MultikomTimeout),
		sagaramobile.New(cfg.SagaraBaseURL, cfg.SagaraMemberID, cfg.SagaraSign, cfg.SagaraOID, cfg.SagaraPIN, cfg.SagaraPassword, cfg.SagaraTimeout),
		minions.New(cfg.MinionsBaseURL, cfg.MinionsMemberID, cfg.MinionsPIN, cfg.MinionsPassword, cfg.MinionsTimeout),
		trionik.New(cfg.TrionikBaseURL, cfg.TrionikMemberID, cfg.TrionikPIN, cfg.TrionikPassword, cfg.TrionikTimeout),
		ajs.New(cfg.AJSBaseURL, cfg.AJSMemberID, cfg.AJSPIN, cfg.AJSPassword, cfg.AJSTimeout),
		gemilang.New(cfg.GemilangBaseURL, cfg.GemilangMemberID, cfg.GemilangPIN, cfg.GemilangPassword, cfg.GemilangTimeout),
		smb.New(cfg.SMBBaseURL, cfg.SMBDirectBaseURL, cfg.SMBID, cfg.SMBPIN, cfg.SMBUser, cfg.SMBPassword, cfg.SMBTimeout),
		lb,
		ch,
		rj,
	)

	for _, refID := range os.Args[1:] {
		var memberID int64
		if err := dbConn.QueryRowContext(context.Background(), `SELECT member_id FROM public.transaksi_member WHERE ref_id = $1`, refID).Scan(&memberID); err != nil {
			log.Printf("refid=%s lookup error: %v", refID, err)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		usedProvider, providerRowID, err := svc.RecoverPendingRef(ctx, memberID, refID)
		cancel()
		if err != nil {
			log.Printf("refid=%s recover error: %v", refID, err)
			continue
		}
		fmt.Printf("refid=%s fallback_provider=%s provider_row_id=%d\n", refID, usedProvider, providerRowID)
	}
}
