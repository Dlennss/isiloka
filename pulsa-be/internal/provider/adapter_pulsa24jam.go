package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const Pulsa24JamProviderName = "pulsa24jam"

type Pulsa24JamAdapter struct {
	BaseURL       string
	MemberID      string
	APIKey        string
	PIN           string
	Password      string
	Secret        string
	CallbackToken string
	Client        *http.Client
}

type Pulsa24JamConfig struct {
	BaseURL       string
	MemberID      string
	APIKey        string
	PIN           string
	Password      string
	Secret        string
	CallbackToken string
	Timeout       time.Duration
}

func NewPulsa24JamAdapter(cfg Pulsa24JamConfig) *Pulsa24JamAdapter {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Pulsa24JamAdapter{
		BaseURL:       strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		MemberID:      strings.TrimSpace(cfg.MemberID),
		APIKey:        strings.TrimSpace(cfg.APIKey),
		PIN:           strings.TrimSpace(cfg.PIN),
		Password:      strings.TrimSpace(cfg.Password),
		Secret:        strings.TrimSpace(cfg.Secret),
		CallbackToken: strings.TrimSpace(cfg.CallbackToken),
		Client:        &http.Client{Timeout: timeout},
	}
}

func (a *Pulsa24JamAdapter) Configured() bool {
	return a != nil && a.BaseURL != "" && a.APIKey != "" && a.PIN != ""
}

func (a *Pulsa24JamAdapter) Name() string { return Pulsa24JamProviderName }

type pulsa24JamPayRequest struct {
	Commands string `json:"commands"`
	Product  string `json:"product,omitempty"`
	Dest     string `json:"dest,omitempty"`
	Qty      int64  `json:"qty,omitempty"`
	RefID    string `json:"refid,omitempty"`
	PIN      string `json:"pin"`
}

type pulsa24JamPayResponse struct {
	OK            bool                `json:"ok"`
	Success       bool                `json:"success"`
	RC            string              `json:"rc"`
	Code          string              `json:"code"`
	Message       string              `json:"message"`
	Msg           string              `json:"msg"`
	Status        string              `json:"status"`
	Command       string              `json:"command"`
	RefID         string              `json:"refid"`
	LegacyRefID   string              `json:"ref_id"`
	ProviderRef   string              `json:"provider_ref"`
	SN            string              `json:"sn"`
	Keterangan    string              `json:"keterangan"`
	Price         int64               `json:"price"`
	Harga         int64               `json:"harga"`
	Balance       int64               `json:"balance"`
	Amount        int64               `json:"amount"`
	ProviderRefID string              `json:"provider_refid"`
	QRURL         string              `json:"qr_url"`
	PaymentType   string              `json:"payment_type"`
	TransactionID string              `json:"transaction_id"`
	FeeAdmin      int64               `json:"fee_admin"`
	GrossAmount   int64               `json:"gross_amount"`
	ExpiredAt     *time.Time          `json:"expired_at"`
	Actions       []map[string]string `json:"actions"`
}

type Pulsa24JamDepositQRISResponse struct {
	RefID         string
	ProviderRefID string
	Amount        int64
	FeeAdmin      int64
	GrossAmount   int64
	Status        string
	PaymentType   string
	TransactionID string
	QRURL         string
	ExpiredAt     *time.Time
	Actions       []map[string]string
	Balance       int64
}

type Pulsa24JamProduct struct {
	ID             int64      `json:"id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"nama"`
	GroupName      string     `json:"group_name"`
	CategoryName   string     `json:"kategori_nama"`
	BrandName      string     `json:"brand_nama"`
	PriceType      string     `json:"tipe_harga"`
	AppBasePrice   *int64     `json:"harga_dasar_app,omitempty"`
	Price          *int64     `json:"harga,omitempty"`
	AdditionalFee  *int64     `json:"fee_tambahan,omitempty"`
	MaximumNominal *int64     `json:"maksimal_nominal,omitempty"`
	Active         bool       `json:"aktif"`
	CreatedAt      *time.Time `json:"dibuat_pada,omitempty"`
	UpdatedAt      *time.Time `json:"diubah_pada,omitempty"`
}

type pulsa24JamProductsResponse struct {
	OK       *bool           `json:"ok"`
	Success  *bool           `json:"success"`
	Commands string          `json:"commands"`
	Command  string          `json:"command"`
	Message  string          `json:"message"`
	Msg      string          `json:"msg"`
	Items    json.RawMessage `json:"items"`
	Products json.RawMessage `json:"products"`
	Data     json.RawMessage `json:"data"`
}

type pulsa24JamProductWire struct {
	ID             int64           `json:"id"`
	SKU            string          `json:"sku"`
	Code           string          `json:"code"`
	Product        string          `json:"product"`
	ProductCode    string          `json:"product_code"`
	KodeProduk     string          `json:"kode_produk"`
	Name           string          `json:"name"`
	Nama           string          `json:"nama"`
	ProductName    string          `json:"product_name"`
	NamaProduk     string          `json:"nama_produk"`
	GroupName      string          `json:"group_name"`
	Group          string          `json:"group"`
	CategoryName   string          `json:"kategori_nama"`
	Category       string          `json:"category"`
	Kategori       string          `json:"kategori"`
	BrandName      string          `json:"brand_nama"`
	Brand          string          `json:"brand"`
	PriceType      string          `json:"tipe_harga"`
	Type           string          `json:"type"`
	AppBasePrice   *int64          `json:"harga_dasar_app"`
	Price          *int64          `json:"price"`
	Harga          *int64          `json:"harga"`
	AdditionalFee  *int64          `json:"fee_tambahan"`
	MaximumNominal *int64          `json:"maksimal_nominal"`
	MaximumAmount  *int64          `json:"maximum_amount"`
	Active         *bool           `json:"aktif"`
	ActiveEnglish  *bool           `json:"active"`
	Status         json.RawMessage `json:"status"`
	CreatedAt      *time.Time      `json:"dibuat_pada"`
	UpdatedAt      *time.Time      `json:"diubah_pada"`
}

// Products meminta katalog H2H memakai command PRODUK. PIN hanya dikirim dari backend.
func (a *Pulsa24JamAdapter) Products(ctx context.Context, product string) ([]Pulsa24JamProduct, error) {
	if !a.Configured() {
		return nil, fmt.Errorf("pulsa24jam credential belum lengkap")
	}
	payload := pulsa24JamPayRequest{
		Commands: "PRODUK",
		Product:  strings.TrimSpace(product),
		PIN:      a.PIN,
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.trxURL(), bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", a.APIKey)

	res, err := a.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}
	var out pulsa24JamProductsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("response produk Pulsa24Jam tidak valid: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices || (out.OK != nil && !*out.OK) || (out.Success != nil && !*out.Success) {
		return nil, fmt.Errorf("produk Pulsa24Jam gagal: %s", firstNonEmpty(out.Message, out.Msg, string(body)))
	}
	wires, err := decodePulsa24JamProductWires(out)
	if err != nil {
		return nil, fmt.Errorf("response produk Pulsa24Jam tidak valid: %w", err)
	}
	requestedSKU := strings.ToUpper(strings.TrimSpace(product))
	items := make([]Pulsa24JamProduct, 0, len(wires))
	for _, wire := range wires {
		item := wire.product()
		if !item.Active {
			continue
		}
		if requestedSKU != "" && strings.ToUpper(strings.TrimSpace(item.SKU)) != requestedSKU {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("katalog Pulsa24Jam tidak berisi produk aktif")
	}
	return items, nil
}

func decodePulsa24JamProductWires(out pulsa24JamProductsResponse) ([]pulsa24JamProductWire, error) {
	for _, raw := range []json.RawMessage{out.Items, out.Products, out.Data} {
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var items []pulsa24JamProductWire
		if err := json.Unmarshal(raw, &items); err == nil {
			return items, nil
		}
		var nested struct {
			Items    []pulsa24JamProductWire `json:"items"`
			Products []pulsa24JamProductWire `json:"products"`
		}
		if err := json.Unmarshal(raw, &nested); err == nil {
			if len(nested.Items) > 0 {
				return nested.Items, nil
			}
			if len(nested.Products) > 0 {
				return nested.Products, nil
			}
		}
	}
	return nil, fmt.Errorf("daftar produk tidak ditemukan")
}

func (w pulsa24JamProductWire) product() Pulsa24JamProduct {
	active := true
	if w.Active != nil {
		active = *w.Active
	} else if w.ActiveEnglish != nil {
		active = *w.ActiveEnglish
	} else if len(w.Status) > 0 {
		var statusString string
		if json.Unmarshal(w.Status, &statusString) == nil {
			status := strings.ToUpper(strings.TrimSpace(statusString))
			active = status == "ACTIVE" || status == "AKTIF" || status == "OPEN" || status == "1" || status == "AVAILABLE"
		} else {
			var statusNumber int
			if json.Unmarshal(w.Status, &statusNumber) == nil {
				active = statusNumber == 1
			}
		}
	}
	return Pulsa24JamProduct{
		ID:             w.ID,
		SKU:            firstNonEmpty(w.SKU, w.Code, w.Product, w.ProductCode, w.KodeProduk),
		Name:           firstNonEmpty(w.Nama, w.Name, w.ProductName, w.NamaProduk),
		GroupName:      firstNonEmpty(w.GroupName, w.Group),
		CategoryName:   firstNonEmpty(w.CategoryName, w.Category, w.Kategori),
		BrandName:      firstNonEmpty(w.BrandName, w.Brand),
		PriceType:      firstNonEmpty(w.PriceType, w.Type),
		AppBasePrice:   w.AppBasePrice,
		Price:          firstInt64Pointer(w.Price, w.Harga),
		AdditionalFee:  w.AdditionalFee,
		MaximumNominal: firstInt64Pointer(w.MaximumNominal, w.MaximumAmount),
		Active:         active,
		CreatedAt:      w.CreatedAt,
		UpdatedAt:      w.UpdatedAt,
	}
}

func firstInt64Pointer(values ...*int64) *int64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func (a *Pulsa24JamAdapter) Pay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	if !a.Configured() {
		return nil, fmt.Errorf("pulsa24jam credential belum lengkap")
	}
	command := strings.ToUpper(strings.TrimSpace(req.Command))
	if command == "" {
		command = "PAY"
	}
	payload := pulsa24JamPayRequest{
		Commands: command,
		Product:  strings.TrimSpace(req.Product),
		Dest:     strings.TrimSpace(req.Dest),
		Qty:      req.Qty,
		RefID:    strings.TrimSpace(req.RefID),
		PIN:      a.PIN,
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.trxURL(), bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", a.APIKey)

	res, err := a.Client.Do(httpReq)
	if err != nil {
		return &PayResponse{
			RequestRaw: map[string]any{"payload": redactPulsa24JamPayload(payload), "url": a.trxURL()},
			Raw:        map[string]any{"error": err.Error()},
		}, err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	var out pulsa24JamPayResponse
	_ = json.Unmarshal(bodyBytes, &out)

	rc := firstNonEmpty(out.RC, out.Code)
	refID := firstNonEmpty(out.RefID, out.LegacyRefID, payload.RefID)
	providerRef := firstNonEmpty(out.ProviderRef, out.SN, refID)
	message := firstNonEmpty(out.Message, out.Msg, out.Keterangan, out.Status, body)
	price := out.Price
	if price <= 0 {
		price = out.Harga
	}

	return &PayResponse{
		HTTPStatus:  res.StatusCode,
		Body:        body,
		RC:          rc,
		Message:     message,
		ProviderRef: providerRef,
		Price:       price,
		Balance:     out.Balance,
		RequestRaw: map[string]any{
			"payload": redactPulsa24JamPayload(payload),
			"url":     a.trxURL(),
		},
		Raw: map[string]any{
			"ok":           out.OK,
			"success":      out.Success,
			"command":      out.Command,
			"status":       out.Status,
			"refid":        refID,
			"provider_ref": providerRef,
			"sn":           out.SN,
			"body":         body,
		},
	}, nil
}

func (a *Pulsa24JamAdapter) CreateDepositQRIS(ctx context.Context, refID string, amount int64) (*Pulsa24JamDepositQRISResponse, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("nominal deposit QRIS harus > 0")
	}
	return a.depositQRIS(ctx, "DEPOSIT-QRIS", refID, amount)
}

func (a *Pulsa24JamAdapter) DepositQRISStatus(ctx context.Context, refID string) (*Pulsa24JamDepositQRISResponse, error) {
	return a.depositQRIS(ctx, "STATUS-DEPOSIT-QRIS", refID, 0)
}

func (a *Pulsa24JamAdapter) depositQRIS(ctx context.Context, command, refID string, amount int64) (*Pulsa24JamDepositQRISResponse, error) {
	if !a.Configured() {
		return nil, fmt.Errorf("pulsa24jam credential belum lengkap")
	}
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, fmt.Errorf("refid deposit QRIS wajib diisi")
	}
	payload := pulsa24JamPayRequest{
		Commands: strings.ToUpper(strings.TrimSpace(command)),
		Qty:      amount,
		RefID:    refID,
		PIN:      a.PIN,
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.trxURL(), bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", a.APIKey)
	res, err := a.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}
	var out pulsa24JamPayResponse
	if err := json.Unmarshal(bodyBytes, &out); err != nil {
		return nil, fmt.Errorf("response deposit QRIS Pulsa24Jam tidak valid: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices || !out.OK {
		msg := firstNonEmpty(out.Message, out.Msg, out.Keterangan, string(bodyBytes))
		return nil, fmt.Errorf("deposit QRIS Pulsa24Jam gagal: %s", msg)
	}
	return &Pulsa24JamDepositQRISResponse{
		RefID:         firstNonEmpty(out.RefID, out.LegacyRefID, refID),
		ProviderRefID: strings.TrimSpace(out.ProviderRefID),
		Amount:        out.Amount,
		FeeAdmin:      out.FeeAdmin,
		GrossAmount:   out.GrossAmount,
		Status:        strings.ToLower(strings.TrimSpace(out.Status)),
		PaymentType:   strings.TrimSpace(out.PaymentType),
		TransactionID: strings.TrimSpace(out.TransactionID),
		QRURL:         strings.TrimSpace(out.QRURL),
		ExpiredAt:     out.ExpiredAt,
		Actions:       out.Actions,
		Balance:       out.Balance,
	}, nil
}

func (a *Pulsa24JamAdapter) trxURL() string {
	base := strings.TrimRight(strings.TrimSpace(a.BaseURL), "/")
	if strings.HasSuffix(base, "/trx") {
		return base
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/trx"
	}
	return base + "/v1/trx"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func redactPulsa24JamPayload(payload pulsa24JamPayRequest) map[string]any {
	out := map[string]any{
		"commands": payload.Commands,
		"product":  payload.Product,
		"dest":     payload.Dest,
		"qty":      payload.Qty,
		"refid":    payload.RefID,
	}
	if payload.PIN != "" {
		out["pin"] = "***"
	}
	return out
}
