package repository

import (
	"context"
	"fmt"
)

func (r *AuditRepository) ResolveProviderWalletMissingDebitBulk(
	ctx context.Context,
	actorID int64,
	transaksiProviderIDs []int64,
) (*BulkResolveProviderWalletMissingDebitResult, error) {
	seen := make(map[int64]struct{}, len(transaksiProviderIDs))
	ids := make([]int64, 0, len(transaksiProviderIDs))
	for _, id := range transaksiProviderIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("transaksi_provider_ids wajib diisi")
	}

	out := &BulkResolveProviderWalletMissingDebitResult{
		ProcessedIDs: make([]int64, 0, len(ids)),
	}
	for _, id := range ids {
		res, err := r.ResolveProviderWalletMissingDebit(ctx, actorID, id)
		if err != nil {
			return nil, err
		}
		out.Processed++
		out.ProcessedIDs = append(out.ProcessedIDs, id)
		if res.AlreadyResolved {
			out.AlreadyCount++
		} else if res.Resolved {
			out.ResolvedCount++
		}
	}
	return out, nil
}

func (r *AuditRepository) IgnoreProviderWalletMissingDebitBulk(
	ctx context.Context,
	actorID int64,
	transaksiProviderIDs []int64,
	note string,
) (*BulkIgnoreProviderWalletMissingDebitResult, error) {
	seen := make(map[int64]struct{}, len(transaksiProviderIDs))
	ids := make([]int64, 0, len(transaksiProviderIDs))
	for _, id := range transaksiProviderIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("transaksi_provider_ids wajib diisi")
	}

	out := &BulkIgnoreProviderWalletMissingDebitResult{
		ProcessedIDs: make([]int64, 0, len(ids)),
	}
	for _, id := range ids {
		if err := r.IgnoreProviderWalletMissingDebit(ctx, actorID, id, note); err != nil {
			return nil, err
		}
		out.Processed++
		out.ProcessedIDs = append(out.ProcessedIDs, id)
	}
	return out, nil
}
