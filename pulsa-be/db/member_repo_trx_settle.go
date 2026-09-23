package db

import "context"

func (r *MemberRepo) UpdateTransaksiMemberSettle(ctx context.Context, trxID int64, status string, ket string, biayaAktual int64, hargaJavapay int64, hargaMember int64) error {
	_, err := r.DB.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status=$2,
    keterangan=NULLIF($3,''),
    biaya_aktual=$4,
    harga_javapay=$5,
    harga_member=$6
WHERE id=$1
`, trxID, status, ket, biayaAktual, hargaJavapay, hargaMember)
	return err
}

func (r *MemberRepo) UpdateTransaksiMemberFee(ctx context.Context, trxID int64, feeMember int64) error {
	_, err := r.DB.ExecContext(ctx, `
UPDATE public.transaksi_member
SET fee_member_rp=$2
WHERE id=$1
`, trxID, feeMember)
	return err
}
