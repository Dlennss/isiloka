package sagaramobile

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL  string
	MemberID string
	Sign     string
	OID      string
	PIN      string
	Password string
	Timeout  time.Duration

	httpc *http.Client
}

func New(baseURL, memberID, sign, oid, pin, password string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		BaseURL:  baseURL,
		MemberID: strings.TrimSpace(memberID),
		Sign:     strings.TrimSpace(sign),
		OID:      strings.TrimSpace(oid),
		PIN:      strings.TrimSpace(pin),
		Password: strings.TrimSpace(password),
		Timeout:  timeout,
		httpc: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DisableKeepAlives:     true,
				ResponseHeaderTimeout: timeout,
				IdleConnTimeout:       30 * time.Second,
				DialContext:       (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			},
		},
	}
}

func (c *Client) signToken() string {
	if strings.TrimSpace(c.Sign) != "" {
		return strings.TrimSpace(c.Sign)
	}
	return strings.TrimSpace(c.OID)
}

func (c *Client) hasNoSignAuth() bool {
	return strings.TrimSpace(c.PIN) != "" && strings.TrimSpace(c.Password) != ""
}

func (c *Client) validate() error {
	if strings.TrimSpace(c.BaseURL) == "" || strings.TrimSpace(c.MemberID) == "" {
		return errors.New("sagaramobile config missing (baseurl/memberid)")
	}
	if c.hasNoSignAuth() {
		return nil
	}
	if strings.TrimSpace(c.signToken()) == "" {
		return errors.New("sagaramobile config missing (pin/password or sign/oid)")
	}
	return nil
}

func (c *Client) httpClientForContext(ctx context.Context) *http.Client {
	if c == nil || c.httpc == nil {
		return &http.Client{}
	}
	if ctx == nil {
		return c.httpc
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return c.httpc
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		remaining = time.Millisecond
	}
	if c.httpc.Timeout > 0 && remaining >= c.httpc.Timeout {
		return c.httpc
	}
	clone := *c.httpc
	clone.Timeout = remaining
	return &clone
}

type TrxAccepted struct {
	RefID       string
	ProviderRef string
	Body        string
	Price       int64
}

var (
	reStatus     = regexp.MustCompile(`(?i)\bstatus\s*=\s*([0-9]{1,3})\b`)
	reRefID      = regexp.MustCompile(`(?i)R#\s*([A-Z0-9_-]{6,})`)
	rePriceEq    = regexp.MustCompile(`(?i)\bharga\s*=\s*([0-9][0-9\.,]*)`)
	rePriceColon = regexp.MustCompile(`(?i)\bharga\s*:\s*([0-9][0-9\.,]*)`)
	reSN         = regexp.MustCompile(`(?i)\bSN(?:\s*/\s*Ref)?\s*:\s*([^\r\n]+)`)
	reSaldoEq    = regexp.MustCompile(`(?i)\bSaldo\b[^=\r\n]*=\s*([0-9][0-9\.,]*)`)
)

func ExtractStatusCode(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	m := reStatus.FindStringSubmatch(body)
	if len(m) != 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func ParseRefIDFromMsg(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	m := reRefID.FindStringSubmatch(strings.ToUpper(body))
	if len(m) != 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func ParsePriceFromMsg(body string) int64 {
	body = strings.TrimSpace(body)
	if body == "" {
		return 0
	}
	for _, re := range []*regexp.Regexp{rePriceEq, rePriceColon} {
		if m := re.FindStringSubmatch(body); len(m) == 2 {
			return parseAmount(m[1])
		}
	}
	if i := strings.LastIndex(body, " - "); i >= 0 {
		seg := strings.TrimSpace(body[i+3:])
		if j := strings.Index(seg, " = "); j >= 0 {
			return parseAmount(seg[:j])
		}
	}
	return 0
}

func ParseLastBalanceFromMsg(body string) (int64, bool) {
	body = strings.TrimSpace(body)
	if body == "" {
		return 0, false
	}
	if m := reSaldoEq.FindStringSubmatch(body); len(m) == 2 {
		n := parseAmount(m[1])
		return n, n > 0
	}
	idx := strings.LastIndex(body, "=")
	if idx < 0 || idx+1 >= len(body) {
		return 0, false
	}
	seg := strings.TrimSpace(body[idx+1:])
	if i := strings.Index(seg, "@"); i >= 0 {
		seg = strings.TrimSpace(seg[:i])
	}
	if seg == "" {
		return 0, false
	}
	n := parseAmount(seg)
	return n, n > 0
}

func parseAmount(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func ParseProviderRefFromMsg(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	m := reSN.FindStringSubmatch(body)
	if len(m) != 2 {
		return ""
	}
	seg := strings.TrimSpace(m[1])
	if i := strings.Index(strings.ToUpper(seg), "SALDO"); i >= 0 {
		seg = strings.TrimSpace(seg[:i])
	}
	seg = strings.Trim(seg, " .,:;")
	if seg == "" || seg == "." {
		return ""
	}
	return seg
}

func (c *Client) Trx(ctx context.Context, product, dest string, qty int64, refid string) (acc TrxAccepted, httpStatus int, body string, err error) {
	if err := c.validate(); err != nil {
		return TrxAccepted{}, 0, "", err
	}
	refid = strings.TrimSpace(refid)
	if refid == "" {
		return TrxAccepted{}, 0, "", errors.New("refid required")
	}
	product = strings.TrimSpace(product)
	dest = strings.TrimSpace(dest)
	if product == "" {
		return TrxAccepted{}, 0, "", errors.New("product required")
	}
	if dest == "" {
		return TrxAccepted{}, 0, "", errors.New("dest required")
	}
	if qty <= 0 {
		return TrxAccepted{}, 0, "", errors.New("qty required")
	}

	u, _ := url.Parse(c.BaseURL + "/trx")
	q := u.Query()
	q.Set("product", product)
	q.Set("dest", dest)
	q.Set("qty", strconv.FormatInt(qty, 10))
	q.Set("refID", refid)
	q.Set("memberID", c.MemberID)
	if c.hasNoSignAuth() {
		q.Set("pin", c.PIN)
		q.Set("password", c.Password)
	} else {
		q.Set("sign", c.signToken())
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.httpClientForContext(ctx).Do(req)
	if err != nil {
		return TrxAccepted{}, 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	body = strings.TrimSpace(string(b))
	return TrxAccepted{
		RefID:       ParseRefIDFromMsg(body),
		ProviderRef: ParseProviderRefFromMsg(body),
		Body:        body,
		Price:       ParsePriceFromMsg(body),
	}, resp.StatusCode, body, nil
}

func LooksLikeSuccess(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL")
}

func LooksLikePending(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "AKAN DIPROSES") || strings.Contains(up, "MENUNGGU JAWABAN") || strings.Contains(up, "SEDANG DIPROSES")
}

func LooksLikeAccepted(body string) bool {
	return LooksLikeSuccess(body) || LooksLikePending(body)
}

func LooksLikeImmediateReject(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "GAGAL") ||
		strings.Contains(up, "PRODUK TIDAK TERDAFTAR") ||
		strings.Contains(up, "SALDO TIDAK CUKUP") ||
		strings.Contains(up, "NOMOR TUJUAN SALAH") ||
		strings.Contains(up, "NOMOR TIDAK DAPAT DIPROSES") ||
		strings.Contains(up, "DIBATALKAN") ||
		strings.Contains(up, "STOK SEDANG KOSONG") ||
		strings.Contains(up, "PRODUK SEDANG GANGGUAN") ||
		strings.Contains(up, "TIMEOUT")
}

func LooksLikeSystemIssue(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return true
	}
	if strings.Contains(up, "MAINTEN") || strings.Contains(up, "SIBUK") || strings.Contains(up, "INVALID SIGN") || strings.Contains(up, "PASSWORD IP") {
		return true
	}
	return !LooksLikeAccepted(body) && !LooksLikeImmediateReject(body)
}
