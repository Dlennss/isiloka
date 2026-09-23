package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	produkdto "pulsa2/internal/dto/produk"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type ProdukController struct {
	svc  *service.ProdukService
	base string
}

func NewProdukController(svc *service.ProdukService, base string) *ProdukController {
	return &ProdukController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *ProdukController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/produk"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			if helper.QueryInt(r, "groups", 0) == 1 {
				groups, err := h.svc.ListGroupNames(r.Context())
				if err != nil {
					helper.WriteJSON(w, http.StatusInternalServerError, produkdto.MapError(err.Error()))
					return
				}
				helper.WriteJSON(w, http.StatusOK, map[string]any{
					"ok":    true,
					"items": groups,
				})
				return
			}

			q := helper.QueryString(r, "q")
			sku := helper.QueryString(r, "sku")
			groupName := helper.QueryString(r, "group_name")
			kategoriID := int64(helper.QueryInt(r, "kategori_id", 0))
			brandID := int64(helper.QueryInt(r, "brand_id", 0))
			limit := helper.QueryInt(r, "limit", 10)
			page := helper.QueryInt(r, "page", 1)
			if page <= 0 {
				page = 1
			}
			if limit <= 0 {
				limit = 10
			}
			offset := (page - 1) * limit

			rows, total, err := h.svc.List(r.Context(), q, sku, groupName, kategoriID, brandID, limit, offset)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, produkdto.MapError(err.Error()))
				return
			}
			totalPages := int((total + int64(limit) - 1) / int64(limit))
			helper.WriteJSON(w, http.StatusOK, map[string]any{
				"ok":          true,
				"items":       rows,
				"total":       total,
				"limit":       limit,
				"offset":      offset,
				"page":        page,
				"total_pages": totalPages,
			})
		case http.MethodPost:
			var req produkdto.Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("invalid json"))
				return
			}
			if strings.TrimSpace(req.SKU) == "" || strings.TrimSpace(req.Nama) == "" {
				helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("sku/nama required"))
				return
			}
			if req.KategoriID <= 0 || req.BrandID <= 0 {
				helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("kategori_id/brand_id invalid"))
				return
			}
			if strings.TrimSpace(req.TipeHarga) == "" {
				helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("tipe_harga required"))
				return
			}
			req.TipeHarga = strings.ToUpper(strings.TrimSpace(req.TipeHarga))
			if req.TipeHarga != "FIXED" && req.TipeHarga != "OPEN_AMOUNT" {
				helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("tipe_harga harus FIXED atau OPEN_AMOUNT"))
				return
			}
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			id, err := h.svc.Create(r.Context(), repository.ProdukUpsertInput{
				SKU:             req.SKU,
				Nama:            req.Nama,
				GroupName:       req.GroupName,
				KategoriID:      req.KategoriID,
				BrandID:         req.BrandID,
				TipeHarga:       req.TipeHarga,
				Nominal:         req.Nominal,
				MaksimalNominal: req.MaksimalNominal,
				Aktif:           aktif,
			})
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, produkdto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, produkdto.MapID(id))
		default:
			helper.WriteJSON(w, http.StatusMethodNotAllowed, produkdto.MapError("method not allowed"))
		}
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, produkdto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, produkdto.MapError("produk not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, produkdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, produkdto.MapItem(row))
	case http.MethodPut:
		var req produkdto.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("invalid json"))
			return
		}
		if strings.TrimSpace(req.SKU) == "" || strings.TrimSpace(req.Nama) == "" {
			helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("sku/nama required"))
			return
		}
		if req.KategoriID <= 0 || req.BrandID <= 0 {
			helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("kategori_id/brand_id invalid"))
			return
		}
		if strings.TrimSpace(req.TipeHarga) == "" {
			helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("tipe_harga required"))
			return
		}
		req.TipeHarga = strings.ToUpper(strings.TrimSpace(req.TipeHarga))
		if req.TipeHarga != "FIXED" && req.TipeHarga != "OPEN_AMOUNT" {
			helper.WriteJSON(w, http.StatusBadRequest, produkdto.MapError("tipe_harga harus FIXED atau OPEN_AMOUNT"))
			return
		}
		aktif := true
		if req.Aktif != nil {
			aktif = *req.Aktif
		}
		if err := h.svc.Update(r.Context(), repository.ProdukUpsertInput{
			ID:              id,
			SKU:             req.SKU,
			Nama:            req.Nama,
			GroupName:       req.GroupName,
			KategoriID:      req.KategoriID,
			BrandID:         req.BrandID,
			TipeHarga:       req.TipeHarga,
			Nominal:         req.Nominal,
			MaksimalNominal: req.MaksimalNominal,
			Aktif:           aktif,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, produkdto.MapError("produk not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, produkdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, produkdto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, produkdto.MapError("produk not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, produkdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, produkdto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, produkdto.MapError("method not allowed"))
	}
}
