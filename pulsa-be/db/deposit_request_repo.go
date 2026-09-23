package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type DepositRequest struct {
	ID           int64      `json:"id"`
	MemberID     int64      `json:"member_id"`
	MemberNama   *string    `json:"member_nama,omitempty"`
	Amount       int64      `json:"amount"`
	Metode       string     `json:"metode"`
	BuktiURL     string     `json:"bukti_url"`
	Status       string     `json:"status"` // pending | approved | rejected
	Note         string     `json:"note"`
	DibuatPada   time.Time  `json:"dibuat_pada"`
	DiprosesPada *time.Time `json:"diproses_pada"`
	DiprosesOleh *int64     `json:"diproses_oleh"`
	DiprosesNama *string    `json:"diproses_nama,omitempty"`
}

// ===============================
// MEMBER: Create deposit request
// ===============================
func (r *MemberRepo) CreateDepositRequest(ctx context.Context, memberID, amount int64, metode, buktiURL string) error {
	if memberID <= 0 || amount <= 0 {
		return errors.New("invalid member_id/amount")
	}
	if metode == "" {
		return errors.New("metode required")
	}

	_, err := r.DB.ExecContext(ctx, `
INSERT INTO deposit_request (member_id, amount, metode, bukti_url, status)
VALUES ($1,$2,$3,$4,'pending')
`, memberID, amount, metode, buktiURL)
	return err
}

// ===============================
// MEMBER: List deposit requests
// ===============================
func (r *MemberRepo) ListDepositRequestsByMember(ctx context.Context, memberID int64, limit int) ([]DepositRequest, error) {
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

const q = `
SELECT d.id, d.member_id, m.nama AS member_nama, d.amount, d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM deposit_request d
LEFT JOIN member m ON m.id = d.member_id
LEFT JOIN member p ON p.id = d.diproses_oleh
WHERE d.member_id = $1
ORDER BY d.dibuat_pada DESC
LIMIT $2
`
	rows, err := r.DB.QueryContext(ctx, q, memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DepositRequest, 0, 32)
	for rows.Next() {
		var d DepositRequest
		var note sql.NullString
		var memberNama sql.NullString
		var diprosesNama sql.NullString

		if err := rows.Scan(
			&d.ID,
			&d.MemberID,
			&memberNama,
			&d.Amount,
			&d.Metode,
			&d.BuktiURL,
			&d.Status,
			&note,
			&d.DibuatPada,
			&d.DiprosesPada,
			&d.DiprosesOleh,
			&diprosesNama,
		); err != nil {
			return nil, err
		}
		if note.Valid {
			d.Note = note.String
		}
		if memberNama.Valid {
			v := memberNama.String
			d.MemberNama = &v
		}
		if diprosesNama.Valid {
			v := diprosesNama.String
			d.DiprosesNama = &v
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ===============================
// ADMIN: List requests (filter by status)
// ===============================
func (r *MemberRepo) AdminListDepositRequests(
	ctx context.Context,
	status string,
	memberID int64,
	fromStr, toStr string,
	limit, offset int,
) ([]DepositRequest, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	status = strings.TrimSpace(status)
	args = append(args, status)
	wheres = append(wheres, fmt.Sprintf("($%d = '' OR status = $%d)", len(args), len(args)))

	if memberID > 0 {
		args = append(args, memberID)
		wheres = append(wheres, fmt.Sprintf("d.member_id = $%d", len(args)))
	}

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("d.dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("d.dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT d.id, d.member_id, m.nama AS member_nama, d.amount, d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM deposit_request d
LEFT JOIN member m ON m.id = d.member_id
LEFT JOIN member p ON p.id = d.diproses_oleh
WHERE %s
ORDER BY d.dibuat_pada DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DepositRequest, 0, 32)
	for rows.Next() {
		var d DepositRequest
		var note sql.NullString
		var memberNama sql.NullString
		var diprosesNama sql.NullString

		if err := rows.Scan(
			&d.ID,
			&d.MemberID,
			&memberNama,
			&d.Amount,
			&d.Metode,
			&d.BuktiURL,
			&d.Status,
			&note,
			&d.DibuatPada,
			&d.DiprosesPada,
			&d.DiprosesOleh,
			&diprosesNama,
		); err != nil {
			return nil, err
		}
		if note.Valid {
			d.Note = note.String
		}
		if memberNama.Valid {
			v := memberNama.String
			d.MemberNama = &v
		}
		if diprosesNama.Valid {
			v := diprosesNama.String
			d.DiprosesNama = &v
		}
		out = append(out, d)
	}

	return out, rows.Err()
}

// ===============================
// ADMIN: Approve request
// - atomic: lock request -> credit wallet -> mark approved
// ===============================
func (r *MemberRepo) AdminApproveDeposit(ctx context.Context, reqID, adminID int64, note, refID string) error {
	if reqID <= 0 || adminID <= 0 {
		return errors.New("invalid req_id/admin_id")
	}
	if refID == "" {
		return errors.New("ref_id required")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var memberID, amount int64
	var status string

	// lock request row
	err = tx.QueryRowContext(ctx, `
SELECT member_id, amount, status
FROM deposit_request
WHERE id = $1
FOR UPDATE
`, reqID).Scan(&memberID, &amount, &status)
	if err != nil {
		return err
	}

	if status != "pending" {
		return errors.New("request already processed")
	}
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	// credit saldo + insert mutasi_dompet (uses schema: arah/jumlah/alasan/catatan/saldo_* )
	if err := r.creditDompetTx(ctx, tx, memberID, amount, "deposit approve", note, refID); err != nil {
		return err
	}

	// update request status
	_, err = tx.ExecContext(ctx, `
UPDATE deposit_request
SET status = 'approved',
    note = $1,
    diproses_pada = now(),
    diproses_oleh = $2
WHERE id = $3
`, note, adminID, reqID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ===============================
// ADMIN: Reject request
// ===============================
func (r *MemberRepo) AdminRejectDeposit(ctx context.Context, reqID, adminID int64, note string) error {
	if reqID <= 0 || adminID <= 0 {
		return errors.New("invalid req_id/admin_id")
	}

	res, err := r.DB.ExecContext(ctx, `
UPDATE deposit_request
SET status = 'rejected',
    note = $1,
    diproses_pada = now(),
    diproses_oleh = $2
WHERE id = $3 AND status = 'pending'
`, note, adminID, reqID)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("request already processed")
	}
	return nil
}
