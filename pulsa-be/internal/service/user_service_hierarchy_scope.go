package service

import (
	"context"
	"errors"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *UserService) resolveRetailHierarchyAssignment(ctx context.Context, memberID int64, role string, agentRow, masterRow *repository.UserRow, preview repository.HierarchyAssignPreviewRow, originalRole string) (*resolvedHierarchyAssignment, error) {
	var err error
	if agentRow != nil && helper.NormalizeRole(agentRow.Role) != helper.RoleRetailAgent {
		return nil, errors.New("agent retail tidak valid")
	}
	if masterRow != nil && helper.NormalizeRole(masterRow.Role) != helper.RoleRetailMaster {
		return nil, errors.New("master retail tidak valid")
	}
	switch role {
	case helper.RoleUser:
	case helper.RoleRetailAgent:
		if agentRow != nil {
			return nil, errors.New("akun agent retail tidak boleh punya agent di atasnya")
		}
	case helper.RoleRetailMaster:
		if agentRow != nil || masterRow != nil {
			return nil, errors.New("akun master retail tidak boleh punya upline")
		}
	default:
		return nil, errors.New("role retail tidak valid")
	}
	if agentRow != nil && masterRow == nil && agentRow.RetailMasterID != nil {
		masterRow, err = s.repo.Get(ctx, *agentRow.RetailMasterID)
		if err != nil {
			return nil, err
		}
		preview.DerivedMasterFromAgent = true
	}
	if agentRow != nil && agentRow.RetailMasterID != nil && masterRow != nil && *agentRow.RetailMasterID != masterRow.ID {
		return nil, errors.New("master retail tidak cocok dengan master milik agent")
	}
	count, countErr := s.retailRepo.CountSuccessfulOrdersByMember(ctx, memberID)
	if countErr != nil {
		return nil, countErr
	}
	preview.TransactionCount = count
	if agentRow != nil {
		v := agentRow.ID
		preview.NextAgentID = &v
		preview.Agent = &repository.HierarchyAssignPreviewTarget{
			MemberID:        agentRow.ID,
			Email:           agentRow.Email,
			Nama:            agentRow.Nama,
			Role:            helper.NormalizeRole(agentRow.Role),
			CommissionRp:    agentRow.RetailAgentCommissionRp,
			CalculatedCount: count,
			CalculatedTotal: count * agentRow.RetailAgentCommissionRp,
		}
	}
	if masterRow != nil {
		v := masterRow.ID
		preview.NextMasterID = &v
		preview.Master = &repository.HierarchyAssignPreviewTarget{
			MemberID:        masterRow.ID,
			Email:           masterRow.Email,
			Nama:            masterRow.Nama,
			Role:            helper.NormalizeRole(masterRow.Role),
			CommissionRp:    masterRow.RetailMasterCommissionRp,
			CalculatedCount: count,
			CalculatedTotal: count * masterRow.RetailMasterCommissionRp,
		}
	}
	return &resolvedHierarchyAssignment{preview: preview, originalRole: originalRole}, nil
}

func (s *UserService) resolveH2HHierarchyAssignment(ctx context.Context, memberID int64, role string, agentRow, masterRow *repository.UserRow, preview repository.HierarchyAssignPreviewRow, originalRole string) (*resolvedHierarchyAssignment, error) {
	var err error
	if agentRow != nil && helper.NormalizeRole(agentRow.Role) != helper.RoleH2HAgent {
		return nil, errors.New("agent h2h tidak valid")
	}
	if masterRow != nil && helper.NormalizeRole(masterRow.Role) != helper.RoleH2HMaster {
		return nil, errors.New("master h2h tidak valid")
	}
	switch role {
	case helper.RoleMember:
	case helper.RoleH2HAgent:
		if agentRow != nil {
			return nil, errors.New("akun agent h2h tidak boleh punya agent di atasnya")
		}
	case helper.RoleH2HMaster:
		if agentRow != nil || masterRow != nil {
			return nil, errors.New("akun master h2h tidak boleh punya upline")
		}
	default:
		return nil, errors.New("role h2h tidak valid")
	}
	if agentRow != nil && masterRow == nil && agentRow.H2HMasterID != nil {
		masterRow, err = s.repo.Get(ctx, *agentRow.H2HMasterID)
		if err != nil {
			return nil, err
		}
		preview.DerivedMasterFromAgent = true
	}
	if agentRow != nil && agentRow.H2HMasterID != nil && masterRow != nil && *agentRow.H2HMasterID != masterRow.ID {
		return nil, errors.New("master h2h tidak cocok dengan master milik agent")
	}
	count, countErr := s.h2hRepo.CountSuccessfulTransactionsByMember(ctx, memberID)
	if countErr != nil {
		return nil, countErr
	}
	preview.TransactionCount = count
	if agentRow != nil {
		v := agentRow.ID
		preview.NextAgentID = &v
		preview.Agent = &repository.HierarchyAssignPreviewTarget{
			MemberID:        agentRow.ID,
			Email:           agentRow.Email,
			Nama:            agentRow.Nama,
			Role:            helper.NormalizeRole(agentRow.Role),
			CommissionRp:    agentRow.H2HAgentCommissionRp,
			CalculatedCount: count,
			CalculatedTotal: count * agentRow.H2HAgentCommissionRp,
		}
	}
	if masterRow != nil {
		v := masterRow.ID
		preview.NextMasterID = &v
		preview.Master = &repository.HierarchyAssignPreviewTarget{
			MemberID:        masterRow.ID,
			Email:           masterRow.Email,
			Nama:            masterRow.Nama,
			Role:            helper.NormalizeRole(masterRow.Role),
			CommissionRp:    masterRow.H2HMasterCommissionRp,
			CalculatedCount: count,
			CalculatedTotal: count * masterRow.H2HMasterCommissionRp,
		}
	}
	return &resolvedHierarchyAssignment{preview: preview, originalRole: originalRole}, nil
}
