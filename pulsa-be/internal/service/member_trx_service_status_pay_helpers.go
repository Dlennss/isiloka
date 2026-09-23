package service

import (
	"context"
	"strings"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

func detectAdminManualFinalStatus(currentStatus string, ket string) (string, string, bool) {
	status := strings.ToLower(strings.TrimSpace(currentStatus))
	ket = strings.TrimSpace(ket)
	lowerKet := strings.ToLower(ket)

	if status == "failed" && strings.Contains(lowerKet, "dibatalkan admin") {
		if ket == "" {
			ket = "dibatalkan admin"
		}
		return "failed", ket, true
	}
	if status == "success" && strings.Contains(lowerKet, "diselesaikan admin") {
		if ket == "" {
			ket = "diselesaikan admin"
		}
		return "success", ket, true
	}
	return "", "", false
}

func (h *MemberTrxService) buildAlreadyFinalStatusPayResponse(ctx context.Context, trx *repository.TrxMemberFull, existP24, existJP, existYS, existTL, existMK, existSG, existMN, existTR, existAJ, existGM, existSM, existLB *model.JavapayTrxRow) *serviceResponse {
	finalStatus := strings.TrimSpace(strings.ToLower(trx.Status))
	ket := ""
	providerRef := ""
	sn := ""
	price := int64(0)

	switch {
	case existP24 != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("pulsa24jam", existP24, finalStatus)
	case existJP != nil:
		msg := ""
		if existJP.Pesan != nil {
			msg = strings.TrimSpace(*existJP.Pesan)
		}
		var info helper.ProviderInfo
		ket, info = helper.SafeMemberKeterangan(finalStatus, msg)
		providerRef = strings.TrimSpace(info.Reff)
		sn = strings.TrimSpace(info.SN)
		if existJP.NoReferensi != nil {
			noreff := strings.TrimSpace(*existJP.NoReferensi)
			if providerRef == "" {
				providerRef = noreff
			}
			if sn == "" {
				sn = noreff
			}
		}
		if existJP.Harga != nil {
			price = *existJP.Harga
		}
	case existYS != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("yuscom", existYS, finalStatus)
	case existTL != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("talentapay", existTL, finalStatus)
	case existMK != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("multikom", existMK, finalStatus)
	case existSG != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("sagaramobile", existSG, finalStatus)
	case existMN != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("minions", existMN, finalStatus)
	case existTR != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("trionik", existTR, finalStatus)
	case existAJ != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("ajs", existAJ, finalStatus)
	case existGM != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("gemilang", existGM, finalStatus)
	case existSM != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("smb", existSM, finalStatus)
	case existLB != nil:
		ket, providerRef, sn, price = providerRowWebhookInfo("loketbayar", existLB, finalStatus)
	default:
		ket, _ = helper.SafeMemberKeterangan(finalStatus, "")
	}

	st, _, whErr := h.sendFinalWebhook(ctx, trx, finalStatus, ket, providerRef, sn, price)
	return &serviceResponse{Body: trxmemberdto.MapAlreadyFinalResponse(trx.RefID, trx.Status, existP24 != nil || existJP != nil || existYS != nil || existTL != nil || existMK != nil || existSG != nil || existMN != nil || existTR != nil || existAJ != nil || existGM != nil || existSM != nil || existLB != nil, mapCallbackDelivery(st, whErr))}
}
