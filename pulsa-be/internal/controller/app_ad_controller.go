package controller

import (
	"net/http"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppAdController struct {
	svc  *service.AppAdService
	base string
}

func NewAppAdController(svc *service.AppAdService, base string) *AppAdController {
	return &AppAdController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *AppAdController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/ads"
	if r.URL.Path != base {
		helper.WriteJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "not found"})
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}

	rows, err := h.svc.ListActive(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
}
