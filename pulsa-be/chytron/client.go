package chytron

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
	ID       string
	PIN      string
	User     string
	Password string
	Timeout  time.Duration

	httpc *http.Client
}

type Request struct {
	KodeProduk string
	Tujuan     string
	Qty        int64
	RefID      string
}

type StepResult struct {
	Kind       string         `json:"kind"`
	URL        string         `json:"url"`
	HTTPStatus int            `json:"http_status"`
	Body       string         `json:"body"`
	Request    map[string]any `json:"request,omitempty"`
}

type DispatchResult struct {
	Pay StepResult `json:"pay"`
}

func New(baseURL, id, pin, user, password string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		BaseURL:  baseURL,
		ID:       strings.TrimSpace(id),
		PIN:      strings.TrimSpace(pin),
		User:     strings.TrimSpace(user),
		Password: strings.TrimSpace(password),
		Timeout:  timeout,
		httpc: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DisableKeepAlives:     true,
				ResponseHeaderTimeout: timeout,
				IdleConnTimeout:       30 * time.Second,
				DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			},
		},
	}
}

func (c *Client) Validate() error {
	if c == nil {
		return errors.New("chytron client nil")
	}
	if c.BaseURL == "" {
		return errors.New("chytron config missing (baseurl)")
	}
	if c.ID == "" || c.PIN == "" || c.User == "" || c.Password == "" {
		return errors.New("chytron config missing (id/pin/user/password)")
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

func (c *Client) Pay(ctx context.Context, req Request) (*DispatchResult, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	req.KodeProduk = strings.ToUpper(strings.TrimSpace(req.KodeProduk))
	req.Tujuan = strings.TrimSpace(req.Tujuan)
	req.RefID = strings.TrimSpace(req.RefID)
	if req.KodeProduk == "" || req.Tujuan == "" || req.RefID == "" {
		return nil, errors.New("chytron request tidak lengkap")
	}
	if req.Qty <= 0 {
		return nil, errors.New("chytron qty harus > 0")
	}

	params := buildParams(c, req)
	step, err := c.call(ctx, params, redactedParams(c, req))
	if step != nil {
		step.Kind = "pay"
	}
	if err != nil {
		return nil, err
	}
	return &DispatchResult{Pay: *step}, nil
}

func buildParams(c *Client, req Request) url.Values {
	q := url.Values{}
	q.Set("id", c.ID)
	q.Set("pin", c.PIN)
	q.Set("user", c.User)
	q.Set("pass", c.Password)
	q.Set("kodeproduk", req.KodeProduk)
	q.Set("tujuan", strings.TrimSpace(req.Tujuan))
	q.Set("idtrx", req.RefID)
	q.Set("counter", "1")
	q.Set("amount", strconv.FormatInt(req.Qty, 10))
	q.Set("jenis", "1")
	return q
}

func redactedParams(c *Client, req Request) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"pin":        "***",
		"user":       c.User,
		"pass":       "***",
		"kodeproduk": req.KodeProduk,
		"tujuan":     strings.TrimSpace(req.Tujuan),
		"idtrx":      req.RefID,
		"counter":    "1",
		"amount":     strconv.FormatInt(req.Qty, 10),
		"jenis":      "1",
	}
}

func (c *Client) call(ctx context.Context, q url.Values, safeReq map[string]any) (*StepResult, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(c.BaseURL), "/") + "/api/h2h")
	if err != nil {
		return nil, err
	}
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := c.httpClientForContext(ctx).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return &StepResult{
		URL:        redactedURL(u),
		HTTPStatus: resp.StatusCode,
		Body:       strings.TrimSpace(string(b)),
		Request:    safeReq,
	}, nil
}

func redactedURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	clone := *u
	q := clone.Query()
	if q.Has("pin") {
		q.Set("pin", "***")
	}
	if q.Has("pass") {
		q.Set("pass", "***")
	}
	clone.RawQuery = q.Encode()
	return clone.String()
}

func ParseAmount(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
