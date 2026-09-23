package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func BrandRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	repo := repository.NewBrandRepository(db)
	svc := service.NewBrandService(repo)
	ctrl := controller.NewBrandController(svc, "/v1/admin")

	mux.HandleFunc("/v1/admin/brand", wrap(requireAdmin(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/brand/", wrap(requireAdmin(ctrl.Handle)))
}
