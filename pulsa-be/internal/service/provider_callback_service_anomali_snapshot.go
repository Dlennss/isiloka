package service

import (
	"context"
	"encoding/json"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) insertAnomaliSnapshot(ctx context.Context, provider, refID, sumber string, saldo int64, raw any) {
	if s == nil || s.repo == nil || saldo <= 0 {
		return
	}
	rawJSON, _ := json.Marshal(raw)
	if err := s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
		Provider:      provider,
		SaldoProvider: saldo,
		RefID:         strings.TrimSpace(refID),
		Sumber:        strings.TrimSpace(sumber),
		RawJSON:       rawJSON,
	}); err != nil {
		helper.AppendProviderServiceLog("provider_anomali.log", "insert snapshot anomali gagal provider=%s refid=%s saldo=%d err=%v", provider, refID, saldo, err)
	}
}
