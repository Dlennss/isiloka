package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type MemberSelfController struct {
	svc *service.MemberSelfService
}

func NewMemberSelfController(svc *service.MemberSelfService) *MemberSelfController {
	return &MemberSelfController{svc: svc}
}

func (h *MemberSelfController) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPatch && r.Method != http.MethodPut {
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	if r.Method == http.MethodGet {
		out, err := h.svc.Profile(r.Context(), a.MemberID)
		if err != nil {
			helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, 200, out)
		return
	}

	var in struct {
		Nama            string `json:"nama"`
		Phone           string `json:"phone"`
		ProfilePhotoURL string `json:"profile_photo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	out, err := h.svc.UpdateProfile(r.Context(), a.MemberID, in.Nama, in.Phone, in.ProfilePhotoURL)
	if err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, 200, out)
}

func (h *MemberSelfController) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	out, err := h.svc.Stats(r.Context(), a.MemberID)
	if err != nil {
		helper.WriteJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, 200, out)
}

func (h *MemberSelfController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	if err := h.svc.ChangePassword(r.Context(), a.MemberID, in.OldPassword, in.NewPassword); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, 200, map[string]any{"ok": true})
}

func (h *MemberSelfController) ChangePIN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in struct {
		OldPin string `json:"old_pin"`
		NewPin string `json:"new_pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	if err := h.svc.ChangePIN(r.Context(), a.MemberID, in.OldPin, in.NewPin); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, 200, map[string]any{"ok": true})
}

func (h *MemberSelfController) ResetAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	apiKey, err := h.svc.ResetAPIKey(r.Context(), a.MemberID)
	if err != nil {
		helper.WriteJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, 200, map[string]any{"ok": true, "api_key": apiKey})
}

func (h *MemberSelfController) ChargeReceiver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}
	var in struct {
		ChargeReceiver bool `json:"charge_receiver"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	if err := h.svc.UpdateChargeReceiver(r.Context(), a.MemberID, a.Role, in.ChargeReceiver); err != nil {
		helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, 200, map[string]any{"ok": true, "charge_receiver": in.ChargeReceiver})
}

func (h *MemberSelfController) IPWhitelist(w http.ResponseWriter, r *http.Request) {
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, 401, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := h.svc.ListIPWhitelist(r.Context(), a.MemberID)
		if err != nil {
			helper.WriteJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, 200, map[string]any{"ok": true, "rows": rows})
		return
	case http.MethodPost:
		var in struct {
			IP         string `json:"ip"`
			Label      string `json:"label"`
			WebhookURL string `json:"webhook_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": "invalid json"})
			return
		}
		if err := h.svc.AddIPWhitelist(r.Context(), a.MemberID, strings.TrimSpace(in.IP), strings.TrimSpace(in.Label), strings.TrimSpace(in.WebhookURL)); err != nil {
			helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, 200, map[string]any{"ok": true})
		return
	case http.MethodDelete:
		idStr := strings.TrimSpace(r.URL.Query().Get("id"))
		if idStr == "" {
			helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": "id required"})
			return
		}
		if err := h.svc.DeleteIPWhitelist(r.Context(), a.MemberID, idStr); err != nil {
			helper.WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		helper.WriteJSON(w, 200, map[string]any{"ok": true})
		return
	default:
		helper.WriteJSON(w, 405, map[string]any{"ok": false, "error": "method not allowed"})
	}
}
