package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type fakeStaffRepo struct {
	memberships map[uuid.UUID]domain.TenantUser
	invitations map[uuid.UUID]domain.TenantInvitation
	byTokenHash map[string]uuid.UUID
}

func (f *fakeStaffRepo) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission domain.Permission) (bool, error) {
	return false, errors.New("unused")
}

func (f *fakeStaffRepo) GetEffectiveRole(ctx context.Context, tenantID, userID uuid.UUID) (domain.Role, error) {
	return "", errors.New("unused")
}

func (f *fakeStaffRepo) CreateTenantUser(ctx context.Context, item *domain.TenantUser) error {
	if f.memberships == nil {
		f.memberships = map[uuid.UUID]domain.TenantUser{}
	}
	f.memberships[item.ID] = *item
	return nil
}

func (f *fakeStaffRepo) ListTenantUsers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantUser, error) {
	return nil, nil
}

func (f *fakeStaffRepo) GetTenantUserByID(ctx context.Context, tenantID, tenantUserID uuid.UUID) (*domain.TenantUser, error) {
	item, ok := f.memberships[tenantUserID]
	if !ok {
		return nil, errors.New("staff member not found")
	}
	return &item, nil
}

func (f *fakeStaffRepo) GetTenantUserByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantUser, error) {
	for _, item := range f.memberships {
		if item.UserID == userID && item.TenantID == tenantID {
			copy := item
			return &copy, nil
		}
	}
	return nil, errors.New("staff member not found")
}

func (f *fakeStaffRepo) UpdateTenantUserRole(ctx context.Context, tenantID, tenantUserID uuid.UUID, role domain.Role) error {
	item, ok := f.memberships[tenantUserID]
	if !ok {
		return errors.New("staff member not found")
	}
	item.Role = role
	f.memberships[tenantUserID] = item
	return nil
}

func (f *fakeStaffRepo) DeleteTenantUser(ctx context.Context, tenantID, tenantUserID uuid.UUID) error {
	if _, ok := f.memberships[tenantUserID]; !ok {
		return errors.New("staff member not found")
	}
	delete(f.memberships, tenantUserID)
	return nil
}

func (f *fakeStaffRepo) CreateInvitation(ctx context.Context, item *domain.TenantInvitation) error {
	if f.invitations == nil {
		f.invitations = map[uuid.UUID]domain.TenantInvitation{}
	}
	if f.byTokenHash == nil {
		f.byTokenHash = map[string]uuid.UUID{}
	}
	f.invitations[item.ID] = *item
	f.byTokenHash[item.TokenHash] = item.ID
	return nil
}

func (f *fakeStaffRepo) ListInvitations(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantInvitation, error) {
	return nil, nil
}

func (f *fakeStaffRepo) GetInvitationByID(ctx context.Context, tenantID, invitationID uuid.UUID) (*domain.TenantInvitation, error) {
	item, ok := f.invitations[invitationID]
	if !ok {
		return nil, errors.New("invitation not found")
	}
	return &item, nil
}

func (f *fakeStaffRepo) GetInvitationByTokenHash(ctx context.Context, tenantID uuid.UUID, tokenHash string) (*domain.TenantInvitation, error) {
	id, ok := f.byTokenHash[tokenHash]
	if !ok {
		return nil, errors.New("invitation not found")
	}
	item := f.invitations[id]
	return &item, nil
}

func (f *fakeStaffRepo) MarkInvitationAccepted(ctx context.Context, tenantID, invitationID uuid.UUID, acceptedAt time.Time) error {
	item, ok := f.invitations[invitationID]
	if !ok {
		return errors.New("invitation not found")
	}
	item.AcceptedAt = &acceptedAt
	f.invitations[invitationID] = item
	return nil
}

func (f *fakeStaffRepo) DeleteInvitation(ctx context.Context, tenantID, invitationID uuid.UUID) error {
	if _, ok := f.invitations[invitationID]; !ok {
		return errors.New("invitation not found")
	}
	delete(f.invitations, invitationID)
	return nil
}

type fakeStaffUserRepo struct {
	users       map[uuid.UUID]domain.User
	usersByMail map[string]uuid.UUID
}

func (f *fakeStaffUserRepo) Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error {
	if f.users == nil {
		f.users = map[uuid.UUID]domain.User{}
	}
	if f.usersByMail == nil {
		f.usersByMail = map[string]uuid.UUID{}
	}
	f.users[user.ID] = *user
	f.usersByMail[user.Email] = user.ID
	return nil
}

func (f *fakeStaffUserRepo) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	user, ok := f.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (f *fakeStaffUserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	id, ok := f.usersByMail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	user := f.users[id]
	return &user, nil
}

func (f *fakeStaffUserRepo) GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return nil, nil
}
func (f *fakeStaffUserRepo) UpdateLocale(ctx context.Context, tenantID, userID uuid.UUID, locale domain.Locale) error {
	return nil
}
func (f *fakeStaffUserRepo) UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error {
	return nil
}
func (f *fakeStaffUserRepo) UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error {
	return nil
}
func (f *fakeStaffUserRepo) UpdateRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.Role) error {
	user, ok := f.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	user.Role = role
	f.users[userID] = user
	return nil
}

type fakeStaffHasher struct{}

func (f *fakeStaffHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (f *fakeStaffHasher) Verify(password, hash string) bool    { return true }

type fakeStaffTokens struct{}

func (f *fakeStaffTokens) Issue(userID, tenantID uuid.UUID, role domain.Role, locale domain.Locale) (string, error) {
	return "token", nil
}
func (f *fakeStaffTokens) Parse(token string) (*port.TokenClaims, error) {
	return nil, errors.New("unused")
}

func TestStaffInvite_AdminCannotInviteOwner(t *testing.T) {
	tenantID := uuid.New()
	uc := &StaffUseCase{}

	_, _, err := uc.Invite(authCtx(tenantID, domain.RoleAdmin), InviteStaffInput{
		TenantID:  tenantID,
		Email:     "person@example.com",
		Role:      domain.RoleOwner,
		InvitedBy: uuid.New(),
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestStaffInvite_NonStaffManageCannotInvite(t *testing.T) {
	tenantID := uuid.New()
	uc := &StaffUseCase{}

	_, _, err := uc.Invite(authCtx(tenantID, domain.RoleViewer), InviteStaffInput{
		TenantID:  tenantID,
		Email:     "person@example.com",
		Role:      domain.RoleStaff,
		InvitedBy: uuid.New(),
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestStaffInvite_TokenCannotBeReused(t *testing.T) {
	tenantID := uuid.New()
	invitationID := uuid.New()
	repo := &fakeStaffRepo{
		invitations: map[uuid.UUID]domain.TenantInvitation{
			invitationID: {
				ID:         invitationID,
				TenantID:   tenantID,
				Email:      "person@example.com",
				Role:       domain.RoleStaff,
				TokenHash:  invitationTokenHash("abc"),
				AcceptedAt: timePtr(time.Now().UTC()),
				ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
			},
		},
		byTokenHash: map[string]uuid.UUID{invitationTokenHash("abc"): invitationID},
	}
	uc := &StaffUseCase{Staff: repo, Users: &fakeStaffUserRepo{}, Hasher: &fakeStaffHasher{}, Tokens: &fakeStaffTokens{}}

	_, _, err := uc.AcceptInvitation(context.Background(), AcceptInvitationInput{TenantID: tenantID, Token: "abc", Password: "password123", Locale: "en"})
	if err == nil || err.Error() != "invitation already accepted" {
		t.Fatalf("expected token reuse rejection, got %v", err)
	}
}

func TestStaffInvite_ExpiredTokenRejected(t *testing.T) {
	tenantID := uuid.New()
	invitationID := uuid.New()
	repo := &fakeStaffRepo{
		invitations: map[uuid.UUID]domain.TenantInvitation{
			invitationID: {
				ID:        invitationID,
				TenantID:  tenantID,
				Email:     "person@example.com",
				Role:      domain.RoleStaff,
				TokenHash: invitationTokenHash("abc"),
				ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
			},
		},
		byTokenHash: map[string]uuid.UUID{invitationTokenHash("abc"): invitationID},
	}
	uc := &StaffUseCase{Staff: repo, Users: &fakeStaffUserRepo{}, Hasher: &fakeStaffHasher{}, Tokens: &fakeStaffTokens{}}

	_, _, err := uc.AcceptInvitation(context.Background(), AcceptInvitationInput{TenantID: tenantID, Token: "abc", Password: "password123", Locale: "en"})
	if err == nil || err.Error() != "invitation expired" {
		t.Fatalf("expected expired token rejection, got %v", err)
	}
}

func TestStaffRemove_OwnerCannotBeRemoved(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	tenantUserID := uuid.New()
	repo := &fakeStaffRepo{
		memberships: map[uuid.UUID]domain.TenantUser{
			tenantUserID: {
				ID:             tenantUserID,
				TenantID:       tenantID,
				UserID:         userID,
				Role:           domain.RoleOwner,
				IsPrimaryOwner: true,
			},
		},
	}
	users := &fakeStaffUserRepo{users: map[uuid.UUID]domain.User{userID: {ID: userID, TenantID: tenantID, Email: "owner@example.com", Role: domain.RoleOwner}}}
	uc := &StaffUseCase{Staff: repo, Users: users}

	err := uc.RemoveUser(authCtx(tenantID, domain.RoleOwner), tenantID, tenantUserID)
	if err == nil || err.Error() != "owner cannot be removed without ownership transfer" {
		t.Fatalf("expected owner removal rejection, got %v", err)
	}
}

func timePtr(v time.Time) *time.Time { return &v }
