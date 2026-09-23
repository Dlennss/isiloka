package controller

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func (h *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError("ADMIN_TOKEN not set"))
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Admin-Token")), []byte(adminToken)) != 1 {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	var req struct {
		Email                    string `json:"email"`
		Nama                     string `json:"nama"`
		Phone                    string `json:"phone"`
		Password                 string `json:"password"`
		PIN                      string `json:"pin"`
		Role                     string `json:"role"`
		RetailAgentCommissionRp  *int64 `json:"retail_agent_commission_rp"`
		RetailMasterCommissionRp *int64 `json:"retail_master_commission_rp"`
		H2HAgentCommissionRp     *int64 `json:"h2h_agent_commission_rp"`
		H2HMasterCommissionRp    *int64 `json:"h2h_master_commission_rp"`
		FeeDana                  *int64 `json:"fee_dana"`
		FeeGopay                 *int64 `json:"fee_gopay"`
		FeeLinkAja               *int64 `json:"fee_linkaja"`
		FeeOvo                   *int64 `json:"fee_ovo"`
		FeeShopee                *int64 `json:"fee_shopee"`
		FeeBank                  *int64 `json:"fee_bank"`
		FeeLainnya               *int64 `json:"fee_lainnya"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("password min 8 chars"))
		return
	}
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if role == "" {
		role = "member"
	}
	pin := strings.TrimSpace(req.PIN)
	if helper.IsH2HRole(role) && (len(pin) < 4 || len(pin) > 12) {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("pin length 4-12"))
		return
	}
	commissions := []*int64{
		req.RetailAgentCommissionRp,
		req.RetailMasterCommissionRp,
		req.H2HAgentCommissionRp,
		req.H2HMasterCommissionRp,
	}
	for _, commission := range commissions {
		if commission != nil && *commission < 0 {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("komisi tidak boleh negatif"))
			return
		}
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError("hash password failed"))
		return
	}
	pinHash := ""
	if helper.IsH2HRole(role) {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
		if hashErr != nil {
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError("hash pin failed"))
			return
		}
		pinHash = string(hash)
	}

	var initialFees *repository.H2HInitialFeeSetupInput
	if helper.IsH2HRole(role) {
		if req.FeeDana == nil || req.FeeGopay == nil || req.FeeLinkAja == nil || req.FeeOvo == nil || req.FeeShopee == nil || req.FeeBank == nil || req.FeeLainnya == nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("fee dana, gopay, linkaja, ovo, shopee, bank, dan lainnya wajib diisi"))
			return
		}
		fields := []*int64{req.FeeDana, req.FeeGopay, req.FeeLinkAja, req.FeeOvo, req.FeeShopee, req.FeeBank, req.FeeLainnya}
		for _, fee := range fields {
			if fee == nil || *fee < 0 {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("semua fee awal member harus >= 0"))
				return
			}
		}
		initialFees = &repository.H2HInitialFeeSetupInput{
			FeeDana:    *req.FeeDana,
			FeeGopay:   *req.FeeGopay,
			FeeLinkAja: *req.FeeLinkAja,
			FeeOvo:     *req.FeeOvo,
			FeeShopee:  *req.FeeShopee,
			FeeBank:    *req.FeeBank,
			FeeLainnya: *req.FeeLainnya,
		}
	}

	id, apiKey, err := h.svc.Register(r.Context(), repository.UserCreateInput{
		Email:                    req.Email,
		Nama:                     req.Nama,
		Phone:                    req.Phone,
		PasswordHash:             string(passHash),
		PinHash:                  pinHash,
		Role:                     role,
		APIKey:                   helper.RandHex(32),
		Label:                    "default",
		Aktif:                    true,
		RetailAgentCommissionRp:  derefInt64(req.RetailAgentCommissionRp),
		RetailMasterCommissionRp: derefInt64(req.RetailMasterCommissionRp),
		H2HAgentCommissionRp:     derefInt64(req.H2HAgentCommissionRp),
		H2HMasterCommissionRp:    derefInt64(req.H2HMasterCommissionRp),
	}, initialFees)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "member_id": id, "api_key": apiKey})
}

func (h *AuthController) RegisterPublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	var req struct {
		Email           string `json:"email"`
		Nama            string `json:"nama"`
		Phone           string `json:"phone"`
		Password        string `json:"password"`
		RefundInvoiceID string `json:"refund_invoice_id"`
		GuestEmail      string `json:"guest_email"`
		GuestPhone      string `json:"guest_phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("password min 8 chars"))
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError("hash password failed"))
		return
	}

	id, refundRow, err := h.svc.RegisterPublic(r.Context(), req.Email, req.Nama, req.Phone, string(passHash), &service.RegisterPublicGuestRefundInput{
		InvoiceID:  req.RefundInvoiceID,
		GuestEmail: req.GuestEmail,
		GuestPhone: req.GuestPhone,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	resp := map[string]any{"ok": true, "member_id": id, "role": "user"}
	if refundRow != nil {
		resp["auto_refund_claimed"] = true
		resp["auto_refund_amount"] = refundRow.AmountRefund
		resp["refund_invoice_id"] = refundRow.InvoiceID
	}
	helper.WriteJSON(w, http.StatusOK, resp)
}
