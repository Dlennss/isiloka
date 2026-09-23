package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	apporderdto "pulsa2/internal/dto/app_order"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AppOrderProviderAdminController struct {
	svc *service.AppOrderAdminService
}

func NewAppOrderProviderAdminController(svc *service.AppOrderAdminService) *AppOrderProviderAdminController {
	return &AppOrderProviderAdminController{svc: svc}
}

func (h *AppOrderProviderAdminController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
		return
	}
	filter, err := h.buildFilter(r)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}
	rows, err := h.svc.ListProviderTrx(r.Context(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, apporderdto.MapList(rows))
}

func (h *AppOrderProviderAdminController) buildFilter(r *http.Request) (repository.AppOrderProviderTrxListFilter, error) {
	q := r.URL.Query()
	filter := repository.AppOrderProviderTrxListFilter{
		Provider:  strings.TrimSpace(q.Get("provider")),
		Status:    strings.TrimSpace(q.Get("status")),
		RefID:     strings.TrimSpace(q.Get("ref_id")),
		InvoiceID: strings.TrimSpace(q.Get("invoice_id")),
		Limit:     20,
		Offset:    0,
	}
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return repository.AppOrderProviderTrxListFilter{}, errInvalid("limit")
		}
		filter.Limit = n
	}
	if v := strings.TrimSpace(q.Get("offset")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return repository.AppOrderProviderTrxListFilter{}, errInvalid("offset")
		}
		filter.Offset = n
	}
	if v := strings.TrimSpace(q.Get("date_from")); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return repository.AppOrderProviderTrxListFilter{}, errInvalid("date_from")
		}
		filter.DateFrom = &t
	}
	if v := strings.TrimSpace(q.Get("date_to")); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return repository.AppOrderProviderTrxListFilter{}, errInvalid("date_to")
		}
		t = t.AddDate(0, 0, 1)
		filter.DateTo = &t
	}
	return filter, nil
}
