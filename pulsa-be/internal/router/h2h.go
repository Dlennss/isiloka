package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func H2HRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewH2HRepository(db)
	authRepo := repository.NewAuthRepository(db)
	bankRepo := repository.NewBankRepository(db)
	svc := service.NewH2HService(repo, authRepo, bankRepo)
	ctrl := controller.NewH2HController(svc)

	mux.HandleFunc("/v1/me/h2h/downlines", wrap(ctrl.Downlines))
	mux.HandleFunc("/v1/me/h2h/commissions", wrap(ctrl.Commissions))
	mux.HandleFunc("/v1/me/h2h/commissions/summary", wrap(ctrl.CommissionSummary))
	mux.HandleFunc("/v1/me/h2h/withdraw-requests", wrap(ctrl.WithdrawRequests))
}

func H2HAdminRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewH2HRepository(db)
	authRepo := repository.NewAuthRepository(db)
	bankRepo := repository.NewBankRepository(db)
	svc := service.NewH2HService(repo, authRepo, bankRepo)
	ctrl := controller.NewH2HAdminController(svc)

	adminOrWallet := helper.RequireRoles("admin", "operator_wallet")
	mux.HandleFunc("/v1/admin/h2h/withdraw-requests", wrap(adminOrWallet(ctrl.ListWithdrawRequests)))
	mux.HandleFunc("/v1/admin/h2h/withdraw-requests/approve", wrap(adminOrWallet(ctrl.ApproveWithdrawRequest)))
	mux.HandleFunc("/v1/admin/h2h/withdraw-requests/reject", wrap(adminOrWallet(ctrl.RejectWithdrawRequest)))
}
