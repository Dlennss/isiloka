package trxmemberdto

import (
	"encoding/json"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/repository"
)

func MapErrorResponse(msg string) ErrorResponse {
	return commondto.MapError(msg)
}

func MapBusinessStatusResponse(refID string, status int, msg string) map[string]any {
	out := map[string]any{
		"ok":      true,
		"refid":   strings.TrimSpace(refID),
		"status":  status,
		"message": strings.TrimSpace(msg),
	}
	if out["message"] == "" {
		out["message"] = "Transaksi gagal"
	}
	return out
}

func MapExistingResponse(trx any) ExistingResponse {
	return ExistingResponse{
		Ok:              true,
		Existing:        true,
		TransaksiMember: trx,
	}
}

func MapStatusResponse(refID, status, rc, msg string, callback *CallbackDelivery) StatusResponse {
	return StatusResponse{
		Ok:           true,
		RefID:        refID,
		Status:       status,
		RC:           rc,
		Msg:          msg,
		ResponseKind: "status_pay_api_response",
		Callback:     callback,
	}
}

func MapAlreadyFinalResponse(refID, status string, providerJPOK bool, callback *CallbackDelivery) AlreadyFinalResponse {
	return AlreadyFinalResponse{
		Ok:           true,
		AlreadyFinal: true,
		RefID:        refID,
		Status:       status,
		WebhookRetry: callback != nil && callback.Attempted,
		ProviderJPOK: providerJPOK,
		ResponseKind: "status_pay_api_response",
		Callback:     callback,
	}
}

func MapRetryResponse(refID, status, retry, provider, reason string) RetryResponse {
	return RetryResponse{
		Ok:       true,
		RefID:    refID,
		Status:   status,
		Retry:    retry,
		Provider: provider,
		Reason:   reason,
	}
}

func MapTransaksiMemberResponse(id int64, refID, status string, qtyProvider, biayaPerkiraan, feeMemberRp int64, chargeReceiverApplied bool, keterangan, provider string) TransaksiMemberResponse {
	return TransaksiMemberResponse{
		Ok: true,
		TransaksiMember: TransaksiMemberItem{
			ID:                    id,
			RefID:                 refID,
			Status:                status,
			QtyProvider:           qtyProvider,
			ChargeReceiverApplied: chargeReceiverApplied,
			BiayaPerkiraan:        biayaPerkiraan,
			FeeMemberRp:           feeMemberRp,
			Keterangan:            keterangan,
			Provider:              provider,
		},
	}
}

func MapProdukListResponse(product string, items []repository.H2HProdukRow) ProdukListResponse {
	out := MapH2HProdukItems(items)
	return ProdukListResponse{
		Ok:       true,
		Commands: "PRODUK",
		Product:  product,
		Items:    out,
	}
}

func MapH2HProdukItems(items []repository.H2HProdukRow) []H2HProdukItemDTO {
	out := make([]H2HProdukItemDTO, 0, len(items))
	for _, item := range items {
		out = append(out, H2HProdukItemDTO{
			ID:              item.ID,
			SKU:             item.SKU,
			Nama:            item.Nama,
			GroupName:       item.GroupName,
			KategoriNama:    item.KategoriNama,
			BrandNama:       item.BrandNama,
			TipeHarga:       item.TipeHarga,
			Harga:           item.Harga,
			FeeTambahan:     item.FeeTambahan,
			MaksimalNominal: item.MaksimalNominal,
			DibuatPada:      item.DibuatPada,
			DiubahPada:      item.DiubahPada,
		})
	}
	return out
}

func NormalizeBusinessStatusPayload(body any) any {
	b, err := json.Marshal(body)
	if err != nil {
		return body
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return body
	}
	return normalizeBusinessStatusValue(out)
}

func normalizeBusinessStatusValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if strings.EqualFold(k, "status") {
				x[k] = NormalizeBusinessStatus(val)
				continue
			}
			x[k] = normalizeBusinessStatusValue(val)
		}
		return x
	case []any:
		for i, val := range x {
			x[i] = normalizeBusinessStatusValue(val)
		}
		return x
	default:
		return v
	}
}

func NormalizeBusinessStatus(v any) any {
	switch x := v.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(x)) {
		case "pending", "process", "proses", "1":
			return 1
		case "success", "sukses", "2":
			return 2
		case "failed", "gagal", "3":
			return 3
		default:
			return x
		}
	case float64:
		if x == 1 || x == 2 || x == 3 {
			return int(x)
		}
		return x
	default:
		return v
	}
}
