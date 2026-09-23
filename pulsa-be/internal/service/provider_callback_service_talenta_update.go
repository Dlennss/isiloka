package service

import (
	"context"
	"encoding/json"
	"strings"

	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) handleTalentaMissingRef(ctx context.Context, rawBody string, payload map[string]any, data talentaCallbackData) (int, map[string]any) {
	var qtyPtr *int64
	if data.qty > 0 {
		qtyPtr = &data.qty
	}
	dest := strings.TrimSpace(data.dest)
	var destPtr *string
	if dest != "" {
		destPtr = &dest
	}
	var pricePtr *int64
	if data.price > 0 {
		pricePtr = &data.price
	}
	_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{
		Provider:   "talentapay",
		RefID:      "",
		KodeRespon: &data.rcStr,
		Pesan:      &data.msg,
		Harga:      pricePtr,
		Tujuan:     destPtr,
		Qty:        qtyPtr,
		RawQuery:   "",
		RawBody:    rawBody,
		Payload:    payload,
	})
	s.insertAnomaliSnapshot(ctx, "talentapay", data.refid, "callback_talentapay_anomali", data.lastBalance, payload)
	return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
}

func (s *ProviderCallbackService) handleTalentaMissingRow(ctx context.Context, rawBody string, payload map[string]any, data talentaCallbackData) (int, map[string]any) {
	var qtyPtr *int64
	if data.qty > 0 {
		qtyPtr = &data.qty
	}
	dest := strings.TrimSpace(data.dest)
	var destPtr *string
	if dest != "" {
		destPtr = &dest
	}
	var pricePtr *int64
	if data.price > 0 {
		pricePtr = &data.price
	}
	_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{
		Provider:   "talentapay",
		RefID:      data.refid,
		KodeRespon: &data.rcStr,
		Pesan:      &data.msg,
		Harga:      pricePtr,
		Tujuan:     destPtr,
		Qty:        qtyPtr,
		RawQuery:   "",
		RawBody:    rawBody,
		Payload:    payload,
	})
	s.insertAnomaliSnapshot(ctx, "talentapay", data.refid, "callback_talentapay_anomali", data.lastBalance, payload)
	return 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid}
}

func (s *ProviderCallbackService) updateTalentaProviderResult(ctx context.Context, row *repository.ProviderTrxRefRow, payload map[string]any, data talentaCallbackData) {
	httpStatus := 200
	var noreffPtr *string
	if strings.TrimSpace(data.noreff) != "" {
		noreffPtr = &data.noreff
	}
	upd := repository.UpdateResult{
		HTTPStatus:  &httpStatus,
		KodeRespon:  &data.rcStr,
		Pesan:       &data.msg,
		Harga:       &data.price,
		NoReferensi: noreffPtr,
		ResponMentah: map[string]any{
			"raw": payload,
		},
	}
	if data.lastBalance > 0 {
		upd.SaldoTerakhir = &data.lastBalance
	}
	_ = s.repo.UpdateResult(ctx, row.ID, upd)

	if upd.SaldoTerakhir != nil && *upd.SaldoTerakhir > 0 {
		tmID := row.TransaksiMemberID
		tpID := row.ID
		rawJSON, _ := json.Marshal(payload)
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:            "talentapay",
			SaldoProvider:       *upd.SaldoTerakhir,
			RefID:               data.refid,
			TransaksiMemberID:   &tmID,
			TransaksiProviderID: &tpID,
			Sumber:              "callback_talentapay",
			RawJSON:             rawJSON,
		})
	}
}
