-- Read-only P24 SMPAY marker coverage preflight.
--
-- Required psql variables:
--   cutover: quoted timestamptz literal, for example '2026-06-20 22:10:00+07'
--   ids: quoted comma-separated P24 member id list from P24_SMPAY_SOURCE_MEMBER_IDS
--
-- Example:
--   psql "$DATABASE_URL" -v cutover="'2026-06-20 22:10:00+07'" -v ids="'1,2,3'" -f p24-smpay-marker-coverage-preflight.sql

\pset pager off
\pset format aligned

WITH params AS (
  SELECT :cutover::timestamptz AS cutover,
         string_to_array(:ids, ',')::bigint[] AS source_member_ids
), smpay_members AS (
  SELECT unnest(source_member_ids) AS member_id FROM params
), candidates AS (
  SELECT tm.id, tm.member_id, tm.ref_id, tm.perintah, tm.kode_produk, tm.status,
         COALESCE(tm.biaya_aktual, tm.biaya_perkiraan, tm.qty, 0) AS amount,
         tm.dibuat_pada, tm.diperbarui_pada
  FROM public.transaksi_member tm
  JOIN smpay_members sm ON sm.member_id = tm.member_id
  JOIN params p ON true
  WHERE tm.dibuat_pada >= p.cutover
    AND btrim(COALESCE(tm.ref_id, '')) <> ''
), joined AS (
  SELECT c.*,
         rs.id AS ref_source_id,
         rs.smpay_transaction_id AS ref_smpay_transaction_id,
         rs.raw_request,
         ts.transaksi_member_id AS tx_source_id,
         ts.smpay_transaction_id AS tx_smpay_transaction_id
  FROM candidates c
  LEFT JOIN public.smpay_ref_sources rs ON rs.member_id = c.member_id AND rs.ref_id = c.ref_id
  LEFT JOIN public.smpay_transaction_sources ts ON ts.transaksi_member_id = c.id
)
SELECT 'created_since_cutover' AS check_name,
       count(*) AS total,
       count(*) FILTER (WHERE ref_source_id IS NOT NULL) AS with_ref_source,
       count(*) FILTER (WHERE tx_source_id IS NOT NULL) AS with_tx_source,
       count(*) FILTER (
         WHERE ref_source_id IS NOT NULL
           AND COALESCE((raw_request->>'p24_request_has_marker')::boolean, true) = false
       ) AS fallback_ref_source_rows,
       count(*) FILTER (WHERE COALESCE(ref_smpay_transaction_id, tx_smpay_transaction_id, 0) > 0) AS with_smpay_tx_id,
       count(*) FILTER (WHERE ref_source_id IS NULL) AS missing_ref_source,
       count(*) FILTER (WHERE tx_source_id IS NULL) AS missing_tx_source,
       count(*) FILTER (WHERE COALESCE(ref_smpay_transaction_id, tx_smpay_transaction_id, 0) <= 0) AS missing_smpay_tx_id,
       COALESCE(sum(amount), 0) AS amount_total
FROM joined;

WITH params AS (
  SELECT :cutover::timestamptz AS cutover,
         string_to_array(:ids, ',')::bigint[] AS source_member_ids
), smpay_members AS (
  SELECT unnest(source_member_ids) AS member_id FROM params
), candidates AS (
  SELECT tm.id, tm.member_id, tm.ref_id, tm.perintah, tm.kode_produk, tm.status,
         COALESCE(tm.biaya_aktual, tm.biaya_perkiraan, tm.qty, 0) AS amount,
         tm.dibuat_pada, tm.diperbarui_pada
  FROM public.transaksi_member tm
  JOIN smpay_members sm ON sm.member_id = tm.member_id
  JOIN params p ON true
  WHERE tm.dibuat_pada >= p.cutover
    AND btrim(COALESCE(tm.ref_id, '')) <> ''
), joined AS (
  SELECT c.*,
         rs.id AS ref_source_id,
         rs.raw_request,
         ts.transaksi_member_id AS tx_source_id
  FROM candidates c
  LEFT JOIN public.smpay_ref_sources rs ON rs.member_id = c.member_id AND rs.ref_id = c.ref_id
  LEFT JOIN public.smpay_transaction_sources ts ON ts.transaksi_member_id = c.id
)
SELECT 'missing_marker_samples' AS section,
       id, member_id, ref_id, perintah, kode_produk, status, amount,
       dibuat_pada AT TIME ZONE 'Asia/Jakarta' AS dibuat_wib,
       diperbarui_pada AT TIME ZONE 'Asia/Jakarta' AS diperbarui_wib,
       COALESCE((raw_request->>'p24_request_has_marker')::boolean, true) AS p24_request_has_marker,
       (ref_source_id IS NOT NULL) AS has_ref_source,
       (tx_source_id IS NOT NULL) AS has_tx_source
FROM joined
WHERE ref_source_id IS NULL OR tx_source_id IS NULL
ORDER BY dibuat_pada DESC, id DESC
LIMIT 50;

WITH orphan_tx_source AS (
  SELECT ts.*
  FROM public.smpay_transaction_sources ts
  LEFT JOIN public.transaksi_member tm ON tm.id = ts.transaksi_member_id
  WHERE tm.id IS NULL
), orphan_ref_link AS (
  SELECT ts.*
  FROM public.smpay_transaction_sources ts
  LEFT JOIN public.smpay_ref_sources rs ON rs.id = ts.smpay_ref_source_id
  WHERE ts.smpay_ref_source_id IS NOT NULL
    AND rs.id IS NULL
), duplicate_ref AS (
  SELECT member_id, ref_id, count(*) AS rows
  FROM public.smpay_ref_sources
  GROUP BY member_id, ref_id
  HAVING count(*) > 1
)
SELECT 'integrity' AS check_name,
       (SELECT count(*) FROM orphan_tx_source) AS orphan_tx_source_rows,
       (SELECT count(*) FROM orphan_ref_link) AS orphan_ref_link_rows,
       (SELECT count(*) FROM duplicate_ref) AS duplicate_member_ref_rows;
