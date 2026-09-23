package router

import (
	"database/sql"
	"net/http"

	"pulsa2/smb"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AuditRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB, smClient *smb.Client) {
	repo := repository.NewAuditRepository(db)
	svc := service.NewAuditService(repo, smClient)
	ctrl := controller.NewAuditController(svc)

	mux.HandleFunc("/v1/admin/audit/transaksi/status-mismatch", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminStatusMismatch)))
	mux.HandleFunc("/v1/admin/audit/transaksi/provider-empty-response", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminProviderEmptyResponse)))
	mux.HandleFunc("/v1/admin/audit/transaksi/transaksi-suspect", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminProviderSuccessSuspiciousMessage)))
	mux.HandleFunc("/v1/admin/audit/transaksi/provider-wallet-missing-debit", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminProviderWalletMissingDebit)))
	mux.HandleFunc("/v1/admin/audit/app-order/guest-refund-missing", wrap(helper.RequireRoles("admin")(ctrl.AdminGuestRefundMissing)))
}
