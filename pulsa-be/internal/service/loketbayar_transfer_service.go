package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/loketbayar"
)

const (
	loketBayarTransferMinAmount = int64(10000)
	loketBayarTransferMaxAmount = int64(50000000)
)

type LoketBayarTransferService struct {
	repo *repository.LoketBayarTransferRepository
	lb   *loketbayar.Client
}

type LoketBayarTransferCreateRequest struct {
	Provider           string `json:"provider"`
	ProviderRekeningID int64  `json:"provider_rekening_id"`
	Amount             int64  `json:"amount"`
	Note               string `json:"note"`
}

func NewLoketBayarTransferService(repo *repository.LoketBayarTransferRepository, lb *loketbayar.Client) *LoketBayarTransferService {
	return &LoketBayarTransferService{repo: repo, lb: lb}
}

func (s *LoketBayarTransferService) Summary(ctx context.Context) (*repository.LoketBayarTransferSummary, error) {
	return s.repo.Summary(ctx, loketBayarTransferMaxAmount)
}

func (s *LoketBayarTransferService) List(ctx context.Context, provider, status, search string, limit, offset int) ([]repository.LoketBayarTransferLedgerRow, int64, error) {
	return s.repo.List(ctx, provider, status, search, limit, offset)
}

func (s *LoketBayarTransferService) CreateRequest(ctx context.Context, actorID int64, in LoketBayarTransferCreateRequest) (*repository.LoketBayarTransferRow, error) {
	if s.lb == nil || s.lb.Validate() != nil {
		return nil, errors.New("credential LoketBayar belum dikonfigurasi")
	}
	in.Provider = strings.TrimSpace(strings.ToLower(in.Provider))
	in.Note = strings.TrimSpace(in.Note)
	if in.Provider == "" || in.ProviderRekeningID <= 0 {
		return nil, errors.New("provider dan rekening tujuan wajib dipilih")
	}
	if in.Provider == "loketbayar" {
		return nil, errors.New("tujuan provider tidak boleh LoketBayar")
	}
	if in.Amount < loketBayarTransferMinAmount || in.Amount > loketBayarTransferMaxAmount {
		return nil, fmt.Errorf("nominal transfer wajib Rp %d sampai Rp %d", loketBayarTransferMinAmount, loketBayarTransferMaxAmount)
	}

	summary, err := s.repo.Summary(ctx, loketBayarTransferMaxAmount)
	if err != nil {
		return nil, err
	}
	if summary.SaldoInternal < in.Amount {
		return nil, errors.New("saldo internal LoketBayar tidak cukup")
	}

	account, err := s.repo.GetProviderRekening(ctx, in.ProviderRekeningID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(strings.ToLower(account.Provider)) != in.Provider {
		return nil, errors.New("rekening tujuan tidak sesuai provider")
	}
	if !account.Aktif {
		return nil, errors.New("rekening provider tidak aktif")
	}
	if strings.TrimSpace(account.NomorRekeningDigits) == "" {
		return nil, errors.New("nomor rekening provider kosong")
	}

	bankCode, bankName, err := s.repo.ResolveBankCode(ctx, account.Bank)
	if err != nil {
		return nil, err
	}

	return s.repo.InsertRequest(ctx, repository.LoketBayarTransferInsert{
		RefID:              makeLoketBayarTransferRef(),
		Provider:           in.Provider,
		ProviderRekeningID: account.ID,
		BankCode:           bankCode,
		BankName:           firstNonEmpty(bankName, account.Bank),
		AccountNo:          account.NomorRekeningDigits,
		AccountName:        account.Nama,
		Amount:             in.Amount,
		Note:               in.Note,
		CreatedBy:          actorID,
	})
}

func (s *LoketBayarTransferService) Process(ctx context.Context, actorID, id int64) (*repository.LoketBayarTransferRow, error) {
	if id <= 0 {
		return nil, errors.New("request invalid")
	}
	if s.lb == nil || s.lb.Validate() != nil {
		return nil, errors.New("credential LoketBayar belum dikonfigurasi")
	}

	claimed, err := s.repo.ClaimInternalTransfer(ctx, id, actorID)
	if err != nil {
		return nil, err
	}

	resp, httpStatus, reqRaw, callErr := s.lb.Topup(ctx, loketbayar.TopupRequest{
		ProductCode: loketBayarBankTransferProduct,
		Dest:        buildLoketBayarTransferDest(claimed.BankCode, claimed.AccountNo),
		RefID:       claimed.RefID,
		Nominal:     claimed.Amount,
	})
	raw := loketBayarTransferRaw(reqRaw, resp, httpStatus, callErr)
	status, errText := loketBayarTransferStatus(resp, httpStatus, callErr)
	var snapshot *int64
	if resp.Saldo > 0 {
		v := resp.Saldo
		snapshot = &v
	}
	updated, markErr := s.repo.MarkProviderResponse(
		ctx,
		claimed.ID,
		status,
		firstNonEmpty(resp.Reff, resp.SN, resp.TrxID),
		resp.Keterangan,
		errText,
		resp.Price,
		snapshot,
		raw,
	)
	if markErr != nil {
		return nil, markErr
	}
	if snapshot != nil {
		_ = s.repo.InsertSourceSnapshot(ctx, claimed.RefID, *snapshot, raw, "loketbayar_transfer_response")
	}
	if callErr != nil {
		return updated, callErr
	}
	if httpStatus >= 400 && status == "create_failed" {
		return updated, fmt.Errorf("LoketBayar HTTP %d", httpStatus)
	}
	return updated, nil
}

func makeLoketBayarTransferRef() string {
	var buf [5]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "LKB-" + time.Now().Format("20060102150405")
	}
	return "LKB-" + time.Now().Format("20060102150405") + "-" + hex.EncodeToString(buf[:])
}

func buildLoketBayarTransferDest(bankCode, accountNo string) string {
	return strings.TrimSpace(bankCode) + strings.TrimSpace(accountNo)
}

func loketBayarTransferStatus(resp loketbayar.TopupResponse, httpStatus int, callErr error) (string, string) {
	if callErr != nil {
		return "create_failed", callErr.Error()
	}
	state := helper.ProviderResponseStateOf("loketbayar", resp.Status, resp.Keterangan)
	switch state {
	case helper.ProviderResponseSuccess:
		return "success", ""
	case helper.ProviderResponseFailed:
		return "failed", ""
	case helper.ProviderResponsePending:
		return "processing", ""
	default:
		if httpStatus >= 400 {
			return "create_failed", fmt.Sprintf("LoketBayar HTTP %d", httpStatus)
		}
		return "processing", ""
	}
}

func loketBayarTransferRaw(reqRaw map[string]any, resp loketbayar.TopupResponse, httpStatus int, callErr error) []byte {
	responseRaw := resp.Raw
	if responseRaw == nil {
		responseRaw = map[string]any{}
	}
	payload := map[string]any{
		"request":     reqRaw,
		"response":    responseRaw,
		"http_status": httpStatus,
		"parsed": map[string]any{
			"status":       resp.Status,
			"keterangan":   resp.Keterangan,
			"product_code": resp.ProductCode,
			"dest":         resp.Dest,
			"price":        resp.Price,
			"trx_id":       resp.TrxID,
			"reff":         resp.Reff,
			"sn":           resp.SN,
			"saldo":        resp.Saldo,
		},
	}
	if callErr != nil {
		payload["error"] = callErr.Error()
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}
