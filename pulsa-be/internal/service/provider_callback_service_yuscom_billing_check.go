package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"strconv"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) processYuscomBillingCheckCallback(ctx context.Context, refid string, statusNum int, price int64, msg string, q url.Values) (int, map[string]any) {
	row, err := s.billingCheckRepo.GetByRefID(ctx, refid)
	if err != nil || row == nil {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	finalStatus := resolveYuscomFinalStatus(statusNum, strconv.Itoa(statusNum), msg)
	status := "processing_provider"
	switch finalStatus {
	case "success":
		status = "success"
	case "failed":
		status = "failed"
	case "pending":
		status = "processing_provider"
	}
	rc := strconv.Itoa(statusNum)
	pricePtr := &price
	rawCallback, _ := json.Marshal(map[string]any{
		"t":       q.Get("t"),
		"refid":   refid,
		"status":  statusNum,
		"price":   price,
		"message": msg,
	})
	if err := s.billingCheckRepo.UpdateResult(ctx, repository.AppBillingCheckUpdateInput{
		ID:            row.ID,
		HargaProvider: pricePtr,
		Status:        status,
		KodeRespon:    &rc,
		Pesan:         &msg,
		RawCallback:   string(rawCallback),
	}); err != nil && err != sql.ErrNoRows {
		helper.AppendProviderServiceLog("provider_callback_error.log", "update billing check callback failed refid=%s billing_check_id=%d err=%v", refid, row.ID, err)
	}
	return 200, map[string]any{"ok": true, "refid": refid, "status": status}
}
