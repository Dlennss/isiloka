package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	apporderrefunddto "pulsa2/internal/dto/app_order_refund"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppOrderRefundMeController struct {
	svc       *service.AppOrderRefundService
	jwtSecret []byte
}

func NewAppOrderRefundMeController(svc *service.AppOrderRefundService, jwtSecret []byte) *AppOrderRefundMeController {
	return &AppOrderRefundMeController{svc: svc, jwtSecret: jwtSecret}
}

func (h *AppOrderRefundMeController) HandleClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderrefunddto.MapError("method not allowed"))
		return
	}

	memberID, status, errMsg := h.resolveUser(r)
	if status != 0 {
		helper.WriteJSON(w, status, apporderrefunddto.MapError(errMsg))
		return
	}

	var req apporderrefunddto.ClaimGuestRefundRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderrefunddto.MapError("invalid json"))
		return
	}

	row, err := h.svc.ClaimGuestRefund(r.Context(), memberID, req.InvoiceID, req.GuestEmail, req.GuestPhone)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderrefunddto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, apporderrefunddto.MapItem(row))
}

func (h *AppOrderRefundMeController) resolveUser(r *http.Request) (int64, int, string) {
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
