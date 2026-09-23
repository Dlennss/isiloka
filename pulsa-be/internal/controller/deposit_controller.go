package controller

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type DepositController struct {
	svc *service.DepositService
}

func NewDepositController(svc *service.DepositService) *DepositController {
	return &DepositController{svc: svc}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if err == sql.ErrNoRows {
		return true
	}

	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "no rows")
}

func (h *DepositController) MemberCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	var in struct {
		Amount   int64  `json:"amount"`
		BankID   int64  `json:"bank_id"`
		Metode   string `json:"metode"`
		BuktiURL string `json:"bukti_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	item, err := h.svc.CreateRequest(r.Context(), a.MemberID, a.Role, in.BankID, in.Amount, in.Metode, in.BuktiURL)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *DepositController) MemberConfirmTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	var in struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	item, err := h.svc.ConfirmTicketTransfer(r.Context(), a.MemberID, in.ID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *DepositController) MemberCancelTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	var in struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	item, err := h.svc.CancelTicket(r.Context(), a.MemberID, in.ID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *DepositController) MemberCreateQris(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	var in struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	out, err := h.svc.CreateQrisRequest(r.Context(), a.MemberID, a.Role, in.Amount)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
}

func (h *DepositController) MemberCreateVA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	var in struct {
		Amount int64  `json:"amount"`
		Bank   string `json:"bank"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	item, err := h.svc.CreateVARequest(r.Context(), a.MemberID, a.Role, in.Amount, in.Bank)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *DepositController) MemberQrisStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	refID := strings.TrimSpace(helper.QueryString(r, "ref_id"))
	refresh := strings.EqualFold(strings.TrimSpace(helper.QueryString(r, "refresh")), "1") ||
		strings.EqualFold(strings.TrimSpace(helper.QueryString(r, "refresh")), "true")
	item, row, err := h.svc.GetQrisStatusByRefID(r.Context(), a.MemberID, a.Role, refID, refresh)
	if err != nil {
		status := http.StatusBadRequest
		if isNotFoundError(err) {
			status = http.StatusNotFound
		}
		helper.WriteJSON(w, status, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"item": item,
		"row":  row,
	})
}

func (h *DepositController) MemberBanks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	rows, err := h.svc.ActiveBanks(r.Context(), a.MemberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *DepositController) MemberList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	limit := helper.QueryInt(r, "limit", 50)
	rows, err := h.svc.ListMemberRequests(r.Context(), a.MemberID, limit)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "rows": rows})
}

func (h *DepositController) AdminList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	rows, err := h.svc.AdminList(
		r.Context(),
		helper.QueryString(r, "status"),
		helper.QueryInt64(r, "member_id", 0),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		helper.QueryInt(r, "limit", 50),
		helper.QueryInt(r, "offset", 0),
		helper.QueryString(r, "order"),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *DepositController) AdminListVA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	rows, err := h.svc.AdminListVA(
		r.Context(),
		helper.QueryString(r, "status"),
		helper.QueryInt64(r, "member_id", 0),
		helper.QueryString(r, "ref_id"),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		helper.QueryInt(r, "limit", 50),
		helper.QueryInt(r, "offset", 0),
		helper.QueryString(r, "order"),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}

func (h *DepositController) AdminApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin or operator_wallet only"))
		return
	}
	reqID, _ := strconv.ParseInt(helper.QueryString(r, "id"), 10, 64)
	var in struct {
		Note           string   `json:"note"`
		ApprovedAmount int64    `json:"approved_amount"`
		CreditAmount   int64    `json:"credit_amount"`
		BankRefID      string   `json:"bank_ref_id"`
		BankRefIDs     []string `json:"bank_ref_ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	approvedAmount := in.ApprovedAmount
	if approvedAmount <= 0 {
		approvedAmount = in.CreditAmount
	}
	bankRefIDs := append([]string{}, in.BankRefIDs...)
	if strings.TrimSpace(in.BankRefID) != "" {
		bankRefIDs = append(bankRefIDs, in.BankRefID)
	}
	refID, creditedAmount, err := h.svc.AdminApprove(r.Context(), reqID, a.MemberID, approvedAmount, in.Note, bankRefIDs)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "ref_id": refID, "approved_amount": creditedAmount})
}

func (h *DepositController) AdminApproveVA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin or operator_wallet only"))
		return
	}
	reqID, _ := strconv.ParseInt(helper.QueryString(r, "id"), 10, 64)
	var in struct {
		Note           string `json:"note"`
		ApprovedAmount int64  `json:"approved_amount"`
		TransferAmount int64  `json:"transfer_amount"`
		CreditAmount   int64  `json:"credit_amount"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	approvedAmount := in.ApprovedAmount
	if approvedAmount <= 0 {
		approvedAmount = in.TransferAmount
	}
	if approvedAmount <= 0 {
		approvedAmount = in.CreditAmount
	}
	item, err := h.svc.AdminApproveVA(r.Context(), reqID, a.MemberID, approvedAmount, in.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item, "approved_amount": item.Approved, "ref_id": item.RefID})
}

func (h *DepositController) AdminReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin or operator_wallet only"))
		return
	}
	reqID, _ := strconv.ParseInt(helper.QueryString(r, "id"), 10, 64)
	var in struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.svc.AdminReject(r.Context(), reqID, a.MemberID, in.Note); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}

func (h *DepositController) AdminRejectVA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin or operator_wallet only"))
		return
	}
	reqID, _ := strconv.ParseInt(helper.QueryString(r, "id"), 10, 64)
	var in struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	item, err := h.svc.AdminRejectVA(r.Context(), reqID, a.MemberID, in.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *DepositController) AdminCreditInternal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	adminToken := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
	if adminToken == "" {
		adminToken = strings.TrimSpace(r.Header.Get("Authorization"))
	}
	if strings.HasPrefix(adminToken, "Bearer ") {
		adminToken = strings.TrimSpace(strings.TrimPrefix(adminToken, "Bearer "))
	}
	if adminToken == "" || subtle.ConstantTimeCompare([]byte(adminToken), []byte(os.Getenv("ADMIN_TOKEN"))) != 1 {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	var req struct {
		MemberID int64  `json:"member_id"`
		Amount   int64  `json:"amount"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("bad json"))
		return
	}
	refID, err := h.svc.AdminCreditInternal(r.Context(), req.MemberID, req.Amount, req.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "ref_id": refID})
}
