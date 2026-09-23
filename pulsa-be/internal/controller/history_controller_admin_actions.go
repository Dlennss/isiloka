package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (h *HistoryController) AdminMutasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	items, total, err := h.svc.AdminListMutasiByMember(
		r.Context(),
		helper.QueryInt64(r, "member_id", 0),
		helper.QueryInt(r, "limit", 50),
		helper.QueryInt(r, "offset", 0),
		helper.QueryString(r, "ref_id"),
		helper.QueryString(r, "arah"),
		helper.QueryString(r, "date"),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		strings.EqualFold(strings.TrimSpace(helper.QueryString(r, "wallet_only")), "1") ||
			strings.EqualFold(strings.TrimSpace(helper.QueryString(r, "wallet_only")), "true"),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total})
}

func (h *HistoryController) AdminTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	items, total, err := h.svc.AdminListTransaksi(
		r.Context(),
		helper.QueryInt64(r, "member_id", 0),
		helper.QueryInt(r, "limit", 50),
		helper.QueryInt(r, "offset", 0),
		helper.QueryString(r, "q"),
		helper.QueryString(r, "status"),
		helper.QueryString(r, "kode_produk"),
		helper.QueryString(r, "ref_id"),
		helper.QueryString(r, "dest"),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		helper.QueryString(r, "member_role"),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total})
}

func (h *HistoryController) AdminCancelTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorTrx && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator/wallet only"))
		return
	}

	var in struct {
		TrxID  int64   `json:"trx_id"`
		TrxIDs []int64 `json:"trx_ids"`
		Reason string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	allowSuccessCancel := helper.IsAdminLikeRole(a.Role)
	if len(in.TrxIDs) > 0 {
		items, failed, err := h.svc.AdminCancelPendingTransaksiBulk(r.Context(), a.MemberID, in.TrxIDs, strings.TrimSpace(in.Reason), allowSuccessCancel)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{
			"ok":              true,
			"items":           items,
			"failed":          failed,
			"processed_count": len(items),
			"failed_count":    len(failed),
		})
		return
	}

	item, err := h.svc.AdminCancelPendingTransaksi(r.Context(), a.MemberID, in.TrxID, strings.TrimSpace(in.Reason), allowSuccessCancel)
	if err != nil {
		if errors.Is(err, repository.ErrTransaksiNotPending) {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *HistoryController) AdminCompleteTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorTrx && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator/wallet only"))
		return
	}

	var in struct {
		TrxID  int64   `json:"trx_id"`
		TrxIDs []int64 `json:"trx_ids"`
		Reason string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	if len(in.TrxIDs) > 0 {
		items, failed, err := h.svc.AdminCompletePendingTransaksiBulk(r.Context(), a.MemberID, in.TrxIDs, strings.TrimSpace(in.Reason))
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{
			"ok":              true,
			"items":           items,
			"failed":          failed,
			"processed_count": len(items),
			"failed_count":    len(failed),
		})
		return
	}

	item, err := h.svc.AdminCompletePendingTransaksi(r.Context(), a.MemberID, in.TrxID, strings.TrimSpace(in.Reason))
	if err != nil {
		if errors.Is(err, repository.ErrTransaksiNotPending) {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *HistoryController) AdminSendTransaksiCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorTrx && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator/wallet only"))
		return
	}

	var in struct {
		TrxID  int64   `json:"trx_id"`
		TrxIDs []int64 `json:"trx_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	if len(in.TrxIDs) > 0 {
		items, failed, err := h.svc.AdminSendTransaksiCallbackBulk(r.Context(), a.MemberID, in.TrxIDs)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{
			"ok":              true,
			"items":           items,
			"failed":          failed,
			"processed_count": len(items),
			"failed_count":    len(failed),
		})
		return
	}

	item, err := h.svc.AdminSendTransaksiCallback(r.Context(), a.MemberID, in.TrxID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *HistoryController) AdminResendTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorTrx && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator/wallet only"))
		return
	}

	var in struct {
		ProviderTrxID int64  `json:"provider_trx_id"`
		TrxID         int64  `json:"trx_id"`
		Mode          string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	mode := strings.TrimSpace(strings.ToLower(in.Mode))
	if mode == "refund_no_success" {
		if !helper.IsAdminLikeRole(a.Role) {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}
		item, err := h.svc.AdminResendRefundNoSuccessTransaksi(r.Context(), a.MemberID, in.ProviderTrxID)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
		return
	}

	item, err := h.svc.AdminResendPendingTransaksi(r.Context(), a.MemberID, in.ProviderTrxID, in.TrxID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}
