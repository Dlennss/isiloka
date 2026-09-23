package service

import (
	"pulsa2/gemilang"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/yuscom"
)

type AppOrderFulfillmentService struct {
	orderRepo       *repository.AppOrderRepository
	providerTrxRepo *repository.AppOrderProviderTrxRepository
	callbackRepo    *repository.ProviderCallbackRepository
	pricingRepo     *repository.ProdukAppPricingRepository
	ysClient        *yuscom.Client
	gmClient        *gemilang.Client
	providerClients map[string]provider.Client
}

func NewAppOrderFulfillmentService(orderRepo *repository.AppOrderRepository, providerTrxRepo *repository.AppOrderProviderTrxRepository, callbackRepo *repository.ProviderCallbackRepository, pricingRepo *repository.ProdukAppPricingRepository, ysClient *yuscom.Client, gmClient *gemilang.Client, extraClients ...provider.Client) *AppOrderFulfillmentService {
	providerClients := map[string]provider.Client{}
	for _, client := range extraClients {
		if client != nil && client.Name() != "" {
			providerClients[client.Name()] = client
		}
	}
	return &AppOrderFulfillmentService{
		orderRepo:       orderRepo,
		providerTrxRepo: providerTrxRepo,
		callbackRepo:    callbackRepo,
		pricingRepo:     pricingRepo,
		ysClient:        ysClient,
		gmClient:        gmClient,
		providerClients: providerClients,
	}
}
