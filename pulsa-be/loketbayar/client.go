package loketbayar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL         string
	Username        string
	Password        string
	Timeout         time.Duration
	ProductBaseURLs map[string]string

	httpc *http.Client
}

func New(baseURL, username, password string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		BaseURL:  baseURL,
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
		Timeout:  timeout,
		httpc: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) SetProductBaseURL(product, baseURL string) {
	if c == nil {
		return
	}
	product = strings.ToUpper(strings.TrimSpace(product))
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if product == "" || baseURL == "" {
		return
	}
	if c.ProductBaseURLs == nil {
		c.ProductBaseURLs = map[string]string{}
	}
	c.ProductBaseURLs[product] = baseURL
}

func (c *Client) Validate() error {
	if c.BaseURL == "" || c.Username == "" || c.Password == "" {
		return errors.New("loketbayar config missing (baseurl/username/password)")
	}
	return nil
}

func (c *Client) baseURLForProduct(product string) (string, bool) {
	product = strings.ToUpper(strings.TrimSpace(product))
	if c != nil && c.ProductBaseURLs != nil {
		if baseURL := strings.TrimSpace(c.ProductBaseURLs[product]); baseURL != "" {
			return baseURL, true
		}
	}
	return c.BaseURL, false
}

type TopupRequest struct {
	ProductCode string
	Dest        string
	RefID       string
	Nominal     int64
}

type TopupResponse struct {
	Status      string
	Keterangan  string
	ProductCode string
	Dest        string
	Price       int64
	TrxID       string
	Reff        string
	SN          string
	Saldo       int64
	Raw         map[string]any
}

type DepositTicketRequest struct {
	BankCode string
	Nominal  int64
}

type DepositTicketResponse struct {
	Status      string
	TicketID    string
	BankCode    string
	Destination string
	AccountName string
	Keterangan  string
	Raw         map[string]any
}

type DepositTicketCancelRequest struct {
	TicketID string
}

type DepositTicketCancelResponse struct {
	Status     string
	Keterangan string
	Raw        map[string]any
}

func (c *Client) Topup(ctx context.Context, in TopupRequest) (TopupResponse, int, map[string]any, error) {
	return c.trx(ctx, in, false)
}

func (c *Client) Advice(ctx context.Context, in TopupRequest) (TopupResponse, int, map[string]any, error) {
	// Dokumentasi Otomax terbaru tidak menyediakan endpoint advice terpisah.
	// Recheck memakai /trx dengan refID yang sama agar provider menangani idempotensi.
	return c.trx(ctx, in, true)
}

func (c *Client) trx(ctx context.Context, in TopupRequest, advice bool) (TopupResponse, int, map[string]any, error) {
	if err := c.Validate(); err != nil {
		return TopupResponse{}, 0, nil, err
	}

	product := strings.TrimSpace(in.ProductCode)
	dest := strings.TrimSpace(in.Dest)
	refID := strings.TrimSpace(in.RefID)
	if product == "" || dest == "" || refID == "" {
		return TopupResponse{}, 0, nil, errors.New("loketbayar trx requires product, dest, refid")
	}

	endpoint := "/trx"
	baseURL, productEndpoint := c.baseURLForProduct(product)
	reqRaw := map[string]any{
		"method":   http.MethodGet,
		"endpoint": endpoint,
		"username": "<set>",
		"password": "<redacted>",
		"product":  product,
		"dest":     dest,
		"refID":    refID,
	}
	if in.Nominal > 0 {
		reqRaw["qty"] = in.Nominal
	}
	if advice {
		reqRaw["advice"] = true
	}
	if productEndpoint {
		reqRaw["product_endpoint"] = strings.ToUpper(product)
	}

	reqURL, err := url.Parse(baseURL + endpoint)
	if err != nil {
		return TopupResponse{}, 0, reqRaw, err
	}
	q := reqURL.Query()
	q.Set("username", c.Username)
	q.Set("password", c.Password)
	q.Set("product", product)
	q.Set("dest", dest)
	q.Set("refID", refID)
	if in.Nominal > 0 {
		q.Set("qty", strconv.FormatInt(in.Nominal, 10))
	}
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return TopupResponse{}, 0, reqRaw, err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return TopupResponse{}, 0, reqRaw, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	out := parseOtomaxResponse(strings.TrimSpace(string(body)))
	if out.Raw == nil {
		out.Raw = map[string]any{}
	}
	out.Raw["endpoint"] = endpoint
	out.Raw["method"] = http.MethodGet
	if advice {
		out.Raw["advice"] = true
	}
	return out, resp.StatusCode, reqRaw, nil
}

var (
	trxProductDestRE = regexp.MustCompile(`(?i)\bTRX\s+(?:TOPUP\s+)?(\S+)\s+ke\s+(\S+)\s+status\b`)
	labelStopRE      = regexp.MustCompile(`(?i)\.\s+(?:HARGA|SALDO)\s*:`)
	snLabelStopRE    = regexp.MustCompile(`(?i)\.\s*(?:HARGA|SALDO)\s*:`)

	depositTicketIDRE      = regexp.MustCompile(`(?i)\bTIKET\s*:\s*([A-Za-z0-9._-]+)`)
	depositTicketBankRE    = regexp.MustCompile(`(?i)\bBANK\s*:\s*([A-Za-z0-9._-]+)`)
	depositTicketDestRE    = regexp.MustCompile(`(?i)\bTUJUAN\s*:\s*([0-9]+)`)
	depositTicketAccountRE = regexp.MustCompile(`(?i)\bNAMA\s*:\s*(.+?)(?:/|$)`)
)

func (c *Client) ValidateDepositTicket() error {
	if c == nil || c.BaseURL == "" || c.Username == "" {
		return errors.New("loketbayar config missing (baseurl/username)")
	}
	return nil
}

func (c *Client) DepositTicket(ctx context.Context, in DepositTicketRequest) (DepositTicketResponse, int, map[string]any, error) {
	if err := c.ValidateDepositTicket(); err != nil {
		return DepositTicketResponse{}, 0, nil, err
	}

	bankCode := strings.ToUpper(strings.TrimSpace(in.BankCode))
	if bankCode == "" || in.Nominal <= 0 {
		return DepositTicketResponse{}, 0, nil, errors.New("loketbayar deposit ticket requires bank and nominal")
	}

	const endpoint = "/tiketdeposit"
	reqRaw := map[string]any{
		"method":   http.MethodGet,
		"endpoint": endpoint,
		"username": "<set>",
		"bank":     bankCode,
		"nominal":  in.Nominal,
	}

	reqURL, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return DepositTicketResponse{}, 0, reqRaw, err
	}
	q := reqURL.Query()
	q.Set("username", c.Username)
	q.Set("bank", bankCode)
	q.Set("nominal", strconv.FormatInt(in.Nominal, 10))
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return DepositTicketResponse{}, 0, reqRaw, err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return DepositTicketResponse{}, 0, reqRaw, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	out := parseDepositTicketResponse(strings.TrimSpace(string(body)))
	out.Raw["endpoint"] = endpoint
	out.Raw["method"] = http.MethodGet
	return out, resp.StatusCode, reqRaw, nil
}

func (c *Client) CancelDepositTicket(ctx context.Context, in DepositTicketCancelRequest) (DepositTicketCancelResponse, int, map[string]any, error) {
	if err := c.ValidateDepositTicket(); err != nil {
		return DepositTicketCancelResponse{}, 0, nil, err
	}

	ticketID := strings.TrimSpace(in.TicketID)
	if ticketID == "" {
		return DepositTicketCancelResponse{}, 0, nil, errors.New("loketbayar deposit ticket cancel requires kode_tiket")
	}

	const endpoint = "/tiket/cancel"
	reqRaw := map[string]any{
		"method":     http.MethodGet,
		"endpoint":   endpoint,
		"username":   "<set>",
		"kode_tiket": ticketID,
	}

	reqURL, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return DepositTicketCancelResponse{}, 0, reqRaw, err
	}
	q := reqURL.Query()
	q.Set("username", c.Username)
	q.Set("kode_tiket", ticketID)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return DepositTicketCancelResponse{}, 0, reqRaw, err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return DepositTicketCancelResponse{}, 0, reqRaw, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	out := parseDepositTicketCancelResponse(strings.TrimSpace(string(body)))
	out.Raw["endpoint"] = endpoint
	out.Raw["method"] = http.MethodGet
	return out, resp.StatusCode, reqRaw, nil
}

func parseDepositTicketResponse(text string) DepositTicketResponse {
	raw := map[string]any{
		"_raw":   text,
		"format": "loketbayar_deposit_ticket_text",
	}
	out := DepositTicketResponse{
		Status:      detectOtomaxStatus(text),
		TicketID:    depositTicketLabel(text, depositTicketIDRE),
		BankCode:    strings.ToUpper(depositTicketLabel(text, depositTicketBankRE)),
		Destination: depositTicketLabel(text, depositTicketDestRE),
		AccountName: strings.TrimSpace(depositTicketLabel(text, depositTicketAccountRE)),
		Keterangan:  text,
		Raw:         raw,
	}
	if out.Status != "" {
		raw["status"] = out.Status
	}
	if out.TicketID != "" {
		raw["ticket"] = out.TicketID
	}
	if out.BankCode != "" {
		raw["bank"] = out.BankCode
	}
	if out.Destination != "" {
		raw["tujuan"] = out.Destination
	}
	if out.AccountName != "" {
		raw["nama"] = out.AccountName
	}
	return out
}

func parseDepositTicketCancelResponse(text string) DepositTicketCancelResponse {
	raw := map[string]any{
		"_raw":   text,
		"format": "loketbayar_deposit_ticket_cancel_text",
	}
	status := detectDepositTicketCancelStatus(text)
	out := DepositTicketCancelResponse{
		Status:     status,
		Keterangan: text,
		Raw:        raw,
	}
	if status != "" {
		raw["status"] = status
	}
	return out
}

func detectDepositTicketCancelStatus(text string) string {
	up := strings.ToUpper(strings.TrimSpace(text))
	switch {
	case strings.Contains(up, "SUKSES") || strings.Contains(up, "SUCCESS") ||
		(strings.Contains(up, "BERHASIL") && (strings.Contains(up, "BATAL") || strings.Contains(up, "CANCEL"))):
		return "SUKSES"
	case strings.Contains(up, "GAGAL") || strings.Contains(up, "FAILED") ||
		strings.Contains(up, "DITOLAK") || strings.Contains(up, "TIDAK DITEMUKAN") ||
		strings.Contains(up, "NOT FOUND"):
		return "GAGAL"
	case strings.Contains(up, "PENDING") || strings.Contains(up, "MENUNGGU") ||
		strings.Contains(up, "DALAM PROSES") || strings.Contains(up, "SEDANG DIPROSES"):
		return "PENDING"
	default:
		return detectOtomaxStatus(text)
	}
}

func depositTicketLabel(text string, re *regexp.Regexp) string {
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(match[1]), ".")
}

func parseOtomaxResponse(text string) TopupResponse {
	raw := map[string]any{
		"_raw":   text,
		"format": "otomax_text",
	}
	out := TopupResponse{
		Status:     detectOtomaxStatus(text),
		Keterangan: text,
		TrxID:      parseLeadingTrxID(text),
		Reff:       parseSNRef(text),
		SN:         parseSNRef(text),
		Price:      parseLabelAmount(text, "HARGA"),
		Saldo:      parseLabelAmount(text, "SALDO"),
		Raw:        raw,
	}
	if m := trxProductDestRE.FindStringSubmatch(text); len(m) == 3 {
		out.ProductCode = strings.TrimSpace(m[1])
		out.Dest = strings.TrimSpace(m[2])
	}
	if out.Status != "" {
		raw["status"] = out.Status
	}
	if out.TrxID != "" {
		raw["trx_id"] = out.TrxID
	}
	if out.ProductCode != "" {
		raw["produk"] = out.ProductCode
	}
	if out.Dest != "" {
		raw["idpel"] = out.Dest
	}
	if out.Reff != "" {
		raw["sn"] = out.Reff
		raw["reff"] = out.Reff
	}
	if out.Price > 0 {
		raw["harga"] = out.Price
	}
	if out.Saldo > 0 {
		raw["saldo"] = out.Saldo
	}
	return out
}

func detectOtomaxStatus(text string) string {
	up := strings.ToUpper(strings.TrimSpace(text))
	switch {
	case strings.Contains(up, "STATUS SUKSES") || strings.Contains(up, " SUKSES"):
		return "SUKSES"
	case strings.Contains(up, "STATUS PENDING") || strings.Contains(up, "PENDING") ||
		strings.Contains(up, "DALAM PROSES") || strings.Contains(up, "SEDANG DIPROSES") ||
		strings.Contains(up, "MENUNGGU"):
		return "PENDING"
	case strings.Contains(up, "STATUS GAGAL") || strings.Contains(up, "GAGAL") ||
		strings.Contains(up, "FAILED") || strings.Contains(up, "DITOLAK") ||
		strings.Contains(up, "BATAL"):
		return "GAGAL"
	default:
		return ""
	}
}

func parseLeadingTrxID(text string) string {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "#") {
		return ""
	}
	first := strings.Fields(text)
	if len(first) == 0 {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(first[0]), "#")
}

func parseSNRef(text string) string {
	idx := -1
	for _, label := range []string{"SN/REF:", "SN:"} {
		if i := strings.Index(strings.ToUpper(text), label); i >= 0 {
			idx = i + len(label)
			break
		}
	}
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(text[idx:])
	if stop := labelStopRE.FindStringIndex(rest); stop != nil {
		rest = strings.TrimSpace(rest[:stop[0]])
	} else if stop := snLabelStopRE.FindStringIndex(rest); stop != nil {
		rest = strings.TrimSpace(rest[:stop[0]])
	}
	return strings.Trim(strings.TrimSpace(rest), ".")
}

func parseLabelAmount(text, label string) int64 {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\s*:\s*([0-9][0-9.,]*)`, regexp.QuoteMeta(label)))
	m := re.FindStringSubmatch(text)
	if len(m) != 2 {
		return 0
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(m[1], ".", ""), ",", "")
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}
