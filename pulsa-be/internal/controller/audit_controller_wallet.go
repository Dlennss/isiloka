package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
)

func (h *AuditController) AdminProviderWalletMissingDebit(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit := helper.QueryInt(r, "limit", 10)
		offset := helper.QueryInt(r, "offset", 0)
		items, total, err := h.svc.AdminListProviderWalletMissingDebit(
			r.Context(),
			limit,
			offset,
			helper.QueryString(r, "provider"),
			helper.QueryString(r, "ref_id"),
			helper.QueryString(r, "from"),
			helper.QueryString(r, "to"),
		)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
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
			"filters": map[string]any{
				"provider": strings.TrimSpace(helper.QueryString(r, "provider")),
				"ref_id":   strings.TrimSpace(helper.QueryString(r, "ref_id")),
				"from":     strings.TrimSpace(helper.QueryString(r, "from")),
				"to":       strings.TrimSpace(helper.QueryString(r, "to")),
			},
		})
	case http.MethodPost:
		a, ok := helper.GetAuth(r.Context())
		role := ""
		if ok {
			role = helper.NormalizeRole(a.Role)
		}
		if !ok || (!helper.IsAdminLikeRole(role) && role != helper.RoleOperatorWallet) {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator_wallet only"))
			return
		}

		var in struct {
			TransaksiProviderID  int64   `json:"transaksi_provider_id"`
			TransaksiProviderIDs []int64 `json:"transaksi_provider_ids"`
			Ignore               bool    `json:"ignore"`
			Note                 string  `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("payload tidak valid"))
			return
		}
		ids := make([]int64, 0, len(in.TransaksiProviderIDs)+1)
		if in.TransaksiProviderID > 0 {
			ids = append(ids, in.TransaksiProviderID)
		}
		for _, id := range in.TransaksiProviderIDs {
			if id > 0 {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("transaksi_provider_id atau transaksi_provider_ids wajib diisi"))
			return
		}

		if in.Ignore {
			if !helper.IsAdminLikeRole(role) {
				helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
				return
			}
			if len(ids) > 1 {
				out, err := h.svc.IgnoreProviderWalletMissingDebitBulk(r.Context(), a.MemberID, ids, strings.TrimSpace(in.Note))
				if err != nil {
					helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
					return
				}
				helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
				return
			}
			if err := h.svc.IgnoreProviderWalletMissingDebit(r.Context(), a.MemberID, ids[0], strings.TrimSpace(in.Note)); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{
				"ok": true,
				"item": map[string]any{
					"ignored":               true,
					"transaksi_provider_id": ids[0],
				},
			})
			return
		}

		if len(ids) > 1 {
			out, err := h.svc.ResolveProviderWalletMissingDebitBulk(r.Context(), a.MemberID, ids)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
			return
		}

		out, err := h.svc.ResolveProviderWalletMissingDebit(r.Context(), a.MemberID, ids[0])
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}
