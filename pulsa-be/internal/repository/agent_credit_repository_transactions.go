package repository

import (
	"context"
	"strings"
	"time"
)

type AgentCreditAgentTransaction struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	MemberName     string    `json:"member_nama"`
	MemberEmail    string    `json:"member_email"`
	RefID          string    `json:"ref_id"`
	ProductCode    string    `json:"kode_produk"`
	ProductName    string    `json:"produk_nama"`
	Destination    string    `json:"tujuan"`
	Quantity       int64     `json:"qty"`
	Status         string    `json:"status"`
	EstimatedPrice int64     `json:"biaya_perkiraan"`
	ActualPrice    int64     `json:"biaya_aktual"`
	CreatedAt      time.Time `json:"dibuat_pada"`
	UpdatedAt      time.Time `json:"diperbarui_pada"`
}

func (r *AgentCreditRepository) ListAgentTransactions(ctx context.Context, status, search string, limit int) ([]AgentCreditAgentTransaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
WITH agent_activity AS (
  SELECT
    ao.id,
    m.id AS member_id,
    COALESCE(m.nama, '') AS member_nama,
    COALESCE(m.email, '') AS member_email,
    ao.invoice_id,
    COALESCE(ao.produk_sku_snapshot, '') AS produk_sku_snapshot,
    COALESCE(ao.produk_nama_snapshot, '') AS produk_nama_snapshot,
    COALESCE(ao.dest, '') AS dest,
    COALESCE(ao.qty, 0) AS qty,
    COALESCE(ao.status, '') AS status,
    COALESCE(ao.harga_final, 0) AS harga_final,
    COALESCE(ao.dibuat_pada, NOW()) AS dibuat_pada,
    COALESCE(ao.diubah_pada, ao.dibuat_pada, NOW()) AS diubah_pada
  FROM public.app_order ao
  JOIN public.member m ON m.id = ao.member_id
  WHERE LOWER(COALESCE(m.role, '')) = 'agent'
    AND LOWER(COALESCE(ao.buyer_type, '')) = 'user'
    AND LOWER(COALESCE(ao.status, '')) IN ('success', 'failed', 'refunded')
  UNION ALL
  SELECT
    -rw.id AS id,
    m.id AS member_id,
    COALESCE(m.nama, '') AS member_nama,
    COALESCE(m.email, '') AS member_email,
    rw.ref_id AS invoice_id,
    'WITHDRAW' AS produk_sku_snapshot,
    CONCAT('Penarikan ', COALESCE(NULLIF(rw.bank_name, ''), 'Saldo')) AS produk_nama_snapshot,
    COALESCE(rw.account_number, '') AS dest,
    COALESCE(rw.amount, 0) AS qty,
    CASE LOWER(COALESCE(rw.status, ''))
      WHEN 'approved' THEN 'success'
      WHEN 'rejected' THEN 'refunded'
      ELSE 'processing_provider'
    END AS status,
    COALESCE(rw.amount, 0) AS harga_final,
    COALESCE(rw.created_at, NOW()) AS dibuat_pada,
    COALESCE(rw.updated_at, rw.created_at, NOW()) AS diubah_pada
  FROM public.retail_withdraw_request rw
  JOIN public.member m ON m.id = rw.member_id
  WHERE LOWER(COALESCE(m.role, '')) = 'agent'
    AND LOWER(COALESCE(rw.status, '')) IN ('approved', 'rejected')
), latest_agent_order AS (
  SELECT
    agent_activity.*,
    ROW_NUMBER() OVER (PARTITION BY member_id ORDER BY dibuat_pada DESC, id DESC) AS urutan
  FROM agent_activity
)
SELECT
  id,
  member_id,
  member_nama,
  member_email,
  invoice_id,
  produk_sku_snapshot,
  produk_nama_snapshot,
  dest,
  qty,
  status,
  harga_final,
  harga_final,
  dibuat_pada,
  diubah_pada
FROM latest_agent_order
WHERE urutan = 1
  AND (
    $1 = ''
    OR ($1 = 'refunded' AND LOWER(status) IN ('failed', 'refunded'))
    OR ($1 <> 'refunded' AND LOWER(status) = $1)
  )
  AND (
    $2 = ''
    OR member_nama ILIKE '%' || $2 || '%'
    OR member_email ILIKE '%' || $2 || '%'
    OR invoice_id ILIKE '%' || $2 || '%'
    OR produk_sku_snapshot ILIKE '%' || $2 || '%'
    OR produk_nama_snapshot ILIKE '%' || $2 || '%'
    OR dest ILIKE '%' || $2 || '%'
  )
ORDER BY dibuat_pada DESC, id DESC
LIMIT $3
`, strings.ToLower(strings.TrimSpace(status)), strings.TrimSpace(search), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentCreditAgentTransaction, 0)
	for rows.Next() {
		var item AgentCreditAgentTransaction
		if err := rows.Scan(
			&item.ID, &item.MemberID, &item.MemberName, &item.MemberEmail,
			&item.RefID, &item.ProductCode, &item.ProductName, &item.Destination,
			&item.Quantity, &item.Status, &item.EstimatedPrice, &item.ActualPrice,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
