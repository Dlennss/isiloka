package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

func parseLiveCatalog(raw []byte) ([]productRow, error) {
	var response struct {
		OK    bool `json:"ok"`
		Items []struct {
			provider.Pulsa24JamProduct
			Active *bool `json:"aktif"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("provider did not return a successful catalogue")
	}
	rows := make([]productRow, 0, len(response.Items))
	items := make([]repository.Pulsa24JamCatalogItem, 0, len(response.Items))
	for _, item := range response.Items {
		if item.Active != nil && !*item.Active {
			continue
		}
		price, err := item.CatalogPrice()
		if err != nil {
			return nil, err
		}
		items = append(items, repository.Pulsa24JamCatalogItem{SKU: item.SKU, Name: item.Name, GroupName: item.GroupName, CategoryName: item.CategoryName, BrandName: item.BrandName, PriceType: item.PriceType, Price: price, MaximumNominal: item.MaximumNominal})
		maximum := ""
		if item.MaximumNominal != nil {
			maximum = strconv.FormatInt(*item.MaximumNominal, 10)
		}
		rows = append(rows, productRow{SKU: item.SKU, Name: item.Name, GroupName: item.GroupName, Category: item.CategoryName, Brand: item.BrandName, PriceType: item.PriceType, Price: price, MaximumNominal: maximum})
	}
	if err := repository.ValidatePulsa24JamCatalog(items); err != nil {
		return nil, err
	}
	return rows, nil
}
