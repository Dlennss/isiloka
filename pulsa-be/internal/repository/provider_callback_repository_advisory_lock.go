package repository

import (
	"context"
	"database/sql"
)

// AcquireCallbackLock mengambil advisory lock berdasar transaksi_member_id.
// Lock otomatis release saat context selesai (session-level).
// Return true jika lock berhasil diambil.
func (r *ProviderCallbackRepository) AcquireCallbackLock(ctx context.Context, trxMemberID int64) (bool, error) {
	var acquired bool
	err := r.db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", trxMemberID).Scan(&acquired)
	if err != nil {
		return false, err
	}
	return acquired, nil
}

// ReleaseCallbackLock melepas advisory lock.
func (r *ProviderCallbackRepository) ReleaseCallbackLock(ctx context.Context, trxMemberID int64) {
	_, _ = r.db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", trxMemberID)
}

// SettleAndCheck wraps UpdateTransaksiMemberSettle dan return apakah status benar-benar berubah.
// Jika return sql.ErrNoRows, berarti status sudah final → caller harus SKIP financial ops.
func (r *ProviderCallbackRepository) SettleAndCheck(ctx context.Context, trxID int64, status, keterangan string, biayaAktual, hargaJavapay, hargaMember int64) (changed bool, err error) {
	err = r.UpdateTransaksiMemberSettle(ctx, trxID, status, keterangan, biayaAktual, hargaJavapay, hargaMember)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
