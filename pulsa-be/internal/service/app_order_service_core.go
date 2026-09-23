package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

type AppOrderService struct {
	orderRepo        *repository.AppOrderRepository
	paymentRepo      *repository.AppOrderPaymentRepository
	produkRepo       *repository.ProdukRepository
	pricingRepo      *repository.ProdukAppPricingRepository
	kategoriFeeRepo  *repository.KategoriFeeAppRepository
	appProviderRepo  *repository.AppOrderProviderTrxRepository
	billingCheckRepo *repository.AppBillingCheckRepository
	pulsa24JamClient *provider.Pulsa24JamAdapter
}

func (s *AppOrderService) SetPulsa24JamClient(client *provider.Pulsa24JamAdapter) {
	s.pulsa24JamClient = client
}

func (s *AppOrderService) validatePulsa24JamProduct(ctx context.Context, productCode string) (*provider.Pulsa24JamProduct, error) {
	if s.pulsa24JamClient == nil || !s.pulsa24JamClient.Configured() {
		return nil, nil
	}
	items, err := s.pulsa24JamClient.Products(ctx, productCode)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa produk Pulsa24Jam: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("produk tidak tersedia di Pulsa24Jam")
	}
	item := items[0]
	if !item.Active {
		return nil, fmt.Errorf("produk sedang tidak aktif di Pulsa24Jam")
	}
	return &item, nil
}

const appOrderPaymentFeeBps int64 = 7

func NewAppOrderService(
	orderRepo *repository.AppOrderRepository,
	paymentRepo *repository.AppOrderPaymentRepository,
	produkRepo *repository.ProdukRepository,
	pricingRepo *repository.ProdukAppPricingRepository,
	kategoriFeeRepo *repository.KategoriFeeAppRepository,
	appProviderRepo *repository.AppOrderProviderTrxRepository,
	billingCheckRepo *repository.AppBillingCheckRepository,
) *AppOrderService {
	return &AppOrderService{
		orderRepo:        orderRepo,
		paymentRepo:      paymentRepo,
		produkRepo:       produkRepo,
		pricingRepo:      pricingRepo,
		kategoriFeeRepo:  kategoriFeeRepo,
		appProviderRepo:  appProviderRepo,
		billingCheckRepo: billingCheckRepo,
	}
}
func computeAppOrderPaymentFee(subtotal int64) int64 {
	if subtotal <= 0 {
		return 0
	}
	return (subtotal*appOrderPaymentFeeBps + 999) / 1000
}

func normalizeGuestEmail(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizeGuestPhone(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
func mapBillingInquiry(parsed *helper.AppBillingInquiryParsed) *repository.AppBillingInquiryInfo {
	if parsed == nil {
		return nil
	}
	out := &repository.AppBillingInquiryInfo{
		Provider:        parsed.Provider,
		ProviderStatus:  parsed.ProviderStatus,
		ProviderMessage: parsed.ProviderMessage,
		DisplayMessage:  parsed.DisplayMessage,
		BillAmount:      parsed.BillAmount,
		PenaltyAmount:   parsed.PenaltyAmount,
		AdminAmount:     parsed.AdminAmount,
		TotalAmount:     parsed.TotalAmount,
		CanPay:          parsed.CanPay,
	}
	if strings.TrimSpace(parsed.CustomerName) != "" {
		v := strings.TrimSpace(parsed.CustomerName)
		out.CustomerName = &v
	}
	if strings.TrimSpace(parsed.UsageLabel) != "" {
		v := strings.TrimSpace(parsed.UsageLabel)
		out.UsageLabel = &v
	}
	if strings.TrimSpace(parsed.MeterType) != "" {
		v := strings.TrimSpace(parsed.MeterType)
		out.MeterType = &v
	}
	if strings.TrimSpace(parsed.PeriodLabel) != "" {
		v := strings.TrimSpace(parsed.PeriodLabel)
		out.PeriodLabel = &v
	}
	if strings.TrimSpace(parsed.MeterRange) != "" {
		v := strings.TrimSpace(parsed.MeterRange)
		out.MeterRange = &v
	}
	if strings.TrimSpace(parsed.TransactionTime) != "" {
		v := strings.TrimSpace(parsed.TransactionTime)
		out.TransactionTime = &v
	}
	return out
}

func looksLikeCheckProduct(sku, nama string) bool {
	upperSKU := strings.TrimSpace(strings.ToUpper(sku))
	upperName := strings.TrimSpace(strings.ToUpper(nama))
	return strings.HasPrefix(upperSKU, "CEK") || strings.HasSuffix(upperSKU, "C") || strings.Contains(upperName, "CEK ")
}

func valueOrEmptyString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func normalizeBuyer(buyerType string, memberID *int64) (string, *int64, error) {
	buyerType = strings.TrimSpace(strings.ToLower(buyerType))
	switch buyerType {
	case "user":
		if memberID == nil || *memberID <= 0 {
			return "", nil, fmt.Errorf("user tidak valid")
		}
		return "user", memberID, nil
	case "guest", "":
		return "guest", nil, nil
	default:
		return "", nil, fmt.Errorf("buyer_type tidak valid")
	}
}

func resolveOrderNominal(produk *repository.ProdukRow, qty int64, hargaDasar int64, isCheckProduct bool) (nominal int64, qtyFinal int64, err error) {
	switch strings.ToUpper(strings.TrimSpace(produk.TipeHarga)) {
	case "FIXED":
		if qty != 1 {
			return 0, 0, fmt.Errorf("qty untuk produk FIXED harus 1")
		}
		if isCheckProduct {
			return 0, 1, nil
		}
		if hargaDasar <= 0 {
			return 0, 0, fmt.Errorf("harga dasar produk FIXED belum valid")
		}
		return hargaDasar, 1, nil
	case "OPEN_AMOUNT":
		if qty <= 0 {
			return 0, 0, fmt.Errorf("qty untuk produk OPEN_AMOUNT harus > 0")
		}
		if produk.MaksimalNominal != nil && *produk.MaksimalNominal > 0 && qty > *produk.MaksimalNominal {
			return 0, 0, fmt.Errorf("qty melebihi maksimal nominal (%d)", *produk.MaksimalNominal)
		}
		return qty, qty, nil
	default:
		return 0, 0, fmt.Errorf("tipe_harga produk tidak didukung")
	}
}

func isAppCheckProduct(produk *repository.ProdukRow) bool {
	if produk == nil {
		return false
	}
	sku := strings.ToUpper(strings.TrimSpace(produk.SKU))
	nama := strings.ToUpper(strings.TrimSpace(produk.Nama))
	if strings.HasPrefix(sku, "CEK") {
		return true
	}
	if sku == "PLNC" {
		return true
	}
	if strings.Contains(nama, "CEK ") || strings.HasPrefix(nama, "CEK") {
		return true
	}
	return false
}

func buildAppOrderInvoiceID() string {
	return fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102150405"), strings.ToUpper(helper.RandHex(4)))
}
