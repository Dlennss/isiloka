package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string

	// Legacy direct provider: YUSCOM. For the current H2H flow, Pulsa24Jam is
	// the primary upstream and legacy providers should remain optional.
	YuscomBaseURL  string
	YuscomMemberID string
	YuscomPIN      string
	YuscomPassword string
	YuscomTimeout  time.Duration

	// Legacy fallback provider: JAVAPAY
	JavaPayBaseURL  string
	JavaPayMemberID string
	JavaPayAPIKey   string
	JavaPayPIN      string
	JavaPayTimeout  time.Duration

	// Special provider: TALENTA (LinkAja)
	TalentaBaseURL  string
	TalentaMemberID string
	TalentaPIN      string
	TalentaPassword string
	TalentaTimeout  time.Duration

	// Otomax provider: MULTIKOM
	MultikomBaseURL  string
	MultikomMemberID string
	MultikomPIN      string
	MultikomPassword string
	MultikomTimeout  time.Duration

	// Direct provider: LOKET BAYAR
	LoketBayarBaseURL     string
	LoketBayarDanaBaseURL string
	LoketBayarUsername    string
	LoketBayarPassword    string
	LoketBayarTimeout     time.Duration

	// IP provider: SAGARA MOBILE
	SagaraBaseURL  string
	SagaraMemberID string
	SagaraSign     string
	SagaraOID      string
	SagaraPIN      string
	SagaraPassword string
	SagaraTimeout  time.Duration

	// IP provider: MINIONS
	MinionsBaseURL  string
	MinionsMemberID string
	MinionsPIN      string
	MinionsPassword string
	MinionsTimeout  time.Duration

	// Otomax provider: TRIONIK
	TrionikBaseURL  string
	TrionikMemberID string
	TrionikPIN      string
	TrionikPassword string
	TrionikTimeout  time.Duration

	// Otomax provider: AJS
	AJSBaseURL  string
	AJSMemberID string
	AJSPIN      string
	AJSPassword string
	AJSTimeout  time.Duration

	// Otomax provider: GEMILANG
	GemilangBaseURL  string
	GemilangMemberID string
	GemilangPIN      string
	GemilangPassword string
	GemilangTimeout  time.Duration

	// IRS9 provider: SMB
	SMBBaseURL       string
	SMBDirectBaseURL string
	SMBID            string
	SMBPIN           string
	SMBUser          string
	SMBPassword      string
	SMBTimeout       time.Duration

	// IRS/H2H provider: CHYTRON
	ChytronBaseURL  string
	ChytronID       string
	ChytronPIN      string
	ChytronUser     string
	ChytronPassword string
	ChytronTimeout  time.Duration

	// JSON provider: RAJABILLER
	RajabillerBaseURL    string
	RajabillerUID        string
	RajabillerPIN        string
	RajabillerMerchantID string
	RajabillerTimeout    time.Duration

	// Primary H2H provider: PULSA24JAM
	Pulsa24JamBaseURL       string
	Pulsa24JamMemberID      string
	Pulsa24JamAPIKey        string
	Pulsa24JamPIN           string
	Pulsa24JamPassword      string
	Pulsa24JamSecret        string
	Pulsa24JamCallbackToken string
	Pulsa24JamTimeout       time.Duration

	Port string
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env: %s", key)
	}
	return v
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvAny(def string, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return def
}

func Load() Config {
	// Untuk local development: load .env jika ada.
	// Jika file tidak ada (mis. production with real env vars), lanjut tanpa error.
	_ = godotenv.Load(".env")

	yto, err := time.ParseDuration(getEnv("YUSCOM_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid YUSCOM_TIMEOUT: %v", err)
	}
	jto, err := time.ParseDuration(getEnv("JAVAPAY_TIMEOUT", "15s"))
	if err != nil {
		log.Fatalf("invalid JAVAPAY_TIMEOUT: %v", err)
	}
	tto, err := time.ParseDuration(getEnv("TALENTA_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid TALENTA_TIMEOUT: %v", err)
	}
	mto, err := time.ParseDuration(getEnv("MULTIKOM_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid MULTIKOM_TIMEOUT: %v", err)
	}
	lto, err := time.ParseDuration(getEnv("LOKETBAYAR_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid LOKETBAYAR_TIMEOUT: %v", err)
	}
	sto, err := time.ParseDuration(getEnv("SAGARA_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid SAGARA_TIMEOUT: %v", err)
	}
	mnto, err := time.ParseDuration(getEnv("MINIONS_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid MINIONS_TIMEOUT: %v", err)
	}
	tro, err := time.ParseDuration(getEnv("TRIONIK_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid TRIONIK_TIMEOUT: %v", err)
	}
	ajsto, err := time.ParseDuration(getEnv("AJS_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid AJS_TIMEOUT: %v", err)
	}
	gto, err := time.ParseDuration(getEnv("GEMILANG_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid GEMILANG_TIMEOUT: %v", err)
	}
	smbto, err := time.ParseDuration(getEnv("SMB_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid SMB_TIMEOUT: %v", err)
	}
	chto, err := time.ParseDuration(getEnv("CHYTRON_TIMEOUT", "20s"))
	if err != nil {
		log.Fatalf("invalid CHYTRON_TIMEOUT: %v", err)
	}
	rjto, err := time.ParseDuration(getEnv("RAJABILLER_TIMEOUT", "45s"))
	if err != nil {
		log.Fatalf("invalid RAJABILLER_TIMEOUT: %v", err)
	}
	p24to, err := time.ParseDuration(getEnv("PULSA24JAM_TIMEOUT", "30s"))
	if err != nil {
		log.Fatalf("invalid PULSA24JAM_TIMEOUT: %v", err)
	}

	return Config{
		DatabaseURL: mustEnv("DATABASE_URL"),

		YuscomBaseURL:  getEnv("YUSCOM_BASE_URL", ""),
		YuscomMemberID: getEnv("YUSCOM_MEMBERID", ""),
		YuscomPIN:      getEnv("YUSCOM_PIN", ""),
		YuscomPassword: getEnv("YUSCOM_PASSWORD", ""),
		YuscomTimeout:  yto,

		JavaPayBaseURL:  getEnv("JAVAPAY_BASE_URL", ""),
		JavaPayMemberID: getEnv("JAVAPAY_MEMBERID", ""),
		JavaPayAPIKey:   getEnv("JAVAPAY_APIKEY", ""),
		JavaPayPIN:      getEnv("JAVAPAY_PIN", ""),
		JavaPayTimeout:  jto,

		TalentaBaseURL:  getEnv("TALENTA_BASE_URL", ""),
		TalentaMemberID: getEnv("TALENTA_MEMBERID", ""),
		TalentaPIN:      getEnv("TALENTA_PIN", ""),
		TalentaPassword: getEnv("TALENTA_PASSWORD", ""),
		TalentaTimeout:  tto,

		MultikomBaseURL:  getEnv("MULTIKOM_BASE_URL", ""),
		MultikomMemberID: getEnv("MULTIKOM_MEMBERID", ""),
		MultikomPIN:      getEnv("MULTIKOM_PIN", ""),
		MultikomPassword: getEnv("MULTIKOM_PASSWORD", ""),
		MultikomTimeout:  mto,

		LoketBayarBaseURL:     getEnv("LOKETBAYAR_BASE_URL", ""),
		LoketBayarDanaBaseURL: getEnv("LOKETBAYAR_DANA_BASE_URL", ""),
		LoketBayarUsername:    getEnv("LOKETBAYAR_USERNAME", ""),
		LoketBayarPassword:    getEnv("LOKETBAYAR_PASSWORD", ""),
		LoketBayarTimeout:     lto,

		SagaraBaseURL:  getEnv("SAGARA_BASE_URL", ""),
		SagaraMemberID: getEnvAny("", "SAGARA_MEMBERID", "SAGARA_MEMBER_ID"),
		SagaraSign:     getEnv("SAGARA_SIGN", ""),
		SagaraOID:      getEnv("SAGARA_OID", ""),
		SagaraPIN:      getEnv("SAGARA_PIN", ""),
		SagaraPassword: getEnv("SAGARA_PASSWORD", ""),
		SagaraTimeout:  sto,

		MinionsBaseURL:  getEnv("MINIONS_BASE_URL", ""),
		MinionsMemberID: getEnvAny("", "MINIONS_MEMBERID", "MINIONS_MEMBER_ID"),
		MinionsPIN:      getEnv("MINIONS_PIN", ""),
		MinionsPassword: getEnv("MINIONS_PASSWORD", ""),
		MinionsTimeout:  mnto,

		TrionikBaseURL:  getEnv("TRIONIK_BASE_URL", ""),
		TrionikMemberID: getEnvAny("", "TRIONIK_MEMBERID", "TRIONIK_MEMBER_ID"),
		TrionikPIN:      getEnv("TRIONIK_PIN", ""),
		TrionikPassword: getEnv("TRIONIK_PASSWORD", ""),
		TrionikTimeout:  tro,

		AJSBaseURL:  getEnv("AJS_BASE_URL", ""),
		AJSMemberID: getEnvAny("", "AJS_MEMBERID", "AJS_MEMBER_ID"),
		AJSPIN:      getEnv("AJS_PIN", ""),
		AJSPassword: getEnv("AJS_PASSWORD", ""),
		AJSTimeout:  ajsto,

		GemilangBaseURL:  getEnv("GEMILANG_BASE_URL", ""),
		GemilangMemberID: getEnvAny("", "GEMILANG_MEMBERID", "GEMILANG_MEMBER_ID"),
		GemilangPIN:      getEnv("GEMILANG_PIN", ""),
		GemilangPassword: getEnv("GEMILANG_PASSWORD", ""),
		GemilangTimeout:  gto,

		SMBBaseURL:       getEnv("SMB_BASE_URL", ""),
		SMBDirectBaseURL: getEnv("SMB_DIRECT_BASE_URL", ""),
		SMBID:            getEnvAny("", "SMB_ID", "SMB_MEMBERID", "SMB_MEMBER_ID"),
		SMBPIN:           getEnv("SMB_PIN", ""),
		SMBUser:          getEnvAny("", "SMB_USER", "SMB_USERNAME", "SMB_ID", "SMB_MEMBERID"),
		SMBPassword:      getEnvAny("", "SMB_PASSWORD", "SMB_PASS"),
		SMBTimeout:       smbto,

		ChytronBaseURL:  getEnv("CHYTRON_BASE_URL", ""),
		ChytronID:       getEnvAny("", "CHYTRON_ID", "CHYTRON_IDRS", "CHYTRON_MEMBERID", "CHYTRON_MEMBER_ID"),
		ChytronPIN:      getEnv("CHYTRON_PIN", ""),
		ChytronUser:     getEnvAny("", "CHYTRON_USER", "CHYTRON_USERNAME"),
		ChytronPassword: getEnvAny("", "CHYTRON_PASSWORD", "CHYTRON_PASS"),
		ChytronTimeout:  chto,

		RajabillerBaseURL:    getEnv("RAJABILLER_BASE_URL", ""),
		RajabillerUID:        getEnv("RAJABILLER_UID", ""),
		RajabillerPIN:        getEnv("RAJABILLER_PIN", ""),
		RajabillerMerchantID: getEnv("RAJABILLER_MERCHANT_ID", ""),
		RajabillerTimeout:    rjto,

		Pulsa24JamBaseURL:       getEnv("PULSA24JAM_BASE_URL", ""),
		Pulsa24JamMemberID:      getEnvAny("", "PULSA24JAM_MEMBERID", "PULSA24JAM_MEMBER_ID"),
		Pulsa24JamAPIKey:        getEnv("PULSA24JAM_API_KEY", ""),
		Pulsa24JamPIN:           getEnv("PULSA24JAM_PIN", ""),
		Pulsa24JamPassword:      getEnv("PULSA24JAM_PASSWORD", ""),
		Pulsa24JamSecret:        getEnv("PULSA24JAM_SECRET", ""),
		Pulsa24JamCallbackToken: getEnv("PULSA24JAM_CALLBACK_TOKEN", ""),
		Pulsa24JamTimeout:       p24to,

		Port: getEnv("PORT", "8080"),
	}
}
