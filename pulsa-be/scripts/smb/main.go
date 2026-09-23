package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"pulsa2/config"
	"pulsa2/smb"
)

func main() {
	modeFlag := flag.String("mode", "DIRECT", "SMB mode: ELEKTRIK, PPOB, WALLET_PPOB, DIRECT")
	codeFlag := flag.String("code", "", "provider product code, example ELDN or DANA")
	destFlag := flag.String("dest", "6282124307365", "destination/account number. Wajib gunakan nomor test SMB.")
	qtyFlag := flag.Int64("qty", 0, "qty/amount")
	refFlag := flag.String("refid", "", "reference id")
	cmdFlag := flag.String("command", "PAY", "PAY or INQ")
	timeoutFlag := flag.Duration("timeout", 20*time.Second, "request timeout")
	flag.Parse()

	cfg := config.Load()
	client := smb.New(cfg.SMBBaseURL, cfg.SMBDirectBaseURL, cfg.SMBID, cfg.SMBPIN, cfg.SMBUser, cfg.SMBPassword, *timeoutFlag)
	if err := client.Validate(); err != nil {
		log.Fatal(err)
	}
	res, err := client.Dispatch(context.Background(), smb.Mode(*modeFlag), smb.Request{
		KodeProduk: *codeFlag,
		Tujuan:     *destFlag,
		Qty:        *qtyFlag,
		RefID:      *refFlag,
	}, *cmdFlag)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", *res)
}
