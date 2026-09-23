package repository

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (r *MemberSelfRepository) ChangePassword(ctx context.Context, memberID int64, oldPassword string, newPasswordHash string) error {
	if memberID <= 0 || oldPassword == "" || newPasswordHash == "" {
		return errors.New("invalid payload")
	}
	var oldHash sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM public.member WHERE id=$1`, memberID).Scan(&oldHash)
	if err != nil {
		return err
	}
	if !oldHash.Valid || oldHash.String == "" {
		return errors.New("password belum diset")
	}
	if bcrypt.CompareHashAndPassword([]byte(oldHash.String), []byte(oldPassword)) != nil {
		return errors.New("old_password salah")
	}
	_, err = r.db.ExecContext(ctx, `UPDATE public.member SET password_hash=$1 WHERE id=$2`, newPasswordHash, memberID)
	return err
}

func (r *MemberSelfRepository) ChangePIN(ctx context.Context, memberID int64, oldPIN string, newPinHash string) error {
	if memberID <= 0 || oldPIN == "" || newPinHash == "" {
		return errors.New("invalid payload")
	}
	var oldHash sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT pin_hash FROM public.member WHERE id=$1`, memberID).Scan(&oldHash)
	if err != nil {
		return err
	}
	if !oldHash.Valid || oldHash.String == "" {
		return errors.New("pin belum diset")
	}
	if bcrypt.CompareHashAndPassword([]byte(oldHash.String), []byte(oldPIN)) != nil {
		return errors.New("old_pin salah")
	}
	_, err = r.db.ExecContext(ctx, `UPDATE public.member SET pin_hash=$1 WHERE id=$2`, newPinHash, memberID)
	return err
}

func (r *MemberSelfRepository) ResetAPIKey(ctx context.Context, memberID int64, apiKey string, label string) error {
	if memberID <= 0 || apiKey == "" {
		return errors.New("invalid payload")
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `UPDATE public.member_api_key SET aktif=false WHERE member_id=$1`, memberID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.member_api_key (member_id, api_key, label, aktif)
VALUES ($1,$2,NULLIF($3,''),true)
`, memberID, apiKey, label); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MemberSelfRepository) ListIPWhitelist(ctx context.Context, memberID int64) ([]MemberIPWhitelist, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, ip::text, label, webhook_url, aktif
FROM public.member_ip_whitelist
WHERE member_id=$1
ORDER BY id DESC
LIMIT 200
`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemberIPWhitelist, 0, 8)
	for rows.Next() {
		var row MemberIPWhitelist
		var label sql.NullString
		var webhook sql.NullString
		if err := rows.Scan(&row.ID, &row.IP, &label, &webhook, &row.Aktif); err != nil {
			return nil, err
		}
		if label.Valid {
			v := label.String
			row.Label = &v
		}
		if webhook.Valid {
			v := webhook.String
			row.WebhookURL = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *MemberSelfRepository) AddIPWhitelist(ctx context.Context, memberID int64, ip string, label string, webhookURL string) error {
	if memberID <= 0 {
		return errors.New("invalid member")
	}
	ip = strings.TrimSpace(ip)
	label = strings.TrimSpace(label)
	webhookURL = strings.TrimSpace(webhookURL)
	if ip == "" {
		return errors.New("ip required")
	}
	if webhookURL == "" {
		return errors.New("webhook_url required")
	}
	if net.ParseIP(ip) == nil {
		return errors.New("format ip tidak valid")
	}
	u, err := url.ParseRequestURI(webhookURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("webhook_url tidak valid")
	}

	_, err = r.db.ExecContext(ctx, `
INSERT INTO public.member_ip_whitelist (member_id, ip, label, webhook_url, aktif)
VALUES ($1, $2::inet, NULLIF($3,''), NULLIF($4,''), true)
ON CONFLICT (member_id, ip) DO UPDATE
SET aktif=true,
    label = NULLIF(EXCLUDED.label,''),
    webhook_url = NULLIF(EXCLUDED.webhook_url,'')
`, memberID, ip, label, webhookURL)
	return err
}

func (r *MemberSelfRepository) DeleteIPWhitelist(ctx context.Context, memberID int64, idStr string) error {
	if memberID <= 0 || strings.TrimSpace(idStr) == "" {
		return errors.New("invalid payload")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return errors.New("id tidak valid")
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM public.member_ip_whitelist WHERE id=$1 AND member_id=$2`, id, memberID)
	return err
}
