package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type MemberRepo struct {
	DB *sql.DB
}

func NewMemberRepo(db *sql.DB) *MemberRepo { return &MemberRepo{DB: db} }

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInsufficientBalance = errors.New("saldo tidak cukup")
)

type MemberAuth struct {
	MemberID int64
	Email    sql.NullString
	Nama     sql.NullString
	Aktif    bool
	Role     string

	PasswordHash sql.NullString
	PinHash      sql.NullString
}

// ====================== AUTH ======================

func (r *MemberRepo) AuthByAPIKey(ctx context.Context, apiKey string) (*MemberAuth, error) {
	const q = `
SELECT m.id, m.email, m.nama, m.aktif, m.role, m.password_hash, m.pin_hash
FROM public.member_api_key k
JOIN public.member m ON m.id = k.member_id
WHERE k.api_key = $1 AND k.aktif = true
LIMIT 1`
	var out MemberAuth
	err := r.DB.QueryRowContext(ctx, q, apiKey).Scan(
		&out.MemberID, &out.Email, &out.Nama, &out.Aktif, &out.Role, &out.PasswordHash, &out.PinHash,
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

// ====================== TRANSAKSI MEMBER ======================

// Insert transaksi_member dan idempotent by (member_id, ref_id)
// Jika conflict, tidak buat transaksi baru dan hanya touch diperbarui_pada.
// Return trxID dan created=true kalau insert baru.
func (r *MemberRepo) CreateTransaksiMember(
	ctx context.Context,
	memberID int64,
	clientRefID string,
	perintah string,
	kodeProduk string,
	tujuan string,
	qty int64,
	biayaPerkiraan int64,
) (trxID int64, created bool, err error) {
	const q = `
	INSERT INTO public.transaksi_member
	  (member_id, ref_id, perintah, kode_produk, tujuan, qty, status, biaya_perkiraan, biaya_aktual, harga_member)
	VALUES
	  ($1,$2,$3,$4,$5,$6,'pending',$7,0,$7)
	ON CONFLICT (member_id, ref_id) DO UPDATE
	SET diperbarui_pada = now()
	RETURNING id, (xmax = 0) AS created`
	err = r.DB.QueryRowContext(ctx, q,
		memberID, clientRefID, perintah, kodeProduk, tujuan, qty, biayaPerkiraan,
	).Scan(&trxID, &created)
	return trxID, created, err
}

func (r *MemberRepo) GetTransaksiMemberByRef(ctx context.Context, memberID int64, clientRefID string) (map[string]any, error) {
	const q = `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, status, keterangan, dibuat_pada, diperbarui_pada, biaya_perkiraan, biaya_aktual
FROM public.transaksi_member
WHERE member_id = $1 AND ref_id = $2
LIMIT 1`
	row := r.DB.QueryRowContext(ctx, q, memberID, clientRefID)

	var (
		id, mid, qty, biayaPerkiraan, biayaAktual   int64
		refid, perintah, kodeProduk, tujuan, status string
		ket                                         sql.NullString
		dibuat, diupd                               time.Time
	)
	err := row.Scan(
		&id, &mid, &refid, &perintah, &kodeProduk, &tujuan, &qty, &status, &ket,
		&dibuat, &diupd, &biayaPerkiraan, &biayaAktual,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"id":              id,
		"member_id":       mid,
		"ref_id":          refid,
		"perintah":        perintah,
		"kode_produk":     kodeProduk,
		"tujuan":          tujuan,
		"qty":             qty,
		"status":          status,
		"dibuat_pada":     dibuat,
		"diperbarui_pada": diupd,
		"biaya_perkiraan": biayaPerkiraan,
		"biaya_aktual":    biayaAktual,
	}
	if ket.Valid {
		out["keterangan"] = ket.String
	}
	return out, nil
}

func (r *MemberRepo) UpdateTransaksiMemberStatus(ctx context.Context, trxID int64, status string, keterangan string, biayaAktual int64) error {
	const q = `
UPDATE public.transaksi_member
SET status = $2,
    keterangan = NULLIF($3,''),
    biaya_aktual = $4
WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, q, trxID, status, keterangan, biayaAktual)
	return err
}

// ====================== DOMPET + MUTASI ======================

func (r *MemberRepo) DebitDompet(ctx context.Context, memberID int64, refID string, jumlah int64, alasan string, catatan string) (before int64, after int64, err error) {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	if err != nil {
		return 0, 0, err
	}
	if before < jumlah {
		return 0, 0, ErrInsufficientBalance
	}
	after = before - jumlah

	_, err = tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after)
	if err != nil {
		return 0, 0, err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'DEBIT',$3,$4,NULLIF($5,''),$6,$7)
`, memberID, refID, jumlah, alasan, catatan, before, after)
	if err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}

func (r *MemberRepo) CreditDompet(ctx context.Context, memberID int64, refID string, jumlah int64, alasan string, catatan string) error {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var before int64
	err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	if err != nil {
		return err
	}

	var existingCount int64
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
  AND LOWER(COALESCE(arah, '')) = 'credit'
  AND LOWER(COALESCE(alasan, '')) = LOWER($3)
`, memberID, refID, alasan).Scan(&existingCount); err != nil {
		return err
	}
	if existingCount > 0 {
		return tx.Commit()
	}

	after := before + jumlah

	_, err = tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'CREDIT',$3,$4,NULLIF($5,''),$6,$7)
`, memberID, refID, jumlah, alasan, catatan, before, after)
	if err != nil {
		return err
	}

	return tx.Commit()
}
