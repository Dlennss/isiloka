package controller

import (
	"net/http"
	"strings"

	kategoridto "pulsa2/internal/dto/kategori"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppKategoriController struct {
	svc  *service.AppKategoriService
	base string
}

func NewAppKategoriController(svc *service.AppKategoriService, base string) *AppKategoriController {
	return &AppKategoriController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *AppKategoriController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/kategori"
	if r.URL.Path != base {
		helper.WriteJSON(w, http.StatusNotFound, kategoridto.MapError("not found"))
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, kategoridto.MapError("method not allowed"))
		return
	}

	rows, err := h.svc.List(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, kategoridto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, kategoridto.MapList(rows))
}
