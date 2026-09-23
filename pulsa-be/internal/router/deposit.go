package router

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
	"pulsa2/loketbayar"
)

func DepositRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB, lbClient *loketbayar.Client, p24Client *provider.Pulsa24JamAdapter) {
	repo := repository.NewDepositRepository(db)
	if err := repo.EnsureTicketSchema(context.Background()); err != nil {
		log.Printf("[deposit_ticket_setup] gagal menyiapkan skema tiket deposit: %v", err)
	}
	bankRepo := repository.NewBankRepository(db)
	if err := bankRepo.EnsurePulsaKilatDepositBank(context.Background()); err != nil {
		log.Printf("[deposit_bank_setup] gagal mengaktifkan rekening deposit BNI: %v", err)
	}
	svc := service.NewDepositService(repo, bankRepo, lbClient)
	svc.SetPulsa24JamClient(p24Client)
	ctrl := controller.NewDepositController(svc)

	// member
	mux.HandleFunc("/v1/deposit/banks", wrap(ctrl.MemberBanks))
	mux.HandleFunc("/v1/deposit/request", wrap(ctrl.MemberCreate))
	mux.HandleFunc("/v1/deposit/request/confirm-transfer", wrap(ctrl.MemberConfirmTransfer))
	mux.HandleFunc("/v1/deposit/request/cancel-ticket", wrap(ctrl.MemberCancelTicket))
	mux.HandleFunc("/v1/deposit/request/qris", wrap(ctrl.MemberCreateQris))
	mux.HandleFunc("/v1/deposit/request/qris/status", wrap(ctrl.MemberQrisStatus))
	mux.HandleFunc("/v1/deposit/request/va", wrap(ctrl.MemberCreateVA))
	mux.HandleFunc("/v1/history/deposit", wrap(ctrl.MemberList))

	// admin / operator read-only list
	mux.HandleFunc("/v1/admin/deposit/requests", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminList)))
	mux.HandleFunc("/v1/admin/deposit/requests/va", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminListVA)))
	// admin / operator wallet write actions
	mux.HandleFunc("/v1/admin/deposit/requests/approve", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminApprove)))
	mux.HandleFunc("/v1/admin/deposit/requests/reject", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminReject)))
	mux.HandleFunc("/v1/admin/deposit/requests/va/approve", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminApproveVA)))
	mux.HandleFunc("/v1/admin/deposit/requests/va/reject", wrap(helper.RequireRoles("admin", "operator_wallet")(ctrl.AdminRejectVA)))

	// internal admin token
	mux.HandleFunc("/admin/deposit/credit", ctrl.AdminCreditInternal)

	if strings.EqualFold(strings.TrimSpace(os.Getenv("DEPOSIT_AUTO_APPROVE_ENABLED")), "true") {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				n, err := svc.AutoApprovePendingFromBankMutations(ctx, 5)
				cancel()
				if err != nil {
					log.Printf("[deposit_auto_approve] error: %v", err)
					continue
				}
				if n > 0 {
					log.Printf("[deposit_auto_approve] %d deposit approved", n)
				}
			}
		}()
	}

	if p24Client != nil && p24Client.Configured() {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				n, err := svc.ReconcilePulsa24JamQris(ctx, 25)
				cancel()
				if err != nil {
					log.Printf("[pulsa24jam_qris_reconcile] error: %v", err)
					continue
				}
				if n > 0 {
					log.Printf("[pulsa24jam_qris_reconcile] %d deposit diperiksa", n)
				}
			}
		}()
	}
}
