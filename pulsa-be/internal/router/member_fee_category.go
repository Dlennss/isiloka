package router

import (
	"database/sql"
	"net/http"

	"pulsa2/internal/controller"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

func MemberFeeCategoryRouter(mux *http.ServeMux, wrap Middleware, requireAdmin Middleware, db *sql.DB) {
	repo := repository.NewMemberFeeCategoryRepository(db)
	svc := service.NewMemberFeeCategoryService(repo)
	ctrl := controller.NewMemberFeeCategoryController(svc)

	mux.HandleFunc("/v1/admin/members/fee/categories", wrap(requireAdmin(ctrl.List)))
	mux.HandleFunc("/v1/admin/members/fee/categories/upsert", wrap(requireAdmin(ctrl.Upsert)))
	mux.HandleFunc("/v1/admin/members/fee/categories/delete", wrap(requireAdmin(ctrl.Delete)))
}
