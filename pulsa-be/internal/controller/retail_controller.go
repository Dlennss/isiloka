package controller

import (
	"encoding/json"
	"log"
	"net/http"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type RetailController struct {
	svc *service.RetailService
}

func NewRetailController(svc *service.RetailService) *RetailController {
	return &RetailController{svc: svc}
}

func (h *RetailController) Downlines(w http.ResponseWriter, r *http.Request) {
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := h.svc.ListDownlines(r.Context(), a.MemberID)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
	case http.MethodPost:
		var req struct {
			Email     string `json:"email"`
			Nama      string `json:"nama"`
			Phone     string `json:"phone"`
			StoreName string `json:"store_name"`
			Password  string `json:"password"`
			Role      string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		id, err := h.svc.RegisterDownline(r.Context(), a.MemberID, service.RetailRegisterDownlineInput{
			Email:     req.Email,
			Nama:      req.Nama,
			Phone:     req.Phone,
			StoreName: req.StoreName,
			Password:  req.Password,
			Role:      req.Role,
		})
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "member_id": id})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *RetailController) CommissionSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	item, err := h.svc.CommissionSummary(r.Context(), a.MemberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *RetailController) Commissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	rows, err := h.svc.ListCommissions(
		r.Context(),
		a.MemberID,
		helper.QueryInt(r, "limit", 20),
		helper.QueryInt(r, "offset", 0),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *RetailController) WithdrawRequests(w http.ResponseWriter, r *http.Request) {
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := h.svc.ListOwnWithdrawRequests(
			r.Context(),
			a.MemberID,
			helper.QueryInt(r, "limit", 20),
			helper.QueryInt(r, "offset", 0),
		)
		if err != nil {
			helper.SafeErrorResponse(w, http.StatusInternalServerError, "Gagal memuat riwayat penarikan.", err, "list own retail withdraw requests")
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
	case http.MethodPost:
		var req struct {
			Amount        int64  `json:"amount"`
			SourceType    string `json:"source_type"`
			BankName      string `json:"bank_name"`
			AccountName   string `json:"account_name"`
			AccountNumber string `json:"account_number"`
			Note          string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		item, err := h.svc.CreateWithdrawRequest(r.Context(), a.MemberID, service.RetailWithdrawCreateInput{
			Amount:        req.Amount,
			SourceType:    req.SourceType,
			BankName:      req.BankName,
			AccountName:   req.AccountName,
			AccountNumber: req.AccountNumber,
			Note:          req.Note,
		})
		if err != nil {
			log.Printf("[retail_withdraw_create] member_id=%d source=%s amount=%d err=%v", a.MemberID, req.SourceType, req.Amount, err)
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}
