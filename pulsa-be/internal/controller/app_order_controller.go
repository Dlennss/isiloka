package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	apporderdto "pulsa2/internal/dto/app_order"
	apporderpaymentdto "pulsa2/internal/dto/app_order_payment"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AppOrderController struct {
	svc        *service.AppOrderService
	paymentSvc *service.AppOrderPaymentService
	base       string
	jwtSecret  []byte
}

func NewAppOrderController(svc *service.AppOrderService, paymentSvc *service.AppOrderPaymentService, base string, jwtSecret []byte) *AppOrderController {
	return &AppOrderController{svc: svc, paymentSvc: paymentSvc, base: strings.TrimRight(base, "/"), jwtSecret: jwtSecret}
}

func (h *AppOrderController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/orders"
	if r.URL.Path == base {
		h.handleCollection(w, r)
		return
	}

	invoiceID, action, ok := parseOrderPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("not found"))
		return
	}

	if action == "pay" {
		h.pay(w, r, invoiceID)
		return
	}

	h.handleItem(w, r, invoiceID)
}

func (h *AppOrderController) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.create(w, r)
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
	}
}

func (h *AppOrderController) handleItem(w http.ResponseWriter, r *http.Request, invoiceID string) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r, invoiceID)
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
	}
}

func (h *AppOrderController) create(w http.ResponseWriter, r *http.Request) {
	in, ok := h.decodeCreateRequest(w, r)
	if !ok {
		return
	}

	buyerType, memberID, status, roleOrErr := h.resolveBuyerFromHeader(r)
	if status != 0 {
		helper.WriteJSON(w, status, apporderdto.MapError(roleOrErr))
		return
	}
	in.BuyerType = buyerType
	in.MemberID = memberID
	in.BuyerRole = buyerType
	if buyerType == "user" && strings.TrimSpace(roleOrErr) != "" {
		in.BuyerRole = strings.TrimSpace(roleOrErr)
	}

	row, err := h.svc.Create(r.Context(), in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, apporderdto.MapItem(row))
}

func (h *AppOrderController) pay(w http.ResponseWriter, r *http.Request, invoiceID string) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderpaymentdto.MapError("method not allowed"))
		return
	}
	order, err := h.svc.GetByInvoiceID(r.Context(), invoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			helper.WriteJSON(w, http.StatusNotFound, apporderpaymentdto.MapError("order not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, apporderpaymentdto.MapError(err.Error()))
		return
	}
	if !h.canAccessOrder(w, r, order) {
		return
	}
	payment, err := h.paymentSvc.CreateByInvoiceID(r.Context(), invoiceID)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderpaymentdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, apporderpaymentdto.MapItem(payment))
}

func (h *AppOrderController) canAccessOrder(w http.ResponseWriter, r *http.Request, row *repository.AppOrderRow) bool {
	if row.BuyerType != "user" {
		guestEmail := normalizeGuestEmailHeader(r.Header.Get("X-Guest-Email"))
		guestPhone := normalizeGuestPhoneHeader(r.Header.Get("X-Guest-Phone"))
		if guestEmail == "" || guestPhone == "" {
			helper.WriteJSON(w, http.StatusUnauthorized, apporderdto.MapError("guest_email dan guest_phone wajib diisi"))
			return false
		}
		dbEmail := normalizeGuestEmailHeader(valueOrEmpty(row.GuestEmail))
		dbPhone := normalizeGuestPhoneHeader(valueOrEmpty(row.GuestPhone))
		if dbEmail == "" || dbPhone == "" || dbEmail != guestEmail || dbPhone != guestPhone {
			helper.WriteJSON(w, http.StatusForbidden, apporderdto.MapError("data guest tidak cocok"))
			return false
		}
		maskGuestIdentity(row)
		return true
	}
	buyerType, memberID, status, errMsg := h.resolveBuyerFromHeader(r)
	if status != 0 {
		helper.WriteJSON(w, status, apporderdto.MapError(errMsg))
		return false
	}
	if buyerType != "user" || memberID == nil || row.MemberID == nil || *memberID != *row.MemberID {
		helper.WriteJSON(w, http.StatusForbidden, apporderdto.MapError("forbidden"))
		return false
	}
	return true
}

func normalizeGuestEmailHeader(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizeGuestPhoneHeader(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func maskGuestIdentity(row *repository.AppOrderRow) {
	if row == nil || strings.TrimSpace(strings.ToLower(row.BuyerType)) != "guest" {
		return
	}
	row.GuestEmail = nil
	row.GuestPhone = nil
}

func (h *AppOrderController) get(w http.ResponseWriter, r *http.Request, invoiceID string) {
	row, err := h.svc.GetByInvoiceID(r.Context(), invoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("order not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}

	if !h.canAccessOrder(w, r, row) {
		return
	}

	_ = h.paymentSvc.RefreshStatusByOrderID(r.Context(), invoiceID)
	if refreshed, err := h.svc.GetByInvoiceID(r.Context(), invoiceID); err == nil && refreshed != nil {
		row = refreshed
		if strings.TrimSpace(strings.ToLower(row.BuyerType)) == "guest" {
			maskGuestIdentity(row)
		}
	}

	helper.WriteJSON(w, http.StatusOK, apporderdto.MapItem(row))
}

func (h *AppOrderController) decodeCreateRequest(w http.ResponseWriter, r *http.Request) (repository.AppOrderCreateInput, bool) {
	var req apporderdto.CreateRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError("invalid json"))
		return repository.AppOrderCreateInput{}, false
	}
	return apporderdto.MapCreateRequestToInput(req), true
}

func (h *AppOrderController) resolveBuyerFromHeader(r *http.Request) (string, *int64, int, string) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return "guest", nil, 0, ""
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", nil, http.StatusUnauthorized, "missing bearer token"
	}
	claims, err := helper.ParseJWT(h.jwtSecret, strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")))
	if err != nil {
		return "", nil, http.StatusUnauthorized, "invalid token"
	}
	role := helper.NormalizeRole(claims.Role)
	if !helper.IsRetailRole(role) {
		return "", nil, http.StatusForbidden, "retail user only"
	}
	memberID := claims.Sub
	return "user", &memberID, 0, role
}

func parseOrderPath(path, base string) (string, string, bool) {
	if !strings.HasPrefix(path, base+"/") {
		return "", "", false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(path, base+"/"))
	if rest == "" {
		return "", "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), "", strings.TrimSpace(parts[0]) != ""
	}
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) == "pay" {
		return strings.TrimSpace(parts[0]), "pay", true
	}
	return "", "", false
}
