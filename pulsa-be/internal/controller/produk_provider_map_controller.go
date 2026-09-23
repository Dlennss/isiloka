package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	produkprovidermapdto "pulsa2/internal/dto/produk_provider_map"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type ProdukProviderMapController struct {
	svc  *service.ProdukProviderMapService
	base string
}

func NewProdukProviderMapController(svc *service.ProdukProviderMapService, base string) *ProdukProviderMapController {
	return &ProdukProviderMapController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *ProdukProviderMapController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/produk-provider-map"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			q := helper.QueryString(r, "q")
			rows, err := h.svc.List(r.Context(), q)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, produkprovidermapdto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, produkprovidermapdto.MapList(rows))
		case http.MethodPost:
			var req produkprovidermapdto.Request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, produkprovidermapdto.MapError("invalid json"))
				return
			}
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			id, err := h.svc.Create(r.Context(), repository.ProdukProviderMapUpsertInput{
				ProdukID:        req.ProdukID,
				Provider:        req.Provider,
				KodeProvider:    req.KodeProvider,
				SpecialCode:     req.SpecialCode,
				Mode:            req.Mode,
				MinimalNominal:  req.MinimalNominal,
				MaksimalNominal: req.MaksimalNominal,
				FeeRp:           req.FeeRp,
				JamBuka:         req.JamBuka,
				JamTutup:        req.JamTutup,
				Aktif:           aktif,
			})
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, produkprovidermapdto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, produkprovidermapdto.MapID(id))
		default:
			helper.WriteJSON(w, http.StatusMethodNotAllowed, produkprovidermapdto.MapError("method not allowed"))
		}
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, produkprovidermapdto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, produkprovidermapdto.MapError("produk_provider_map not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, produkprovidermapdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, produkprovidermapdto.MapItem(row))
	case http.MethodPut:
		var req produkprovidermapdto.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, produkprovidermapdto.MapError("invalid json"))
			return
		}
		aktif := true
		if req.Aktif != nil {
			aktif = *req.Aktif
		}
		if err := h.svc.Update(r.Context(), repository.ProdukProviderMapUpsertInput{
			ID:              id,
			ProdukID:        req.ProdukID,
			Provider:        req.Provider,
			KodeProvider:    req.KodeProvider,
			SpecialCode:     req.SpecialCode,
			Mode:            req.Mode,
			MinimalNominal:  req.MinimalNominal,
			MaksimalNominal: req.MaksimalNominal,
			FeeRp:           req.FeeRp,
			JamBuka:         req.JamBuka,
			JamTutup:        req.JamTutup,
			Aktif:           aktif,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, produkprovidermapdto.MapError("produk_provider_map not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, produkprovidermapdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, produkprovidermapdto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, produkprovidermapdto.MapError("produk_provider_map not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, produkprovidermapdto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, produkprovidermapdto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, produkprovidermapdto.MapError("method not allowed"))
	}
}

func (h *ProdukProviderMapController) HandleToggleActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch && r.Method != http.MethodPut {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, produkprovidermapdto.MapError("method not allowed"))
		return
	}

	base := h.base + "/produk-provider-map/"
	path := strings.TrimSuffix(strings.TrimRight(r.URL.Path, "/"), "/toggle-active")
	id, ok := helper.ParseIDFromPath(path, strings.TrimRight(base, "/"))
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, produkprovidermapdto.MapError("not found"))
		return
	}

	var req produkprovidermapdto.ToggleAktifRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, produkprovidermapdto.MapError("invalid json"))
		return
	}

	if err := h.svc.UpdateAktif(r.Context(), id, req.Aktif); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, produkprovidermapdto.MapError("produk_provider_map not found"))
			return
		}
		helper.WriteJSON(w, http.StatusInternalServerError, produkprovidermapdto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, produkprovidermapdto.MapOK())
}
