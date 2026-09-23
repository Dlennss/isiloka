package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func UserRouter(mux *http.ServeMux, wrap Middleware, db *sql.DB) {
	repo := repository.NewUserRepository(db)
	retailRepo := repository.NewRetailRepository(db)
	h2hRepo := repository.NewH2HRepository(db)
	svc := service.NewUserService(repo, retailRepo, h2hRepo)
	ctrl := controller.NewUserController(svc, "/v1/admin")
	denyStaff := helper.ForbidRoles(helper.RoleStaff)
	userReadRoles := helper.RequireRoles("admin", "operator_trx", "operator_wallet", helper.RoleRetailMaster, helper.RoleRetailMarketing, helper.RoleRetailAnalyst)
	adminOnly := func(next http.HandlerFunc) http.HandlerFunc {
		return helper.RequireRoles("admin")(denyStaff(next))
	}
	adminOrWallet := helper.RequireRoles("admin", "operator_wallet")

	mux.HandleFunc("/v1/admin/users", wrap(userReadRoles(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/users/", wrap(userReadRoles(ctrl.Handle)))

	// actions
	mux.HandleFunc("/v1/admin/users/fee", wrap(adminOnly(ctrl.SetFee)))
	mux.HandleFunc("/v1/admin/users/password", wrap(adminOnly(ctrl.SetPassword)))
	mux.HandleFunc("/v1/admin/users/pin", wrap(adminOnly(ctrl.SetPIN)))
	mux.HandleFunc("/v1/admin/users/stats", wrap(adminOrWallet(ctrl.Stats)))
	mux.HandleFunc("/v1/admin/users/hierarchy/preview", wrap(adminOnly(ctrl.PreviewHierarchy)))
	mux.HandleFunc("/v1/admin/users/hierarchy/apply", wrap(adminOnly(ctrl.ApplyHierarchy)))

	// backward-compatible aliases from old member path
	mux.HandleFunc("/v1/admin/members", wrap(userReadRoles(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/members/", wrap(userReadRoles(ctrl.Handle)))
	mux.HandleFunc("/v1/admin/members/fee", wrap(adminOnly(ctrl.SetFee)))
	mux.HandleFunc("/v1/admin/members/password", wrap(adminOnly(ctrl.SetPassword)))
	mux.HandleFunc("/v1/admin/members/pin", wrap(adminOnly(ctrl.SetPIN)))
	mux.HandleFunc("/v1/admin/members/stats", wrap(adminOrWallet(ctrl.Stats)))
	mux.HandleFunc("/v1/admin/members/hierarchy/preview", wrap(adminOnly(ctrl.PreviewHierarchy)))
	mux.HandleFunc("/v1/admin/members/hierarchy/apply", wrap(adminOnly(ctrl.ApplyHierarchy)))
	mux.HandleFunc("/v1/master/operator/marketing", wrap(helper.RequireRoles("admin", helper.RoleRetailAnalyst)(ctrl.CreateMarketing)))
}
