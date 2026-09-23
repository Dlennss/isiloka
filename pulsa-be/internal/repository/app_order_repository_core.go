package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type AppOrderRepository struct {
	db *sql.DB
}

func NewAppOrderRepository(db *sql.DB) *AppOrderRepository {
	return &AppOrderRepository{db: db}
}

func (r *AppOrderRepository) Create(ctx context.Context, in AppOrderCreateInput) error {
	return r.db.QueryRowContext(ctx, `
INSERT INTO public.app_order
  (invoice_id, member_id, guest_nama, guest_email, guest_phone, produk_id, produk_sku_snapshot, produk_nama_snapshot, dest, qty, nominal, buyer_type, buyer_role, harga_dasar, fee, harga_final, fee_user_snapshot, fee_agent_snapshot, fee_master_snapshot, retail_agent_id_snapshot, retail_master_id_snapshot, status, catatan, alasan_gagal, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,now(),now())
RETURNING id
`, in.InvoiceID, nullableInt64(in.MemberID), nullableString(in.GuestNama), nullableString(in.GuestEmail), nullableString(in.GuestPhone), in.ProdukID, in.ProdukSKUSnapshot, in.ProdukNamaSnapshot, in.Dest, in.Qty, in.Nominal, in.BuyerType, in.BuyerRole, in.HargaDasar, in.Fee, in.HargaFinal, in.FeeUserSnapshot, in.FeeAgentSnapshot, in.FeeMasterSnapshot, nullableInt64(in.RetailAgentIDSnapshot), nullableInt64(in.RetailMasterIDSnapshot), in.Status, nullableString(in.Catatan), nullableString(in.AlasanGagal)).Scan(&in.ID)
}

func (r *AppOrderRepository) GetRetailMemberContext(ctx context.Context, memberID int64) (*RetailMemberContextRow, error) {
	if memberID <= 0 {
		return nil, sql.ErrNoRows
	}

	var (
		row            RetailMemberContextRow
		retailAgentID  sql.NullInt64
		retailMasterID sql.NullInt64
		h2hAgentID     sql.NullInt64
		h2hMasterID    sql.NullInt64
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  m.id,
  m.email,
  COALESCE(m.nama, ''),
  COALESCE(m.role, ''),
  m.aktif,
  COALESCE(d.saldo, 0),
  m.retail_agent_id,
  m.retail_master_id,
  m.h2h_agent_member_id,
  m.h2h_master_member_id
FROM public.member m
LEFT JOIN public.dompet_member d ON d.member_id = m.id
WHERE m.id = $1
LIMIT 1
`, memberID).Scan(
		&row.MemberID,
		&row.Email,
		&row.Nama,
		&row.Role,
		&row.Aktif,
		&row.Saldo,
		&retailAgentID,
		&retailMasterID,
		&h2hAgentID,
		&h2hMasterID,
	)
	if err != nil {
		return nil, err
	}
	if retailAgentID.Valid {
		v := retailAgentID.Int64
		row.RetailAgentID = &v
	}
	if retailMasterID.Valid {
		v := retailMasterID.Int64
		row.RetailMasterID = &v
	}
	if h2hAgentID.Valid {
		v := h2hAgentID.Int64
		row.H2HAgentID = &v
	}
	if h2hMasterID.Valid {
		v := h2hMasterID.Int64
		row.H2HMasterID = &v
	}
	return &row, nil
}

func nullableString(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func normalizePhone(v string) string {
	v = strings.TrimSpace(v)
	var b strings.Builder
	for _, r := range v {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func shouldApplyAppOrderStatusTransition(current, next string) bool {
	current = strings.TrimSpace(strings.ToLower(current))
	next = strings.TrimSpace(strings.ToLower(next))
	if current == "" || next == "" {
		return false
	}
	if current == next {
		return true
	}

	switch current {
	case "success":
		return next == "refunded"
	case "refunded":
		return false
	case "processing_provider":
		switch next {
		case "processing_provider", "success", "failed", "refunded":
			return true
		default:
			return false
		}
	case "failed", "expired", "cancelled":
		if current == "failed" && next == "refunded" {
			return true
		}
		return false
	case "paid":
		switch next {
		case "paid", "processing_provider", "success", "failed", "refunded":
			return true
		default:
			return false
		}
	case "pending_payment":
		switch next {
		case "paid", "failed", "expired", "cancelled":
			return true
		default:
			return false
		}
	default:
		return true
	}
}

func (r *AppOrderRepository) UpdateStatusByInvoiceID(ctx context.Context, invoiceID, status string) error {
	invoiceID = strings.TrimSpace(invoiceID)
	status = strings.TrimSpace(status)

	var current string
	err := r.db.QueryRowContext(ctx, `
SELECT status
FROM public.app_order
WHERE invoice_id = $1
LIMIT 1
`, invoiceID).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}
	if !shouldApplyAppOrderStatusTransition(current, status) {
		return nil
	}

	res, err := r.db.ExecContext(ctx, `
UPDATE public.app_order
SET status = $2,
    diubah_pada = now()
WHERE invoice_id = $1
`, invoiceID, status)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *AppOrderRepository) UpdateStatusByID(ctx context.Context, id int64, status string) error {
	status = strings.TrimSpace(status)

	var current string
	err := r.db.QueryRowContext(ctx, `
SELECT status
FROM public.app_order
WHERE id = $1
LIMIT 1
`, id).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}
	if !shouldApplyAppOrderStatusTransition(current, status) {
		return nil
	}

	res, err := r.db.ExecContext(ctx, `
UPDATE public.app_order
SET status = $2,
    diubah_pada = now()
WHERE id = $1
`, id, status)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}
