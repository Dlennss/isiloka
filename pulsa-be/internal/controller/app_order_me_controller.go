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

type AppOrderMeController struct {
	svc       *service.AppOrderService
	jwtSecret []byte
}

func NewAppOrderMeController(svc *service.AppOrderService, jwtSecret []byte) *AppOrderMeController {
	return &AppOrderMeController{svc: svc, jwtSecret: jwtSecret}
}

func (h *AppOrderMeController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
		return
	}

	memberID, status, errMsg := h.resolveUser(r)
	if status != 0 {
		helper.WriteJSON(w, status, apporderdto.MapError(errMsg))
		return
	}

	filter, err := h.buildFilter(r, memberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}

	rows, err := h.svc.ListByMemberID(r.Context(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, apporderdto.MapList(rows))
}

func (h *AppOrderMeController) resolveUser(r *http.Request) (int64, int, string) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return 0, http.StatusUnauthorized, "missing bearer token"
	}
	claims, err := helper.ParseJWT(h.jwtSecret, strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")))
	if err != nil {
		return 0, http.StatusUnauthorized, "invalid token"
	}
	if !helper.IsRetailRole(claims.Role) {
		return 0, http.StatusForbidden, "retail user only"
	}
	if claims.Sub <= 0 {
		return 0, http.StatusUnauthorized, "user tidak valid"
	}
	return claims.Sub, 0, ""
}

func (h *AppOrderMeController) buildFilter(r *http.Request, memberID int64) (repository.AppOrderListFilter, error) {
	q := r.URL.Query()
	filter := repository.AppOrderListFilter{
		MemberID: memberID,
		Status:   strings.TrimSpace(q.Get("status")),
		Limit:    20,
		Offset:   0,
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

func errInvalid(field string) error {
	return &badFieldError{field: field}
}

type badFieldError struct {
	field string
}

func (e *badFieldError) Error() string {
	return e.field + " tidak valid"
}
