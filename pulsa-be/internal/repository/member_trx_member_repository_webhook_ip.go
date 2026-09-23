package repository

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"strings"
)

func (r *MemberTrxMemberRepository) IsIPAllowedForMember(ctx context.Context, memberID int64, ip string) (bool, error) {
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
	if err := r.db.QueryRowContext(ctx, `
SELECT count(*) FROM public.member_ip_whitelist WHERE member_id=$1 AND aktif=true
`, memberID).Scan(&cnt); err != nil {
		return false, err
	}
	if cnt == 0 {
		return false, nil
	}

	var ok bool
	if err := r.db.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1 FROM public.member_ip_whitelist
  WHERE member_id=$1 AND aktif=true AND ip=$2::inet
)
`, memberID, ip).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *MemberTrxMemberRepository) GetMemberWebhookURL(ctx context.Context, memberID int64) (string, error) {
	if memberID <= 0 {
		return "", errors.New("invalid member")
	}
	var wh sql.NullString
	err := r.db.QueryRowContext(ctx, `
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

func (r *MemberTrxMemberRepository) DeleteIPWhitelist(ctx context.Context, memberID int64, idStr string) error {
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
