package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type H2HController struct {
	svc *service.H2HService
}

func NewH2HController(svc *service.H2HService) *H2HController {
	return &H2HController{svc: svc}
}

func (h *H2HController) Downlines(w http.ResponseWriter, r *http.Request) {
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
			Email      string `json:"email"`
			Nama       string `json:"nama"`
			Password   string `json:"password"`
			PIN        string `json:"pin"`
			Role       string `json:"role"`
			FeeDana    int64  `json:"fee_dana"`
			FeeGopay   int64  `json:"fee_gopay"`
			FeeLinkAja int64  `json:"fee_linkaja"`
			FeeOvo     int64  `json:"fee_ovo"`
			FeeShopee  int64  `json:"fee_shopee"`
			FeeBank    int64  `json:"fee_bank"`
			FeeLainnya int64  `json:"fee_lainnya"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		if req.FeeDana < 0 || req.FeeGopay < 0 || req.FeeLinkAja < 0 || req.FeeOvo < 0 || req.FeeShopee < 0 || req.FeeBank < 0 || req.FeeLainnya < 0 {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("semua fee awal member harus >= 0"))
			return
		}
		id, apiKey, err := h.svc.RegisterDownline(r.Context(), a.MemberID, service.H2HRegisterDownlineInput{
			Email:    req.Email,
			Nama:     req.Nama,
			Password: req.Password,
			PIN:      req.PIN,
			Role:     req.Role,
			InitialFees: repository.H2HInitialFeeSetupInput{
				FeeDana:    req.FeeDana,
				FeeGopay:   req.FeeGopay,
				FeeLinkAja: req.FeeLinkAja,
				FeeOvo:     req.FeeOvo,
				FeeShopee:  req.FeeShopee,
				FeeBank:    req.FeeBank,
				FeeLainnya: req.FeeLainnya,
			},
		})
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "member_id": id, "api_key": apiKey})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *H2HController) CommissionSummary(w http.ResponseWriter, r *http.Request) {
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

func (h *H2HController) Commissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	rows, err := h.svc.ListCommissions(r.Context(), a.MemberID, helper.QueryInt(r, "limit", 20), helper.QueryInt(r, "offset", 0))
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *H2HController) WithdrawRequests(w http.ResponseWriter, r *http.Request) {
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := h.svc.ListOwnWithdrawRequests(r.Context(), a.MemberID, helper.QueryInt(r, "limit", 20), helper.QueryInt(r, "offset", 0))
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
	case http.MethodPost:
		var req struct {
			Amount        int64  `json:"amount"`
			BankName      string `json:"bank_name"`
			AccountName   string `json:"account_name"`
			AccountNumber string `json:"account_number"`
			Note          string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		item, err := h.svc.CreateWithdrawRequest(r.Context(), a.MemberID, service.H2HWithdrawCreateInput{
			Amount:        req.Amount,
			BankName:      req.BankName,
			AccountName:   req.AccountName,
			AccountNumber: req.AccountNumber,
			Note:          req.Note,
		})
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

type H2HAdminController struct {
	svc *service.H2HService
}

func NewH2HAdminController(svc *service.H2HService) *H2HAdminController {
	return &H2HAdminController{svc: svc}
}

func (h *H2HAdminController) ListWithdrawRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	rows, err := h.svc.AdminListWithdrawRequests(r.Context(), helper.QueryString(r, "status"), helper.QueryString(r, "q"), helper.QueryInt(r, "limit", 20), helper.QueryInt(r, "offset", 0))
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *H2HAdminController) ApproveWithdrawRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	reqID, _ := strconv.ParseInt(helper.QueryString(r, "id"), 10, 64)
	var req struct {
		BankID int64  `json:"bank_id"`
		Fee    int64  `json:"fee"`
		Note   string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.svc.AdminApproveWithdrawRequest(r.Context(), reqID, a.MemberID, a.Role, req.BankID, req.Fee, req.Note); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}

func (h *H2HAdminController) RejectWithdrawRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	reqID, _ := strconv.ParseInt(helper.QueryString(r, "id"), 10, 64)
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	if err := h.svc.AdminRejectWithdrawRequest(r.Context(), reqID, a.MemberID, req.Reason); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}
