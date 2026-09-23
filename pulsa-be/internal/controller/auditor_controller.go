package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AuditorController struct {
	svc *service.AuditorService
}

func NewAuditorController(svc *service.AuditorService) *AuditorController {
	return &AuditorController{svc: svc}
}

func (h *AuditorController) TradingSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}
	items, err := h.svc.ListTradingSummary(r.Context(), repository.AuditorTradingSummaryArgs{
		Scope:   helper.QueryString(r, "scope"),
		Period:  helper.QueryString(r, "period"),
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
}

func (h *AuditorController) TradingDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}
	items, total, err := h.svc.ListTradingDetails(r.Context(), repository.AuditorTradingDetailArgs{
		Scope:   helper.QueryString(r, "scope"),
		RefID:   helper.QueryString(r, "ref_id"),
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
		Limit:   helper.QueryInt(r, "limit", 100),
		Offset:  helper.QueryInt(r, "offset", 0),
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total})
}

func (h *AuditorController) Finance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}
	items, total, err := h.svc.ListFinance(r.Context(), a.Role, repository.AuditorFinanceArgs{
		BankID:  helper.QueryInt64(r, "bank_id", 0),
		Type:    helper.QueryString(r, "type"),
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
		Limit:   helper.QueryInt(r, "limit", 100),
		Offset:  helper.QueryInt(r, "offset", 0),
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total})
}

func (h *AuditorController) ProfitLoss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	month := helper.QueryString(r, "month")
	loc, _ := time.LoadLocation("Asia/Jakarta")
	target := time.Now().In(loc)
	if month != "" {
		t, err := time.ParseInLocation("2006-01", month, loc)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("month invalid"))
			return
		}
		target = t
	}
	item, err := h.svc.GetProfitLoss(r.Context(), a.Role, repository.AuditorProfitLossArgs{Month: target})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *AuditorController) OpeningBalances(w http.ResponseWriter, r *http.Request) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	switch r.Method {
	case http.MethodGet:
		month := helper.QueryString(r, "month")
		target := time.Now().In(loc)
		if month != "" {
			t, err := time.ParseInLocation("2006-01", month, loc)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("month invalid"))
				return
			}
			target = t
		}
		items, err := h.svc.ListOpeningBalances(r.Context(), target)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
	case http.MethodPost:
		a, ok := helper.GetAuth(r.Context())
		if !ok || !helper.IsAdminLikeRole(a.Role) {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}
		var in struct {
			Month string                                 `json:"month"`
			Items []repository.OpeningBalanceUpsertInput `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		target := time.Now().In(loc)
		if in.Month != "" {
			t, err := time.ParseInLocation("2006-01", in.Month, loc)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("month invalid"))
				return
			}
			target = t
		}
		if err := h.svc.UpsertOpeningBalances(r.Context(), a.MemberID, target, in.Items); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *AuditorController) InternalFinance(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a, ok := helper.GetAuth(r.Context())
		if !ok {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
			return
		}
		items, total, err := h.svc.ListInternalFinanceEntries(
			r.Context(),
			a.Role,
			helper.QueryInt64(r, "bank_id", 0),
			helper.QueryString(r, "entry_type"),
			helper.QueryString(r, "from"),
			helper.QueryString(r, "to"),
			helper.QueryInt(r, "limit", 100),
			helper.QueryInt(r, "offset", 0),
		)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total})
	case http.MethodPost:
		a, ok := helper.GetAuth(r.Context())
		if !ok || (!helper.IsAdminLikeRole(a.Role) && helper.NormalizeRole(a.Role) != helper.RoleOperatorWallet) {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin/operator wallet only"))
			return
		}
		var in struct {
			EntryType    string `json:"entry_type"`
			Category     string `json:"category"`
			BankID       int64  `json:"bank_id"`
			Amount       int64  `json:"amount"`
			Fee          int64  `json:"fee"`
			Counterparty string `json:"counterparty"`
			Note         string `json:"note"`
			OccurredAt   string `json:"occurred_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		var occurredAt time.Time
		if strings.TrimSpace(in.OccurredAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(in.OccurredAt))
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("occurred_at invalid"))
				return
			}
			occurredAt = t
		}
		item, err := h.svc.CreateInternalFinanceEntry(r.Context(), a.MemberID, a.Role, repository.InternalFinanceCreateInput{
			EntryType:    in.EntryType,
			Category:     in.Category,
			BankID:       in.BankID,
			Amount:       in.Amount,
			Fee:          in.Fee,
			Counterparty: in.Counterparty,
			Note:         in.Note,
			OccurredAt:   occurredAt,
		})
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}
