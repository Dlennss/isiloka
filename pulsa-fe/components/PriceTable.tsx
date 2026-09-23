'use client';

import * as React from 'react';
import { getErrorMessage } from '@/lib/error';

type Product = {
  id?: number | string;
  code?: string;
  name?: string;
  operator?: string;
  brand?: string;
  category?: string;
  face_value?: number;
  provider_price?: number;
  price?: number;
  sell_price?: number;
};

const API_BASE =
  (typeof window !== 'undefined' && process.env.NEXT_PUBLIC_API_BASE) || '';

function asCurrency(n: number | undefined) {
  if (typeof n !== 'number' || Number.isNaN(n)) return '-';
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(n);
}

type Row = {
  id: string;
  code: string;
  name: string;
  operator: string;
  price?: number;
  face?: number;
  category: string;
};

export default function PriceTable() {
  const [data, setData] = React.useState<Product[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [q, setQ] = React.useState('');
  const [operator, setOperator] = React.useState<string>('Semua');
  const [selected, setSelected] = React.useState<Row | null>(null);

  const fetchData = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const url = `${API_BASE}/instant`;
      const res = await fetch(url, { method: 'GET', cache: 'no-store' });
      const json = await res.json();
      if (!res.ok) throw new Error(json?.message || 'Gagal memuat data');
      const arr: Product[] = Array.isArray(json?.data) ? json.data
                        : Array.isArray(json) ? json
                        : Array.isArray(json?.products) ? json.products
                        : [];
      setData(arr);
    } catch (e: unknown) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => { fetchData(); }, [fetchData]);

  // normalisasi → Row
  const rows = React.useMemo<Row[]>(() => {
    return data.map((p) => {
      const price =
        typeof p.price === 'number' ? p.price :
        typeof p.sell_price === 'number' ? p.sell_price :
        typeof p.provider_price === 'number' ? Math.round(p.provider_price * 1.02) :
        undefined;

      return {
        id: String(p.id ?? p.code ?? p.name),
        code: String(p.code ?? ''),
        name: String(p.name ?? ''),
        operator: String((p.operator || p.brand || '').trim()),
        category: String(p.category || ''),
        face: p.face_value,
        price,
      };
    });
  }, [data]);

  // daftar kategori (operator/brand/category digabung agar praktis)
  const categories = React.useMemo(() => {
    const s = new Set<string>();
    s.add('Semua');
    for (const r of rows) {
      const base = r.operator || r.category || 'Lainnya';
      s.add(base);
    }
    return Array.from(s).filter(Boolean).sort((a, b) => a.localeCompare(b, 'id'));
  }, [rows]);

  // filter teks + kategori
  const filtered = React.useMemo(() => {
    const ql = q.trim().toLowerCase();
    return rows.filter((r) => {
      const inCat = operator === 'Semua' || r.operator === operator || r.category === operator;
      const textHit =
        !ql ||
        r.name.toLowerCase().includes(ql) ||
        r.operator.toLowerCase().includes(ql) ||
        r.category.toLowerCase().includes(ql) ||
        r.code.toLowerCase().includes(ql);
      return inCat && textHit;
    });
  }, [rows, q, operator]);

  return (
    <div>
      {/* filter bar tetap */}
      <div className="filterbar" style={{ display: 'grid', gridTemplateColumns: '1fr 220px 140px', gap: 12 }}>
        <input
          placeholder="Cari produk / kode / operator"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <select value={operator} onChange={(e) => setOperator(e.target.value)}>
          {categories.map((c) => (
            <option key={c} value={c}>{c === 'Semua' ? 'Semua Kategori' : c}</option>
          ))}
        </select>
        <button onClick={fetchData} disabled={loading}>
          {loading ? 'Memuat...' : 'Refresh'}
        </button>
      </div>

      {error && <div className="alert error" style={{ marginTop: 12 }}>{error}</div>}

      {/* GRID KARTU – kode TIDAK ditampilkan di sini */}
      <div style={{ marginTop: 16 }}>
        {loading ? (
          <div className="muted" style={{ textAlign: 'center' }}>Memuat data…</div>
        ) : filtered.length === 0 ? (
          <div className="muted" style={{ textAlign: 'center' }}>Tidak ada data</div>
        ) : (
          <div
            className="grid"
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))',
              gap: 12,
            }}
          >
            {filtered.map((r) => (
              <div
                key={r.id}
                className="card"
                style={{ padding: 12, borderRadius: 12 }}
              >
                <div className="muted" style={{ fontSize: 12 }}>{r.operator || r.category || '-'}</div>
                <div style={{ fontWeight: 600, marginTop: 4 }}>{r.name}</div>
                {r.face ? <div className="muted" style={{ fontSize: 12, marginTop: 2 }}>{asCurrency(r.face)}</div> : null}
                <div style={{ marginTop: 6, fontWeight: 600 }}>{asCurrency(r.price)}</div>
                <div style={{ marginTop: 10, display: 'flex', gap: 8 }}>
                  <button
                    onClick={() => setSelected(r)}
                    className="btn primary"
                  >
                    Pesan
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* MODAL: tampilkan kode hanya saat pemesanan */}
      {selected && (
        <div
          className="modal-backdrop"
          onClick={() => setSelected(null)}
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0,0,0,.5)',
            display: 'grid',
            placeItems: 'center',
            zIndex: 50,
          }}
        >
          <div
            className="card"
            onClick={(e) => e.stopPropagation()}
            style={{ width: 420, maxWidth: '90vw', borderRadius: 14, padding: 16 }}
          >
            <h3 style={{ margin: 0 }}>Konfirmasi Pesanan</h3>
            <div className="muted" style={{ marginTop: 6 }}>{selected.operator || selected.category || '-'}</div>
            <div style={{ marginTop: 8, fontWeight: 600 }}>{selected.name}</div>
            <div style={{ marginTop: 4 }}>Harga: <strong>{asCurrency(selected.price)}</strong></div>
            <div style={{ marginTop: 4 }}>Kode Produk: <code>{selected.code}</code></div>

            <div style={{ marginTop: 14, display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <button onClick={() => setSelected(null)} className="btn">Batal</button>
              <button className="btn success">Lanjutkan Pembayaran</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
