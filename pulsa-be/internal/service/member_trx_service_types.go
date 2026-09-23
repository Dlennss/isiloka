package service

import (
	"sync"
	"time"

	"pulsa2/ajs"
	"pulsa2/chytron"
	"pulsa2/gemilang"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/javapay"
	"pulsa2/loketbayar"
	"pulsa2/minions"
	"pulsa2/model"
	"pulsa2/multikom"
	"pulsa2/rajabiller"
	"pulsa2/sagaramobile"
	"pulsa2/smb"
	"pulsa2/talenta"
	"pulsa2/trionik"
	"pulsa2/yuscom"
)

type providerRouteAttempt struct {
	Name                string
	Need                int64
	Fee                 int64
	Src                 string
	ProdukSKUSnapshot   string
	ProdukProviderMapID *int64
	KodeProduk          string
	SpecialCode         string
	Mode                string
}

type providerState struct {
	name string
	row  *model.JavapayTrxRow
}

const retrySameRefCooldown = 30 * time.Second
const sameProviderHTTPRetryLimit = 5
const sameProviderHTTPRetryTimeout = 5 * time.Second
const providerCallTimeout = time.Duration((sameProviderHTTPRetryLimit*2)-1) * sameProviderHTTPRetryTimeout
const loketBayarRetryInterval = 5 * time.Second
const loketBayarRetryMaxWindow = 1 * time.Minute
const rajabillerCallTimeout = 45 * time.Second
const payInqHandleTimeout = 10 * time.Minute
const defaultHandleTimeout = 25 * time.Second
const javapayPendingRetryDelay = 5 * time.Minute
const smbCheckCallbackTimeout = 2 * time.Minute

type MemberTrxService struct {
	MemberRepo   *repository.MemberTrxMemberRepository
	Clients      map[string]provider.Client
	retryMu      sync.Mutex
	pinMu        sync.Mutex
	pinCache     map[pinCacheKey]time.Time
	pinLocks     sync.Map
	authMu       sync.Mutex
	authCache    map[authCacheKey]authCacheEntry
	ipAllowCache map[ipAllowCacheKey]time.Time

	YSClient *yuscom.Client
	JPClient *javapay.Client
	TLClient *talenta.Client
	MKClient *multikom.Client
	SGClient *sagaramobile.Client
	MNClient *minions.Client
	TRClient *trionik.Client
	AJClient *ajs.Client
	GMClient *gemilang.Client
	SMClient *smb.Client
	LBClient *loketbayar.Client
	CHClient *chytron.Client
	RJClient *rajabiller.Client

	JPRepo   *repository.MemberTrxProviderTrxRepository
	BankRepo *repository.BankRepository

	ProviderWallet      *repository.MemberTrxProviderWalletRepository
	ProviderMerchantIDs *repository.ProviderMerchantIDRepository
	H2HRepo             *repository.H2HRepository
	H2HProdukRepo       *repository.H2HProdukRepository
}

type ServiceErrorKind string

const (
	ErrBadRequest   ServiceErrorKind = "bad_request"
	ErrUnauthorized ServiceErrorKind = "unauthorized"
	ErrForbidden    ServiceErrorKind = "forbidden"
	ErrUpstream     ServiceErrorKind = "upstream"
)

type ServiceError struct {
	Kind    ServiceErrorKind
	Message string
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewMemberTrxService(repo *repository.MemberTrxRepository, ysClient *yuscom.Client, jpClient *javapay.Client, tlClient *talenta.Client, mkClient *multikom.Client, sgClient *sagaramobile.Client, mnClient *minions.Client, trClient *trionik.Client, ajClient *ajs.Client, gmClient *gemilang.Client, smClient *smb.Client, lbClient *loketbayar.Client, chClient *chytron.Client, rjClient *rajabiller.Client, extraClients ...provider.Client) *MemberTrxService {
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
	for _, client := range extraClients {
		if client != nil && client.Name() != "" {
			clients[client.Name()] = client
		}
	}

	return &MemberTrxService{
		MemberRepo:          repo.MemberRepo,
		Clients:             clients,
		YSClient:            ysClient,
		JPClient:            jpClient,
		TLClient:            tlClient,
		MKClient:            mkClient,
		SGClient:            sgClient,
		MNClient:            mnClient,
		TRClient:            trClient,
		AJClient:            ajClient,
		GMClient:            gmClient,
		SMClient:            smClient,
		LBClient:            lbClient,
		CHClient:            chClient,
		RJClient:            rjClient,
		JPRepo:              repo.ProviderTrx,
		BankRepo:            repository.NewBankRepository(repo.MemberRepo.DB()),
		ProviderWallet:      repo.ProviderWallet,
		ProviderMerchantIDs: repository.NewProviderMerchantIDRepository(repo.MemberRepo.DB()),
		H2HRepo:             repository.NewH2HRepository(repo.MemberRepo.DB()),
		H2HProdukRepo:       repo.H2HProduk,
	}
}
