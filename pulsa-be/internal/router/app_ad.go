package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AppAdRouter(mux *http.ServeMux, db *sql.DB) {
	repo := repository.NewAppAdRepository(db)
	svc := service.NewAppAdService(repo)
	ctrl := controller.NewAppAdController(svc, "/v1/app")

	mux.HandleFunc("/v1/app/ads", ctrl.Handle)
}
