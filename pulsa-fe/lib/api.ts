/**
 * Generic API Fetch Helper
 * Mendukung caching dengan revalidate dan cache: 'no-store'
 */

type FetchOptions = {
  revalidate?: number; // Detik, untuk ISR (Incremental Static Regeneration)
  cache?: 'no-store' | 'force-cache';
  headers?: Record<string, string>;
  method?: string;
  body?: BodyInit | null;
};

type ApiResponse<T> = {
  ok?: boolean;
  items?: T[];
  item?: T;
  [key: string]: unknown;
};

/**
 * Fetch API dengan opsi caching
 * @param endpoint - URL endpoint (relative atau absolute)
 * @param options - Opsi fetch (revalidate, cache, headers)
 * @returns Data dari API
 *
 * @example
 * // Dengan revalidate (ISR - cache 1 jam)
 * const brands = await fetchAPI<Brand[]>('/v1/app/brand?kategori_id=1', { revalidate: 3600 });
 *
 * @example
 * // Tanpa cache (realtime)
 * const transactions = await fetchAPI<Transaction[]>('/v1/app/transactions', { cache: 'no-store' });
 */
export async function fetchAPI<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<T[]> {
  const base = (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || 'http://127.0.0.1:8083').replace(/\/+$/, '');
  const url = endpoint.startsWith('http') ? endpoint : `${base}${endpoint}`;

  const fetchOptions: RequestInit = {
    method: options.method,
    body: options.body,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  };

  // Konfigurasi caching
  if (options.cache === 'no-store') {
    // Realtime - tidak cache sama sekali
    fetchOptions.cache = 'no-store';
  } else if (options.revalidate !== undefined) {
    // ISR - cache dengan revalidate otomatis
    fetchOptions.next = {
      revalidate: options.revalidate,
    };
  } else {
    // Default: cache 60 detik
    fetchOptions.next = {
      revalidate: 60,
    };
  }

  try {
    const res = await fetch(url, fetchOptions);

    if (!res.ok) {
      console.error(`API Error [${res.status}]:`, url);
      return [];
    }

    const json = (await res.json().catch(() => ({}))) as ApiResponse<T>;

    // Support untuk both { items: [...] } dan { item: {...} }
    if (json.items && Array.isArray(json.items)) {
      return json.items;
    }

    if (json.item) {
      return [json.item] as T[];
    }

    return [];
  } catch (error) {
    console.error('Fetch API Error:', error);
    return [];
  }
}

/**
 * Fetch API dengan response object (untuk flexibility)
 * Gunakan ini kalau perlu access `ok`, `error`, atau fields lain
 */
export async function fetchAPIFull<T>(
  endpoint: string,
  options: FetchOptions = {}
): Promise<ApiResponse<T>> {
  const base = (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || 'http://127.0.0.1:8083').replace(/\/+$/, '');
  const url = endpoint.startsWith('http') ? endpoint : `${base}${endpoint}`;

  const fetchOptions: RequestInit = {
    method: options.method,
    body: options.body,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  };

  if (options.cache === 'no-store') {
    fetchOptions.cache = 'no-store';
  } else if (options.revalidate !== undefined) {
    fetchOptions.next = { revalidate: options.revalidate };
  } else {
    fetchOptions.next = { revalidate: 60 };
  }

  try {
    const res = await fetch(url, fetchOptions);
    const json = (await res.json().catch(() => ({}))) as ApiResponse<T>;

    if (!res.ok) {
      return {
        ok: false,
        error: json.error || `API Error: ${res.statusText}`,
      };
    }

    return {
      ok: true,
      ...json,
    };
  } catch (error) {
    console.error('Fetch API Error:', error);
    return {
      ok: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

/**
 * Cache strategy recommendations:
 *
 * ✅ revalidate: 3600 (1 jam)
 *   - Kategori, Brand, Static content
 *
 * ✅ revalidate: 300 (5 menit)
 *   - Harga produk, Stok
 *
 * ✅ cache: 'no-store'
 *   - Transaksi, Order, User-specific data
 *
 * ✅ revalidate: 60 (default)
 *   - General purpose
 */
