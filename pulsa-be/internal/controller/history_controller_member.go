package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (h *HistoryController) MemberSaldo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	saldo, err := h.svc.GetSaldo(r.Context(), a.MemberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "saldo": saldo})
}

func (h *HistoryController) MemberMutasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	filter := repository.MutasiFilter{
		RefID:  helper.QueryString(r, "ref_id"),
		Arah:   helper.QueryString(r, "arah"),
		Date:   helper.QueryString(r, "date"),
		From:   helper.QueryString(r, "from"),
		To:     helper.QueryString(r, "to"),
		Limit:  helper.QueryInt(r, "limit", 50),
		Offset: helper.QueryInt(r, "offset", 0),
	}
	rows, total, err := h.svc.ListMutasi(r.Context(), a.MemberID, filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	totalPages := 1
	if filter.Limit > 0 {
		totalPages = int((total + int64(filter.Limit) - 1) / int64(filter.Limit))
		if totalPages <= 0 {
			totalPages = 1
		}
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"rows":        rows,
		"total":       total,
		"total_pages": totalPages,
	})
}

func (h *HistoryController) MemberMutasiDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/v1/history/mutasi/")
	idStr = strings.TrimSpace(strings.Trim(idStr, "/"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("id invalid"))
		return
	}

	item, err := h.svc.GetMutasiByID(r.Context(), a.MemberID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("data tidak ditemukan"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *HistoryController) MemberTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}

	limit := helper.QueryInt(r, "limit", 50)
	offset := helper.QueryInt(r, "offset", 0)
	rows, total, err := h.svc.ListTransaksi(
		r.Context(),
		a.MemberID,
		limit,
		offset,
		helper.QueryString(r, "q"),
		helper.QueryString(r, "status"),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
		if totalPages <= 0 {
			totalPages = 1
		}
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"rows":        rows,
		"total":       total,
		"total_pages": totalPages,
	})
}
