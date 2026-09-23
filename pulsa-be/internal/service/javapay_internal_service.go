package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"pulsa2/db"
	"pulsa2/internal/repository"
	"pulsa2/javapay"
	"pulsa2/model"
)

type JavapayInternalService struct {
	repo *repository.JavapayInternalRepository
	jp   *javapay.Client
}

func NewJavapayInternalService(repo *repository.JavapayInternalRepository, jp *javapay.Client) *JavapayInternalService {
	return &JavapayInternalService{repo: repo, jp: jp}
}

func (s *JavapayInternalService) HandleTrx(ctx context.Context, in model.JavapayTrxCreateIn) (int, map[string]any) {
	in.RefID = strings.TrimSpace(in.RefID)
	in.Perintah = strings.TrimSpace(in.Perintah)
	in.KodeProduk = strings.TrimSpace(in.KodeProduk)
	in.Tujuan = strings.TrimSpace(in.Tujuan)
	command := strings.ToUpper(strings.TrimSpace(in.Perintah))
	if in.Qty <= 0 {
		in.Qty = 1
	}
	if in.TransaksiMemberID <= 0 || in.Perintah == "" {
		return 400, map[string]any{"ok": false, "error": "missing fields"}
	}
	if command != "STATUS" && command != "STATUS-PAY" && (in.KodeProduk == "" || in.Tujuan == "") {
		return 400, map[string]any{"ok": false, "error": "missing fields"}
	}

	var row *model.JavapayTrxRow
	if in.RefID != "" {
		existing, err := s.repo.GetLatestByRefIDAndPerintah(ctx, in.RefID, in.Perintah)
		if err != nil {
			return 502, map[string]any{"ok": false, "error": err.Error()}
		}
		if existing != nil {
			if existing.HTTPStatus != nil {
				return 200, map[string]any{"ok": true, "existing": true, "row": existing}
			}
			row = existing
		}
	}

	if row == nil {
		reqShadow := map[string]any{
			"perintah":    in.Perintah,
			"kode_produk": in.KodeProduk,
			"tujuan":      in.Tujuan,
			"qty":         in.Qty,
			"ref_id":      in.RefID,
		}
		created, err := s.repo.Create(ctx, in, reqShadow)
		if err != nil {
			return 502, map[string]any{"ok": false, "error": err.Error()}
		}
		row = created
	}

	callCtx, cancel := context.WithTimeout(ctx, s.jp.Timeout)
	defer cancel()

	refID := row.RefID
	if command == "STATUS-PAY" {
		command = "STATUS"
	}

	var (
		respMap    map[string]any
		httpStatus int
		reqMentah  map[string]any
		err        error
	)
	if command == "STATUS" {
		respMap, httpStatus, reqMentah, err = s.jp.Status(callCtx, refID)
	} else {
		respMap, httpStatus, reqMentah, err = s.jp.Trx(callCtx, command, in.KodeProduk, in.Tujuan, in.Qty, refID)
	}
	if err != nil {
		hs := 0
		_ = s.repo.UpdateResult(ctx, row.ID, db.UpdateResult{HTTPStatus: &hs, ResponMentah: map[string]any{"status": false, "error": err.Error()}})
		return 502, map[string]any{"ok": false, "error": err.Error(), "ref_id": refID}
	}

	_ = s.repo.UpdateRequestMentah(ctx, row.ID, reqMentah)

	upd := db.UpdateResult{HTTPStatus: &httpStatus, ResponMentah: respMap}
	if data, ok := respMap["data"].(map[string]any); ok {
		if v := toString(data["trxid"]); v != "" {
			upd.TrxIDJavapay = &v
		}
		if v := toString(data["rc"]); v != "" {
			upd.KodeRespon = &v
		}
		if v := toString(data["message"]); v != "" {
			upd.Pesan = &v
		}
		if v := toString(data["noreff"]); v != "" {
			upd.NoReferensi = &v
		}
		if n, ok := toInt64(data["price"]); ok {
			upd.Harga = &n
		}
		if n, ok := toInt64(data["last_balance"]); ok {
			upd.SaldoTerakhir = &n
		}
	}
	_ = s.repo.UpdateResult(ctx, row.ID, upd)

	latest, _ := s.repo.GetLatestByRefIDAndPerintah(ctx, refID, in.Perintah)
	return 200, map[string]any{
		"ok":      true,
		"row":     latest,
		"javapay": model.JavapayCallResult{HTTPStatus: httpStatus, Body: respMap},
		"ts":      time.Now().Format(time.RFC3339),
	}
}

func (s *JavapayInternalService) ProdukStatus(ctx context.Context) (int, map[string]any) {
	callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	resp, httpStatus, err := s.jp.ProdukStatus(callCtx)
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	return 200, map[string]any{"ok": true, "http_status": httpStatus, "javapay": resp}
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		return ""
	}
}

func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int:
		return int64(t), true
	case int64:
		return t, true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err == nil {
			return n, true
		}
		return 0, false
	default:
		return 0, false
	}
}
