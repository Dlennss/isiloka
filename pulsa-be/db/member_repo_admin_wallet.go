package db

import (
	"context"
	"errors"
)

func (r *MemberRepo) AdminAdjustSaldo(
	ctx context.Context,
	memberID int64,
	jumlah int64,
	direction string,
	alasan string,
	catatan string,
	refID string,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch direction {
	case "credit":
		err = r.creditDompetTx(ctx, tx, memberID, jumlah, alasan, catatan, refID)
	case "debit":
		err = r.debitDompetTx(ctx, tx, memberID, jumlah, alasan, catatan, refID)
	default:
		return errors.New("direction invalid")
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
