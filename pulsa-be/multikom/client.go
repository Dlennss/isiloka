package multikom

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pulsa2/otomax"
)

type Client struct {
	base *otomax.Client
}

func New(baseURL, memberID, pin, password string, timeout time.Duration) *Client {
	return &Client{
		base: otomax.New(baseURL, memberID, pin, password, timeout),
	}
}

func (c *Client) validate() error {
	return c.base.Validate("multikom")
}

type TrxAccepted struct {
	RefID  string
	Ticket string
	Body   string
	Price  int64
}

var (
	reR         = regexp.MustCompile(`R#([0-9]{6,})`)
	reT         = regexp.MustCompile(`T#([0-9]{3,})`)
	rePriceDash = regexp.MustCompile(`-\s*([0-9][0-9\.\,]*)\s*=`)
	reAccepted  = regexp.MustCompile(`akan\s+diproses|AKAN\s+DIPROSES`)
)

func parseIDDigits(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "R#")
	s = strings.TrimPrefix(s, "T#")
	s = strings.TrimSpace(s)
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			out = append(out, s[i])
		}
	}
	return string(out)
}

func parsePriceIDFormat(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func (c *Client) Balance(ctx context.Context) (httpStatus int, body string, err error) {
	if err := c.validate(); err != nil {
		return 0, "", err
	}
	return c.base.Balance(ctx)
}

func (c *Client) TrxNoSign(ctx context.Context, product string, qty int64, dest string, refid string) (acc TrxAccepted, httpStatus int, body string, err error) {
	if err := c.validate(); err != nil {
		return TrxAccepted{}, 0, "", err
	}

	httpStatus, body, err = c.base.TrxRaw(ctx, product, qty, dest, refid)
	if err != nil {
		return TrxAccepted{}, 0, "", err
	}

	rid := ""
	tid := ""
	if m := reR.FindStringSubmatch(body); len(m) >= 2 {
		rid = parseIDDigits(m[1])
	}
	if m := reT.FindStringSubmatch(body); len(m) >= 2 {
		tid = parseIDDigits(m[1])
	}

	price := int64(0)
	if m := rePriceDash.FindStringSubmatch(body); len(m) >= 2 {
		price = parsePriceIDFormat(m[1])
	}

	return TrxAccepted{
		RefID:  rid,
		Ticket: tid,
		Body:   body,
		Price:  price,
	}, httpStatus, body, nil
}

func LooksLikeSystemIssue(body string) bool {
	s := strings.ToUpper(strings.TrimSpace(body))
	if s == "" {
		return true
	}
	if strings.Contains(s, "TIMEOUT") || strings.Contains(s, "GANGGUAN") || strings.Contains(s, "MAINTEN") || strings.Contains(s, "SIBUK") {
		return true
	}
	if !reAccepted.MatchString(s) && !strings.Contains(s, "GAGAL") && !strings.Contains(s, "SUKSES") {
		return true
	}
	return false
}
