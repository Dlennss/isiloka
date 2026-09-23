package service

import (
	"strings"
	"time"

	"pulsa2/internal/repository"
)

const (
	depositQrisMethod = "qris"
	depositQrisPrefix = "DQR-"
)

type DepositQrisCreateResult struct {
	RefID         string              `json:"ref_id"`
	Amount        int64               `json:"amount"`
	FeeAdmin      int64               `json:"fee_admin"`
	GrossAmount   int64               `json:"gross_amount"`
	Status        string              `json:"status"`
	PaymentType   string              `json:"payment_type"`
	TransactionID string              `json:"transaction_id"`
	QRURL         string              `json:"qr_url"`
	ExpiredAt     *time.Time          `json:"expired_at,omitempty"`
	Actions       []map[string]string `json:"actions,omitempty"`
	RawResponse   map[string]any      `json:"raw_response,omitempty"`
}

func buildDepositQrisResultFromRow(row *repository.DepositRequestRow) *DepositQrisCreateResult {
	feeAdmin := calcDepositQrisFee()
	return &DepositQrisCreateResult{
		RefID:       row.RefID,
		Amount:      row.Amount,
		FeeAdmin:    feeAdmin,
		GrossAmount: row.Amount + feeAdmin,
		Status:      row.Status,
		QRURL:       strings.TrimSpace(row.BuktiURL),
	}
}
