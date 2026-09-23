package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/config"
	"pulsa2/gemilang"
	"pulsa2/httpapi"
	"pulsa2/internal/provider"
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

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	return conn
}

func main() {
	cfg := config.Load()

	dbConn := connectDB(cfg.DatabaseURL)
	defer dbConn.Close()

	var ys *yuscom.Client
	if cfg.YuscomBaseURL != "" && cfg.YuscomMemberID != "" && cfg.YuscomPIN != "" && cfg.YuscomPassword != "" {
		ys = yuscom.New(
			cfg.YuscomBaseURL,
			cfg.YuscomMemberID,
			cfg.YuscomPIN,
			cfg.YuscomPassword,
			cfg.YuscomTimeout,
		)
	} else {
		log.Printf("yuscom client nonaktif: env YUSCOM_* belum lengkap")
	}

	var jp *javapay.Client
	if cfg.JavaPayBaseURL != "" && cfg.JavaPayMemberID != "" && cfg.JavaPayAPIKey != "" && cfg.JavaPayPIN != "" {
		jp = javapay.New(
			cfg.JavaPayBaseURL,
			cfg.JavaPayMemberID,
			cfg.JavaPayAPIKey,
			cfg.JavaPayPIN,
			cfg.JavaPayTimeout,
		)
	} else {
		log.Printf("javapay client nonaktif: env JAVAPAY_* belum lengkap")
	}

	var tl *talenta.Client
	if cfg.TalentaBaseURL != "" && cfg.TalentaMemberID != "" && cfg.TalentaPIN != "" && cfg.TalentaPassword != "" {
		tl = talenta.New(
			cfg.TalentaBaseURL,
			cfg.TalentaMemberID,
			cfg.TalentaPIN,
			cfg.TalentaPassword,
			cfg.TalentaTimeout,
		)
	} else {
		log.Printf("talenta client nonaktif: env TALENTA_* belum lengkap")
	}

	var mk *multikom.Client
	if cfg.MultikomBaseURL != "" && cfg.MultikomMemberID != "" && cfg.MultikomPIN != "" && cfg.MultikomPassword != "" {
		mk = multikom.New(
			cfg.MultikomBaseURL,
			cfg.MultikomMemberID,
			cfg.MultikomPIN,
			cfg.MultikomPassword,
			cfg.MultikomTimeout,
		)
	} else {
		log.Printf("multikom client nonaktif: env MULTIKOM_* belum lengkap")
	}

	var lb *loketbayar.Client
	if cfg.LoketBayarBaseURL != "" && cfg.LoketBayarUsername != "" && cfg.LoketBayarPassword != "" {
		lb = loketbayar.New(
			cfg.LoketBayarBaseURL,
			cfg.LoketBayarUsername,
			cfg.LoketBayarPassword,
			cfg.LoketBayarTimeout,
		)
		lb.SetProductBaseURL("DANAPLUS", cfg.LoketBayarDanaBaseURL)
	} else {
		log.Printf("loketbayar client nonaktif: env LOKETBAYAR_* belum lengkap")
	}

	var sg *sagaramobile.Client
	if cfg.SagaraBaseURL != "" && cfg.SagaraMemberID != "" && ((cfg.SagaraPIN != "" && cfg.SagaraPassword != "") || cfg.SagaraSign != "" || cfg.SagaraOID != "") {
		sg = sagaramobile.New(
			cfg.SagaraBaseURL,
			cfg.SagaraMemberID,
			cfg.SagaraSign,
			cfg.SagaraOID,
			cfg.SagaraPIN,
			cfg.SagaraPassword,
			cfg.SagaraTimeout,
		)
	} else {
		log.Printf("sagaramobile client nonaktif: env SAGARA_* belum lengkap")
	}

	var mn *minions.Client
	if cfg.MinionsBaseURL != "" && cfg.MinionsMemberID != "" && cfg.MinionsPIN != "" && cfg.MinionsPassword != "" {
		mn = minions.New(
			cfg.MinionsBaseURL,
			cfg.MinionsMemberID,
			cfg.MinionsPIN,
			cfg.MinionsPassword,
			cfg.MinionsTimeout,
		)
	} else {
		log.Printf("minions client nonaktif: env MINIONS_* belum lengkap")
	}

	var tr *trionik.Client
	if cfg.TrionikBaseURL != "" && cfg.TrionikMemberID != "" && cfg.TrionikPIN != "" && cfg.TrionikPassword != "" {
		tr = trionik.New(
			cfg.TrionikBaseURL,
			cfg.TrionikMemberID,
			cfg.TrionikPIN,
			cfg.TrionikPassword,
			cfg.TrionikTimeout,
		)
	} else {
		log.Printf("trionik client nonaktif: env TRIONIK_* belum lengkap")
	}

	var aj *ajs.Client
	if cfg.AJSBaseURL != "" && cfg.AJSMemberID != "" && cfg.AJSPIN != "" && cfg.AJSPassword != "" {
		aj = ajs.New(
			cfg.AJSBaseURL,
			cfg.AJSMemberID,
			cfg.AJSPIN,
			cfg.AJSPassword,
			cfg.AJSTimeout,
		)
	} else {
		log.Printf("ajs client nonaktif: env AJS_* belum lengkap")
	}

	var gm *gemilang.Client
	if cfg.GemilangBaseURL != "" && cfg.GemilangMemberID != "" && cfg.GemilangPIN != "" && cfg.GemilangPassword != "" {
		gm = gemilang.New(
			cfg.GemilangBaseURL,
			cfg.GemilangMemberID,
			cfg.GemilangPIN,
			cfg.GemilangPassword,
			cfg.GemilangTimeout,
		)
	} else {
		log.Printf("gemilang client nonaktif: env GEMILANG_* belum lengkap")
	}

	var smbc *smb.Client
	if cfg.SMBBaseURL != "" && cfg.SMBDirectBaseURL != "" && cfg.SMBID != "" && cfg.SMBPIN != "" && cfg.SMBUser != "" && cfg.SMBPassword != "" {
		smbc = smb.New(
			cfg.SMBBaseURL,
			cfg.SMBDirectBaseURL,
			cfg.SMBID,
			cfg.SMBPIN,
			cfg.SMBUser,
			cfg.SMBPassword,
			cfg.SMBTimeout,
		)
	} else {
		log.Printf("smb client nonaktif: env SMB_* belum lengkap")
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
	} else {
		log.Printf("chytron client nonaktif: env CHYTRON_* belum lengkap")
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
	} else {
		log.Printf("rajabiller client nonaktif: env RAJABILLER_* belum lengkap")
	}

	var p24 provider.Client
	if cfg.Pulsa24JamBaseURL != "" && cfg.Pulsa24JamAPIKey != "" && cfg.Pulsa24JamPIN != "" {
		p24 = provider.NewPulsa24JamAdapter(provider.Pulsa24JamConfig{
			BaseURL:       cfg.Pulsa24JamBaseURL,
			MemberID:      cfg.Pulsa24JamMemberID,
			APIKey:        cfg.Pulsa24JamAPIKey,
			PIN:           cfg.Pulsa24JamPIN,
			Password:      cfg.Pulsa24JamPassword,
			Secret:        cfg.Pulsa24JamSecret,
			CallbackToken: cfg.Pulsa24JamCallbackToken,
			Timeout:       cfg.Pulsa24JamTimeout,
		})
	} else {
		log.Printf("pulsa24jam client nonaktif: env PULSA24JAM_* belum lengkap")
	}

	handler := httpapi.Routes(httpapi.Deps{
		DB:  dbConn,
		YS:  ys,
		JP:  jp,
		TL:  tl,
		MK:  mk,
		LB:  lb,
		SG:  sg,
		MN:  mn,
		TR:  tr,
		AJ:  aj,
		GM:  gm,
		SM:  smbc,
		CH:  ch,
		RJ:  rj,
		P24: p24,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("pulsa2 running on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
