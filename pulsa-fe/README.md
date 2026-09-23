# Dokumentasi Jaringan Retail PulsaKilat

README ini menjelaskan modul jaringan retail pada aplikasi PulsaKilat, mulai dari role, struktur downline, komisi, withdraw, endpoint, sampai tabel database yang dipakai.

## Ringkasan

Jaringan retail dipakai untuk akun aplikasi retail, berbeda dari jaringan H2H. Role retail terdiri dari:

| Role | Nama tampilan | Fungsi utama |
| --- | --- | --- |
| `master` | Master | Upline tertinggi di jaringan retail. Bisa memiliki agent dan user. |
| `agent` | Agent | Upline menengah. Bisa memiliki user. Bisa berada di bawah master. |
| `user` | User | Pembeli retail/aplikasi. Bisa berada langsung di bawah agent atau master. |

Rule dasar:

- Master boleh membuat downline `agent` atau `user`.
- Agent hanya boleh membuat downline `user`.
- User tidak boleh membuat downline.
- Akun `agent` tidak boleh punya agent di atasnya.
- Akun `master` tidak boleh punya upline.
- Jika user dipasang ke agent yang punya master, master dapat otomatis diturunkan dari data agent saat assignment admin.

## Struktur Jaringan

Kolom relasi retail ada di tabel `public.member`:

| Kolom | Isi |
| --- | --- |
| `retail_agent_id` | ID agent retail di atas akun. Umumnya dipakai oleh user. |
| `retail_master_id` | ID master retail di atas akun. Dipakai oleh user atau agent. |
| `retail_agent_commission_rp` | Nominal komisi per order sukses untuk agent. |
| `retail_master_commission_rp` | Nominal komisi per order sukses untuk master. |

Contoh struktur:

```text
Master A
  Agent B
    User C
    User D
  User E
```

Pada struktur di atas:

- Order sukses dari `User C` atau `User D` dapat memberi komisi ke `Agent B` dan `Master A`.
- Order sukses dari `User E` dapat memberi komisi langsung ke `Master A`.
- Order sukses dari `Agent B` dapat memberi komisi ke `Master A`.

## Default Komisi

Default komisi retail adalah `Rp100` per transaksi sukses untuk role `agent` dan `master` jika nilai komisi belum diisi eksplisit.

Kode sumber:

- Backend helper role: `pulsa-be/internal/helper/roles.go`
- Kolom komisi member: `retail_agent_commission_rp`, `retail_master_commission_rp`

Catatan:

- Role `user` tidak otomatis mendapat default komisi sebagai upline.
- Nilai komisi dapat diubah oleh admin lewat data member.
- Komisi yang dipakai saat pembagian adalah nilai komisi dari akun penerima komisi, bukan dari akun pembeli.

## Alur Pembuatan Downline

### Dari Dashboard Retail

Menu user retail memakai komponen:

- `components/user/RetailDownlineManager.tsx`
- Proxy frontend: `/api/me/retail/downlines`
- Backend: `/v1/me/retail/downlines`

Request tambah downline:

```json
{
  "email": "user@example.com",
  "nama": "Nama User",
  "password": "minimal8karakter",
  "role": "user"
}
```

Validasi:

- `email`, `nama`, dan `password` wajib diisi.
- Password minimal 8 karakter.
- Role bawahan hanya boleh `user` atau `agent`.
- Actor `master` boleh membuat `agent` atau `user`.
- Actor `agent` hanya boleh membuat `user`.
- Email harus unik.

Saat akun berhasil dibuat:

- Data masuk ke `public.member`.
- Dompet member dibuat di `public.dompet_member` dengan saldo awal `0`.
- Relasi `retail_agent_id` dan/atau `retail_master_id` diisi sesuai actor pembuat.

### Dari Dashboard Admin

Admin dapat mengelola akun retail melalui halaman master member.

Endpoint backend:

- `POST /v1/admin/users/hierarchy/preview`
- `POST /v1/admin/users/hierarchy/apply`
- Alias lama juga tersedia di `/v1/admin/members/hierarchy/preview` dan `/v1/admin/members/hierarchy/apply`

Fungsi admin:

- Mengubah role akun ke `user`, `agent`, atau `master`.
- Memasang agent dan master untuk akun retail.
- Preview dampak assignment sebelum diterapkan.
- Apply komisi historis bila `applyHistorical` diaktifkan.

## Alur Order dan Snapshot Jaringan

Saat user retail membuat order aplikasi:

1. Backend membaca role pembeli dari `public.member`.
2. Fee harga aplikasi dipilih berdasarkan role pembeli:
   - `master` memakai `fee_master`.
   - `agent` memakai `fee_agent`.
   - role lain retail memakai `fee_user`.
3. Backend menyimpan snapshot jaringan ke `public.app_order`:
   - `retail_agent_id_snapshot`
   - `retail_master_id_snapshot`
   - `fee_agent_snapshot`
   - `fee_master_snapshot`
   - `fee_user_snapshot`
4. Order dibuat dengan status awal `pending_payment`.

Kode sumber:

- Create order: `pulsa-be/internal/service/app_order_service_create.go`
- Insert order: `pulsa-be/internal/repository/app_order_repository_core.go`

Snapshot penting karena komisi dihitung berdasarkan jaringan saat order dibuat, bukan jaringan terbaru setelah order berubah status.

## Alur Komisi Retail

Komisi retail diproses saat order aplikasi menjadi `success`.

Kode sumber:

- Apply komisi: `pulsa-be/internal/repository/retail_repository_commission_apply.go`
- Trigger dari callback provider:
  - `provider_callback_service_yuscom_app_order.go`
  - `provider_callback_service_gemilang_app.go`

Syarat komisi masuk:

- `app_order.status = success`
- `buyer_type = user`
- `member_id` order valid
- Ada snapshot `retail_agent_id_snapshot` dan/atau `retail_master_id_snapshot`
- Nominal komisi target lebih dari `0`

Pembagian komisi:

| Role pembeli | Penerima komisi |
| --- | --- |
| `user` punya agent dan master | Agent menerima `retail_agent_commission_rp`, master menerima `retail_master_commission_rp`. |
| `user` hanya punya master | Master menerima `retail_master_commission_rp`. |
| `agent` punya master | Master menerima `retail_master_commission_rp`. |
| `master` | Tidak ada upline retail, jadi tidak ada komisi jaringan. |

Saat komisi masuk:

- Baris dibuat di `public.retail_commission_ledger`.
- Saldo penerima di `public.dompet_member` bertambah.
- Mutasi dibuat di `public.mutasi_dompet` dengan alasan `RETAIL_COMMISSION`.

Idempotensi:

- Index unik `retail_commission_ledger_member_order_uidx` mencegah komisi ganda untuk kombinasi `member_id` dan `source_app_order_id`.

Reversal:

- Jika order yang sudah sukses kemudian direfund/gagal, komisi dapat dibalik.
- Saldo penerima didebit kembali.
- Mutasi dibuat dengan alasan `RETAIL_COMMISSION_REVERSAL`.
- Amount ledger diset menjadi `0` dan note diberi keterangan reversal.

## Komisi Historis

Admin dapat menerapkan komisi historis saat memperbaiki/menata jaringan lewat endpoint hierarchy apply.

Sumber:

- `UserService.ApplyHierarchyAssignment`
- `RetailRepository.ApplyHistoricalCommission`

Cara kerja:

- Sistem menghitung order sukses historis milik member target.
- Komisi dihitung dari nilai komisi akun agent/master yang dipasang.
- Ledger dibuat dengan note `retail commission historical backfill`.
- Saldo penerima ditambah total backfill.
- Mutasi dompet dibuat dengan alasan `RETAIL_COMMISSION_BACKFILL`.

## Ringkasan Komisi di UI

Halaman komisi retail memakai:

- Komponen: `components/user/RetailCommissionClient.tsx`
- Summary: `/api/me/retail/commissions/summary`
- Riwayat: `/api/me/retail/commissions`

Field summary:

| Field | Arti |
| --- | --- |
| `total_earned` | Total komisi yang tercatat di ledger. |
| `available_saldo` | Saldo dompet member saat ini. |
| `total_pending_withdraw` | Total withdraw retail yang masih pending. |
| `total_approved_withdraw` | Total withdraw retail yang sudah approved. |
| `total_rejected_withdraw` | Total withdraw retail yang rejected. |

## Alur Withdraw Fee Retail

User retail dapat mengajukan pencairan komisi lewat:

- Komponen: `components/user/RetailWithdrawClient.tsx`
- Proxy frontend: `/api/me/retail/withdraw-requests`
- Backend: `/v1/me/retail/withdraw-requests`

Request:

```json
{
  "amount": 50000,
  "bank_name": "BCA",
  "account_name": "Nama Pemilik",
  "account_number": "1234567890",
  "note": "Catatan opsional"
}
```

Validasi:

- `amount` harus lebih dari `0`.
- `bank_name`, `account_name`, dan `account_number` wajib diisi.
- Saldo dompet harus cukup.

Saat withdraw dibuat:

- Ref ID dibuat dengan format `RWD-YYYYMMDDHHMMSS-XXXX`.
- Saldo dompet langsung dikurangi/di-hold.
- Mutasi dompet dibuat dengan alasan `RETAIL_WITHDRAW_HOLD`.
- Request masuk ke `public.retail_withdraw_request` dengan status `pending`.

## Proses Withdraw oleh Admin/Operator Wallet

Halaman admin:

- Admin: `/dashboard/admin/retail-withdraws`
- Wallet: `/dashboard/wallet/retail-withdraws`
- Komponen: `components/dashboard/RetailWithdrawRequestsPage.tsx`

Endpoint backend:

- `GET /v1/admin/retail/withdraw-requests`
- `POST /v1/admin/retail/withdraw-requests/approve?id={id}`
- `POST /v1/admin/retail/withdraw-requests/reject?id={id}`

Akses:

- `admin`
- `operator_wallet`

Approve withdraw:

```json
{
  "bank_id": 1,
  "fee": 2500,
  "note": "Diproses via BCA"
}
```

Saat approve:

- Request harus status `pending`.
- Bank sumber wajib aktif.
- Saldo bank sumber harus cukup untuk `amount + fee`.
- Saldo bank didebit.
- Mutasi bank dibuat dengan alasan `RETAIL_WITHDRAW_APPROVE`.
- Request menjadi `approved`.
- Saldo member tidak dikurangi lagi karena sudah di-hold saat request dibuat.

Reject withdraw:

```json
{
  "reason": "Data rekening tidak valid"
}
```

Saat reject:

- Request harus status `pending`.
- Saldo member dikembalikan sebesar `amount`.
- Mutasi dompet dibuat dengan alasan `RETAIL_WITHDRAW_REJECT_REFUND`.
- Request menjadi `rejected`.

## Endpoint Retail

### Member Retail

| Method | Endpoint backend | Fungsi |
| --- | --- | --- |
| `GET` | `/v1/me/retail/downlines` | Melihat jaringan downline retail. |
| `POST` | `/v1/me/retail/downlines` | Membuat downline retail. |
| `GET` | `/v1/me/retail/commissions` | Melihat riwayat komisi retail. |
| `GET` | `/v1/me/retail/commissions/summary` | Melihat ringkasan saldo dan komisi. |
| `GET` | `/v1/me/retail/withdraw-requests` | Melihat riwayat withdraw sendiri. |
| `POST` | `/v1/me/retail/withdraw-requests` | Mengajukan withdraw komisi retail. |

### Admin/Operator Wallet

| Method | Endpoint backend | Fungsi |
| --- | --- | --- |
| `GET` | `/v1/admin/retail/withdraw-requests` | List request withdraw retail. |
| `POST` | `/v1/admin/retail/withdraw-requests/approve?id={id}` | Approve withdraw retail. |
| `POST` | `/v1/admin/retail/withdraw-requests/reject?id={id}` | Reject withdraw retail. |
| `POST` | `/v1/admin/users/hierarchy/preview` | Preview assignment jaringan retail/H2H. |
| `POST` | `/v1/admin/users/hierarchy/apply` | Apply assignment jaringan retail/H2H. |

### Proxy Frontend

| Endpoint frontend | Backend tujuan |
| --- | --- |
| `/api/me/retail/downlines` | `/v1/me/retail/downlines` |
| `/api/me/retail/commissions` | `/v1/me/retail/commissions` |
| `/api/me/retail/commissions/summary` | `/v1/me/retail/commissions/summary` |
| `/api/me/retail/withdraw-requests` | `/v1/me/retail/withdraw-requests` |
| `/api/admin/retail/withdraw-requests` | `/v1/admin/retail/withdraw-requests` |
| `/api/admin/retail/withdraw-requests/approve` | `/v1/admin/retail/withdraw-requests/approve` |
| `/api/admin/retail/withdraw-requests/reject` | `/v1/admin/retail/withdraw-requests/reject` |

## Tabel Database

### `public.member`

Kolom retail penting:

- `role`
- `retail_agent_id`
- `retail_master_id`
- `retail_agent_commission_rp`
- `retail_master_commission_rp`

### `public.app_order`

Kolom snapshot retail:

- `buyer_type`
- `buyer_role`
- `fee_user_snapshot`
- `fee_agent_snapshot`
- `fee_master_snapshot`
- `retail_agent_id_snapshot`
- `retail_master_id_snapshot`

### `public.retail_commission_ledger`

Menyimpan ledger komisi retail.

Kolom penting:

- `member_id`: penerima komisi
- `source_member_id`: pembeli/sumber transaksi
- `source_app_order_id`: order aplikasi sumber komisi
- `invoice_id`
- `level_name`: `agent` atau `master`
- `amount`
- `note`
- `created_at`

### `public.retail_withdraw_request`

Menyimpan request withdraw fee retail.

Kolom penting:

- `member_id`
- `amount`
- `bank_name`
- `account_name`
- `account_number`
- `status`: `pending`, `approved`, `rejected`
- `note`
- `reject_reason`
- `ref_id`
- `processed_by`
- `processed_at`

### `public.dompet_member`

Saldo komisi dan saldo member disimpan di tabel ini.

Mutasi yang terkait retail:

| Alasan mutasi | Arah | Keterangan |
| --- | --- | --- |
| `RETAIL_COMMISSION` | `CREDIT` | Komisi order sukses. |
| `RETAIL_COMMISSION_REVERSAL` | `DEBIT` | Pembalikan komisi saat order direfund/gagal. |
| `RETAIL_COMMISSION_BACKFILL` | `CREDIT` | Komisi historis dari assignment jaringan. |
| `RETAIL_WITHDRAW_HOLD` | `DEBIT` | Saldo ditahan saat user mengajukan withdraw. |
| `RETAIL_WITHDRAW_REJECT_REFUND` | `CREDIT` | Saldo dikembalikan saat withdraw ditolak. |

### `public.mutasi_bank`

Mutasi bank saat withdraw retail di-approve:

- `arah = DEBIT`
- `alasan = RETAIL_WITHDRAW_APPROVE`
- `jumlah = amount + fee`
- `meta.type = retail_withdraw_approve`

## File Penting

Backend:

- `pulsa-be/internal/service/retail_service.go`
- `pulsa-be/internal/repository/retail_repository_core.go`
- `pulsa-be/internal/repository/retail_repository_commission_apply.go`
- `pulsa-be/internal/repository/retail_repository_commission_query.go`
- `pulsa-be/internal/repository/retail_repository_withdraw.go`
- `pulsa-be/internal/controller/retail_controller.go`
- `pulsa-be/internal/controller/retail_admin_controller.go`
- `pulsa-be/internal/router/retail.go`
- `pulsa-be/internal/router/retail_admin.go`

Frontend:

- `pulsakilat-fe/components/user/RetailDownlineManager.tsx`
- `pulsakilat-fe/components/user/RetailCommissionClient.tsx`
- `pulsakilat-fe/components/user/RetailWithdrawClient.tsx`
- `pulsakilat-fe/components/dashboard/RetailWithdrawRequestsPage.tsx`
- `pulsakilat-fe/lib/memberRoles.ts`

Migrasi SQL:

- `pulsa-be/sql/20260325_add_member_hierarchy_columns.sql`
- `pulsa-be/sql/20260325_add_member_commission_columns.sql`
- `pulsa-be/sql/20260325_add_retail_commission_and_withdraw.sql`
- `pulsa-be/sql/20260715_fix_retail_schema.sql`

## Checklist Setup

1. Pastikan migrasi retail sudah dijalankan.
2. Pastikan role retail memakai salah satu dari `user`, `agent`, atau `master`.
3. Pastikan akun agent/master memiliki nilai komisi:
   - `retail_agent_commission_rp`
   - `retail_master_commission_rp`
4. Pastikan user/agent sudah punya upline:
   - user di bawah agent: isi `retail_agent_id`, opsional/otomatis `retail_master_id`
   - user langsung di bawah master: isi `retail_master_id`
   - agent di bawah master: isi `retail_master_id`
5. Pastikan order aplikasi sukses memanggil proses `ApplyCommissionForOrder`.
6. Untuk withdraw, pastikan bank sumber aktif dan saldo bank cukup.

## Troubleshooting

### Downline tidak muncul

- Cek role actor. Hanya `master` dan `agent` yang bisa melihat/mengelola downline.
- Cek kolom `retail_agent_id` dan `retail_master_id` pada akun downline.
- Master melihat akun dengan `retail_master_id = master_id` atau `retail_agent_id = master_id`.
- Agent melihat akun dengan `retail_agent_id = agent_id`.

### Komisi tidak masuk

- Cek `app_order.status` harus `success`.
- Cek `buyer_type` harus `user`.
- Cek snapshot `retail_agent_id_snapshot` atau `retail_master_id_snapshot`.
- Cek nilai komisi penerima lebih dari `0`.
- Cek apakah ledger sudah pernah dibuat untuk order yang sama.

### Withdraw gagal dibuat

- Cek saldo di `dompet_member`.
- Cek `amount > 0`.
- Cek data rekening lengkap.

### Withdraw gagal di-approve

- Cek status request masih `pending`.
- Cek `bank_id` aktif.
- Cek saldo bank cukup untuk `amount + fee`.
- Cek actor punya role `admin` atau `operator_wallet`.
