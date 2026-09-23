package httpapi

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/gemilang"
	"pulsa2/internal/handler"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	internalrouter "pulsa2/internal/router"
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

type Deps struct {
	DB  *sql.DB
	YS  *yuscom.Client
	JP  *javapay.Client
	TL  *talenta.Client
	MK  *multikom.Client
	SG  *sagaramobile.Client
	MN  *minions.Client
	TR  *trionik.Client
	AJ  *ajs.Client
	GM  *gemilang.Client
	SM  *smb.Client
	LB  *loketbayar.Client
	CH  *chytron.Client
	RJ  *rajabiller.Client
	P24 provider.Client
}

func Routes(d Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", helper.NoDirListing(http.FileServer(http.Dir("uploads")))))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	internalrouter.JavapayInternalRouter(mux, internalrouter.JavapayInternalDeps{
		DB:       d.DB,
		JPClient: d.JP,
	})

	internalrouter.ProviderCallbackRouter(mux, internalrouter.ProviderCallbackDeps{
		DB:       d.DB,
		JPClient: d.JP,
		YSClient: d.YS,
		TLClient: d.TL,
		MKClient: d.MK,
		SGClient: d.SG,
		MNClient: d.MN,
		TRClient: d.TR,
		AJClient: d.AJ,
		GMClient: d.GM,
		SMClient: d.SM,
		LBClient: d.LB,
		CHClient: d.CH,
		RJClient: d.RJ,
	})
	internalrouter.AppOrderPaymentRouter(mux, d.DB, d.YS, d.GM, d.P24)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) < 32 {
		log.Fatalf("FATAL: JWT_SECRET must be at least 32 bytes, got %d", len(jwtSecret))
	}
	jwt := &helper.JWTAuthMiddleware{Secret: jwtSecret}

	internalrouter.Register(
		mux,
		jwt.Wrap,
		d.DB,
		jwtSecret,
		d.YS,
		d.JP,
		d.TL,
		d.MK,
		d.SG,
		d.MN,
		d.TR,
		d.AJ,
		d.GM,
		d.SM,
		d.LB,
		d.CH,
		d.RJ,
		d.P24,
	)

	internalrouter.MemberTrxRouter(mux, internalrouter.MemberTrxDeps{
		DB:           d.DB,
		YSClient:     d.YS,
		JPClient:     d.JP,
		TLClient:     d.TL,
		MKClient:     d.MK,
		SGClient:     d.SG,
		MNClient:     d.MN,
		TRClient:     d.TR,
		AJClient:     d.AJ,
		GMClient:     d.GM,
		SMClient:     d.SM,
		LBClient:     d.LB,
		CHClient:     d.CH,
		RJClient:     d.RJ,
		ExtraClients: []provider.Client{d.P24},
	})

	reconcileH := handler.NewReconcileHandler(d.DB)
	mux.HandleFunc("/v1/internal/reconcile", func(w http.ResponseWriter, r *http.Request) {
		if !helper.IsLocalRequest(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		reconcileH.Reconcile(w, r)
	})

	// Credit applications carry four compressed survey photos in one JSON payload.
	// Keep a bounded but practical limit so the request is not truncated into invalid JSON.
	return helper.PlayIntegrityGuard(helper.SanitizeErrors(helper.CORS(helper.MaxBodySize(8<<20, mux))))
}
