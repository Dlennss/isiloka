package controller

import (
	"net/http"
	"time"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AdminBusinessReportController struct {
	svc *service.AdminBusinessReportService
}

func NewAdminBusinessReportController(svc *service.AdminBusinessReportService) *AdminBusinessReportController {
	return &AdminBusinessReportController{svc: svc}
}

func (h *AdminBusinessReportController) CommissionBySource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}

	items, err := h.svc.ListCommissionBySource(r.Context(), repository.AdminCommissionBySourceArgs{
		Scope:   helper.QueryString(r, "scope"),
		Limit:   helper.QueryInt(r, "limit", 1000),
		Offset:  helper.QueryInt(r, "offset", 0),
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

func (h *AdminBusinessReportController) DailyBusiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}

	items, err := h.svc.ListDailyBusiness(r.Context(), repository.AdminDailyBusinessArgs{
		Scope:   helper.QueryString(r, "scope"),
		Months:  helper.QueryInt(r, "months", 3),
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
	})
	if err != nil {
		helper.SafeErrorResponse(w, http.StatusBadRequest, "Gagal memuat laporan bisnis.", err, "admin daily business report")
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
}

func (h *AdminBusinessReportController) RefreshDailyBusinessCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	item, err := h.svc.RefreshDailyBusinessCache(
		r.Context(),
		helper.QueryInt(r, "months", 3),
		helper.QueryInt(r, "days", 0),
	)
	if err != nil {
		helper.SafeErrorResponse(w, http.StatusBadRequest, "Gagal update data laporan bisnis.", err, "admin daily business cache refresh")
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}
