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
	"pulsa2/javapay"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	jpClient := javapay.New(
		os.Getenv("JAVAPAY_BASE_URL"),
		os.Getenv("JAVAPAY_MEMBER_ID"),
		os.Getenv("JAVAPAY_API_KEY"),
		os.Getenv("JAVAPAY_PIN"),
		15*time.Second,
	)

	rows, err := db.Query(`
SELECT tm.id, tm.ref_id, tm.kode_produk, tm.tujuan, tm.qty, tm.biaya_perkiraan, tm.member_id, tp.id
FROM transaksi_member tm
JOIN transaksi_provider tp ON tp.transaksi_member_id = tm.id
WHERE tp.provider = 'javapay' AND tm.status = 'pending'
  AND tm.dibuat_pada > '2026-04-10 17:00:00+07'
ORDER BY tm.dibuat_pada`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type stk struct {
		TrxID, ProvID, MemberID, Qty, Biaya int64
		RefID, Produk, Tujuan               string
	}
	var stuck []stk
	for rows.Next() {
		var s stk
		rows.Scan(&s.TrxID, &s.RefID, &s.Produk, &s.Tujuan, &s.Qty, &s.Biaya, &s.MemberID, &s.ProvID)
		stuck = append(stuck, s)
	}
	log.Printf("Found %d stuck JavaPay transactions", len(stuck))

	for _, s := range stuck {
		log.Printf("Checking refid=%s produk=%s tujuan=%s qty=%d", s.RefID, s.Produk, s.Tujuan, s.Qty)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		resp, httpCode, _, callErr := jpClient.Status(ctx, s.RefID)
		cancel()

		if callErr != nil {
			log.Printf("  ERROR: %v", callErr)
			continue
		}

		respJSON, _ := json.Marshal(resp)
		log.Printf("  RESPONSE http=%d body=%s", httpCode, truncStr(string(respJSON), 300))

		status := strings.ToLower(fmt.Sprintf("%v", resp["status"]))
		sn := fmt.Sprintf("%v", resp["sn"])
		price := int64(0)
		if p, ok := resp["price"]; ok {
			switch v := p.(type) {
			case float64:
				price = int64(v)
			case string:
				fmt.Sscanf(v, "%d", &price)
			}
		}

		switch {
		case status == "success" || status == "1" || status == "true":
			log.Printf("  SUCCESS sn=%s price=%d — updating", sn, price)
			db.Exec(`UPDATE transaksi_provider SET status='final', kode_respon='00', pesan=$2, no_referensi=$3, respon_mentah=$4 WHERE id=$1`,
				s.ProvID, "Transaksi berhasil", sn, string(respJSON))
			db.Exec(`UPDATE transaksi_member SET status='success', keterangan=$2, biaya_aktual=$3, harga_javapay=$4 WHERE id=$1 AND status='pending'`,
				s.TrxID, sn, s.Biaya, price)
			if price > 0 {
				debitWallet(db, s.RefID, price, s.TrxID, s.ProvID)
			}

		case status == "failed" || status == "2" || status == "false":
			log.Printf("  FAILED — refunding")
			res, _ := db.Exec(`UPDATE transaksi_member SET status='failed', keterangan='JavaPay gagal (status check)' WHERE id=$1 AND status='pending'`, s.TrxID)
			if aff, _ := res.RowsAffected(); aff > 0 {
				refundMember(db, s.MemberID, s.RefID, s.Biaya)
			}

		default:
			log.Printf("  STILL PENDING status=%s — skipping", status)
		}
	}
	log.Println("Done")
}

func refundMember(db *sql.DB, memberID int64, refID string, biaya int64) {
	tx, _ := db.Begin()
	defer tx.Rollback()
	var before int64
	tx.QueryRow(`SELECT saldo FROM dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	after := before + biaya
	tx.Exec(`UPDATE dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after)
	tx.Exec(`INSERT INTO mutasi_dompet (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah) VALUES ($1,$2,'CREDIT',$3,'REFUND','JavaPay status check (manual fix)',$4,$5)`,
		memberID, refID, biaya, before, after)
	tx.Commit()
	log.Printf("  REFUNDED member=%d amount=%d", memberID, biaya)
}

func debitWallet(db *sql.DB, refID string, price, trxID, provID int64) {
	tx, _ := db.Begin()
	defer tx.Rollback()
	var before int64
	tx.QueryRow(`SELECT saldo FROM dompet_provider WHERE provider='javapay' FOR UPDATE`).Scan(&before)
	after := before - price
	tx.Exec(`UPDATE dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider='javapay'`, after)
	tx.Exec(`INSERT INTO mutasi_dompet_provider (provider,ref_id,arah,jumlah,alasan,catatan,saldo_sebelum,saldo_sesudah,transaksi_member_id,transaksi_provider_id,dibuat_pada,meta) VALUES ('javapay',$1,'debit',$2,'TRX_SUCCESS_COST','status check fix',$3,$4,$5,$6,now(),'{}'::jsonb)`,
		refID, price, before, after, trxID, provID)
	tx.Commit()
}

func truncStr(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
