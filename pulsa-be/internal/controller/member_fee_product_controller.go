package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	memberfeeprodukdto "pulsa2/internal/dto/member_fee_produk"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type MemberFeeProductController struct {
	svc *service.MemberFeeProductService
}

func NewMemberFeeProductController(svc *service.MemberFeeProductService) *MemberFeeProductController {
	return &MemberFeeProductController{svc: svc}
}

func (h *MemberFeeProductController) Upsert(w http.ResponseWriter, r *http.Request) {
	var req memberfeeprodukdto.UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("invalid json"))
		return
	}
	req.KodeProduk = strings.ToUpper(strings.TrimSpace(req.KodeProduk))
	if req.MemberID <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("member_id invalid"))
		return
	}
	if req.ProdukID <= 0 && req.KodeProduk == "" {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("produk_id atau kode_produk wajib diisi"))
		return
	}

	hasPersen := req.FeePersen != nil
	hasRp := req.FeeRp != nil
	if hasPersen == hasRp {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("isi salah satu: fee_persen atau fee_rp"))
		return
	}
	if hasPersen && *req.FeePersen < 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("fee_persen must be >= 0"))
		return
	}
	if hasRp && *req.FeeRp < 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("fee_rp must be >= 0"))
		return
	}

	if err := h.svc.Upsert(r.Context(), req.MemberID, req.ProdukID, req.KodeProduk, req.FeePersen, req.FeeRp); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, memberfeeprodukdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, memberfeeprodukdto.MapOK())
}

func (h *MemberFeeProductController) Delete(w http.ResponseWriter, r *http.Request) {
	var req memberfeeprodukdto.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("invalid json"))
		return
	}
	req.KodeProduk = strings.ToUpper(strings.TrimSpace(req.KodeProduk))
	if req.MemberID <= 0 || (req.ProdukID <= 0 && req.KodeProduk == "") {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("member_id dan (produk_id/kode_produk) wajib"))
		return
	}
	if err := h.svc.Delete(r.Context(), req.MemberID, req.ProdukID, req.KodeProduk); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, memberfeeprodukdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, memberfeeprodukdto.MapOK())
}

func (h *MemberFeeProductController) List(w http.ResponseWriter, r *http.Request) {
	memberID := helper.QueryInt64(r, "member_id", 0)
	if memberID <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, memberfeeprodukdto.MapError("member_id invalid"))
		return
	}

	rows, err := h.svc.List(r.Context(), memberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, memberfeeprodukdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, memberfeeprodukdto.MapList(rows))
}
