package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func AppOrderMeRouter(mux *http.ServeMux, db *sql.DB, jwtSecret []byte) {
	orderRepo := repository.NewAppOrderRepository(db)
	produkRepo := repository.NewProdukRepository(db)
	pricingRepo := repository.NewProdukAppPricingRepository(db)
	feeRepo := repository.NewKategoriFeeAppRepository(db)
	paymentRepo := repository.NewAppOrderPaymentRepository(db)
	appProviderRepo := repository.NewAppOrderProviderTrxRepository(db)
	billingCheckRepo := repository.NewAppBillingCheckRepository(db)
	svc := service.NewAppOrderService(orderRepo, paymentRepo, produkRepo, pricingRepo, feeRepo, appProviderRepo, billingCheckRepo)
	ctrl := controller.NewAppOrderMeController(svc, jwtSecret)

	mux.HandleFunc("/v1/app/me/orders", ctrl.Handle)
}
