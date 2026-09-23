package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type SMPAYSourceMarker struct {
	MemberID           int64
	TransaksiMemberID  int64
	RefID              string
	SMPAYTransactionID int64
	SMPAYWebsiteID     int64
	SMPAYDivisionID    int64
	SourceSystem       string
	SkipH2HCommission  bool
	RawRequestJSON     string
}

func normalizeSMPAYSourceMarker(m SMPAYSourceMarker) SMPAYSourceMarker {
	m.RefID = strings.TrimSpace(m.RefID)
	m.SourceSystem = strings.ToUpper(strings.TrimSpace(m.SourceSystem))
	if m.SourceSystem == "" {
		m.SourceSystem = "SMPAY"
	}
	if strings.TrimSpace(m.RawRequestJSON) == "" {
		m.RawRequestJSON = "{}"
	}
	return m
}

func (r *MemberTrxMemberRepository) UpsertSMPAYRefSource(ctx context.Context, marker SMPAYSourceMarker) error {
	marker = normalizeSMPAYSourceMarker(marker)
	if marker.MemberID <= 0 || marker.RefID == "" || marker.SourceSystem != "SMPAY" {
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.smpay_ref_sources
  (member_id, ref_id, smpay_transaction_id, smpay_website_id, smpay_division_id, source_system, skip_h2h_commission, raw_request, created_at)
VALUES
  ($1,$2,NULLIF($3,0),NULLIF($4,0),NULLIF($5,0),'SMPAY',$6,COALESCE(NULLIF($7,''),'{}')::jsonb,now())
ON CONFLICT (member_id, ref_id) DO UPDATE
SET smpay_transaction_id = COALESCE(EXCLUDED.smpay_transaction_id, public.smpay_ref_sources.smpay_transaction_id),
    smpay_website_id = COALESCE(EXCLUDED.smpay_website_id, public.smpay_ref_sources.smpay_website_id),
    smpay_division_id = COALESCE(EXCLUDED.smpay_division_id, public.smpay_ref_sources.smpay_division_id),
    source_system = 'SMPAY',
    skip_h2h_commission = EXCLUDED.skip_h2h_commission,
    raw_request = CASE
      WHEN EXCLUDED.raw_request = '{}'::jsonb THEN public.smpay_ref_sources.raw_request
      ELSE EXCLUDED.raw_request
    END
`, marker.MemberID, marker.RefID, marker.SMPAYTransactionID, marker.SMPAYWebsiteID, marker.SMPAYDivisionID, marker.SkipH2HCommission, marker.RawRequestJSON)
	return err
}

func (r *MemberTrxMemberRepository) AttachSMPAYTransactionSource(ctx context.Context, marker SMPAYSourceMarker) error {
	marker = normalizeSMPAYSourceMarker(marker)
	if marker.MemberID <= 0 || marker.TransaksiMemberID <= 0 || marker.RefID == "" || marker.SourceSystem != "SMPAY" {
		return nil
	}
	if err := r.UpsertSMPAYRefSource(ctx, marker); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.smpay_transaction_sources
  (transaksi_member_id, smpay_ref_source_id, ref_id, smpay_transaction_id, smpay_website_id, smpay_division_id, source_system, skip_h2h_commission, created_at)
SELECT
  $1, s.id, s.ref_id, s.smpay_transaction_id, s.smpay_website_id, s.smpay_division_id, 'SMPAY', s.skip_h2h_commission, now()
FROM public.smpay_ref_sources s
WHERE s.member_id = $2 AND s.ref_id = $3
ON CONFLICT (transaksi_member_id) DO UPDATE
SET smpay_ref_source_id = EXCLUDED.smpay_ref_source_id,
    ref_id = EXCLUDED.ref_id,
    smpay_transaction_id = COALESCE(EXCLUDED.smpay_transaction_id, public.smpay_transaction_sources.smpay_transaction_id),
    smpay_website_id = COALESCE(EXCLUDED.smpay_website_id, public.smpay_transaction_sources.smpay_website_id),
    smpay_division_id = COALESCE(EXCLUDED.smpay_division_id, public.smpay_transaction_sources.smpay_division_id),
    source_system = 'SMPAY',
    skip_h2h_commission = EXCLUDED.skip_h2h_commission
`, marker.TransaksiMemberID, marker.MemberID, marker.RefID)
	return err
}

func lookupSMPAYSkipH2HCommission(ctx context.Context, tx *sql.Tx, trxID, memberID int64, refID string) (bool, error) {
	refID = strings.TrimSpace(refID)
	if tx == nil || trxID <= 0 || memberID <= 0 || refID == "" {
		return false, nil
	}

	var skip bool
	err := tx.QueryRowContext(ctx, `
SELECT COALESCE(skip_h2h_commission, false)
FROM public.smpay_transaction_sources
WHERE transaksi_member_id = $1
LIMIT 1
`, trxID).Scan(&skip)
	if err == nil {
		return skip, nil
	}
	if isUndefinedTableErr(err) {
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	var (
		sourceID           int64
		smpayTransactionID sql.NullInt64
		smpayWebsiteID     sql.NullInt64
		smpayDivisionID    sql.NullInt64
	)
	err = tx.QueryRowContext(ctx, `
SELECT id, smpay_transaction_id, smpay_website_id, smpay_division_id, COALESCE(skip_h2h_commission, false)
FROM public.smpay_ref_sources
WHERE member_id = $1 AND ref_id = $2
LIMIT 1
`, memberID, refID).Scan(&sourceID, &smpayTransactionID, &smpayWebsiteID, &smpayDivisionID, &skip)
	if isUndefinedTableErr(err) {
		return false, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.smpay_transaction_sources
  (transaksi_member_id, smpay_ref_source_id, ref_id, smpay_transaction_id, smpay_website_id, smpay_division_id, source_system, skip_h2h_commission, created_at)
VALUES
  ($1,$2,$3,$4,$5,$6,'SMPAY',$7,now())
ON CONFLICT (transaksi_member_id) DO NOTHING
`, trxID, sourceID, refID, smpayTransactionID, smpayWebsiteID, smpayDivisionID, skip)
	if err != nil {
		return false, err
	}
	return skip, nil
}

func isUndefinedTableErr(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "42P01"
}
