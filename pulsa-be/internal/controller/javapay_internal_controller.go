package controller

import (
	"encoding/json"
	"net/http"

	"pulsa2/internal/helper"
	"pulsa2/internal/service"
	"pulsa2/model"
)

type JavapayInternalController struct {
	svc *service.JavapayInternalService
}

func NewJavapayInternalController(svc *service.JavapayInternalService) *JavapayInternalController {
	return &JavapayInternalController{svc: svc}
}

func (h *JavapayInternalController) HandleTrx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in model.JavapayTrxCreateIn
	if err := decodeJSON(r, &in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}

	status, out := h.svc.HandleTrx(r.Context(), in)
	helper.WriteJSON(w, status, out)
}

func (h *JavapayInternalController) HandleProduk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status, out := h.svc.ProdukStatus(r.Context())
	helper.WriteJSON(w, status, out)
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}
