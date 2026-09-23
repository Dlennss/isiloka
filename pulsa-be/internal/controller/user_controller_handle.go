package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func (h *UserController) Handle(w http.ResponseWriter, r *http.Request) {
	a, _ := helper.GetAuth(r.Context())
	isAdmin := a.Role == "admin"
	isRetailMaster := helper.NormalizeRole(a.Role) == helper.RoleRetailMaster
	base := h.base + "/users"
	if r.URL.Path == base {
		switch r.Method {
		case http.MethodGet:
			q := helper.QueryString(r, "q")
			role := helper.QueryString(r, "role")
			scope := helper.QueryString(r, "scope")
			rows, err := h.svc.List(
				r.Context(),
				q,
				role,
				scope,
				helper.QueryInt(r, "limit", 20),
				helper.QueryInt(r, "offset", 0),
			)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
				return
			}
			totalCount, err := h.svc.Count(r.Context(), q, role, scope)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
				return
			}
			totalSaldo, err := h.svc.SumSaldo(r.Context(), q, role, scope)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "rows": rows, "total_count": totalCount, "total_saldo": totalSaldo})
		case http.MethodPost:
			if !isAdmin {
				helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
				return
			}
			var req struct {
				Email    string `json:"email"`
				Nama     string `json:"nama"`
				Password string `json:"password"`
				PIN      string `json:"pin"`
				Role     string `json:"role"`
				Aktif    *bool  `json:"aktif"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
				return
			}
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			id, err := h.svc.Create(r.Context(), req.Email, req.Nama, req.Password, req.PIN, req.Role, aktif)
			if err != nil {
				helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
				return
			}
			helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
		default:
			helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		}
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		if row == nil {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("user not found"))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
	case http.MethodPut:
		if !isAdmin && !isRetailMaster {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}
		var req struct {
			Email                    string `json:"email"`
			Nama                     string `json:"nama"`
			Phone                    string `json:"phone"`
			Role                     string `json:"role"`
			Aktif                    *bool  `json:"aktif"`
			FeeMemberRp              int64  `json:"fee_member_rp"`
			RetailAgentCommissionRp  int64  `json:"retail_agent_commission_rp"`
			RetailMasterCommissionRp int64  `json:"retail_master_commission_rp"`
			H2HAgentCommissionRp     int64  `json:"h2h_agent_commission_rp"`
			H2HMasterCommissionRp    int64  `json:"h2h_master_commission_rp"`
			NewPassword              string `json:"new_password"`
			NewPIN                   string `json:"new_pin"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		if isRetailMaster {
			row, err := h.svc.Get(r.Context(), id)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
				return
			}
			if row == nil {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("user not found"))
				return
			}
			if helper.NormalizeRole(row.Role) != helper.RoleRetailAgent || helper.NormalizeRole(req.Role) != helper.RoleRetailAgent {
				helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("master hanya bisa mengelola akun agent"))
				return
			}
		}
		aktif := true
		if req.Aktif != nil {
			aktif = *req.Aktif
		}
		err := h.svc.Update(r.Context(), repository.UserUpdateInput{
			ID:                       id,
			Email:                    req.Email,
			Nama:                     req.Nama,
			Phone:                    req.Phone,
			Role:                     req.Role,
			Aktif:                    aktif,
			FeeMemberRp:              req.FeeMemberRp,
			RetailAgentCommissionRp:  req.RetailAgentCommissionRp,
			RetailMasterCommissionRp: req.RetailMasterCommissionRp,
			H2HAgentCommissionRp:     req.H2HAgentCommissionRp,
			H2HMasterCommissionRp:    req.H2HMasterCommissionRp,
		}, req.NewPassword, req.NewPIN)
		if err != nil {
			code := http.StatusBadRequest
			if service.IsNotFound(err) {
				code = http.StatusNotFound
			}
			helper.WriteJSON(w, code, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	case http.MethodDelete:
		if !isAdmin && !isRetailMaster {
			helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
			return
		}
		if isRetailMaster {
			row, err := h.svc.Get(r.Context(), id)
			if err != nil {
				helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
				return
			}
			if row == nil {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("user not found"))
				return
			}
			if helper.NormalizeRole(row.Role) != helper.RoleRetailAgent {
				helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("master hanya bisa mengelola akun agent"))
				return
			}
		}
		err := h.svc.Delete(r.Context(), id)
		if err != nil {
			code := http.StatusBadRequest
			if service.IsNotFound(err) {
				code = http.StatusNotFound
			}
			helper.WriteJSON(w, code, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *UserController) CreateMarketing(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		accounts, err := h.svc.ListMarketingManagement(r.Context())
		if err != nil {
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "accounts": accounts})
		return
	}
	if r.Method == http.MethodPatch {
		var req struct {
			Action      string `json:"action"`
			ID          int64  `json:"id"`
			Nama        string `json:"nama"`
			Phone       string `json:"phone"`
			NewPassword string `json:"new_password"`
			Aktif       *bool  `json:"aktif"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
			return
		}
		var err error
		switch req.Action {
		case "update_account":
			aktif := true
			if req.Aktif != nil {
				aktif = *req.Aktif
			}
			err = h.svc.UpdateMarketingAccount(r.Context(), req.ID, req.Nama, req.Phone, req.NewPassword, aktif)
		default:
			err = errors.New("aksi tidak dikenali")
		}
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
		return
	}
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	var req struct {
		Email    string `json:"email"`
		Nama     string `json:"nama"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	id, err := h.svc.Create(r.Context(), req.Email, req.Nama, req.Password, "", helper.RoleRetailMarketing, true)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	if req.Phone != "" {
		if err := h.svc.Update(r.Context(), repository.UserUpdateInput{
			ID:    id,
			Email: req.Email,
			Nama:  req.Nama,
			Phone: req.Phone,
			Role:  helper.RoleRetailMarketing,
			Aktif: true,
		}, "", ""); err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "member_id": id, "phone": req.Phone})
}
