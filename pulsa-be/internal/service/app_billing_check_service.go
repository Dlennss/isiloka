package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

type AppBillingCheckService struct {
	checkRepo   *repository.AppBillingCheckRepository
	produkRepo  *repository.ProdukRepository
	pricingRepo *repository.ProdukAppPricingRepository
	p24Client   provider.Client
}

func NewAppBillingCheckService(checkRepo *repository.AppBillingCheckRepository, produkRepo *repository.ProdukRepository, pricingRepo *repository.ProdukAppPricingRepository, p24Client provider.Client) *AppBillingCheckService {
	return &AppBillingCheckService{checkRepo: checkRepo, produkRepo: produkRepo, pricingRepo: pricingRepo, p24Client: p24Client}
}

func (s *AppBillingCheckService) Create(ctx context.Context, in repository.AppBillingCheckCreateInput) (*repository.AppBillingCheckRow, error) {
	if s == nil || s.checkRepo == nil || s.produkRepo == nil || s.pricingRepo == nil || s.p24Client == nil || s.p24Client.Name() != provider.Pulsa24JamProviderName {
		return nil, fmt.Errorf("service billing check belum siap")
	}
	buyerType, memberID, err := normalizeBuyer(in.BuyerType, in.MemberID)
	if err != nil {
		return nil, err
	}
	produk, err := s.produkRepo.Get(ctx, in.ProdukID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("produk tidak ditemukan")
		}
		return nil, err
	}
	if !isAppCheckProduct(produk) {
		return nil, fmt.Errorf("produk ini bukan produk cek tagihan")
	}
	pricingRow, err := s.pricingRepo.GetByProdukIDProviderActive(ctx, in.ProdukID, provider.Pulsa24JamProviderName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("produk cek belum tersedia untuk app commerce")
		}
		return nil, err
	}
	if !pricingRow.Aktif {
		return nil, fmt.Errorf("produk cek tidak aktif")
	}

	dest := strings.TrimSpace(in.Dest)
	if dest == "" {
		return nil, fmt.Errorf("tujuan wajib diisi")
	}
	in.MemberID = memberID
	in.BuyerType = buyerType
	in.BuyerRole = strings.TrimSpace(in.BuyerRole)
	if in.BuyerRole == "" {
		in.BuyerRole = buyerType
	}
	if buyerType == "guest" {
		in.GuestNama = strings.TrimSpace(in.GuestNama)
		in.GuestEmail = normalizeGuestEmail(in.GuestEmail)
		in.GuestPhone = normalizeGuestPhone(in.GuestPhone)
		if in.GuestNama == "" || in.GuestEmail == "" || in.GuestPhone == "" {
			return nil, fmt.Errorf("guest_nama/guest_email/guest_phone wajib diisi")
		}
	}

	refID := buildAppBillingCheckRefID()
	rawReqBytes, _ := json.Marshal(map[string]any{
		"provider": provider.Pulsa24JamProviderName,
		"commands": "INQ",
		"product":  produk.SKU,
		"dest":     dest,
		"qty":      1,
		"refid":    refID,
	})
	rawReq := string(rawReqBytes)
	createIn := repository.AppBillingCheckCreateInput{
		RefID:              refID,
		MemberID:           memberID,
		BuyerRole:          in.BuyerRole,
		GuestNama:          in.GuestNama,
		GuestEmail:         in.GuestEmail,
		GuestPhone:         in.GuestPhone,
		ProdukID:           produk.ID,
		ProdukSKUSnapshot:  produk.SKU,
		ProdukNamaSnapshot: produk.Nama,
		Dest:               dest,
		BuyerType:          buyerType,
		Provider:           provider.Pulsa24JamProviderName,
		HargaProvider:      0,
		Status:             "processing_provider",
		RawRequest:         rawReq,
	}
	if err := s.checkRepo.Create(ctx, createIn); err != nil {
		return nil, err
	}
	createdRow, err := s.checkRepo.GetByRefID(ctx, refID)
	if err != nil {
		return nil, err
	}

	resp, callErr := s.p24Client.Pay(ctx, provider.PayRequest{
		Command: "INQ",
		Product: produk.SKU,
		Dest:    dest,
		Qty:     1,
		RefID:   refID,
	})
	body := ""
	if resp != nil {
		body = resp.Body
	}
	if callErr != nil {
		msg := callErr.Error()
		_ = s.checkRepo.UpdateResult(ctx, repository.AppBillingCheckUpdateInput{
			ID:          createdRow.ID,
			Status:      "failed",
			Pesan:       &msg,
			RawCallback: body,
		})
		return nil, fmt.Errorf("gagal mengirim cek ke provider")
	}
	if resp == nil {
		msg := "respons cek tagihan Pulsa24Jam kosong"
		_ = s.checkRepo.UpdateResult(ctx, repository.AppBillingCheckUpdateInput{
			ID:     createdRow.ID,
			Status: "failed",
			Pesan:  &msg,
		})
		return nil, fmt.Errorf("respons cek tagihan Pulsa24Jam kosong")
	}
	var pricePtr *int64
	if resp.Price != 0 {
		pricePtr = helper.PtrI64(resp.Price)
	}
	msg := strings.TrimSpace(resp.Message)
	if msg == "" {
		msg = strings.TrimSpace(resp.Body)
	}
	rc := strings.TrimSpace(resp.RC)
	status := "processing_provider"
	stateInput := msg
	if rawStatus, ok := resp.Raw["status"].(string); ok && strings.TrimSpace(rawStatus) != "" {
		stateInput = rawStatus
	}
	switch helper.ProviderResponseStateOf(provider.Pulsa24JamProviderName, rc, stateInput) {
	case helper.ProviderResponseSuccess:
		status = "success"
	case helper.ProviderResponseFailed:
		status = "failed"
	}
	_ = s.checkRepo.UpdateResult(ctx, repository.AppBillingCheckUpdateInput{
		ID:            createdRow.ID,
		HargaProvider: pricePtr,
		Status:        status,
		KodeRespon:    &rc,
		Pesan:         &msg,
		RawCallback:   body,
	})
	return s.GetByRefID(ctx, refID)
}

func (s *AppBillingCheckService) GetByRefID(ctx context.Context, refID string) (*repository.AppBillingCheckRow, error) {
	row, err := s.checkRepo.GetByRefID(ctx, refID)
	if err != nil {
		return nil, err
	}
	row.BillingInquiry = parseBillingInquiryFromCheck(row)
	return row, nil
}

func parseBillingInquiryFromCheck(row *repository.AppBillingCheckRow) *repository.AppBillingInquiryInfo {
	if row == nil {
		return nil
	}
	parsed := helper.ParseAppBillingInquiry(row.Provider, row.Status, valueOrEmptyString(row.Pesan))
	if parsed == nil {
		return nil
	}
	return mapBillingInquiry(parsed)
}

func buildAppBillingCheckRefID() string {
	return fmt.Sprintf("BILCHK-%s-%s", time.Now().Format("20060102150405"), strings.ToUpper(helper.RandHex(4)))
}
