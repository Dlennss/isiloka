GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLE public.retail_commission_ledger
TO syarif;

GRANT USAGE, SELECT
ON SEQUENCE public.retail_commission_ledger_id_seq
TO syarif;

GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLE public.retail_withdraw_request
TO syarif;

GRANT USAGE, SELECT
ON SEQUENCE public.retail_withdraw_request_id_seq
TO syarif;
