package service

import (
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
	"pulsa2/model"
)

func providerRowWebhookInfo(provider string, row *model.JavapayTrxRow, finalStatus string) (ket string, providerRef string, sn string, price int64) {
	if row == nil {
		ket, _ = helper.SafeMemberKeterangan(finalStatus, "")
		return ket, "", "", 0
	}

	msg := ""
	noreff := ""
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	if row.NoReferensi != nil {
		noreff = strings.TrimSpace(*row.NoReferensi)
	}
	if row.Harga != nil {
		price = *row.Harga
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	providerRef = strings.TrimSpace(info.Reff)
	sn = strings.TrimSpace(info.SN)

	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "talentapay":
		pr, parsedSN := providersn.ParseTalentaSNRefFromMsg(msg)
		if providerRef == "" || isWeakProviderRef(providerRef) {
			providerRef = strings.TrimSpace(pr)
		}
		if sn == "" || isWeakProviderRef(sn) {
			sn = strings.TrimSpace(parsedSN)
		}
	case "multikom":
		pr, parsedSN := providersn.ParseMultikomSNRefFromMsg(msg)
		if providerRef == "" || isWeakProviderRef(providerRef) {
			providerRef = strings.TrimSpace(pr)
		}
		if sn == "" || isWeakProviderRef(sn) {
			sn = strings.TrimSpace(parsedSN)
		}
	case "yuscom":
		pr, parsedSN := providersn.ParseYuscomSNRefFromMsg(msg)
		if providerRef == "" || isWeakProviderRef(providerRef) {
			providerRef = strings.TrimSpace(pr)
		}
		if sn == "" || isWeakProviderRef(sn) {
			sn = strings.TrimSpace(parsedSN)
		}
	case "sagaramobile":
		pr, parsedSN := providersn.ParseSagaraSNRefFromMsg(msg)
		if providerRef == "" || isWeakProviderRef(providerRef) {
			providerRef = strings.TrimSpace(pr)
		}
		if sn == "" || isWeakProviderRef(sn) {
			sn = strings.TrimSpace(parsedSN)
		}
	case "minions":
		pr, parsedSN := providersn.ParseMinionsSNRefFromMsg(msg)
		if providerRef == "" || isWeakProviderRef(providerRef) {
			providerRef = strings.TrimSpace(pr)
		}
		if sn == "" || isWeakProviderRef(sn) {
			sn = strings.TrimSpace(parsedSN)
		}
	case "trionik":
		pr, parsedSN := providersn.ParseYuscomSNRefFromMsg(msg)
		if providerRef == "" || isWeakProviderRef(providerRef) {
			providerRef = strings.TrimSpace(pr)
		}
		if sn == "" || isWeakProviderRef(sn) {
			sn = strings.TrimSpace(parsedSN)
		}
	}

	if (sn == "" || isWeakProviderRef(sn) || len(strings.TrimSpace(sn)) <= 3) && noreff != "" {
		sn = noreff
	}
	if (providerRef == "" || isWeakProviderRef(providerRef) || len(strings.TrimSpace(providerRef)) <= 3) && noreff != "" {
		providerRef = noreff
	}
	if strings.EqualFold(strings.TrimSpace(finalStatus), "success") {
		if providerRef != "" {
			ket = "Transaksi berhasil. Ref: " + providerRef
		} else {
			ket = "Transaksi berhasil"
		}
	}
	return ket, providerRef, sn, price
}

func isWeakProviderRef(v string) bool {
	v = strings.TrimSpace(strings.ToUpper(v))
	switch v {
	case "", "NO", "N/A", "-":
		return true
	default:
		return false
	}
}

func providerRowSuccessState(provider string, row *model.JavapayTrxRow) (bool, string, string, int64, string) {
	if row == nil {
		return false, "", "", 0, ""
	}
	rc := ""
	msg := ""
	noreff := ""
	if row.KodeRespon != nil {
		rc = strings.TrimSpace(*row.KodeRespon)
	}
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	if row.NoReferensi != nil {
		noreff = strings.TrimSpace(*row.NoReferensi)
	}
	price := int64(0)
	if row.Harga != nil {
		price = *row.Harga
	}
	if strings.EqualFold(row.Status, "success") {
		return true, rc, msg, price, noreff
	}
	if strings.TrimSpace(row.Status) == "" && helper.ProviderResponseStateOf(provider, rc, msg) == helper.ProviderResponseSuccess {
		return true, rc, msg, price, noreff
	}
	return false, rc, msg, price, noreff
}

func providerRowFailureState(provider string, row *model.JavapayTrxRow) (bool, bool, string, string, int64, string) {
	if row == nil {
		return false, false, "", "", 0, ""
	}

	rc := ""
	msg := ""
	noreff := ""
	if row.KodeRespon != nil {
		rc = strings.TrimSpace(*row.KodeRespon)
	}
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	if row.NoReferensi != nil {
		noreff = strings.TrimSpace(*row.NoReferensi)
	}

	price := int64(0)
	if row.Harga != nil {
		price = *row.Harga
	}

	classified := helper.ProviderResponseStateOf(provider, rc, msg)
	if classified == helper.ProviderResponsePending && (rc != "" || msg != "") {
		return false, false, rc, msg, price, noreff
	}

	if strings.EqualFold(row.Status, "failed") {
		return true, !helper.ShouldBlockProviderFallback(msg), rc, msg, price, noreff
	}
	if strings.TrimSpace(row.Status) == "" && classified == helper.ProviderResponseFailed {
		return true, !helper.ShouldBlockProviderFallback(msg), rc, msg, price, noreff
	}
	return false, false, rc, msg, price, noreff
}

func providerRowPendingState(provider string, row *model.JavapayTrxRow) (bool, string, string) {
	if row == nil {
		return false, "", ""
	}
	rc := ""
	msg := ""
	if row.KodeRespon != nil {
		rc = strings.TrimSpace(*row.KodeRespon)
	}
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	classified := helper.ProviderResponseStateOf(provider, rc, msg)

	if strings.EqualFold(row.Status, "pending") {
		return true, rc, msg
	}
	if classified == helper.ProviderResponsePending && (rc != "" || msg != "") {
		return true, rc, msg
	}
	return false, "", ""
}

func providerRowDefinitelyFailed(provider string, row *model.JavapayTrxRow) bool {
	if row == nil {
		return false
	}
	rc := ""
	msg := ""
	if row.KodeRespon != nil {
		rc = strings.TrimSpace(*row.KodeRespon)
	}
	if row.Pesan != nil {
		msg = strings.TrimSpace(*row.Pesan)
	}
	if helper.ProviderResponseStateOf(provider, rc, msg) == helper.ProviderResponsePending && (rc != "" || msg != "") {
		return false
	}
	return strings.EqualFold(row.Status, "failed")
}
