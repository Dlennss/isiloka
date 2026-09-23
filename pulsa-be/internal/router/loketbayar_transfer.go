package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
	"pulsa2/loketbayar"
)

func LoketBayarTransferRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB, lbClient *loketbayar.Client) {
	repo := repository.NewLoketBayarTransferRepository(db)
	svc := service.NewLoketBayarTransferService(repo, lbClient)
	ctrl := controller.NewLoketBayarTransferController(svc)
	roles := helper.RequireRoles("admin", "operator_wallet")

	mux.HandleFunc("/v1/admin/loketbayar-transfer/summary", wrap(roles(ctrl.Summary)))
	mux.HandleFunc("/v1/admin/loketbayar-transfer/transfers", wrap(roles(ctrl.List)))
	mux.HandleFunc("/v1/admin/loketbayar-transfer/transfers/create", wrap(roles(ctrl.Create)))
	mux.HandleFunc("/v1/admin/loketbayar-transfer/transfers/", wrap(roles(ctrl.Process)))
}
