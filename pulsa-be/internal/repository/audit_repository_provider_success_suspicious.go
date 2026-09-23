package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const providerSuccessSuspectTextMessageSQL = `(
  LOWER(COALESCE(tp.pesan, '')) LIKE '%pernah selesai%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%refund%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%saldo dikembalikan%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%stok pulsa kembali%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%pulsa kembali%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%sempat sukses%'
)`

const providerSuccessSuspectRefundMessageSQL = `(
  LOWER(COALESCE(tp.pesan, '')) LIKE '%refund%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%saldo dikembalikan%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%dikembalikan%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%stok pulsa kembali%'
  OR LOWER(COALESCE(tp.pesan, '')) LIKE '%pulsa kembali%'
)`

const providerSuccessSuspectFailedAnomalyRecordSQL = `(
  LOWER(COALESCE(a.pesan, '')) LIKE '%gagal%'
  OR LOWER(COALESCE(a.pesan, '')) LIKE '%failed%'
  OR LOWER(COALESCE(a.pesan, '')) LIKE '%batal%'
  OR LOWER(COALESCE(a.pesan, '')) LIKE '%ditolak%'
  OR (
    COALESCE(BTRIM(a.kode_respon), '') <> ''
    AND LOWER(TRIM(COALESCE(a.kode_respon, ''))) NOT IN ('00', '0', '9', '28', '68')
  )
)`

const providerSuccessSuspectMessageSQL = `(
  ` + providerSuccessSuspectTextMessageSQL + `
  OR failed_anomaly.id IS NOT NULL
  OR prior_suspect.id IS NOT NULL
)`

const providerSuccessSuspectAllProvidersFailedSQL = `(
  LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
  AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
  AND NOT EXISTS (
    SELECT 1
    FROM public.transaksi_provider tp_other
    WHERE tp_other.transaksi_member_id = tp.transaksi_member_id
      AND LOWER(TRIM(COALESCE(tp_other.status, ''))) <> 'failed'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM public.transaksi_provider tp_newer
    WHERE tp_newer.transaksi_member_id = tp.transaksi_member_id
      AND tp_newer.id > tp.id
  )
)`

const providerSuccessSuspectEvidenceSQL = `(
  ` + providerSuccessSuspectAllProvidersFailedSQL + `
  OR (
    COALESCE(BTRIM(COALESCE(failed_anomaly.pesan, tp.pesan, prior_suspect.pesan)), '') <> ''
    AND ` + providerSuccessSuspectMessageSQL + `
  )
)`

const providerSuccessSuspectEligibleSQL = `(
  LOWER(TRIM(COALESCE(tp.status, ''))) = 'success'
  OR (
    LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
    AND LOWER(COALESCE(tp.pesan, '')) LIKE '%sempat sukses%'
    AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'failed'
  )
  OR (
    LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
    AND ` + providerSuccessSuspectRefundMessageSQL + `
    AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
  )
)`

const providerSuccessSuspectEligibleWithAnomalySQL = `(
  ` + providerSuccessSuspectEligibleSQL + `
  OR ` + providerSuccessSuspectAllProvidersFailedSQL + `
  OR (
    LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
    AND failed_anomaly.id IS NOT NULL
  )
  OR prior_suspect.id IS NOT NULL
)`

const providerSuccessSuspectBankSettlementReason = "TRANSAKSI_SUSPECT_SELESAI"
const providerSuccessSuspectBankSettlementRefPrefix = "TRX-SUSPECT-"
const providerSuccessSuspectBankAccountNumber = "8761518283"
const providerSuccessSuspectBankLabel = "BCA SUSPECTTT"

const providerSuccessSuspectDirectSettlementLateralSQL = `
LEFT JOIN LATERAL (
  SELECT
    mb.id,
    mb.dibuat_pada AS resolved_at,
    mb.diubah_oleh AS user_id,
    COALESCE(mb.catatan, '') AS note
  FROM public.mutasi_bank mb
  WHERE mb.ref_id = '` + providerSuccessSuspectBankSettlementRefPrefix + `' || tp.id::text
    AND mb.alasan = '` + providerSuccessSuspectBankSettlementReason + `'
  ORDER BY mb.id DESC
  LIMIT 1
) direct_settlement ON true`

const providerSuccessSuspectSameRefResolvedLateralSQL = `
LEFT JOIN LATERAL (
  SELECT tp_same.id AS provider_row_id
  FROM public.transaksi_provider tp_same
  LEFT JOIN public.marking_trx_resolve mtr_same
    ON mtr_same.trx_type = 'provider'
   AND mtr_same.trx_id = tp_same.id
   AND mtr_same.mark_type = 'provider_success_suspect'
  LEFT JOIN LATERAL (
    SELECT mb_same.id
    FROM public.mutasi_bank mb_same
    WHERE mb_same.ref_id = '` + providerSuccessSuspectBankSettlementRefPrefix + `' || tp_same.id::text
      AND mb_same.alasan = '` + providerSuccessSuspectBankSettlementReason + `'
    ORDER BY mb_same.id DESC
    LIMIT 1
  ) same_settlement ON true
  WHERE tp_same.id <> tp.id
    AND tp_same.ref_id = COALESCE(NULLIF(BTRIM(tp.ref_id), ''), NULLIF(BTRIM(tm.ref_id), ''))
    AND (mtr_same.id IS NOT NULL OR same_settlement.id IS NOT NULL)
  ORDER BY tp_same.id DESC
  LIMIT 1
) same_ref_resolved ON true`

func (r *AuditRepository) AdminListProviderSuccessSuspiciousMessage(
	ctx context.Context,
	limit, offset int,
	provider, refID, resolveStatus, fromStr, toStr string,
	includeTotal bool,
) ([]AdminProviderSuccessSuspiciousMessageRow, int64, bool, error) {
	if err := r.ensureMarkingTrxResolveTable(ctx); err != nil {
		return nil, 0, false, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	var (
		args           []any
		wheres         []string
		providerArgPos int
		refIDArgPos    int
		fromArgPos     int
		toArgPos       int
	)

	wheres = append(wheres, providerSuccessSuspectEligibleWithAnomalySQL)
	wheres = append(wheres, "LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'")
	wheres = append(wheres, providerSuccessSuspectEvidenceSQL)
	wheres = append(wheres, "COALESCE(retry_guard.other_success_count, 0) = 0")
	wheres = append(wheres, "COALESCE(retry_guard.other_pending_count, 0) = 0")
	wheres = append(wheres, "(prior_suspect.id IS NOT NULL OR later_failed_loket.id IS NULL)")
	wheres = append(wheres, "(mtr.id IS NOT NULL OR direct_settlement.id IS NOT NULL OR same_ref_resolved.provider_row_id IS NULL)")

	resolvedExpr := "(mtr.id IS NOT NULL OR direct_settlement.id IS NOT NULL OR auto_b2.provider_row_id IS NOT NULL)"

	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider != "" {
		args = append(args, provider)
		providerArgPos = len(args)
		wheres = append(wheres, fmt.Sprintf("LOWER(TRIM(tp.provider)) = $%d", providerArgPos))
	}
	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		refIDArgPos = len(args)
		wheres = append(wheres, fmt.Sprintf("(tp.ref_id ILIKE '%%' || $%d || '%%' OR tm.ref_id ILIKE '%%' || $%d || '%%')", len(args), len(args)))
	}
	resolveStatus = strings.TrimSpace(strings.ToLower(resolveStatus))
	switch resolveStatus {
	case "", "unresolved":
		wheres = append(wheres, "NOT "+resolvedExpr)
	case "resolved":
		wheres = append(wheres, resolvedExpr)
	case "all":
	default:
		return nil, 0, false, fmt.Errorf("invalid resolve_status")
	}
	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, 0, false, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		fromArgPos = len(args)
		wheres = append(wheres, fmt.Sprintf("COALESCE(failed_anomaly.dibuat_pada, tp.dibuat_pada) >= $%d", len(args)))
	}
	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, 0, false, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		toArgPos = len(args)
		wheres = append(wheres, fmt.Sprintf("COALESCE(failed_anomaly.dibuat_pada, tp.dibuat_pada) < $%d", len(args)))
	}

	whereSQL := strings.Join(wheres, " AND ")
	textMessageTP0 := strings.ReplaceAll(providerSuccessSuspectTextMessageSQL, "tp.", "tp0.")
	refundMessageTP0 := strings.ReplaceAll(providerSuccessSuspectRefundMessageSQL, "tp.", "tp0.")
	failedAnomalyA0 := strings.ReplaceAll(providerSuccessSuspectFailedAnomalyRecordSQL, "a.", "a0.")
	priorTextTP := strings.ReplaceAll(providerSuccessSuspectTextMessageSQL, "tp.", "tp_prior.")
	allProvidersFailedTP0 := strings.ReplaceAll(
		strings.ReplaceAll(providerSuccessSuspectAllProvidersFailedSQL, "tp.", "tp0."),
		"tm.",
		"tm0.",
	)

	var candidateTextExtra, candidateAnomalyExtra string
	if providerArgPos > 0 {
		filter := fmt.Sprintf("\n  AND LOWER(TRIM(tp0.provider)) = $%d", providerArgPos)
		candidateTextExtra += filter
		candidateAnomalyExtra += filter
	}
	if refIDArgPos > 0 {
		filter := fmt.Sprintf("\n  AND (tp0.ref_id ILIKE '%%' || $%d || '%%' OR tm0.ref_id ILIKE '%%' || $%d || '%%')", refIDArgPos, refIDArgPos)
		candidateTextExtra += filter
		candidateAnomalyExtra += filter
	}
	if fromArgPos > 0 {
		candidateTextExtra += fmt.Sprintf("\n  AND tp0.dibuat_pada >= $%d", fromArgPos)
		candidateAnomalyExtra += fmt.Sprintf("\n  AND a0.dibuat_pada >= $%d", fromArgPos)
	}
	if toArgPos > 0 {
		candidateTextExtra += fmt.Sprintf("\n  AND tp0.dibuat_pada < $%d", toArgPos)
		candidateAnomalyExtra += fmt.Sprintf("\n  AND a0.dibuat_pada < $%d", toArgPos)
	}
	candidateResolvedSQL := `
  AND (
    EXISTS (
      SELECT 1
      FROM public.marking_trx_resolve mtr0
      WHERE mtr0.trx_type = 'provider'
        AND mtr0.trx_id = tp0.id
        AND mtr0.mark_type = 'provider_success_suspect'
    )
    OR EXISTS (
      SELECT 1
      FROM public.mutasi_bank mb0
      WHERE mb0.ref_id = '` + providerSuccessSuspectBankSettlementRefPrefix + `' || tp0.id::text
        AND mb0.alasan = '` + providerSuccessSuspectBankSettlementReason + `'
    )
    OR EXISTS (
      SELECT 1
      FROM public.transaksi_provider tp_b2
      WHERE tp_b2.transaksi_member_id = tp0.transaksi_member_id
        AND tp_b2.id <> tp0.id
        AND LOWER(TRIM(COALESCE(tp_b2.provider, ''))) = 'smb'
        AND LOWER(TRIM(COALESCE(tp_b2.status, ''))) = 'success'
        AND UPPER(TRIM(COALESCE(tp_b2.kode_produk, ''))) LIKE 'BIFASTOPEN2:%'
    )
  )`
	candidateUnresolvedSQL := `
  AND NOT EXISTS (
    SELECT 1
    FROM public.marking_trx_resolve mtr0
    WHERE mtr0.trx_type = 'provider'
      AND mtr0.trx_id = tp0.id
      AND mtr0.mark_type = 'provider_success_suspect'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM public.mutasi_bank mb0
    WHERE mb0.ref_id = '` + providerSuccessSuspectBankSettlementRefPrefix + `' || tp0.id::text
      AND mb0.alasan = '` + providerSuccessSuspectBankSettlementReason + `'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM public.transaksi_provider tp_b2
    WHERE tp_b2.transaksi_member_id = tp0.transaksi_member_id
      AND tp_b2.id <> tp0.id
      AND LOWER(TRIM(COALESCE(tp_b2.provider, ''))) = 'smb'
      AND LOWER(TRIM(COALESCE(tp_b2.status, ''))) = 'success'
      AND UPPER(TRIM(COALESCE(tp_b2.kode_produk, ''))) LIKE 'BIFASTOPEN2:%'
  )`
	switch resolveStatus {
	case "", "unresolved":
		candidateTextExtra += candidateUnresolvedSQL
		candidateAnomalyExtra += candidateUnresolvedSQL
	case "resolved":
		candidateTextExtra += candidateResolvedSQL
		candidateAnomalyExtra += candidateResolvedSQL
	}

	// Build a small candidate set before running per-row lateral checks.
	candidateCTE := fmt.Sprintf(`WITH candidate_ids AS (
  SELECT DISTINCT tp0.id
  FROM public.transaksi_provider tp0
  LEFT JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  WHERE LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND (
      LOWER(TRIM(COALESCE(tp0.status, ''))) = 'success'
      OR (
        LOWER(TRIM(COALESCE(tp0.status, ''))) = 'failed'
        AND %s
      )
    )
    AND %s%s
  UNION
  SELECT DISTINCT tp0.id
  FROM public.transaksi_provider tp0
  LEFT JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  WHERE %s%s
  UNION
  SELECT DISTINCT tp0.id
  FROM public.transaksi_anomasi_provider a0
  JOIN public.transaksi_provider tp0
    ON tp0.ref_id = a0.ref_id
   AND LOWER(TRIM(COALESCE(tp0.provider, ''))) = LOWER(TRIM(COALESCE(a0.provider, '')))
   AND a0.dibuat_pada >= tp0.dibuat_pada
  LEFT JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  WHERE LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND (
      LOWER(TRIM(COALESCE(tp0.status, ''))) = 'success'
      OR LOWER(TRIM(COALESCE(tp0.status, ''))) = 'failed'
    )
    AND %s%s
  UNION
  SELECT DISTINCT tp0.id
  FROM public.transaksi_provider tp0
  LEFT JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  JOIN LATERAL (
    SELECT tp_prior.id
    FROM public.transaksi_provider tp_prior
    WHERE tp_prior.transaksi_member_id = tp0.transaksi_member_id
      AND tp_prior.id < tp0.id
      AND LOWER(TRIM(COALESCE(tp_prior.status, ''))) = 'failed'
      AND %s
    ORDER BY tp_prior.id DESC
    LIMIT 1
  ) prior0 ON true
  WHERE LOWER(TRIM(COALESCE(tp0.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp0.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND COALESCE(BTRIM(tp0.pesan), '') <> ''
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp0.transaksi_member_id
        AND newer.id > tp0.id
    )%s
)`, refundMessageTP0, textMessageTP0, candidateTextExtra, allProvidersFailedTP0, candidateTextExtra, failedAnomalyA0, candidateAnomalyExtra, priorTextTP, candidateTextExtra)
	baseFrom := `
FROM candidate_ids c
JOIN public.transaksi_provider tp ON tp.id = c.id
LEFT JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
LEFT JOIN public.member m ON m.id = tm.member_id
LEFT JOIN public.marking_trx_resolve mtr
  ON mtr.trx_type = 'provider'
 AND mtr.trx_id = tp.id
 AND mtr.mark_type = 'provider_success_suspect'
` + providerSuccessSuspectDirectSettlementLateralSQL + `
LEFT JOIN public.member resolved_user ON resolved_user.id = COALESCE(mtr.user_id, direct_settlement.user_id)
` + providerSuccessSuspectSameRefResolvedLateralSQL + `
LEFT JOIN LATERAL (
  SELECT
    COUNT(*) FILTER (
      WHERE LOWER(TRIM(COALESCE(tp2.status, ''))) = 'success'
    ) AS other_success_count,
    COUNT(*) FILTER (
      WHERE LOWER(TRIM(COALESCE(tp2.status, ''))) = 'pending'
    ) AS other_pending_count
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id <> tp.id
) retry_guard ON true
LEFT JOIN LATERAL (
  SELECT
    a.id,
    a.kode_respon,
    a.pesan,
    a.harga,
    a.dibuat_pada
  FROM public.transaksi_anomasi_provider a
  WHERE LOWER(TRIM(COALESCE(a.provider, ''))) = LOWER(TRIM(COALESCE(tp.provider, '')))
    AND a.ref_id = tp.ref_id
    AND a.dibuat_pada >= tp.dibuat_pada
    AND ` + providerSuccessSuspectFailedAnomalyRecordSQL + `
  ORDER BY a.id DESC
  LIMIT 1
) failed_anomaly ON true
LEFT JOIN LATERAL (
  SELECT
    tp2.id,
    tp2.kode_respon,
    tp2.pesan,
    tp2.harga,
    tp2.dibuat_pada
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id < tp.id
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tp.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
    AND ` + strings.ReplaceAll(providerSuccessSuspectTextMessageSQL, "tp.", "tp2.") + `
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp.transaksi_member_id
        AND newer.id > tp.id
    )
  ORDER BY tp2.id DESC
  LIMIT 1
) prior_suspect ON true
LEFT JOIN LATERAL (
  SELECT tp2.id
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id > tp.id
    AND LOWER(TRIM(COALESCE(tp2.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'failed'
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp2.transaksi_member_id
        AND newer.id > tp2.id
    )
  ORDER BY tp2.id DESC
  LIMIT 1
) later_failed_loket ON true
LEFT JOIN LATERAL (
  SELECT
    tp2.id AS provider_row_id,
    tp2.dibuat_pada AS resolved_at,
    'AUTO BIFASTOPEN2'::text AS resolved_by_name,
    CONCAT('Otomatis selesai karena callback sukses ', COALESCE(NULLIF(BTRIM(tp2.kode_produk), ''), 'BIFASTOPEN2'))::text AS resolve_note
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id <> tp.id
    AND LOWER(TRIM(COALESCE(tp2.provider, ''))) = 'smb'
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'success'
    AND UPPER(TRIM(COALESCE(tp2.kode_produk, ''))) LIKE 'BIFASTOPEN2:%'
  ORDER BY tp2.id DESC
  LIMIT 1
) auto_b2 ON true
`

	queryLimit := limit
	if !includeTotal {
		queryLimit = limit + 1
		// Cache candidates still pass through lateral eligibility checks below, so keep a wide buffer.
		prefetchLimit := offset + (queryLimit * 10) + 2000
		fastResolvedExpr := `(
    EXISTS (
      SELECT 1
      FROM public.marking_trx_resolve mtr0
      WHERE mtr0.trx_type = 'provider'
        AND mtr0.trx_id = tp0.id
        AND mtr0.mark_type = 'provider_success_suspect'
    )
    OR EXISTS (
      SELECT 1
      FROM public.mutasi_bank mb0
      WHERE mb0.ref_id = '` + providerSuccessSuspectBankSettlementRefPrefix + `' || tp0.id::text
        AND mb0.alasan = '` + providerSuccessSuspectBankSettlementReason + `'
    )
    OR EXISTS (
      SELECT 1
      FROM public.transaksi_provider tp_b2
      WHERE tp_b2.transaksi_member_id = tp0.transaksi_member_id
        AND tp_b2.id <> tp0.id
        AND LOWER(TRIM(COALESCE(tp_b2.provider, ''))) = 'smb'
        AND LOWER(TRIM(COALESCE(tp_b2.status, ''))) = 'success'
        AND UPPER(TRIM(COALESCE(tp_b2.kode_produk, ''))) LIKE 'BIFASTOPEN2:%'
    )
  )`
		fastCacheFrom := `FROM public.provider_success_suspect_cache c
  JOIN public.transaksi_provider tp0 ON tp0.id = c.transaksi_provider_id
  LEFT JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id`
		fastCacheWheres := make([]string, 0, 4)
		if providerArgPos > 0 {
			fastCacheWheres = append(fastCacheWheres, fmt.Sprintf("LOWER(TRIM(tp0.provider)) = $%d", providerArgPos))
		}
		if refIDArgPos > 0 {
			fastCacheWheres = append(fastCacheWheres, fmt.Sprintf("(tp0.ref_id ILIKE '%%' || $%d || '%%' OR tm0.ref_id ILIKE '%%' || $%d || '%%')", refIDArgPos, refIDArgPos))
		}
		if fromArgPos > 0 {
			fastCacheWheres = append(fastCacheWheres, fmt.Sprintf("c.candidate_at >= $%d", fromArgPos))
		}
		if toArgPos > 0 {
			fastCacheWheres = append(fastCacheWheres, fmt.Sprintf("c.candidate_at < $%d", toArgPos))
		}
		fastCacheOrderSQL := "c.candidate_at DESC, c.transaksi_provider_id DESC"
		switch resolveStatus {
		case "", "unresolved":
			fastCacheWheres = append(fastCacheWheres, "NOT "+fastResolvedExpr)
		case "resolved":
			fastCacheWheres = append(fastCacheWheres, fastResolvedExpr)
		case "all":
			fastCacheOrderSQL = fastResolvedExpr + " ASC, c.candidate_at DESC, c.transaksi_provider_id DESC"
		}
		fastCacheWhereSQL := ""
		if len(fastCacheWheres) > 0 {
			fastCacheWhereSQL = "\n  WHERE " + strings.Join(fastCacheWheres, "\n    AND ")
		}
		candidateCTE = fmt.Sprintf(`WITH candidate_ids AS (
  SELECT c.transaksi_provider_id AS id
  %s%s
  ORDER BY %s
  LIMIT %d
)`, fastCacheFrom, fastCacheWhereSQL, fastCacheOrderSQL, prefetchLimit)
	}
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, queryLimit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	rankedSelectSQL := fmt.Sprintf(`
SELECT
  tp.id AS transaksi_provider_id,
  tp.transaksi_member_id,
  tm.member_id,
  m.nama AS member_nama,
  tm.status AS status_member,
  tm.ref_id AS ref_id_member,
  tm.kode_produk AS produk_member,
  tm.tujuan AS tujuan_member,
  tm.qty AS qty_member,
  tp.provider,
  COALESCE(tp.status, 'success') AS status_provider,
  tp.ref_id AS ref_id_provider,
  tp.perintah,
  tp.kode_produk,
  tp.tujuan,
  tp.qty,
  COALESCE(NULLIF(tp.qty, 0), NULLIF(tm.qty_provider, 0), NULLIF(tm.qty, 0), 0) AS nominal_provider_request,
  COALESCE(failed_anomaly.harga, tp.harga, prior_suspect.harga) AS harga,
  COALESCE(NULLIF(BTRIM(failed_anomaly.kode_respon), ''), tp.kode_respon, prior_suspect.kode_respon) AS kode_respon,
  COALESCE(NULLIF(BTRIM(failed_anomaly.pesan), ''), NULLIF(BTRIM(tp.pesan), ''), NULLIF(BTRIM(prior_suspect.pesan), ''), 'Member success dan semua provider gagal') AS pesan,
  tp.no_referensi,
  COALESCE(failed_anomaly.dibuat_pada, tp.dibuat_pada) AS provider_dibuat_pada,
  `+resolvedExpr+` AS resolved,
  COALESCE(mtr.resolved_at, direct_settlement.resolved_at, auto_b2.resolved_at) AS resolved_at,
  COALESCE(mtr.user_id, direct_settlement.user_id) AS resolved_by_user_id,
  COALESCE(resolved_user.nama, auto_b2.resolved_by_name) AS resolved_by_name,
  COALESCE(NULLIF(mtr.note, ''), NULLIF(direct_settlement.note, ''), auto_b2.resolve_note) AS resolve_note,
  (
    NOT `+resolvedExpr+`
    AND UPPER(TRIM(COALESCE(tp.perintah, ''))) = 'PAY'
    AND `+strings.ReplaceAll(providerSuccessSuspectRefundMessageSQL, "%", "%%")+`
    AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
    AND COALESCE(retry_guard.other_success_count, 0) = 0
    AND COALESCE(retry_guard.other_pending_count, 0) = 0
  ) AS can_retry_refund,
  ROW_NUMBER() OVER (
    PARTITION BY COALESCE(NULLIF(BTRIM(tp.ref_id), ''), NULLIF(BTRIM(tm.ref_id), ''), tp.id::text)
    ORDER BY (`+resolvedExpr+`) DESC, COALESCE(failed_anomaly.dibuat_pada, tp.dibuat_pada) DESC, tp.id DESC
  ) AS ref_rank
%s
WHERE %s`, baseFrom, whereSQL)

	qList := fmt.Sprintf(`
%s,
ranked_suspects AS (
%s
)
SELECT
  transaksi_provider_id,
  transaksi_member_id,
  member_id,
  member_nama,
  status_member,
  ref_id_member,
  produk_member,
  tujuan_member,
  qty_member,
  provider,
  status_provider,
  ref_id_provider,
  perintah,
  kode_produk,
  tujuan,
  qty,
  nominal_provider_request,
  harga,
  kode_respon,
  pesan,
  no_referensi,
  provider_dibuat_pada,
  resolved,
  resolved_at,
  resolved_by_user_id,
  resolved_by_name,
  resolve_note,
  can_retry_refund
FROM ranked_suspects
WHERE ref_rank = 1
ORDER BY resolved ASC, provider_dibuat_pada DESC, transaksi_provider_id DESC
LIMIT $%d OFFSET $%d
`, candidateCTE, rankedSelectSQL, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, listArgs...)
	if err != nil {
		return nil, 0, false, err
	}
	defer rows.Close()

	out := make([]AdminProviderSuccessSuspiciousMessageRow, 0, queryLimit)
	for rows.Next() {
		var (
			x           AdminProviderSuccessSuspiciousMessageRow
			memberID    sql.NullInt64
			memberNama  sql.NullString
			statusMem   sql.NullString
			refMem      sql.NullString
			produkMem   sql.NullString
			tujuanMem   sql.NullString
			qtyMem      sql.NullInt64
			nominalReq  sql.NullInt64
			harga       sql.NullInt64
			kodeRespon  sql.NullString
			pesan       sql.NullString
			noReferensi sql.NullString
			resolvedAt  sql.NullTime
			resolvedBy  sql.NullInt64
			resolvedNm  sql.NullString
			resolveNote sql.NullString
		)
		if err := rows.Scan(
			&x.TransaksiProviderID, &x.TransaksiMemberID, &memberID, &memberNama, &statusMem, &refMem,
			&produkMem, &tujuanMem, &qtyMem, &x.Provider, &x.StatusProvider, &x.RefIDProvider, &x.Perintah,
			&x.KodeProduk, &x.Tujuan, &x.Qty, &nominalReq, &harga, &kodeRespon, &pesan, &noReferensi, &x.ProviderDibuatPada,
			&x.Resolved, &resolvedAt, &resolvedBy, &resolvedNm, &resolveNote, &x.CanRetryRefund,
		); err != nil {
			return nil, 0, false, err
		}
		x.KodeProduk = normalizeProviderProductCodeDisplay(x.Provider, x.KodeProduk)
		if memberID.Valid {
			v := memberID.Int64
			x.MemberID = &v
		}
		if memberNama.Valid {
			v := strings.TrimSpace(memberNama.String)
			x.MemberNama = &v
		}
		if statusMem.Valid {
			v := strings.TrimSpace(statusMem.String)
			x.StatusMember = &v
		}
		if refMem.Valid {
			v := strings.TrimSpace(refMem.String)
			x.RefIDMember = &v
		}
		if produkMem.Valid {
			v := strings.TrimSpace(produkMem.String)
			x.ProdukMember = &v
		}
		if tujuanMem.Valid {
			v := strings.TrimSpace(tujuanMem.String)
			x.TujuanMember = &v
		}
		if qtyMem.Valid {
			v := qtyMem.Int64
			x.QtyMember = &v
		}
		if nominalReq.Valid && nominalReq.Int64 > 0 {
			v := nominalReq.Int64
			x.NominalProviderRequest = &v
		}
		if harga.Valid {
			v := harga.Int64
			x.Harga = &v
		}
		if kodeRespon.Valid {
			v := strings.TrimSpace(kodeRespon.String)
			x.KodeRespon = &v
		}
		if pesan.Valid {
			x.Pesan = strings.TrimSpace(pesan.String)
		}
		if noReferensi.Valid {
			v := strings.TrimSpace(noReferensi.String)
			x.NoReferensi = &v
		}
		if resolvedAt.Valid {
			v := resolvedAt.Time
			x.ResolvedAt = &v
		}
		if resolvedBy.Valid {
			v := resolvedBy.Int64
			x.ResolvedByUserID = &v
		}
		if resolvedNm.Valid {
			v := strings.TrimSpace(resolvedNm.String)
			x.ResolvedByName = &v
		}
		if resolveNote.Valid {
			v := strings.TrimSpace(resolveNote.String)
			x.ResolveNote = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, err
	}

	hasNext := false
	if len(out) > limit {
		hasNext = true
		out = out[:limit]
	}
	if includeTotal {
		qCount := fmt.Sprintf(`%s,
ranked_suspects AS (
%s
)
SELECT COUNT(1)
FROM ranked_suspects
WHERE ref_rank = 1`, candidateCTE, rankedSelectSQL)
		var total int64
		if err := r.db.QueryRowContext(ctx, qCount, args...).Scan(&total); err != nil {
			return nil, 0, false, err
		}
		return out, total, int64(offset+len(out)) < total, nil
	}
	total := int64(offset + len(out))
	if hasNext {
		total++
	}
	return out, total, hasNext, nil
}

func (r *AuditRepository) ResolveProviderSuccessSuspiciousMessage(
	ctx context.Context,
	userID int64,
	transaksiProviderID int64,
	note string,
) (*MarkingTrxResolveResult, error) {
	if err := r.ensureMarkingTrxResolveTable(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, fmt.Errorf("user_id invalid")
	}
	if transaksiProviderID <= 0 {
		return nil, fmt.Errorf("transaksi_provider_id harus > 0")
	}

	var exists int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.transaksi_provider tp
LEFT JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
LEFT JOIN LATERAL (
  SELECT
    a.id,
    a.kode_respon,
    a.pesan,
    a.harga,
    a.dibuat_pada
  FROM public.transaksi_anomasi_provider a
  WHERE LOWER(TRIM(COALESCE(a.provider, ''))) = LOWER(TRIM(COALESCE(tp.provider, '')))
    AND a.ref_id = tp.ref_id
    AND a.dibuat_pada >= tp.dibuat_pada
    AND `+providerSuccessSuspectFailedAnomalyRecordSQL+`
  ORDER BY a.id DESC
  LIMIT 1
) failed_anomaly ON true
LEFT JOIN LATERAL (
  SELECT
    tp2.id,
    tp2.kode_respon,
    tp2.pesan,
    tp2.harga,
    tp2.dibuat_pada
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id < tp.id
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tp.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
    AND `+strings.ReplaceAll(providerSuccessSuspectTextMessageSQL, "tp.", "tp2.")+`
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp.transaksi_member_id
        AND newer.id > tp.id
    )
  ORDER BY tp2.id DESC
  LIMIT 1
) prior_suspect ON true
LEFT JOIN LATERAL (
  SELECT tp2.id
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id > tp.id
    AND LOWER(TRIM(COALESCE(tp2.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'failed'
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp2.transaksi_member_id
        AND newer.id > tp2.id
    )
  ORDER BY tp2.id DESC
  LIMIT 1
) later_failed_loket ON true
`+providerSuccessSuspectSameRefResolvedLateralSQL+`
WHERE tp.id = $1
  AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
  AND `+providerSuccessSuspectEvidenceSQL+`
  AND `+providerSuccessSuspectEligibleWithAnomalySQL+`
  AND (prior_suspect.id IS NOT NULL OR later_failed_loket.id IS NULL)
  AND same_ref_resolved.provider_row_id IS NULL
`, transaksiProviderID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("transaksi provider suspect tidak ditemukan")
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.marking_trx_resolve
  (trx_type, trx_id, mark_type, user_id, note)
VALUES
  ('provider', $1, 'provider_success_suspect', $2, $3)
ON CONFLICT (trx_type, trx_id, mark_type)
DO UPDATE
SET user_id = EXCLUDED.user_id,
    note = EXCLUDED.note,
    resolved_at = now(),
    updated_at = now()
`, transaksiProviderID, userID, strings.TrimSpace(note))
	if err != nil {
		return nil, err
	}
	return &MarkingTrxResolveResult{Resolved: true, TrxID: transaksiProviderID, UserID: userID}, nil
}

func (r *AuditRepository) SettleProviderSuccessSuspiciousMessageWithBankDebit(
	ctx context.Context,
	userID int64,
	transaksiProviderID int64,
	nominal int64,
	fee int64,
	note string,
) (*ProviderSuspectBankSettlementResult, error) {
	if err := r.ensureMarkingTrxResolveTable(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, fmt.Errorf("user_id invalid")
	}
	if transaksiProviderID <= 0 {
		return nil, fmt.Errorf("transaksi_provider_id harus > 0")
	}
	if fee < 0 {
		return nil, fmt.Errorf("fee tidak boleh minus")
	}
	note = strings.TrimSpace(note)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	lockKey := fmt.Sprintf("provider-success-suspect-bank-settlement:%d", transaksiProviderID)
	if _, err := tx.ExecContext(ctx, `
SELECT pg_advisory_xact_lock(hashtextextended($1::text, 1006202601))
`, lockKey); err != nil {
		return nil, err
	}

	var (
		transaksiMemberID int64
		memberID          sql.NullInt64
		provider          string
		providerRefID     string
		transferAmount    int64
	)
	if err := tx.QueryRowContext(ctx, `
SELECT
  tp.transaksi_member_id,
  tm.member_id,
  COALESCE(tp.provider, ''),
  COALESCE(tp.ref_id, ''),
  COALESCE(NULLIF(tp.qty, 0), NULLIF(tm.qty_provider, 0), NULLIF(tm.qty, 0), 0)
FROM public.transaksi_provider tp
LEFT JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
LEFT JOIN LATERAL (
  SELECT
    a.id,
    a.kode_respon,
    a.pesan,
    a.harga,
    a.dibuat_pada
  FROM public.transaksi_anomasi_provider a
  WHERE LOWER(TRIM(COALESCE(a.provider, ''))) = LOWER(TRIM(COALESCE(tp.provider, '')))
    AND a.ref_id = tp.ref_id
    AND a.dibuat_pada >= tp.dibuat_pada
    AND `+providerSuccessSuspectFailedAnomalyRecordSQL+`
  ORDER BY a.id DESC
  LIMIT 1
) failed_anomaly ON true
LEFT JOIN LATERAL (
  SELECT
    tp2.id,
    tp2.kode_respon,
    tp2.pesan,
    tp2.harga,
    tp2.dibuat_pada
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id < tp.id
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tp.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp.status, ''))) = 'failed'
    AND `+strings.ReplaceAll(providerSuccessSuspectTextMessageSQL, "tp.", "tp2.")+`
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp.transaksi_member_id
        AND newer.id > tp.id
    )
  ORDER BY tp2.id DESC
  LIMIT 1
) prior_suspect ON true
LEFT JOIN LATERAL (
  SELECT tp2.id
  FROM public.transaksi_provider tp2
  WHERE tp2.transaksi_member_id = tp.transaksi_member_id
    AND tp2.id > tp.id
    AND LOWER(TRIM(COALESCE(tp2.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp2.status, ''))) = 'failed'
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp2.transaksi_member_id
        AND newer.id > tp2.id
    )
  ORDER BY tp2.id DESC
  LIMIT 1
) later_failed_loket ON true
`+providerSuccessSuspectSameRefResolvedLateralSQL+`
WHERE tp.id = $1
  AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
  AND `+providerSuccessSuspectEvidenceSQL+`
  AND `+providerSuccessSuspectEligibleWithAnomalySQL+`
  AND (prior_suspect.id IS NOT NULL OR later_failed_loket.id IS NULL)
  AND same_ref_resolved.provider_row_id IS NULL
LIMIT 1
`, transaksiProviderID).Scan(&transaksiMemberID, &memberID, &provider, &providerRefID, &transferAmount); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaksi provider suspect tidak ditemukan")
		}
		return nil, err
	}
	provider = strings.TrimSpace(provider)
	providerRefID = strings.TrimSpace(providerRefID)
	if transferAmount <= 0 {
		return nil, fmt.Errorf("nominal request provider tidak valid")
	}
	if nominal > 0 && nominal != transferAmount {
		return nil, fmt.Errorf("nominal debit tidak sesuai dengan nominal transaksi provider")
	}
	nominal = transferAmount
	totalAmount := nominal + fee
	if totalAmount <= 0 || totalAmount < nominal {
		return nil, fmt.Errorf("total debit tidak valid")
	}

	var (
		bankID      int64
		bankName    string
		bankAccount string
		bankOwner   string
		bankBefore  int64
	)
	if err := tx.QueryRowContext(ctx, `
SELECT id, nama, nomor_rekening, atas_nama, saldo
FROM public.bank
WHERE regexp_replace(COALESCE(nomor_rekening, ''), '[^0-9]', '', 'g') = $1
  AND LOWER(TRIM(COALESCE(nama, ''))) LIKE '%bca%'
  AND LOWER(TRIM(COALESCE(nama, ''))) LIKE '%suspect%'
ORDER BY id ASC
LIMIT 1
FOR UPDATE
`, providerSuccessSuspectBankAccountNumber).Scan(&bankID, &bankName, &bankAccount, &bankOwner, &bankBefore); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rekening %s %s tidak ditemukan", providerSuccessSuspectBankLabel, providerSuccessSuspectBankAccountNumber)
		}
		return nil, err
	}

	refID := fmt.Sprintf("%s%d", providerSuccessSuspectBankSettlementRefPrefix, transaksiProviderID)
	var (
		existingMutationID int64
		existingAmount     int64
		existingBefore     int64
		existingAfter      int64
	)
	existingErr := tx.QueryRowContext(ctx, `
SELECT id, jumlah, saldo_sebelum, saldo_sesudah
FROM public.mutasi_bank
WHERE bank_id = $1
  AND ref_id = $2
  AND alasan = $3
ORDER BY id DESC
LIMIT 1
`, bankID, refID, providerSuccessSuspectBankSettlementReason).Scan(&existingMutationID, &existingAmount, &existingBefore, &existingAfter)
	if existingErr != nil && existingErr != sql.ErrNoRows {
		return nil, existingErr
	}

	resolveNote := providerSuspectBankSettlementNote(note, refID, nominal, fee, totalAmount)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.marking_trx_resolve
  (trx_type, trx_id, mark_type, user_id, note)
VALUES
  ('provider', $1, 'provider_success_suspect', $2, $3)
ON CONFLICT (trx_type, trx_id, mark_type)
DO UPDATE
SET user_id = EXCLUDED.user_id,
    note = EXCLUDED.note,
    resolved_at = now(),
    updated_at = now()
`, transaksiProviderID, userID, resolveNote); err != nil {
		return nil, err
	}

	if existingErr == nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &ProviderSuspectBankSettlementResult{
			Resolved:             true,
			AlreadySettled:       true,
			TransaksiProviderID:  transaksiProviderID,
			TransaksiMemberID:    transaksiMemberID,
			BankID:               bankID,
			BankNama:             strings.TrimSpace(bankName),
			BankNomorRekening:    strings.TrimSpace(bankAccount),
			Nominal:              nominal,
			Fee:                  fee,
			TotalAmount:          existingAmount,
			SaldoSebelum:         existingBefore,
			SaldoSesudah:         existingAfter,
			MutasiBankID:         existingMutationID,
			RefID:                refID,
			ProviderRefID:        providerRefID,
			TransferAmountSource: transferAmount,
		}, nil
	}

	if bankBefore < totalAmount {
		return nil, fmt.Errorf("saldo rekening %s tidak cukup: saldo %d, total debit %d", providerSuccessSuspectBankLabel, bankBefore, totalAmount)
	}
	bankAfter := bankBefore - totalAmount
	if _, err := tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, bankAfter); err != nil {
		return nil, err
	}

	var memberIDArg any
	if memberID.Valid {
		memberIDArg = memberID.Int64
	}
	metaJSON, _ := json.Marshal(map[string]any{
		"type":                   "transaksi_suspect_bank_settlement",
		"source":                 "admin_transaksi_suspect",
		"transaksi_provider_id":  transaksiProviderID,
		"transaksi_member_id":    transaksiMemberID,
		"provider":               provider,
		"provider_ref_id":        providerRefID,
		"nominal":                nominal,
		"fee":                    fee,
		"total_amount":           totalAmount,
		"transfer_amount_source": transferAmount,
		"bank_account_number":    strings.TrimSpace(bankAccount),
		"bank_account_name":      strings.TrimSpace(bankOwner),
		"admin_note":             note,
	})
	var mutasiBankID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, member_id, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,$4,NULLIF($5,''),$6,$7,NULLIF($8,''),$9,$10,now(),$11::jsonb)
RETURNING id
`, bankID, refID, totalAmount, providerSuccessSuspectBankSettlementReason, resolveNote, bankBefore, bankAfter, provider, memberIDArg, userID, string(metaJSON)).Scan(&mutasiBankID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &ProviderSuspectBankSettlementResult{
		Resolved:             true,
		AlreadySettled:       false,
		TransaksiProviderID:  transaksiProviderID,
		TransaksiMemberID:    transaksiMemberID,
		BankID:               bankID,
		BankNama:             strings.TrimSpace(bankName),
		BankNomorRekening:    strings.TrimSpace(bankAccount),
		Nominal:              nominal,
		Fee:                  fee,
		TotalAmount:          totalAmount,
		SaldoSebelum:         bankBefore,
		SaldoSesudah:         bankAfter,
		MutasiBankID:         mutasiBankID,
		RefID:                refID,
		ProviderRefID:        providerRefID,
		TransferAmountSource: transferAmount,
	}, nil
}

func providerSuspectBankSettlementNote(note string, refID string, nominal int64, fee int64, totalAmount int64) string {
	base := fmt.Sprintf("Diselesaikan via debit %s ref=%s nominal=%d fee=%d total=%d", providerSuccessSuspectBankLabel, refID, nominal, fee, totalAmount)
	note = strings.TrimSpace(note)
	if note == "" {
		return base
	}
	return base + " | " + note
}

func (r *AuditRepository) ensureMarkingTrxResolveTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS public.marking_trx_resolve (
  id BIGSERIAL PRIMARY KEY,
  trx_type TEXT NOT NULL,
  trx_id BIGINT NOT NULL,
  mark_type TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES public.member(id),
  note TEXT NOT NULL DEFAULT '',
  resolved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (trx_type, trx_id, mark_type)
)`)
	return err
}

func (r *AuditRepository) GetProviderSuccessSuspiciousResendTarget(ctx context.Context, transaksiProviderID int64) (*AdminProviderSuccessSuspiciousResendTarget, error) {
	if transaksiProviderID <= 0 {
		return nil, fmt.Errorf("transaksi_provider_id harus > 0")
	}
	var (
		out          AdminProviderSuccessSuspiciousResendTarget
		produkMember sql.NullString
		mapID        sql.NullInt64
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  tp.id,
  tp.transaksi_member_id,
  tp.ref_id,
  COALESCE(tm.kode_produk, ''),
  tp.tujuan,
  tp.qty,
  tp.kode_produk,
  tp.provider,
  COALESCE(tp.status, ''),
  COALESCE(tp.pesan, ''),
  tp.produk_provider_map_id
FROM public.transaksi_provider tp
JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
WHERE tp.id = $1
  AND LOWER(TRIM(tp.provider)) = 'smb'
  AND COALESCE(BTRIM(tp.pesan), '') <> ''
  AND `+providerSuccessSuspectRefundMessageSQL+`
  AND `+providerSuccessSuspectEligibleSQL+`
LIMIT 1
`, transaksiProviderID).Scan(
		&out.TransaksiProviderID,
		&out.TransaksiMemberID,
		&out.RefID,
		&produkMember,
		&out.Tujuan,
		&out.Qty,
		&out.KodeProdukProvider,
		&out.Provider,
		&out.StatusProvider,
		&out.Pesan,
		&mapID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaksi provider suspect SMB tidak ditemukan")
		}
		return nil, err
	}
	out.ProdukMember = strings.ToUpper(strings.TrimSpace(produkMember.String))
	out.Provider = strings.ToLower(strings.TrimSpace(out.Provider))
	out.StatusProvider = strings.ToLower(strings.TrimSpace(out.StatusProvider))
	out.KodeProdukProvider = strings.ToUpper(strings.TrimSpace(out.KodeProdukProvider))
	out.Pesan = strings.TrimSpace(out.Pesan)
	out.RefID = strings.TrimSpace(out.RefID)
	out.Tujuan = strings.TrimSpace(out.Tujuan)
	if mapID.Valid {
		v := mapID.Int64
		out.ProdukProviderMapID = &v
	}
	if out.Provider != "smb" {
		return nil, fmt.Errorf("hanya suspect SMB yang bisa dikirim ulang")
	}
	if !strings.HasPrefix(out.KodeProdukProvider, "BIFASTOPEN") {
		return nil, fmt.Errorf("hanya row BIFASTOPEN yang bisa dikirim ulang ke BIFASTOPEN2")
	}
	if strings.HasPrefix(out.KodeProdukProvider, "BIFASTOPEN2") {
		return nil, fmt.Errorf("row ini sudah berada di jalur BIFASTOPEN2")
	}
	if out.ProdukMember == "" || out.RefID == "" || out.Tujuan == "" || out.Qty <= 0 {
		return nil, fmt.Errorf("data transaksi suspect tidak lengkap untuk kirim ulang")
	}
	return &out, nil
}

func (r *AuditRepository) FindBackupSMBMapForInternalSKU(ctx context.Context, internalSKU string, nominal int64) (*int64, string, error) {
	internalSKU = strings.ToUpper(strings.TrimSpace(internalSKU))
	if internalSKU == "" {
		return nil, "", fmt.Errorf("produk internal kosong")
	}
	var (
		id           int64
		kodeProvider string
	)
	err := r.db.QueryRowContext(ctx, `
SELECT ppm.id, ppm.kode_provider
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
JOIN public.provider pr ON LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
WHERE UPPER(TRIM(p.sku)) = $1
  AND LOWER(TRIM(ppm.provider)) = 'smb'
  AND pr.aktif = true
  AND ppm.aktif = true
  AND UPPER(TRIM(COALESCE(ppm.special_code, ''))) = 'BIFASTOPEN2'
  AND UPPER(TRIM(COALESCE(ppm.mode, ''))) = 'DIRECT'
  AND ($2 <= 0 OR (ppm.minimal_nominal IS NULL OR ppm.minimal_nominal <= $2))
  AND ($2 <= 0 OR (ppm.maksimal_nominal IS NULL OR ppm.maksimal_nominal >= $2))
  AND (
    ppm.jam_buka IS NULL OR ppm.jam_tutup IS NULL
    OR (CURRENT_TIME AT TIME ZONE 'Asia/Jakarta')::time BETWEEN ppm.jam_buka AND ppm.jam_tutup
  )
ORDER BY ppm.id DESC
LIMIT 1
`, internalSKU, nominal).Scan(&id, &kodeProvider)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("mapping SMB bank tidak ditemukan untuk produk ini")
		}
		return nil, "", err
	}
	kodeProvider = strings.ToUpper(strings.TrimSpace(kodeProvider))
	if kodeProvider == "" {
		return nil, "", fmt.Errorf("kode bank SMB kosong")
	}
	product := "BIFASTOPEN2:" + kodeProvider
	return nil, product, nil
}
