package rajabiller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	UID        string
	PIN        string
	MerchantID string
	Timeout    time.Duration

	httpc *http.Client
}

func New(baseURL, uid, pin string, timeout time.Duration) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return &Client{
		BaseURL: baseURL,
		UID:     strings.TrimSpace(uid),
		PIN:     strings.TrimSpace(pin),
		Timeout: timeout,
		httpc: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Validate() error {
	if c.BaseURL == "" || c.UID == "" || c.PIN == "" {
		return errors.New("rajabiller config missing (baseurl/uid/pin)")
	}
	return nil
}

type TransactionRequest struct {
	Method      string
	Product     string
	Dest        string
	RefID       string
	Nominal     int64
	SendNominal bool
	Server      string
	KodeBank    string
	HP          string
	Berita      string
	MerchantID  string
}

type TransactionResponse struct {
	TrxID          string
	RC             string
	Status         string
	Product        string
	Dest           string
	SN             string
	Token          string
	Ref            string
	ProviderRefID  string
	Price          int64
	Balance        int64
	SaldoTerpotong int64
	Raw            map[string]any
}

func (c *Client) Transaction(ctx context.Context, in TransactionRequest) (TransactionResponse, int, map[string]any, error) {
	if err := c.Validate(); err != nil {
		return TransactionResponse{}, 0, nil, err
	}

	method := strings.ToLower(strings.TrimSpace(in.Method))
	if method == "" {
		method = "bayar"
	}
	product := strings.TrimSpace(in.Product)
	dest := strings.TrimSpace(in.Dest)
	refID := strings.TrimSpace(in.RefID)
	if product == "" || dest == "" || refID == "" {
		return TransactionResponse{}, 0, nil, errors.New("rajabiller transaction requires product, dest, refid")
	}

	reqBody := map[string]any{
		"method": method,
		"uid":    c.UID,
		"pin":    c.PIN,
		"produk": product,
		"idpel":  dest,
		"ref1":   refID,
	}
	if in.SendNominal && in.Nominal > 0 {
		reqBody["nominal"] = strconv.FormatInt(in.Nominal, 10)
	}
	if server := strings.TrimSpace(in.Server); server != "" {
		reqBody["server"] = server
	}
	if kodeBank := strings.TrimSpace(in.KodeBank); kodeBank != "" {
		reqBody["kodebank"] = kodeBank
	}
	if hp := strings.TrimSpace(in.HP); hp != "" {
		reqBody["hp"] = hp
	}
	if berita := strings.TrimSpace(in.Berita); berita != "" {
		reqBody["berita"] = berita
	}
	merchantID := strings.TrimSpace(in.MerchantID)
	if merchantID == "" {
		merchantID = strings.TrimSpace(c.MerchantID)
	}
	if merchantID != "" {
		reqBody["id_merchant"] = merchantID
	}

	return c.postJSON(ctx, reqBody)
}

func (c *Client) Status(ctx context.Context, refID string, tanggal time.Time) (TransactionResponse, int, map[string]any, error) {
	if err := c.Validate(); err != nil {
		return TransactionResponse{}, 0, nil, err
	}
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return TransactionResponse{}, 0, nil, errors.New("rajabiller status requires refid")
	}
	if tanggal.IsZero() {
		tanggal = time.Now()
	}
	reqBody := map[string]any{
		"method":  "status",
		"uid":     c.UID,
		"pin":     c.PIN,
		"ref":     refID,
		"tanggal": tanggal.Format("2006-01-02"),
	}
	return c.postJSON(ctx, reqBody)
}

func (c *Client) postJSON(ctx context.Context, reqBody map[string]any) (TransactionResponse, int, map[string]any, error) {
	reqLog := make(map[string]any, len(reqBody))
	for k, v := range reqBody {
		reqLog[k] = v
	}
	reqLog["pin"] = "***"

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(b))
	if err != nil {
		return TransactionResponse{}, 0, reqLog, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return TransactionResponse{}, 0, reqLog, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	raw := map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		raw = map[string]any{"_raw": strings.TrimSpace(string(body))}
	}
	return parseTransactionResponse(raw), resp.StatusCode, reqLog, nil
}

func parseTransactionResponse(raw map[string]any) TransactionResponse {
	return TransactionResponse{
		TrxID:          strings.TrimSpace(toString(raw["trxid"])),
		RC:             strings.TrimSpace(toString(raw["rc"])),
		Status:         strings.TrimSpace(toString(raw["status"])),
		Product:        strings.TrimSpace(toString(raw["produk"])),
		Dest:           strings.TrimSpace(toString(raw["idpel"])),
		SN:             strings.TrimSpace(toString(raw["sn"])),
		Token:          strings.TrimSpace(toString(raw["token"])),
		Ref:            strings.TrimSpace(toString(raw["ref"])),
		ProviderRefID:  strings.TrimSpace(toString(raw["refid"])),
		Price:          firstPositiveInt64(raw["harga"], raw["saldo_terpotong"], raw["total_bayar"], raw["tagihan"]),
		Balance:        firstPositiveInt64(raw["saldo_akhir"], raw["sisa_saldo"], raw["saldo"]),
		SaldoTerpotong: toInt64(raw["saldo_terpotong"]),
		Raw:            raw,
	}
}

func firstPositiveInt64(values ...any) int64 {
	for _, v := range values {
		n := toInt64(v)
		if n > 0 {
			return n
		}
	}
	return 0
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	case json.Number:
		n, _ := x.Int64()
		return n
	case string:
		clean := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(x), ".", ""), ",", "")
		n, _ := strconv.ParseInt(clean, 10, 64)
		return n
	default:
		return 0
	}
}
