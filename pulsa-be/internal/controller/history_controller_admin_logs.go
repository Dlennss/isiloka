package controller

import (
	"net/http"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (h *HistoryController) AdminTransaksiStatusLogsManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorTrx && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator/wallet only"))
		return
	}

	filter := repository.TrxMemberStatusLogFilter{
		TrxID:      helper.QueryInt64(r, "trx_id", 0),
		RefID:      helper.QueryString(r, "ref_id"),
		DiubahOleh: helper.QueryInt64(r, "diubah_oleh", 0),
		From:       helper.QueryString(r, "from"),
		To:         helper.QueryString(r, "to"),
		Limit:      helper.QueryInt(r, "limit", 10),
		Offset:     helper.QueryInt(r, "offset", 0),
	}
	if a.Role == "operator_trx" {
		filter.DiubahOleh = a.MemberID
	}

	items, total, totalPages, err := h.svc.AdminListTransaksiStatusLogsManual(r.Context(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	limit := helper.QueryInt(r, "limit", 10)
	if limit <= 0 {
		limit = 10
	}
	page := 1 + (helper.QueryInt(r, "offset", 0) / limit)
	if page < 1 {
		page = 1
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"items":       items,
		"total":       total,
		"total_pages": totalPages,
		"page":        page,
		"limit":       limit,
		"manual_only": true,
	})
}

func (h *HistoryController) AdminTransaksiStatusLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorTrx && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator/wallet only"))
		return
	}

	filter := repository.TrxMemberStatusLogFilter{
		TrxID:      helper.QueryInt64(r, "trx_id", 0),
		RefID:      helper.QueryString(r, "ref_id"),
		DiubahOleh: helper.QueryInt64(r, "diubah_oleh", 0),
		From:       helper.QueryString(r, "from"),
		To:         helper.QueryString(r, "to"),
		Limit:      helper.QueryInt(r, "limit", 10),
		Offset:     helper.QueryInt(r, "offset", 0),
	}
	if a.Role == "operator_trx" {
		filter.DiubahOleh = a.MemberID
	}

	items, total, totalPages, err := h.svc.AdminListTransaksiStatusLogs(r.Context(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	limit := helper.QueryInt(r, "limit", 10)
	if limit <= 0 {
		limit = 10
	}
	offset := helper.QueryInt(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}
	page := (offset / limit) + 1

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"items":       items,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
		"page":        page,
		"total_pages": totalPages,
		"filters": map[string]any{
			"trx_id":      helper.QueryInt64(r, "trx_id", 0),
			"ref_id":      helper.QueryString(r, "ref_id"),
			"diubah_oleh": helper.QueryInt64(r, "diubah_oleh", 0),
			"from":        helper.QueryString(r, "from"),
			"to":          helper.QueryString(r, "to"),
		},
	})
}
