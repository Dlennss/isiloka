package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AppAdAdminRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	repo := repository.NewAppAdRepository(db)
	svc := service.NewAppAdService(repo)
	ctrl := controller.NewAppAdAdminController(svc, "/v1/admin")

	mux.HandleFunc("/v1/admin/app-ads", wrap(requireAdmin(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/app-ads/", wrap(requireAdmin(ctrl.Handle)))
}
