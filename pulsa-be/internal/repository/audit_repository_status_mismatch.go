package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

func (r *AuditRepository) AdminListStatusMismatch(
	ctx context.Context,
	limit, offset int,
	provider, statusMember, mismatchType, fromStr, toStr string,
) ([]AdminStatusMismatchRow, int64, error) {
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
		args   []any
		wheres []string
	)

	provider = strings.TrimSpace(strings.ToLower(provider))

	statusMember = strings.TrimSpace(strings.ToLower(statusMember))
	if statusMember != "" {
		args = append(args, statusMember)
		wheres = append(wheres, fmt.Sprintf("LOWER(TRIM(tm.status)) = $%d", len(args)))
	}

	wheres = append(wheres, "NOT ((LOWER(TRIM(tm.status)) = 'failed' AND LOWER(COALESCE(tm.keterangan, '')) LIKE '%dibatalkan admin%') OR (LOWER(TRIM(tm.status)) = 'success' AND LOWER(COALESCE(tm.keterangan, '')) LIKE '%diselesaikan admin%'))")

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada < $%d", len(args)))
	}

	switch strings.TrimSpace(strings.ToLower(mismatchType)) {
	case "member_success_provider_not20", "member_not_success_provider_20", "":
	default:
		return nil, 0, fmt.Errorf("invalid mismatch_type: %s", mismatchType)
	}

	whereSQL := "1=1"
	if len(wheres) > 0 {
		whereSQL = strings.Join(wheres, " AND ")
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  tm.id AS transaksi_member_id,
  tm.ref_id,
  tm.status AS status_member,
  tm.dibuat_pada AS member_created,
  tp.id AS transaksi_provider_id,
  tp.provider,
  tp.kode_respon,
  tp.pesan,
  tp.dibuat_pada AS provider_created
FROM public.transaksi_member tm
JOIN public.transaksi_provider tp ON tp.transaksi_member_id = tm.id
WHERE %s
ORDER BY tm.dibuat_pada DESC, tm.id DESC, tp.id DESC
`, whereSQL), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	type auditProviderAttempt struct {
		id              int64
		provider        string
		kodeRespon      sql.NullString
		pesan           sql.NullString
		providerCreated time.Time
	}
	type auditMismatchGroup struct {
		row      AdminStatusMismatchRow
		attempts []auditProviderAttempt
	}

	finalizeGroup := func(group *auditMismatchGroup) (AdminStatusMismatchRow, bool) {
		if group == nil {
			return AdminStatusMismatchRow{}, false
		}
		var (
			bestSuccess *auditProviderAttempt
			bestPending *auditProviderAttempt
			bestFailed  *auditProviderAttempt
			bestUnknown *auditProviderAttempt
		)
		for i := range group.attempts {
			attempt := group.attempts[i]
			rc := ""
			msg := ""
			if attempt.kodeRespon.Valid {
				rc = strings.TrimSpace(attempt.kodeRespon.String)
			}
			if attempt.pesan.Valid {
				msg = strings.TrimSpace(attempt.pesan.String)
			}
			switch helper.ProviderResponseStateOf(attempt.provider, rc, msg) {
			case helper.ProviderResponseSuccess:
				if bestSuccess == nil || attempt.id > bestSuccess.id {
					a := attempt
					bestSuccess = &a
				}
			case helper.ProviderResponsePending:
				if bestPending == nil || attempt.id > bestPending.id {
					a := attempt
					bestPending = &a
				}
			case helper.ProviderResponseFailed:
				if bestFailed == nil || attempt.id > bestFailed.id {
					a := attempt
					bestFailed = &a
				}
			default:
				if bestUnknown == nil || attempt.id > bestUnknown.id {
					a := attempt
					bestUnknown = &a
				}
			}
		}

		var picked *auditProviderAttempt
		switch {
		case bestSuccess != nil:
			picked = bestSuccess
		case bestPending != nil:
			picked = bestPending
		case bestFailed != nil:
			picked = bestFailed
		default:
			picked = bestUnknown
		}
		if picked == nil {
			return AdminStatusMismatchRow{}, false
		}

		rc := ""
		msg := ""
		if picked.kodeRespon.Valid {
			rc = strings.TrimSpace(picked.kodeRespon.String)
		}
		if picked.pesan.Valid {
			msg = strings.TrimSpace(picked.pesan.String)
		}
		state := helper.ProviderResponseStateOf(picked.provider, rc, msg)
		memberSuccess := strings.EqualFold(strings.TrimSpace(group.row.StatusMember), "success")

		mismatch := false
		switch strings.TrimSpace(strings.ToLower(mismatchType)) {
		case "member_success_provider_not20":
			mismatch = memberSuccess && state != helper.ProviderResponseSuccess
		case "member_not_success_provider_20":
			mismatch = !memberSuccess && state == helper.ProviderResponseSuccess
		default:
			mismatch = (memberSuccess && state != helper.ProviderResponseSuccess) || (!memberSuccess && state == helper.ProviderResponseSuccess)
		}
		if !mismatch {
			return AdminStatusMismatchRow{}, false
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(picked.provider)) != provider {
			return AdminStatusMismatchRow{}, false
		}

		row := group.row
		row.TransaksiProviderID = picked.id
		row.Provider = strings.ToLower(strings.TrimSpace(picked.provider))
		v := string(state)
		row.StatusProvider = &v
		row.ProviderCreated = picked.providerCreated
		return row, true
	}

	matches := make([]AdminStatusMismatchRow, 0, limit)
	var (
		current *auditMismatchGroup
		total   int64
	)

	for rows.Next() {
		var (
			memberID        int64
			refID           string
			memberStatus    string
			memberCreated   time.Time
			providerID      int64
			providerName    string
			kodeRespon      sql.NullString
			pesan           sql.NullString
			providerCreated time.Time
		)
		if err := rows.Scan(
			&memberID, &refID, &memberStatus, &memberCreated, &providerID, &providerName, &kodeRespon, &pesan, &providerCreated,
		); err != nil {
			return nil, 0, err
		}

		if current != nil && current.row.TransaksiMemberID != memberID {
			if row, ok := finalizeGroup(current); ok {
				total++
				if total > int64(offset) && len(matches) < limit {
					matches = append(matches, row)
				}
			}
			current = nil
		}
		if current == nil {
			current = &auditMismatchGroup{
				row: AdminStatusMismatchRow{
					TransaksiMemberID: memberID,
					RefID:             refID,
					StatusMember:      memberStatus,
					MemberCreated:     memberCreated,
				},
				attempts: make([]auditProviderAttempt, 0, 4),
			}
		}
		current.attempts = append(current.attempts, auditProviderAttempt{
			id: providerID, provider: providerName, kodeRespon: kodeRespon, pesan: pesan, providerCreated: providerCreated,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if row, ok := finalizeGroup(current); ok {
		total++
		if total > int64(offset) && len(matches) < limit {
			matches = append(matches, row)
		}
	}

	return matches, total, nil
}
