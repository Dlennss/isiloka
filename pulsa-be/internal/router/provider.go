package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func ProviderRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	repo := repository.NewProviderRepository(db)
	svc := service.NewProviderService(repo)
	ctrl := controller.NewProviderController(svc, "/v1/admin")

	mux.HandleFunc("/v1/admin/provider", wrap(requireAdmin(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/provider/", wrap(requireAdmin(ctrl.Handle)))
}
