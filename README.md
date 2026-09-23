# Isiloka

Kebutuhan harian, dekat dan tenang. Isi hari, dari sini.

Salinan mandiri dari PulsaKilat; kode Go, SQL, kontrak API, login, checkout,
refid, callback, saldo, dan aturan status transaksi dipertahankan.
Logo, banner, nama tampilan, metadata, font dan warna memakai identitas Isiloka.
Nama folder pulsa-fe/pulsa-be, modul Go pulsa2, SKU dan ID provider sengaja tetap kompatibel.

## Jalankan Frontend

Di folder pulsa-fe, jalankan npm ci lalu npm run dev.
Alamat lokal: http://localhost:3102
.env.local berisi koneksi lokal dan secret sesi baru, tanpa kredensial produksi.
Tanpa backend/database, tampilan dapat dibuka tetapi login dan pembelian belum tersedia.

## Hubungkan Backend

Konfigurasikan pulsa-be/.env berdasarkan .env.example. Gunakan database isiloka
dan akun H2HR yang ditetapkan untuk aplikasi ini. Jalankan migrasi sesuai dokumentasi
proyek asal, lalu jalankan go run . dari pulsa-be. Port backend: 8102.
Jangan menjalankan seluruh seed historis secara membabi buta: sebagian berisi katalog lama.
Sinkronkan katalog Pulsa24Jam H2HR sebelum melakukan pembelian.

Domain produksi belum ditetapkan. Ubah NEXTAUTH_URL, NEXT_PUBLIC_SITE_URL,
API_BASE, NEXT_PUBLIC_API_BASE dan CORS_ORIGINS saat domain tersedia.
API_BASE adalah alamat backend yang dapat diakses server frontend; NEXT_PUBLIC_API_BASE
harus dapat diakses browser jika dipakai oleh client. Di produksi gunakan HTTPS.
Daftarkan URL callback /api/v1/webhooks/pulsa24jam beserta token baru ke H2HR.
Endpoint upstream tetap https://api.pulsa24jam.net/v2/trx.

Gunakan database dan akun/provider namespace yang terpisah antar aplikasi:
refid berbasis ID order dari kode asal dapat bertabrakan jika enam database memakai
satu akun H2HR tanpa pengaturan namespace dari provider.
Rekening deposit dan kontak bantuan bawaan sumber tetap dipertahankan; tinjau
kepemilikannya sebelum rilis. Data iklan/nama dari database perlu disesuaikan tersendiri.

## Verifikasi dan Rilis

Frontend: npm run build. Backend: go test ./internal/service ./internal/helper ./internal/provider ./internal/repository.
Periksa CLONE-MANIFEST.json untuk hash sumber dan salinan.
Workflows dan skrip deploy produksi asal tidak disalin agar tidak menimpa PulsaKilat.
Jangan build di direktori runtime release. Buat release baru setelah build berhasil,
lalu arahkan service khusus isiloka ke release tersebut.
Jangan mengubah status transaksi manual menjadi sukses: status final berasal dari provider.
