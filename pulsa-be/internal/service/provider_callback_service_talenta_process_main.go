package service

import "context"

func (s *ProviderCallbackService) ProcessTalentaCallback(ctx context.Context, rawBody string, payload map[string]any) (int, map[string]any) {
	data := extractTalentaCallbackData(payload)

	if data.refid == "" {
		return s.handleTalentaMissingRef(ctx, rawBody, payload, data)
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, data.refid, "talentapay")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		return s.handleTalentaMissingRow(ctx, rawBody, payload, data)
	}

	s.updateTalentaProviderResult(ctx, row, payload, data)

	trx, err := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": data.refid}
	}

	return s.finalizeTalentaMemberTrx(ctx, row, trx, data)
}
