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

type targetRow struct {
    ID    int64
    RefID string
}

func main() {
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

    const q = `
WITH candidates AS (
    SELECT tm.id, tm.ref_id
    FROM transaksi_member tm
    JOIN LATERAL (
        SELECT tp.provider, tp.status, tp.kode_produk, tp.http_status, tp.kode_respon
        FROM transaksi_provider tp
        WHERE tp.transaksi_member_id = tm.id
        ORDER BY tp.id DESC
        LIMIT 1
    ) lastp ON TRUE
    WHERE tm.status = $1
      AND tm.dibuat_pada < now() - make_interval(mins => 5)
      AND lastp.provider = $2
      AND lastp.status = $1
      AND lastp.kode_produk LIKE $3
      AND COALESCE(lastp.http_status, 0) = 200
      AND COALESCE(lastp.kode_respon::text, $4) = $5
)
SELECT id, ref_id FROM candidates ORDER BY id;`

    rows, err := db.QueryContext(ctx, q, "pending", "smb", "BIFASTOPEN2:%", "", "68")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    var targets []targetRow
    for rows.Next() {
        var t targetRow
        if err := rows.Scan(&t.ID, &t.RefID); err != nil {
            log.Fatal(err)
        }
        targets = append(targets, t)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("targets=%d\n", len(targets))
    if len(targets) == 0 {
        return
    }

    pcbRepo := repository.NewProviderCallbackRepository(db)
    histRepo := repository.NewHistoryRepository(db)
    histSvc := service.NewHistoryService(histRepo, nil)

    for _, t := range targets {
        trx, err := pcbRepo.GetTransaksiMemberByID(ctx, t.ID)
        if err != nil {
            fmt.Printf("skip refid=%s id=%d get_err=%v\n", t.RefID, t.ID, err)
            continue
        }
        note := "failed cleanup after stale BIFASTOPEN2 pending rc68"
        if err := pcbRepo.UpdateTransaksiMemberSettle(ctx, t.ID, "failed", note, 0, 0, trx.HargaMember); err != nil {
            fmt.Printf("fail refid=%s id=%d settle_err=%v\n", t.RefID, t.ID, err)
            continue
        }
        item, err := histSvc.AdminSendTransaksiCallback(ctx, 1, t.ID)
        if err != nil {
            fmt.Printf("partial refid=%s id=%d callback_err=%v\n", t.RefID, t.ID, err)
            continue
        }
        fmt.Printf("ok refid=%s id=%d callback_http=%d callback_body=%s\n", t.RefID, t.ID, item.CallbackHTTP, item.CallbackBody)
    }
}
