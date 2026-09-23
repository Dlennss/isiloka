package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	apporderpaymentdto "pulsa2/internal/dto/app_order_payment"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppOrderPaymentController struct {
	svc        *service.AppOrderPaymentService
	depositSvc *service.DepositService
}

func NewAppOrderPaymentController(svc *service.AppOrderPaymentService, depositSvc *service.DepositService) *AppOrderPaymentController {
	return &AppOrderPaymentController{svc: svc, depositSvc: depositSvc}
}

func (h *AppOrderPaymentController) MidtransWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const maxBody = int64(2 << 20)
	raw, _ := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	_ = r.Body.Close()
	if int64(len(raw)) > maxBody {
		raw = raw[:maxBody]
	}
	helper.AppendProviderCallbackLog("midtrans", r, raw)

	r.Body = io.NopCloser(bytes.NewReader(raw))
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()

	var payload map[string]any
	if err := dec.Decode(&payload); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderpaymentdto.MapError("invalid json"))
		return
	}

	orderID := helper.ToString(payload["order_id"])
	status := http.StatusInternalServerError
	out := map[string]any{"ok": false, "error": "handler not configured"}
	if service.IsDepositQrisOrderID(orderID) {
		if h.depositSvc == nil {
			helper.WriteJSON(w, http.StatusInternalServerError, apporderpaymentdto.MapError("deposit qris handler belum diset"))
			return
		}
		status, out = h.depositSvc.HandleMidtransWebhook(r.Context(), raw, payload)
	} else {
		status, out = h.svc.HandleMidtransWebhook(r.Context(), raw, payload)
	}
	helper.WriteJSON(w, status, out)
}
