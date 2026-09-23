package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func MemberSelfRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewMemberSelfRepository(db)
	svc := service.NewMemberSelfService(repo)
	ctrl := controller.NewMemberSelfController(svc)

	mux.HandleFunc("/v1/me/profile", wrap(ctrl.Profile))
	mux.HandleFunc("/v1/me/charge-receiver", wrap(ctrl.ChargeReceiver))
	mux.HandleFunc("/v1/stats", wrap(ctrl.Stats))
	mux.HandleFunc("/v1/me/password", wrap(ctrl.ChangePassword))
	mux.HandleFunc("/v1/me/pin", wrap(ctrl.ChangePIN))
	mux.HandleFunc("/v1/me/api-key/reset", wrap(ctrl.ResetAPIKey))
	mux.HandleFunc("/v1/me/ip-whitelist", wrap(ctrl.IPWhitelist))
}
