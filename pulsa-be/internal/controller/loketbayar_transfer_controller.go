package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/service"
)

type LoketBayarTransferController struct {
	svc *service.LoketBayarTransferService
}

func NewLoketBayarTransferController(svc *service.LoketBayarTransferService) *LoketBayarTransferController {
	return &LoketBayarTransferController{svc: svc}
}

func (h *LoketBayarTransferController) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	summary, err := h.svc.Summary(r.Context())
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"summary":    summary,
		"max_amount": 50000000,
	})
}

func (h *LoketBayarTransferController) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	provider := strings.TrimSpace(helper.QueryString(r, "provider"))
	status := strings.TrimSpace(helper.QueryString(r, "status"))
	search := strings.TrimSpace(helper.QueryString(r, "q"))
	limit := helper.QueryInt(r, "limit", 50)
	offset := helper.QueryInt(r, "offset", 0)
	items, total, err := h.svc.List(r.Context(), provider, status, search, limit, offset)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *LoketBayarTransferController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	var in service.LoketBayarTransferCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	row, err := h.svc.CreateRequest(r.Context(), a.MemberID, in)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
}

func (h *LoketBayarTransferController) Process(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	id, ok := loketBayarTransferPathID(w, r)
	if !ok {
		return
	}
	row, err := h.svc.Process(r.Context(), a.MemberID, id)
	if err != nil {
		payload := map[string]any{"ok": false, "error": err.Error()}
		if row != nil {
			payload["item"] = row
		}
		helper.WriteJSON(w, http.StatusBadRequest, payload)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
}

func loketBayarTransferPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/admin/loketbayar-transfer/transfers/")
	path = strings.TrimSuffix(path, "/process")
	path = strings.Trim(path, "/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid id"))
		return 0, false
	}
	return id, true
}
