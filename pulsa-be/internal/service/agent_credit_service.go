package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

const (
	minimumAgentCreditAmount int64 = 500000
	maximumAgentCreditAmount int64 = 2000000
)

type AgentCreditService struct {
	repo *repository.AgentCreditRepository
}

func NewAgentCreditService(repo *repository.AgentCreditRepository) *AgentCreditService {
	return &AgentCreditService{repo: repo}
}

type AgentCreditSubmitInput struct {
	ID              int64          `json:"id"`
	MemberID        int64          `json:"member_id"`
	RequestedAmount int64          `json:"requested_amount"`
	ApplicantData   map[string]any `json:"applicant_data"`
	DocumentData    map[string]any `json:"document_data"`
	AgentSignature  string         `json:"agent_signature"`
	TermsAccepted   bool           `json:"terms_accepted"`
}

type AgentCreditDecisionInput struct {
	ID                int64    `json:"id"`
	Decision          string   `json:"decision"`
	ApprovedAmount    int64    `json:"approved_amount"`
	Note              string   `json:"note"`
	ReviewerMode      string   `json:"reviewer_mode"`
	SignatureData     string   `json:"signature_data"`
	RiskLevel         string   `json:"risk_level"`
	RiskScore         int64    `json:"risk_score"`
	RevisionDocuments []string `json:"revision_documents"`
}

type AgentCreditPaymentInput struct {
	ApplicationID int64          `json:"application_id"`
	MemberID      int64          `json:"member_id"`
	RefID         string         `json:"ref_id"`
	Amount        int64          `json:"amount"`
	Note          string         `json:"note"`
	PaymentMethod string         `json:"payment_method"`
	PaymentProof  map[string]any `json:"payment_proof"`
}

type AgentCreditTransferInput struct {
	ApplicationID int64 `json:"application_id"`
	Amount        int64 `json:"amount"`
}

type AgentCreditRankChangeInput struct {
	MemberID    int64  `json:"member_id"`
	RankID      int64  `json:"rank_id"`
	LimitAmount int64  `json:"limit_amount"`
	Reason      string `json:"reason"`
}

type AgentCreditLoanStatusInput struct {
	ApplicationID int64  `json:"application_id"`
	Suspended     bool   `json:"suspended"`
	Reason        string `json:"reason"`
}

type AgentCreditManualInput struct {
	ApplicationID   int64          `json:"application_id"`
	MemberID        int64          `json:"member_id"`
	RequestedAmount int64          `json:"requested_amount"`
	Action          string         `json:"action"`
	Note            string         `json:"note"`
	ApplicantData   map[string]any `json:"applicant_data"`
	DocumentData    map[string]any `json:"document_data"`
}

func (s *AgentCreditService) UpdateManualApplication(ctx context.Context, auth helper.AuthInfo, in AgentCreditManualInput) error {
	if !canManageManualCredit(auth.Role) {
		return errors.New("input data agent hanya dapat dilakukan operator")
	}
	if in.ApplicationID <= 0 || in.MemberID <= 0 {
		return errors.New("data migrasi tidak valid")
	}
	if in.RequestedAmount < minimumAgentCreditAmount || in.RequestedAmount > maximumAgentCreditAmount {
		return errors.New("nominal kredit harus antara Rp500.000 dan Rp2.000.000")
	}
	if err := validateAgentCreditDocuments(in.DocumentData); err != nil {
		return fmt.Errorf("dokumen migrasi: %w", err)
	}
	if in.ApplicantData == nil {
		in.ApplicantData = map[string]any{}
	}
	in.ApplicantData["entry_source"] = "operator_manual"
	in.ApplicantData["entry_source_label"] = "Migrasi Data Operator"
	in.ApplicantData["manual_entry"] = true
	in.ApplicantData["manual_entry_operator_id"] = auth.MemberID
	in.ApplicantData["manual_entry_updated_at"] = time.Now().UTC().Format(time.RFC3339)
	in.ApplicantData["system_validation_status"] = "operator_verified"
	stampAgentCreditDocumentValidation(in.ApplicantData)
	return s.repo.UpdateManualApplication(ctx, in.ApplicationID, in.MemberID, in.RequestedAmount, in.ApplicantData, in.DocumentData)
}

type AgentCreditManualResult struct {
	Item          *repository.AgentCreditApplication `json:"item"`
	Approved      bool                               `json:"approved"`
	ApprovalError string                             `json:"approval_error,omitempty"`
}

func canManageManualCredit(role string) bool {
	normalized := helper.NormalizeRole(role)
	return normalized == helper.RoleAdmin || normalized == helper.RoleStaff || normalized == helper.RoleRetailAnalyst
}

func (s *AgentCreditService) ListManualEntryAgents(ctx context.Context, auth helper.AuthInfo, search string, limit int) ([]repository.AgentCreditManualAgent, error) {
	if !canManageManualCredit(auth.Role) {
		return nil, errors.New("input data agent hanya dapat dilakukan operator")
	}
	return s.repo.ListManualEntryAgents(ctx, search, limit)
}

func (s *AgentCreditService) CreateManualApplication(ctx context.Context, auth helper.AuthInfo, in AgentCreditManualInput) (*AgentCreditManualResult, error) {
	if !canManageManualCredit(auth.Role) {
		return nil, errors.New("input data agent hanya dapat dilakukan operator")
	}
	if in.MemberID <= 0 {
		return nil, errors.New("akun agent wajib dipilih")
	}
	if in.RequestedAmount == 0 {
		in.RequestedAmount = minimumAgentCreditAmount
	}
	if in.RequestedAmount < minimumAgentCreditAmount || in.RequestedAmount > maximumAgentCreditAmount {
		return nil, errors.New("nominal kredit harus antara Rp500.000 dan Rp2.000.000")
	}
	if err := validateAgentCreditDocuments(in.DocumentData); err != nil {
		return nil, fmt.Errorf("dokumen migrasi: %w", err)
	}
	if openID, err := s.repo.FindOpenEarlyApplicationID(ctx, in.MemberID); err != nil {
		return nil, err
	} else if openID > 0 {
		return nil, fmt.Errorf("agent masih memiliki pengajuan aktif KRD-%08d", openID)
	}
	if err := s.repo.EnsureAgentCanSubmitNextApplication(ctx, in.MemberID); err != nil {
		return nil, err
	}
	if in.ApplicantData == nil {
		in.ApplicantData = map[string]any{}
	}
	in.ApplicantData["entry_source"] = "operator_manual"
	in.ApplicantData["entry_source_label"] = "Migrasi Data Operator"
	in.ApplicantData["manual_entry"] = true
	in.ApplicantData["manual_entry_operator_id"] = auth.MemberID
	in.ApplicantData["manual_entry_at"] = time.Now().UTC().Format(time.RFC3339)
	in.ApplicantData["system_validation_status"] = "operator_verified"
	stampAgentCreditDocumentValidation(in.ApplicantData)

	item, err := s.repo.CreateManualApplication(ctx, repository.AgentCreditApplicationInput{
		MemberID:        in.MemberID,
		RequestedAmount: in.RequestedAmount,
		ApplicantData:   in.ApplicantData,
		DocumentData:    in.DocumentData,
	})
	if err != nil {
		return nil, err
	}
	result := &AgentCreditManualResult{Item: item}
	if strings.EqualFold(strings.TrimSpace(in.Action), "approve") {
		approved, approvalErr := s.DecideApplication(ctx, auth, AgentCreditDecisionInput{
			ID:             item.ID,
			Decision:       "approve",
			ApprovedAmount: in.RequestedAmount,
			Note:           fallbackNote(strings.TrimSpace(in.Note), "Data manual operator diverifikasi dan disetujui."),
			RiskLevel:      "low",
		})
		if approvalErr != nil {
			result.ApprovalError = approvalErr.Error()
			return result, nil
		}
		result.Item = approved
		result.Approved = true
	}
	return result, nil
}

func (s *AgentCreditService) ListTeamActivity(ctx context.Context, auth helper.AuthInfo, actorRole string, limit int) ([]repository.AgentCreditTeamActivity, error) {
	if helper.NormalizeRole(auth.Role) != helper.RoleAdmin {
		return nil, errors.New("super admin only")
	}
	actorRole = strings.TrimSpace(strings.ToLower(actorRole))
	if actorRole != "" && actorRole != "marketing" && actorRole != "operator_credit" && actorRole != "super_admin" {
		return nil, errors.New("filter role tidak valid")
	}
	return s.repo.ListTeamActivity(ctx, actorRole, limit)
}

func (s *AgentCreditService) ListInactiveAgents(ctx context.Context, auth helper.AuthInfo, days int, search string, limit int) ([]repository.AgentCreditInactiveAgent, error) {
	role := helper.NormalizeRole(auth.Role)
	if role != helper.RoleAdmin && role != helper.RoleRetailAnalyst {
		return nil, errors.New("admin/operator kredit only")
	}
	return s.repo.ListInactiveAgents(ctx, days, search, limit)
}

func (s *AgentCreditService) ListAgentTransactions(ctx context.Context, auth helper.AuthInfo, status, search string, limit int) ([]repository.AgentCreditAgentTransaction, error) {
	role := helper.NormalizeRole(auth.Role)
	if role != helper.RoleAdmin && role != helper.RoleRetailAnalyst {
		return nil, errors.New("admin/operator kredit only")
	}
	return s.repo.ListAgentTransactions(ctx, status, search, limit)
}

func isCreditReviewer(role string) bool {
	normalized := helper.NormalizeRole(role)
	return normalized == helper.RoleAdmin ||
		normalized == helper.RoleStaff ||
		normalized == helper.RoleRetailAnalyst
}

func canViewCreditApplications(role string) bool {
	normalized := helper.NormalizeRole(role)
	return isCreditReviewer(normalized) || normalized == helper.RoleRetailMarketing || normalized == helper.RoleRetailMaster
}

func (s *AgentCreditService) SubmitApplication(ctx context.Context, auth helper.AuthInfo, in AgentCreditSubmitInput) (*repository.AgentCreditApplication, error) {
	role := helper.NormalizeRole(auth.Role)
	targetMemberID := auth.MemberID
	if role == helper.RoleRetailAgent {
		if in.MemberID > 0 && in.MemberID != auth.MemberID {
			return nil, errors.New("agent hanya bisa mengajukan untuk akun sendiri")
		}
	} else if isCreditReviewer(role) {
		if in.MemberID <= 0 {
			return nil, errors.New("member_id peminjam wajib diisi")
		}
		targetMemberID = in.MemberID
	} else {
		return nil, errors.New("tidak punya akses pengajuan kredit")
	}
	if in.ApplicantData == nil {
		in.ApplicantData = map[string]any{}
	}
	if !in.TermsAccepted {
		if accepted, ok := in.ApplicantData["terms_accepted"].(bool); ok {
			in.TermsAccepted = accepted
		}
	}
	if role == helper.RoleRetailAgent {
		if !in.TermsAccepted {
			return nil, errors.New("syarat dan ketentuan wajib disetujui")
		}
		in.ApplicantData["terms_accepted"] = in.TermsAccepted
		in.ApplicantData["terms_version"] = "pulsakilat-agent-credit-2026-07"
	} else if in.TermsAccepted {
		in.ApplicantData["terms_accepted"] = true
		in.ApplicantData["terms_version"] = "pulsakilat-agent-credit-2026-07"
	}
	if role == helper.RoleRetailAgent {
		profile, err := s.repo.GetMemberCreditProfile(ctx, auth.MemberID)
		if err != nil {
			return nil, err
		}
		if in.RequestedAmount <= 0 {
			in.RequestedAmount = profile.LimitAmount
		}
	}
	if in.RequestedAmount <= 0 {
		return nil, errors.New("nominal kredit wajib diisi")
	}
	if in.RequestedAmount < minimumAgentCreditAmount {
		return nil, errors.New("nominal kredit minimal Rp500.000")
	}
	if in.RequestedAmount > maximumAgentCreditAmount {
		return nil, errors.New("nominal kredit maksimal Rp2.000.000")
	}
	if in.DocumentData == nil {
		in.DocumentData = map[string]any{}
	}
	hasStoredAgentSignature := false
	validateFullSubmission := role == helper.RoleRetailAgent || (isCreditReviewer(role) && in.ID <= 0)
	if validateFullSubmission {
		validationInput := in
		if role == helper.RoleRetailAgent && in.ID > 0 {
			storedDocuments, storedSignature, err := s.repo.GetApplicationRevisionStateForMember(ctx, in.ID, targetMemberID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if err == nil {
				hasStoredAgentSignature = storedSignature
				mergedDocuments := make(map[string]any, len(storedDocuments)+len(in.DocumentData))
				for key, value := range storedDocuments {
					mergedDocuments[key] = value
				}
				for key, value := range in.DocumentData {
					mergedDocuments[key] = value
				}
				if selfie, ok := mergedDocuments["selfie"]; ok {
					if _, exists := mergedDocuments["selfie_marketing"]; !exists {
						mergedDocuments["selfie_marketing"] = selfie
					}
				}
				validationInput.DocumentData = mergedDocuments
			}
		}
		if err := validateAgentCreditSubmission(&validationInput, true); err != nil {
			return nil, fmt.Errorf(" %w", err)
		}
	}
	if role == helper.RoleRetailAgent {
		if strings.TrimSpace(in.AgentSignature) == "" && !hasStoredAgentSignature {
			return nil, errors.New("tanda tangan wajib diisi")
		}
		if in.ID > 0 {
			item, err := s.repo.CompleteApplicationConsent(ctx, repository.AgentCreditApplicationInput{
				ID:              in.ID,
				MemberID:        targetMemberID,
				RequestedAmount: in.RequestedAmount,
				ApplicantData:   in.ApplicantData,
				DocumentData:    in.DocumentData,
				AgentSignature:  strings.TrimSpace(in.AgentSignature),
			})
			if err == nil {
				return item, nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			// The client may hold an old application ID after a rejected or
			// deleted record. Resolve the current open application before
			// falling back to creating a new submitted application.
			existingID, lookupErr := s.repo.FindOpenEarlyApplicationID(ctx, targetMemberID)
			if lookupErr != nil {
				return nil, lookupErr
			}
			if existingID > 0 && existingID != in.ID {
				return s.repo.CompleteApplicationConsent(ctx, repository.AgentCreditApplicationInput{
					ID:              existingID,
					MemberID:        targetMemberID,
					RequestedAmount: in.RequestedAmount,
					ApplicantData:   in.ApplicantData,
					DocumentData:    in.DocumentData,
					AgentSignature:  strings.TrimSpace(in.AgentSignature),
				})
			}
			// No open row remains, so continue below and create a fresh
			// submitted application for the agent.
		}
		existingID, err := s.repo.FindOpenEarlyApplicationID(ctx, targetMemberID)
		if err != nil {
			return nil, err
		}
		if existingID > 0 {
			return s.repo.CompleteApplicationConsent(ctx, repository.AgentCreditApplicationInput{
				ID:              existingID,
				MemberID:        targetMemberID,
				RequestedAmount: in.RequestedAmount,
				ApplicantData:   in.ApplicantData,
				DocumentData:    in.DocumentData,
				AgentSignature:  strings.TrimSpace(in.AgentSignature),
			})
		}
		if err := s.repo.EnsureAgentCanSubmitNextApplication(ctx, targetMemberID); err != nil {
			return nil, err
		}
	} else if isCreditReviewer(role) && in.ID > 0 {
		if selfie, hasLegacySelfie := in.DocumentData["selfie"]; hasLegacySelfie {
			if _, ok := in.DocumentData["selfie_ktp"]; !ok {
				in.DocumentData["selfie_ktp"] = selfie
			}
		}
		if role == helper.RoleRetailMarketing || role == helper.RoleRetailMaster {
			if err := validateAgentCreditDocuments(in.DocumentData); err != nil {
				return nil, fmt.Errorf("validasi dokumen sistem: %w", err)
			}
			stampAgentCreditDocumentValidation(in.ApplicantData)
		}
		return s.repo.CompleteApplicationConsent(ctx, repository.AgentCreditApplicationInput{
			ID:              in.ID,
			MemberID:        targetMemberID,
			MarketingID:     creditApplicationMarketingID(role, auth.MemberID),
			RequestedAmount: in.RequestedAmount,
			ApplicantData:   in.ApplicantData,
			DocumentData:    in.DocumentData,
			AgentSignature:  "",
		})
	}
	return s.repo.CreateApplication(ctx, repository.AgentCreditApplicationInput{
		ID:              in.ID,
		MemberID:        targetMemberID,
		MarketingID:     creditApplicationMarketingID(role, auth.MemberID),
		RequestedAmount: in.RequestedAmount,
		ApplicantData:   in.ApplicantData,
		DocumentData:    in.DocumentData,
		AgentSignature:  strings.TrimSpace(in.AgentSignature),
	})
}

func (s *AgentCreditService) ListApplications(ctx context.Context, auth helper.AuthInfo, limit int) ([]repository.AgentCreditApplication, error) {
	if !canViewCreditApplications(auth.Role) {
		return nil, errors.New("akses pemantauan kredit tidak tersedia")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	items, err := s.repo.ListApplications(ctx, limit)
	if err != nil {
		return nil, err
	}
	if helper.NormalizeRole(auth.Role) == helper.RoleRetailMarketing {
		// Marketing only receives a non-financial operational view of its own
		// agents. Never leak balance, loan, limit, payment, or document data.
		visible := make([]repository.AgentCreditApplication, 0, len(items))
		for _, item := range items {
			if item.MarketingID != auth.MemberID {
				continue
			}
			item.RequestedAmount = 0
			item.ApprovedAmount = 0
			item.ApplicantData = map[string]any{
				"agent_name":    item.ApplicantData["agent_name"],
				"store_name":    item.ApplicantData["store_name"],
				"whatsapp":      item.ApplicantData["whatsapp"],
				"store_address": item.ApplicantData["store_address"],
			}
			item.DocumentData = nil
			item.AgentSignatureData = ""
			item.HasAgentSignature = false
			item.OutstandingAmount = 0
			item.CreditAvailableAmount = 0
			item.PaidAmount = 0
			item.PaymentCount = 0
			item.CreditLimitAmount = 0
			item.LoanApprovedAt = nil
			item.LoanDueDate = nil
			visible = append(visible, item)
		}
		return visible, nil
	}
	// Operator and admin receive the complete operational and financial view.
	return items, nil
}

func (s *AgentCreditService) ListMyApplications(ctx context.Context, auth helper.AuthInfo) ([]repository.AgentCreditApplication, error) {
	role := helper.NormalizeRole(auth.Role)
	if role != helper.RoleRetailAgent && role != helper.RoleUser {
		return nil, errors.New("user only")
	}
	return s.repo.ListMemberApplications(ctx, auth.MemberID, 10)
}

func (s *AgentCreditService) ListRanks(ctx context.Context, auth helper.AuthInfo) ([]repository.AgentCreditRank, error) {
	role := helper.NormalizeRole(auth.Role)
	if role != helper.RoleRetailAnalyst && role != helper.RoleAdmin {
		return nil, errors.New("operator kredit only")
	}
	return s.repo.ListActiveRanks(ctx)
}

func (s *AgentCreditService) ChangeMemberCreditRank(ctx context.Context, auth helper.AuthInfo, in AgentCreditRankChangeInput) (*repository.AgentCreditRank, error) {
	role := helper.NormalizeRole(auth.Role)
	if role != helper.RoleRetailAnalyst && role != helper.RoleAdmin {
		return nil, errors.New("operator kredit only")
	}
	if in.MemberID <= 0 || (in.RankID <= 0 && in.LimitAmount <= 0) {
		return nil, errors.New("agent dan nominal limit wajib diisi")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return nil, errors.New("catatan keputusan wajib diisi")
	}
	return s.repo.ChangeMemberCreditRank(ctx, in.MemberID, in.RankID, in.LimitAmount, auth.MemberID, reason)
}

func (s *AgentCreditService) SetLoanOperationalStatus(ctx context.Context, auth helper.AuthInfo, in AgentCreditLoanStatusInput) (*repository.AgentCreditLoanStatusResult, error) {
	if helper.NormalizeRole(auth.Role) != helper.RoleAdmin {
		return nil, errors.New("super admin only")
	}
	if in.ApplicationID <= 0 {
		return nil, errors.New("pengajuan kredit tidak valid")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return nil, errors.New("alasan perubahan status kredit wajib diisi")
	}
	return s.repo.SetLoanOperationalStatus(ctx, in.ApplicationID, auth.MemberID, in.Suspended, reason)
}

func (s *AgentCreditService) DecideApplication(ctx context.Context, auth helper.AuthInfo, in AgentCreditDecisionInput) (*repository.AgentCreditApplication, error) {
	if !isCreditReviewer(auth.Role) {
		return nil, errors.New("keputusan kredit hanya dapat dilakukan operator")
	}
	if in.ID <= 0 {
		return nil, errors.New("pengajuan tidak valid")
	}

	decision := strings.TrimSpace(strings.ToLower(in.Decision))
	role := helper.NormalizeRole(auth.Role)
	note := strings.TrimSpace(in.Note)
	approvedAmount := in.ApprovedAmount
	limitAmount := maximumAgentCreditAmount
	reviewState, err := s.repo.GetApplicationReviewState(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	switch role {
	case helper.RoleAdmin, helper.RoleStaff:
		return s.decideAsAdmin(ctx, auth.MemberID, in.ID, reviewState, decision, note, strings.TrimSpace(in.SignatureData), strings.TrimSpace(in.RiskLevel), in.RiskScore, approvedAmount, limitAmount)
	case helper.RoleRetailAnalyst:
		return s.decideAsAnalyst(ctx, auth.MemberID, in.ID, reviewState, decision, note, strings.TrimSpace(in.SignatureData), strings.TrimSpace(in.RiskLevel), in.RiskScore, approvedAmount, limitAmount, in.RevisionDocuments)
	default:
		return nil, errors.New("keputusan kredit hanya dapat dilakukan operator")
	}
}

func (s *AgentCreditService) DeleteRejectedApplication(ctx context.Context, auth helper.AuthInfo, applicationID int64) error {
	if !isCreditReviewer(auth.Role) {
		return errors.New("penghapusan pengajuan hanya dapat dilakukan operator")
	}
	if applicationID <= 0 {
		return errors.New("pengajuan tidak valid")
	}
	return s.repo.DeleteRejectedApplication(ctx, applicationID)
}

func creditApplicationMarketingID(role string, memberID int64) int64 {
	if helper.NormalizeRole(role) == helper.RoleRetailMarketing {
		return memberID
	}
	return 0
}

func (s *AgentCreditService) decideAsAdmin(ctx context.Context, adminID, applicationID int64, reviewState *repository.AgentCreditReviewState, decision, note, signatureData, riskLevel string, riskScore int64, approvedAmount, limitAmount int64) (*repository.AgentCreditApplication, error) {
	if reviewState.Status == "approved" || reviewState.Status == "rejected" {
		return nil, errors.New("pengajuan sudah memiliki keputusan akhir")
	}
	riskLevel = normalizeRiskLevel(riskLevel)
	switch decision {
	case "approve", "approved", "setujui":
		if approvedAmount <= 0 {
			approvedAmount = limitAmount
		}
		if approvedAmount > limitAmount {
			return nil, fmt.Errorf("nominal disetujui maksimal Rp%d", limitAmount)
		}
		if approvedAmount <= 0 {
			return nil, errors.New("nominal disetujui wajib diisi")
		}
		return s.repo.AnalystFinalDecision(ctx, repository.AgentCreditDecisionInput{
			ID:             applicationID,
			AnalystID:      adminID,
			Status:         "approved",
			ApprovedAmount: approvedAmount,
			AnalystNote:    fallbackNote(note, "Admin menyetujui pengajuan kredit secara manual."),
			Recommendation: "approved_by_admin",
			ActorRole:      "super_admin",
		}, signatureData, riskLevel, riskScore)
	case "reject", "rejected", "tolak":
		return s.repo.AnalystFinalDecision(ctx, repository.AgentCreditDecisionInput{
			ID:             applicationID,
			AnalystID:      adminID,
			Status:         "rejected",
			ApprovedAmount: 0,
			AnalystNote:    fallbackNote(note, "Ditolak oleh admin."),
			Recommendation: "rejected_by_admin",
			ActorRole:      "super_admin",
		}, signatureData, riskLevel, riskScore)
	default:
		return nil, errors.New("admin wajib memilih setuju atau tolak")
	}
}

func (s *AgentCreditService) decideAsMarketing(ctx context.Context, marketingID, applicationID int64, reviewState *repository.AgentCreditReviewState, decision, note, signatureData, actorRole string) (*repository.AgentCreditApplication, error) {
	if reviewState.Status != "submitted" && reviewState.Status != "marketing_review" {
		return nil, errors.New("status pengajuan belum bisa diverifikasi marketing")
	}
	switch decision {
	case "approve", "approved", "setujui", "forward_to_analysis", "kirim_analis":
		if !reviewState.TermsAccepted {
			return nil, errors.New("agent wajib menyetujui syarat dan ketentuan sebelum dikirim ke operator")
		}
		if !reviewState.HasAgentSignature {
			return nil, errors.New("agent wajib tanda tangan sebelum dikirim ke operator")
		}
		if !reviewState.HasKTPDocument || !reviewState.HasStoreDocument || !reviewState.HasSelfieKTPDocument || !reviewState.HasSelfieMarketing {
			missing := make([]string, 0, 4)
			if !reviewState.HasKTPDocument {
				missing = append(missing, "foto KTP")
			}
			if !reviewState.HasStoreDocument {
				missing = append(missing, "foto toko")
			}
			if !reviewState.HasSelfieKTPDocument {
				missing = append(missing, "dokumen formulir")
			}
			if !reviewState.HasSelfieMarketing {
				missing = append(missing, "foto bersama marketing")
			}
			return nil, fmt.Errorf("%s wajib lengkap sebelum dikirim ke operator", strings.Join(missing, ", "))
		}
		if !strings.HasPrefix(signatureData, "data:image/") {
			return nil, errors.New("tanda tangan marketing wajib diisi")
		}
		defaultReviewNote := "Marketing sudah verifikasi lapangan dan dokumen."
		if actorRole == "super_admin" {
			defaultReviewNote = "Super Admin sudah memverifikasi dokumen dan survei lapangan."
		}
		return s.repo.MarketingReviewApplication(ctx, repository.AgentCreditDecisionInput{
			ID:             applicationID,
			MarketingID:    marketingID,
			Status:         "analysis_review",
			MarketingNote:  fallbackNote(note, defaultReviewNote),
			Recommendation: "marketing_verified",
			ApprovedAmount: 0,
			ActorRole:      actorRole,
		}, signatureData)
	case "reject", "rejected", "tolak":
		return nil, errors.New("marketing hanya bisa verifikasi dan mengirim pengajuan ke operator")
	default:
		return nil, errors.New("marketing wajib memilih verifikasi dan kirim ke operator")
	}
}

func (s *AgentCreditService) decideAsAnalyst(ctx context.Context, analystID, applicationID int64, reviewState *repository.AgentCreditReviewState, decision, note, signatureData, riskLevel string, riskScore int64, approvedAmount, limitAmount int64, revisionDocuments []string) (*repository.AgentCreditApplication, error) {
	if reviewState.Status != "submitted" && reviewState.Status != "analysis_review" && reviewState.Status != "marketing_review" && reviewState.Status != "master_review" && reviewState.Status != "ready_to_disburse" {
		return nil, errors.New("pengajuan belum masuk tahap operator")
	}
	riskLevel = normalizeRiskLevel(riskLevel)
	switch decision {
	case "revision_required", "revision", "revisi", "perlu_revisi":
		documents := normalizeRevisionDocuments(revisionDocuments)
		if len(documents) == 0 {
			return nil, errors.New("pilih minimal satu dokumen yang perlu diperbaiki")
		}
		note = fallbackNote(note, "Operator meminta perbaikan pada dokumen yang dipilih.")
		return s.repo.ReturnApplicationToMarketing(ctx, repository.AgentCreditDecisionInput{
			ID:             applicationID,
			AnalystID:      analystID,
			AnalystNote:    note,
			Recommendation: "revision_required",
			ActorRole:      "operator_credit",
		}, documents)
	case "approve", "approved", "setujui":
		if approvedAmount <= 0 {
			approvedAmount = limitAmount
		}
		if approvedAmount > limitAmount {
			return nil, fmt.Errorf("nominal disetujui maksimal Rp%d", limitAmount)
		}
		if approvedAmount <= 0 {
			return nil, errors.New("nominal disetujui wajib diisi")
		}
		return s.repo.AnalystFinalDecision(ctx, repository.AgentCreditDecisionInput{
			ID:             applicationID,
			AnalystID:      analystID,
			Status:         "approved",
			ApprovedAmount: approvedAmount,
			AnalystNote:    fallbackNote(note, "Operator menyetujui risiko dan nominal. Pinjaman saldo aktif."),
			Recommendation: "approved",
			ActorRole:      "operator_credit",
		}, signatureData, riskLevel, riskScore)
	case "reject", "rejected", "tolak":
		return s.repo.AnalystFinalDecision(ctx, repository.AgentCreditDecisionInput{
			ID:             applicationID,
			AnalystID:      analystID,
			Status:         "rejected",
			ApprovedAmount: 0,
			AnalystNote:    fallbackNote(note, "Ditolak oleh operator."),
			Recommendation: "rejected",
			ActorRole:      "operator_credit",
		}, signatureData, riskLevel, riskScore)
	default:
		return nil, errors.New("operator wajib memilih setuju atau tolak")
	}
}

func normalizeRevisionDocuments(values []string) []string {
	allowed := map[string]bool{"ktp": true, "store": true, "selfie_ktp": true, "selfie_marketing": true}
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.TrimSpace(strings.ToLower(value))
		if allowed[key] && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}

func (s *AgentCreditService) decideAsMaster(ctx context.Context, masterID, applicationID int64, reviewState *repository.AgentCreditReviewState, decision, note, signatureData string) (*repository.AgentCreditApplication, error) {
	if reviewState.Status == "submitted" || reviewState.Status == "marketing_review" {
		switch decision {
		case "approve", "approved", "setujui", "forward_to_analysis", "kirim_analis":
			if !reviewState.TermsAccepted {
				return nil, errors.New("agent wajib menyetujui syarat dan ketentuan sebelum dikirim ke operator")
			}
			if !reviewState.HasAgentSignature {
				return nil, errors.New("agent wajib tanda tangan sebelum dikirim ke operator")
			}
			if !reviewState.HasKTPDocument || !reviewState.HasStoreDocument || !reviewState.HasSelfieKTPDocument || !reviewState.HasSelfieMarketing {
				return nil, errors.New("foto KTP, foto toko, dokumen formulir, dan selfie bersama marketing wajib lengkap")
			}
			if !strings.HasPrefix(signatureData, "data:image/") {
				return nil, errors.New("tanda tangan marketing wajib diisi")
			}
			return s.repo.MarketingReviewApplication(ctx, repository.AgentCreditDecisionInput{
				ID:             applicationID,
				MarketingID:    masterID,
				Status:         "analysis_review",
				MarketingNote:  fallbackNote(note, "Marketing sudah verifikasi data agent dan dokumen, lalu dikirim ke operator."),
				Recommendation: "marketing_verified",
				ApprovedAmount: 0,
			}, signatureData)
		case "reject", "rejected", "tolak":
			return nil, errors.New("marketing tidak bisa menolak pengajuan agent; keputusan tolak ada di operator")
		default:
			return nil, errors.New("marketing wajib tanda tangan dan kirim ke operator")
		}
	}
	return nil, errors.New("keputusan akhir pinjaman ada di operator")
}

func fallbackNote(note, fallback string) string {
	if strings.TrimSpace(note) != "" {
		return strings.TrimSpace(note)
	}
	return fallback
}

func normalizeRiskLevel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "aman", "low", "rendah":
		return "aman"
	case "perhatian", "medium", "sedang":
		return "perhatian"
	case "tinggi", "high":
		return "tinggi"
	default:
		return "perhatian"
	}
}

func (s *AgentCreditService) PayInstallment(ctx context.Context, auth helper.AuthInfo, in AgentCreditPaymentInput) error {
	if helper.NormalizeRole(auth.Role) != helper.RoleRetailAgent {
		return errors.New("agent only")
	}
	if in.ApplicationID <= 0 {
		return errors.New("pengajuan tidak valid")
	}
	if in.Amount <= 0 {
		return errors.New("nominal pelunasan wajib diisi")
	}
	if strings.TrimSpace(in.PaymentMethod) == "" {
		in.PaymentMethod = "transfer"
	}
	proofURL, _ := in.PaymentProof["data_url"].(string)
	if !strings.HasPrefix(strings.TrimSpace(proofURL), "data:image/") {
		return errors.New("bukti pembayaran wajib diupload")
	}
	refID := strings.TrimSpace(in.RefID)
	if refID == "" {
		refID = fmt.Sprintf("KREDIT-BAYAR-%d-%d", auth.MemberID, time.Now().UnixNano())
	}
	return s.repo.PayInstallment(ctx, repository.AgentCreditPaymentInput{
		ApplicationID: in.ApplicationID,
		MemberID:      auth.MemberID,
		RefID:         refID,
		Amount:        in.Amount,
		Note:          strings.TrimSpace(in.Note),
		PaymentMethod: strings.TrimSpace(in.PaymentMethod),
		PaymentProof:  in.PaymentProof,
	})
}

func (s *AgentCreditService) TransferToMainBalance(ctx context.Context, auth helper.AuthInfo, in AgentCreditTransferInput) (*repository.AgentCreditTransferResult, error) {
	return nil, errors.New("saldo kredit sudah tidak digunakan; kredit yang disetujui langsung masuk ke saldo utama")
}
