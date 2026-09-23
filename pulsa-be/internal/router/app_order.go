package router

import (
	"database/sql"
	"net/http"

	"pulsa2/gemilang"
	"pulsa2/internal/controller"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
	"pulsa2/yuscom"
)

func AppOrderRouter(mux *http.ServeMux, db *sql.DB, jwtSecret []byte, ysClient *yuscom.Client, gmClient *gemilang.Client, extraClients ...provider.Client) {
	var p24Client provider.Client
	var p24Adapter *provider.Pulsa24JamAdapter
	for _, client := range extraClients {
		if client != nil && client.Name() == provider.Pulsa24JamProviderName {
			p24Client = client
			p24Adapter, _ = client.(*provider.Pulsa24JamAdapter)
			break
		}
	}
	orderRepo := repository.NewAppOrderRepository(db)
	produkRepo := repository.NewProdukRepository(db)
	pricingRepo := repository.NewProdukAppPricingRepository(db)
	feeRepo := repository.NewKategoriFeeAppRepository(db)
	paymentRepo := repository.NewAppOrderPaymentRepository(db)
	bankRepo := repository.NewBankRepository(db)
	appProviderRepo := repository.NewAppOrderProviderTrxRepository(db)
	billingCheckRepo := repository.NewAppBillingCheckRepository(db)
	svc := service.NewAppOrderService(orderRepo, paymentRepo, produkRepo, pricingRepo, feeRepo, appProviderRepo, billingCheckRepo)
	svc.SetPulsa24JamClient(p24Adapter)
	billingCheckSvc := service.NewAppBillingCheckService(billingCheckRepo, produkRepo, pricingRepo, p24Client)
	callbackRepo := repository.NewProviderCallbackRepository(db)
	fulfillmentSvc := service.NewAppOrderFulfillmentService(orderRepo, appProviderRepo, callbackRepo, pricingRepo, ysClient, gmClient, extraClients...)
	paymentSvc := service.NewAppOrderPaymentService(paymentRepo, orderRepo, bankRepo, fulfillmentSvc)
	ctrl := controller.NewAppOrderController(svc, paymentSvc, "/v1/app", jwtSecret)
	billingCtrl := controller.NewAppBillingCheckController(billingCheckSvc, "/v1/app", jwtSecret)

	mux.HandleFunc("/v1/app/orders", ctrl.Handle)
	mux.HandleFunc("/v1/app/orders/", ctrl.Handle)
	mux.HandleFunc("/v1/app/billing-checks", billingCtrl.Handle)
	mux.HandleFunc("/v1/app/billing-checks/", billingCtrl.Handle)
}
