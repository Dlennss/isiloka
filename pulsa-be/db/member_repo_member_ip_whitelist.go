package db

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type IPWhitelistRow struct {
	ID         int64   `json:"id"`
	IP         string  `json:"ip"`
	Label      *string `json:"label,omitempty"`
	WebhookURL *string `json:"webhook_url,omitempty"`
	Aktif      bool    `json:"aktif"`
}

func (r *MemberRepo) ListIPWhitelist(ctx context.Context, memberID int64) ([]IPWhitelistRow, error) {
	rows, err := r.DB.QueryContext(ctx, `
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

	out := make([]IPWhitelistRow, 0, 8)
	for rows.Next() {
		var row IPWhitelistRow
		var label sql.NullString
		var wh sql.NullString
		if err := rows.Scan(&row.ID, &row.IP, &label, &wh, &row.Aktif); err != nil {
			return nil, err
		}
		if label.Valid {
			v := label.String
			row.Label = &v
		}
		if wh.Valid {
			v := wh.String
			row.WebhookURL = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *MemberRepo) AddIPWhitelist(ctx context.Context, memberID int64, ip string, label string, webhookURL string) error {
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

	_, err = r.DB.ExecContext(ctx, `
INSERT INTO public.member_ip_whitelist (member_id, ip, label, webhook_url, aktif)
VALUES ($1, $2::inet, NULLIF($3,''), NULLIF($4,''), true)
ON CONFLICT (member_id, ip) DO UPDATE
SET aktif=true,
    label = NULLIF(EXCLUDED.label,''),
    webhook_url = NULLIF(EXCLUDED.webhook_url,'')
`, memberID, ip, label, webhookURL)
	return err
}

func (r *MemberRepo) DeleteIPWhitelist(ctx context.Context, memberID int64, idStr string) error {
	if memberID <= 0 || strings.TrimSpace(idStr) == "" {
		return errors.New("invalid payload")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return errors.New("id tidak valid")
	}
	_, err = r.DB.ExecContext(ctx, `DELETE FROM public.member_ip_whitelist WHERE id=$1 AND member_id=$2`, id, memberID)
	return err
}

// STRICT: whitelist kosong => reject
func (r *MemberRepo) IsIPAllowedForMember(ctx context.Context, memberID int64, ip string) (bool, error) {
	if memberID <= 0 {
		return false, errors.New("invalid member")
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false, errors.New("missing ip")
	}
	if net.ParseIP(ip) == nil {
		return false, errors.New("format ip tidak valid")
	}

	var cnt int64
	if err := r.DB.QueryRowContext(ctx, `
SELECT count(*) FROM public.member_ip_whitelist WHERE member_id=$1 AND aktif=true
`, memberID).Scan(&cnt); err != nil {
		return false, err
	}
	if cnt == 0 {
		return false, nil
	}

	var ok bool
	if err := r.DB.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1 FROM public.member_ip_whitelist
  WHERE member_id=$1 AND aktif=true AND ip=$2::inet
)
`, memberID, ip).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

// Ambil webhook terbaru dari whitelist aktif
func (r *MemberRepo) GetMemberWebhookURL(ctx context.Context, memberID int64) (string, error) {
	if memberID <= 0 {
		return "", errors.New("invalid member")
	}
	var wh sql.NullString
	err := r.DB.QueryRowContext(ctx, `
SELECT webhook_url
FROM public.member_ip_whitelist
WHERE member_id=$1 AND aktif=true AND webhook_url IS NOT NULL AND webhook_url <> ''
ORDER BY dibuat_pada DESC, id DESC
LIMIT 1
`, memberID).Scan(&wh)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if !wh.Valid {
		return "", nil
	}
	return strings.TrimSpace(wh.String), nil
}
