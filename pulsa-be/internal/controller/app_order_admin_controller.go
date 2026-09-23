package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	apporderdto "pulsa2/internal/dto/app_order"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AppOrderAdminController struct {
	svc      *service.AppOrderService
	adminSvc *service.AppOrderAdminService
	base     string
}

func NewAppOrderAdminController(svc *service.AppOrderService, adminSvc *service.AppOrderAdminService, base string) *AppOrderAdminController {
	return &AppOrderAdminController{svc: svc, adminSvc: adminSvc, base: strings.TrimRight(base, "/")}
}

func (h *AppOrderAdminController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/orders"
	if r.URL.Path == base {
		h.list(w, r)
		return
	}

	if !strings.HasPrefix(r.URL.Path, base+"/") {
		helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("not found"))
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
		return
	}
	invoiceID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, base+"/"))
	if invoiceID == "" || strings.Contains(invoiceID, "/") {
		helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("not found"))
		return
	}
	h.detail(w, r, invoiceID)
}

func (h *AppOrderAdminController) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
		return
	}
	filter, err := h.buildFilter(r)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}

	rows, err := h.svc.List(r.Context(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, apporderdto.MapList(rows))
}

func (h *AppOrderAdminController) detail(w http.ResponseWriter, r *http.Request, invoiceID string) {
	row, err := h.adminSvc.GetDetailByInvoiceID(r.Context(), invoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("order not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, apporderdto.MapItem(row))
}

func (h *AppOrderAdminController) buildFilter(r *http.Request) (repository.AppOrderListFilter, error) {
	q := r.URL.Query()
	filter := repository.AppOrderListFilter{
		All:       true,
		Q:         strings.TrimSpace(q.Get("q")),
		Status:    strings.TrimSpace(q.Get("status")),
		BuyerType: strings.TrimSpace(q.Get("buyer_type")),
		Limit:     20,
		Offset:    0,
	}
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return repository.AppOrderListFilter{}, errInvalid("limit")
		}
		filter.Limit = n
	}
	if v := strings.TrimSpace(q.Get("offset")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return repository.AppOrderListFilter{}, errInvalid("offset")
		}
		filter.Offset = n
	}
	if v := strings.TrimSpace(q.Get("date_from")); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return repository.AppOrderListFilter{}, errInvalid("date_from")
		}
		filter.DateFrom = &t
	}
	if v := strings.TrimSpace(q.Get("date_to")); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return repository.AppOrderListFilter{}, errInvalid("date_to")
		}
		t = t.AddDate(0, 0, 1)
		filter.DateTo = &t
	}
	return filter, nil
}
