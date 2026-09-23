package controller

import (
	"net/http"
	"strings"

	branddto "pulsa2/internal/dto/brand"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppBrandController struct {
	svc  *service.AppBrandService
	base string
}

func NewAppBrandController(svc *service.AppBrandService, base string) *AppBrandController {
	return &AppBrandController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *AppBrandController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/brand"
	if r.URL.Path != base {
		helper.WriteJSON(w, http.StatusNotFound, branddto.MapError("not found"))
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, branddto.MapError("method not allowed"))
		return
	}

	kategoriID := helper.QueryInt64(r, "kategori_id", 0)
	rows, err := h.svc.List(r.Context(), kategoriID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, branddto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, branddto.MapList(rows))
}
