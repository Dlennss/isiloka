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

type KategoriFeeAppController struct {
	svc  *service.KategoriFeeAppService
	base string
}

func NewKategoriFeeAppController(svc *service.KategoriFeeAppService, base string) *KategoriFeeAppController {
	return &KategoriFeeAppController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *KategoriFeeAppController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/kategori-fee-app"
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

func (h *KategoriFeeAppController) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := h.svc.List(r.Context(), helper.QueryString(r, "q"))
		if err != nil {
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
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

func (h *KategoriFeeAppController) handleItem(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("kategori_fee_app not found"))
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
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("kategori_fee_app not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("kategori_fee_app not found"))
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

func (h *KategoriFeeAppController) decodeRequest(w http.ResponseWriter, r *http.Request, id int64) (repository.KategoriFeeAppUpsertInput, bool) {
	var req struct {
		KategoriID int64 `json:"kategori_id"`
		FeeMaster  int64 `json:"fee_master"`
		FeeAgent   int64 `json:"fee_agent"`
		FeeUser    int64 `json:"fee_user"`
		FeeNonUser int64 `json:"fee_non_user"`
		Aktif      bool  `json:"aktif"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return repository.KategoriFeeAppUpsertInput{}, false
	}
	return repository.KategoriFeeAppUpsertInput{
		ID:         id,
		KategoriID: req.KategoriID,
		FeeMaster:  req.FeeMaster,
		FeeAgent:   req.FeeAgent,
		FeeUser:    req.FeeUser,
		FeeNonUser: req.FeeNonUser,
		Aktif:      req.Aktif,
	}, true
}
