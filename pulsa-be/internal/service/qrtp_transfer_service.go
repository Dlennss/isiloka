package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/repository"
	"pulsa2/internal/tukangpay"
)

const (
	qrtpMinTransferAmount = int64(10000)
	qrtpMaxTransferAmount = int64(50000000)
)

type QRTPTransferService struct {
	repo *repository.QRTPTransferRepository
	tp   *tukangpay.Client
}

type QRTPInquiryRequest struct {
	Provider           string `json:"provider"`
	ProviderRekeningID int64  `json:"provider_rekening_id"`
	Amount             int64  `json:"amount"`
	Note               string `json:"note"`
}

type QRTPCallbackPayout struct {
	ID                    string `json:"id"`
	OrderID               string `json:"order_id"`
	Amount                int64  `json:"amount"`
	Fee                   int64  `json:"fee"`
	Currency              string `json:"currency"`
	Status                string `json:"status"`
	Reason                string `json:"reason"`
	Provider              string `json:"provider"`
	MID                   string `json:"mid"`
	ProviderTransactionID string `json:"provider_transaction_id"`
	BankCode              string `json:"bank_code"`
	BankName              string `json:"bank_name"`
	PaidAt                string `json:"paid_at"`
}

func NewQRTPTransferService(repo *repository.QRTPTransferRepository, tp *tukangpay.Client) *QRTPTransferService {
	return &QRTPTransferService{repo: repo, tp: tp}
}

func (s *QRTPTransferService) Summary(ctx context.Context) (*repository.BankRow, error) {
	return s.repo.QRTPBank(ctx)
}

func (s *QRTPTransferService) ProviderBalances(ctx context.Context) (*tukangpay.ProviderBalanceResponse, error) {
	if s.tp == nil || !s.tp.Configured() {
		return nil, errors.New("credential TukangPay belum dikonfigurasi")
	}
	return s.tp.ProviderBalances(ctx)
}

func (s *QRTPTransferService) CreateInquiry(ctx context.Context, actorID int64, in QRTPInquiryRequest) (*repository.QRTPTransferRow, error) {
	if s.tp == nil || !s.tp.Configured() {
		return nil, errors.New("credential TukangPay belum dikonfigurasi")
	}
	in.Provider = strings.TrimSpace(strings.ToLower(in.Provider))
	in.Note = strings.TrimSpace(in.Note)
	if in.Amount < qrtpMinTransferAmount || in.Amount > qrtpMaxTransferAmount {
		return nil, fmt.Errorf("nominal transfer wajib Rp %d sampai Rp %d", qrtpMinTransferAmount, qrtpMaxTransferAmount)
	}
	if in.Provider == "" || in.ProviderRekeningID <= 0 {
		return nil, errors.New("provider dan rekening tujuan wajib dipilih")
	}
	bank, err := s.repo.QRTPBank(ctx)
	if err != nil {
		return nil, err
	}
	if bank.Saldo < in.Amount+repository.DefaultQRTPAdminFee() {
		return nil, errors.New("saldo QRTP tidak cukup")
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
	bankCode, bankName, err := resolveQRTPBankCode(account.Bank)
	if err != nil {
		return nil, err
	}
	refID := makeQRTPRef("QRTP")
	inquiryOrderID := refID

	resp, callErr := s.tp.Inquiry(ctx, tukangpay.InquiryRequest{
		OrderID:   inquiryOrderID,
		BankCode:  bankCode,
		BankName:  bankName,
		AccountNo: account.NomorRekeningDigits,
	})
	status := "inquiry_success"
	errText := ""
	item := tukangpay.InquiryItem{}
	raw := []byte(`{}`)
	if resp != nil {
		item = resp.Inquiry
		raw = resp.RawBody
	}
	if callErr != nil || strings.TrimSpace(item.AccountName) == "" || strings.EqualFold(item.Status, "failed") {
		status = "inquiry_failed"
		errText = strings.TrimSpace(fmt.Sprint(callErr))
		if errText == "" {
			errText = "inquiry rekening gagal"
		}
	}
	row, insertErr := s.repo.InsertInquiry(ctx, repository.QRTPInquiryInsert{
		RefID:              refID,
		InquiryOrderID:     inquiryOrderID,
		Provider:           in.Provider,
		ProviderRekeningID: account.ID,
		BankID:             bank.ID,
		BankCode:           bankCode,
		BankName:           firstNonEmpty(item.BankName, bankName),
		AccountNo:          account.NomorRekeningDigits,
		AccountName:        item.AccountName,
		Amount:             in.Amount,
		AdminFee:           repository.DefaultQRTPAdminFee(),
		Note:               in.Note,
		Status:             status,
		InquiryStatus:      item.InquiryStatus,
		AccountStatus:      item.AccountStatus,
		InquiryPublicID:    item.PublicID,
		InquiryError:       errText,
		InquiryRaw:         raw,
		CreatedBy:          actorID,
	})
	if insertErr != nil {
		return nil, insertErr
	}
	if status != "inquiry_success" {
		return row, errors.New(errText)
	}
	return row, nil
}

func (s *QRTPTransferService) Process(ctx context.Context, actorID, id int64) (*repository.QRTPTransferRow, error) {
	if id <= 0 {
		return nil, errors.New("request invalid")
	}
	if s.tp == nil || !s.tp.Configured() {
		return nil, errors.New("credential TukangPay belum dikonfigurasi")
	}
	claimed, err := s.repo.ClaimInternalTransfer(ctx, id, actorID)
	if err != nil {
		return nil, err
	}
	resp, callErr := s.tp.CreatePayout(ctx, tukangpay.PayoutRequest{
		OrderID:     claimed.RefID,
		Amount:      claimed.Amount,
		Currency:    "IDR",
		Description: "QRTP",
		BankCode:    claimed.BankCode,
		BankName:    claimed.BankName,
		AccountNo:   claimed.AccountNo,
		AccountName: claimed.AccountName,
	})
	item := tukangpay.PayoutItem{}
	raw := []byte(`{}`)
	errText := ""
	if resp != nil {
		item = resp.Payout
		raw = resp.RawBody
		errText = resp.Error
	}
	if callErr != nil && errText == "" {
		errText = callErr.Error()
	}
	status := strings.TrimSpace(strings.ToLower(item.Status))
	if status == "" {
		if callErr != nil {
			status = "create_failed"
		} else {
			status = "requested"
		}
	}
	updated, markErr := s.repo.MarkPayoutResponse(ctx, claimed.ID, status, item.PublicID, item.ProviderTransactionID, item.Reason, errText, raw)
	if markErr != nil {
		return nil, markErr
	}
	if callErr != nil {
		return updated, callErr
	}
	return updated, nil
}

func (s *QRTPTransferService) List(ctx context.Context, provider, status, direction, search string, limit, offset int) ([]repository.QRTPTransferLedgerRow, int64, error) {
	return s.repo.List(ctx, provider, status, direction, search, limit, offset)
}

func (s *QRTPTransferService) HandleCallback(ctx context.Context, event string, payout QRTPCallbackPayout, raw []byte) (*repository.QRTPTransferRow, error) {
	status := strings.TrimSpace(strings.ToLower(payout.Status))
	if status == "" {
		status = strings.TrimSpace(strings.ToLower(event))
	}
	if payout.OrderID == "" {
		return nil, errors.New("order_id callback kosong")
	}
	return s.repo.ApplyCallback(ctx, payout.OrderID, status, payout.ID, payout.ProviderTransactionID, payout.Reason, raw)
}

func resolveQRTPBankCode(bank string) (string, string, error) {
	upper := strings.ToUpper(strings.TrimSpace(bank))
	switch {
	case strings.Contains(upper, "BCA"):
		return "014", "BCA", nil
	case strings.Contains(upper, "BRI"):
		return "002", "BRI", nil
	case strings.Contains(upper, "BNI"):
		return "009", "BNI", nil
	case strings.Contains(upper, "MANDIRI"):
		return "008", "MANDIRI", nil
	default:
		return "", "", fmt.Errorf("bank provider %q belum punya mapping TukangPay", bank)
	}
}

func makeQRTPRef(prefix string) string {
	var buf [5]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strings.ToUpper(prefix) + "-" + time.Now().Format("20060102150405")
	}
	return strings.ToUpper(prefix) + "-" + time.Now().Format("20060102150405") + "-" + hex.EncodeToString(buf[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
