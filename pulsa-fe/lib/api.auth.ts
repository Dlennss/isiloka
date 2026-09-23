/**
 * Authentication API Calls
 * Login, register, user profile, logout
 */

import { fetchAPIFull } from './api';
import type { RetailCommissionSummary, UserSession, UserProfile } from '@/components/user/types';

export type AgentCreditApplication = {
  id: number;
  member_id: number;
  member_name: string;
  member_email: string;
  member_phone: string;
  marketing_id?: number;
  analyst_id?: number;
  agent_name?: string;
  agent_email?: string;
  store_name?: string;
  requested_amount: number;
  approved_amount: number;
  status: string;
  applicant_data?: Record<string, unknown>;
  document_data?: Record<string, unknown>;
  has_agent_signature?: boolean;
  agent_signature_data?: string;
  agent_signature_at?: string;
  marketing_note?: string;
  analyst_note?: string;
  analyst_recommendation?: string;
  analyst_recommended_amount?: number;
  recommended_amount?: number;
  loan_status?: string;
  outstanding_amount?: number;
  credit_available_amount?: number;
  paid_amount?: number;
  payment_count?: number;
  payments?: AgentCreditPayment[];
  credit_level_code?: string;
  credit_level_name?: string;
  credit_needs_repair?: boolean;
  qualified_paid_total?: number;
  credit_limit_amount?: number;
	operator_revision_documents?: string[];
	operator_revision_required?: boolean;
	operator_revision_requested_at?: string;
	operator_revision_resolved_at?: string;
  loan_approved_at?: string;
  loan_due_date?: string;
  last_transaction_at?: string;
  created_at: string;
  updated_at: string;
};

export type AgentCreditPayment = {
  id: number;
  loan_id: number;
  application_id: number;
  member_id: number;
  amount: number;
  due_date: string;
  paid_at: string;
  days_late: number;
  status: string;
  note?: string;
  payment_method?: string;
  payment_proof?: {
    name?: string;
    type?: string;
    size?: number;
    data_url?: string;
  };
};

// ============================================
// USER PROFILE
// ============================================

/**
 * Get current user profile
 * Cache: tidak ada (user-specific, sensitive)
 */
export async function getUserProfile(token: string): Promise<UserProfile | null> {
  const response = await fetchAPIFull<UserProfile>('/v1/me/profile', {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (response.ok && response.profile) {
    return response.profile as UserProfile;
  }

  return null;
}

/**
 * Update user profile
 * Cache: tidak ada
 */
export async function updateUserProfile(
  token: string,
  data: Partial<UserSession>
) {
  const response = await fetchAPIFull<UserProfile>('/v1/me/profile', {
    method: 'PATCH',
    body: JSON.stringify(data),
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  return response;
}

// ============================================
// SECURITY
// ============================================

/**
 * Get user IP whitelist
 * Cache: 1 hari (jarang berubah)
 */
export async function getUserIPWhitelist(token: string) {
  return fetchAPIFull('/v1/user/security/ip-whitelist', {
    revalidate: 86400, // 1 hari
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function getRetailCommissionSummary(token: string): Promise<RetailCommissionSummary | null> {
  const response = await fetchAPIFull<RetailCommissionSummary>('/v1/me/retail/commissions/summary', {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (response.ok && response.item) {
    return response.item as RetailCommissionSummary;
  }

  return null;
}

export async function getAgentCreditApplications(token: string, limit = 50): Promise<AgentCreditApplication[]> {
  const response = await fetchAPIFull<AgentCreditApplication>(`/v1/master/agent-credit/applications?limit=${Math.min(200, Math.max(1, limit))}`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return Array.isArray(response.items) ? response.items : [];
}

export async function getMyAgentCreditApplications(token: string): Promise<AgentCreditApplication[]> {
  const response = await fetchAPIFull<AgentCreditApplication>('/v1/me/agent-credit/applications', {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return Array.isArray(response.items) ? response.items : [];
}
