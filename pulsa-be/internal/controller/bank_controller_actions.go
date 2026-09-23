package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
)

func (h *BankController) ToggleActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	role := bankAuthRole(a.Role)
	if role != "admin" && role != "operator_wallet" {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}

	var in struct {
		BankID int64 `json:"bank_id"`
		Aktif  bool  `json:"aktif"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	if err := h.svc.EnsureVisibleToRole(r.Context(), in.BankID, a.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	if err := h.svc.ToggleActive(r.Context(), in.BankID, in.Aktif); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
}

func (h *BankController) AdminAdjustSaldo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	role := ""
	if ok {
		role = bankAuthRole(a.Role)
	}
	if !ok || role != "admin" {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
		return
	}

	var in struct {
		BankID    int64  `json:"bank_id"`
		Amount    int64  `json:"amount"`
		Direction string `json:"direction"`
		Note      string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	refID, saldo, err := h.svc.AdjustSaldo(r.Context(), a.MemberID, in.BankID, in.Amount, in.Direction, in.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "ref_id": refID, "saldo": saldo})
}

func (h *BankController) ManualIncomingMutation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	role := ""
	if ok {
		role = bankAuthRole(a.Role)
	}
	if !ok || (role != "admin" && role != "operator_wallet") {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}

	var in struct {
		BankID          int64  `json:"bank_id"`
		Amount          int64  `json:"amount"`
		Sender          string `json:"sender"`
		Pengirim        string `json:"pengirim"`
		Receiver        string `json:"receiver"`
		Penerima        string `json:"penerima"`
		Note            string `json:"note"`
		ExternalRef     string `json:"external_ref"`
		Direction       string `json:"direction"`
		Balance         int64  `json:"balance"`
		SaldoSesudah    int64  `json:"saldo_sesudah"`
		WaktuMutasiBank string `json:"waktu_mutasi_bank"`
		BankMutationAt  string `json:"bank_mutation_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	actualBalance := in.Balance
	if actualBalance <= 0 {
		actualBalance = in.SaldoSesudah
	}
	sender := in.Sender
	if sender == "" {
		sender = in.Pengirim
	}
	receiver := in.Penerima
	if receiver == "" {
		receiver = in.Receiver
	}
	mutationTime := in.WaktuMutasiBank
	if mutationTime == "" {
		mutationTime = in.BankMutationAt
	}

	refID, saldo, duplicate, err := h.svc.ManualIncomingMutation(r.Context(), a.MemberID, a.Role, in.BankID, in.Amount, sender, receiver, in.Note, in.ExternalRef, in.Direction, actualBalance, mutationTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "ref_id": refID, "saldo": saldo, "duplicate": duplicate})
}

func (h *BankController) Kantor24IncomingMutation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}

	var in struct {
		BankID          int64  `json:"bank_id"`
		Amount          int64  `json:"amount"`
		Sender          string `json:"sender"`
		Pengirim        string `json:"pengirim"`
		Receiver        string `json:"receiver"`
		Penerima        string `json:"penerima"`
		Note            string `json:"note"`
		ExternalRef     string `json:"external_ref"`
		Direction       string `json:"direction"`
		Balance         int64  `json:"balance"`
		SaldoSesudah    int64  `json:"saldo_sesudah"`
		WaktuMutasiBank string `json:"waktu_mutasi_bank"`
		BankMutationAt  string `json:"bank_mutation_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}
	actualBalance := in.Balance
	if actualBalance <= 0 {
		actualBalance = in.SaldoSesudah
	}
	sender := in.Sender
	if sender == "" {
		sender = in.Pengirim
	}
	receiver := in.Penerima
	if receiver == "" {
		receiver = in.Receiver
	}
	mutationTime := in.WaktuMutasiBank
	if mutationTime == "" {
		mutationTime = in.BankMutationAt
	}

	refID, saldo, duplicate, err := h.svc.Kantor24IncomingMutation(r.Context(), in.BankID, in.Amount, sender, receiver, in.Note, in.ExternalRef, in.Direction, actualBalance, mutationTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"ref_id":    refID,
		"saldo":     saldo,
		"duplicate": duplicate,
		"source":    "kantor24",
	})
}

func (h *BankController) AdminTransferOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	role := ""
	if ok {
		role = bankAuthRole(a.Role)
	}
	if !ok || role != "admin" {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("admin only"))
		return
	}

	var in struct {
		BankID int64  `json:"bank_id"`
		Tujuan string `json:"tujuan"`
		Amount int64  `json:"amount"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	refID, saldo, tujuan, err := h.svc.TransferOut(r.Context(), a.MemberID, in.BankID, in.Tujuan, in.Amount, in.Note)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"ref_id": refID,
		"saldo":  saldo,
		"tujuan": tujuan,
	})
}

func (h *BankController) TransferToBCAOperational(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	role := ""
	if ok {
		role = bankAuthRole(a.Role)
	}
	if !ok || (role != "admin" && role != "operator_wallet") {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}

	var in struct {
		BankID   int64  `json:"bank_id"`
		Amount   int64  `json:"amount"`
		AdminFee int64  `json:"admin_fee"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	refID, sourceSaldo, destinationSaldo, destinationName, destinationAccount, err := h.svc.TransferToBCAOperational(
		r.Context(),
		a.MemberID,
		a.Role,
		in.BankID,
		in.Amount,
		in.AdminFee,
		in.Note,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":                     true,
		"ref_id":                 refID,
		"source_bank_saldo":      sourceSaldo,
		"destination_bank_saldo": destinationSaldo,
		"destination_bank":       destinationName,
		"destination_account":    destinationAccount,
		"admin_fee":              in.AdminFee,
		"bank_debit":             in.Amount + in.AdminFee,
	})
}

func (h *BankController) CreditProviderFromBankMutation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	role := ""
	if ok {
		role = bankAuthRole(a.Role)
	}
	if !ok || (role != "admin" && role != "operator_wallet") {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}

	var in struct {
		MutasiBankID int64  `json:"mutasi_bank_id"`
		Provider     string `json:"provider"`
		Note         string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return
	}

	result, err := h.svc.CreditProviderFromBankMutation(r.Context(), a.MemberID, a.Role, in.MutasiBankID, in.Provider, in.Note)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("mutasi/provider not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"provider":        result.Provider,
		"saldo_internal":  result.ProviderSaldo,
		"amount":          result.Amount,
		"ref_id":          result.RefID,
		"bank_id":         result.BankID,
		"bank_nama":       result.BankNama,
		"mutasi_bank_id":  in.MutasiBankID,
		"manual_provider": true,
	})
}

func (h *BankController) Kantor24LatestMutation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	bankID := helper.QueryInt64(r, "bank_id", 0)
	item, saldoTerakhir, err := h.svc.Kantor24LatestMutation(r.Context(), bankID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	resp := map[string]any{
		"ok":             true,
		"item":           item,
		"bank_id":        bankID,
		"saldo_terakhir": saldoTerakhir,
	}
	if item != nil {
		resp["ref_id_terakhir"] = item.RefID
		resp["mutasi_saldo_terakhir"] = item.SaldoSesudah
	}
	helper.WriteJSON(w, http.StatusOK, resp)
}

func (h *BankController) AdminHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	bankID := helper.QueryInt64(r, "bank_id", 0)
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	if err := h.svc.EnsureVisibleToRole(r.Context(), bankID, a.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("bank not found"))
			return
		}
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	items, total, err := h.svc.History(
		r.Context(),
		bankID,
		helper.QueryString(r, "arah"),
		helper.QueryString(r, "ref_id"),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		helper.QueryString(r, "q"),
		bankHistoryPrioritizeUnassigned(r),
		helper.QueryInt(r, "limit", 50),
		helper.QueryInt(r, "offset", 0),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total, "bank_id": bankID})
}

func (h *BankController) UnpairedDebitMutasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
		return
	}
	a, ok := helper.GetAuth(r.Context())
	if !ok {
		helper.WriteJSON(w, http.StatusForbidden, commondto.MapError("forbidden"))
		return
	}
	items, total, err := h.svc.UnpairedDebitMutasi(
		r.Context(),
		a.Role,
		helper.QueryInt64(r, "bank_id", 0),
		helper.QueryString(r, "from"),
		helper.QueryString(r, "to"),
		helper.QueryString(r, "q"),
		helper.QueryInt(r, "limit", 50),
		helper.QueryInt(r, "offset", 0),
	)
	if err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items, "total": total})
}

func bankHistoryPrioritizeUnassigned(r *http.Request) bool {
	switch strings.ToLower(helper.QueryString(r, "prioritize_unassigned")) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
