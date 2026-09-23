package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"

	"pulsa2/config"
	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

type anomalyRow struct {
	TrxID               int64
	MemberID            int64
	RefID               string
	MemberStatus        string
	BiayaPerkiraan      int64
	HargaMember         int64
	ProviderID          int64
	Provider            string
	ProviderState       string
	ProviderHarga       int64
	ProviderRank        int64
	MemberDebitCount    int64
	MemberCreditCount   int64
	MemberDebitTotal    int64
	MemberCreditTotal   int64
	ProviderDebitCount  int64
	ProviderCreditCount int64
	ProviderDebitTotal  int64
	ProviderCreditTotal int64
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	cfg := config.Load()
	dbConn := connectDB(cfg.DatabaseURL)
	defer dbConn.Close()
	unlock, locked, err := tryAcquireRunLock(context.Background(), dbConn)
	if err != nil {
		log.Fatalf("acquire lock: %v", err)
	}
	if !locked {
		log.Printf("reconcile-provider-truth skipped: another run is still active")
		return
	}
	defer unlock()

	memberRepo := repository.NewMemberTrxMemberRepository(dbConn)
	callbackRepo := repository.NewProviderCallbackRepository(dbConn)
	historyRepo := repository.NewHistoryRepository(dbConn)
	window := reconcileWindow()
	backlogWindow := reconcileBacklogWindow()
	refIDs := reconcileRefIDs()

	const batchLimit = 200
	totalRows := 0
	fixed := 0
	for {
		rows, err := loadActionableAnomalies(context.Background(), dbConn, batchLimit, window, backlogWindow, refIDs)
		if err != nil {
			log.Fatalf("load anomalies: %v", err)
		}
		if len(rows) == 0 {
			break
		}
		totalRows += len(rows)
		fixedThisBatch := 0
		for _, row := range rows {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			err := reconcileOne(ctx, memberRepo, callbackRepo, historyRepo, row)
			cancel()
			if err != nil {
				log.Printf("reconcile refid=%s trx_id=%d provider_id=%d failed: %v", row.RefID, row.TrxID, row.ProviderID, err)
				continue
			}
			fixed++
			fixedThisBatch++
		}
		if len(rows) < batchLimit || len(refIDs) > 0 || fixedThisBatch == 0 {
			break
		}
	}

	log.Printf("reconcile-provider-truth done rows=%d fixed=%d", totalRows, fixed)
}

func tryAcquireRunLock(ctx context.Context, db *sql.DB) (func(), bool, error) {
	const lockKey int64 = 202604060001
	var locked bool
	if err := db.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, lockKey).Scan(&locked); err != nil {
		return nil, false, err
	}
	if !locked {
		return func() {}, false, nil
	}
	return func() {
		_, _ = db.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)
	}, true, nil
}

func reconcileWindow() string {
	v := strings.TrimSpace(os.Getenv("RECONCILE_PROVIDER_WINDOW"))
	if v == "" {
		return "10 minutes"
	}
	return v
}

func reconcileBacklogWindow() string {
	v := strings.TrimSpace(os.Getenv("RECONCILE_PROVIDER_BACKLOG_WINDOW"))
	if v == "" {
		return "2 days"
	}
	return v
}

func reconcileRefIDs() []string {
	v := strings.TrimSpace(os.Getenv("RECONCILE_PROVIDER_REFIDS"))
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func connectDB(dsn string) *sql.DB {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	return conn
}

func loadActionableAnomalies(ctx context.Context, db *sql.DB, limit int, window string, backlogWindow string, refIDs []string) ([]anomalyRow, error) {
	q := `
WITH latest_provider AS (
  SELECT
    tp.transaksi_member_id,
    tp.id AS transaksi_provider_id,
    tp.provider,
    tp.ref_id,
    COALESCE(tp.harga, 0) AS provider_harga,
    tp.dibuat_pada,
    ROW_NUMBER() OVER (PARTITION BY tp.transaksi_member_id ORDER BY tp.dibuat_pada DESC, tp.id DESC) AS provider_rank,
    lower(COALESCE(trim(tp.status), 'unknown')) AS provider_state
  FROM public.transaksi_provider tp
),
member_mutasi AS (
  SELECT ref_id,
    count(*) FILTER (WHERE lower(trim(COALESCE(arah, ''))) = 'debit') AS member_debit_count,
    count(*) FILTER (WHERE lower(trim(COALESCE(arah, ''))) = 'credit') AS member_credit_count,
    COALESCE(sum(CASE WHEN lower(trim(COALESCE(arah, ''))) = 'debit' THEN jumlah ELSE 0 END), 0) AS member_debit_total,
    COALESCE(sum(CASE WHEN lower(trim(COALESCE(arah, ''))) = 'credit' THEN jumlah ELSE 0 END), 0) AS member_credit_total
  FROM public.mutasi_dompet
  GROUP BY ref_id
),
provider_mutasi AS (
  SELECT transaksi_provider_id,
    count(*) FILTER (WHERE lower(trim(COALESCE(arah, ''))) = 'debit') AS provider_debit_count,
    count(*) FILTER (WHERE lower(trim(COALESCE(arah, ''))) = 'credit') AS provider_credit_count,
    COALESCE(sum(CASE WHEN lower(trim(COALESCE(arah, ''))) = 'debit' THEN jumlah ELSE 0 END), 0) AS provider_debit_total,
    COALESCE(sum(CASE WHEN lower(trim(COALESCE(arah, ''))) = 'credit' THEN jumlah ELSE 0 END), 0) AS provider_credit_total
  FROM public.mutasi_dompet_provider
  WHERE transaksi_provider_id IS NOT NULL
  GROUP BY transaksi_provider_id
)
SELECT
  tm.id,
  tm.member_id,
  tm.ref_id,
  tm.status,
  COALESCE(tm.biaya_perkiraan, 0),
  COALESCE(tm.harga_member, 0),
  lp.transaksi_provider_id,
  lp.provider,
  lp.provider_state,
  lp.provider_harga,
  lp.provider_rank,
  COALESCE(mm.member_debit_count, 0),
  COALESCE(mm.member_credit_count, 0),
  COALESCE(mm.member_debit_total, 0),
  COALESCE(mm.member_credit_total, 0),
  COALESCE(pm.provider_debit_count, 0),
  COALESCE(pm.provider_credit_count, 0),
  COALESCE(pm.provider_debit_total, 0),
  COALESCE(pm.provider_credit_total, 0)
FROM public.transaksi_member tm
JOIN latest_provider lp ON lp.transaksi_member_id = tm.id
LEFT JOIN member_mutasi mm ON mm.ref_id = tm.ref_id
LEFT JOIN provider_mutasi pm ON pm.transaksi_provider_id = lp.transaksi_provider_id
WHERE tm.dibuat_pada >= now() - interval '3 days'
  AND ($3::text[] IS NULL OR tm.ref_id = ANY($3))
  AND lp.provider_state IN ('success', 'failed')
  AND (
    $3::text[] IS NOT NULL
    OR EXISTS (
      SELECT 1
      FROM public.transaksi_provider tp_recent
      WHERE tp_recent.id = lp.transaksi_provider_id
        AND lp.dibuat_pada >= now() - ($2::interval)
    )
    OR EXISTS (
      SELECT 1
      FROM public.transaksi_provider tp_backlog
      WHERE tp_backlog.id = lp.transaksi_provider_id
        AND lp.dibuat_pada >= now() - ($4::interval)
        AND (
          (lp.provider_state = 'success' AND ((lp.provider_rank = 1 AND (lower(trim(tm.status)) <> 'success' OR COALESCE(mm.member_debit_total, 0) - COALESCE(mm.member_credit_total, 0) <> COALESCE(tm.biaya_perkiraan, 0))) OR COALESCE(pm.provider_debit_total, 0) - COALESCE(pm.provider_credit_total, 0) <> COALESCE(lp.provider_harga, 0)))
          OR
          (lp.provider_state = 'failed' AND ((lp.provider_rank = 1 AND (lower(trim(tm.status)) <> 'failed' OR COALESCE(mm.member_debit_total, 0) <> COALESCE(mm.member_credit_total, 0))) OR COALESCE(pm.provider_debit_total, 0) <> COALESCE(pm.provider_credit_total, 0)))
        )
    )
  )
  AND (
    (lp.provider_state = 'success' AND ((lp.provider_rank = 1 AND (lower(trim(tm.status)) <> 'success' OR COALESCE(mm.member_debit_total, 0) - COALESCE(mm.member_credit_total, 0) <> COALESCE(tm.biaya_perkiraan, 0))) OR COALESCE(pm.provider_debit_total, 0) - COALESCE(pm.provider_credit_total, 0) <> COALESCE(lp.provider_harga, 0)))
    OR
    (lp.provider_state = 'failed' AND ((lp.provider_rank = 1 AND (lower(trim(tm.status)) <> 'failed' OR COALESCE(mm.member_debit_total, 0) <> COALESCE(mm.member_credit_total, 0))) OR COALESCE(pm.provider_debit_total, 0) <> COALESCE(pm.provider_credit_total, 0)))
    OR
    ($3::text[] IS NOT NULL AND tm.ref_id = ANY($3))
  )
ORDER BY tm.id DESC, lp.provider_rank ASC, lp.transaksi_provider_id DESC
LIMIT $1`

	var refArg any
	if len(refIDs) > 0 {
		refArg = pq.Array(refIDs)
	}
	rs, err := db.QueryContext(ctx, q, limit, window, refArg, backlogWindow)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var out []anomalyRow
	for rs.Next() {
		var row anomalyRow
		if err := rs.Scan(
			&row.TrxID,
			&row.MemberID,
			&row.RefID,
			&row.MemberStatus,
			&row.BiayaPerkiraan,
			&row.HargaMember,
			&row.ProviderID,
			&row.Provider,
			&row.ProviderState,
			&row.ProviderHarga,
			&row.ProviderRank,
			&row.MemberDebitCount,
			&row.MemberCreditCount,
			&row.MemberDebitTotal,
			&row.MemberCreditTotal,
			&row.ProviderDebitCount,
			&row.ProviderCreditCount,
			&row.ProviderDebitTotal,
			&row.ProviderCreditTotal,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rs.Err()
}

func reconcileOne(ctx context.Context, memberRepo *repository.MemberTrxMemberRepository, callbackRepo *repository.ProviderCallbackRepository, historyRepo *repository.HistoryRepository, row anomalyRow) error {
	price := row.ProviderHarga
	providerName := strings.ToLower(strings.TrimSpace(row.Provider))
	changed := false

	switch row.ProviderState {
	case "success":
		if row.ProviderRank == 1 {
			target, err := historyRepo.GetAdminTrxCallbackTarget(ctx, row.TrxID)
			if err != nil {
				return err
			}
			providerRow, err := historyRepo.GetLatestProviderByRefID(ctx, row.RefID)
			if err != nil {
				return err
			}
			if providerRow == nil {
				return errors.New("latest provider row not found")
			}

			ket, providerRef, sn, callbackPrice, callbackProviderName := buildCallbackInfo(row.ProviderState, target, providerRow)
			if callbackPrice > 0 {
				price = callbackPrice
			}
			if callbackProviderName != "" {
				providerName = callbackProviderName
			}
			hargaMember := effectiveMemberSellingPrice(row.HargaMember, row.BiayaPerkiraan)
			memberNet := row.MemberDebitTotal - row.MemberCreditTotal

			if strings.TrimSpace(strings.ToLower(target.Status)) != "success" {
				if err := memberRepo.ForceReconcileFailedToSuccess(ctx, target.ID, ket, row.BiayaPerkiraan, price, hargaMember); err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				changed = true
			}
			switch {
			case memberNet < row.BiayaPerkiraan:
				diff := row.BiayaPerkiraan - memberNet
				if row.MemberCreditTotal > 0 {
					if err := callbackRepo.RecoverMemberRefund(ctx, target.MemberID, target.RefID, diff, "CALLBACK_SUCCESS_RECOVERY", "recovery refund nominal selisih saat provider success"); err != nil && !errors.Is(err, repository.ErrInsufficientBalance) {
						return err
					}
				} else if diff > 0 {
					if _, _, err := memberRepo.DebitDompet(ctx, target.MemberID, target.RefID, diff, "STATUS_PAY_REDEBIT", "debit nominal selisih saat reconcile provider success"); err != nil {
						return err
					}
				}
				changed = changed || diff > 0
			case memberNet > row.BiayaPerkiraan:
				diff := memberNet - row.BiayaPerkiraan
				if diff > 0 {
					if err := callbackRepo.CreditDompet(ctx, target.MemberID, target.RefID, diff, "CALLBACK_SUCCESS_OVERDEBIT_REFUND", "refund kelebihan potong nominal saat provider success"); err != nil {
						return err
					}
					changed = true
				}
			}

			if changed {
				httpStatus, body, err := sendMemberCallback(ctx, historyRepo, row.TrxID, row.ProviderState, ket, providerRef, sn, providerName)
				if err != nil {
					return fmt.Errorf("callback failed http=%d body=%s err=%w", httpStatus, body, err)
				}
				log.Printf("reconciled refid=%s trx_id=%d provider=%s final=%s callback_http=%d", row.RefID, row.TrxID, providerName, row.ProviderState, httpStatus)
			}
		}
		providerNet := row.ProviderDebitTotal - row.ProviderCreditTotal
		switch {
		case providerNet < price && price > 0:
			diff := price - providerNet
			if _, _, err := callbackRepo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
				Provider:            row.Provider,
				RefID:               row.RefID,
				Arah:                "debit",
				Jumlah:              diff,
				AllowNegative:       true,
				Alasan:              "TRX_SUCCESS_COST_ADJUST",
				Catatan:             "debit selisih nominal provider saat reconcile success",
				TransaksiMemberID:   &row.TrxID,
				TransaksiProviderID: &row.ProviderID,
			}); err != nil {
				return err
			}
			changed = true
		case providerNet > price:
			diff := providerNet - price
			if _, _, err := callbackRepo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
				Provider:            row.Provider,
				RefID:               row.RefID,
				Arah:                "credit",
				Jumlah:              diff,
				Alasan:              "TRX_SUCCESS_COST_REFUND_ADJUST",
				Catatan:             "refund kelebihan potong nominal provider saat reconcile success",
				TransaksiMemberID:   &row.TrxID,
				TransaksiProviderID: &row.ProviderID,
			}); err != nil {
				return err
			}
			changed = true
		}
	case "failed":
		if row.ProviderRank == 1 {
			target, err := historyRepo.GetAdminTrxCallbackTarget(ctx, row.TrxID)
			if err != nil {
				return err
			}
			providerRow, err := historyRepo.GetLatestProviderByRefID(ctx, row.RefID)
			if err != nil {
				return err
			}
			if providerRow == nil {
				return errors.New("latest provider row not found")
			}

			ket, providerRef, sn, callbackPrice, callbackProviderName := buildCallbackInfo(row.ProviderState, target, providerRow)
			if callbackPrice > 0 {
				price = callbackPrice
			}
			if callbackProviderName != "" {
				providerName = callbackProviderName
			}
			hargaMember := effectiveMemberSellingPrice(row.HargaMember, row.BiayaPerkiraan)
			memberNet := row.MemberDebitTotal - row.MemberCreditTotal

			memberStatus := strings.TrimSpace(strings.ToLower(target.Status))
			if memberStatus == "success" {
				log.Printf("[SKIP] refid=%s trx_id=%d member sudah success, tidak boleh force failed", row.RefID, row.TrxID)
			} else if memberStatus != "failed" {
				if err := forceSettleFailed(ctx, callbackRepo.DB(), target.ID, ket, price, hargaMember); err != nil {
					return err
				}
				changed = true
			}
			switch {
			case memberNet > 0:
				if err := callbackRepo.CreditDompet(ctx, target.MemberID, target.RefID, memberNet, "REFUND", "refund nominal selisih saat reconcile provider failed"); err != nil {
					return err
				}
				changed = true
			case memberNet < 0:
				diff := -memberNet
				if _, _, err := memberRepo.DebitDompet(ctx, target.MemberID, target.RefID, diff, "FAILED_REFUND_RECOVERY", "recovery kelebihan refund saat reconcile provider failed"); err != nil {
					return err
				}
				changed = true
			}

			if changed {
				httpStatus, body, err := sendMemberCallback(ctx, historyRepo, row.TrxID, row.ProviderState, ket, providerRef, sn, providerName)
				if err != nil {
					return fmt.Errorf("callback failed http=%d body=%s err=%w", httpStatus, body, err)
				}
				log.Printf("reconciled refid=%s trx_id=%d provider=%s final=%s callback_http=%d", row.RefID, row.TrxID, providerName, row.ProviderState, httpStatus)
			}
		}
		providerNet := row.ProviderDebitTotal - row.ProviderCreditTotal
		switch {
		case providerNet > 0:
			if _, _, err := callbackRepo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
				Provider:            row.Provider,
				RefID:               row.RefID,
				Arah:                "credit",
				Jumlah:              providerNet,
				Alasan:              "TRX_FAILED_REFUND_ADJUST",
				Catatan:             "refund nominal selisih provider saat reconcile failed",
				TransaksiMemberID:   &row.TrxID,
				TransaksiProviderID: &row.ProviderID,
			}); err != nil {
				return err
			}
			changed = true
		case providerNet < 0:
			diff := -providerNet
			if _, _, err := callbackRepo.ApplyProviderWalletTx(ctx, repository.CallbackProviderWalletTxIn{
				Provider:            row.Provider,
				RefID:               row.RefID,
				Arah:                "debit",
				Jumlah:              diff,
				AllowNegative:       true,
				Alasan:              "TRX_FAILED_REFUND_RECOVER",
				Catatan:             "recovery kelebihan refund provider saat reconcile failed",
				TransaksiMemberID:   &row.TrxID,
				TransaksiProviderID: &row.ProviderID,
			}); err != nil {
				return err
			}
			changed = true
		}
	}
	return nil
}

func forceSettleFailed(ctx context.Context, db *sql.DB, trxID int64, ket string, hargaJavapay int64, hargaMember int64) error {
	res, err := db.ExecContext(ctx, `
WITH prev AS (
  SELECT
    id,
    status,
    COALESCE(keterangan, '') AS keterangan,
    COALESCE(biaya_aktual, 0) AS biaya_aktual
  FROM public.transaksi_member
  WHERE id = $1
),
upd AS (
  UPDATE public.transaksi_member
  SET status = 'failed',
      keterangan = NULLIF($2,''),
      biaya_aktual = 0,
      harga_javapay = $3,
      harga_member = $4,
      diperbarui_pada = now()
  WHERE id = $1
    AND LOWER(COALESCE(status, '')) != 'success'
  RETURNING id
)
INSERT INTO public.transaksi_member_status_log
  (transaksi_member_id, status_sebelum, status_sesudah, keterangan_sebelum, keterangan_sesudah, biaya_aktual_sebelum, biaya_aktual_sesudah, aksi, diubah_oleh)
SELECT
  prev.id, prev.status, 'failed', prev.keterangan, COALESCE(NULLIF($2,''), ''), prev.biaya_aktual, 0, 'provider_truth_force_failed', NULL
FROM prev
WHERE EXISTS (SELECT 1 FROM upd)
`, trxID, ket, hargaJavapay, hargaMember)
	if err != nil {
		return err
	}
	if aff, affErr := res.RowsAffected(); affErr == nil && aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func sendMemberCallback(ctx context.Context, historyRepo *repository.HistoryRepository, trxID int64, finalStatus, ket, providerRef, sn, providerName string) (int, string, error) {
	trx, err := historyRepo.GetAdminTrxCallbackTarget(ctx, trxID)
	if err != nil {
		return 0, "", err
	}
	webhookURL, err := historyRepo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil {
		return 0, "", err
	}
	if strings.TrimSpace(webhookURL) == "" {
		return 0, "", errors.New("member webhook tidak diatur")
	}
	memberBalance, err := historyRepo.GetSaldo(ctx, trx.MemberID)
	if err != nil {
		return 0, "", err
	}
	biayaAktualOut := int64(0)
	if finalStatus == "success" {
		biayaAktualOut = trx.BiayaAktual
		if biayaAktualOut <= 0 {
			biayaAktualOut = trx.BiayaPerkiraan
		}
	}
	payload := buildCallbackPayload(trx, finalStatus, ket, providerRef, sn, memberBalance, biayaAktualOut)
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	httpStatus, body, err := helper.PostJSON(cctx, webhookURL, payload)
	if err == nil {
		log.Printf("callback sent refid=%s trx_id=%d provider=%s status=%s http=%d", trx.RefID, trx.ID, providerName, finalStatus, httpStatus)
	}
	return httpStatus, body, err
}

func buildCallbackPayload(trx *repository.AdminTrxCallbackTarget, finalStatus, ket, providerRef, sn string, memberBalance, biayaAktualOut int64) map[string]any {
	qtyProvider := trx.QtyProvider
	if qtyProvider <= 0 {
		qtyProvider = trx.Qty
	}
	snField := strings.TrimSpace(sn)
	if strings.EqualFold(strings.TrimSpace(finalStatus), "failed") && snField == "" {
		snField = strings.TrimSpace(ket)
	}
	return map[string]any{
		"refid":          trx.RefID,
		"status":         finalStatus,
		"member_balance": memberBalance,
		"trx": map[string]any{
			"id":           trx.ID,
			"commands":     trx.Perintah,
			"product":      trx.KodeProduk,
			"dest":         trx.Tujuan,
			"qty":          trx.Qty,
			"qty_provider": qtyProvider,
			"biaya_aktual": biayaAktualOut,
			"message":      ket,
			"provider_ref": providerRef,
			"sn":           snField,
		},
	}
}

func buildCallbackInfo(finalStatus string, trx *repository.AdminTrxCallbackTarget, row *model.JavapayTrxRow) (ket string, providerRef string, sn string, price int64, provider string) {
	if row == nil {
		ket, _ = helper.SafeMemberKeterangan(finalStatus, strings.TrimSpace(trx.Keterangan))
		if ket == "" {
			ket = strings.TrimSpace(trx.Keterangan)
		}
		return ket, "", "", 0, ""
	}
	provider = strings.ToLower(strings.TrimSpace(row.Provider))
	msg := ""
	noreff := ""
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	if row.NoReferensi != nil {
		noreff = strings.TrimSpace(*row.NoReferensi)
	}
	if row.Harga != nil {
		price = *row.Harga
	}
	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	providerRef = strings.TrimSpace(info.Reff)
	sn = strings.TrimSpace(info.SN)
	switch provider {
	case "talentapay":
		parsedRef, parsedSN := providersn.ParseTalentaSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "multikom":
		parsedRef, parsedSN := providersn.ParseMultikomSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "yuscom":
		parsedRef, parsedSN := providersn.ParseYuscomSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "sagaramobile":
		parsedRef, parsedSN := providersn.ParseSagaraSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "minions":
		parsedRef, parsedSN := providersn.ParseMinionsSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "trionik":
		parsedRef, parsedSN := providersn.ParseYuscomSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	case "ajs":
		parsedRef, parsedSN := providersn.ParseAJSSNRefFromMsg(msg)
		providerRef, sn = mergeProviderInfo(providerRef, sn, parsedRef, parsedSN)
	}
	if sn == "" {
		sn = noreff
	}
	if providerRef == "" {
		providerRef = noreff
	}
	if ket == "" {
		ket = strings.TrimSpace(trx.Keterangan)
	}
	return ket, providerRef, sn, price, provider
}

func mergeProviderInfo(providerRef, sn string, parsedRef, parsedSN string) (string, string) {
	if isWeak(providerRef) {
		providerRef = strings.TrimSpace(parsedRef)
	}
	if isWeak(sn) {
		sn = strings.TrimSpace(parsedSN)
	}
	return providerRef, sn
}

func isWeak(v string) bool {
	switch strings.TrimSpace(strings.ToUpper(v)) {
	case "", "NO", "N/A", "-":
		return true
	default:
		return false
	}
}

func effectiveMemberSellingPrice(hargaMember, biayaPerkiraan int64) int64 {
	if hargaMember > 0 {
		return hargaMember
	}
	if biayaPerkiraan > 0 {
		return biayaPerkiraan
	}
	return 0
}
