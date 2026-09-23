package service

import (
	"context"
	"database/sql"
	"errors"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *UserService) PreviewHierarchyAssignment(ctx context.Context, scope string, memberID int64, agentID, masterID *int64, targetRole string) (*repository.HierarchyAssignPreviewRow, error) {
	resolved, err := s.resolveHierarchyAssignment(ctx, scope, memberID, agentID, masterID, targetRole)
	if err != nil {
		return nil, err
	}
	return &resolved.preview, nil
}

func (s *UserService) ApplyHierarchyAssignment(ctx context.Context, scope string, memberID int64, agentID, masterID *int64, targetRole string, applyHistorical bool) (*repository.HierarchyAssignApplyRow, error) {
	resolved, err := s.resolveHierarchyAssignment(ctx, scope, memberID, agentID, masterID, targetRole)
	if err != nil {
		return nil, err
	}

	out := &repository.HierarchyAssignApplyRow{Preview: resolved.preview}

	switch resolved.preview.Scope {
	case "retail":
		out.AgentAppliedCount, out.AgentAppliedTotal, out.MasterAppliedCount, out.MasterAppliedTotal, err =
			s.retailRepo.ApplyHierarchyAssignment(ctx, memberID, resolved.preview.MemberRole, resolved.preview.NextAgentID, resolved.preview.NextMasterID, applyHistorical)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "h2h":
		if resolved.originalRole != resolved.preview.MemberRole {
			if err := s.repo.SetRole(ctx, memberID, resolved.preview.MemberRole); err != nil {
				return nil, err
			}
		}
		if err := s.repo.SetH2HHierarchy(ctx, memberID, resolved.preview.NextAgentID, resolved.preview.NextMasterID); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("scope tidak valid")
	}

	if !applyHistorical {
		return out, nil
	}

	switch resolved.preview.Scope {
	case "h2h":
		out.AgentAppliedCount, out.AgentAppliedTotal, out.MasterAppliedCount, out.MasterAppliedTotal, err =
			s.h2hRepo.ApplyHistoricalCommission(ctx, memberID, resolved.preview.MemberRole, resolved.preview.NextAgentID, resolved.preview.NextMasterID)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

type resolvedHierarchyAssignment struct {
	preview      repository.HierarchyAssignPreviewRow
	originalRole string
}

func (s *UserService) resolveHierarchyAssignment(ctx context.Context, scope string, memberID int64, agentID, masterID *int64, targetRole string) (*resolvedHierarchyAssignment, error) {
	scope = normalizeScope(scope)
	if scope != "retail" && scope != "h2h" {
		return nil, errors.New("scope tidak valid")
	}
	if memberID <= 0 {
		return nil, errors.New("member_id tidak valid")
	}

	member, err := s.repo.Get(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, sql.ErrNoRows
	}

	originalRole := helper.NormalizeRole(member.Role)
	role := helper.NormalizeRole(targetRole)
	if role == "" {
		role = originalRole
	}
	if scope == "retail" && !helper.IsRetailRole(role) {
		return nil, errors.New("akun bukan member retail")
	}
	if scope == "h2h" && !helper.IsH2HRole(role) {
		return nil, errors.New("akun bukan member h2h")
	}

	var (
		agentRow  *repository.UserRow
		masterRow *repository.UserRow
	)
	if agentID != nil && *agentID > 0 {
		if *agentID == memberID {
			return nil, errors.New("akun tidak boleh menjadi agent dirinya sendiri")
		}
		agentRow, err = s.repo.Get(ctx, *agentID)
		if err != nil {
			return nil, err
		}
		if agentRow == nil {
			return nil, errors.New("agent tidak ditemukan")
		}
	}
	if masterID != nil && *masterID > 0 {
		if *masterID == memberID {
			return nil, errors.New("akun tidak boleh menjadi master dirinya sendiri")
		}
		masterRow, err = s.repo.Get(ctx, *masterID)
		if err != nil {
			return nil, err
		}
		if masterRow == nil {
			return nil, errors.New("master tidak ditemukan")
		}
	}
	if agentRow != nil && masterRow != nil && agentRow.ID == masterRow.ID {
		return nil, errors.New("agent dan master tidak boleh akun yang sama")
	}

	preview := repository.HierarchyAssignPreviewRow{
		Scope:       scope,
		MemberID:    member.ID,
		MemberEmail: member.Email,
		MemberNama:  member.Nama,
		MemberRole:  role,
	}

	if scope == "retail" {
		preview.CurrentAgentID = member.RetailAgentID
		preview.CurrentMasterID = member.RetailMasterID
	} else {
		preview.CurrentAgentID = member.H2HAgentID
		preview.CurrentMasterID = member.H2HMasterID
	}

	switch scope {
	case "retail":
		return s.resolveRetailHierarchyAssignment(ctx, memberID, role, agentRow, masterRow, preview, originalRole)
	case "h2h":
		return s.resolveH2HHierarchyAssignment(ctx, memberID, role, agentRow, masterRow, preview, originalRole)
	default:
		return nil, errors.New("scope tidak valid")
	}
}
