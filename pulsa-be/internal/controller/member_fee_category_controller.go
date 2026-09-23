package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	memberfeekategoridto "pulsa2/internal/dto/member_fee_kategori"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type MemberFeeCategoryController struct {
	svc *service.MemberFeeCategoryService
}

func NewMemberFeeCategoryController(svc *service.MemberFeeCategoryService) *MemberFeeCategoryController {
	return &MemberFeeCategoryController{svc: svc}
}

func (h *MemberFeeCategoryController) Upsert(w http.ResponseWriter, r *http.Request) {
	var req memberfeekategoridto.UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("invalid json"))
		return
	}
	if req.MemberID <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("member_id invalid"))
		return
	}
	if strings.TrimSpace(req.FeeCode) == "" && req.KategoriID <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("fee_code invalid"))
		return
	}
	if req.FeeRp < 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("fee_rp must be >= 0"))
		return
	}

	aktif := true
	if req.Aktif != nil {
		aktif = *req.Aktif
	}

	if err := h.svc.Upsert(r.Context(), req.MemberID, req.FeeCode, req.KategoriID, req.FeeRp, aktif); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, memberfeekategoridto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, memberfeekategoridto.MapOK())
}

func (h *MemberFeeCategoryController) Delete(w http.ResponseWriter, r *http.Request) {
	var req memberfeekategoridto.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("invalid json"))
		return
	}
	if req.MemberID <= 0 || (strings.TrimSpace(req.FeeCode) == "" && req.KategoriID <= 0) {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("member_id dan fee_code wajib"))
		return
	}
	if err := h.svc.Delete(r.Context(), req.MemberID, req.FeeCode, req.KategoriID); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, memberfeekategoridto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, memberfeekategoridto.MapOK())
}

func (h *MemberFeeCategoryController) List(w http.ResponseWriter, r *http.Request) {
	memberID := helper.QueryInt64(r, "member_id", 0)
	if memberID <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeekategoridto.MapError("member_id invalid"))
		return
	}

	rows, err := h.svc.List(r.Context(), memberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, memberfeekategoridto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, memberfeekategoridto.MapList(rows))
}
