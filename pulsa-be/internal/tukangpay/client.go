package tukangpay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type InquiryRequest struct {
	OrderID   string `json:"order_id"`
	BankCode  string `json:"bank_code"`
	BankName  string `json:"bank_name"`
	AccountNo string `json:"account_no"`
}

type InquiryItem struct {
	ID              int64  `json:"id"`
	PublicID        string `json:"public_id"`
	OrderID         string `json:"order_id"`
	BankCode        string `json:"bank_code"`
	BankName        string `json:"bank_name"`
	AccountNo       string `json:"account_no"`
	AccountName     string `json:"account_name"`
	InquiryStatus   string `json:"inquiry_status"`
	AccountStatus   string `json:"account_status"`
	Status          string `json:"status"`
	ProviderCode    string `json:"provider_code"`
	ProviderMIDCode string `json:"provider_mid_code"`
}

type InquiryResponse struct {
	MemberID int64       `json:"member_id"`
	Inquiry  InquiryItem `json:"inquiry"`
	Error    string      `json:"error"`
	RawBody  []byte      `json:"-"`
}

type PayoutRequest struct {
	OrderID     string `json:"order_id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	BankCode    string `json:"bank_code"`
	BankName    string `json:"bank_name"`
	AccountNo   string `json:"account_no"`
	AccountName string `json:"account_name"`
}

type PayoutItem struct {
	ID                    int64   `json:"id"`
	PublicID              string  `json:"public_id"`
	OrderID               string  `json:"order_id"`
	ProviderTransactionID string  `json:"provider_transaction_id"`
	Amount                int64   `json:"amount"`
	Fee                   int64   `json:"fee"`
	Currency              string  `json:"currency"`
	Description           string  `json:"description"`
	BankCode              string  `json:"bank_code"`
	BankName              string  `json:"bank_name"`
	AccountNo             string  `json:"account_no"`
	AccountName           string  `json:"account_name"`
	Status                string  `json:"status"`
	Reason                string  `json:"reason"`
	PaidAt                *string `json:"paid_at"`
	ProviderCode          string  `json:"provider_code"`
	ProviderMIDCode       string  `json:"provider_mid_code"`
}

type PayoutResponse struct {
	Deduped bool       `json:"deduped"`
	Payout  PayoutItem `json:"payout"`
	Error   string     `json:"error"`
	RawBody []byte     `json:"-"`
}

type ProviderBalance struct {
	ProviderID       int64   `json:"provider_id"`
	ProviderMIDID    int64   `json:"provider_mid_id"`
	ProviderCode     string  `json:"provider_code"`
	ProviderMIDCode  string  `json:"provider_mid_code"`
	BalanceID        string  `json:"balance_id"`
	BalanceAvailable string  `json:"balance_available"`
	BalancePending   string  `json:"balance_pending"`
	Currency         string  `json:"currency"`
	RequestID        string  `json:"request_id"`
	Status           string  `json:"status"`
	LastError        string  `json:"last_error"`
	CheckedAt        *string `json:"checked_at"`
}

type ProviderBalanceTotal struct {
	Currency         string `json:"currency"`
	BalanceAvailable int64  `json:"balance_available"`
	BalancePending   int64  `json:"balance_pending"`
	OKCount          int    `json:"ok_count"`
}

type ProviderBalanceResponse struct {
	Balances              []ProviderBalance      `json:"balances"`
	Count                 int                    `json:"count"`
	Refreshed             bool                   `json:"refreshed"`
	Currency              string                 `json:"currency"`
	TotalBalanceAvailable int64                  `json:"total_balance_available"`
	TotalBalancePending   int64                  `json:"total_balance_pending"`
	OKCount               int                    `json:"ok_count"`
	FailedCount           int                    `json:"failed_count"`
	UnsupportedCount      int                    `json:"unsupported_count"`
	Totals                []ProviderBalanceTotal `json:"totals"`
	Error                 string                 `json:"error"`
	RawBody               []byte                 `json:"-"`
}

func NewFromEnv() *Client {
	timeout := 20 * time.Second
	if raw := strings.TrimSpace(os.Getenv("TUKANGPAY_TIMEOUT")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	return New(os.Getenv("TUKANGPAY_BASE_URL"), os.Getenv("TUKANGPAY_API_KEY"), timeout)
}

func New(baseURL, apiKey string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://tukangpay.com"
	}
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Configured() bool {
	return c != nil && strings.TrimSpace(c.baseURL) != "" && strings.TrimSpace(c.apiKey) != ""
}

func (c *Client) Inquiry(ctx context.Context, req InquiryRequest) (*InquiryResponse, error) {
	var out InquiryResponse
	raw, err := c.postJSON(ctx, "/v1/payout/inquiry", req, &out)
	out.RawBody = raw
	if err != nil {
		return &out, err
	}
	if strings.TrimSpace(out.Error) != "" {
		return &out, errors.New(out.Error)
	}
	return &out, nil
}

func (c *Client) CreatePayout(ctx context.Context, req PayoutRequest) (*PayoutResponse, error) {
	var out PayoutResponse
	raw, err := c.postJSON(ctx, "/v1/payout/transactions", req, &out)
	out.RawBody = raw
	if err != nil {
		return &out, err
	}
	if strings.TrimSpace(out.Error) != "" {
		return &out, errors.New(out.Error)
	}
	return &out, nil
}

func (c *Client) ProviderBalances(ctx context.Context) (*ProviderBalanceResponse, error) {
	var out ProviderBalanceResponse
	raw, err := c.getJSON(ctx, "/v1/provider-balances?summary=1", &out)
	out.RawBody = raw
	if err != nil {
		return &out, err
	}
	if strings.TrimSpace(out.Error) != "" {
		return &out, errors.New(out.Error)
	}
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, path string, out any) ([]byte, error) {
	if !c.Configured() {
		return nil, errors.New("credential TukangPay belum dikonfigurasi")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, out)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = resp.Status
		}
		return raw, fmt.Errorf("TukangPay HTTP %d: %s", resp.StatusCode, msg)
	}
	return raw, nil
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, out any) ([]byte, error) {
	if !c.Configured() {
		return nil, errors.New("credential TukangPay belum dikonfigurasi")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, out)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = resp.Status
		}
		return raw, fmt.Errorf("TukangPay HTTP %d: %s", resp.StatusCode, msg)
	}
	return raw, nil
}
