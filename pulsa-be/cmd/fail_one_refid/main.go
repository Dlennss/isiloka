package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"

    "pulsa2/internal/repository"
    "pulsa2/internal/service"
)

func main() {
    refid := "smpay238426c306"
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL is empty")
    }
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx := context.Background()
    var trxID int64
    if err := db.QueryRowContext(ctx, "SELECT id FROM transaksi_member WHERE ref_id=$1", refid).Scan(&trxID); err != nil {
        log.Fatal(err)
    }

    pcbRepo := repository.NewProviderCallbackRepository(db)
    histRepo := repository.NewHistoryRepository(db)
    histSvc := service.NewHistoryService(histRepo, nil)

    trx, err := pcbRepo.GetTransaksiMemberByID(ctx, trxID)
    if err != nil {
        log.Fatal(err)
    }
    note := "failed cleanup after stale SMB no-response"
    if err := pcbRepo.UpdateTransaksiMemberSettle(ctx, trxID, "failed", note, 0, 0, trx.HargaMember); err != nil {
        log.Fatal(err)
    }
    item, err := histSvc.AdminSendTransaksiCallback(ctx, 1, trxID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("ok refid=%s id=%d callback_http=%d callback_body=%s\n", refid, trxID, item.CallbackHTTP, item.CallbackBody)
}
