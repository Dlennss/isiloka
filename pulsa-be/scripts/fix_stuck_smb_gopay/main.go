package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"pulsa2/smb"
)

type stuckTrx struct {
	TrxID       int64
	ProviderID  int64
	RefID       string
	KodeProduk  string
	Tujuan      string
	Qty         int64
	QtyProvider int64
	MinutesAgo  float64
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	smbBaseURL := os.Getenv("SMB_BASE_URL")
	smbDirectURL := os.Getenv("SMB_DIRECT_BASE_URL")
	smbID := os.Getenv("SMB_ID")
	smbPIN := os.Getenv("SMB_PIN")
	smbUser := os.Getenv("SMB_USER")
	smbPass := os.Getenv("SMB_PASSWORD")
	if smbBaseURL == "" || smbID == "" {
		log.Fatal("SMB env vars required")
	}
	smbClient := smb.New(smbBaseURL, smbDirectURL, smbID, smbPIN, smbUser, smbPass, 30*time.Second)
	rows, err := db.Query(`
SELECT tm.id, tp.id, tm.ref_id, tm.kode_produk, tm.tujuan, tm.qty,
       COALESCE(tm.qty_provider, tm.qty),
       extract(epoch from now() - tm.dibuat_pada)/60
FROM transaksi_member tm
JOIN transaksi_provider tp ON tp.transaksi_member_id = tm.id
WHERE tp.provider = 'smb' AND tp.pesan = 'CEK' AND tm.status = 'pending'
ORDER BY tm.dibuat_pada`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var stuck []stuckTrx
	for rows.Next() {
		var s stuckTrx
		rows.Scan(&s.TrxID, &s.ProviderID, &s.RefID, &s.KodeProduk, &s.Tujuan, &s.Qty, &s.QtyProvider, &s.MinutesAgo)
		stuck = append(stuck, s)
	}
	log.Printf("Found %d stuck SMB GOPAY transactions", len(stuck))

	for _, s := range stuck {
		log.Printf("Processing refid=%s tujuan=%s qty=%d age=%.0fmin", s.RefID, s.Tujuan, s.QtyProvider, s.MinutesAgo)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		mode, code, parseErr := smb.ParseMappedCode(s.KodeProduk)
		if parseErr != nil {
			log.Printf("  SKIP parse error: %v", parseErr)
			cancel()
			continue
		}

		dispatch, callErr := smbClient.DispatchPayOnly(ctx, mode, smb.Request{
			KodeProduk: code, Tujuan: s.Tujuan, Qty: s.QtyProvider, RefID: s.RefID,
		})
		cancel()

		if callErr != nil || dispatch == nil {
			errMsg := "nil dispatch"
			if callErr != nil {
				errMsg = callErr.Error()
			}
			log.Printf("  DISPATCH FAILED: %s — refunding", errMsg)
			refundTrx(db, s)
			continue
		}

		body := strings.TrimSpace(dispatch.Final.Body)
		log.Printf("  DISPATCH: status=%s body=%s", dispatch.StatusCode, truncStr(body, 200))

		respJSON, _ := json.Marshal(map[string]any{"stage": "pay_manual_fix", "dispatch": dispatch, "message": body})

		if smb.LooksLikeSuccess(body) {
			price := dispatch.Price
			provRef := dispatch.ProviderRef
			log.Printf("  SUCCESS price=%d ref=%s", price, provRef)
			db.Exec(`UPDATE transaksi_provider SET pesan=$2, kode_respon=$3, no_referensi=$4, respon_mentah=$5 WHERE id=$1`,
				s.ProviderID, body, dispatch.StatusCode, provRef, string(respJSON))
			db.Exec(`UPDATE transaksi_member SET status='success', keterangan=$2, biaya_aktual=$3, harga_javapay=$4 WHERE id=$1 AND status='pending'`,
				s.TrxID, provRef, s.Qty, price)
			debitWallet(db, s.RefID, price, s.TrxID, s.ProviderID)
		} else if smb.LooksLikeImmediateReject(body) {
			log.Printf("  REJECTED — refunding")
			db.Exec(`UPDATE transaksi_provider SET pesan=$2, kode_respon=$3, respon_mentah=$4 WHERE id=$1`,
				s.ProviderID, body, dispatch.StatusCode, string(respJSON))
			refundTrx(db, s)
		} else {
			log.Printf("  STILL PENDING")
			db.Exec(`UPDATE transaksi_provider SET pesan=$2, kode_respon=$3, respon_mentah=$4 WHERE id=$1`,
				s.ProviderID, body, dispatch.StatusCode, string(respJSON))
		}
	}
	log.Println("All done")
}

func refundTrx(db *sql.DB, s stuckTrx) {
	res, _ := db.Exec(`UPDATE transaksi_member SET status='failed', keterangan='SMB check->pay gagal (auto fix)' WHERE id=$1 AND status='pending'`, s.TrxID)
	aff, _ := res.RowsAffected()
	if aff == 0 {
		log.Printf("  SKIP refund already changed trx=%d", s.TrxID)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		log.Printf("  ERROR begin tx: %v", err)
		return
	}
	defer tx.Rollback()
	var memberID, biaya, before int64
	tx.QueryRow(`SELECT member_id, biaya_perkiraan FROM transaksi_member WHERE id=$1`, s.TrxID).Scan(&memberID, &biaya)
	tx.QueryRow(`SELECT saldo FROM dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	after := before + biaya
	tx.Exec(`UPDATE dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after)
	tx.Exec(`INSERT INTO mutasi_dompet (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah) VALUES ($1,$2,'CREDIT',$3,'REFUND','SMB check->pay gagal (auto fix)',$4,$5)`,
		memberID, s.RefID, biaya, before, after)
	tx.Commit()
	log.Printf("  REFUNDED member=%d amount=%d", memberID, biaya)
}

func debitWallet(db *sql.DB, refID string, price, trxID, provID int64) {
	if price <= 0 {
		return
	}
	tx, _ := db.Begin()
	defer tx.Rollback()
	var before int64
	tx.QueryRow(`SELECT saldo FROM dompet_provider WHERE provider='smb' FOR UPDATE`).Scan(&before)
	after := before - price
	tx.Exec(`UPDATE dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider='smb'`, after)
	tx.Exec(`INSERT INTO mutasi_dompet_provider (provider,ref_id,arah,jumlah,alasan,catatan,saldo_sebelum,saldo_sesudah,transaksi_member_id,transaksi_provider_id,dibuat_pada,meta) VALUES ('smb',$1,'debit',$2,'TRX_SUCCESS_COST','manual fix smb stuck',$3,$4,$5,$6,now(),'{}'::jsonb)`,
		refID, price, before, after, trxID, provID)
	tx.Commit()
}

func truncStr(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
