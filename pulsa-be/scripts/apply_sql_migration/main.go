package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./scripts/apply_sql_migration <migration.sql>")
	}

	_ = godotenv.Load(".env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is empty")
	}

	path := filepath.Clean(os.Args[1])
	query, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read migration %s: %v", path, err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	if _, err := db.ExecContext(ctx, string(query)); err != nil {
		log.Fatalf("apply migration %s: %v", path, err)
	}

	fmt.Printf("Applied migration: %s\n", path)
}
