package controller

import (
	"encoding/json"
	"net/http"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

func (h *UserController) SetFee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		UserID   int64 `json:"user_id"`
		MemberID int64 `json:"member_id"`
		FeeRp    int64 `json:"fee_member_rp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	userID := req.UserID
	if userID <= 0 {
		userID = req.MemberID
	}
	if err := h.svc.SetFee(r.Context(), userID, req.FeeRp); err != nil {
		code := http.StatusBadRequest
		if service.IsNotFound(err) {
			code = http.StatusNotFound
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}

func (h *UserController) SetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		UserID      int64  `json:"user_id"`
		MemberID    int64  `json:"member_id"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	userID := req.UserID
	if userID <= 0 {
		userID = req.MemberID
	}
	if err := h.svc.SetPassword(r.Context(), userID, req.NewPassword); err != nil {
		code := http.StatusBadRequest
		if service.IsNotFound(err) {
			code = http.StatusNotFound
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}

func (h *UserController) SetPIN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		UserID   int64  `json:"user_id"`
		MemberID int64  `json:"member_id"`
		NewPIN   string `json:"new_pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	userID := req.UserID
	if userID <= 0 {
		userID = req.MemberID
	}
	if err := h.svc.SetPIN(r.Context(), userID, req.NewPIN); err != nil {
		code := http.StatusBadRequest
		if service.IsNotFound(err) {
			code = http.StatusNotFound
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}
