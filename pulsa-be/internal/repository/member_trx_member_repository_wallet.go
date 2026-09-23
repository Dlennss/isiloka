package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *MemberTrxMemberRepository) DebitDompet(ctx context.Context, memberID int64, refID string, jumlah int64, alasan string, catatan string) (before int64, after int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
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

func (r *MemberTrxMemberRepository) DebitDompetHoldToTarget(ctx context.Context, memberID int64, refID string, targetHold int64, alasan string, catatan string) (before int64, after int64, debited int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	if err != nil {
		return 0, 0, 0, err
	}

	outstanding, err := memberWalletOutstandingHoldByRef(ctx, tx, memberID, refID)
	if err != nil {
		return 0, 0, 0, err
	}

	debited = ComputeWalletHoldDebitAmount(targetHold, outstanding)
	if debited <= 0 {
		if err := tx.Commit(); err != nil {
			return 0, 0, 0, err
		}
		return before, before, 0, nil
	}
	if before < debited {
		return 0, 0, 0, ErrInsufficientBalance
	}
	after = before - debited

	_, err = tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after)
	if err != nil {
		return 0, 0, 0, err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'DEBIT',$3,$4,NULLIF($5,''),$6,$7)
`, memberID, refID, debited, alasan, catatan, before, after)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, err
	}
	return before, after, debited, nil
}

func (r *MemberTrxMemberRepository) GetOutstandingHoldAmountByRef(ctx context.Context, memberID int64, refID string) (int64, error) {
	return memberWalletOutstandingHoldByRef(ctx, r.db, memberID, refID)
}

func (r *MemberTrxMemberRepository) CreditDompet(ctx context.Context, memberID int64, refID string, jumlah int64, alasan string, catatan string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var before int64
	err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	if err != nil {
		return err
	}

	if strings.EqualFold(strings.TrimSpace(alasan), "REFUND") {
		outstanding, err := memberWalletOutstandingHoldByRef(ctx, tx, memberID, refID)
		if err != nil {
			return err
		}
		jumlah = clampWalletRefundAmount(jumlah, outstanding)
		if jumlah <= 0 {
			return tx.Commit()
		}
	} else {
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

func (r *MemberTrxMemberRepository) GetSaldo(ctx context.Context, memberID int64) (int64, error) {
	var saldo int64
	err := r.db.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id=$1
`, memberID).Scan(&saldo)
	return saldo, err
}
