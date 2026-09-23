package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *AppOrderService) Create(ctx context.Context, in repository.AppOrderCreateInput) (*repository.AppOrderRow, error) {
	buyerType, memberID, err := normalizeBuyer(in.BuyerType, in.MemberID)
	if err != nil {
		return nil, err
	}

	in.Dest = strings.TrimSpace(in.Dest)
	if in.Dest == "" {
		return nil, fmt.Errorf("dest wajib diisi")
	}

	produk, err := s.produkRepo.Get(ctx, in.ProdukID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("produk tidak ditemukan")
		}
		return nil, err
	}
	if !produk.Aktif {
		return nil, fmt.Errorf("produk tidak aktif")
	}

	// Cek jam online produk dari database
	if !helper.IsProductAvailableNow(produk.JamBuka, produk.JamTutup) {
		return nil, fmt.Errorf("produk ini tidak tersedia pada jam ini, silakan coba lagi nanti")
	}

	pricingRow, err := s.pricingRepo.GetEffectiveByProdukIDActive(ctx, in.ProdukID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("produk tidak tersedia untuk app commerce")
		}
		return nil, err
	}
	if !pricingRow.Aktif {
		return nil, fmt.Errorf("produk tidak aktif")
	}
	if !strings.EqualFold(strings.TrimSpace(pricingRow.Provider), "pulsa24jam") {
		return nil, fmt.Errorf("produk retail wajib memakai provider Pulsa24Jam")
	}
	providerCode := strings.TrimSpace(pricingRow.YuscomSKU)
	if providerCode == "" {
		providerCode = strings.TrimSpace(produk.SKU)
	}
	liveProduct, err := s.validatePulsa24JamProduct(ctx, providerCode)
	if err != nil {
		return nil, err
	}

	feeKategoriID := produk.KategoriID
	feeRow, err := s.kategoriFeeRepo.GetByKategoriIDActive(ctx, produk.KategoriID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("fee kategori belum tersedia")
		}
		return nil, err
	}

	if strings.ToUpper(strings.TrimSpace(produk.TipeHarga)) == "OPEN_AMOUNT" {
		openAmountFeeRow, openErr := s.kategoriFeeRepo.GetByKategoriNameActive(ctx, repository.OpenAmountFeeKategoriName)
		if openErr == nil && openAmountFeeRow != nil {
			feeRow = openAmountFeeRow
			feeKategoriID = openAmountFeeRow.KategoriID
		} else if openErr != nil && openErr != sql.ErrNoRows {
			return nil, openErr
		}
	}

	qty := in.Qty
	if qty <= 0 {
		qty = 1
	}

	hargaDasar := pricingRow.Harga
	if liveProduct != nil {
		if liveProduct.AppBasePrice != nil {
			hargaDasar = *liveProduct.AppBasePrice
		} else if liveProduct.Price != nil {
			hargaDasar = *liveProduct.Price
		} else if liveProduct.AdditionalFee != nil {
			hargaDasar = *liveProduct.AdditionalFee
		}
	}
	isCheckProduct := isAppCheckProduct(produk)
	billingAmount, err := s.resolveBillingAmountFromSourceCheck(ctx, produk, buyerType, memberID, in)
	if err != nil {
		return nil, err
	}
	var nominal int64
	var qtyFinal int64
	if billingAmount > 0 {
		nominal = billingAmount
		qtyFinal = 1
	} else {
		nominal, qtyFinal, err = resolveOrderNominal(produk, qty, hargaDasar, isCheckProduct)
		if err != nil {
			return nil, err
		}
	}

	buyerRole := helper.NormalizeRole(in.BuyerRole)
	var retailCtx *repository.RetailMemberContextRow
	fee := feeRow.FeeNonUser
	if buyerType == "user" {
		if memberID == nil || *memberID <= 0 {
			return nil, fmt.Errorf("user tidak valid")
		}
		retailCtx, err = s.orderRepo.GetRetailMemberContext(ctx, *memberID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("user tidak ditemukan")
			}
			return nil, err
		}
		if buyerRole == "" {
			buyerRole = helper.NormalizeRole(retailCtx.Role)
		}
		switch buyerRole {
		case helper.RoleRetailMaster:
			fee = feeRow.FeeMaster
		case helper.RoleRetailAgent:
			fee = feeRow.FeeAgent
		default:
			fee = feeRow.FeeUser
		}
	}

	if isCheckProduct || (billingAmount <= 0 && strings.ToUpper(strings.TrimSpace(produk.TipeHarga)) == "FIXED" && hargaDasar <= 0) {
		fee = 0
	}

	if buyerType == "guest" {
		in.GuestNama = strings.TrimSpace(in.GuestNama)
		in.GuestEmail = normalizeGuestEmail(in.GuestEmail)
		in.GuestPhone = normalizeGuestPhone(in.GuestPhone)
		if in.GuestNama == "" || in.GuestEmail == "" || in.GuestPhone == "" {
			return nil, fmt.Errorf("guest_nama/guest_email/guest_phone wajib diisi")
		}
		if !strings.Contains(in.GuestEmail, "@") || strings.HasPrefix(in.GuestEmail, "@") || strings.HasSuffix(in.GuestEmail, "@") {
			return nil, fmt.Errorf("guest_email tidak valid")
		}
		if len(in.GuestPhone) < 8 {
			return nil, fmt.Errorf("guest_phone tidak valid")
		}
	}

	hargaDasarFinal := hargaDasar
	if isCheckProduct {
		hargaDasarFinal = 0
	}
	if strings.ToUpper(strings.TrimSpace(produk.TipeHarga)) == "OPEN_AMOUNT" {
		hargaDasarFinal = nominal + hargaDasar
	} else if billingAmount > 0 {
		hargaDasarFinal = nominal + hargaDasar
	}

	subtotal := hargaDasarFinal + fee
	paymentFee := computeAppOrderPaymentFee(subtotal)
	if buyerType == "user" && memberID != nil && *memberID > 0 && subtotal > 0 && s.paymentRepo != nil {
		if wallet, _, saldoErr := s.paymentRepo.GetMemberFundingBalance(ctx, *memberID); saldoErr == nil && wallet >= subtotal {
			paymentFee = 0
		}
	}
	totalFee := fee + paymentFee

	invoiceID := buildAppOrderInvoiceID()
	createIn := repository.AppOrderCreateInput{
		InvoiceID:          invoiceID,
		MemberID:           memberID,
		GuestNama:          strings.TrimSpace(in.GuestNama),
		GuestEmail:         normalizeGuestEmail(in.GuestEmail),
		GuestPhone:         normalizeGuestPhone(in.GuestPhone),
		ProdukID:           produk.ID,
		ProdukSKUSnapshot:  produk.SKU,
		ProdukNamaSnapshot: produk.Nama,
		Dest:               in.Dest,
		Qty:                qtyFinal,
		Nominal:            nominal,
		BuyerType:          buyerType,
		BuyerRole:          buyerRole,
		HargaDasar:         hargaDasarFinal,
		Fee:                totalFee,
		HargaFinal:         subtotal + paymentFee,
		FeeUserSnapshot:    feeRow.FeeUser,
		FeeAgentSnapshot:   feeRow.FeeAgent,
		FeeMasterSnapshot:  feeRow.FeeMaster,
		Status:             "pending_payment",
	}
	if retailCtx != nil {
		createIn.RetailAgentIDSnapshot = retailCtx.RetailAgentID
		createIn.RetailMasterIDSnapshot = retailCtx.RetailMasterID
	}
	_ = feeKategoriID
	if err := s.orderRepo.Create(ctx, createIn); err != nil {
		return nil, err
	}
	return s.orderRepo.GetByInvoiceID(ctx, invoiceID)
}
func (s *AppOrderService) resolveBillingAmountFromSourceCheck(ctx context.Context, produk *repository.ProdukRow, buyerType string, memberID *int64, in repository.AppOrderCreateInput) (int64, error) {
	sourceCheckRefID := strings.TrimSpace(in.SourceCheckRefID)
	if sourceCheckRefID != "" {
		if produk == nil {
			return 0, fmt.Errorf("produk tidak valid")
		}
		if isAppCheckProduct(produk) {
			return 0, fmt.Errorf("produk cek tidak dapat memakai source_check_ref_id")
		}
		if s.billingCheckRepo == nil {
			return 0, fmt.Errorf("service billing check belum siap")
		}
		sourceCheck, err := s.billingCheckRepo.GetByRefID(ctx, sourceCheckRefID)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("ref cek tagihan tidak ditemukan")
			}
			return 0, err
		}
		if strings.TrimSpace(sourceCheck.Dest) != strings.TrimSpace(in.Dest) {
			return 0, fmt.Errorf("tujuan pembayaran tidak cocok dengan hasil cek")
		}
		switch buyerType {
		case "user":
			if memberID == nil || sourceCheck.MemberID == nil || *memberID != *sourceCheck.MemberID {
				return 0, fmt.Errorf("hasil cek tidak dimiliki user ini")
			}
		case "guest":
			if normalizeGuestEmail(in.GuestEmail) == "" || normalizeGuestPhone(in.GuestPhone) == "" {
				return 0, fmt.Errorf("data guest untuk hasil cek tidak valid")
			}
			if normalizeGuestEmail(in.GuestEmail) != normalizeGuestEmail(valueOrEmptyString(sourceCheck.GuestEmail)) ||
				normalizeGuestPhone(in.GuestPhone) != normalizeGuestPhone(valueOrEmptyString(sourceCheck.GuestPhone)) {
				return 0, fmt.Errorf("hasil cek tidak cocok dengan data guest ini")
			}
		default:
			return 0, fmt.Errorf("buyer_type tidak valid")
		}
		parsed := parseBillingInquiryFromCheck(sourceCheck)
		if parsed == nil || !parsed.CanPay || parsed.TotalAmount <= 0 {
			return 0, fmt.Errorf("hasil cek tagihan belum memiliki total bayar yang valid")
		}
		return parsed.TotalAmount, nil
	}

	sourceInvoiceID := strings.TrimSpace(in.SourceCheckInvoiceID)
	if sourceInvoiceID == "" {
		return 0, nil
	}
	if produk == nil {
		return 0, fmt.Errorf("produk tidak valid")
	}
	if isAppCheckProduct(produk) {
		return 0, fmt.Errorf("produk cek tidak dapat memakai source_check_invoice_id")
	}

	sourceOrder, err := s.orderRepo.GetByInvoiceID(ctx, sourceInvoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invoice cek tidak ditemukan")
		}
		return 0, err
	}
	if !looksLikeCheckProduct(sourceOrder.ProdukSKUSnapshot, sourceOrder.ProdukNamaSnapshot) {
		return 0, fmt.Errorf("invoice sumber bukan hasil cek tagihan")
	}
	if strings.TrimSpace(strings.ToLower(sourceOrder.Status)) != "success" {
		return 0, fmt.Errorf("hasil cek tagihan belum siap dibayar")
	}
	if strings.TrimSpace(sourceOrder.Dest) != strings.TrimSpace(in.Dest) {
		return 0, fmt.Errorf("tujuan pembayaran tidak cocok dengan hasil cek")
	}

	switch buyerType {
	case "user":
		if memberID == nil || sourceOrder.MemberID == nil || *memberID != *sourceOrder.MemberID {
			return 0, fmt.Errorf("hasil cek tidak dimiliki user ini")
		}
	case "guest":
		if normalizeGuestEmail(in.GuestEmail) == "" || normalizeGuestPhone(in.GuestPhone) == "" {
			return 0, fmt.Errorf("data guest untuk hasil cek tidak valid")
		}
		if normalizeGuestEmail(in.GuestEmail) != normalizeGuestEmail(valueOrEmptyString(sourceOrder.GuestEmail)) ||
			normalizeGuestPhone(in.GuestPhone) != normalizeGuestPhone(valueOrEmptyString(sourceOrder.GuestPhone)) {
			return 0, fmt.Errorf("hasil cek tidak cocok dengan data guest ini")
		}
	default:
		return 0, fmt.Errorf("buyer_type tidak valid")
	}

	providerRow, err := s.appProviderRepo.GetLatestByAppOrderID(ctx, sourceOrder.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("hasil cek provider belum tersedia")
		}
		return 0, err
	}
	parsed := helper.ParseAppBillingInquiry(providerRow.Provider, providerRow.Status, valueOrEmptyString(providerRow.Pesan))
	if parsed == nil || !parsed.CanPay || parsed.TotalAmount <= 0 {
		return 0, fmt.Errorf("hasil cek tagihan belum memiliki total bayar yang valid")
	}
	return parsed.TotalAmount, nil
}
