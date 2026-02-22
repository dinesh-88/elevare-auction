package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type OnboardingUseCase struct {
	Tenants port.TenantRepository
	Hasher  port.PasswordHasher
}

type TenantSignupInput struct {
	OrganizationName       string
	DesiredSubdomain       string
	AdminEmail             string
	Password               string
	Country                string
	EstimatedAuctionVolume *int
}

var subdomainPattern = regexp.MustCompile(`^[a-z0-9-]{3,40}$`)

func normalizeSubdomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	s = strings.Split(s, ".")[0]
	return s
}

func (uc *OnboardingUseCase) PublicSignup(ctx context.Context, in TenantSignupInput) (*domain.Tenant, *domain.User, error) {
	name := strings.TrimSpace(in.OrganizationName)
	subdomain := normalizeSubdomain(in.DesiredSubdomain)
	email := strings.ToLower(strings.TrimSpace(in.AdminEmail))
	country := strings.TrimSpace(in.Country)

	if name == "" || email == "" || country == "" || len(in.Password) < 8 || !subdomainPattern.MatchString(subdomain) {
		return nil, nil, ErrValidation
	}

	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return nil, nil, err
	}

	tenant := &domain.Tenant{
		ID:                     uuid.New(),
		Name:                   name,
		Subdomain:              subdomain,
		Status:                 domain.TenantPending,
		Country:                country,
		EstimatedAuctionVolume: in.EstimatedAuctionVolume,
		RequireAdminApproval:   true,
		RequireKYC:             false,
	}
	admin := &domain.User{
		ID:             uuid.New(),
		TenantID:       tenant.ID,
		Email:          email,
		PasswordHash:   hash,
		Role:           domain.RoleAdmin,
		Status:         domain.UserStatusActive,
		EmailVerified:  false,
		ApprovalStatus: domain.ApprovalPending,
		KYCStatus:      domain.KYCNone,
	}

	err = uc.Tenants.CreatePendingTenantWithAdmin(ctx, port.CreatePendingTenantWithAdminInput{
		TenantID:               tenant.ID,
		TenantName:             tenant.Name,
		Subdomain:              tenant.Subdomain,
		Country:                tenant.Country,
		EstimatedAuctionVolume: tenant.EstimatedAuctionVolume,
		RequireAdminApproval:   tenant.RequireAdminApproval,
		RequireKYC:             tenant.RequireKYC,
		AdminUserID:            admin.ID,
		AdminEmail:             admin.Email,
		AdminPasswordHash:      admin.PasswordHash,
		AdminEmailVerified:     admin.EmailVerified,
		AdminApprovalStatus:    admin.ApprovalStatus,
	})
	if err != nil {
		return nil, nil, err
	}

	return tenant, admin, nil
}

func (uc *OnboardingUseCase) ListTenantRequests(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error) {
	if status == "" {
		status = domain.TenantPending
	}
	return uc.Tenants.ListByStatus(ctx, status)
}

func (uc *OnboardingUseCase) ApproveTenant(ctx context.Context, tenantID uuid.UUID) error {
	return uc.Tenants.UpdateStatusAndAdminApproval(ctx, tenantID, domain.TenantActive, domain.ApprovalApproved)
}

func (uc *OnboardingUseCase) RejectTenant(ctx context.Context, tenantID uuid.UUID) error {
	return uc.Tenants.UpdateStatusAndAdminApproval(ctx, tenantID, domain.TenantRejected, domain.ApprovalRejected)
}
