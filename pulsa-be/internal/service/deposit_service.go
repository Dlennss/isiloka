package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/loketbayar"
)

type DepositService struct {
	repo     *repository.DepositRepository
	bankRepo *repository.BankRepository
	lbClient *loketbayar.Client
	p24      *provider.Pulsa24JamAdapter
}

const (
	minDepositAmount            int64 = 1000000
	minRetailDepositAmount      int64 = 100000
	minDepositVAAmount          int64 = 10000000
	specialDepositBankBCA8ID    int64 = 18
	specialDepositBankBCA8Email       = "makan@makin.com"
	depositTicketOfflineMessage       = "tidak bisa request tiket pada saat ini; tiket bisa dibuat jam 00.31"
	depositVAMethod                   = "va"
)

type depositVABank struct {
	Code string
	Name string
}

var depositVABanks = map[string]depositVABank{
	"VA24MAN":  {Code: "VA24MAN", Name: "VA Mandiri"},
	"VA24BRI":  {Code: "VA24BRI", Name: "VA BRI"},
	"VA24PRMT": {Code: "VA24PRMT", Name: "VA Permata"},
	"VA24DNMN": {Code: "VA24DNMN", Name: "VA Danamon"},
	"VA24OCBC": {Code: "VA24OCBC", Name: "VA OCBC"},
}

func NewDepositService(repo *repository.DepositRepository, bankRepo *repository.BankRepository, lbClient *loketbayar.Client) *DepositService {
	return &DepositService{repo: repo, bankRepo: bankRepo, lbClient: lbClient}
}

func (s *DepositService) SetPulsa24JamClient(client *provider.Pulsa24JamAdapter) {
	if s != nil {
		s.p24 = client
	}
}

func depositTicketLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

func depositTicketRequestOfflineAt(t time.Time) bool {
	now := t.In(depositTicketLocation())
	minuteOfDay := now.Hour()*60 + now.Minute()
	return minuteOfDay >= 23*60+30 || minuteOfDay <= 30
}

func (s *DepositService) CreateRequest(ctx context.Context, memberID int64, role string, bankID, amount int64, metode, buktiURL string) (*repository.DepositRequestRow, error) {
	if memberID <= 0 {
		return nil, errors.New("unauthorized")
	}
	if depositTicketRequestOfflineAt(time.Now()) {
		return nil, errors.New(depositTicketOfflineMessage)
	}
	metode = strings.TrimSpace(metode)
	if metode == "" {
		metode = "transfer"
	}
	if bankID <= 0 || amount <= 0 {
		return nil, errors.New("invalid payload")
	}
	if !helper.IsH2HRole(role) && amount < minRetailDepositAmount {
		return nil, errors.New("minimal topup Rp 100.000")
	}
	if helper.IsH2HRole(role) && amount < minDepositAmount {
		return nil, errors.New("minimal deposit Rp 1.000.000")
	}
	bank, err := s.bankRepo.GetVisible(ctx, bankID, false)
	if err != nil {
		return nil, errors.New("bank not found")
	}
	if !isMemberDepositBank(*bank) {
		return nil, errors.New("bank not found")
	}
	if !bank.Aktif {
		allowed, err := s.canUseSpecialDepositBank(ctx, memberID, bank.ID)
		if err != nil || !allowed {
			return nil, errors.New("bank tidak aktif")
		}
	}
	if strings.TrimSpace(bank.Nama) == "" || strings.TrimSpace(bank.NomorRekening) == "" || strings.TrimSpace(bank.AtasNama) == "" {
		return nil, errors.New("data rekening bank belum lengkap")
	}
	refID := "DTK-" + time.Now().Format("20060102150405") + "-" + strings.ToUpper(helper.RandHex(4))
	return s.repo.CreateTicketRequest(ctx, memberID, bankID, amount, metode, bank.Nama, bank.NomorRekening, bank.AtasNama, refID)
}

func (s *DepositService) CreateVARequest(ctx context.Context, memberID int64, _ string, amount int64, bankCode string) (*repository.DepositRequestRow, error) {
	if memberID <= 0 {
		return nil, errors.New("unauthorized")
	}
	if amount < minDepositVAAmount {
		return nil, errors.New("minimal deposit VA Rp 10.000.000")
	}
	if s == nil || s.lbClient == nil {
		return nil, errors.New("loketbayar belum aktif")
	}
	bank, ok := normalizeDepositVABank(bankCode)
	if !ok {
		return nil, errors.New("bank VA tidak valid")
	}
	activeCount, err := s.repo.ActiveNonQrisTicketCount(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if activeCount >= 5 {
		return nil, errors.New("maksimal 5 tiket aktif")
	}

	ticket, httpStatus, _, err := s.lbClient.DepositTicket(ctx, loketbayar.DepositTicketRequest{
		BankCode: bank.Code,
		Nominal:  amount,
	})
	if err != nil {
		return nil, err
	}
	if httpStatus < 200 || httpStatus >= 300 {
		return nil, errors.New("loketbayar tiket deposit gagal")
	}
	if !strings.EqualFold(strings.TrimSpace(ticket.Status), "SUKSES") {
		msg := strings.TrimSpace(ticket.Keterangan)
		if msg == "" {
			msg = "loketbayar tiket deposit gagal"
		}
		return nil, errors.New(msg)
	}
	ticketID := strings.TrimSpace(ticket.TicketID)
	if ticketID == "" {
		return nil, errors.New("loketbayar tidak mengirim nomor tiket")
	}
	ticketAmount, err := parseDepositVATicketAmount(ticketID)
	if err != nil {
		return nil, err
	}
	if ticketAmount < amount {
		return nil, errors.New("nominal tiket loketbayar lebih kecil dari nominal request")
	}
	if ticket.BankCode != "" && !strings.EqualFold(ticket.BankCode, bank.Code) {
		return nil, errors.New("kode bank tiket loketbayar tidak sesuai")
	}
	if strings.TrimSpace(ticket.Destination) == "" {
		return nil, errors.New("loketbayar tidak mengirim rekening tujuan")
	}

	accountName := strings.TrimSpace(ticket.AccountName)
	if accountName == "" {
		accountName = "SM PAY"
	}
	note := fmt.Sprintf("LoketBayar VA ticket %s bank %s requested_amount=%d unique_code=%d", ticketID, bank.Code, amount, ticketAmount-amount)
	return s.repo.CreateVARequest(ctx, memberID, amount, ticketAmount, bank.Code, bank.Name, strings.TrimSpace(ticket.Destination), accountName, ticketID, note)
}

func parseDepositVATicketAmount(ticketID string) (int64, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return 0, errors.New("nominal tiket loketbayar kosong")
	}
	n, err := strconv.ParseInt(ticketID, 10, 64)
	if err != nil || n <= 0 {
		return 0, errors.New("nominal tiket loketbayar tidak valid")
	}
	return n, nil
}

func normalizeDepositVABank(bankCode string) (depositVABank, bool) {
	code := strings.ToUpper(strings.TrimSpace(bankCode))
	bank, ok := depositVABanks[code]
	return bank, ok
}

func (s *DepositService) ActiveBanks(ctx context.Context, memberID int64) ([]repository.BankRow, error) {
	rows, err := s.bankRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]repository.BankRow, 0, len(rows))
	for _, row := range rows {
		if !isMemberDepositBank(row) {
			continue
		}
		out = append(out, row)
	}
	allowed, err := s.canUseSpecialDepositBank(ctx, memberID, specialDepositBankBCA8ID)
	if err != nil || !allowed || depositBankExists(out, specialDepositBankBCA8ID) {
		return out, err
	}
	bank, err := s.bankRepo.GetVisible(ctx, specialDepositBankBCA8ID, false)
	if err != nil {
		return out, nil
	}
	if !isMemberDepositBank(*bank) {
		return out, nil
	}
	out = append(out, *bank)
	return out, nil
}

func isMemberDepositBank(row repository.BankRow) bool {
	if row.AdminStaffOnly {
		return false
	}
	name := strings.ToUpper(strings.TrimSpace(row.Nama))
	if name == "" || name == "QRIS" || strings.Contains(name, "QRTP") {
		return false
	}
	return true
}

func (s *DepositService) canUseSpecialDepositBank(ctx context.Context, memberID, bankID int64) (bool, error) {
	if bankID != specialDepositBankBCA8ID {
		return false, nil
	}
	return s.repo.MemberEmailMatches(ctx, memberID, specialDepositBankBCA8Email)
}

func depositBankExists(rows []repository.BankRow, bankID int64) bool {
	for _, row := range rows {
		if row.ID == bankID {
			return true
		}
	}
	return false
}

func (s *DepositService) ListMemberRequests(ctx context.Context, memberID int64, limit int) ([]repository.DepositRequestRow, error) {
	return s.repo.ListByMember(ctx, memberID, limit)
}

func (s *DepositService) ConfirmTicketTransfer(ctx context.Context, memberID, reqID int64) (*repository.DepositRequestRow, error) {
	return s.repo.ConfirmTicketTransfer(ctx, memberID, reqID)
}

func (s *DepositService) CancelTicket(ctx context.Context, memberID, reqID int64) (*repository.DepositRequestRow, error) {
	row, err := s.repo.GetMemberTicket(ctx, memberID, reqID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return s.repo.CancelTicket(ctx, memberID, reqID)
	}
	cancelNote := ""
	if isDepositVATicket(row) {
		cancelNote, err = s.cancelLoketBayarDepositTicket(ctx, row.RefID)
		if err != nil {
			return nil, err
		}
	}
	return s.repo.CancelTicketWithNote(ctx, memberID, reqID, cancelNote)
}

func (s *DepositService) AdminList(ctx context.Context, status string, memberID int64, from, to string, limit, offset int, order string) ([]repository.DepositRequestRow, error) {
	return s.repo.AdminList(ctx, status, memberID, from, to, limit, offset, order)
}

func (s *DepositService) AdminListVA(ctx context.Context, status string, memberID int64, refID, from, to string, limit, offset int, order string) ([]repository.DepositRequestRow, error) {
	return s.repo.AdminListVA(ctx, status, memberID, refID, from, to, limit, offset, order)
}

func (s *DepositService) AdminApprove(ctx context.Context, reqID, adminID, approvedAmount int64, note string, bankRefIDs []string) (string, int64, error) {
	refID := "DEP-" + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	return s.repo.Approve(ctx, reqID, adminID, approvedAmount, note, refID, bankRefIDs)
}

func (s *DepositService) AdminApproveVA(ctx context.Context, reqID, adminID, approvedAmount int64, note string) (*repository.DepositRequestRow, error) {
	row, err := s.repo.GetVARequestByID(ctx, reqID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil && shouldCancelLoketBayarTicket(row) {
		cancelNote, cancelErr := s.cancelLoketBayarDepositTicketForManualApprove(ctx, row.RefID)
		if cancelErr != nil {
			return nil, cancelErr
		}
		note = appendDepositNote(note, cancelNote)
	}
	return s.repo.ApproveVA(ctx, reqID, adminID, approvedAmount, note)
}

func (s *DepositService) AutoApprovePendingFromBankMutations(ctx context.Context, limit int) (int, error) {
	return s.repo.AutoApprovePendingFromBankMutations(ctx, limit)
}

func (s *DepositService) AdminReject(ctx context.Context, reqID, adminID int64, note string) error {
	return s.repo.Reject(ctx, reqID, adminID, note)
}

func (s *DepositService) AdminRejectVA(ctx context.Context, reqID, adminID int64, note string) (*repository.DepositRequestRow, error) {
	row, err := s.repo.GetVARequestByID(ctx, reqID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil && shouldCancelLoketBayarTicket(row) {
		cancelNote, cancelErr := s.cancelLoketBayarDepositTicket(ctx, row.RefID)
		if cancelErr != nil {
			return nil, cancelErr
		}
		note = appendDepositNote(note, cancelNote)
	}
	return s.repo.RejectVA(ctx, reqID, adminID, note)
}

func (s *DepositService) AdminCreditInternal(ctx context.Context, memberID, amount int64, note string) (string, error) {
	if memberID <= 0 || amount <= 0 {
		return "", errors.New("member_id and amount must be > 0")
	}
	if strings.TrimSpace(note) == "" {
		note = "deposit"
	}
	refID := "DEP-" + time.Now().Format("20060102150405")
	return refID, s.repo.CreditInternal(ctx, memberID, amount, note, refID)
}

func isDepositVATicket(row *repository.DepositRequestRow) bool {
	return row != nil && strings.EqualFold(strings.TrimSpace(row.Metode), depositVAMethod)
}

func shouldCancelLoketBayarTicket(row *repository.DepositRequestRow) bool {
	if !isDepositVATicket(row) || strings.TrimSpace(row.RefID) == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(row.Status)) {
	case "ticket", "pending":
		return true
	default:
		return false
	}
}

func (s *DepositService) cancelLoketBayarDepositTicket(ctx context.Context, ticketID string) (string, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return "", errors.New("kode tiket LoketBayar kosong")
	}
	if s == nil || s.lbClient == nil {
		return "", errors.New("loketbayar belum aktif")
	}
	resp, httpStatus, _, err := s.lbClient.CancelDepositTicket(ctx, loketbayar.DepositTicketCancelRequest{TicketID: ticketID})
	if err != nil {
		return "", err
	}
	msg := normalizeDepositNote(resp.Keterangan)
	if msg == "" {
		msg = strings.TrimSpace(resp.Status)
	}
	if httpStatus < 200 || httpStatus >= 300 {
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", httpStatus)
		}
		return "", fmt.Errorf("loketbayar cancel tiket gagal: %s", msg)
	}
	if !strings.EqualFold(strings.TrimSpace(resp.Status), "SUKSES") {
		if msg == "" {
			msg = "response tidak sukses"
		}
		return "", fmt.Errorf("loketbayar cancel tiket gagal: %s", msg)
	}
	return fmt.Sprintf("LoketBayar VA cancel SUKSES kode_tiket=%s response=%s", ticketID, msg), nil
}

func (s *DepositService) cancelLoketBayarDepositTicketForManualApprove(ctx context.Context, ticketID string) (string, error) {
	cancelNote, err := s.cancelLoketBayarDepositTicket(ctx, ticketID)
	if err == nil {
		return cancelNote, nil
	}
	msg := strings.TrimSpace(err.Error())
	msg = strings.TrimPrefix(msg, "loketbayar cancel tiket gagal:")
	msg = normalizeDepositNote(msg)
	if looksLikeLoketBayarTicketAlreadyClosed(msg) {
		return fmt.Sprintf("LoketBayar VA cancel GAGAL kode_tiket=%s response=%s", strings.TrimSpace(ticketID), msg), nil
	}
	return "", err
}

func looksLikeLoketBayarTicketAlreadyClosed(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	return strings.Contains(up, "TIDAK DITEMUKAN") ||
		strings.Contains(up, "NOT FOUND") ||
		strings.Contains(up, "SUDAH TIDAK") ||
		strings.Contains(up, "TIDAK DALAM STATUS PENDING")
}

func appendDepositNote(base, extra string) string {
	base = strings.TrimSpace(base)
	extra = strings.TrimSpace(extra)
	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}
	return base + " | " + extra
}

func normalizeDepositNote(note string) string {
	note = strings.Join(strings.Fields(strings.TrimSpace(note)), " ")
	if len(note) > 500 {
		note = note[:500]
	}
	return note
}
