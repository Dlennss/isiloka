package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (h *BankController) Handle(w http.ResponseWriter, r *http.Request) {
	base := "/v1/admin/banks"
	a, okAuth := helper.GetAuth(r.Context())
	role := ""
	if okAuth {
		role = bankAuthRole(a.Role)
	}
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			if role != "admin" && role != "auditor" && role != "operator_wallet" && role != "operator_trx" {
				helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
				return
			}
			rows, err := h.svc.List(r.Context(), a.Role)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
				return
			}
			if role == "operator_trx" {
				rows = sanitizeBankRowsForOperatorTrx(rows)
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
		case http.MethodPost:
			if role != "admin" {
				helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
				return
			}
			var in repository.BankUpsertInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
				return
			}
			var actorID int64
			if okAuth {
				actorID = a.MemberID
			}
			id, err := h.svc.Create(r.Context(), actorID, in)
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
		if role != "admin" && role != "auditor" && role != "operator_wallet" && role != "operator_trx" {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
			return
		}
		item, err := h.svc.Get(r.Context(), id, a.Role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		if role == "operator_trx" {
			item = sanitizeBankRowForOperatorTrx(item)
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
	case http.MethodPut:
		var in repository.BankUpsertInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		in.ID = id

		if role == "operator_wallet" {
			if err := h.svc.EnsureVisibleToRole(r.Context(), id, a.Role); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
					return
				}
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			if err := h.svc.ToggleActive(r.Context(), id, in.Aktif); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
					return
				}
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
			return
		}

		if role != "admin" {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}

		if err := h.svc.Update(r.Context(), in); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	case http.MethodDelete:
		if role != "admin" {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
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
