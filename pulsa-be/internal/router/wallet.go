package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func WalletRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewWalletRepository(db)
	bankRepo := repository.NewBankRepository(db)
	svc := service.NewWalletService(repo, bankRepo)
	ctrl := controller.NewWalletController(svc)

	// member wallet (admin)
	mux.HandleFunc("/v1/admin/wallets/members/adjust", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminAdjustMemberWallet)))
	mux.HandleFunc("/v1/admin/wallets/activity", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminWalletActivity)))
	mux.HandleFunc("/v1/admin/wallets/members/deposit-history", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminMemberDepositHistory)))

	// provider wallet (admin)
	mux.HandleFunc("/v1/admin/wallets/providers", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminListProviderWallets)))
	mux.HandleFunc("/v1/admin/wallets/providers/deposit", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminDepositProviderWallet)))
	mux.HandleFunc("/v1/admin/wallets/providers/adjust", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminAdjustProviderWallet)))
	mux.HandleFunc("/v1/admin/wallets/providers/history", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminProviderWalletHistory)))
	mux.HandleFunc("/v1/admin/wallets/providers/deposit-history", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminProviderDepositHistory)))
}
