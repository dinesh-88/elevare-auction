package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuthUseCase struct {
	Users        port.UserRepository
	TenantLookup port.UserTenantLookupRepository
	Hasher       port.PasswordHasher
	Tokens       port.TokenService
}

type RegisterInput struct {
	Tenant   *domain.Tenant
	Email    string
	Password string
	Role     domain.Role
}

type LoginInput struct {
	TenantID uuid.UUID
	Email    string
	Password string
}

type TenantLoginMatch struct {
	TenantID     uuid.UUID
	TenantName   string
	Subdomain    string
	CustomDomain *string
}

func (uc *AuthUseCase) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	if in.Tenant == nil || in.Email == "" || len(in.Password) < 8 {
		return nil, ErrValidation
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	approval := domain.ApprovalApproved
	if in.Tenant.RequireAdminApproval {
		approval = domain.ApprovalPending
	}
	user := &domain.User{
		ID:             uuid.New(),
		TenantID:       in.Tenant.ID,
		Email:          in.Email,
		PasswordHash:   hash,
		Role:           in.Role,
		Status:         domain.UserStatusActive,
		EmailVerified:  false,
		ApprovalStatus: approval,
		KYCStatus:      domain.KYCNone,
	}
	if err := uc.Users.Create(ctx, in.Tenant.ID, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *AuthUseCase) VerifyEmail(ctx context.Context, tenantID, userID uuid.UUID) error {
	return uc.Users.UpdateEmailVerified(ctx, tenantID, userID, true)
}

func (uc *AuthUseCase) Login(ctx context.Context, in LoginInput) (string, *domain.User, error) {
	user, err := uc.Users.GetByEmail(ctx, in.TenantID, strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil {
		return "", nil, err
	}
	if !uc.Hasher.Verify(in.Password, user.PasswordHash) {
		return "", nil, ErrInvalidCredential
	}
	if !user.EmailVerified {
		return "", nil, errors.New("email not verified")
	}
	if user.ApprovalStatus != domain.ApprovalApproved {
		return "", nil, errors.New("user not approved")
	}
	if user.Status == domain.UserStatusSuspended {
		return "", nil, errors.New("user is suspended")
	}
	token, err := uc.Tokens.Issue(user.ID, user.TenantID, user.Role)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (uc *AuthUseCase) CurrentUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrValidation
	}
	return uc.Users.GetByID(ctx, tenantID, userID)
}

func (uc *AuthUseCase) DiscoverTenantsByEmail(ctx context.Context, email string) ([]TenantLoginMatch, error) {
	if uc.TenantLookup == nil {
		return nil, errors.New("tenant lookup unavailable")
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || !strings.Contains(normalizedEmail, "@") {
		return nil, ErrValidation
	}
	rows, err := uc.TenantLookup.ListTenantLoginCandidatesByEmail(ctx, normalizedEmail)
	if err != nil {
		return nil, err
	}
	out := make([]TenantLoginMatch, 0, len(rows))
	for _, row := range rows {
		out = append(out, TenantLoginMatch{
			TenantID:     row.TenantID,
			TenantName:   row.TenantName,
			Subdomain:    row.Subdomain,
			CustomDomain: row.CustomDomain,
		})
	}
	return out, nil
}
