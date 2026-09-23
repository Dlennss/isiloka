package controller

import (
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type H2HProdukController struct {
	svc *service.H2HProdukService
}

func NewH2HProdukController(svc *service.H2HProdukService) *H2HProdukController {
	return &H2HProdukController{svc: svc}
}

func (c *H2HProdukController) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	apiKey := strings.TrimSpace(r.Header.Get("X-Api-Key"))
	if apiKey == "" {
		helper.WriteJSON(w, http.StatusUnauthorized, commondto.MapError("missing X-Api-Key"))
		return
	}

	rows, err := c.svc.List(
		r.Context(),
		apiKey,
		helper.QueryString(r, "q"),
		helper.QueryString(r, "kategori"),
		helper.QueryString(r, "brand"),
	)
	if svcErr, ok := err.(*service.ServiceError); ok {
		helper.WriteJSON(w, mapServiceErrStatus(svcErr.Kind), commondto.MapError(svcErr.Message))
		return
	}
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"items": trxmemberdto.MapH2HProdukItems(rows),
	})
}
