package router

import (
	"database/sql"
	"net/http"

	"pulsa2/db"
	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
	"pulsa2/javapay"
)

type JavapayInternalDeps struct {
	DB       *sql.DB
	JPClient *javapay.Client
}

func JavapayInternalRouter(mux *http.ServeMux, deps JavapayInternalDeps) {
	dbRepo := db.NewJavapayRepo(deps.DB)
	repo := repository.NewJavapayInternalRepository(dbRepo)
	svc := service.NewJavapayInternalService(repo, deps.JPClient)
	ctrl := controller.NewJavapayInternalController(svc)

	mux.HandleFunc("/internal/javapay/trx", helper.RequireInternalSecret(ctrl.HandleTrx))
	mux.HandleFunc("/internal/javapay/produk", helper.RequireInternalSecret(ctrl.HandleProduk))
}
