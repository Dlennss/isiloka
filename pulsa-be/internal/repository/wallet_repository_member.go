package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type MemberWalletMutasiRow struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	MemberNama     *string   `json:"member_nama,omitempty"`
	BankID         *int64    `json:"bank_id,omitempty"`
	BankNama       *string   `json:"bank_nama,omitempty"`
	BankNomor      *string   `json:"bank_nomor_rekening,omitempty"`
	BankAtasNama   *string   `json:"bank_atas_nama,omitempty"`
	RefID          string    `json:"ref_id"`
	Arah           string    `json:"arah"`
	Jumlah         int64     `json:"jumlah"`
	Alasan         string    `json:"alasan"`
	Catatan        string    `json:"catatan"`
	SaldoSebelum   int64     `json:"saldo_sebelum"`
	SaldoSesudah   int64     `json:"saldo_sesudah"`
	DiubahOleh     *int64    `json:"diubah_oleh,omitempty"`
	DiubahOlehNama *string   `json:"diubah_oleh_nama,omitempty"`
	DibuatPada     time.Time `json:"dibuat_pada"`
}

func (r *WalletRepository) AdminAdjustMemberWallet(ctx context.Context, actorID, memberID, amount int64, direction, reason, note, refID string) error {
	if actorID <= 0 || memberID <= 0 || amount <= 0 {
		return errors.New("member_id/amount invalid")
	}
	direction = strings.TrimSpace(strings.ToLower(direction))
	if direction != "credit" && direction != "debit" {
		return errors.New("direction invalid")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var before int64
	if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE`, memberID).Scan(&before); err != nil {
		return err
	}

	var after int64
	arahMutasi := "CREDIT"
	if direction == "credit" {
		after = before + amount
	} else {
		if before < amount {
			return errors.New("saldo tidak cukup")
		}
		after = before - amount
		arahMutasi = "DEBIT"
	}

	if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo = $1, diperbarui_pada = now() WHERE member_id = $2`, after, memberID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada)
VALUES
  ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,now())
`, memberID, refID, arahMutasi, amount, reason, note, before, after, actorID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *WalletRepository) ListMemberDepositMutasi(ctx context.Context, memberID int64, arah, refID, fromStr, toStr string, limit, offset int) ([]MemberWalletMutasiRow, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	arah = strings.TrimSpace(strings.ToLower(arah))
	refID = strings.TrimSpace(refID)

	var (
		args   []any
		wheres []string
	)
	wheres = append(wheres, `(
		(LOWER(TRIM(COALESCE(md.arah, ''))) = 'credit' AND LOWER(TRIM(COALESCE(md.alasan, ''))) IN ('deposit approve', 'deposit qris', 'deposit va', 'deposit', 'deposit_api', 'admin manual credit'))
		OR
		(LOWER(TRIM(COALESCE(md.arah, ''))) = 'debit' AND LOWER(TRIM(COALESCE(md.alasan, ''))) = 'admin manual debit')
	)`)

	if memberID > 0 {
		args = append(args, memberID)
		wheres = append(wheres, fmt.Sprintf("md.member_id = $%d", len(args)))
	}
	if arah != "" {
		if arah != "credit" && arah != "debit" {
			return nil, 0, errors.New("arah must be credit|debit")
		}
		args = append(args, arah)
		wheres = append(wheres, fmt.Sprintf("LOWER(TRIM(COALESCE(md.arah, ''))) = $%d", len(args)))
	}
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("md.ref_id ILIKE '%%' || $%d || '%%'", len(args)))
	}
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("md.dibuat_pada >= $%d", len(args)))
	}
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("md.dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	qList := fmt.Sprintf(`
SELECT
  md.id, md.member_id, mb.nama, dr.bank_id, NULLIF(dr.bank_nama,''), NULLIF(dr.bank_nomor_rekening,''), NULLIF(dr.bank_atas_nama,''), md.ref_id, md.arah, md.jumlah, md.alasan, COALESCE(md.catatan,''), md.saldo_sebelum, md.saldo_sesudah,
  md.diubah_oleh, actor.nama, md.dibuat_pada
FROM public.mutasi_dompet md
JOIN public.member mb ON mb.id = md.member_id
LEFT JOIN public.deposit_request dr ON dr.member_id = md.member_id AND dr.status = 'approved' AND dr.ref_id = md.ref_id
LEFT JOIN public.member actor ON actor.id = md.diubah_oleh
WHERE %s
ORDER BY md.dibuat_pada DESC, md.id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]MemberWalletMutasiRow, 0, limit)
	for rows.Next() {
		var (
			x         MemberWalletMutasiRow
			memberNm  sql.NullString
			bankID    sql.NullInt64
			bankNama  sql.NullString
			bankNomor sql.NullString
			bankAtas  sql.NullString
			actorID   sql.NullInt64
			actorName sql.NullString
		)
		if err := rows.Scan(
			&x.ID, &x.MemberID, &memberNm, &bankID, &bankNama, &bankNomor, &bankAtas, &x.RefID, &x.Arah, &x.Jumlah,
			&x.Alasan, &x.Catatan, &x.SaldoSebelum, &x.SaldoSesudah, &actorID, &actorName, &x.DibuatPada,
		); err != nil {
			return nil, 0, err
		}
		if memberNm.Valid {
			v := memberNm.String
			x.MemberNama = &v
		}
		if bankID.Valid {
			v := bankID.Int64
			x.BankID = &v
		}
		if bankNama.Valid {
			v := bankNama.String
			x.BankNama = &v
		}
		if bankNomor.Valid {
			v := bankNomor.String
			x.BankNomor = &v
		}
		if bankAtas.Valid {
			v := bankAtas.String
			x.BankAtasNama = &v
		}
		if actorID.Valid {
			v := actorID.Int64
			x.DiubahOleh = &v
		}
		if actorName.Valid {
			v := actorName.String
			x.DiubahOlehNama = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := args[:limitPos-1]
	var total int64
	qCount := fmt.Sprintf(`SELECT COUNT(1) FROM public.mutasi_dompet md WHERE %s`, strings.Join(wheres, " AND "))
	if err := r.db.QueryRowContext(ctx, qCount, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
