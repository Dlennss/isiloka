package service

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestParsePulsa24JamNestedCallback(t *testing.T) {
	for _, tc := range []struct{ raw, ref, status string }{
		{`{"ok":true,"transaksi_member":{"ref_id":"PKADANA","status":2,"keterangan":"NAMA:TEST,NOMINAL:20000,REFF:123"}}`, "PKADANA", "success"},
		{`{"ok":true,"transaksi_member":{"ref_id":"PKAGOPAY","status":1}}`, "PKAGOPAY", "pending"},
		{`{"ok":true,"transaksi_member":{"ref_id":"PKAFAIL","status":3}}`, "PKAFAIL", "failed"},
		{`{"ok":true,"refid":"PKAFINAL","already_final":true,"status":2}`, "PKAFINAL", "success"},
		{`{"ok":true,"refid":"PKAUNKNOWN"}`, "PKAUNKNOWN", "pending"},
	} {
		var payload map[string]any
		if err := json.Unmarshal([]byte(tc.raw), &payload); err != nil {
			t.Fatal(err)
		}
		got := parsePulsa24JamCallback(tc.raw, url.Values{}, payload)
		if got.refid != tc.ref || pulsa24JamFinalStatus(got) != tc.status {
			t.Fatalf("unexpected callback: ref=%q status=%q", got.refid, pulsa24JamFinalStatus(got))
		}
		if got.price != 0 {
			t.Fatal("must not infer actual cost from receipt nominal")
		}
	}
}
