package service

import (
	"net/url"
	"strconv"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
)

type talentaCallbackData struct {
	msg         string
	refid       string
	rcNum       int
	rcStr       string
	price       int64
	qty         int64
	dest        string
	noreff      string
	lastBalance int64
	rawVals     url.Values
}

func extractTalentaCallbackData(payload map[string]any) talentaCallbackData {
	data := helper.AsMap(payload["data"])
	trxObj := helper.AsMap(payload["trx"])
	rawObj := helper.AsMap(payload["raw"])
	rawStr := strings.TrimSpace(helper.ToString(payload["raw"]))

	var rawVals url.Values
	if rawStr != "" {
		if parsed, err := url.ParseQuery(rawStr); err == nil {
			rawVals = parsed
		}
	}

	msg := helper.PickStr(payload, data, "message", "pesan", "status_desc")
	if msg == "" && trxObj != nil {
		msg = helper.PickStr(payload, trxObj, "message", "pesan", "status_desc")
	}
	if msg == "" && rawObj != nil {
		msg = helper.PickStr(payload, rawObj, "message", "pesan", "status_desc")
	}
	if msg == "" && rawVals != nil {
		msg = strings.TrimSpace(rawVals.Get("message"))
	}

	refid := helper.PickStr(payload, data, "refid", "ref_id", "refID", "reffid")
	if refid == "" && trxObj != nil {
		refid = helper.PickStr(payload, trxObj, "refid", "ref_id", "refID", "reffid")
	}
	if refid == "" && rawObj != nil {
		refid = helper.PickStr(payload, rawObj, "refid", "ref_id", "refID", "reffid")
	}
	if refid == "" && rawVals != nil {
		refid = strings.TrimSpace(rawVals.Get("refid"))
	}
	if refid == "" {
		refid = helper.PickStr(payload, nil, "refid", "ref_id", "refID", "reffid")
	}
	if refid == "" {
		refid = helper.ExtractRefIDFromPesan(msg)
	}
	refid = strings.TrimSpace(refid)

	rcNum := int(helper.PickI64(payload, data, "rc", "status", "code"))
	if rcNum == 0 && trxObj != nil {
		rcNum = int(helper.PickI64(payload, trxObj, "rc", "status", "code"))
	}
	if rcNum == 0 && rawObj != nil {
		rcNum = int(helper.PickI64(payload, rawObj, "rc", "status", "code"))
	}
	rcStr := strings.TrimSpace(helper.PickStr(payload, data, "rc", "status", "code"))
	if rcStr == "" && trxObj != nil {
		rcStr = strings.TrimSpace(helper.PickStr(payload, trxObj, "rc", "status", "code"))
	}
	if rcStr == "" && rawObj != nil {
		rcStr = strings.TrimSpace(helper.PickStr(payload, rawObj, "rc", "status", "code"))
	}
	if rcStr == "" && rawVals != nil {
		rcStr = strings.TrimSpace(rawVals.Get("status"))
	}
	if rcStr == "" {
		rcStr = helper.ExtractProviderStatusCode(msg)
	}
	if rcStr == "" && rcNum != 0 {
		rcStr = strconv.Itoa(rcNum)
	}

	price := helper.PickI64(payload, data, "price", "harga", "amount")
	if price <= 0 && trxObj != nil {
		price = helper.PickI64(payload, trxObj, "price", "harga", "amount")
	}
	if price <= 0 && rawObj != nil {
		price = helper.PickI64(payload, rawObj, "price", "harga", "amount")
	}
	if price <= 0 && rawVals != nil {
		rawPrice := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(rawVals.Get("price")), ".", ""), ",", "")
		price, _ = strconv.ParseInt(rawPrice, 10, 64)
	}

	qty := helper.PickI64(payload, data, "qty", "jumlah", "amount")
	if qty <= 0 && trxObj != nil {
		qty = helper.PickI64(payload, trxObj, "qty", "jumlah", "amount")
	}
	if qty <= 0 && rawObj != nil {
		qty = helper.PickI64(payload, rawObj, "qty", "jumlah", "amount")
	}
	if qty <= 0 && rawVals != nil {
		qty = helper.ToI64(strings.TrimSpace(rawVals.Get("qty")))
	}

	dest := helper.PickStr(payload, data, "dest", "tujuan")
	if dest == "" && trxObj != nil {
		dest = helper.PickStr(payload, trxObj, "dest", "tujuan")
	}
	if dest == "" && rawObj != nil {
		dest = helper.PickStr(payload, rawObj, "dest", "tujuan")
	}
	if dest == "" && rawVals != nil {
		dest = strings.TrimSpace(rawVals.Get("dest"))
	}

	noreff := helper.PickStr(payload, data, "noreff", "sn", "provider_ref", "ticket")
	if noreff == "" && trxObj != nil {
		noreff = helper.PickStr(payload, trxObj, "noreff", "sn", "provider_ref", "ticket")
	}
	if noreff == "" && rawObj != nil {
		noreff = helper.PickStr(payload, rawObj, "noreff", "sn", "provider_ref", "ticket")
	}
	if noreff == "" && rawVals != nil {
		noreff = strings.TrimSpace(rawVals.Get("noreff"))
	}
	if strings.TrimSpace(noreff) == "" {
		pr, sn := providersn.ParseTalentaSNRefFromMsg(msg)
		if strings.TrimSpace(sn) != "" {
			noreff = sn
		} else if strings.TrimSpace(pr) != "" {
			noreff = pr
		}
	}

	lastBalance := helper.PickI64(payload, data, "last_balance", "saldo", "balance", "saldo_akhir")
	if lastBalance <= 0 && trxObj != nil {
		lastBalance = helper.PickI64(payload, trxObj, "last_balance", "saldo", "balance", "saldo_akhir")
	}
	if lastBalance <= 0 && rawObj != nil {
		lastBalance = helper.PickI64(payload, rawObj, "last_balance", "saldo", "balance", "saldo_akhir")
	}
	if lastBalance <= 0 {
		if bal, ok := helper.ExtractSaldoTerakhirFromMsg(msg); ok {
			lastBalance = bal
		}
	}

	return talentaCallbackData{
		msg:         msg,
		refid:       refid,
		rcNum:       rcNum,
		rcStr:       rcStr,
		price:       price,
		qty:         qty,
		dest:        dest,
		noreff:      noreff,
		lastBalance: lastBalance,
		rawVals:     rawVals,
	}
}
