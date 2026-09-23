package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
)

func (h *AuditController) AdminStatusMismatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	limit := helper.QueryInt(r, "limit", 10)
	offset := helper.QueryInt(r, "offset", 0)
	items, total, err := h.svc.AdminListStatusMismatch(
		r.Context(),
		limit,
		offset,
		helper.QueryString(r, "provider"),
		helper.QueryString(r, "status_member"),
		helper.QueryString(r, "mismatch_type"),
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
			"provider":      strings.TrimSpace(helper.QueryString(r, "provider")),
			"status_member": strings.TrimSpace(helper.QueryString(r, "status_member")),
			"mismatch_type": strings.TrimSpace(helper.QueryString(r, "mismatch_type")),
			"from":          strings.TrimSpace(helper.QueryString(r, "from")),
			"to":            strings.TrimSpace(helper.QueryString(r, "to")),
		},
	})
}

func (h *AuditController) AdminGuestRefundMissing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	limit := helper.QueryInt(r, "limit", 10)
	offset := helper.QueryInt(r, "offset", 0)
	items, total, err := h.svc.AdminListGuestRefundMissing(
		r.Context(),
		limit,
		offset,
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		helper.QueryString(r, "invoice_id"),
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
			"from":       strings.TrimSpace(helper.QueryString(r, "from")),
			"to":         strings.TrimSpace(helper.QueryString(r, "to")),
			"invoice_id": strings.TrimSpace(helper.QueryString(r, "invoice_id")),
		},
	})
}

func (h *AuditController) AdminProviderEmptyResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	limit := helper.QueryInt(r, "limit", 10)
	offset := helper.QueryInt(r, "offset", 0)
	items, total, err := h.svc.AdminListProviderEmptyResponse(
		r.Context(),
		limit,
		offset,
		helper.QueryString(r, "provider"),
		helper.QueryString(r, "ref_id"),
		helper.QueryString(r, "kode_produk"),
		helper.QueryString(r, "tujuan"),
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
			"provider":    strings.TrimSpace(helper.QueryString(r, "provider")),
			"ref_id":      strings.TrimSpace(helper.QueryString(r, "ref_id")),
			"kode_produk": strings.TrimSpace(helper.QueryString(r, "kode_produk")),
			"tujuan":      strings.TrimSpace(helper.QueryString(r, "tujuan")),
			"from":        strings.TrimSpace(helper.QueryString(r, "from")),
			"to":          strings.TrimSpace(helper.QueryString(r, "to")),
		},
	})
}

func (h *AuditController) AdminProviderSuccessSuspiciousMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		a, ok := helper.GetAuth(r.Context())
		if !ok || a.MemberID <= 0 {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("auth required"))
			return
		}
		if !helper.IsAdminLikeRole(a.Role) {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}
		var in struct {
			Action               string  `json:"action"`
			TransaksiProviderID  int64   `json:"transaksi_provider_id"`
			TransaksiProviderIDs []int64 `json:"transaksi_provider_ids"`
			Nominal              int64   `json:"nominal"`
			Fee                  int64   `json:"fee"`
			Note                 string  `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("payload tidak valid"))
			return
		}
		action := strings.TrimSpace(strings.ToLower(in.Action))
		if action == "resend_bifastopen2" {
			if in.TransaksiProviderID <= 0 {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("transaksi_provider_id wajib diisi"))
				return
			}
			out, err := h.svc.ResendProviderSuccessSuspiciousMessageToBIFASTOPEN2(r.Context(), a.MemberID, in.TransaksiProviderID)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
			return
		}
		if action == "settle_bank_debit" {
			if in.TransaksiProviderID <= 0 {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("transaksi_provider_id wajib diisi"))
				return
			}
			out, err := h.svc.SettleProviderSuccessSuspiciousMessageWithBankDebit(
				r.Context(),
				a.MemberID,
				in.TransaksiProviderID,
				in.Nominal,
				in.Fee,
				strings.TrimSpace(in.Note),
			)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
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
		if len(ids) > 1 {
			out, err := h.svc.ResolveProviderSuccessSuspiciousMessageBulk(
				r.Context(),
				a.MemberID,
				ids,
				strings.TrimSpace(in.Note),
			)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
			return
		}
		out, err := h.svc.ResolveProviderSuccessSuspiciousMessage(
			r.Context(),
			a.MemberID,
			ids[0],
			strings.TrimSpace(in.Note),
		)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
		return
	}

	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	limit := helper.QueryInt(r, "limit", 10)
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	offset := helper.QueryInt(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}
	includeTotal := false
	switch strings.ToLower(strings.TrimSpace(helper.QueryString(r, "include_total"))) {
	case "1", "true", "yes":
		includeTotal = true
	}
	switch strings.ToLower(strings.TrimSpace(helper.QueryString(r, "fast_page"))) {
	case "1", "true", "yes":
		includeTotal = false
	}
	items, total, hasNext, err := h.svc.AdminListProviderSuccessSuspiciousMessage(
		r.Context(),
		limit,
		offset,
		helper.QueryString(r, "provider"),
		helper.QueryString(r, "ref_id"),
		helper.QueryString(r, "resolve_status"),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		includeTotal,
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
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
		"has_next":    hasNext,
		"total_exact": includeTotal,
		"filters": map[string]any{
			"provider":       strings.TrimSpace(helper.QueryString(r, "provider")),
			"ref_id":         strings.TrimSpace(helper.QueryString(r, "ref_id")),
			"resolve_status": strings.TrimSpace(helper.QueryString(r, "resolve_status")),
			"from":           strings.TrimSpace(helper.QueryString(r, "from")),
			"to":             strings.TrimSpace(helper.QueryString(r, "to")),
		},
	})
}
