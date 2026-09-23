package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/config"
)

func main() {
	log.SetFlags(0)

	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlBytes, err := os.ReadFile("sql/20260406_add_transaksi_provider_status.sql")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		log.Fatal(err)
	}

	log.Println("applied transaksi_provider.status migration and backfill")
}
