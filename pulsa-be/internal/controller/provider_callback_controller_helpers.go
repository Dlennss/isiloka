package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"pulsa2/internal/helper"
)

func readProviderCallbackBody(r *http.Request) []byte {
	const maxBody = int64(2 << 20)
	raw, _ := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	_ = r.Body.Close()
	if int64(len(raw)) > maxBody {
		raw = raw[:maxBody]
	}
	return raw
}

func decodeJSONMap(raw []byte) (map[string]any, error) {
	payload := map[string]any{}
	if len(bytes.TrimSpace(raw)) == 0 {
		return payload, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func mergeCallbackQuery(raw []byte, base url.Values) url.Values {
	q, _ := url.ParseQuery(string(raw))
	for k, vals := range base {
		if len(vals) == 0 {
			continue
		}
		if _, ok := q[k]; !ok {
			q[k] = vals
		}
	}
	return q
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func appendProviderLog(provider string, r *http.Request, raw []byte) {
	helper.AppendProviderCallbackLog(provider, r, raw)
}
