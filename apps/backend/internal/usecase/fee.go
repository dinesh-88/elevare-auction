package usecase

import (
	"context"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type FeeUseCase struct {
	Fees port.FeeRepository
}

type CreateFeeComponentInput struct {
	TenantID        uuid.UUID
	Name            string
	CalculationType domain.FeeCalculationType
	Value           float64
	AppliesTo       domain.FeeAppliesTo
	Taxable         bool
}

func (uc *FeeUseCase) CreateComponent(ctx context.Context, in CreateFeeComponentInput) (*domain.FeeComponent, error) {
	if in.TenantID == uuid.Nil || in.Name == "" || in.Value < 0 {
		return nil, ErrValidation
	}
	if in.CalculationType != domain.FeeCalculationPercentage && in.CalculationType != domain.FeeCalculationFixed {
		return nil, ErrValidation
	}
	if in.AppliesTo != domain.FeeAppliesToHammer && in.AppliesTo != domain.FeeAppliesToPremium && in.AppliesTo != domain.FeeAppliesToSubtotal {
		return nil, ErrValidation
	}
	component := &domain.FeeComponent{
		ID:              uuid.New(),
		TenantID:        in.TenantID,
		Name:            in.Name,
		CalculationType: in.CalculationType,
		Value:           in.Value,
		AppliesTo:       in.AppliesTo,
		Taxable:         in.Taxable,
		Active:          true,
	}
	if err := uc.Fees.CreateComponent(ctx, component); err != nil {
		return nil, err
	}
	return component, nil
}

func (uc *FeeUseCase) UpdateComponent(ctx context.Context, tenantID, componentID uuid.UUID, in CreateFeeComponentInput) (*domain.FeeComponent, error) {
	current, err := uc.Fees.GetComponentByID(ctx, tenantID, componentID)
	if err != nil {
		return nil, err
	}
	if in.Name == "" || in.Value < 0 {
		return nil, ErrValidation
	}
	if in.CalculationType != domain.FeeCalculationPercentage && in.CalculationType != domain.FeeCalculationFixed {
		return nil, ErrValidation
	}
	if in.AppliesTo != domain.FeeAppliesToHammer && in.AppliesTo != domain.FeeAppliesToPremium && in.AppliesTo != domain.FeeAppliesToSubtotal {
		return nil, ErrValidation
	}
	current.Name = in.Name
	current.CalculationType = in.CalculationType
	current.Value = in.Value
	current.AppliesTo = in.AppliesTo
	current.Taxable = in.Taxable
	if err := uc.Fees.UpdateComponent(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (uc *FeeUseCase) ListComponents(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]domain.FeeComponent, error) {
	return uc.Fees.ListComponents(ctx, tenantID, activeOnly)
}

func (uc *FeeUseCase) DeactivateComponent(ctx context.Context, tenantID, componentID uuid.UUID) error {
	return uc.Fees.DeactivateComponent(ctx, tenantID, componentID)
}

type CreateFeeAssignmentInput struct {
	TenantID       uuid.UUID
	ScopeType      domain.FeeScopeType
	ScopeID        uuid.UUID
	FeeComponentID uuid.UUID
}

func (uc *FeeUseCase) CreateAssignment(ctx context.Context, in CreateFeeAssignmentInput) (*domain.FeeAssignment, error) {
	if in.TenantID == uuid.Nil || in.ScopeID == uuid.Nil || in.FeeComponentID == uuid.Nil {
		return nil, ErrValidation
	}
	if in.ScopeType != domain.FeeScopeTenant && in.ScopeType != domain.FeeScopeAuction && in.ScopeType != domain.FeeScopeLot {
		return nil, ErrValidation
	}
	assignment := &domain.FeeAssignment{
		ID:             uuid.New(),
		ScopeType:      in.ScopeType,
		ScopeID:        in.ScopeID,
		FeeComponentID: in.FeeComponentID,
	}
	if err := uc.Fees.CreateAssignment(ctx, in.TenantID, assignment); err != nil {
		return nil, err
	}
	return assignment, nil
}

func (uc *FeeUseCase) ListAssignments(ctx context.Context, tenantID uuid.UUID, scopeType domain.FeeScopeType, scopeID uuid.UUID) ([]domain.FeeAssignment, error) {
	if scopeID == uuid.Nil {
		return nil, ErrValidation
	}
	if scopeType != domain.FeeScopeTenant && scopeType != domain.FeeScopeAuction && scopeType != domain.FeeScopeLot {
		return nil, ErrValidation
	}
	return uc.Fees.ListAssignments(ctx, tenantID, scopeType, scopeID)
}

func (uc *FeeUseCase) DeleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error {
	return uc.Fees.DeleteAssignment(ctx, tenantID, assignmentID)
}
