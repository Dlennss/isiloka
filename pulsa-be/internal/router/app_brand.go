package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AppBrandRouter(mux *http.ServeMux, db *sql.DB) {
	repo := repository.NewAppBrandRepository(db)
	svc := service.NewAppBrandService(repo)
	ctrl := controller.NewAppBrandController(svc, "/v1/app")

	mux.HandleFunc("/v1/app/brand", ctrl.Handle)
}
