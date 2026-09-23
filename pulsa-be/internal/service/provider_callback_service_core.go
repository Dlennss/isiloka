package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/gemilang"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/javapay"
	"pulsa2/loketbayar"
	"pulsa2/minions"
	"pulsa2/multikom"
	"pulsa2/rajabiller"
	"pulsa2/sagaramobile"
	"pulsa2/smb"
	"pulsa2/talenta"
	"pulsa2/trionik"
	"pulsa2/yuscom"
)

var providerCallbackLocks sync.Map

func lockProviderCallback(key int64) func() {
	if key <= 0 {
		return func() {}
	}
	actual, _ := providerCallbackLocks.LoadOrStore(key, &sync.Mutex{})
	mu := actual.(*sync.Mutex)
	mu.Lock()
	return func() {
		mu.Unlock()
	}
}

type callbackFallbackCandidate struct {
	Provider            string
	ProdukSKUSnapshot   string
	ProdukProviderMapID *int64
	KodeProduk          string
	SpecialCode         *string
	Mode                *string
	Need                int64
	Fee                 int64
	Source              string
}

type ProviderCallbackService struct {
	repo                *repository.ProviderCallbackRepository
	depositRepo         *repository.DepositRepository
	loketTransferRepo   *repository.LoketBayarTransferRepository
	clients             map[string]provider.Client
	appProviderRepo     *repository.AppOrderProviderTrxRepository
	appPricingRepo      *repository.ProdukAppPricingRepository
	billingCheckRepo    *repository.AppBillingCheckRepository
	appOrderRepo        *repository.AppOrderRepository
	retailRepo          *repository.RetailRepository
	h2hRepo             *repository.H2HRepository
	providerMerchantIDs *repository.ProviderMerchantIDRepository
	jpClient            *javapay.Client
	ysClient            *yuscom.Client
	tlClient            *talenta.Client
	mkClient            *multikom.Client
	sgClient            *sagaramobile.Client
	mnClient            *minions.Client
	trClient            *trionik.Client
	ajClient            *ajs.Client
	gmClient            *gemilang.Client
	smClient            *smb.Client
	lbClient            *loketbayar.Client
	chClient            *chytron.Client
	rjClient            *rajabiller.Client
}

func (s *ProviderCallbackService) VerifyJavapayCallbackSignature(rawBody []byte, signature string) bool {
	if s == nil || s.jpClient == nil {
		return false
	}
	return s.jpClient.VerifyCallbackSignature(rawBody, signature)
}

func buildDirectMemberWebhookPayload(trx *repository.CallbackTrxMemberFull, finalStatus, ket, providerRef, sn string, memberSaldo, biayaAktual int64) map[string]any {
	if trx == nil {
		return map[string]any{}
	}
	qtyProvider := trx.QtyProvider
	if qtyProvider <= 0 {
		qtyProvider = trx.Qty
	}
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	storedKet := strings.TrimSpace(trx.Keterangan)
	messageField, providerRefField, snField := normalizeMemberWebhookFields(finalStatus, ket, providerRef, sn, storedKet)

	return map[string]any{
		"refid":          trx.RefID,
		"status":         finalStatus,
		"member_balance": memberSaldo,
		"trx": map[string]any{
			"id":           trx.ID,
			"commands":     trx.Perintah,
			"product":      trx.KodeProduk,
			"dest":         trx.Tujuan,
			"qty":          trx.Qty,
			"qty_provider": qtyProvider,
			"harga_member": hargaMember,
			"biaya_aktual": biayaAktual,
			"message":      messageField,
			"provider_ref": providerRefField,
			"sn":           snField,
		},
	}
}

func normalizeMemberWebhookFields(finalStatus, ket, providerRef, sn, storedKet string) (messageField, providerRefField, snField string) {
	status := strings.ToLower(strings.TrimSpace(finalStatus))
	messageField = strings.TrimSpace(ket)
	providerRefField = strings.TrimSpace(providerRef)
	snField = strings.TrimSpace(sn)
	storedKet = strings.TrimSpace(storedKet)

	if status == "failed" {
		if messageField == "" {
			messageField = storedKet
		}
		safeKet, _ := helper.SafeMemberKeterangan("failed", messageField)
		if strings.TrimSpace(safeKet) == "" {
			safeKet = "Transaksi gagal"
		}
		return safeKet, "", safeKet
	}

	if storedKet != "" && !helper.IsProviderRawMemberMessage(storedKet) {
		if messageField == "" {
			messageField = storedKet
		}
		if providerRefField == "" {
			providerRefField = storedKet
		}
		if snField == "" {
			snField = storedKet
		}
	}

	if helper.IsProviderRawMemberMessage(messageField) {
		safeKet, _ := helper.SafeMemberKeterangan(status, messageField)
		messageField = safeKet
	}
	if helper.IsProviderRawMemberMessage(providerRefField) {
		providerRefField = ""
	}
	if helper.IsProviderRawMemberMessage(snField) {
		snField = ""
	}
	if status == "pending" && strings.TrimSpace(messageField) == "" {
		messageField = "Sedang diproses"
	}
	return messageField, providerRefField, snField
}

func (s *ProviderCallbackService) sendDirectMemberWebhook(_ context.Context, provider string, trx *repository.CallbackTrxMemberFull, webhookURL, refid, finalStatus, ket, providerRef, sn string, memberSaldo, biayaAktual, price int64) (int, map[string]any) {
	if trx == nil {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	trxCopy := *trx
	go func() {
		callCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = s.sendDirectMemberWebhookSync(callCtx, provider, &trxCopy, webhookURL, refid, finalStatus, ket, providerRef, sn, memberSaldo, biayaAktual, price)
	}()
	return 200, map[string]any{"ok": true, "refid": refid, "member_webhook_queued": true}
}

func (s *ProviderCallbackService) sendDirectMemberWebhookSync(ctx context.Context, provider string, trx *repository.CallbackTrxMemberFull, webhookURL, refid, finalStatus, ket, providerRef, sn string, memberSaldo, biayaAktual, _ int64) (int, map[string]any) {
	if trx == nil {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	if s != nil && s.repo != nil && trx.ID > 0 {
		if latestTrx, err := s.repo.GetTransaksiMemberByID(ctx, trx.ID); err == nil && latestTrx != nil {
			trx = latestTrx
			latestStatus := strings.ToLower(strings.TrimSpace(latestTrx.Status))
			if latestStatus == "success" || latestStatus == "failed" {
				finalStatus = latestStatus
			}
		}
	}

	out := buildDirectMemberWebhookPayload(trx, finalStatus, ket, providerRef, sn, memberSaldo, biayaAktual)

	callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	st, body, callErr := helper.PostJSON(callCtx, webhookURL, out)
	if callErr != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log", "WEBHOOK gagal provider=%s refid=%s trx_id=%d err=%v", provider, refid, trx.ID, callErr)
		return 200, map[string]any{
			"ok":                   true,
			"refid":                refid,
			"member_webhook_url":   webhookURL,
			"member_webhook_error": callErr.Error(),
		}
	}

	return 200, map[string]any{
		"ok":    true,
		"refid": refid,
		"member_webhook": map[string]any{
			"url":    webhookURL,
			"status": st,
			"body":   body,
		},
	}
}

func isWeakProviderValue(v string) bool {
	v = strings.TrimSpace(strings.ToUpper(v))
	switch v {
	case "", "NO", "N/A", "-":
		return true
	default:
		return false
	}
}

func minionsFeeByCodeForCallback(code string) int64 {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "HDANA":
		return 55
	case "HDANAB":
		return 100
	case "HGOPAY":
		return 975
	case "HLINK":
		return 625
	case "HLINKB":
		return 700
	case "HOVO":
		return 580
	case "HSHOPEE":
		return 300
	case "HSHOPEEB":
		return 500
	default:
		return 0
	}
}

func (s *ProviderCallbackService) applyH2HCommission(ctx context.Context, trxID int64, refID string) {
	if s.h2hRepo == nil || trxID <= 0 {
		return
	}
	if err := s.h2hRepo.ApplyCommissionForSuccessTrx(ctx, trxID); err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "h2h commission apply failed refid=%s trx_member_id=%d err=%v", refID, trxID, err)
	}
}

func shouldKeepExistingMemberFinalStatus(currentStatus, _ string) bool {
	currentStatus = strings.ToLower(strings.TrimSpace(currentStatus))
	switch currentStatus {
	case "success":
		return true
	case "failed":
		return true
	default:
		return false
	}
}

func (s *ProviderCallbackService) prepareMemberTrxSuccessTransition(ctx context.Context, _ string, trx *repository.CallbackTrxMemberFull) error {
	if s == nil || s.repo == nil || trx == nil {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(trx.Status), "failed") {
		return nil
	}
	return s.repo.RecoverMemberRefund(ctx, trx.MemberID, trx.RefID, trx.BiayaPerkiraan, "CALLBACK_SUCCESS_RECOVERY", "reverse refund after provider success")
}

func NewProviderCallbackService(repo *repository.ProviderCallbackRepository, jpClient *javapay.Client, ysClient *yuscom.Client, tlClient *talenta.Client, mkClient *multikom.Client, sgClient *sagaramobile.Client, mnClient *minions.Client, trClient *trionik.Client, ajClient *ajs.Client, gmClient *gemilang.Client, smClient *smb.Client, lbClient *loketbayar.Client, chClient *chytron.Client, rjClient *rajabiller.Client) *ProviderCallbackService {
	clients := map[string]provider.Client{}
	if ysClient != nil {
		clients["yuscom"] = &provider.YuscomAdapter{C: ysClient}
	}
	if jpClient != nil {
		clients["javapay"] = &provider.JavapayAdapter{C: jpClient}
	}
	if tlClient != nil {
		clients["talentapay"] = &provider.TalentaAdapter{C: tlClient}
	}
	if mkClient != nil {
		clients["multikom"] = &provider.MultikomAdapter{C: mkClient}
	}
	if sgClient != nil {
		clients["sagaramobile"] = &provider.SagaramobileAdapter{C: sgClient}
	}
	if mnClient != nil {
		clients["minions"] = &provider.MinionsAdapter{C: mnClient}
	}
	if trClient != nil {
		clients["trionik"] = &provider.TrionikAdapter{C: trClient}
	}
	if ajClient != nil {
		clients["ajs"] = &provider.AJSAdapter{C: ajClient}
	}
	if gmClient != nil {
		clients["gemilang"] = &provider.GemilangAdapter{C: gmClient}
	}
	if smClient != nil {
		clients["smb"] = &provider.SMBAdapter{C: smClient}
	}
	if lbClient != nil {
		clients["loketbayar"] = &provider.LoketBayarAdapter{C: lbClient}
	}
	if chClient != nil {
		clients["chytron"] = &provider.ChytronAdapter{C: chClient}
	}
	if rjClient != nil {
		clients["rajabiller"] = &provider.RajabillerAdapter{C: rjClient}
	}

	return &ProviderCallbackService{
		repo:                repo,
		depositRepo:         repository.NewDepositRepository(repo.DB()),
		loketTransferRepo:   repository.NewLoketBayarTransferRepository(repo.DB()),
		clients:             clients,
		appProviderRepo:     repository.NewAppOrderProviderTrxRepository(repo.DB()),
		appPricingRepo:      repository.NewProdukAppPricingRepository(repo.DB()),
		billingCheckRepo:    repository.NewAppBillingCheckRepository(repo.DB()),
		appOrderRepo:        repository.NewAppOrderRepository(repo.DB()),
		retailRepo:          repository.NewRetailRepository(repo.DB()),
		h2hRepo:             repository.NewH2HRepository(repo.DB()),
		providerMerchantIDs: repository.NewProviderMerchantIDRepository(repo.DB()),
		jpClient:            jpClient,
		ysClient:            ysClient,
		tlClient:            tlClient,
		mkClient:            mkClient,
		sgClient:            sgClient,
		mnClient:            mnClient,
		trClient:            trClient,
		ajClient:            ajClient,
		gmClient:            gmClient,
		smClient:            smClient,
		lbClient:            lbClient,
		chClient:            chClient,
		rjClient:            rjClient,
	}
}
