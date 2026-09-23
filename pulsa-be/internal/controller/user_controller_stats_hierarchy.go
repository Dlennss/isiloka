package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

func (h *UserController) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	userID := helper.QueryInt64(r, "user_id", 0)
	if userID <= 0 {
		userID, _ = strconv.ParseInt(helper.QueryString(r, "member_id"), 10, 64)
	}
	items, err := h.svc.StatsLast3Months(r.Context(), userID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
}

func (h *UserController) PreviewHierarchy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		Scope      string `json:"scope"`
		MemberID   int64  `json:"member_id"`
		AgentID    *int64 `json:"agent_id"`
		MasterID   *int64 `json:"master_id"`
		TargetRole string `json:"target_role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	item, err := h.svc.PreviewHierarchyAssignment(r.Context(), req.Scope, req.MemberID, req.AgentID, req.MasterID, req.TargetRole)
	if err != nil {
		code := http.StatusBadRequest
		if service.IsNotFound(err) {
			code = http.StatusNotFound
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *UserController) ApplyHierarchy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		Scope           string `json:"scope"`
		MemberID        int64  `json:"member_id"`
		AgentID         *int64 `json:"agent_id"`
		MasterID        *int64 `json:"master_id"`
		TargetRole      string `json:"target_role"`
		ApplyHistorical bool   `json:"apply_historical"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	item, err := h.svc.ApplyHierarchyAssignment(r.Context(), req.Scope, req.MemberID, req.AgentID, req.MasterID, req.TargetRole, req.ApplyHistorical)
	if err != nil {
		code := http.StatusBadRequest
		if service.IsNotFound(err) {
			code = http.StatusNotFound
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}
