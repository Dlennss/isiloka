package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type MemberTrxMemberRepository struct {
	db *sql.DB
}

func NewMemberTrxMemberRepository(db *sql.DB) *MemberTrxMemberRepository {
	return &MemberTrxMemberRepository{db: db}
}

func (r *MemberTrxMemberRepository) DB() *sql.DB {
	return r.db
}

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInsufficientBalance = errors.New("saldo tidak cukup")
)

type MemberAuth struct {
	MemberID       int64
	Email          sql.NullString
	Nama           sql.NullString
	Aktif          bool
	Role           string
	ChargeReceiver bool

	PasswordHash sql.NullString
	PinHash      sql.NullString
}

type ProdukPricingRule struct {
	TipeHarga       string
	Nominal         *int64
	MaksimalNominal *int64
	JamBuka         string
	JamTutup        string
	Aktif           bool
	KategoriNama    string
	BrandNama       string
}

type TrxMemberFull struct {
	ID                    int64
	MemberID              int64
	RefID                 string
	Perintah              string
	KodeProduk            string
	Tujuan                string
	Qty                   int64
	QtyProvider           int64
	Status                string
	Keterangan            string
	ChargeReceiverApplied bool
	BiayaPerkiraan        int64
	BiayaAktual           int64
	FeeMemberRp           int64
	HargaJavapay          int64
	HargaMember           int64
}

func (r *MemberTrxMemberRepository) AuthByAPIKey(ctx context.Context, apiKey string) (*MemberAuth, error) {
	const q = `
SELECT m.id, m.email, m.nama, m.aktif, m.role, COALESCE(m.charge_receiver, false), m.password_hash, m.pin_hash
FROM public.member_api_key k
JOIN public.member m ON m.id = k.member_id
WHERE k.api_key = $1 AND k.aktif = true
LIMIT 1`
	var out MemberAuth
	err := r.db.QueryRowContext(ctx, q, apiKey).Scan(
		&out.MemberID, &out.Email, &out.Nama, &out.Aktif, &out.Role, &out.ChargeReceiver, &out.PasswordHash, &out.PinHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if !out.Aktif {
		return nil, ErrForbidden
	}
	return &out, nil
}

func (r *MemberTrxMemberRepository) CreateTransaksiMember(
	ctx context.Context,
	memberID int64,
	clientRefID string,
	perintah string,
	kodeProduk string,
	tujuan string,
	qty int64,
	qtyProvider int64,
	chargeReceiverApplied bool,
	biayaPerkiraan int64,
) (trxID int64, created bool, err error) {
	const q = `
	INSERT INTO public.transaksi_member
	  (member_id, ref_id, perintah, kode_produk, tujuan, qty, qty_provider, charge_receiver_applied, status, biaya_perkiraan, biaya_aktual, harga_member)
	VALUES
	  ($1,$2,$3,$4,$5,$6,$7,$8,'pending',$9,0,$9)
	ON CONFLICT (member_id, ref_id) DO UPDATE
	SET diperbarui_pada = now()
	RETURNING id, (xmax = 0) AS created`
	err = r.db.QueryRowContext(ctx, q,
		memberID, clientRefID, perintah, kodeProduk, tujuan, qty, qtyProvider, chargeReceiverApplied, biayaPerkiraan,
	).Scan(&trxID, &created)
	return trxID, created, err
}

func (r *MemberTrxMemberRepository) GetTransaksiMemberByRef(ctx context.Context, memberID int64, clientRefID string) (map[string]any, error) {
	const q = `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, COALESCE(qty_provider, qty), status, COALESCE(charge_receiver_applied, false), keterangan, dibuat_pada, diperbarui_pada, biaya_perkiraan, biaya_aktual
FROM public.transaksi_member
WHERE member_id = $1 AND ref_id = $2
LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, memberID, clientRefID)

	var (
		id, mid, qty, qtyProvider, biayaPerkiraan, biayaAktual int64
		refid, perintah, kodeProduk, tujuan, status            string
		chargeReceiverApplied                                  bool
		ket                                                    sql.NullString
		dibuat, diupd                                          time.Time
	)
	err := row.Scan(
		&id, &mid, &refid, &perintah, &kodeProduk, &tujuan, &qty, &qtyProvider, &status, &chargeReceiverApplied, &ket,
		&dibuat, &diupd, &biayaPerkiraan, &biayaAktual,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"id":                      id,
		"member_id":               mid,
		"ref_id":                  refid,
		"perintah":                perintah,
		"kode_produk":             kodeProduk,
		"tujuan":                  tujuan,
		"qty":                     qty,
		"qty_provider":            qtyProvider,
		"status":                  status,
		"charge_receiver_applied": chargeReceiverApplied,
		"dibuat_pada":             dibuat,
		"diperbarui_pada":         diupd,
		"biaya_perkiraan":         biayaPerkiraan,
		"biaya_aktual":            biayaAktual,
	}
	if ket.Valid {
		out["keterangan"] = ket.String
	}
	return out, nil
}

func (r *MemberTrxMemberRepository) GetTransaksiMemberByID(ctx context.Context, trxID int64) (*TrxMemberFull, error) {
	const q = `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, status,
       COALESCE(qty_provider, qty), COALESCE(keterangan, ''), COALESCE(charge_receiver_applied, false),
       biaya_perkiraan, biaya_aktual, fee_member_rp, harga_javapay, harga_member
FROM public.transaksi_member
WHERE id=$1
LIMIT 1`
	var t TrxMemberFull
	err := r.db.QueryRowContext(ctx, q, trxID).Scan(
		&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty, &t.Status,
		&t.QtyProvider, &t.Keterangan, &t.ChargeReceiverApplied,
		&t.BiayaPerkiraan, &t.BiayaAktual, &t.FeeMemberRp, &t.HargaJavapay, &t.HargaMember,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
