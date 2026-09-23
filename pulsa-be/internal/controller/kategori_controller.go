package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	kategoridto "pulsa2/internal/dto/kategori"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type KategoriController struct {
	svc  *service.KategoriService
	base string
}

func NewKategoriController(svc *service.KategoriService, base string) *KategoriController {
	return &KategoriController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *KategoriController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/kategori"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			rows, err := h.svc.List(r.Context())
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, kategoridto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, kategoridto.MapList(rows))
		case http.MethodPost:
			var req kategoridto.Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, kategoridto.MapError("invalid json"))
				return
			}
			req.Nama = strings.TrimSpace(req.Nama)
			if req.Nama == "" {
				helper.WriteJSON(w, http.StatusBadRequest, kategoridto.MapError("nama required"))
				return
			}
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			id, err := h.svc.Create(r.Context(), req.Nama, aktif)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, kategoridto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, kategoridto.MapID(id))
		default:
			helper.WriteJSON(w, http.StatusMethodNotAllowed, kategoridto.MapError("method not allowed"))
		}
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, kategoridto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, kategoridto.MapError("kategori not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, kategoridto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, kategoridto.MapItem(row))
	case http.MethodPut:
		var req kategoridto.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, kategoridto.MapError("invalid json"))
			return
		}
		req.Nama = strings.TrimSpace(req.Nama)
		if req.Nama == "" {
			helper.WriteJSON(w, http.StatusBadRequest, kategoridto.MapError("nama required"))
			return
		}
		aktif := true
		if req.Aktif != nil {
			aktif = *req.Aktif
		}
		if err := h.svc.Update(r.Context(), id, req.Nama, aktif); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, kategoridto.MapError("kategori not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, kategoridto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, kategoridto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, kategoridto.MapError("kategori not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, kategoridto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, kategoridto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, kategoridto.MapError("method not allowed"))
	}
}
