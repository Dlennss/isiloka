package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type ProviderMerchantIDController struct {
	svc *service.ProviderMerchantIDService
}

func NewProviderMerchantIDController(svc *service.ProviderMerchantIDService) *ProviderMerchantIDController {
	return &ProviderMerchantIDController{svc: svc}
}

func (h *ProviderMerchantIDController) Handle(w http.ResponseWriter, r *http.Request) {
	base := "/v1/admin/provider-merchant-ids"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			provider := helper.QueryString(r, "provider")
			q := helper.QueryString(r, "q")
			aktifOnly := strings.EqualFold(helper.QueryString(r, "aktif"), "1") ||
				strings.EqualFold(helper.QueryString(r, "aktif"), "true")
			limit := helper.QueryInt(r, "limit", 50)
			offset := helper.QueryInt(r, "offset", 0)
			items, total, err := h.svc.List(r.Context(), provider, q, aktifOnly, limit, offset)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total, "limit": limit, "offset": offset})
		case http.MethodPost:
			var in repository.ProviderMerchantIDUpsertInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
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
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("merchant id not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
	case http.MethodPut:
		var in repository.ProviderMerchantIDUpsertInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		if err := h.svc.Update(r.Context(), id, in); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("merchant id not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("merchant id not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}
