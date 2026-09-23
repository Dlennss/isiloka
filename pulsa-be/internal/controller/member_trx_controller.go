package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type MemberTrxController struct {
	svc *service.MemberTrxService
}

func NewMemberTrxController(svc *service.MemberTrxService) *MemberTrxController {
	return &MemberTrxController{svc: svc}
}

func mapServiceErrStatus(kind service.ServiceErrorKind) int {
	switch kind {
	case service.ErrBadRequest:
		return http.StatusBadRequest
	case service.ErrUnauthorized:
		return http.StatusUnauthorized
	case service.ErrForbidden:
		return http.StatusForbidden
	case service.ErrUpstream:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func (c *MemberTrxController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, trxmemberdto.MapErrorResponse("method not allowed"))
		return
	}

	apiKey := strings.TrimSpace(r.Header.Get("X-Api-Key"))
	if apiKey == "" {
		helper.WriteJSON(w, http.StatusUnauthorized, trxmemberdto.MapErrorResponse("missing X-Api-Key"))
		return
	}

	var in trxmemberdto.TrxRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, trxmemberdto.MapErrorResponse("invalid json"))
		return
	}

	clientIP := helper.GetClientIP(r)
	body, svcErr := c.svc.Handle(r.Context(), apiKey, clientIP, in)
	if svcErr != nil {
		helper.WriteJSON(w, mapServiceErrStatus(svcErr.Kind), trxmemberdto.MapErrorResponse(svcErr.Message))
		return
	}
	helper.WriteJSON(w, http.StatusOK, trxmemberdto.NormalizeBusinessStatusPayload(body))
}
