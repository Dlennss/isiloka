package otomax

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL  string
	MemberID string
	PIN      string
	Password string
	Timeout  time.Duration

	httpc *http.Client
}

func New(baseURL, memberID, pin, password string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		BaseURL:  baseURL,
		MemberID: strings.TrimSpace(memberID),
		PIN:      strings.TrimSpace(pin),
		Password: strings.TrimSpace(password),
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
	}
}

func (c *Client) Validate(providerName string) error {
	if c.BaseURL == "" || c.MemberID == "" {
		if strings.TrimSpace(providerName) == "" {
			providerName = "otomax"
		}
		return errors.New(strings.TrimSpace(providerName) + " config missing (baseurl/memberid)")
	}
	if c.PIN == "" || c.Password == "" {
		if strings.TrimSpace(providerName) == "" {
			providerName = "otomax"
		}
		return errors.New(strings.TrimSpace(providerName) + " config missing (pin/password)")
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

func (c *Client) Balance(ctx context.Context) (httpStatus int, body string, err error) {
	u, _ := url.Parse(c.BaseURL + "/balance")
	q := u.Query()
	q.Set("memberID", c.MemberID)
	q.Set("pin", c.PIN)
	q.Set("password", c.Password)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.httpClientForContext(ctx).Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, strings.TrimSpace(string(b)), nil
}

func (c *Client) TrxRaw(ctx context.Context, product string, qty int64, dest string, refid string) (httpStatus int, body string, err error) {
	refid = strings.TrimSpace(refid)
	if refid == "" {
		return 0, "", errors.New("refid required")
	}

	u, _ := url.Parse(c.BaseURL + "/trx")
	q := u.Query()
	q.Set("product", strings.TrimSpace(product))
	q.Set("qty", strconv.FormatInt(qty, 10))
	q.Set("dest", strings.TrimSpace(dest))
	q.Set("refID", refid)
	q.Set("memberID", c.MemberID)
	q.Set("pin", c.PIN)
	q.Set("password", c.Password)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.httpClientForContext(ctx).Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, strings.TrimSpace(string(b)), nil
}
