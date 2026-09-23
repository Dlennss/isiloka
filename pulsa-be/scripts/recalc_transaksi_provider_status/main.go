package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"pulsa2/config"
)

// Script ini sudah tidak melakukan recalc dari RC/pesan.
// Kolom status adalah single source of truth.
// Script hanya melaporkan statistik status saat ini.

func main() {
	log.SetFlags(0)
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
SELECT COALESCE(LOWER(TRIM(status)), 'null'), count(*)
FROM public.transaksi_provider
GROUP BY 1
ORDER BY 2 DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== Status distribution ===")
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %s = %d\n", status, count)
	}
	fmt.Println("Kolom status adalah single source of truth. Tidak ada recalc dari RC/pesan.")
}
