package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type RetailAdminController struct {
	svc *service.RetailService
}

func NewRetailAdminController(svc *service.RetailService) *RetailAdminController {
	return &RetailAdminController{svc: svc}
}

func (h *RetailAdminController) ListWithdrawRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	rows, err := h.svc.AdminListWithdrawRequests(
		r.Context(),
		helper.QueryString(r, "status"),
		helper.QueryString(r, "q"),
		helper.QueryInt(r, "limit", 20),
		helper.QueryInt(r, "offset", 0),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *RetailAdminController) ListWithdrawSourceBanks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	rows, err := h.svc.ListWithdrawSourceBanks(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *RetailAdminController) ApproveWithdrawRequest(w http.ResponseWriter, r *http.Request) {
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

func (h *RetailAdminController) RejectWithdrawRequest(w http.ResponseWriter, r *http.Request) {
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
