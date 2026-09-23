'use client';

import { useEffect } from 'react';
import { saveGuestTransaction } from '@/lib/guest-transaction-storage';

type Props = {
  invoiceId: string;
  guestEmail: string;
  guestPhone: string;
  dest: string;
  title: string;
  status: string;
  amount: number;
  serialNumber: string;
  createdAt: string;
  updatedAt: string;
};

/**
 * Invisible component yang menyimpan data transaksi guest ke localStorage
 * saat halaman detail dibuka. Mirip dengan upsertFromOrder di mobile.
 */
export function GuestTransactionDetailSaver(props: Props) {
  useEffect(() => {
    if (!props.invoiceId || !props.guestEmail || !props.guestPhone) return;
    saveGuestTransaction({
      invoice_id: props.invoiceId,
      guest_email: props.guestEmail,
      guest_phone: props.guestPhone,
      dest: props.dest,
      title: props.title || 'Transaksi Guest',
      created_at: props.createdAt,
      updated_at: props.updatedAt || props.createdAt,
      status: props.status,
      amount: props.amount,
      serial_number: props.serialNumber,
    });
  }, [props]);

  return null;
}
