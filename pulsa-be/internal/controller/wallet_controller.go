package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type WalletController struct {
	svc *service.WalletService
}

func NewWalletController(svc *service.WalletService) *WalletController {
	return &WalletController{svc: svc}
}

func (h *WalletController) AdminAdjustMemberWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator wallet only"))
		return
	}

	var in struct {
		MemberID  int64  `json:"member_id"`
		Amount    int64  `json:"amount"`
		Direction string `json:"direction"`
		Note      string `json:"note"`
		Reason    string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	refID, err := h.svc.AdminAdjustMemberWallet(r.Context(), a.MemberID, in.MemberID, in.Amount, in.Direction, in.Note, in.Reason)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "ref_id": refID})
}

func (h *WalletController) AdminListProviderWallets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	rows, err := h.svc.ListProviderWallets(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusBadGateway, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "data": rows})
}

func (h *WalletController) AdminDepositProviderWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator wallet only"))
		return
	}

	var in struct {
		BankID   int64  `json:"bank_id"`
		Provider string `json:"provider"`
		Amount   int64  `json:"amount"`
		AdminFee int64  `json:"admin_fee"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	refID, bankSaldo, saldo, err := h.svc.DepositProviderWallet(r.Context(), a.MemberID, a.Role, in.BankID, in.Provider, in.Amount, in.AdminFee, in.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "refid": refID, "bank_saldo": bankSaldo, "saldo_internal": saldo, "admin_fee": in.AdminFee, "bank_debit": in.Amount + in.AdminFee})
}

func (h *WalletController) AdminAdjustProviderWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	a, ok := helper.GetAuth(r.Context())
	if !ok || !helper.IsAdminLikeRole(a.Role) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
		return
	}

	var in struct {
		Provider string `json:"provider"`
		Amount   int64  `json:"amount"`
		Mode     string `json:"mode"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	out, err := h.svc.AdjustProviderWallet(r.Context(), a.MemberID, in.Provider, in.Amount, in.Mode, in.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"mode":           out.Mode,
		"unchanged":      out.Unchanged,
		"arah":           out.Arah,
		"amount_applied": out.AmountApplied,
		"refid":          out.RefID,
		"saldo_internal": out.SaldoInternal,
	})
}

func (h *WalletController) AdminProviderWalletHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	provider := strings.TrimSpace(strings.ToLower(helper.QueryString(r, "provider")))
	arah := strings.TrimSpace(strings.ToLower(helper.QueryString(r, "arah")))
	refID := strings.TrimSpace(helper.QueryString(r, "ref_id"))
	from := strings.TrimSpace(helper.QueryString(r, "from"))
	to := strings.TrimSpace(helper.QueryString(r, "to"))
	limit := helper.QueryInt(r, "limit", 50)
	offset := helper.QueryInt(r, "offset", 0)

	items, total, err := h.svc.ProviderHistory(r.Context(), provider, arah, refID, from, to, limit, offset)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"items":  items,
		"total":  total,
		"ref_id": refID,
		"arah":   arah,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *WalletController) AdminMemberDepositHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	memberID := helper.QueryInt64(r, "member_id", 0)
	arah := strings.TrimSpace(strings.ToLower(helper.QueryString(r, "arah")))
	refID := strings.TrimSpace(helper.QueryString(r, "ref_id"))
	from := strings.TrimSpace(helper.QueryString(r, "from"))
	to := strings.TrimSpace(helper.QueryString(r, "to"))
	limit := helper.QueryInt(r, "limit", 50)
	offset := helper.QueryInt(r, "offset", 0)

	items, total, err := h.svc.MemberDepositHistory(r.Context(), memberID, arah, refID, from, to, limit, offset)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"items":     items,
		"total":     total,
		"member_id": memberID,
		"arah":      arah,
		"ref_id":    refID,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *WalletController) AdminProviderDepositHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	provider := strings.TrimSpace(strings.ToLower(helper.QueryString(r, "provider")))
	refID := strings.TrimSpace(helper.QueryString(r, "ref_id"))
	from := strings.TrimSpace(helper.QueryString(r, "from"))
	to := strings.TrimSpace(helper.QueryString(r, "to"))
	limit := helper.QueryInt(r, "limit", 50)
	offset := helper.QueryInt(r, "offset", 0)

	items, total, err := h.svc.ProviderDepositHistory(r.Context(), provider, refID, from, to, limit, offset)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"items":    items,
		"total":    total,
		"provider": provider,
		"ref_id":   refID,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *WalletController) AdminWalletActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	a, ok := helper.GetAuth(r.Context())
	if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator wallet only"))
		return
	}

	from := strings.TrimSpace(helper.QueryString(r, "from"))
	to := strings.TrimSpace(helper.QueryString(r, "to"))
	limit := helper.QueryInt(r, "limit", 10)
	actorID := a.MemberID
	if helper.IsAdminLikeRole(a.Role) || helper.NormalizeRole(a.Role) == helper.RoleOperatorWallet {
		if v := helper.QueryInt64(r, "actor_id", 0); v > 0 {
			actorID = v
		}
	}

	out, err := h.svc.ActorActivity(r.Context(), actorID, from, to, limit)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"item": out,
	})
}
