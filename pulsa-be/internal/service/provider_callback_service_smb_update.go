package service

import (
	"context"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) handleSMBMissingRef(ctx context.Context, rawQuery string, data smbCallbackData) (int, map[string]any) {
	var pricePtr *int64
	if data.price > 0 {
		pricePtr = &data.price
	}
	_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "smb", RefID: "", KodeRespon: helper.PtrString(data.statusRaw), Pesan: helper.PtrString(data.msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: "", Payload: data.payload})
	return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
}

func (s *ProviderCallbackService) handleSMBMissingRow(ctx context.Context, rawQuery string, data smbCallbackData) (int, map[string]any) {
	var pricePtr *int64
	if data.price > 0 {
		pricePtr = &data.price
	}
	_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "smb", RefID: data.refid, KodeRespon: helper.PtrString(data.statusRaw), Pesan: helper.PtrString(data.msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: "", Payload: data.payload})
	return 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid}
}

func (s *ProviderCallbackService) updateSMBProviderResult(ctx context.Context, row *repository.ProviderTrxRefRow, data smbCallbackData) error {
	httpStatus := 200
	persistedMsg := smbPersistedMessage(data.stage, data.msg)
	upd := repository.UpdateResult{
		HTTPStatus:   &httpStatus,
		KodeRespon:   helper.PtrString(data.statusRaw),
		Pesan:        helper.PtrString(persistedMsg),
		Harga:        helper.PtrI64(data.price),
		NoReferensi:  helper.PtrString(data.providerRef),
		ResponMentah: data.payload,
	}
	if data.hasLastBalance && data.lastBalance > 0 {
		upd.SaldoTerakhir = &data.lastBalance
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:            "smb",
			SaldoProvider:       data.lastBalance,
			RefID:               data.refid,
			TransaksiMemberID:   &row.TransaksiMemberID,
			TransaksiProviderID: &row.ID,
			Sumber:              "callback_smb",
			RawJSON:             data.payloadJSON,
		})
	}
	return s.repo.UpdateResult(ctx, row.ID, upd)
}
