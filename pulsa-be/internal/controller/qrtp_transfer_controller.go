package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type QRTPTransferController struct {
	svc *service.QRTPTransferService
}

func NewQRTPTransferController(svc *service.QRTPTransferService) *QRTPTransferController {
	return &QRTPTransferController{svc: svc}
}

func (h *QRTPTransferController) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	bank, err := h.svc.Summary(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"bank":       bank,
		"admin_fee":  2500,
		"max_amount": 50000000,
	})
}

func (h *QRTPTransferController) ProviderBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	balances, err := h.svc.ProviderBalances(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusBadGateway, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":                      true,
		"refreshed":               balances.Refreshed,
		"currency":                balances.Currency,
		"total_balance_available": balances.TotalBalanceAvailable,
	})
}

func (h *QRTPTransferController) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	provider := strings.TrimSpace(helper.QueryString(r, "provider"))
	status := strings.TrimSpace(helper.QueryString(r, "status"))
	direction := strings.TrimSpace(helper.QueryString(r, "direction"))
	search := strings.TrimSpace(helper.QueryString(r, "q"))
	limit := helper.QueryInt(r, "limit", 50)
	offset := helper.QueryInt(r, "offset", 0)
	items, total, err := h.svc.List(r.Context(), provider, status, direction, search, limit, offset)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *QRTPTransferController) Inquiry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	var in service.QRTPInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	row, err := h.svc.CreateInquiry(r.Context(), a.MemberID, in)
	if err != nil {
		payload := map[string]any{"ok": false, "error": err.Error()}
		if row != nil {
			payload["item"] = row
		}
		helper.WriteJSON(w, http.StatusBadRequest, payload)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
}

func (h *QRTPTransferController) Process(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	id, ok := qrtpPathID(w, r)
	if !ok {
		return
	}
	row, err := h.svc.Process(r.Context(), a.MemberID, id)
	if err != nil {
		payload := map[string]any{"ok": false, "error": err.Error()}
		if row != nil {
			payload["item"] = row
		}
		helper.WriteJSON(w, http.StatusBadRequest, payload)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
}

func (h *QRTPTransferController) TukangPayCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid body"))
		return
	}
	if !validTukangPaySignature(raw, r.Header.Get("X-TukangPay-Signature")) {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("invalid signature"))
		return
	}
	var payload struct {
		Event  string                     `json:"event"`
		Payout service.QRTPCallbackPayout `json:"payout"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	row, err := h.svc.HandleCallback(r.Context(), payload.Event, payload.Payout, raw)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
}

func qrtpPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/admin/qrtp/transfers/")
	path = strings.TrimSuffix(path, "/process")
	path = strings.Trim(path, "/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid id"))
		return 0, false
	}
	return id, true
}

func validTukangPaySignature(body []byte, signature string) bool {
	secret := strings.TrimSpace(os.Getenv("TUKANGPAY_CALLBACK_SECRET"))
	signature = strings.TrimSpace(signature)
	if secret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
