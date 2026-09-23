package repository

import (
	"context"
	"database/sql"
)

type AppBillingCheckRepository struct {
	db *sql.DB
}

func NewAppBillingCheckRepository(db *sql.DB) *AppBillingCheckRepository {
	return &AppBillingCheckRepository{db: db}
}

func (r *AppBillingCheckRepository) Create(ctx context.Context, in AppBillingCheckCreateInput) error {
	return r.db.QueryRowContext(ctx, `
INSERT INTO public.app_billing_check
  (ref_id, member_id, guest_nama, guest_email, guest_phone, produk_id, produk_sku_snapshot, produk_nama_snapshot, dest, buyer_type, buyer_role, provider, harga_provider, status, kode_respon, pesan, sn, raw_request, raw_callback, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18::jsonb,$19::jsonb,now(),now())
RETURNING id
`, in.RefID, nullableInt64(in.MemberID), nullableString(in.GuestNama), nullableString(in.GuestEmail), nullableString(in.GuestPhone),
		in.ProdukID, in.ProdukSKUSnapshot, in.ProdukNamaSnapshot, in.Dest, in.BuyerType, in.BuyerRole, in.Provider, in.HargaProvider,
		in.Status, nullableString(in.KodeRespon), nullableString(in.Pesan), nullableString(in.SN), nullableJSON(in.RawRequest), nullableJSON(in.RawCallback)).Scan(&in.ID)
}

func (r *AppBillingCheckRepository) GetByRefID(ctx context.Context, refID string) (*AppBillingCheckRow, error) {
	var (
		row        AppBillingCheckRow
		memberNama sql.NullString
		guestNama  sql.NullString
		guestEmail sql.NullString
		guestPhone sql.NullString
		kodeRespon sql.NullString
		pesan      sql.NullString
		sn         sql.NullString
		dibuatPada sql.NullTime
		diubahPada sql.NullTime
		memberID   sql.NullInt64
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  abc.id, abc.ref_id, abc.member_id, m.nama, abc.guest_nama, abc.guest_email, abc.guest_phone,
  abc.produk_id, abc.produk_sku_snapshot, abc.produk_nama_snapshot, abc.dest, abc.buyer_type, abc.buyer_role,
  abc.provider, abc.harga_provider, abc.status, abc.kode_respon, abc.pesan, abc.sn, abc.dibuat_pada, abc.diubah_pada
FROM public.app_billing_check abc
LEFT JOIN public.member m ON m.id = abc.member_id
WHERE TRIM(abc.ref_id) = $1
LIMIT 1
`, refID).Scan(
		&row.ID, &row.RefID, &memberID, &memberNama, &guestNama, &guestEmail, &guestPhone,
		&row.ProdukID, &row.ProdukSKUSnapshot, &row.ProdukNamaSnapshot, &row.Dest, &row.BuyerType, &row.BuyerRole,
		&row.Provider, &row.HargaProvider, &row.Status, &kodeRespon, &pesan, &sn, &dibuatPada, &diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if memberID.Valid {
		v := memberID.Int64
		row.MemberID = &v
	}
	if memberNama.Valid {
		v := memberNama.String
		row.MemberNama = &v
	}
	if guestNama.Valid {
		v := guestNama.String
		row.GuestNama = &v
	}
	if guestEmail.Valid {
		v := guestEmail.String
		row.GuestEmail = &v
	}
	if guestPhone.Valid {
		v := guestPhone.String
		row.GuestPhone = &v
	}
	if kodeRespon.Valid {
		v := kodeRespon.String
		row.KodeRespon = &v
	}
	if pesan.Valid {
		v := pesan.String
		row.Pesan = &v
	}
	if sn.Valid {
		v := sn.String
		row.SN = &v
	}
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	if diubahPada.Valid {
		v := diubahPada.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *AppBillingCheckRepository) UpdateResult(ctx context.Context, in AppBillingCheckUpdateInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.app_billing_check
SET harga_provider = COALESCE($2, harga_provider),
    status = CASE
      WHEN status IN ('success', 'failed', 'refunded', 'expired', 'cancelled')
        AND COALESCE(NULLIF($3, ''), '') = 'processing_provider'
      THEN status
      ELSE COALESCE(NULLIF($3, ''), status)
    END,
    kode_respon = COALESCE($4, kode_respon),
    pesan = COALESCE($5, pesan),
    sn = COALESCE($6, sn),
    raw_callback = COALESCE($7::jsonb, raw_callback),
    diubah_pada = now()
WHERE id = $1
`, in.ID, in.HargaProvider, nullableStringValue(in.Status), nullableStringValuePtr(in.KodeRespon), nullableStringValuePtr(in.Pesan), nullableStringValuePtr(in.SN), nullableJSON(in.RawCallback))
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

func nullableStringValuePtr(v *string) any {
	if v == nil {
		return nil
	}
	return nullableStringValue(*v)
}
