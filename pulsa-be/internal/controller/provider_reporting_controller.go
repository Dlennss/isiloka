package controller

import (
	"net/http"
	"time"

	commondto "pulsa2/internal/dto/common"
	providerreportingdto "pulsa2/internal/dto/provider_reporting"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type ProviderReportingController struct {
	svc *service.ProviderReportingService
}

var providerReportingJakartaLocation = time.FixedZone("Asia/Jakarta", 7*60*60)

func NewProviderReportingController(svc *service.ProviderReportingService) *ProviderReportingController {
	return &ProviderReportingController{svc: svc}
}

func (h *ProviderReportingController) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	memberID := helper.QueryInt64(r, "member_id", 0)
	limit := helper.QueryInt(r, "limit", 100)
	offset := helper.QueryInt(r, "offset", 0)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	kodeRespon := helper.QueryString(r, "rc")
	if kodeRespon == "" {
		kodeRespon = helper.QueryString(r, "kode_respon")
	}

	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}

	items, total, err := h.svc.ListTransactions(r.Context(), repository.ProviderTransactionListArgs{
		Provider:   helper.QueryString(r, "provider"),
		Q:          helper.QueryString(r, "q"),
		MemberID:   memberID,
		KodeProduk: helper.QueryString(r, "kode_produk"),
		RefID:      helper.QueryString(r, "ref_id"),
		Tujuan:     helper.QueryString(r, "tujuan"),
		KodeRespon: kodeRespon,
		Status:     helper.QueryString(r, "status"),
		From:       from,
		HasFrom:    hasFrom,
		To:         to,
		HasTo:      hasTo,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, providerreportingdto.MapListTransactions(items, total))
}

func (h *ProviderReportingController) Analytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}

	data, err := h.svc.Analytics(r.Context(), repository.ProviderAnalyticsArgs{
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, providerreportingdto.MapAnalytics(data))
}

func (h *ProviderReportingController) RefreshAnalyticsCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})

	item, err := h.svc.RefreshAnalyticsCache(
		r.Context(),
		helper.QueryInt(r, "months", 3),
		helper.QueryInt(r, "days", 0),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
}

func (h *ProviderReportingController) DailyProductSuccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if !hasFrom {
		now := time.Now().In(providerReportingJakartaLocation)
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, providerReportingJakartaLocation)
		hasFrom = true
	}
	if hasTo {
		to = to.Add(24 * time.Hour)
	} else {
		to = from.Add(24 * time.Hour)
		hasTo = true
	}

	limit := helper.QueryInt(r, "limit", 500)
	offset := helper.QueryInt(r, "offset", 0)
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	data, err := h.svc.DailyProductSuccess(r.Context(), repository.DailyProductSuccessArgs{
		Q:       helper.QueryString(r, "q"),
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, providerreportingdto.MapDailyProductSuccess(data))
}

func (h *ProviderReportingController) ListAnomalies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	limit := helper.QueryInt(r, "limit", 100)
	offset := helper.QueryInt(r, "offset", 0)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	from, hasFrom := helper.QueryDate(r, "from")
	to, hasTo := helper.QueryDate(r, "to")
	if hasTo {
		to = to.Add(24 * time.Hour)
	}

	kodeRespon := helper.QueryString(r, "rc")
	if kodeRespon == "" {
		kodeRespon = helper.QueryString(r, "kode_respon")
	}

	items, total, err := h.svc.ListAnomalies(r.Context(), repository.ProviderAnomalyListArgs{
		Provider:   helper.QueryString(r, "provider"),
		Q:          helper.QueryString(r, "q"),
		RefID:      helper.QueryString(r, "ref_id"),
		KodeRespon: kodeRespon,
		Status:     helper.QueryString(r, "status"), // duplicate / fraud
		From:       from,
		HasFrom:    hasFrom,
		To:         to,
		HasTo:      hasTo,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, providerreportingdto.MapListAnomalies(items, total))
}
