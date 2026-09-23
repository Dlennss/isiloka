package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"pulsa2/ajs"
	"pulsa2/gemilang"
	"pulsa2/javapay"
	"pulsa2/minions"
	"pulsa2/multikom"
	ridgen "pulsa2/refid"
	"pulsa2/sagaramobile"
	"pulsa2/smb"
	"pulsa2/talenta"
	"pulsa2/trionik"
	"pulsa2/yuscom"
)

type result struct {
	Provider    string
	Code        string
	RefID       string
	HTTPStatus  int
	ParsedRef   string
	ParsedProv  string
	ParsedPrice int64
	Body        string
	Err         error
}

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func main() {
	_ = godotenv.Load(".env")

	dest := flag.String("dest", "082124307365", "destination number")
	qty := flag.Int64("qty", 10000, "nominal/qty")
	timeout := flag.Duration("timeout", 25*time.Second, "request timeout")
	flag.Parse()

	tests := []struct {
		provider string
		code     string
		run      func(context.Context, string) result
	}{
		{
			provider: "ajs",
			code:     "DND",
			run: func(ctx context.Context, ref string) result {
				c := ajs.New(env("AJS_BASE_URL"), env("AJS_MEMBERID"), env("AJS_PIN"), env("AJS_PASSWORD"), *timeout)
				acc, hs, body, err := c.TrxNoSign(ctx, "DND", *qty, *dest, ref)
				return result{"ajs", "DND", ref, hs, acc.RefID, acc.Ticket, acc.Price, body, err}
			},
		},
		{
			provider: "gemilang",
			code:     "DNBS",
			run: func(ctx context.Context, ref string) result {
				c := gemilang.New(env("GEMILANG_BASE_URL"), env("GEMILANG_MEMBERID"), env("GEMILANG_PIN"), env("GEMILANG_PASSWORD"), *timeout)
				acc, hs, body, err := c.TrxNoSign(ctx, "DNBS", *qty, *dest, ref)
				return result{"gemilang", "DNBS", ref, hs, acc.RefID, acc.Ticket, acc.Price, body, err}
			},
		},
		{
			provider: "javapay",
			code:     "DNID",
			run: func(ctx context.Context, ref string) result {
				c := javapay.New(env("JAVAPAY_BASE_URL"), env("JAVAPAY_MEMBERID"), env("JAVAPAY_APIKEY"), env("JAVAPAY_PIN"), *timeout)
				resp, hs, _, err := c.Trx(ctx, "PAY", "DNID", *dest, *qty, ref)
				body := fmt.Sprintf("%v", resp)
				parsedRef, _ := resp["refid"].(string)
				parsedProv, _ := resp["noreff"].(string)
				return result{"javapay", "DNID", ref, hs, parsedRef, parsedProv, 0, body, err}
			},
		},
		{
			provider: "minions",
			code:     "HDANAB",
			run: func(ctx context.Context, ref string) result {
				c := minions.New(env("MINIONS_BASE_URL"), env("MINIONS_MEMBERID"), env("MINIONS_PIN"), env("MINIONS_PASSWORD"), *timeout)
				acc, hs, body, err := c.Trx(ctx, "HDANAB", *dest, *qty, ref)
				return result{"minions", "HDANAB", ref, hs, acc.RefID, acc.ProviderRef, acc.Price, body, err}
			},
		},
		{
			provider: "multikom",
			code:     "DANARP",
			run: func(ctx context.Context, ref string) result {
				c := multikom.New(env("MULTIKOM_BASE_URL"), env("MULTIKOM_MEMBERID"), env("MULTIKOM_PIN"), env("MULTIKOM_PASSWORD"), *timeout)
				acc, hs, body, err := c.TrxNoSign(ctx, "DANARP", *qty, *dest, ref)
				return result{"multikom", "DANARP", ref, hs, acc.RefID, acc.Ticket, acc.Price, body, err}
			},
		},
		{
			provider: "sagaramobile",
			code:     "DANARP",
			run: func(ctx context.Context, ref string) result {
				c := sagaramobile.New(env("SAGARA_BASE_URL"), env("SAGARA_MEMBER_ID"), env("SAGARA_SIGN"), env("SAGARA_OID"), env("SAGARA_PIN"), env("SAGARA_PASSWORD"), *timeout)
				acc, hs, body, err := c.Trx(ctx, "DANARP", *dest, *qty, ref)
				return result{"sagaramobile", "DANARP", ref, hs, acc.RefID, acc.ProviderRef, acc.Price, body, err}
			},
		},
		{
			provider: "smb",
			code:     "DANA",
			run: func(ctx context.Context, ref string) result {
				c := smb.New(env("SMB_BASE_URL"), env("SMB_DIRECT_BASE_URL"), env("SMB_ID"), env("SMB_PIN"), env("SMB_USER"), env("SMB_PASSWORD"), *timeout)
				mode, code, err := smb.ParseMappedCode("DANA")
				if err != nil {
					return result{"smb", "DANA", ref, 0, "", "", 0, "", err}
				}
				out, err := c.Dispatch(ctx, mode, smb.Request{KodeProduk: code, Tujuan: *dest, Qty: *qty, RefID: ref}, "PAY")
				if err != nil {
					return result{"smb", "DANA", ref, 0, "", "", 0, "", err}
				}
				body := out.Final.Body
				hs := out.Final.HTTPStatus
				if out.Pay != nil {
					hs = out.Pay.HTTPStatus
				}
				return result{"smb", "DANA", ref, hs, out.ProviderRef, out.ProviderRef, out.Price, body, nil}
			},
		},
		{
			provider: "talentapay",
			code:     "TDBN",
			run: func(ctx context.Context, ref string) result {
				c := talenta.New(env("TALENTA_BASE_URL"), env("TALENTA_MEMBERID"), env("TALENTA_PIN"), env("TALENTA_PASSWORD"), *timeout)
				acc, hs, body, err := c.TrxNoSign(ctx, "TDBN", *qty, *dest, ref)
				return result{"talentapay", "TDBN", ref, hs, acc.RefID, acc.Ticket, acc.Price, body, err}
			},
		},
		{
			provider: "trionik",
			code:     "DANA",
			run: func(ctx context.Context, ref string) result {
				c := trionik.New(env("TRIONIK_BASE_URL"), env("TRIONIK_MEMBERID"), env("TRIONIK_PIN"), env("TRIONIK_PASSWORD"), *timeout)
				acc, hs, body, err := c.TrxNoSign(ctx, "DANA", *qty, *dest, ref)
				return result{"trionik", "DANA", ref, hs, acc.RefID, acc.Ticket, acc.Price, body, err}
			},
		},
		{
			provider: "yuscom",
			code:     "DANA",
			run: func(ctx context.Context, ref string) result {
				c := yuscom.New(env("YUSCOM_BASE_URL"), env("YUSCOM_MEMBERID"), env("YUSCOM_PIN"), env("YUSCOM_PASSWORD"), *timeout)
				acc, hs, body, err := c.TrxNoSign(ctx, "DANA", *qty, *dest, ref)
				return result{"yuscom", "DANA", ref, hs, acc.RefID, acc.Ticket, acc.Price, body, err}
			},
		},
	}

	for _, tc := range tests {
		ref := ridgen.Generate(strings.ToUpper(tc.provider[:2]))
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		res := tc.run(ctx, ref)
		cancel()

		fmt.Printf("=== %s ===\n", strings.ToUpper(res.Provider))
		fmt.Printf("code       : %s\n", res.Code)
		fmt.Printf("refid      : %s\n", res.RefID)
		fmt.Printf("http       : %d\n", res.HTTPStatus)
		fmt.Printf("parsed_ref : %s\n", strings.TrimSpace(res.ParsedRef))
		fmt.Printf("provider_rf: %s\n", strings.TrimSpace(res.ParsedProv))
		fmt.Printf("price      : %d\n", res.ParsedPrice)
		if res.Err != nil {
			fmt.Printf("error      : %v\n", res.Err)
		}
		fmt.Printf("body       : %s\n\n", strings.TrimSpace(res.Body))
	}
}
