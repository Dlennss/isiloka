package providerreportingdto

import "pulsa2/internal/repository"

func mapTransactionRow(x repository.ProviderTransactionRecord) TransactionRow {
	return TransactionRow{
		ID:                x.ID,
		DibuatPada:        x.DibuatPada,
		Provider:          x.Provider,
		StatusProvider:    x.StatusProvider,
		TransaksiMemberID: x.TransaksiMemberID,
		MemberID:          x.MemberID,
		MemberNama:        x.MemberNama,
		RefID:             x.RefID,
		Perintah:          x.Perintah,
		KodeProduk:        x.KodeProduk,
		Tujuan:            x.Tujuan,
		Qty:               x.Qty,
		TrxIDJavapay:      x.TrxIDJavapay,
		TrxIDProvider:     x.TrxIDProvider,
		KodeRespon:        x.KodeRespon,
		Pesan:             x.Pesan,
		NoReferensi:       x.NoReferensi,
		Harga:             x.Harga,
		SaldoTerakhir:     x.SaldoTerakhir,
		HTTPStatus:        x.HTTPStatus,
		Percobaan:         x.Percobaan,
	}
}

func mapAnalyticsRow(x repository.ProviderAnalyticsRecord) AnalyticsRow {
	return AnalyticsRow{
		Provider:       x.Provider,
		Total:          x.Total,
		Success:        x.Success,
		Failed:         x.Failed,
		SumQty:         x.SumQty,
		SumHarga:       x.SumHarga,
		SuccessNominal: x.SuccessNominal,
		DepositAmount:  x.DepositAmount,
	}
}

func mapAnalyticsPeriodRow(x repository.ProviderAnalyticsPeriodRecord) AnalyticsPeriodRow {
	return AnalyticsPeriodRow{
		Provider:       x.Provider,
		PeriodStart:    x.PeriodStart,
		SuccessCount:   x.SuccessCount,
		SuccessNominal: x.SuccessNominal,
		DepositAmount:  x.DepositAmount,
	}
}

func mapDailyProductSuccessRow(x repository.DailyProductSuccessRecord) DailyProductSuccessRow {
	return DailyProductSuccessRow{
		PeriodStart:        x.PeriodStart,
		InternalSKU:        x.InternalSKU,
		ProductName:        x.ProductName,
		GroupName:          x.GroupName,
		SuccessCount:       x.SuccessCount,
		TotalQty:           x.TotalQty,
		TotalQtyProvider:   x.TotalQtyProvider,
		TotalProviderPrice: x.TotalProviderPrice,
		TotalMemberPrice:   x.TotalMemberPrice,
		TotalMargin:        x.TotalMargin,
		ProviderCount:      x.ProviderCount,
		Providers:          x.Providers,
		FirstSuccessAt:     x.FirstSuccessAt,
		LastSuccessAt:      x.LastSuccessAt,
	}
}

func mapDailyProductSuccessSummary(x repository.DailyProductSuccessSummary) DailyProductSuccessSummary {
	return DailyProductSuccessSummary{
		GroupCount:         x.GroupCount,
		UniqueSKUCount:     x.UniqueSKUCount,
		SuccessCount:       x.SuccessCount,
		TotalQty:           x.TotalQty,
		TotalQtyProvider:   x.TotalQtyProvider,
		TotalProviderPrice: x.TotalProviderPrice,
		TotalMemberPrice:   x.TotalMemberPrice,
		TotalMargin:        x.TotalMargin,
	}
}

func MapListTransactions(items []repository.ProviderTransactionRecord, total int64) ListTransactionsResponse {
	out := make([]TransactionRow, 0, len(items))
	for _, it := range items {
		out = append(out, mapTransactionRow(it))
	}
	return ListTransactionsResponse{
		Ok:    true,
		Items: out,
		Total: total,
	}
}

func MapAnalytics(data repository.ProviderAnalyticsBundle) AnalyticsResponse {
	out := make([]AnalyticsRow, 0, len(data.Items))
	for _, it := range data.Items {
		out = append(out, mapAnalyticsRow(it))
	}
	daily := make([]AnalyticsPeriodRow, 0, len(data.DailyItems))
	for _, it := range data.DailyItems {
		daily = append(daily, mapAnalyticsPeriodRow(it))
	}
	monthly := make([]AnalyticsPeriodRow, 0, len(data.MonthlyItems))
	for _, it := range data.MonthlyItems {
		monthly = append(monthly, mapAnalyticsPeriodRow(it))
	}
	return AnalyticsResponse{
		Ok:           true,
		Items:        out,
		DailyItems:   daily,
		MonthlyItems: monthly,
	}
}

func MapDailyProductSuccess(data repository.DailyProductSuccessBundle) DailyProductSuccessResponse {
	out := make([]DailyProductSuccessRow, 0, len(data.Items))
	for _, it := range data.Items {
		out = append(out, mapDailyProductSuccessRow(it))
	}
	return DailyProductSuccessResponse{
		Ok:      true,
		Items:   out,
		Total:   data.Total,
		Summary: mapDailyProductSuccessSummary(data.Summary),
	}
}

func mapAnomalyRow(x repository.ProviderAnomalyRecord) AnomalyRow {
	return AnomalyRow{
		ID:               x.ID,
		DibuatPada:       x.DibuatPada,
		Provider:         x.Provider,
		RefID:            x.RefID,
		KodeRespon:       x.KodeRespon,
		Pesan:            x.Pesan,
		Harga:            x.Harga,
		Tujuan:           x.Tujuan,
		Qty:              x.Qty,
		PayloadHash:      x.PayloadHash,
		IsDuplicate:      x.IsDuplicate,
		IsSuspectedFraud: x.IsSuspectedFraud,
		FraudReason:      x.FraudReason,
		RawQuery:         x.RawQuery,
		RawBody:          x.RawBody,
	}
}

func MapListAnomalies(items []repository.ProviderAnomalyRecord, total int64) ListAnomaliesResponse {
	out := make([]AnomalyRow, 0, len(items))
	for _, it := range items {
		out = append(out, mapAnomalyRow(it))
	}
	return ListAnomaliesResponse{
		Ok:    true,
		Items: out,
		Total: total,
	}
}
