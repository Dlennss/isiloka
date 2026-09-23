# Cara Pakai Midtrans Sandbox

Proyek ini sudah disiapkan untuk Midtrans Sandbox dari backend.
Kamu hanya perlu mengganti `MIDTRANS_SERVER_KEY` dengan Server Key Sandbox asli dari dashboard Midtrans.

## 1. Buka Dashboard Sandbox

Buka:

```text
https://dashboard.sandbox.midtrans.com/
```

Login memakai akun Midtrans kamu.

Kalau dashboard menampilkan pilihan environment, pastikan pilih `Sandbox`, bukan `Production`.

## 2. Ambil Server Key

Di dashboard Midtrans:

```text
Settings > Access Keys > Server Key
```

Copy nilai `Server Key`.

Server Key Sandbox biasanya diawali:

```text
SB-Mid-server-
```

Pada sebagian tampilan/dashboard, key yang tampil di environment Sandbox bisa tetap diawali:

```text
Mid-server-
```

Kalau kiri atas dashboard sudah jelas `Environment Sandbox`, gunakan `Server Key` dari halaman itu.
Yang penting jangan mengambil key dari environment `Production`.

## 3. Tempel ke Backend

Buka file:

```text
pulsa-be/.env
```

Ubah bagian ini:

```env
MIDTRANS_SERVER_KEY=SB-Mid-server-GANTI_DENGAN_SERVER_KEY
MIDTRANS_IS_PRODUCTION=false
MIDTRANS_QRIS_ACQUIRER=gopay
```

Menjadi seperti ini:

```env
MIDTRANS_SERVER_KEY=server-key-asli-dari-dashboard-sandbox
MIDTRANS_IS_PRODUCTION=false
MIDTRANS_QRIS_ACQUIRER=gopay
```

Simpan file, lalu restart backend.

## 4. Jalankan Aplikasi

Jalankan backend dari folder `pulsa-be`:

```powershell
go run .
```

Jalankan frontend dari folder `pulsa-fe`:

```powershell
npm run dev
```

Frontend local memakai backend ini:

```env
NEXT_PUBLIC_API_BASE=http://127.0.0.1:8080
```

## 5. Buat Transaksi QRIS

Dari aplikasi, buat pembayaran QRIS seperti biasa.

Backend akan membuat transaksi ke Midtrans Sandbox. Kalau key benar, Midtrans akan mengembalikan data QRIS.

## 6. Simulasikan Pembayaran

Buka simulator:

```text
https://simulator.sandbox.midtrans.com/
```

Untuk QRIS langsung:

```text
https://simulator.sandbox.midtrans.com/v2/qris/index
```

Masukkan data QRIS/URL dari transaksi sandbox, lalu simulasikan pembayaran.

## Catatan Penting

Untuk testing:

```env
MIDTRANS_SERVER_KEY=Server Key dari dashboard Sandbox
MIDTRANS_IS_PRODUCTION=false
```

Untuk live:

```env
MIDTRANS_SERVER_KEY=Mid-server-...
MIDTRANS_IS_PRODUCTION=true
```

Jangan campur key sandbox dengan mode production, atau sebaliknya.
Backend sudah punya validasi untuk menolak konfigurasi yang salah.
