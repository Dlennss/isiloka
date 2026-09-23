package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppOrderRefundAdminController struct {
	svc *service.AppOrderRefundAdminService
}

func NewAppOrderRefundAdminController(svc *service.AppOrderRefundAdminService) *AppOrderRefundAdminController {
	return &AppOrderRefundAdminController{svc: svc}
}

func (h *AppOrderRefundAdminController) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	limit := helper.QueryInt(r, "limit", 10)
	offset := helper.QueryInt(r, "offset", 0)
	items, total, err := h.svc.ListPending(r.Context(), limit, offset, helper.QueryString(r, "invoice_id"))
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	if limit <= 0 {
		limit = 10
	}
	page := (offset / limit) + 1
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"items":       items,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
		"page":        page,
		"total_pages": totalPages,
		"filters": map[string]any{
			"invoice_id": strings.TrimSpace(helper.QueryString(r, "invoice_id")),
		},
	})
}

func (h *AppOrderRefundAdminController) HandleClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	var req struct {
		InvoiceID      string `json:"invoice_id"`
		TargetMemberID int64  `json:"target_member_id"`
		TargetEmail    string `json:"target_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	refundRow, target, err := h.svc.ClaimToUser(r.Context(), req.InvoiceID, req.TargetMemberID, req.TargetEmail)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"item":   refundRow,
		"target": target,
	})
}
