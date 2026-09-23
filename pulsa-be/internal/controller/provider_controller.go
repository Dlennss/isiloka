package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	providerdto "pulsa2/internal/dto/provider"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type ProviderController struct {
	svc  *service.ProviderService
	base string
}

func NewProviderController(svc *service.ProviderService, base string) *ProviderController {
	return &ProviderController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *ProviderController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/provider"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			rows, err := h.svc.List(r.Context())
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, providerdto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, providerdto.MapList(rows))
		case http.MethodPost:
			var req providerdto.Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, providerdto.MapError("invalid json"))
				return
			}
			req.Nama = strings.TrimSpace(req.Nama)
			if req.Nama == "" {
				helper.WriteJSON(w, http.StatusBadRequest, providerdto.MapError("nama required"))
				return
			}
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			id, err := h.svc.Create(r.Context(), req.Nama, aktif)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, providerdto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, providerdto.MapID(id))
		default:
			helper.WriteJSON(w, http.StatusMethodNotAllowed, providerdto.MapError("method not allowed"))
		}
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, providerdto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, providerdto.MapError("provider not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, providerdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, providerdto.MapItem(row))
	case http.MethodPut:
		var req providerdto.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, providerdto.MapError("invalid json"))
			return
		}
		req.Nama = strings.TrimSpace(req.Nama)
		if req.Nama == "" {
			helper.WriteJSON(w, http.StatusBadRequest, providerdto.MapError("nama required"))
			return
		}
		aktif := true
		if req.Aktif != nil {
			aktif = *req.Aktif
		}
		if err := h.svc.Update(r.Context(), id, req.Nama, aktif); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, providerdto.MapError("provider not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, providerdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, providerdto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, providerdto.MapError("provider not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, providerdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, providerdto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, providerdto.MapError("method not allowed"))
	}
}
