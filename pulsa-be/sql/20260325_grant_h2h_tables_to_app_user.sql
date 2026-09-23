DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE
    ON TABLE public.h2h_commission_ledger
    TO app_user;

    GRANT USAGE, SELECT
    ON SEQUENCE public.h2h_commission_ledger_id_seq
    TO app_user;

    GRANT SELECT, INSERT, UPDATE, DELETE
    ON TABLE public.h2h_withdraw_request
    TO app_user;

    GRANT USAGE, SELECT
    ON SEQUENCE public.h2h_withdraw_request_id_seq
    TO app_user;
  END IF;
END $$;
