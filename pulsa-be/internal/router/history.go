package router

import (
	"database/sql"
	"net/http"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/gemilang"
	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
	"pulsa2/javapay"
	"pulsa2/loketbayar"
	"pulsa2/minions"
	"pulsa2/multikom"
	"pulsa2/rajabiller"
	"pulsa2/sagaramobile"
	"pulsa2/smb"
	"pulsa2/talenta"
	"pulsa2/trionik"
	"pulsa2/yuscom"
)

type HistoryDeps struct {
	DB       *sql.DB
	YSClient *yuscom.Client
	JPClient *javapay.Client
	TLClient *talenta.Client
	MKClient *multikom.Client
	SGClient *sagaramobile.Client
	MNClient *minions.Client
	TRClient *trionik.Client
	AJClient *ajs.Client
	GMClient *gemilang.Client
	SMClient *smb.Client
	LBClient *loketbayar.Client
	CHClient *chytron.Client
	RJClient *rajabiller.Client
}

func HistoryRouter(mux *http.ServeMux, wrap Middleware, deps HistoryDeps) {
	repo := repository.NewHistoryRepository(deps.DB)
	memberRepo := repository.NewMemberTrxRepository(deps.DB)
	trxSvc := service.NewMemberTrxService(memberRepo, deps.YSClient, deps.JPClient, deps.TLClient, deps.MKClient, deps.SGClient, deps.MNClient, deps.TRClient, deps.AJClient, deps.GMClient, deps.SMClient, deps.LBClient, deps.CHClient, deps.RJClient)
	svc := service.NewHistoryService(repo, trxSvc)
	ctrl := controller.NewHistoryController(svc)

	mux.HandleFunc("/v1/saldo", wrap(ctrl.MemberSaldo))
	mux.HandleFunc("/v1/history/mutasi", wrap(ctrl.MemberMutasi))
	mux.HandleFunc("/v1/history/mutasi/", wrap(ctrl.MemberMutasiDetail))
	mux.HandleFunc("/v1/history/transaksi", wrap(ctrl.MemberTransaksi))

	mux.HandleFunc("/v1/admin/history/mutasi", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminMutasi)))
	mux.HandleFunc("/v1/admin/history/transaksi", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet", "analis")(ctrl.AdminTransaksi)))
	mux.HandleFunc("/v1/admin/history/transaksi/logs", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminTransaksiStatusLogs)))
	mux.HandleFunc("/v1/admin/history/transaksi/logs/manual", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminTransaksiStatusLogsManual)))
	mux.HandleFunc("/v1/admin/history/transaksi/cancel", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminCancelTransaksi)))
	mux.HandleFunc("/v1/admin/history/transaksi/complete", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminCompleteTransaksi)))
	mux.HandleFunc("/v1/admin/history/transaksi/send-callback", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminSendTransaksiCallback)))
	mux.HandleFunc("/v1/admin/history/transaksi/resend", wrap(helper.RequireRoles("admin", "operator_trx", "operator_wallet")(ctrl.AdminResendTransaksi)))
}
