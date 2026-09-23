package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func ProdukRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	repo := repository.NewProdukRepository(db)
	svc := service.NewProdukService(repo)
	ctrl := controller.NewProdukController(svc, "/v1/admin")

	mux.HandleFunc("/v1/admin/produk", wrap(requireAdmin(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/produk/", wrap(requireAdmin(ctrl.Handle)))
}
