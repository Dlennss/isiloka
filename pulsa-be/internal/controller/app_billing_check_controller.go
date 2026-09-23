package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	apporderdto "pulsa2/internal/dto/app_order"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AppBillingCheckController struct {
	svc       *service.AppBillingCheckService
	base      string
	jwtSecret []byte
}

func NewAppBillingCheckController(svc *service.AppBillingCheckService, base string, jwtSecret []byte) *AppBillingCheckController {
	return &AppBillingCheckController{svc: svc, base: strings.TrimRight(base, "/"), jwtSecret: jwtSecret}
}

func (h *AppBillingCheckController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/billing-checks"
	if r.URL.Path == base {
		if r.Method != http.MethodPost {
			helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
			return
		}
		h.create(w, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, base+"/") {
		helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("not found"))
		return
	}
	refID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, base+"/"))
	if refID == "" {
		helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("not found"))
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, apporderdto.MapError("method not allowed"))
		return
	}
	h.get(w, r, refID)
}

func (h *AppBillingCheckController) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProdukID   int64  `json:"produk_id"`
		Dest       string `json:"dest"`
		GuestNama  string `json:"guest_nama"`
		GuestEmail string `json:"guest_email"`
		GuestPhone string `json:"guest_phone"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError("invalid json"))
		return
	}
	buyerType, memberID, status, roleOrErr := h.resolveBuyerFromHeader(r)
	if status != 0 {
		helper.WriteJSON(w, status, apporderdto.MapError(roleOrErr))
		return
	}
	buyerRole := buyerType
	if buyerType == "user" && strings.TrimSpace(roleOrErr) != "" {
		buyerRole = strings.TrimSpace(roleOrErr)
	}
	row, err := h.svc.Create(r.Context(), repository.AppBillingCheckCreateInput{
		MemberID:   memberID,
		BuyerType:  buyerType,
		BuyerRole:  buyerRole,
		GuestNama:  req.GuestNama,
		GuestEmail: req.GuestEmail,
		GuestPhone: req.GuestPhone,
		ProdukID:   req.ProdukID,
		Dest:       req.Dest,
	})
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}
	if buyerType == "guest" {
		maskGuestBillingCheckIdentity(row)
	}
	helper.WriteJSON(w, http.StatusOK, apporderdto.MapItem(row))
}

func (h *AppBillingCheckController) get(w http.ResponseWriter, r *http.Request, refID string) {
	row, err := h.svc.GetByRefID(r.Context(), refID)
	if err != nil {
		if err == sql.ErrNoRows {
			helper.WriteJSON(w, http.StatusNotFound, apporderdto.MapError("billing check not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, apporderdto.MapError(err.Error()))
		return
	}
	if !h.canAccessBillingCheck(w, r, row) {
		return
	}
	if row.BuyerType == "guest" {
		maskGuestBillingCheckIdentity(row)
	}
	helper.WriteJSON(w, http.StatusOK, apporderdto.MapItem(row))
}

func (h *AppBillingCheckController) canAccessBillingCheck(w http.ResponseWriter, r *http.Request, row *repository.AppBillingCheckRow) bool {
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

func maskGuestBillingCheckIdentity(row *repository.AppBillingCheckRow) {
	if row == nil || strings.TrimSpace(strings.ToLower(row.BuyerType)) != "guest" {
		return
	}
	row.GuestEmail = nil
	row.GuestPhone = nil
}

func (h *AppBillingCheckController) resolveBuyerFromHeader(r *http.Request) (string, *int64, int, string) {
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
