package usecase

import (
	"context"
	"encoding/json"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type TaxUseCase struct {
	Taxes port.TaxRepository
	Audit port.AuditRepository
}

type CreateTaxRuleInput struct {
	TenantID    uuid.UUID
	Name        string
	RatePercent float64
}

func (uc *TaxUseCase) CreateRule(ctx context.Context, in CreateTaxRuleInput) (*domain.TaxRule, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionFeesManage); err != nil {
		return nil, ErrForbidden
	}
	if in.TenantID == uuid.Nil || in.Name == "" || in.RatePercent < 0 {
		return nil, ErrValidation
	}
	rule := &domain.TaxRule{
		ID:          uuid.New(),
		TenantID:    in.TenantID,
		Name:        in.Name,
		RatePercent: in.RatePercent,
		Active:      true,
	}
	if err := uc.Taxes.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (uc *TaxUseCase) UpdateRule(ctx context.Context, tenantID, ruleID uuid.UUID, in CreateTaxRuleInput) (*domain.TaxRule, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionFeesManage); err != nil {
		return nil, ErrForbidden
	}
	current, err := uc.Taxes.GetRuleByID(ctx, tenantID, ruleID)
	if err != nil {
		return nil, err
	}
	if in.Name == "" || in.RatePercent < 0 {
		return nil, ErrValidation
	}
	current.Name = in.Name
	current.RatePercent = in.RatePercent
	if err := uc.Taxes.UpdateRule(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (uc *TaxUseCase) ListRules(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]domain.TaxRule, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Taxes.ListRules(ctx, tenantID, activeOnly)
}

func (uc *TaxUseCase) DeactivateRule(ctx context.Context, tenantID, ruleID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionFeesManage); err != nil {
		return ErrForbidden
	}
	return uc.Taxes.DeactivateRule(ctx, tenantID, ruleID)
}

type CreateTaxAssignmentInput struct {
	TenantID  uuid.UUID
	ScopeType domain.TaxRuleScopeType
	ScopeID   uuid.UUID
	TaxRuleID uuid.UUID
}

func (uc *TaxUseCase) CreateAssignment(ctx context.Context, in CreateTaxAssignmentInput) (*domain.TaxRuleAssignment, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionFeesManage); err != nil {
		return nil, ErrForbidden
	}
	if in.TenantID == uuid.Nil || in.ScopeID == uuid.Nil || in.TaxRuleID == uuid.Nil {
		return nil, ErrValidation
	}
	if in.ScopeType != domain.TaxRuleScopeTenant && in.ScopeType != domain.TaxRuleScopeAuction && in.ScopeType != domain.TaxRuleScopeLot {
		return nil, ErrValidation
	}
	assignment := &domain.TaxRuleAssignment{
		ID:        uuid.New(),
		ScopeType: in.ScopeType,
		ScopeID:   in.ScopeID,
		TaxRuleID: in.TaxRuleID,
	}
	if err := uc.Taxes.CreateAssignment(ctx, in.TenantID, assignment); err != nil {
		return nil, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{
			"scope_type":  assignment.ScopeType,
			"scope_id":    assignment.ScopeID,
			"tax_rule_id": assignment.TaxRuleID,
			"change":      "CREATED",
		})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &in.TenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "TAX_RULE_ASSIGNMENTS_CHANGED",
			EntityType:  "TAX_RULE_ASSIGNMENT",
			EntityID:    &assignment.ID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return assignment, nil
}

func (uc *TaxUseCase) ListAssignments(ctx context.Context, tenantID uuid.UUID, scopeType domain.TaxRuleScopeType, scopeID uuid.UUID) ([]domain.TaxRuleAssignment, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if scopeID == uuid.Nil {
		return nil, ErrValidation
	}
	if scopeType != domain.TaxRuleScopeTenant && scopeType != domain.TaxRuleScopeAuction && scopeType != domain.TaxRuleScopeLot {
		return nil, ErrValidation
	}
	return uc.Taxes.ListAssignments(ctx, tenantID, scopeType, scopeID)
}

func (uc *TaxUseCase) DeleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionFeesManage); err != nil {
		return ErrForbidden
	}
	if err := uc.Taxes.DeleteAssignment(ctx, tenantID, assignmentID); err != nil {
		return err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"change": "DELETED"})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "TAX_RULE_ASSIGNMENTS_CHANGED",
			EntityType:  "TAX_RULE_ASSIGNMENT",
			EntityID:    &assignmentID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return nil
}
