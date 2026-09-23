package repository

import (
	"context"
	"database/sql"
)

type DepositRepository struct {
	db *sql.DB
}

func NewDepositRepository(db *sql.DB) *DepositRepository {
	return &DepositRepository{db: db}
}

// EnsureTicketSchema upgrades older installations that predate bank deposit
// tickets. Without these columns/status values, ticket creation is sanitized
// by the HTTP layer into the unhelpful "internal error" response.
func (r *DepositRepository) EnsureTicketSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
ALTER TABLE public.deposit_request
  ADD COLUMN IF NOT EXISTS requested_amount bigint,
  ADD COLUMN IF NOT EXISTS unique_code bigint,
  ADD COLUMN IF NOT EXISTS approved_amount bigint,
  ADD COLUMN IF NOT EXISTS ref_id text;

UPDATE public.deposit_request
SET requested_amount = amount
WHERE requested_amount IS NULL;

UPDATE public.deposit_request
SET unique_code = 0
WHERE unique_code IS NULL;

UPDATE public.deposit_request
SET approved_amount = amount
WHERE status = 'approved'
  AND approved_amount IS NULL;

ALTER TABLE public.deposit_request
  DROP CONSTRAINT IF EXISTS deposit_request_status_check;

ALTER TABLE public.deposit_request
  ADD CONSTRAINT deposit_request_status_check
  CHECK (status = ANY (ARRAY['ticket'::text, 'pending'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text]));

CREATE INDEX IF NOT EXISTS idx_deposit_request_ref_id
  ON public.deposit_request(ref_id);
`)
	return err
}
