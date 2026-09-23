package service

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"pulsa2/smb"
)

type smbCallbackData struct {
	rawMsg         string
	msg            string
	statusRaw      string
	rawRefID       string
	stage          string
	refid          string
	routeCodeHint  string
	price          int64
	providerRef    string
	lastBalance    int64
	hasLastBalance bool
	payload        map[string]any
	payloadJSON    []byte
}

func extractSMBCallbackData(q url.Values) smbCallbackData {
	msg := smbCallbackValue(q, "message", "msg", "pesan")
	statusRaw := smbCallbackValue(q, "status", "statuscode", "rc")
	if statusRaw == "" {
		statusRaw = smb.ExtractStatusCode(msg)
	}

	rawRefID := smbCallbackValue(q, "refid", "idtrx", "clientid", "reffid")
	if rawRefID == "" {
		rawRefID = smb.ParseRefID(msg)
	}
	stage := smbCallbackStage(rawRefID)
	refid := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(rawRefID), "CEK"), "BYR")

	routeCodeHint := strings.ToUpper(strings.TrimSpace(smbCallbackValue(q, "kp", "kodeproduk", "produk")))
	if i := strings.Index(routeCodeHint, ":"); i > 0 {
		routeCodeHint = strings.TrimSpace(routeCodeHint[:i])
	}

	price := smb.ParsePrice(msg)
	if price == 0 {
		if n, _ := strconv.ParseInt(strings.TrimSpace(q.Get("harga")), 10, 64); n > 0 {
			price = n
		}
	}
	providerRef := strings.TrimSpace(q.Get("sn"))
	if providerRef == "" {
		providerRef = strings.TrimSpace(q.Get("noreff"))
	}
	if providerRef == "" {
		providerRef = smb.ParseProviderRef(msg)
	}
	lastBalance, hasLastBalance := smb.ParseLastBalance(msg)

	payload := map[string]any{}
	for k, vals := range q {
		if len(vals) == 1 {
			payload[k] = vals[0]
		} else {
			payload[k] = vals
		}
	}
	if msg != "" {
		payload["message"] = msg
	}
	if stage != "" {
		payload["stage"] = stage
	}
	if routeCodeHint != "" {
		payload["route_code_hint"] = routeCodeHint
	}
	payloadJSON, _ := json.Marshal(payload)

	return smbCallbackData{
		rawMsg:         msg,
		msg:            msg,
		statusRaw:      statusRaw,
		rawRefID:       rawRefID,
		stage:          stage,
		refid:          refid,
		routeCodeHint:  routeCodeHint,
		price:          price,
		providerRef:    providerRef,
		lastBalance:    lastBalance,
		hasLastBalance: hasLastBalance,
		payload:        payload,
		payloadJSON:    payloadJSON,
	}
}
