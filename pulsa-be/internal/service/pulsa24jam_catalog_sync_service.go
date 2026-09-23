package service

import (
	"context"
	"fmt"

	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

type Pulsa24JamCatalogSyncService struct {
	repo   *repository.Pulsa24JamCatalogRepository
	client *provider.Pulsa24JamAdapter
}

func NewPulsa24JamCatalogSyncService(dbRepo *repository.Pulsa24JamCatalogRepository, client *provider.Pulsa24JamAdapter) *Pulsa24JamCatalogSyncService {
	return &Pulsa24JamCatalogSyncService{repo: dbRepo, client: client}
}

func (s *Pulsa24JamCatalogSyncService) Sync(ctx context.Context) (*repository.Pulsa24JamCatalogSyncResult, error) {
	if s == nil || s.repo == nil || s.client == nil || !s.client.Configured() {
		return nil, fmt.Errorf("sinkronisasi katalog Pulsa24Jam belum dikonfigurasi")
	}
	products, err := s.client.Products(ctx, "")
	if err != nil {
		return nil, err
	}
	items := make([]repository.Pulsa24JamCatalogItem, 0, len(products))
	for _, product := range products {
		items = append(items, pulsa24JamCatalogItemFromProduct(product))
	}
	return s.repo.Sync(ctx, items)
}

func pulsa24JamCatalogItemFromProduct(product provider.Pulsa24JamProduct) repository.Pulsa24JamCatalogItem {
	price := int64(0)
	if product.AppBasePrice != nil {
		price = *product.AppBasePrice
	} else if product.Price != nil {
		price = *product.Price
	} else if product.AdditionalFee != nil {
		price = *product.AdditionalFee
	}
	return repository.Pulsa24JamCatalogItem{
		SKU:            product.SKU,
		Name:           product.Name,
		GroupName:      product.GroupName,
		CategoryName:   product.CategoryName,
		BrandName:      product.BrandName,
		PriceType:      product.PriceType,
		Price:          price,
		MaximumNominal: product.MaximumNominal,
	}
}
