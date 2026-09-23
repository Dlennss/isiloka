package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type ProdukAppPricingController struct {
	svc  *service.ProdukAppPricingService
	base string
}

func NewProdukAppPricingController(svc *service.ProdukAppPricingService, base string) *ProdukAppPricingController {
	return &ProdukAppPricingController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *ProdukAppPricingController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/produk-app-pricing"
	if r.URL.Path == base {
		h.handleCollection(w, r)
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("not found"))
		return
	}
	h.handleItem(w, r, id)
}

func (h *ProdukAppPricingController) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit := helper.QueryInt(r, "limit", 10)
		offset := helper.QueryInt(r, "offset", 0)
		var aktif *bool
		if raw := strings.TrimSpace(strings.ToLower(helper.QueryString(r, "aktif"))); raw != "" {
			switch raw {
			case "true", "1":
				v := true
				aktif = &v
			case "false", "0":
				v := false
				aktif = &v
			default:
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("aktif harus true/false"))
				return
			}
		}

		items, total, err := h.svc.List(r.Context(), helper.QueryString(r, "q"), aktif, limit, offset)
		if err != nil {
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
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
		})
	case http.MethodPost:
		in, ok := h.decodeRequest(w, r, 0)
		if !ok {
			return
		}
		id, err := h.svc.Create(r.Context(), in)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *ProdukAppPricingController) handleItem(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("produk_app_pricing not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
	case http.MethodPut:
		in, ok := h.decodeRequest(w, r, id)
		if !ok {
			return
		}
		if err := h.svc.Update(r.Context(), in); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("produk_app_pricing not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("produk_app_pricing not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *ProdukAppPricingController) decodeRequest(w http.ResponseWriter, r *http.Request, id int64) (repository.ProdukAppPricingUpsertInput, bool) {
	var req struct {
		ProdukID  int64   `json:"produk_id"`
		Harga     int64   `json:"harga"`
		Aktif     bool    `json:"aktif"`
		FetchedAt *string `json:"fetched_at"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return repository.ProdukAppPricingUpsertInput{}, false
	}

	var fetchedAt *time.Time
	if req.FetchedAt != nil {
		raw := strings.TrimSpace(*req.FetchedAt)
		if raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("fetched_at harus RFC3339"))
				return repository.ProdukAppPricingUpsertInput{}, false
			}
			fetchedAt = &t
		}
	}

	return repository.ProdukAppPricingUpsertInput{
		ID:        id,
		ProdukID:  req.ProdukID,
		Harga:     req.Harga,
		Aktif:     req.Aktif,
		FetchedAt: fetchedAt,
	}, true
}
