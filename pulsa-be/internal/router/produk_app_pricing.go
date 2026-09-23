package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func ProdukAppPricingRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	repo := repository.NewProdukAppPricingRepository(db)
	svc := service.NewProdukAppPricingService(repo)
	ctrl := controller.NewProdukAppPricingController(svc, "/v1/admin")

	mux.HandleFunc("/v1/admin/produk-app-pricing", wrap(requireAdmin(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/produk-app-pricing/", wrap(requireAdmin(ctrl.Handle)))
}
