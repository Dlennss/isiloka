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

	"pulsa2/internal/helper/providersn"
)

type weakRow struct {
	TrxID        int64
	RefID        string
	Status       string
	Keterangan   string
	BiayaAktual  int64
	Provider     string
	Pesan        string
	NoReferensi  string
	TrxIDJavapay string
	ResponMentah []byte
}

type updateInput struct {
	TrxID              int64
	StatusSebelum      string
	StatusSesudah      string
	KeteranganSebelum  string
	KeteranganSesudah  string
	BiayaAktualSebelum int64
	BiayaAktualSesudah int64
	Aksi               string
}

func main() {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		log.Fatalf("db ping error: %v", err)
	}
	cancel()

	rows, err := loadWeakRows(context.Background(), db)
	if err != nil {
		log.Fatalf("load rows error: %v", err)
	}

	var updated, skipped int
	updatedByProvider := map[string]int{}
	skippedByProvider := map[string]int{}

	for _, row := range rows {
		strong := deriveStrongKeterangan(row)
		if strong == "" || isWeakValue(strong) {
			skipped++
			skippedByProvider[row.Provider]++
			continue
		}
		if strings.EqualFold(strings.TrimSpace(row.Keterangan), strong) {
			skipped++
			skippedByProvider[row.Provider]++
			continue
		}
		if err := applyUpdate(context.Background(), db, row, strong); err != nil {
			log.Printf("update failed refid=%s provider=%s err=%v", row.RefID, row.Provider, err)
			skipped++
			skippedByProvider[row.Provider]++
			continue
		}
		updated++
		updatedByProvider[row.Provider]++
	}

	fmt.Printf("checked=%d updated=%d skipped=%d\n", len(rows), updated, skipped)
	for _, provider := range []string{"yuscom", "talentapay", "javapay", "multikom"} {
		fmt.Printf("provider=%s updated=%d skipped=%d\n", provider, updatedByProvider[provider], skippedByProvider[provider])
	}
}

func loadWeakRows(ctx context.Context, db *sql.DB) ([]weakRow, error) {
	const q = `
SELECT DISTINCT ON (tm.id)
  tm.id,
  tm.ref_id,
  COALESCE(tm.status, ''),
  COALESCE(tm.keterangan, ''),
  COALESCE(tm.biaya_aktual, 0),
  COALESCE(tp.provider, ''),
  COALESCE(tp.pesan, ''),
  COALESCE(tp.no_referensi, ''),
  COALESCE(tp.trx_id_javapay, ''),
  tp.respon_mentah
FROM public.transaksi_member tm
JOIN public.transaksi_provider tp ON tp.ref_id = tm.ref_id
WHERE tm.status = 'success'
  AND (
    tm.keterangan IS NULL
    OR BTRIM(tm.keterangan) = ''
    OR UPPER(BTRIM(tm.keterangan)) IN ('NO', 'N/A', '-', 'TRANSAKSI BERHASIL')
    OR tm.keterangan ~* '^transaksi berhasil'
  )
ORDER BY tm.id, tp.id DESC
`
	rs, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var out []weakRow
	for rs.Next() {
		var row weakRow
		if err := rs.Scan(
			&row.TrxID,
			&row.RefID,
			&row.Status,
			&row.Keterangan,
			&row.BiayaAktual,
			&row.Provider,
			&row.Pesan,
			&row.NoReferensi,
			&row.TrxIDJavapay,
			&row.ResponMentah,
		); err != nil {
			return nil, err
		}
		row.Provider = strings.ToLower(strings.TrimSpace(row.Provider))
		out = append(out, row)
	}
	return out, rs.Err()
}

func deriveStrongKeterangan(row weakRow) string {
	msg := strings.TrimSpace(row.Pesan)
	noreff := strings.TrimSpace(row.NoReferensi)
	parsedJavaPay := parseJavaPayNoReff(row.ResponMentah)
	switch row.Provider {
	case "talentapay":
		pr, sn := providersn.ParseTalentaSNRefFromMsg(msg)
		if strong := pickStrong(sn, pr, noreff, parsedJavaPay); strong != "" {
			return strong
		}
	case "multikom":
		pr, sn := providersn.ParseMultikomSNRefFromMsg(msg)
		if strong := pickStrong(sn, pr, noreff, parsedJavaPay); strong != "" {
			return strong
		}
	case "yuscom":
		pr, sn := providersn.ParseYuscomSNRefFromMsg(msg)
		if strong := pickStrong(sn, pr, noreff, parsedJavaPay); strong != "" {
			return strong
		}
	case "javapay":
		if strong := pickStrong(noreff, row.TrxIDJavapay, parsedJavaPay); strong != "" {
			return strong
		}
	}
	return ""
}

func parseJavaPayNoReff(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	data, _ := payload["data"].(map[string]any)
	if v := strings.TrimSpace(toString(payload["noreff"])); v != "" {
		return v
	}
	if v := strings.TrimSpace(toString(payload["sn"])); v != "" {
		return v
	}
	if data != nil {
		if v := strings.TrimSpace(toString(data["noreff"])); v != "" {
			return v
		}
		if v := strings.TrimSpace(toString(data["sn"])); v != "" {
			return v
		}
	}
	return ""
}

func toString(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func pickStrong(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || isWeakValue(value) {
			continue
		}
		return value
	}
	return ""
}

func isWeakValue(v string) bool {
	v = strings.TrimSpace(strings.ToUpper(v))
	switch v {
	case "", "NO", "N/A", "-", "TRANSAKSI BERHASIL":
		return true
	default:
		return strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "transaksi berhasil")
	}
}

func applyUpdate(ctx context.Context, db *sql.DB, row weakRow, strong string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const updateQ = `
UPDATE public.transaksi_member
SET keterangan = NULLIF($2, '')
WHERE id = $1
`
	if _, err := tx.ExecContext(ctx, updateQ, row.TrxID, strong); err != nil {
		return err
	}

	in := updateInput{
		TrxID:              row.TrxID,
		StatusSebelum:      row.Status,
		StatusSesudah:      row.Status,
		KeteranganSebelum:  row.Keterangan,
		KeteranganSesudah:  strong,
		BiayaAktualSebelum: row.BiayaAktual,
		BiayaAktualSesudah: row.BiayaAktual,
		Aksi:               "bulk_fix_missing_sn",
	}
	if err := insertStatusLog(ctx, tx, in); err != nil {
		return err
	}

	return tx.Commit()
}

func insertStatusLog(ctx context.Context, exec interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, in updateInput) error {
	const q = `
INSERT INTO public.transaksi_member_status_log
  (transaksi_member_id, status_sebelum, status_sesudah, keterangan_sebelum, keterangan_sesudah, biaya_aktual_sebelum, biaya_aktual_sesudah, aksi, diubah_oleh)
VALUES
  ($1, NULLIF($2,''), NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6, $7, NULLIF($8,''), NULL)
`
	_, err := exec.ExecContext(
		ctx,
		q,
		in.TrxID,
		in.StatusSebelum,
		in.StatusSesudah,
		in.KeteranganSebelum,
		in.KeteranganSesudah,
		in.BiayaAktualSebelum,
		in.BiayaAktualSesudah,
		in.Aksi,
	)
	return err
}
