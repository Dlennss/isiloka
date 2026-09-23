# SMB IRS9 Integration

Integrasi provider `smb` sekarang menyimpan `public.produk_provider_map.kode_provider` dalam bentuk kode asli provider, tanpa prefix mode.

Mode transaksi diinfer dari kodenya:

- `DANA`, `OVO`, `GOPAY`, `SHOPEEPAY`, `LINKAJA` untuk jalur `inq-pay` `jenis=5/6`
- `ELDN`, `GPYOPEN`, `SHPOPEN`, `ELOV`, `ELLI` untuk denom bebas direct `jenis=1`
- kode plain lain akan dianggap jalur `PPOB` biasa `jenis=5/6`

Parser SMB tetap kompatibel dengan format lama ber-prefix selama masa transisi.

Env yang wajib untuk uji lokal:

- `SMB_BASE_URL`
- `SMB_DIRECT_BASE_URL`
- `SMB_ID`
- `SMB_PIN`
- `SMB_USER`
- `SMB_PASSWORD`

Webhook callback yang dipakai provider:

- `/webhook/smb`
- `/v1/webhook/smb`

Nomor test wajib SMB:

- `+62 821-2430-7365`

Gunakan nomor itu untuk seluruh smoke test transaksi sampai ada instruksi lain.
