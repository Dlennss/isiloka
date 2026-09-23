package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AppProdukRouter(mux *http.ServeMux, db *sql.DB) {
	repo := repository.NewAppProdukRepository(db)
	svc := service.NewAppProdukService(repo)
	ctrl := controller.NewAppProdukController(svc, "/v1/app")

	mux.HandleFunc("/v1/app/produk", ctrl.Handle)
}
