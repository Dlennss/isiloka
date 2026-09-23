package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

type ReconcileHandler struct {
	db *sql.DB
}

func NewReconcileHandler(db *sql.DB) *ReconcileHandler {
	return &ReconcileHandler{db: db}
}

type reconcileResult struct {
	MemberFixed        int `json:"member_fixed"`
	MemberDebitFixed   int `json:"member_debit_fixed"`
	ProviderDebitFixed int `json:"provider_debit_fixed"`
}

// Reconcile — cek 10 menit terakhir, fix berdasarkan database
// Prinsip: provider success = kebenaran. Member harus ikut.
func (h *ReconcileHandler) Reconcile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result := reconcileResult{}

	// 1. Provider success tapi member belum success → settle member success
	n1, err := h.fixMemberNotSettled(ctx)
	if err != nil {
		log.Printf("[reconcile] fixMemberNotSettled error: %v", err)
	}
	result.MemberFixed = n1

	// 2. Member success tapi saldo member belum dipotong (debit < biaya_perkiraan) → potong
	n2, err := h.fixMemberSaldoNotDeducted(ctx)
	if err != nil {
		log.Printf("[reconcile] fixMemberSaldoNotDeducted error: %v", err)
	}
	result.MemberDebitFixed = n2

	// 3. Member success tapi saldo provider belum dipotong → potong
	n3, err := h.fixProviderSaldoNotDeducted(ctx)
	if err != nil {
		log.Printf("[reconcile] fixProviderSaldoNotDeducted error: %v", err)
	}
	result.ProviderDebitFixed = n3

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result})
}

// Provider success tapi member masih pending → settle member success
func (h *ReconcileHandler) fixMemberNotSettled(ctx context.Context) (int, error) {
	rows, err := h.db.QueryContext(ctx, `
SELECT DISTINCT ON (tm.id)
  tm.id AS trx_id, tm.ref_id, tm.member_id, tm.biaya_perkiraan, tm.harga_member,
  tp.id AS provider_id, tp.provider, tp.harga AS provider_harga,
  COALESCE(tp.no_referensi, '') AS no_referensi,
  COALESCE(tp.pesan, '') AS pesan
FROM public.transaksi_member tm
JOIN public.transaksi_provider tp ON tp.ref_id = tm.ref_id
WHERE tm.dibuat_pada > now() - interval '10 minutes'
  AND lower(trim(tm.status)) = 'pending'
  AND lower(trim(tp.status)) = 'success'
ORDER BY tm.id, tp.id DESC`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var trxID, memberID, biayaPerkiraan, hargaMember, providerID, providerHarga int64
		var refID, provider, noReferensi, pesan string
		if err := rows.Scan(&trxID, &refID, &memberID, &biayaPerkiraan, &hargaMember,
			&providerID, &provider, &providerHarga, &noReferensi, &pesan); err != nil {
			continue
		}

		// Settle member success
		ket := strings.TrimSpace(noReferensi)
		if ket == "" {
			ket = "Transaksi berhasil"
		}
		hm := hargaMember
		if hm <= 0 {
			hm = biayaPerkiraan
		}

		res, err := h.db.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status = 'success', keterangan = $2, biaya_aktual = $3, harga_member = $4, diperbarui_pada = now()
WHERE id = $1 AND lower(trim(status)) = 'pending'`,
			trxID, ket, biayaPerkiraan, hm)
		if err != nil {
			log.Printf("[reconcile] settle member failed trx=%d ref=%s err=%v", trxID, refID, err)
			continue
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			continue // member sudah bukan pending, skip
		}

		log.Printf("[reconcile] member settled success trx=%d ref=%s provider=%s", trxID, refID, provider)

		// Kirim callback ke member
		h.sendMemberCallback(ctx, memberID, refID, "success", ket, noReferensi, biayaPerkiraan)

		count++
	}
	return count, nil
}

// Member success tapi saldo belum dipotong (net debit < biaya_perkiraan)
func (h *ReconcileHandler) fixMemberSaldoNotDeducted(ctx context.Context) (int, error) {
	rows, err := h.db.QueryContext(ctx, `
SELECT tm.id, tm.ref_id, tm.member_id, tm.biaya_perkiraan,
  COALESCE(SUM(CASE WHEN md.arah = 'DEBIT' THEN md.jumlah ELSE 0 END), 0) AS total_debit,
  COALESCE(SUM(CASE WHEN md.arah = 'CREDIT' THEN md.jumlah ELSE 0 END), 0) AS total_credit
FROM public.transaksi_member tm
LEFT JOIN public.mutasi_dompet md ON md.ref_id = tm.ref_id AND md.member_id = tm.member_id
WHERE tm.dibuat_pada > now() - interval '10 minutes'
  AND lower(trim(tm.status)) = 'success'
  AND tm.biaya_perkiraan > 0
GROUP BY tm.id, tm.ref_id, tm.member_id, tm.biaya_perkiraan
HAVING COALESCE(SUM(CASE WHEN md.arah = 'DEBIT' THEN md.jumlah ELSE 0 END), 0)
     - COALESCE(SUM(CASE WHEN md.arah = 'CREDIT' THEN md.jumlah ELSE 0 END), 0)
     < tm.biaya_perkiraan`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var trxID, memberID, biayaPerkiraan, totalDebit, totalCredit int64
		var refID string
		if err := rows.Scan(&trxID, &refID, &memberID, &biayaPerkiraan, &totalDebit, &totalCredit); err != nil {
			continue
		}

		net := totalDebit - totalCredit
		diff := biayaPerkiraan - net
		if diff <= 0 {
			continue
		}

		// Cek apakah sudah pernah di-fix (cegah double debit)
		var alreadyFixed int
		h.db.QueryRowContext(ctx, `SELECT count(*) FROM public.mutasi_dompet WHERE ref_id = $1 AND member_id = $2 AND alasan = 'CALLBACK_SUCCESS_RECOVERY'`, refID, memberID).Scan(&alreadyFixed)
		if alreadyFixed > 0 {
			continue
		}

		// Debit selisih — atomic: update saldo + insert mutasi dalam satu transaksi
		tx, txErr := h.db.BeginTx(ctx, nil)
		if txErr != nil {
			continue
		}
		var saldoBefore int64
		tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE`, memberID).Scan(&saldoBefore)
		tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo = saldo - $2 WHERE member_id = $1`, memberID, diff)
		tx.ExecContext(ctx, `INSERT INTO public.mutasi_dompet (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah) VALUES ($1, $2, 'DEBIT', $3, 'CALLBACK_SUCCESS_RECOVERY', 'auto reconcile: saldo member kurang potong', $4, $5)`,
			memberID, refID, diff, saldoBefore, saldoBefore-diff)
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			log.Printf("[reconcile] member debit tx failed ref=%s member=%d diff=%d err=%v", refID, memberID, diff, err)
			continue
		}
		log.Printf("[reconcile] member saldo fixed ref=%s member=%d diff=%d", refID, memberID, diff)
		count++
	}
	return count, nil
}

// Provider success tapi saldo provider belum dipotong
func (h *ReconcileHandler) fixProviderSaldoNotDeducted(ctx context.Context) (int, error) {
	rows, err := h.db.QueryContext(ctx, `
SELECT tp.id, tp.ref_id, tp.provider, tp.transaksi_member_id, COALESCE(tp.harga, 0) AS harga
FROM public.transaksi_provider tp
WHERE tp.dibuat_pada > now() - interval '10 minutes'
  AND lower(trim(tp.status)) = 'success'
  AND COALESCE(tp.harga, 0) > 0
  AND NOT EXISTS (
    SELECT 1 FROM public.mutasi_dompet_provider mdp
    WHERE mdp.transaksi_provider_id = tp.id AND lower(trim(mdp.arah)) = 'debit'
  )`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var tpID, trxMemberID, harga int64
		var refID, provider string
		if err := rows.Scan(&tpID, &refID, &provider, &trxMemberID, &harga); err != nil {
			continue
		}

		_, err := h.db.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider (provider, ref_id, arah, jumlah, alasan, catatan, transaksi_member_id, transaksi_provider_id)
VALUES ($1, $2, 'debit', $3, 'TRX_SUCCESS_COST', 'auto reconcile: provider wallet belum dipotong', $4, $5)
ON CONFLICT DO NOTHING`, provider, refID, harga, trxMemberID, tpID)
		if err != nil {
			log.Printf("[reconcile] provider debit failed ref=%s provider=%s err=%v", refID, provider, err)
			continue
		}

		h.db.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo = saldo - $2 WHERE provider = $1`, provider, harga)
		log.Printf("[reconcile] provider saldo fixed ref=%s provider=%s harga=%d", refID, provider, harga)
		count++
	}
	return count, nil
}

func (h *ReconcileHandler) sendMemberCallback(ctx context.Context, memberID int64, refID, status, ket, providerRef string, biayaAktual int64) {
	var webhookURL string
	err := h.db.QueryRowContext(ctx, `SELECT COALESCE(webhook_url, '') FROM public.member WHERE id = $1`, memberID).Scan(&webhookURL)
	if err != nil || strings.TrimSpace(webhookURL) == "" {
		return
	}

	var saldo int64
	h.db.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1`, memberID).Scan(&saldo)

	payload := map[string]any{
		"refid":          refID,
		"status":         status,
		"member_balance": saldo,
		"trx": map[string]any{
			"message":      ket,
			"provider_ref": providerRef,
			"sn":           providerRef,
			"biaya_aktual": biayaAktual,
		},
	}

	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	st, _, cbErr := helper.PostJSON(cctx, webhookURL, payload)
	if cbErr != nil {
		log.Printf("[reconcile] callback failed ref=%s member=%d url=%s err=%v", refID, memberID, webhookURL, cbErr)
	} else {
		log.Printf("[reconcile] callback sent ref=%s member=%d http=%d", refID, memberID, st)
	}
}
