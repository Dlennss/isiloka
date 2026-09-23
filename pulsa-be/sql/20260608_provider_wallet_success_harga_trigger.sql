CREATE OR REPLACE FUNCTION public.fn_provider_wallet_on_success()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
  has_debit BOOLEAN;
  saldo_skrg BIGINT;
  harga_val BIGINT;
  prov TEXT;
  debit_total BIGINT;
  refund_total BIGINT;
  refund_val BIGINT;
BEGIN
  prov := LOWER(TRIM(NEW.provider));
  IF prov = '' THEN
    RETURN NEW;
  END IF;

  IF LOWER(TRIM(COALESCE(NEW.status, ''))) = 'success' THEN
    harga_val := COALESCE(NEW.harga, 0);
    IF harga_val <= 0 THEN
      RETURN NEW;
    END IF;

    SELECT EXISTS(
      SELECT 1
      FROM public.mutasi_dompet_provider
      WHERE transaksi_provider_id = NEW.id
        AND arah = 'debit'
        AND alasan = 'TRX_SUCCESS_COST'
    ) INTO has_debit;

    IF has_debit THEN
      RETURN NEW;
    END IF;

    INSERT INTO public.dompet_provider (provider, saldo)
    VALUES (prov, 0)
    ON CONFLICT (provider) DO NOTHING;

    SELECT COALESCE(saldo, 0)
    INTO saldo_skrg
    FROM public.dompet_provider
    WHERE provider = prov
    FOR UPDATE;

    INSERT INTO public.mutasi_dompet_provider
      (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah,
       transaksi_member_id, transaksi_provider_id, dibuat_pada)
    VALUES
      (prov, NEW.ref_id, 'debit', harga_val, 'TRX_SUCCESS_COST', 'debit',
       saldo_skrg, saldo_skrg - harga_val, NEW.transaksi_member_id, NEW.id, NOW());

    UPDATE public.dompet_provider
    SET saldo = saldo - harga_val,
        diperbarui_pada = NOW()
    WHERE provider = prov;

    RETURN NEW;
  END IF;

  IF LOWER(TRIM(COALESCE(OLD.status, ''))) = 'success'
     AND LOWER(TRIM(COALESCE(NEW.status, ''))) = 'failed' THEN
    SELECT COALESCE(SUM(jumlah), 0)
    INTO debit_total
    FROM public.mutasi_dompet_provider
    WHERE provider = prov
      AND ref_id = NEW.ref_id
      AND transaksi_provider_id = NEW.id
      AND arah = 'debit'
      AND alasan = 'TRX_SUCCESS_COST';

    IF debit_total <= 0 THEN
      RETURN NEW;
    END IF;

    SELECT COALESCE(SUM(jumlah), 0)
    INTO refund_total
    FROM public.mutasi_dompet_provider
    WHERE provider = prov
      AND ref_id = NEW.ref_id
      AND transaksi_provider_id = NEW.id
      AND arah = 'credit'
      AND alasan = 'TRX_SUCCESS_COST_REFUND';

    refund_val := debit_total - refund_total;
    IF refund_val <= 0 THEN
      RETURN NEW;
    END IF;

    INSERT INTO public.dompet_provider (provider, saldo)
    VALUES (prov, 0)
    ON CONFLICT (provider) DO NOTHING;

    SELECT COALESCE(saldo, 0)
    INTO saldo_skrg
    FROM public.dompet_provider
    WHERE provider = prov
    FOR UPDATE;

    INSERT INTO public.mutasi_dompet_provider
      (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah,
       transaksi_member_id, transaksi_provider_id, dibuat_pada)
    VALUES
      (prov, NEW.ref_id, 'credit', refund_val, 'TRX_SUCCESS_COST_REFUND',
       'auto refund provider when transaction provider failed',
       saldo_skrg, saldo_skrg + refund_val, NEW.transaksi_member_id, NEW.id, NOW());

    UPDATE public.dompet_provider
    SET saldo = saldo + refund_val,
        diperbarui_pada = NOW()
    WHERE provider = prov;

    RETURN NEW;
  END IF;

  RETURN NEW;
END;
$function$;

DROP TRIGGER IF EXISTS trg_provider_wallet_on_success ON public.transaksi_provider;

CREATE TRIGGER trg_provider_wallet_on_success
AFTER UPDATE OF status, harga ON public.transaksi_provider
FOR EACH ROW
EXECUTE FUNCTION public.fn_provider_wallet_on_success();
