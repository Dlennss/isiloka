package service

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type BankService struct {
	repo *repository.BankRepository
}

func NewBankService(repo *repository.BankRepository) *BankService {
	return &BankService{repo: repo}
}

func (s *BankService) List(ctx context.Context, role string) ([]repository.BankRow, error) {
	return s.repo.List(ctx, helper.IsAdminLikeRole(role))
}

func (s *BankService) Get(ctx context.Context, id int64, role string) (*repository.BankRow, error) {
	return s.repo.GetVisible(ctx, id, helper.IsAdminLikeRole(role))
}

func (s *BankService) EnsureVisibleToRole(ctx context.Context, bankID int64, role string) error {
	if bankID <= 0 {
		return errors.New("bank_id invalid")
	}
	_, err := s.repo.GetVisible(ctx, bankID, helper.IsAdminLikeRole(role))
	return err
}

func (s *BankService) Create(ctx context.Context, actorID int64, in repository.BankUpsertInput) (int64, error) {
	in.Nama = strings.TrimSpace(in.Nama)
	in.NomorRekening = strings.TrimSpace(in.NomorRekening)
	in.AtasNama = strings.TrimSpace(in.AtasNama)
	if in.Nama == "" {
		return 0, errors.New("nama required")
	}
	if in.AtasNama == "" {
		return 0, errors.New("atas_nama required")
	}
	if in.NomorRekening == "" {
		return 0, errors.New("nomor_rekening required")
	}
	if in.Saldo < 0 {
		return 0, errors.New("saldo must be >= 0")
	}
	refID := ""
	if in.Saldo > 0 {
		refID = "BOPEN-" + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	}
	return s.repo.Create(ctx, actorID, in, refID)
}

func (s *BankService) Update(ctx context.Context, in repository.BankUpsertInput) error {
	in.Nama = strings.TrimSpace(in.Nama)
	in.NomorRekening = strings.TrimSpace(in.NomorRekening)
	in.AtasNama = strings.TrimSpace(in.AtasNama)
	if in.ID <= 0 {
		return errors.New("id invalid")
	}
	if in.Nama == "" {
		return errors.New("nama required")
	}
	if in.AtasNama == "" {
		return errors.New("atas_nama required")
	}
	if in.NomorRekening == "" {
		return errors.New("nomor_rekening required")
	}
	if in.Saldo < 0 {
		return errors.New("saldo must be >= 0")
	}
	return s.repo.Update(ctx, in)
}

func (s *BankService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id invalid")
	}
	return s.repo.Delete(ctx, id)
}

func (s *BankService) ToggleActive(ctx context.Context, id int64, aktif bool) error {
	if id <= 0 {
		return errors.New("id invalid")
	}
	return s.repo.ToggleActive(ctx, id, aktif)
}

func (s *BankService) AdjustSaldo(ctx context.Context, actorID, bankID, amount int64, direction, note string) (string, int64, error) {
	direction = strings.TrimSpace(strings.ToLower(direction))
	note = strings.TrimSpace(note)
	if actorID <= 0 || bankID <= 0 || amount <= 0 {
		return "", 0, errors.New("bank_id/amount invalid")
	}
	if direction != "credit" && direction != "debit" {
		return "", 0, errors.New("direction must be credit or debit")
	}
	if note == "" {
		return "", 0, errors.New("note required")
	}

	refID := "BADJ-" + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	reason := "BANK_ADJUST_" + strings.ToUpper(direction)
	_, after, err := s.repo.AdjustSaldo(ctx, actorID, bankID, amount, direction, reason, note, refID)
	if err != nil {
		return "", 0, err
	}
	return refID, after, nil
}

func (s *BankService) ManualIncomingMutation(ctx context.Context, actorID int64, actorRole string, bankID int64, amount int64, sender string, receiver string, note string, externalRef string, direction string, actualBalance int64, bankMutationTime string) (string, int64, bool, error) {
	sender = strings.TrimSpace(sender)
	receiver = strings.TrimSpace(receiver)
	note = strings.TrimSpace(note)
	externalRef = strings.TrimSpace(externalRef)
	direction = strings.TrimSpace(strings.ToLower(direction))
	if direction == "" {
		direction = "credit"
	}
	if direction != "credit" && direction != "debit" {
		return "", 0, false, errors.New("direction must be credit or debit")
	}
	if actorID <= 0 || bankID <= 0 || amount <= 0 {
		return "", 0, false, errors.New("bank_id/amount invalid")
	}
	if actualBalance < 0 {
		return "", 0, false, errors.New("balance invalid")
	}
	if sender == "" {
		return "", 0, false, errors.New("pengirim required")
	}
	if err := s.EnsureVisibleToRole(ctx, bankID, actorRole); err != nil {
		return "", 0, false, err
	}
	if externalRef != "" && !validExternalRef(externalRef) {
		return "", 0, false, errors.New("external_ref invalid")
	}
	mutationAt, err := parseBankMutationTime(bankMutationTime)
	if err != nil {
		return "", 0, false, err
	}

	catatan := "Pengirim: " + sender
	if receiver != "" {
		catatan += " | Penerima: " + receiver
	}
	if note != "" {
		catatan += " | " + note
	}
	refID := externalRef
	if refID == "" {
		prefix := "BMIN-"
		if direction == "debit" {
			prefix = "BMOUT-"
		}
		refID = prefix + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	}
	_, after, err := s.repo.ManualMutationWithBalanceDetails(ctx, actorID, bankID, amount, direction, catatan, refID, actualBalance, mutationAt, sender, receiver)
	if err != nil {
		var dup *repository.DuplicateBankMutationError
		if errors.As(err, &dup) {
			return dup.RefID, dup.Saldo, true, nil
		}
		return "", 0, false, err
	}
	return refID, after, false, nil
}

func (s *BankService) Kantor24IncomingMutation(ctx context.Context, bankID int64, amount int64, sender string, receiver string, note string, externalRef string, direction string, actualBalance int64, bankMutationTime string) (string, int64, bool, error) {
	externalRef = strings.TrimSpace(externalRef)
	if !strings.HasPrefix(externalRef, "K24-") {
		return "", 0, false, errors.New("kantor24 external_ref must start with K24-")
	}
	return s.ManualIncomingMutation(ctx, kantor24ActorID(), "admin", bankID, amount, sender, receiver, note, externalRef, direction, actualBalance, bankMutationTime)
}

func kantor24ActorID() int64 {
	raw := strings.TrimSpace(os.Getenv("KANTOR24_ACTOR_ID"))
	if raw != "" {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return 365
}

func parseBankMutationTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		parsed := t.In(loc)
		return &parsed, nil
	}
	value = replaceIndonesianMonthNames(value)
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2 January 2006 15:04:05",
		"02 January 2006 15:04:05",
		"2 Jan 2006 15:04:05",
		"02 Jan 2006 15:04:05",
		"2/1/2006 15:04:05",
		"02/01/2006 15:04:05",
		"2-Jan-2006 15:04:05",
		"02-Jan-2006 15:04:05",
		"2006-01-02",
		"2/1/2006",
		"02/01/2006",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return &t, nil
		}
	}
	return nil, errors.New("waktu_mutasi_bank invalid")
}

func replaceIndonesianMonthNames(value string) string {
	replacer := strings.NewReplacer(
		"Januari", "January", "januari", "January",
		"Februari", "February", "februari", "February",
		"Maret", "March", "maret", "March",
		"April", "April", "april", "April",
		"Mei", "May", "mei", "May",
		"Juni", "June", "juni", "June",
		"Juli", "July", "juli", "July",
		"Agustus", "August", "agustus", "August",
		"September", "September", "september", "September",
		"Oktober", "October", "oktober", "October",
		"November", "November", "november", "November",
		"Desember", "December", "desember", "December",
	)
	return replacer.Replace(strings.Join(strings.Fields(value), " "))
}

var externalRefPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,120}$`)

func validExternalRef(value string) bool {
	if !externalRefPattern.MatchString(value) {
		return false
	}
	if strings.HasPrefix(value, "BMIN-") || strings.HasPrefix(value, "K24-") {
		return true
	}
	return false
}

func (s *BankService) TransferOut(ctx context.Context, actorID, bankID int64, tujuan string, amount int64, note string) (string, int64, string, error) {
	if actorID <= 0 || bankID <= 0 || amount <= 0 {
		return "", 0, "", errors.New("bank_id/amount invalid")
	}
	tujuan = strings.TrimSpace(tujuan)
	note = strings.TrimSpace(note)
	if tujuan == "" {
		return "", 0, "", errors.New("tujuan required")
	}
	if note == "" {
		return "", 0, "", errors.New("note required")
	}

	refID := "BTRF-" + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	saldoAfter, err := s.repo.TransferOut(ctx, actorID, bankID, amount, tujuan, note, refID)
	if err != nil {
		return "", 0, "", err
	}
	return refID, saldoAfter, tujuan, nil
}

func (s *BankService) TransferToBCAOperational(ctx context.Context, actorID int64, actorRole string, bankID int64, amount int64, adminFee int64, note string) (string, int64, int64, string, string, error) {
	note = strings.TrimSpace(note)
	if actorID <= 0 || bankID <= 0 || amount <= 0 {
		return "", 0, 0, "", "", errors.New("bank_id/amount invalid")
	}
	if adminFee < 0 {
		return "", 0, 0, "", "", errors.New("admin fee invalid")
	}
	if err := s.EnsureVisibleToRole(ctx, bankID, actorRole); err != nil {
		return "", 0, 0, "", "", err
	}
	if note == "" {
		note = "Transfer ke BCA OPERASIONAL 3432738881"
	}

	refID := "BTRF-" + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	sourceAfter, destinationAfter, destinationName, destinationAccount, err := s.repo.TransferToBCAOperational(ctx, actorID, bankID, amount, adminFee, note, refID)
	if err != nil {
		return "", 0, 0, "", "", err
	}
	return refID, sourceAfter, destinationAfter, destinationName, destinationAccount, nil
}

func (s *BankService) CreditProviderFromBankMutation(ctx context.Context, actorID int64, actorRole string, mutasiBankID int64, provider string, note string) (*repository.BankProviderAssignResult, error) {
	provider = strings.TrimSpace(provider)
	note = strings.TrimSpace(note)
	if actorID <= 0 || mutasiBankID <= 0 {
		return nil, errors.New("mutasi_bank_id invalid")
	}
	if provider == "" {
		return nil, errors.New("provider required")
	}
	return s.repo.CreditProviderFromBankMutation(ctx, actorID, mutasiBankID, provider, note, helper.IsAdminLikeRole(actorRole))
}

func (s *BankService) History(ctx context.Context, bankID int64, arah, refID, from, to, q string, prioritizeUnassigned bool, limit, offset int) ([]repository.BankMutasiRow, int64, error) {
	return s.repo.ListMutasi(ctx, bankID, arah, refID, from, to, q, prioritizeUnassigned, limit, offset)
}

func (s *BankService) UnpairedDebitMutasi(ctx context.Context, actorRole string, bankID int64, from, to, q string, limit, offset int) ([]repository.BankMutasiRow, int64, error) {
	return s.repo.ListUnpairedDebitMutasi(ctx, bankID, from, to, q, helper.IsAdminLikeRole(actorRole), limit, offset)
}

func (s *BankService) Kantor24LatestMutation(ctx context.Context, bankID int64) (*repository.BankMutasiRow, int64, error) {
	if bankID <= 0 {
		return nil, 0, errors.New("bank_id invalid")
	}
	bank, err := s.repo.GetVisible(ctx, bankID, true)
	if err != nil {
		return nil, 0, err
	}
	item, err := s.repo.LatestMutasi(ctx, bankID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, bank.Saldo, nil
	}
	return item, bank.Saldo, err
}
