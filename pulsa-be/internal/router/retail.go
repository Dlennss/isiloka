package router

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func RetailRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB, clients ...provider.Client) {
	repo := repository.NewRetailRepository(db)
	if err := repo.EnsureWithdrawSchema(context.Background()); err != nil {
		log.Printf("[retail_withdraw_schema] gagal menyiapkan skema penarikan: %v", err)
	}
	bankRepo := repository.NewBankRepository(db)
	svc := service.NewRetailService(repo, bankRepo, clients...)
	ctrl := controller.NewRetailController(svc)

	mux.HandleFunc("/v1/me/retail/downlines", wrap(ctrl.Downlines))
	mux.HandleFunc("/v1/me/retail/commissions", wrap(ctrl.Commissions))
	mux.HandleFunc("/v1/me/retail/commissions/summary", wrap(ctrl.CommissionSummary))
	mux.HandleFunc("/v1/me/retail/withdraw-requests", wrap(ctrl.WithdrawRequests))
}
