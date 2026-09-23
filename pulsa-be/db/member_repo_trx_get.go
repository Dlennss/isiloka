package db

import (
	"context"
	"database/sql"
	"errors"
)

type TrxMemberFull struct {
	ID            int64
	MemberID      int64
	RefID         string
	Perintah      string
	KodeProduk    string
	Tujuan        string
	Qty           int64
	Status        string
	BiayaPerkiraan int64
	BiayaAktual    int64
	FeeMemberRp   int64
	HargaJavapay  int64
	HargaMember   int64
}

func (r *MemberRepo) GetTransaksiMemberByID(ctx context.Context, trxID int64) (*TrxMemberFull, error) {
	const q = `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, status,
       biaya_perkiraan, biaya_aktual, fee_member_rp, harga_javapay, harga_member
FROM public.transaksi_member
WHERE id=$1
LIMIT 1`
	var t TrxMemberFull
	err := r.DB.QueryRowContext(ctx, q, trxID).Scan(
		&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty, &t.Status,
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
