package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"pulsa2/model"
)

type MemberTrxProviderTrxRepository struct {
	db *sql.DB
}

func NewMemberTrxProviderTrxRepository(db *sql.DB) *MemberTrxProviderTrxRepository {
	return &MemberTrxProviderTrxRepository{db: db}
}

type rowScan struct {
	id                  int64
	transaksiMemberID   int64
	provider            string
	refID               string
	perintah            string
	produkSKUSnapshot   sql.NullString
	produkProviderMapID sql.NullInt64
	kodeProduk          string
	tujuan              string
	qty                 int64

	trxIDJavapay  sql.NullString
	kodeRespon    sql.NullString
	pesan         sql.NullString
	noReferensi   sql.NullString
	harga         sql.NullInt64
	saldoTerakhir sql.NullInt64

	requestMentah sql.NullString
	responMentah  sql.NullString
	httpStatus    sql.NullInt64
	percobaan     int
	dibuatPada    time.Time
	status        string
}

func toPtrStr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
func toPtrI64(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}
func toPtrInt(ni sql.NullInt64) *int {
	if ni.Valid {
		v := int(ni.Int64)
		return &v
	}
	return nil
}

func parseJSONText(s sql.NullString) any {
	if !s.Valid || s.String == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(s.String), &v); err != nil {
		return map[string]any{"_raw": s.String}
	}
	return v
}

func scanToModel(s *rowScan) *model.JavapayTrxRow {
	return &model.JavapayTrxRow{
		ID:                  s.id,
		TransaksiMemberID:   s.transaksiMemberID,
		Provider:            s.provider,
		RefID:               s.refID,
		Perintah:            s.perintah,
		ProdukSKUSnapshot:   strings.TrimSpace(s.produkSKUSnapshot.String),
		ProdukProviderMapID: toPtrI64(s.produkProviderMapID),
		KodeProduk:          s.kodeProduk,
		Tujuan:              s.tujuan,
		Qty:                 s.qty,
		TrxIDJavapay:        toPtrStr(s.trxIDJavapay),
		KodeRespon:          toPtrStr(s.kodeRespon),
		Pesan:               toPtrStr(s.pesan),
		NoReferensi:         toPtrStr(s.noReferensi),
		Harga:               toPtrI64(s.harga),
		SaldoTerakhir:       toPtrI64(s.saldoTerakhir),
		HTTPStatus:          toPtrInt(s.httpStatus),
		Percobaan:           s.percobaan,
		DibuatPada:          s.dibuatPada,
		RequestMentah:       parseJSONText(s.requestMentah),
		ResponMentah:        parseJSONText(s.responMentah),
		Status:              s.status,
	}
}

func (r *MemberTrxProviderTrxRepository) getProviderByID(ctx context.Context, id int64) string {
	var provider sql.NullString
	if err := r.db.QueryRowContext(ctx, `SELECT provider FROM public.transaksi_provider WHERE id = $1`, id).Scan(&provider); err != nil {
		return ""
	}
	return strings.TrimSpace(provider.String)
}
