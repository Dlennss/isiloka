package router

import (
	"database/sql"
	"net/http"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/gemilang"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
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

func Register(mux *http.ServeMux, wrap Middleware, db *sql.DB, jwtSecret []byte, ysClient *yuscom.Client, jpClient *javapay.Client, tlClient *talenta.Client, mkClient *multikom.Client, sgClient *sagaramobile.Client, mnClient *minions.Client, trClient *trionik.Client, ajClient *ajs.Client, gmClient *gemilang.Client, smClient *smb.Client, lbClient *loketbayar.Client, chClient *chytron.Client, rjClient *rajabiller.Client, extraClients ...provider.Client) {
	var p24Client *provider.Pulsa24JamAdapter
	for _, client := range extraClients {
		if typed, ok := client.(*provider.Pulsa24JamAdapter); ok {
			p24Client = typed
			break
		}
	}
	startPulsa24JamCatalogSync(db, p24Client)
	AuthRouter(mux, wrap, db, jwtSecret)
	AppKategoriRouter(mux, db)
	AppBrandRouter(mux, db)
	AppProdukRouter(mux, db)
	H2HProdukRouter(mux, db)
	AppAdRouter(mux, db)
	AppOrderRouter(mux, db, jwtSecret, ysClient, gmClient, extraClients...)
	AppOrderMeRouter(mux, db, jwtSecret)
	AppOrderRefundMeRouter(mux, db, jwtSecret)
	AppOrderAdminRouter(mux, wrap, db)
	AppOrderRefundAdminRouter(mux, wrap, db)
	MemberSelfRouter(mux, wrap, db)
	RetailRouter(mux, wrap, db, extraClients...)
	AgentCreditRouter(mux, wrap, db)
	H2HRouter(mux, wrap, db)
	UserRouter(mux, wrap, db)
	BankRouter(mux, wrap, db)
	DepositRouter(mux, wrap, db, lbClient, p24Client)
	HistoryRouter(mux, wrap, HistoryDeps{
		DB:       db,
		YSClient: ysClient,
		JPClient: jpClient,
		TLClient: tlClient,
		MKClient: mkClient,
		SGClient: sgClient,
		MNClient: mnClient,
		TRClient: trClient,
		AJClient: ajClient,
		GMClient: gmClient,
		SMClient: smClient,
		LBClient: lbClient,
		CHClient: chClient,
		RJClient: rjClient,
	})
	AuditRouter(mux, wrap, db, smClient)
	KategoriRouter(mux, wrap, helper.RequireRoles("admin"), db)
	BrandRouter(mux, wrap, helper.RequireRoles("admin"), db)
	ProviderRouter(mux, wrap, helper.RequireRoles("admin", "operator_wallet", "operator_trx"), db)
	ProviderRekeningRouter(mux, wrap, db)
	QRTPTransferRouter(mux, wrap, db)
	LoketBayarTransferRouter(mux, wrap, db, lbClient)
	ProviderMerchantIDRouter(mux, wrap, db)
	ProdukRouter(mux, wrap, helper.RequireRoles("admin", "operator_wallet", "operator_trx"), db)
	ProdukAppPricingRouter(mux, wrap, helper.RequireRoles("admin"), db)
	ProdukProviderMapRouter(mux, wrap, helper.RequireRoles("admin"), db)
	KategoriFeeAppRouter(mux, wrap, helper.RequireRoles("admin"), db)
	AppAdAdminRouter(mux, wrap, helper.RequireRoles("admin"), db)
	RetailAdminRouter(mux, wrap, db)
	H2HAdminRouter(mux, wrap, db)
	requireAdminNoStaff := func(next http.HandlerFunc) http.HandlerFunc {
		return helper.RequireRoles("admin")(helper.ForbidRoles(helper.RoleStaff)(next))
	}
	MemberFeeProductRouter(mux, wrap, requireAdminNoStaff, db)
	MemberFeeCategoryRouter(mux, wrap, requireAdminNoStaff, db)
	ProviderReportingRouter(mux, wrap, db)
	AdminBusinessReportRouter(mux, wrap, db)
	AuditorRouter(mux, wrap, db)
	WalletRouter(mux, wrap, db)
}
