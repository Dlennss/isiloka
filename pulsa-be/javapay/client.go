package javapay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL  string
	MemberID string
	APIKey   string
	PIN      string
	Timeout  time.Duration

	httpc *http.Client
	loc   *time.Location
}

func New(baseURL, memberID, apiKey, pin string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")

	return &Client{
		BaseURL:  baseURL,
		MemberID: memberID,
		APIKey:   apiKey,
		PIN:      pin,
		Timeout:  timeout,
		httpc: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DisableKeepAlives:     true,
				ResponseHeaderTimeout: timeout,
				IdleConnTimeout:       30 * time.Second,
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).DialContext,
			},
		},
		loc: loc,
	}
}

func (c *Client) validateBase() error {
	if c.BaseURL == "" || c.MemberID == "" || c.APIKey == "" {
		return errors.New("javapay config missing (baseurl/memberid/apikey)")
	}
	return nil
}

func (c *Client) validateSigned() error {
	if err := c.validateBase(); err != nil {
		return err
	}
	if c.PIN == "" {
		return errors.New("javapay pin missing")
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

// normalizeRefID memaksa refid konsisten (tidak berubah karena spasi/kasus).
// Tidak mengubah isi refid selain trimming; ini memastikan refid yang dipakai untuk
// sign dan payload selalu identik.
func normalizeRefID(refid string) string {
	return strings.TrimSpace(refid)
}

func (c *Client) Trx(ctx context.Context, commands, product, dest string, qty int64, refid string) (map[string]any, int, map[string]any, error) {
	if err := c.validateSigned(); err != nil {
		return nil, 0, nil, err
	}

	commands = strings.ToUpper(strings.TrimSpace(commands))
	if commands == "STATUS" || commands == "STATUS-PAY" {
		return c.Status(ctx, refid)
	}

	refid = normalizeRefID(refid)
	if refid == "" {
		return nil, 0, nil, errors.New("refid required")
	}
	product = strings.TrimSpace(product)
	dest = strings.TrimSpace(dest)
	if commands == "" {
		return nil, 0, nil, errors.New("commands required")
	}
	if product == "" {
		return nil, 0, nil, errors.New("product required")
	}
	if dest == "" {
		return nil, 0, nil, errors.New("dest required")
	}
	if qty <= 0 {
		return nil, 0, nil, errors.New("qty required")
	}

	nowJakarta := time.Now().In(c.loc)
	sign, _ := Sign(c.MemberID, c.APIKey, c.PIN, refid, product, dest, nowJakarta)

	reqBody := map[string]any{
		"commands": commands,
		"memberid": c.MemberID,
		"product":  product,
		"dest":     dest,
		"qty":      qty,
		"refid":    refid,
		"sign":     sign,
	}

	return c.postJSON(ctx, c.BaseURL+"/trx", reqBody)
}

// STATUS tetap dikirim ke endpoint /trx, tetapi payload terbaru hanya
// membutuhkan memberid, refid, dan sign dengan product/dest kosong.
func (c *Client) Status(ctx context.Context, refid string) (map[string]any, int, map[string]any, error) {
	if err := c.validateSigned(); err != nil {
		return nil, 0, nil, err
	}

	refid = normalizeRefID(refid)
	if refid == "" {
		return nil, 0, nil, errors.New("refid required")
	}

	nowJakarta := time.Now().In(c.loc)
	sign, _ := Sign(c.MemberID, c.APIKey, c.PIN, refid, "", "", nowJakarta)

	reqBody := map[string]any{
		"commands": "STATUS",
		"memberid": c.MemberID,
		"refid":    refid,
		"sign":     sign,
	}

	return c.postJSON(ctx, c.BaseURL+"/trx", reqBody)
}

func (c *Client) VerifyCallbackSignature(rawBody []byte, signature string) bool {
	signature = strings.ToLower(strings.TrimSpace(signature))
	if strings.TrimSpace(c.APIKey) == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(c.APIKey))
	_, _ = mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (c *Client) postJSON(ctx context.Context, url string, reqBody map[string]any) (map[string]any, int, map[string]any, error) {
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, 0, reqBody, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClientForContext(ctx).Do(req)
	if err != nil {
		return nil, 0, reqBody, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		out = map[string]any{"_raw": string(body)}
	}
	return out, resp.StatusCode, reqBody, nil
}

func (c *Client) ProdukStatus(ctx context.Context) (map[string]any, int, error) {
	if err := c.validateBase(); err != nil {
		return nil, 0, err
	}

	reqBody := map[string]any{
		"memberid": c.MemberID,
		"apikey":   c.APIKey,
	}

	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/report/produk",
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClientForContext(ctx).Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		out = map[string]any{"_raw": string(body)}
	}

	return out, resp.StatusCode, nil
}
