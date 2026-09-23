package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AdminBusinessReportRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewAdminBusinessReportRepository(db)
	svc := service.NewAdminBusinessReportService(repo)
	ctrl := controller.NewAdminBusinessReportController(svc)

	adminOnly := helper.RequireRoles("admin")

	mux.HandleFunc("/v1/admin/reports/commissions/by-source", wrap(adminOnly(ctrl.CommissionBySource)))
	mux.HandleFunc("/v1/admin/reports/daily-business", wrap(adminOnly(ctrl.DailyBusiness)))
	mux.HandleFunc("/v1/admin/reports/daily-business/refresh-cache", wrap(adminOnly(ctrl.RefreshDailyBusinessCache)))

	mux.HandleFunc("/v1/admin/commissions/by-source", wrap(adminOnly(ctrl.CommissionBySource)))
	mux.HandleFunc("/v1/admin/daily-business", wrap(adminOnly(ctrl.DailyBusiness)))
	mux.HandleFunc("/v1/admin/daily-business/refresh-cache", wrap(adminOnly(ctrl.RefreshDailyBusinessCache)))
}
