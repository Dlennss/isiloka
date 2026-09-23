package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

type RetailService struct {
	repo       *repository.RetailRepository
	bankRepo   *repository.BankRepository
	p24Client  provider.Client
	p24Catalog *provider.Pulsa24JamAdapter
}

type RetailRegisterDownlineInput struct {
	Email     string
	Nama      string
	Phone     string
	StoreName string
	Password  string
	Role      string
}

type RetailWithdrawCreateInput struct {
	Amount        int64
	SourceType    string
	BankName      string
	AccountName   string
	AccountNumber string
	Note          string
}

func NewRetailService(repo *repository.RetailRepository, bankRepo *repository.BankRepository, clients ...provider.Client) *RetailService {
	s := &RetailService{repo: repo, bankRepo: bankRepo}
	for _, client := range clients {
		if client != nil && strings.EqualFold(client.Name(), provider.Pulsa24JamProviderName) {
			s.p24Client = client
			s.p24Catalog, _ = client.(*provider.Pulsa24JamAdapter)
			break
		}
	}
	return s
}

func (s *RetailService) ListDownlines(ctx context.Context, actorID int64) ([]repository.RetailDownlineRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsRetailRole(actor.Role) {
		return nil, errors.New("retail only")
	}
	if actor.Role != helper.RoleRetailMaster && actor.Role != helper.RoleRetailAgent && actor.Role != helper.RoleRetailMarketing {
		return []repository.RetailDownlineRow{}, nil
	}
	return s.repo.ListDownlines(ctx, actor)
}

func (s *RetailService) RegisterDownline(ctx context.Context, actorID int64, in RetailRegisterDownlineInput) (int64, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return 0, err
	}
	if !helper.IsRetailRole(actor.Role) {
		return 0, errors.New("retail only")
	}

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Nama = strings.TrimSpace(in.Nama)
	in.Phone = normalizeMemberPhone(in.Phone)
	in.StoreName = strings.TrimSpace(in.StoreName)
	in.Password = strings.TrimSpace(in.Password)
	in.Role = helper.NormalizeRole(in.Role)

	if in.Email == "" || in.Nama == "" || len(in.Password) < 8 {
		return 0, errors.New("email, nama, dan password valid wajib diisi")
	}
	if in.Role != helper.RoleUser && in.Role != helper.RoleRetailAgent {
		return 0, errors.New("role retail bawahan tidak valid")
	}
	switch actor.Role {
	case helper.RoleRetailMaster:
		// master boleh buat agent atau user
	case helper.RoleRetailAgent:
		if in.Role != helper.RoleUser {
			return 0, errors.New("agent hanya boleh menambahkan user")
		}
	case helper.RoleRetailMarketing:
		if in.Role != helper.RoleRetailAgent {
			return 0, errors.New("marketing hanya boleh mendaftarkan agent")
		}
	default:
		return 0, errors.New("role tidak boleh menambahkan downline")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	createIn := repository.UserCreateInput{
		Email:        in.Email,
		Nama:         in.Nama,
		Phone:        in.Phone,
		StoreName:    in.StoreName,
		PasswordHash: string(passHash),
		Role:         in.Role,
		Aktif:        true,
	}
	createIn.RetailAgentCommissionRp, createIn.RetailMasterCommissionRp = helper.ApplyRetailCommissionDefaults(createIn.Role, createIn.RetailAgentCommissionRp, createIn.RetailMasterCommissionRp)
	switch actor.Role {
	case helper.RoleRetailMaster:
		createIn.RetailMasterID = &actor.MemberID
	case helper.RoleRetailAgent:
		createIn.RetailAgentID = &actor.MemberID
		if actor.RetailMasterID != nil && *actor.RetailMasterID > 0 {
			createIn.RetailMasterID = actor.RetailMasterID
		}
	case helper.RoleRetailMarketing:
		createIn.MarketingID = &actor.MemberID
	}
	return s.repo.CreateRetailChild(ctx, createIn)
}

func (s *RetailService) ListCommissions(ctx context.Context, actorID int64, limit, offset int) ([]repository.RetailCommissionLedgerRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsRetailRole(actor.Role) {
		return nil, errors.New("retail only")
	}
	return s.repo.ListCommissionLedger(ctx, actorID, limit, offset)
}

func (s *RetailService) CommissionSummary(ctx context.Context, actorID int64) (*repository.RetailCommissionSummaryRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsRetailRole(actor.Role) {
		return nil, errors.New("retail only")
	}
	return s.repo.GetCommissionSummary(ctx, actorID)
}

func (s *RetailService) CreateWithdrawRequest(ctx context.Context, actorID int64, in RetailWithdrawCreateInput) (*repository.RetailWithdrawRequestRow, error) {
	// Server lama bisa belum memiliki kolom sumber dana kredit. Pastikan skema
	// tersedia tepat sebelum penarikan dibuat agar error PostgreSQL tidak sampai
	// menjadi pesan "internal error" di aplikasi agent.
	if err := s.repo.EnsureWithdrawSchema(ctx); err != nil {
		return nil, errors.New("sistem penarikan sedang disiapkan. Silakan coba kembali")
	}

	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsRetailRole(actor.Role) {
		return nil, errors.New("retail only")
	}

	in.BankName = strings.TrimSpace(in.BankName)
	in.AccountName = strings.TrimSpace(in.AccountName)
	in.AccountNumber = strings.TrimSpace(in.AccountNumber)
	in.Note = strings.TrimSpace(in.Note)
	in.SourceType = strings.TrimSpace(strings.ToLower(in.SourceType))
	if in.SourceType == "" {
		in.SourceType = "main_balance"
	}
	if in.SourceType != "main_balance" && in.SourceType != "credit" {
		return nil, errors.New("sumber penarikan tidak valid")
	}
	if in.SourceType == "credit" {
		return nil, errors.New("saldo kredit sudah tidak digunakan; gunakan saldo utama")
	}
	if in.Amount <= 0 {
		return nil, errors.New("amount harus > 0")
	}
	if in.BankName == "" || in.AccountName == "" || in.AccountNumber == "" {
		return nil, errors.New("data rekening wajib lengkap")
	}

	// Penarikan saldo utama memakai jalur PAY H2H yang sama seperti pembelian
	// e-wallet. SKU divalidasi sebelum saldo ditahan agar request yang tidak
	// tersedia di katalog provider tidak membuat saldo agent berkurang.
	selection, err := s.resolveWithdrawProduct(ctx, in.BankName, in.Amount)
	if err != nil {
		return nil, err
	}
	providerProduct := selection.SKU
	providerQty := selection.Qty
	refID := fmt.Sprintf("RWD-%s-%s", time.Now().Format("20060102150405"), strings.ToUpper(helper.RandHex(4)))
	item, err := s.repo.CreateWithdrawRequest(ctx, actorID, in.Amount, in.SourceType, in.BankName, in.AccountName, in.AccountNumber, refID, in.Note)
	if err != nil {
		return nil, err
	}
	resp, payErr := s.p24Client.Pay(ctx, provider.PayRequest{
		Command: "PAY",
		Product: providerProduct,
		Dest:    in.AccountNumber,
		Qty:     providerQty,
		RefID:   refID,
	})
	body := ""
	msg := ""
	if resp != nil {
		body = strings.TrimSpace(resp.Body)
		msg = strings.TrimSpace(resp.Message)
	}
	if payErr != nil || (resp != nil && resp.HTTPStatus != 200) || retailPulsa24JamLooksRejected(body, msg) || (!retailPulsa24JamLooksSuccess(body, msg) && !retailPulsa24JamLooksPending(body, msg)) {
		reason := strings.TrimSpace(msg)
		if reason == "" {
			reason = strings.TrimSpace(body)
		}
		if reason == "" && payErr != nil {
			reason = payErr.Error()
		}
		if reason == "" {
			reason = "penarikan ditolak Pulsa24Jam"
		}
		_ = s.repo.RejectWithdrawRequest(ctx, item.ID, actorID, reason)
		if retailWithdrawProviderUnavailable(reason) {
			return nil, errors.New("Pulsa24Jam belum dapat dihubungi. Saldo kredit telah dikembalikan")
		}
		// Detail provider tetap dicatat pada request yang ditolak dan log controller.
		// Agent cukup menerima pesan yang aman serta mudah dipahami.
		return nil, errors.New("Tidak bisa ditarik karena ada kesalahan")
	}
	status := "processing_provider"
	if retailPulsa24JamLooksSuccess(body, msg) {
		status = "approved"
	}
	note := fmt.Sprintf("dikirim ke Pulsa24Jam product=%s qty=%d dest=%s", providerProduct, providerQty, in.AccountNumber)
	if msg != "" {
		note += " | " + msg
	}
	if err := s.repo.UpdateWithdrawRequestProviderStatus(ctx, refID, status, note); err != nil {
		return nil, err
	}
	item.Status = status
	item.Note = note
	return item, nil
}

var retailWithdrawNominalPattern = regexp.MustCompile(`(?i)(?:rp\s*)?(\d{1,3}(?:[.\s]\d{3})+|\d{4,})`)

type retailWithdrawProviderProduct struct {
	SKU string
	Qty int64
}

func (s *RetailService) resolveWithdrawProduct(ctx context.Context, destination string, amount int64) (retailWithdrawProviderProduct, error) {
	if s == nil || s.p24Catalog == nil {
		return retailWithdrawProviderProduct{}, errors.New("katalog Pulsa24Jam untuk penarikan belum aktif")
	}
	items, err := s.p24Catalog.Products(ctx, "")
	if err != nil {
		return retailWithdrawProviderProduct{}, errors.New("katalog Pulsa24Jam belum dapat diakses. Silakan coba kembali")
	}
	destinationKey := retailWithdrawDestinationKey(destination)
	var fixedFallback string
	var openAmountFallback string
	for _, item := range items {
		if !item.Active || !retailWithdrawProductMatches(item, destinationKey) {
			continue
		}

		// Produk fixed Direct lebih aman dari produk Promo atau SKU e-wallet
		// generik. Contoh: DANA 50.000 = UDDND50 dan GoPay 50.000 = UDGY50.
		if retailWithdrawProductNominal(item.Name) == amount {
			sku := strings.TrimSpace(item.SKU)
			if sku == "" {
				continue
			}
			if product, qty, ok := retailWithdrawOpenWalletRequest(item, amount); ok {
				return retailWithdrawProviderProduct{SKU: product, Qty: qty}, nil
			}
			group := strings.ToUpper(strings.TrimSpace(item.GroupName))
			if strings.Contains(group, "DIRECT") {
				return retailWithdrawProviderProduct{SKU: sku, Qty: 1}, nil
			}
			if fixedFallback == "" {
				fixedFallback = sku
			}
			continue
		}
		if retailWithdrawOpenAmountProduct(item) || retailWithdrawDestinationKey(item.SKU) == destinationKey {
			if openAmountFallback == "" {
				openAmountFallback = strings.TrimSpace(item.SKU)
			}
		}
	}
	if fixedFallback != "" {
		return retailWithdrawProviderProduct{SKU: fixedFallback, Qty: 1}, nil
	}
	if openAmountFallback != "" {
		return retailWithdrawProviderProduct{SKU: openAmountFallback, Qty: amount}, nil
	}
	return retailWithdrawProviderProduct{}, fmt.Errorf("produk %s nominal %d belum tersedia di Pulsa24Jam", strings.TrimSpace(destination), amount)
}

// Pulsa24Jam memproses produk fixed DANA dan GoPay reguler melalui command
// bebas nominal. Ini sama dengan request checkout aplikasi yang sudah stabil:
// product DANA/GOPAY dan qty berisi nominal rupiah, bukan SKU fixed dengan qty 1.
func retailWithdrawOpenWalletRequest(item provider.Pulsa24JamProduct, amount int64) (string, int64, bool) {
	sku := strings.ToUpper(strings.TrimSpace(item.SKU))
	name := strings.ToUpper(strings.TrimSpace(item.Name))
	if amount <= 0 {
		return "", 0, false
	}
	switch {
	case strings.HasPrefix(sku, "UDDND") && strings.Contains(name, "DANA"):
		return "DANA", amount, true
	case (strings.HasPrefix(sku, "UDGP") || strings.HasPrefix(sku, "UDGY")) && strings.Contains(name, "GOPAY") && !strings.Contains(name, "DRIVER"):
		return "GOPAY", amount, true
	default:
		return "", 0, false
	}
}

func retailWithdrawDestinationKey(raw string) string {
	key := strings.ToUpper(strings.TrimSpace(raw))
	return strings.NewReplacer(" ", "", "-", "", "_", "", ".", "").Replace(key)
}

func retailWithdrawProductMatches(item provider.Pulsa24JamProduct, destinationKey string) bool {
	if destinationKey == "" {
		return false
	}
	for _, value := range []string{item.BrandName, item.Name, item.GroupName, item.CategoryName, item.SKU} {
		if strings.Contains(retailWithdrawDestinationKey(value), destinationKey) {
			return true
		}
	}
	return false
}

func retailWithdrawOpenAmountProduct(item provider.Pulsa24JamProduct) bool {
	priceType := strings.ToUpper(strings.TrimSpace(item.PriceType))
	return strings.Contains(priceType, "OPEN") || strings.Contains(priceType, "BEBAS") || strings.Contains(priceType, "NOMINAL")
}

func retailWithdrawProductNominal(name string) int64 {
	match := retailWithdrawNominalPattern.FindStringSubmatch(name)
	if len(match) < 2 {
		return 0
	}
	normalized := strings.NewReplacer(".", "", " ", "").Replace(match[1])
	var amount int64
	_, _ = fmt.Sscan(normalized, &amount)
	return amount
}

func retailWithdrawProviderUnavailable(reason string) bool {
	lower := strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(lower, "context deadline") ||
		strings.Contains(lower, "dial tcp") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "i/o timeout") ||
		strings.Contains(lower, " eof")
}

func retailWithdrawPulsa24JamProduct(bankName string) string {
	normalized := strings.ToUpper(strings.TrimSpace(bankName))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "", ".", "")
	key := replacer.Replace(normalized)
	switch {
	case strings.Contains(key, "GOPAY") || strings.Contains(key, "GOJEK") || key == "GPAY":
		return "GOPAY"
	case strings.Contains(key, "DANA"):
		return "DANA"
	case strings.Contains(key, "OVO"):
		return "OVO"
	case strings.Contains(key, "SHOPEE"):
		return "SHOPEEPAY"
	case strings.Contains(key, "LINKAJA"):
		return "LINKAJA"
	case strings.Contains(key, "ISAKU"):
		return "ISAKU"
	default:
		return key
	}
}

func retailPulsa24JamLooksSuccess(values ...string) bool {
	upper := strings.ToUpper(strings.Join(values, " "))
	return strings.Contains(upper, "SUKSES") ||
		strings.Contains(upper, "SUCCESS") ||
		strings.Contains(upper, `"OK":TRUE`) ||
		strings.Contains(upper, `"SUCCESS":TRUE`) ||
		strings.Contains(upper, `"STATUS":2`) ||
		strings.Contains(upper, `"STATUS":"2"`) ||
		strings.Contains(upper, `"STATUS":"SUCCESS"`) ||
		strings.Contains(upper, `"RC":"00"`)
}

func retailPulsa24JamLooksPending(values ...string) bool {
	upper := strings.ToUpper(strings.Join(values, " "))
	return strings.Contains(upper, "PENDING") ||
		strings.Contains(upper, "DIPROSES") ||
		strings.Contains(upper, "PROCESSING")
}

func retailPulsa24JamLooksRejected(values ...string) bool {
	upper := strings.ToUpper(strings.Join(values, " "))
	return strings.Contains(upper, "GAGAL") ||
		strings.Contains(upper, "FAILED") ||
		strings.Contains(upper, "DITOLAK") ||
		strings.Contains(upper, `"STATUS":3`) ||
		strings.Contains(upper, `"STATUS":"3"`) ||
		strings.Contains(upper, `"OK":FALSE`) ||
		strings.Contains(upper, `"SUCCESS":FALSE`) ||
		strings.Contains(upper, `"STATUS":"FAILED"`)
}

func (s *RetailService) ListOwnWithdrawRequests(ctx context.Context, actorID int64, limit, offset int) ([]repository.RetailWithdrawRequestRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsRetailRole(actor.Role) {
		return nil, errors.New("retail only")
	}
	items, err := s.repo.ListWithdrawRequestsByMember(ctx, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.refreshWithdrawStatuses(ctx, items), nil
}

// refreshWithdrawStatuses keeps withdrawals on the same automatic H2H
// path as product purchases. Webhooks remain the primary finalizer, while
// STATUS-PAY resolves a transaction when its callback arrives late.
func (s *RetailService) refreshWithdrawStatuses(ctx context.Context, items []repository.RetailWithdrawRequestRow) []repository.RetailWithdrawRequestRow {
	if s == nil || s.p24Client == nil {
		return items
	}
	for index := range items {
		item := &items[index]
		if item.Status != "processing_provider" {
			continue
		}
		if item.UpdatedAt != nil && time.Since(*item.UpdatedAt) < 10*time.Second {
			continue
		}
		product := retailWithdrawProductFromNote(item.Note)
		qty := retailWithdrawQtyFromNote(item.Note, item.Amount)
		if product == "" || item.RefID == "" || item.AccountNumber == "" || item.Amount <= 0 {
			continue
		}

		response, err := s.p24Client.Pay(ctx, provider.PayRequest{
			Command: "STATUS-PAY",
			Product: product,
			Dest:    item.AccountNumber,
			Qty:     qty,
			RefID:   item.RefID,
		})
		if err != nil || response == nil || response.HTTPStatus < 200 || response.HTTPStatus >= 300 {
			continue
		}
		message := strings.TrimSpace(response.Message)
		if retailPulsa24JamLooksSuccess(response.Body, message) {
			note := "Pulsa24Jam berhasil melalui STATUS-PAY"
			if message != "" {
				note += ": " + message
			}
			if s.repo.UpdateWithdrawRequestProviderStatus(ctx, item.RefID, "approved", note) == nil {
				item.Status = "approved"
				item.Note = note
			}
			continue
		}
		if retailPulsa24JamLooksRejected(response.Body, message) {
			reason := message
			if reason == "" {
				reason = "penarikan Pulsa24Jam gagal"
			}
			if s.repo.RejectWithdrawRequest(ctx, item.ID, item.MemberID, reason) == nil {
				item.Status = "rejected"
				item.RejectReason = reason
			}
		}
	}
	return items
}

func retailWithdrawProductFromNote(note string) string {
	for _, field := range strings.Fields(note) {
		before, value, ok := strings.Cut(field, "product=")
		if before == "" && ok && value != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func retailWithdrawQtyFromNote(note string, fallback int64) int64 {
	for _, field := range strings.Fields(note) {
		before, value, ok := strings.Cut(field, "qty=")
		if before != "" || !ok || value == "" {
			continue
		}
		var qty int64
		if _, err := fmt.Sscan(value, &qty); err == nil && qty > 0 {
			return qty
		}
	}
	return fallback
}

func (s *RetailService) AdminListWithdrawRequests(ctx context.Context, status, q string, limit, offset int) ([]repository.RetailWithdrawRequestRow, error) {
	return s.repo.AdminListWithdrawRequests(ctx, status, q, limit, offset)
}

func (s *RetailService) ListWithdrawSourceBanks(ctx context.Context) ([]repository.BankRow, error) {
	if s.bankRepo == nil {
		return nil, errors.New("repository bank tidak tersedia")
	}
	return s.bankRepo.List(ctx, false)
}

func (s *RetailService) AdminApproveWithdrawRequest(ctx context.Context, reqID, actorID int64, actorRole string, bankID, fee int64, note string) error {
	if reqID <= 0 || actorID <= 0 {
		return errors.New("request invalid")
	}
	if bankID <= 0 {
		return errors.New("rekening sumber wajib dipilih")
	}
	if fee < 0 {
		return errors.New("fee tidak valid")
	}
	if s.bankRepo != nil {
		if _, err := s.bankRepo.GetVisible(ctx, bankID, helper.IsAdminLikeRole(actorRole)); err != nil {
			return err
		}
	}
	err := s.repo.ApproveWithdrawRequest(ctx, reqID, actorID, bankID, fee, strings.TrimSpace(note))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("withdraw request tidak ditemukan atau bukan pending")
	}
	return err
}

func (s *RetailService) AdminRejectWithdrawRequest(ctx context.Context, reqID, actorID int64, reason string) error {
	if reqID <= 0 || actorID <= 0 {
		return errors.New("request invalid")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("alasan reject wajib diisi")
	}
	err := s.repo.RejectWithdrawRequest(ctx, reqID, actorID, reason)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("withdraw request tidak ditemukan")
	}
	return err
}
