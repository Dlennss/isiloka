package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	branddto "pulsa2/internal/dto/brand"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type BrandController struct {
	svc  *service.BrandService
	base string
}

func NewBrandController(svc *service.BrandService, base string) *BrandController {
	return &BrandController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *BrandController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/brand"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			rows, err := h.svc.List(r.Context())
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, branddto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, branddto.MapList(rows))
		case http.MethodPost:
			var req branddto.Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, branddto.MapError("invalid json"))
				return
			}
			req.Nama = strings.TrimSpace(req.Nama)
			if req.Nama == "" {
				helper.WriteJSON(w, http.StatusBadRequest, branddto.MapError("nama required"))
				return
			}
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			id, err := h.svc.Create(r.Context(), req.Nama, aktif)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, branddto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, branddto.MapID(id))
		default:
			helper.WriteJSON(w, http.StatusMethodNotAllowed, branddto.MapError("method not allowed"))
		}
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, branddto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, branddto.MapError("brand not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, branddto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, branddto.MapItem(row))
	case http.MethodPut:
		var req branddto.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, branddto.MapError("invalid json"))
			return
		}
		req.Nama = strings.TrimSpace(req.Nama)
		if req.Nama == "" {
			helper.WriteJSON(w, http.StatusBadRequest, branddto.MapError("nama required"))
			return
		}
		aktif := true
		if req.Aktif != nil {
			aktif = *req.Aktif
		}
		if err := h.svc.Update(r.Context(), id, req.Nama, aktif); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, branddto.MapError("brand not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, branddto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, branddto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, branddto.MapError("brand not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, branddto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, branddto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, branddto.MapError("method not allowed"))
	}
}
