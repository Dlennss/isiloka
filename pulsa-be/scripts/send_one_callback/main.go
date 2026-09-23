package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    "strconv"
    "time"

    _ "github.com/lib/pq"
    "pulsa2/internal/repository"
    "pulsa2/internal/service"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("usage: send_one_callback <trx_id>")
    }
    trxID, err := strconv.ParseInt(os.Args[1], 10, 64)
    if err != nil {
        log.Fatal(err)
    }
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL required")
    }
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    repo := repository.NewHistoryRepository(db)
    svc := service.NewHistoryService(repo, nil)
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()
    out, err := svc.AdminSendTransaksiCallback(ctx, 1, trxID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("trx_id=%d refid=%s status=%s http=%d webhook=%s body=%s\n", out.TrxID, out.RefID, out.Status, out.CallbackHTTP, out.WebhookURL, out.CallbackBody)
}
