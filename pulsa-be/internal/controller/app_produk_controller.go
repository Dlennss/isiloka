package controller

import (
	"net/http"
	"strings"

	appprodukdto "pulsa2/internal/dto/app_produk"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type AppProdukController struct {
	svc  *service.AppProdukService
	base string
}

func NewAppProdukController(svc *service.AppProdukService, base string) *AppProdukController {
	return &AppProdukController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *AppProdukController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/produk"
	if r.URL.Path != base {
		helper.WriteJSON(w, http.StatusNotFound, appprodukdto.MapError("not found"))
		return
	}
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, appprodukdto.MapError("method not allowed"))
		return
	}

	q := helper.QueryString(r, "q")
	kategoriID := helper.QueryInt64(r, "kategori_id", 0)
	brandID := helper.QueryInt64(r, "brand_id", 0)
	rows, err := h.svc.List(r.Context(), q, kategoriID, brandID)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, appprodukdto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, appprodukdto.MapList(rows))
}
