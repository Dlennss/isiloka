package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func ProviderReportingRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewProviderReportingRepository(db)
	svc := service.NewProviderReportingService(repo)
	ctrl := controller.NewProviderReportingController(svc)

	// New tidy resource paths
	mux.HandleFunc("/v1/admin/providers/transactions", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.ListTransactions)))
	mux.HandleFunc("/v1/admin/providers/anomalies", wrap(helper.RequireRoles("admin")(ctrl.ListAnomalies)))
	mux.HandleFunc("/v1/admin/providers/analytics", wrap(helper.RequireRoles("admin")(ctrl.Analytics)))
	mux.HandleFunc("/v1/admin/providers/analytics/refresh-cache", wrap(helper.RequireRoles("admin")(ctrl.RefreshAnalyticsCache)))
	mux.HandleFunc("/v1/admin/providers/products/daily-success", wrap(helper.RequireRoles("admin")(ctrl.DailyProductSuccess)))

	// Backward compatibility (legacy paths)
	mux.HandleFunc("/v1/admin/provider/trx", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.ListTransactions)))
	mux.HandleFunc("/v1/admin/provider/anomalies", wrap(helper.RequireRoles("admin")(ctrl.ListAnomalies)))
	mux.HandleFunc("/v1/admin/provider/analytics", wrap(helper.RequireRoles("admin")(ctrl.Analytics)))
	mux.HandleFunc("/v1/admin/provider/analytics/refresh-cache", wrap(helper.RequireRoles("admin")(ctrl.RefreshAnalyticsCache)))
	mux.HandleFunc("/v1/admin/provider/products/daily-success", wrap(helper.RequireRoles("admin")(ctrl.DailyProductSuccess)))
}
