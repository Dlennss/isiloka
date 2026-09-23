/**
 * Transaction & Order API Calls
 * Semua fetch untuk transaksi, order, dan payment
 */

import { fetchAPI } from './api';

// ============================================
// TRANSACTIONS
// ============================================

/**
 * Get user transactions
 * Cache: tidak ada (realtime data)
 */
export async function getUserTransactions(token: string) {
  return fetchAPI('/v1/user/transactions', {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

/**
 * Get transaction by ID
 * Cache: tidak ada (status bisa berubah)
 */
export async function getTransactionById(id: string, token: string) {
  return fetchAPI(`/v1/user/transactions/${id}`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

// ============================================
// ORDERS
// ============================================

/**
 * Get user orders with optional filters and pagination
 * Cache: tidak ada (realtime)
 * @param token - Backend token for authorization
 * @param status - Optional order status filter
 * @param limit - Items per page (default: 10)
 * @param offset - Pagination offset (default: 0)
 */
export async function getUserOrders(
  token: string,
  status?: string,
  limit: number = 10,
  offset: number = 0
) {
  const qs = new URLSearchParams();
  if (status) qs.set('status', status);
  qs.set('limit', String(limit));
  qs.set('offset', String(offset));

  return fetchAPI(
    `/v1/app/me/orders${qs.toString() ? `?${qs.toString()}` : ''}`,
    {
      cache: 'no-store',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
}

/**
 * Get active orders only
 * Cache: tidak ada (realtime)
 */
export async function getActiveOrders(token: string) {
  return fetchAPI('/v1/app/me/orders/active', {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

/**
 * Get order by ID / Invoice (for both authenticated users and guests)
 * Cache: tidak ada (status bisa berubah tiap detik)
 * @param invoiceId - Invoice ID to lookup
 * @param token - Optional token for authenticated users
 */
export async function getOrderByInvoice(invoiceId: string, token?: string, guestEmail?: string, guestPhone?: string) {
  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  if (guestEmail) {
    headers["X-Guest-Email"] = guestEmail;
  }
  if (guestPhone) {
    headers["X-Guest-Phone"] = guestPhone;
  }

  return fetchAPI(`/v1/app/orders/${encodeURIComponent(invoiceId)}`, {
    cache: 'no-store',
    headers,
  });
}

// ============================================
// WALLET
// ============================================

/**
 * Get user wallet balance
 * Cache: 2 menit (saldo penting tapi tidak perlu instant)
 */
export async function getUserWallet(token: string) {
  return fetchAPI('/v1/user/wallet', {
    revalidate: 120, // 2 menit
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

/**
 * Get wallet transaction history
 * Cache: 5 menit
 */
export async function getWalletHistory(token: string) {
  return fetchAPI('/v1/user/wallet/history', {
    revalidate: 300, // 5 menit
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

// ============================================
// PAYMENT
// ============================================

/**
 * Get payment status / QRIS for order
 * Cache: tidak ada (status bisa berubah)
 */
export async function getPaymentStatus(orderId: string, token: string) {
  return fetchAPI(`/v1/user/orders/${orderId}/payment`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}
