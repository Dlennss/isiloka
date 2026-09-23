package router

import (
	"database/sql"
	"net/http"
	"strings"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func ProdukProviderMapRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	_ = requireAdmin

	repo := repository.NewProdukProviderMapRepository(db)
	svc := service.NewProdukProviderMapService(repo)
	ctrl := controller.NewProdukProviderMapController(svc, "/v1/admin")

	readOrToggle := helper.RequireRoles("admin", "operator_trx")
	adminOnly := helper.RequireRoles("admin")

	mux.HandleFunc("/v1/admin/produk-provider-map", wrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			readOrToggle(ctrl.Handle)(w, r)
			return
		}
		adminOnly(ctrl.Handle)(w, r)
	}))

	mux.HandleFunc("/v1/admin/produk-provider-map/", wrap(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(strings.TrimRight(r.URL.Path, "/"), "/toggle-active") {
			readOrToggle(ctrl.HandleToggleActive)(w, r)
			return
		}
		if r.Method == http.MethodGet {
			readOrToggle(ctrl.Handle)(w, r)
			return
		}
		adminOnly(ctrl.Handle)(w, r)
	}))
}
