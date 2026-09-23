package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    _ "github.com/lib/pq"
    "pulsa2/internal/helper"
    "pulsa2/internal/provider"
    "pulsa2/smb"
)

type target struct {
    TrxID       int64
    RefID       string
    Produk      string
    Tujuan      string
    Qty         int64
    BackupMapID int64
    BackupCode  string
}

func ptrString(s string) *string { s = strings.TrimSpace(s); if s == "" { return nil }; return &s }
func ptrI64(v int64) *int64 { return &v }

func main() {
    if len(os.Args) < 2 { log.Fatal("usage: resend_one_bifastopen2 <refid>") }
    refid := strings.TrimSpace(os.Args[1])
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" { log.Fatal("DATABASE_URL required") }
    db, err := sql.Open("postgres", dsn)
    if err != nil { log.Fatal(err) }
    defer db.Close()

    log.Printf("load target refid=%s", refid)
    var t target
    err = db.QueryRow(`SELECT tm.id, tm.ref_id, tm.kode_produk, tm.tujuan, tm.qty, ppm.id, ppm.kode_provider FROM transaksi_member tm JOIN produk p ON UPPER(TRIM(p.sku)) = UPPER(TRIM(tm.kode_produk)) JOIN produk_provider_map ppm ON ppm.produk_id = p.id JOIN provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider)) WHERE tm.ref_id = $1 AND LOWER(TRIM(tm.status)) = 'pending' AND LOWER(TRIM(ppm.provider)) = 'smb' AND pr.aktif = true AND ppm.aktif = true AND UPPER(TRIM(COALESCE(ppm.mode, ''))) = 'DIRECT' AND UPPER(TRIM(ppm.kode_provider)) LIKE 'BIFASTOPEN2:%' ORDER BY ppm.id DESC LIMIT 1`, refid).Scan(&t.TrxID, &t.RefID, &t.Produk, &t.Tujuan, &t.Qty, &t.BackupMapID, &t.BackupCode)
    if err != nil { log.Fatal(err) }

    log.Printf("target trx_id=%d product=%s tujuan=%s qty=%d map=%d backup=%s", t.TrxID, t.Produk, t.Tujuan, t.Qty, t.BackupMapID, t.BackupCode)
    var attempt int
    if err := db.QueryRow(`SELECT COALESCE(MAX(percobaan),0)+1 FROM transaksi_provider WHERE transaksi_member_id=$1 AND provider='smb' AND produk_provider_map_id=$2`, t.TrxID, t.BackupMapID).Scan(&attempt); err != nil { log.Fatal(err) }
    log.Printf("attempt=%d", attempt)

    reqMentah, _ := json.Marshal(map[string]any{"source":"manual_refid_resend_bifastopen2","refid":t.RefID,"product_in":t.Produk,"product_sent":t.BackupCode,"mode":"DIRECT"})
    var providerRowID int64
    log.Printf("insert provider row")
    err = db.QueryRow(`INSERT INTO transaksi_provider (provider, transaksi_member_id, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty, request_mentah, percobaan, status) VALUES ('smb',$1,$2,'PAY',$3,$4,$5,$6,$7,$8::jsonb,$9,'pending') RETURNING id`, t.TrxID, t.RefID, t.Produk, t.BackupMapID, t.BackupCode, t.Tujuan, t.Qty, string(reqMentah), attempt).Scan(&providerRowID)
    if err != nil { log.Fatal(err) }
    log.Printf("provider_row_id=%d", providerRowID)

    timeout := 30 * time.Second
    if raw := strings.TrimSpace(os.Getenv("SMB_TIMEOUT")); raw != "" { if d, derr := time.ParseDuration(raw); derr == nil && d > 0 { timeout = d } }
    adapter := &provider.SMBAdapter{C: smb.New(os.Getenv("SMB_BASE_URL"), os.Getenv("SMB_DIRECT_BASE_URL"), os.Getenv("SMB_ID"), os.Getenv("SMB_PIN"), os.Getenv("SMB_USER"), os.Getenv("SMB_PASSWORD"), timeout)}

    log.Printf("dispatch pay")
    ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
    defer cancel()
    resp, callErr := adapter.Pay(ctx, provider.PayRequest{Product:t.BackupCode, Mode:"DIRECT", Dest:t.Tujuan, Qty:t.Qty, RefID:t.RefID})
    if resp == nil {
        msg := "manual resend BIFASTOPEN2: respons SMB kosong"
        raw, _ := json.Marshal(map[string]any{"error":msg})
        _, _ = db.Exec(`UPDATE transaksi_provider SET http_status=$2, pesan=$3, respon_mentah=$4::jsonb, status='failed' WHERE id=$1`, providerRowID, 0, msg, string(raw))
        log.Fatalf("nil response")
    }
    status := "failed"
    msg := strings.TrimSpace(resp.Message)
    if resp.HTTPStatus == 200 { status = helper.ProviderResponseStatusString("smb", ptrString(resp.RC), ptrString(msg)) }
    rawJSON, _ := json.Marshal(resp.Raw)
    log.Printf("update result http=%d rc=%s status=%s msg=%s err=%v", resp.HTTPStatus, resp.RC, status, msg, callErr)
    _, err = db.Exec(`UPDATE transaksi_provider SET kode_respon=$2, pesan=$3, no_referensi=$4, harga=$5, saldo_terakhir=$6, http_status=$7, respon_mentah=$8::jsonb, status=$9 WHERE id=$1`, providerRowID, ptrString(resp.RC), ptrString(msg), ptrString(resp.ProviderRef), ptrI64(resp.Price), ptrI64(resp.Balance), resp.HTTPStatus, string(rawJSON), status)
    if err != nil { log.Fatal(err) }
    fmt.Printf("provider_row_id=%d refid=%s product_sent=%s http=%d status=%s rc=%s msg=%s err=%v\n", providerRowID, t.RefID, t.BackupCode, resp.HTTPStatus, status, resp.RC, msg, callErr)
}
