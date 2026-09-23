package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func ProviderMerchantIDRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewProviderMerchantIDRepository(db)
	svc := service.NewProviderMerchantIDService(repo)
	ctrl := controller.NewProviderMerchantIDController(svc)
	roles := helper.RequireRoles("admin")

	mux.HandleFunc("/v1/admin/provider-merchant-ids", wrap(roles(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/provider-merchant-ids/", wrap(roles(ctrl.Handle)))
}
