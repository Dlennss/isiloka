package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (r *AuditRepository) ResolveProviderWalletMissingDebit(
	ctx context.Context,
	actorID int64,
	transaksiProviderID int64,
) (*ResolveProviderWalletMissingDebitResult, error) {
	if err := r.ensureProviderWalletMissingDebitIgnoreTable(ctx); err != nil {
		return nil, err
	}
	if transaksiProviderID <= 0 {
		return nil, fmt.Errorf("transaksi_provider_id harus > 0")
	}

	const q = `
SELECT
  tm.id AS transaksi_member_id,
  tm.member_id,
  tm.status AS status_member,
  tm.ref_id AS ref_id_member,
  tm.kode_produk AS produk_member,
  tm.tujuan AS tujuan_member,
  tm.qty AS qty_member,
  tm.qty_provider,
  tp.id AS transaksi_provider_id,
  tp.provider,
  tp.ref_id AS ref_id_provider,
  tp.perintah,
  tp.kode_produk,
  tp.tujuan,
  tp.qty,
  tp.harga,
  tp.kode_respon,
  tp.pesan,
  tp.no_referensi,
  tp.dibuat_pada AS provider_dibuat_pada
FROM public.transaksi_provider tp
JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
WHERE tp.id = $1
  AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
  AND COALESCE(tp.harga, 0) > 0
  AND LOWER(TRIM(COALESCE(tp.status, ''))) = 'success'
`

	var (
		row         AdminProviderWalletMissingDebitRow
		qtyProvider sql.NullInt64
		kodeRespon  sql.NullString
		pesan       sql.NullString
		noReferensi sql.NullString
	)
	if err := r.db.QueryRowContext(ctx, q, transaksiProviderID).Scan(
		&row.TransaksiMemberID, &row.MemberID, &row.StatusMember, &row.RefIDMember, &row.ProdukMember, &row.TujuanMember,
		&row.QtyMember, &qtyProvider, &row.TransaksiProviderID, &row.Provider, &row.RefIDProvider, &row.Perintah,
		&row.KodeProduk, &row.Tujuan, &row.Qty, &row.Harga, &kodeRespon, &pesan, &noReferensi, &row.ProviderDibuatPada,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaksi audit tidak ditemukan atau sudah tidak eligible")
		}
		return nil, err
	}
	row.KodeProduk = normalizeProviderProductCodeDisplay(row.Provider, row.KodeProduk)
	if qtyProvider.Valid {
		v := qtyProvider.Int64
		row.QtyProvider = &v
	}
	if kodeRespon.Valid {
		v := strings.TrimSpace(kodeRespon.String)
		row.KodeRespon = &v
	}
	if pesan.Valid {
		v := strings.TrimSpace(pesan.String)
		row.Pesan = &v
	}
	if noReferensi.Valid {
		v := strings.TrimSpace(noReferensi.String)
		row.NoReferensi = &v
	}

	var debitCount int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.mutasi_dompet_provider
WHERE transaksi_provider_id = $1
  AND LOWER(TRIM(COALESCE(arah, ''))) = 'debit'
`, transaksiProviderID).Scan(&debitCount); err != nil {
		return nil, err
	}
	if debitCount > 0 {
		return &ResolveProviderWalletMissingDebitResult{
			Resolved: false, AlreadyResolved: true, TransaksiProviderID: transaksiProviderID, Row: &row,
		}, nil
	}

	pcbRepo := NewProviderCallbackRepository(r.db)
	before, after, err := pcbRepo.ApplyProviderWalletTx(ctx, CallbackProviderWalletTxIn{
		Provider:            row.Provider,
		RefID:               row.RefIDProvider,
		Arah:                "debit",
		Jumlah:              row.Harga,
		Alasan:              "TRX_SUCCESS_COST",
		Catatan:             fmt.Sprintf("manual resolve audit missing debit by admin_id=%d", actorID),
		TransaksiMemberID:   &row.TransaksiMemberID,
		TransaksiProviderID: &row.TransaksiProviderID,
	})
	if err != nil {
		return nil, err
	}

	return &ResolveProviderWalletMissingDebitResult{
		Resolved: true, AlreadyResolved: false, SaldoSebelum: before, SaldoSesudah: after, TransaksiProviderID: transaksiProviderID, Row: &row,
	}, nil
}

func (r *AuditRepository) IgnoreProviderWalletMissingDebit(
	ctx context.Context,
	actorID int64,
	transaksiProviderID int64,
	note string,
) error {
	if err := r.ensureProviderWalletMissingDebitIgnoreTable(ctx); err != nil {
		return err
	}
	if transaksiProviderID <= 0 {
		return fmt.Errorf("transaksi_provider_id harus > 0")
	}

	var exists int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.transaksi_provider tp
JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
WHERE tp.id = $1
  AND LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'
  AND COALESCE(tp.harga, 0) > 0
  AND LOWER(TRIM(COALESCE(tp.status, ''))) = 'success'
  AND NOT EXISTS (
    SELECT 1
    FROM public.mutasi_dompet_provider mdp
    WHERE mdp.transaksi_provider_id = tp.id
      AND LOWER(TRIM(COALESCE(mdp.arah, ''))) = 'debit'
  )
`, transaksiProviderID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("transaksi audit tidak ditemukan atau sudah tidak eligible")
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.audit_provider_wallet_missing_debit_ignore
  (transaksi_provider_id, ignored_by_admin_id, note)
VALUES
  ($1, $2, $3)
ON CONFLICT (transaksi_provider_id)
DO UPDATE
SET ignored_by_admin_id = EXCLUDED.ignored_by_admin_id,
    note = EXCLUDED.note
`, transaksiProviderID, actorID, strings.TrimSpace(note))
	return err
}
