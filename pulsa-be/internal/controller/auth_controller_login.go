package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
)

func (h *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("bad json"))
		return
	}
	tok, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		code := http.StatusUnauthorized
		if strings.Contains(strings.ToLower(err.Error()), "required") {
			code = http.StatusBadRequest
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "token": tok})
}

func (h *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("missing bearer token"))
		return
	}
	tok, role, err := h.svc.Refresh(r.Context(), strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer ")))
	if err != nil {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "token": tok, "role": role})
}

func (h *AuthController) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		Email    string `json:"email"`
		Nama     string `json:"nama"`
		GoogleID string `json:"google_id"`
		IDToken  string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("bad json"))
		return
	}

	tok, created, role, err := h.svc.LoginGoogle(r.Context(), req.Email, req.Nama, req.GoogleID, req.IDToken)
	if err != nil {
		code := http.StatusUnauthorized
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "required") {
			code = http.StatusBadRequest
		} else if strings.Contains(lowerErr, "inactive") {
			code = http.StatusForbidden
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"token":   tok,
		"role":    role,
		"created": created,
	})
}

func (h *AuthController) LoginApple(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		Email         string `json:"email"`
		Nama          string `json:"nama"`
		AppleSub      string `json:"apple_sub"`
		IdentityToken string `json:"identity_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("bad json"))
		return
	}

	tok, created, role, err := h.svc.LoginApple(r.Context(), req.Email, req.Nama, req.AppleSub, req.IdentityToken)
	if err != nil {
		code := http.StatusUnauthorized
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "required") {
			code = http.StatusBadRequest
		} else if strings.Contains(lowerErr, "inactive") {
			code = http.StatusForbidden
		}
		helper.WriteJSON(w, code, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"token":   tok,
		"role":    role,
		"created": created,
	})
}

func (h *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("unauthorized"))
		return
	}
	me, err := h.svc.Me(r.Context(), a.MemberID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
		return
	}
	if me == nil {
		helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("member not found"))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"member_id": me.ID,
		"email":     me.Email,
		"nama":      me.Nama,
		"role":      me.Role,
		"aktif":     me.Aktif,
	})
}
