package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AgentCreditController struct {
	svc *service.AgentCreditService
}

func NewAgentCreditController(svc *service.AgentCreditService) *AgentCreditController {
	return &AgentCreditController{svc: svc}
}

func (h *AgentCreditController) MyApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		auth, ok := helper.GetAuth(r.Context())
		if !ok {
			helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		items, err := h.svc.ListMyApplications(r.Context(), auth)
		if err != nil {
			helper.WriteJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
		return
	}
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in service.AgentCreditSubmitInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	item, err := h.svc.SubmitApplication(r.Context(), auth, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *AgentCreditController) MasterApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		auth, ok := helper.GetAuth(r.Context())
		if !ok {
			helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		var in service.AgentCreditSubmitInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
			return
		}
		item, err := h.svc.SubmitApplication(r.Context(), auth, in)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	limit := helper.QueryInt(r, "limit", 50)
	items, err := h.svc.ListApplications(r.Context(), auth, limit)
	if err != nil {
		helper.WriteJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
}

func (h *AgentCreditController) ManualApplications(w http.ResponseWriter, r *http.Request) {
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	if r.Method == http.MethodGet {
		items, err := h.svc.ListManualEntryAgents(r.Context(), auth, strings.TrimSpace(r.URL.Query().Get("q")), helper.QueryInt(r, "limit", 100))
		if err != nil {
			helper.WriteJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
		return
	}
	if r.Method == http.MethodPut {
		var in service.AgentCreditManualInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
			return
		}
		if err := h.svc.UpdateManualApplication(r.Context(), auth, in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "Data migrasi berhasil diperbarui"})
		return
	}
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	var in service.AgentCreditManualInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	result, err := h.svc.CreateManualApplication(r.Context(), auth, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	status := http.StatusOK
	if result.ApprovalError != "" {
		status = http.StatusAccepted
	}
	response := map[string]any{
		"ok":        result.ApprovalError == "",
		"item":      result.Item,
		"approved":  result.Approved,
		"credit_id": fmt.Sprintf("KRD-%08d", result.Item.ID),
		"source":    "operator_manual",
		"status":    result.Item.Status,
		"saved":     true,
		"message":   "Data agent berhasil disimpan",
	}
	if result.ApprovalError != "" {
		response["error"] = "data tersimpan sebagai pending, tetapi persetujuan gagal: " + result.ApprovalError
	}
	if result.Approved {
		response["message"] = "Data agent berhasil disimpan dan kredit disetujui"
	}
	helper.WriteJSON(w, status, response)
}

func (h *AgentCreditController) MasterDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in service.AgentCreditDecisionInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	item, err := h.svc.DecideApplication(r.Context(), auth, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *AgentCreditController) DeleteRejectedApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in struct {
		ApplicationID int64 `json:"application_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	if err := h.svc.DeleteRejectedApplication(r.Context(), auth, in.ApplicationID); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *AgentCreditController) CreditRanks(w http.ResponseWriter, r *http.Request) {
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	if r.Method == http.MethodGet {
		items, err := h.svc.ListRanks(r.Context(), auth)
		if err != nil {
			helper.WriteJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
		return
	}
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	var in service.AgentCreditRankChangeInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	item, err := h.svc.ChangeMemberCreditRank(r.Context(), auth, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *AgentCreditController) AdminLoanStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in service.AgentCreditLoanStatusInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	item, err := h.svc.SetLoanOperationalStatus(r.Context(), auth, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *AgentCreditController) AdminTeamActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	items, err := h.svc.ListTeamActivity(r.Context(), auth, helper.QueryString(r, "role"), helper.QueryInt(r, "limit", 50))
	if err != nil {
		helper.WriteJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
}

func (h *AgentCreditController) AdminInactiveAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	items, err := h.svc.ListInactiveAgents(r.Context(), auth, helper.QueryInt(r, "days", 3), helper.QueryString(r, "q"), helper.QueryInt(r, "limit", 200))
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "only") {
			status = http.StatusForbidden
		}
		helper.WriteJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "days": helper.QueryInt(r, "days", 3)})
}

func (h *AgentCreditController) AdminAgentTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	items, err := h.svc.ListAgentTransactions(r.Context(), auth, helper.QueryString(r, "status"), helper.QueryString(r, "q"), helper.QueryInt(r, "limit", 100))
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "only") {
			status = http.StatusForbidden
		}
		helper.WriteJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": len(items)})
}

func (h *AgentCreditController) PayInstallment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in service.AgentCreditPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	if err := h.svc.PayInstallment(r.Context(), auth, in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *AgentCreditController) TransferToMainBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	auth, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in service.AgentCreditTransferInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	item, err := h.svc.TransferToMainBalance(r.Context(), auth, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}
