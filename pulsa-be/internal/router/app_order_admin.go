package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AppOrderAdminRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	orderRepo := repository.NewAppOrderRepository(db)
	produkRepo := repository.NewProdukRepository(db)
	pricingRepo := repository.NewProdukAppPricingRepository(db)
	feeRepo := repository.NewKategoriFeeAppRepository(db)
	paymentRepo := repository.NewAppOrderPaymentRepository(db)
	appProviderRepo := repository.NewAppOrderProviderTrxRepository(db)
	billingCheckRepo := repository.NewAppBillingCheckRepository(db)
	svc := service.NewAppOrderService(orderRepo, paymentRepo, produkRepo, pricingRepo, feeRepo, appProviderRepo, billingCheckRepo)
	adminSvc := service.NewAppOrderAdminService(orderRepo, paymentRepo, appProviderRepo)
	ctrl := controller.NewAppOrderAdminController(svc, adminSvc, "/v1/admin/app")
	providerCtrl := controller.NewAppOrderProviderAdminController(adminSvc)
	readOnlyAdminApp := helper.RequireRoles("admin", "operator_trx", "operator_wallet")

	mux.HandleFunc("/v1/admin/app/orders", wrap(readOnlyAdminApp(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/app/orders/", wrap(readOnlyAdminApp(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/app/provider-trx", wrap(readOnlyAdminApp(providerCtrl.Handle)))
}
