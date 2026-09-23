package smb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Mode string

const (
	ModeElektrikFixed Mode = "ELEKTRIK"
	ModePPOB          Mode = "PPOB"
	ModeWalletPPOB    Mode = "WALLET_PPOB"
	ModeDirect        Mode = "DIRECT"
)

type Client struct {
	BaseURL       string
	DirectBaseURL string
	ID            string
	PIN           string
	User          string
	Password      string
	Timeout       time.Duration

	httpc *http.Client
}

type Request struct {
	KodeProduk string
	Tujuan     string
	Qty        int64
	RefID      string
}

type StepResult struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	HTTPStatus int    `json:"http_status"`
	Body       string `json:"body"`
}

type DispatchResult struct {
	Mode        Mode        `json:"mode"`
	Check       *StepResult `json:"check,omitempty"`
	Pay         *StepResult `json:"pay,omitempty"`
	Final       StepResult  `json:"final"`
	StatusCode  string      `json:"status_code,omitempty"`
	ProviderRef string      `json:"provider_ref,omitempty"`
	Price       int64       `json:"price,omitempty"`
	LastBalance *int64      `json:"last_balance,omitempty"`
}

func New(baseURL, directBaseURL, id, pin, user, password string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	directBaseURL = strings.TrimRight(strings.TrimSpace(directBaseURL), "/")
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		BaseURL:       baseURL,
		DirectBaseURL: directBaseURL,
		ID:            strings.TrimSpace(id),
		PIN:           strings.TrimSpace(pin),
		User:          strings.TrimSpace(user),
		Password:      strings.TrimSpace(password),
		Timeout:       timeout,
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
		return errors.New("smb client nil")
	}
	if c.BaseURL == "" {
		return errors.New("smb config missing (baseurl)")
	}
	if c.DirectBaseURL == "" {
		return errors.New("smb config missing (direct_baseurl)")
	}
	if c.ID == "" || c.PIN == "" || c.User == "" || c.Password == "" {
		return errors.New("smb config missing (id/pin/user/password)")
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

func ParseMappedCode(raw string) (Mode, string, error) {
	mode, code, _, err := ParseMappedCodeTarget(raw)
	return mode, code, err
}

func ParseMode(raw string) (Mode, error) {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	if raw == "" {
		return "", errors.New("mode smb kosong")
	}
	switch Mode(raw) {
	case ModeElektrikFixed, ModePPOB, ModeWalletPPOB, ModeDirect:
		return Mode(raw), nil
	default:
		return "", fmt.Errorf("mode smb tidak didukung: %s", raw)
	}
}

func ParseMappedCodeTargetWithMode(modeRaw string, raw string) (Mode, string, string, error) {
	inferredMode, code, bankPrefix, err := ParseMappedCodeTarget(raw)
	if err != nil {
		return "", "", "", err
	}
	modeRaw = strings.TrimSpace(modeRaw)
	if modeRaw == "" {
		return inferredMode, code, bankPrefix, nil
	}
	mode, err := ParseMode(modeRaw)
	if err != nil {
		return "", "", "", err
	}
	return mode, code, bankPrefix, nil
}

func ParseMappedCodeTarget(raw string) (Mode, string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", "", errors.New("kode provider kosong")
	}
	raw = strings.ToUpper(raw)
	if strings.HasPrefix(raw, "BIFASTOPEN:") {
		bankCode := strings.TrimSpace(strings.TrimPrefix(raw, "BIFASTOPEN:"))
		if bankCode == "" {
			return "", "", "", errors.New("kode bank BIFASTOPEN kosong")
		}
		return ModeDirect, "BIFASTOPEN", bankCode, nil
	}
	if strings.HasPrefix(raw, "BIFASTOPEN2:") {
		bankCode := strings.TrimSpace(strings.TrimPrefix(raw, "BIFASTOPEN2:"))
		if bankCode == "" {
			return "", "", "", errors.New("kode bank BIFASTOPEN2 kosong")
		}
		return ModeDirect, "BIFASTOPEN2", bankCode, nil
	}

	// Backward compatibility: support the old prefixed format during transition.
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 {
		mode := Mode(strings.ToUpper(strings.TrimSpace(parts[0])))
		code := strings.ToUpper(strings.TrimSpace(parts[1]))
		if code == "" {
			return "", "", "", errors.New("kode provider smb kosong")
		}
		switch mode {
		case ModeElektrikFixed, ModePPOB, ModeWalletPPOB, ModeDirect:
			return mode, code, "", nil
		default:
			return "", "", "", fmt.Errorf("mode smb tidak didukung: %s", mode)
		}
	}

	switch raw {
	case "ELDN", "GPYOPEN", "SHPOPEN", "ELOV", "ELLI", "OVOOPEN", "BIFASTOPEN", "BIFASTOPEN2":
		return ModeDirect, raw, "", nil
	case "DANA", "OVO", "GOPAY", "SHOPEEPAY", "LINKAJA":
		return ModeWalletPPOB, raw, "", nil
	default:
		// Default plain-code fallback for regular PPOB codes.
		return ModePPOB, raw, "", nil
	}
}

func (c *Client) Dispatch(ctx context.Context, mode Mode, req Request, command string) (*DispatchResult, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	command = strings.ToUpper(strings.TrimSpace(command))
	if command != "PAY" && command != "INQ" {
		return nil, fmt.Errorf("smb hanya mendukung PAY/INQ, dapat=%s", command)
	}
	req.KodeProduk = strings.ToUpper(strings.TrimSpace(req.KodeProduk))
	req.Tujuan = strings.TrimSpace(req.Tujuan)
	req.RefID = strings.TrimSpace(req.RefID)
	if req.KodeProduk == "" || req.Tujuan == "" || req.RefID == "" {
		return nil, errors.New("smb request tidak lengkap")
	}
	if req.Qty <= 0 {
		return nil, errors.New("smb qty harus > 0")
	}

	out := &DispatchResult{Mode: mode}
	switch mode {
	case ModeElektrikFixed:
		step, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, req.Tujuan, req.RefID, 0, nil))
		if err != nil {
			return nil, err
		}
		step.Kind = "pay"
		out.Pay = step
		out.Final = *step
	case ModeDirect:
		amount := req.Qty
		step, err := c.call(ctx, c.DirectBaseURL, buildParams(c, req.KodeProduk, req.Tujuan, req.RefID, 1, &amount))
		if err != nil {
			return nil, err
		}
		step.Kind = "pay"
		out.Pay = step
		out.Final = *step
	case ModePPOB:
		// Bank PPOB: tujuan harus format qty@norek
		if !strings.Contains(req.Tujuan, "@") && req.Qty > 0 {
			req.Tujuan = fmt.Sprintf("%d@%s", req.Qty, req.Tujuan)
		}
		check, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, req.Tujuan, withPrefix(req.RefID, "CEK"), 5, nil))
		if err != nil {
			return nil, err
		}
		check.Kind = "check"
		out.Check = check
		if command == "INQ" || LooksLikeImmediateReject(check.Body) {
			out.Final = *check
			break
		}
		pay, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, req.Tujuan, withPrefix(req.RefID, "BYR"), 6, nil))
		if err != nil {
			return nil, err
		}
		pay.Kind = "pay"
		out.Pay = pay
		out.Final = *pay
	case ModeWalletPPOB:
		target := fmt.Sprintf("%d@%s", req.Qty, req.Tujuan)
		check, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, target, withPrefix(req.RefID, "CEK"), 5, nil))
		if err != nil {
			return nil, err
		}
		check.Kind = "check"
		out.Check = check
		if command == "INQ" || LooksLikeImmediateReject(check.Body) {
			out.Final = *check
			break
		}
		pay, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, target, withPrefix(req.RefID, "BYR"), 6, nil))
		if err != nil {
			return nil, err
		}
		pay.Kind = "pay"
		out.Pay = pay
		out.Final = *pay
	default:
		return nil, fmt.Errorf("mode smb tidak didukung: %s", mode)
	}

	out.StatusCode = ExtractStatusCode(out.Final.Body)
	out.ProviderRef = ParseProviderRef(out.Final.Body)
	out.Price = ParsePrice(out.Final.Body)
	if bal, ok := ParseLastBalance(out.Final.Body); ok && bal > 0 {
		out.LastBalance = &bal
	} else if out.Pay != nil {
		if bal, ok := ParseLastBalance(out.Pay.Body); ok && bal > 0 {
			out.LastBalance = &bal
		}
	}
	if out.LastBalance == nil && out.Check != nil {
		if bal, ok := ParseLastBalance(out.Check.Body); ok && bal > 0 {
			out.LastBalance = &bal
		}
	}
	if out.LastBalance == nil {
		if bal, ok := ParseLastBalance(out.Final.Body); ok && bal > 0 {
			out.LastBalance = &bal
		}
	}
	return out, nil
}

func (c *Client) DispatchPayOnly(ctx context.Context, mode Mode, req Request) (*DispatchResult, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	req.KodeProduk = strings.ToUpper(strings.TrimSpace(req.KodeProduk))
	req.Tujuan = strings.TrimSpace(req.Tujuan)
	req.RefID = strings.TrimSpace(req.RefID)
	if req.KodeProduk == "" || req.Tujuan == "" || req.RefID == "" {
		return nil, errors.New("smb request tidak lengkap")
	}
	if req.Qty <= 0 {
		return nil, errors.New("smb qty harus > 0")
	}

	// Bank PPOB: tujuan harus format qty@norek
	if mode == ModePPOB && !strings.Contains(req.Tujuan, "@") && req.Qty > 0 {
		req.Tujuan = fmt.Sprintf("%d@%s", req.Qty, req.Tujuan)
	}

	out := &DispatchResult{Mode: mode}
	switch mode {
	case ModeElektrikFixed:
		step, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, req.Tujuan, req.RefID, 0, nil))
		if err != nil {
			return nil, err
		}
		step.Kind = "pay"
		out.Pay = step
		out.Final = *step
	case ModeDirect:
		amount := req.Qty
		step, err := c.call(ctx, c.DirectBaseURL, buildParams(c, req.KodeProduk, req.Tujuan, req.RefID, 1, &amount))
		if err != nil {
			return nil, err
		}
		step.Kind = "pay"
		out.Pay = step
		out.Final = *step
	case ModePPOB:
		pay, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, req.Tujuan, withPrefix(req.RefID, "BYR"), 6, nil))
		if err != nil {
			return nil, err
		}
		pay.Kind = "pay"
		out.Pay = pay
		out.Final = *pay
	case ModeWalletPPOB:
		target := fmt.Sprintf("%d@%s", req.Qty, req.Tujuan)
		pay, err := c.call(ctx, c.BaseURL, buildParams(c, req.KodeProduk, target, withPrefix(req.RefID, "BYR"), 6, nil))
		if err != nil {
			return nil, err
		}
		pay.Kind = "pay"
		out.Pay = pay
		out.Final = *pay
	default:
		return nil, fmt.Errorf("mode smb tidak didukung: %s", mode)
	}

	out.StatusCode = ExtractStatusCode(out.Final.Body)
	out.ProviderRef = ParseProviderRef(out.Final.Body)
	out.Price = ParsePrice(out.Final.Body)
	if bal, ok := ParseLastBalance(out.Final.Body); ok && bal > 0 {
		out.LastBalance = &bal
	}
	return out, nil
}

func buildParams(c *Client, kodeProduk, tujuan, idtrx string, jenis int, amount *int64) url.Values {
	q := url.Values{}
	q.Set("id", c.ID)
	q.Set("pin", c.PIN)
	q.Set("user", c.User)
	q.Set("pass", c.Password)
	q.Set("kodeproduk", kodeProduk)
	q.Set("tujuan", tujuan)
	q.Set("idtrx", idtrx)
	q.Set("counter", "1")
	q.Set("jenis", strconv.Itoa(jenis))
	if amount != nil && *amount > 0 {
		q.Set("amount", strconv.FormatInt(*amount, 10))
	}
	return q
}

func (c *Client) call(ctx context.Context, baseURL string, q url.Values) (*StepResult, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/api/h2h")
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
		URL:        u.String(),
		HTTPStatus: resp.StatusCode,
		Body:       strings.TrimSpace(string(b)),
	}, nil
}

func withPrefix(refID, prefix string) string {
	refID = strings.TrimSpace(refID)
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	if strings.HasPrefix(strings.ToUpper(refID), prefix) {
		return refID
	}
	return prefix + refID
}

var (
	reStatus    = regexp.MustCompile(`(?i)\b(?:status|rc|kode_respon|kode_response|code)\s*[:=]?\s*([0-9]{1,3})\b`)
	reRefDigits = regexp.MustCompile(`(?i)\b(?:R#|REFID[:=# ]|IDTRX[:=# ])\s*([A-Z0-9_-]{4,})\b`)
	reTicket    = regexp.MustCompile(`(?i)\b(?:T#|SN[:=# ]|NOREF[:=# ]|NOREFF[:=# ]|REF[:=# ])\s*([A-Z0-9_./-]{4,})`)
	rePrice     = regexp.MustCompile(`(?i)\b(?:harga|price)\s*[:=]?\s*([0-9][0-9\.,]*)`)
	reJumlah    = regexp.MustCompile(`(?i)\b(?:jumlah|total\s*tag|total)\s*[:=]?\s*([0-9][0-9\.,]*)`)
	reAmount    = regexp.MustCompile(`(?i)\b(?:nominal|tagihan|amount)\s*[:=]?\s*([0-9][0-9\.,]*)`)
	reSaldoEq   = regexp.MustCompile(`(?i)\b(?:sisa\s*saldo|sisasaldo|saldo\s*akhir|saldo|balance)\b[^=\r\n]*=\s*([0-9][0-9\.,]*)`)
	reSaldo     = regexp.MustCompile(`(?i)\b(?:sisa\s*saldo|sisasaldo|saldo\s*akhir|saldo|balance)\b[^0-9]*([0-9][0-9\.,]*)`)
)

func parseJSONBody(body string) map[string]any {
	body = strings.TrimSpace(body)
	if body == "" || !strings.HasPrefix(body, "{") {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		return nil
	}
	return out
}

func jsonStringField(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if m == nil {
			return ""
		}
		if v, ok := m[key]; ok {
			switch x := v.(type) {
			case string:
				return strings.TrimSpace(x)
			case float64:
				return strconv.FormatInt(int64(x), 10)
			case json.Number:
				return strings.TrimSpace(x.String())
			}
		}
	}
	return ""
}

func ExtractStatusCode(body string) string {
	if m := parseJSONBody(body); m != nil {
		if rc := jsonStringField(m, "rc", "status", "kode_respon"); rc != "" {
			return normalizeStatusCode(rc)
		}
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if m := reStatus.FindStringSubmatch(body); len(m) == 2 {
		return normalizeStatusCode(m[1])
	}
	return ""
}

func normalizeStatusCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimLeft(raw, "0")
	if raw == "" {
		return "0"
	}
	return raw
}

func ParseRefID(body string) string {
	if m := parseJSONBody(body); m != nil {
		refID := jsonStringField(m, "reffid", "refid", "idtrx")
		refID = strings.TrimPrefix(strings.TrimPrefix(refID, "CEK"), "BYR")
		if refID != "" {
			return refID
		}
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if m := reRefDigits.FindStringSubmatch(strings.ToUpper(body)); len(m) == 2 {
		refID := strings.TrimSpace(m[1])
		refID = strings.TrimPrefix(refID, "CEK")
		refID = strings.TrimPrefix(refID, "BYR")
		return refID
	}
	return ""
}

func ParseProviderRef(body string) string {
	if m := parseJSONBody(body); m != nil {
		if ref := jsonStringField(m, "sn", "noreff", "reff", "ref"); ref != "" {
			return strings.Trim(ref, ".,;:")
		}
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if m := reTicket.FindStringSubmatch(strings.ToUpper(body)); len(m) == 2 {
		return strings.Trim(strings.TrimSpace(m[1]), ".,;:")
	}
	return ""
}

func ParsePrice(body string) int64 {
	if m := parseJSONBody(body); m != nil {
		if msg := jsonStringField(m, "msg", "message"); msg != "" {
			if parsed := ParsePrice(msg); parsed > 0 {
				return parsed
			}
		}
		if raw := jsonStringField(m, "harga", "price"); raw != "" {
			return parseAmount(raw)
		}
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return 0
	}
	if m := rePrice.FindStringSubmatch(body); len(m) == 2 {
		return parseAmount(m[1])
	}
	if m := reAmount.FindStringSubmatch(body); len(m) == 2 {
		return parseAmount(m[1])
	}
	return 0
}

func ParsePayableAmount(body string) int64 {
	if m := parseJSONBody(body); m != nil {
		if msg := jsonStringField(m, "msg", "message"); msg != "" {
			if parsed := ParsePayableAmount(msg); parsed > 0 {
				return parsed
			}
		}
		if raw := jsonStringField(m, "jumlah", "total", "total_tag"); raw != "" {
			return parseAmount(raw)
		}
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return 0
	}
	if m := reJumlah.FindStringSubmatch(body); len(m) == 2 {
		return parseAmount(m[1])
	}
	return 0
}

func ParseLastBalance(body string) (int64, bool) {
	if m := parseJSONBody(body); m != nil {
		if msg := jsonStringField(m, "msg", "message"); msg != "" {
			if parsed, ok := ParseLastBalance(msg); ok {
				return parsed, true
			}
		}
		if raw := jsonStringField(m, "saldo", "balance", "last_balance"); raw != "" {
			n := parseAmount(raw)
			return n, n > 0
		}
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return 0, false
	}
	if m := reSaldoEq.FindStringSubmatch(body); len(m) == 2 {
		n := parseAmount(m[1])
		return n, n > 0
	}
	if m := reSaldo.FindStringSubmatch(body); len(m) == 2 {
		n := parseAmount(m[1])
		return n, n > 0
	}
	return 0, false
}

func parseAmount(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func LooksLikeSuccess(body string) bool {
	if m := parseJSONBody(body); m != nil {
		rc := normalizeStatusCode(jsonStringField(m, "rc", "status"))
		msg := strings.ToUpper(jsonStringField(m, "msg", "message"))
		if rc == "20" || rc == "1" {
			if strings.Contains(msg, "PENDING") || strings.Contains(msg, "PROSES") || strings.Contains(msg, "MENUNGGU") {
				return false
			}
			return true
		}
		if okv, exists := m["success"]; exists {
			if okb, ok := okv.(bool); ok && okb && rc == "" && !strings.Contains(msg, "PENDING") && !strings.Contains(msg, "PROSES") && !strings.Contains(msg, "MENUNGGU") {
				return true
			}
		}
	}
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	rc := ExtractStatusCode(body)
	return rc == "20" || rc == "1" || strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL") || strings.Contains(up, "LUNAS")
}

func LooksLikePending(body string) bool {
	if m := parseJSONBody(body); m != nil {
		rc := normalizeStatusCode(jsonStringField(m, "rc", "status"))
		msg := strings.ToUpper(jsonStringField(m, "msg", "message"))
		if smbHasExplicitFailureSignal(msg) {
			return false
		}
		if rc == "0" || rc == "2" || rc == "3" || rc == "27" || rc == "68" {
			return true
		}
		if strings.Contains(msg, "ANTRIAN") || strings.Contains(msg, "ANTRI") {
			return true
		}
	}
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	if smbHasExplicitFailureSignal(up) {
		return false
	}
	rc := ExtractStatusCode(body)
	return rc == "0" || rc == "2" || rc == "3" || rc == "27" || rc == "68" ||
		strings.Contains(up, "PROSES") ||
		strings.Contains(up, "MENUNGGU") ||
		strings.Contains(up, "ANTRIAN") ||
		strings.Contains(up, "ANTRI") ||
		strings.Contains(up, "PENDING")
}

func LooksLikeImmediateReject(body string) bool {
	if m := parseJSONBody(body); m != nil {
		rc := normalizeStatusCode(jsonStringField(m, "rc", "status"))
		msg := strings.ToUpper(jsonStringField(m, "msg", "message"))
		if rc != "" && rc != "20" && rc != "1" && rc != "0" && rc != "2" && rc != "3" && rc != "27" && rc != "68" {
			return true
		}
		if okv, exists := m["success"]; exists {
			if okb, ok := okv.(bool); ok && !okb {
				return true
			}
		}
		if strings.Contains(msg, "INVALID CREDENTIAL") {
			return true
		}
		if strings.Contains(msg, "CEK BALANCE") ||
			strings.Contains(msg, "CS/ADMIN") ||
			strings.Contains(msg, "SALDO PROVIDER KURANG") {
			return true
		}
		if smbHasExplicitFailureSignal(msg) {
			return true
		}
	}
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	if LooksLikeSuccess(body) {
		return false
	}
	if smbHasExplicitFailureSignal(up) {
		return true
	}
	if LooksLikePending(body) {
		return false
	}
	rc := ExtractStatusCode(body)
	if rc != "" && rc != "20" && rc != "1" && rc != "0" && rc != "2" && rc != "3" && rc != "27" && rc != "68" {
		return true
	}
	return strings.Contains(up, "GAGAL") ||
		strings.Contains(up, "FAILED") ||
		strings.Contains(up, "ERROR") ||
		strings.Contains(up, "TIDAK") ||
		strings.Contains(up, "BATAL") ||
		strings.Contains(up, "CEK BALANCE") ||
		strings.Contains(up, "CS/ADMIN") ||
		strings.Contains(up, "SALDO PROVIDER KURANG") ||
		(strings.Contains(up, "SALDO") && strings.Contains(up, "CUKUP"))
}

func LooksLikeAccepted(body string) bool {
	return LooksLikeSuccess(body) || LooksLikePending(body)
}

func LooksLikeSystemIssue(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return true
	}
	if strings.Contains(up, "TIMEOUT") || strings.Contains(up, "GANGGUAN") || strings.Contains(up, "MAINTENANCE") || strings.Contains(up, "MAINTEN") {
		return true
	}
	return !LooksLikeAccepted(body) && !LooksLikeImmediateReject(body)
}

func smbHasExplicitFailureSignal(msg string) bool {
	msg = strings.ToUpper(strings.TrimSpace(msg))
	if msg == "" {
		return false
	}
	return strings.Contains(msg, "GAGAL") ||
		strings.Contains(msg, "FAILED") ||
		strings.Contains(msg, "ERROR") ||
		strings.Contains(msg, "TIDAK") ||
		strings.Contains(msg, "BATAL")
}
