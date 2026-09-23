package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AuditorRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewAuditorRepository(db)
	bankRepo := repository.NewBankRepository(db)
	svc := service.NewAuditorService(repo, bankRepo)
	ctrl := controller.NewAuditorController(svc)

	readRoles := helper.RequireRoles("admin", "auditor", "operator_wallet")
	writeRoles := helper.RequireRoles("admin", "operator_wallet")

	mux.HandleFunc("/v1/auditor/trading/summary", wrap(readRoles(ctrl.TradingSummary)))
	mux.HandleFunc("/v1/auditor/trading/details", wrap(readRoles(ctrl.TradingDetails)))
	mux.HandleFunc("/v1/auditor/finance", wrap(readRoles(ctrl.Finance)))
	mux.HandleFunc("/v1/auditor/profit-loss", wrap(readRoles(ctrl.ProfitLoss)))
	mux.HandleFunc("/v1/auditor/opening-balances", wrap(readRoles(ctrl.OpeningBalances)))
	mux.HandleFunc("/v1/admin/wallets/internal-finance", wrap(writeRoles(ctrl.InternalFinance)))
	mux.HandleFunc("/v1/auditor/internal-finance", wrap(readRoles(ctrl.InternalFinance)))
}
