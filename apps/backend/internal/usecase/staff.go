package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type StaffUseCase struct {
	Staff       port.StaffRepository
	Users       port.UserRepository
	Hasher      port.PasswordHasher
	Tokens      port.TokenService
	EmailOutbox *EmailOutboxService
	Audit       port.AuditRepository
}

type InviteStaffInput struct {
	TenantID        uuid.UUID
	Email           string
	Role            domain.Role
	InvitedBy       uuid.UUID
	InviteBase      string
	InstitutionName string
}

type AcceptInvitationInput struct {
	TenantID uuid.UUID
	Token    string
	Password string
	Locale   string
}

func (uc *StaffUseCase) List(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantUser, []domain.TenantInvitation, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionStaffInvite); err != nil {
		return nil, nil, ErrForbidden
	}
	users, err := uc.Staff.ListTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	invitations, err := uc.Staff.ListInvitations(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	return users, invitations, nil
}

func (uc *StaffUseCase) Invite(ctx context.Context, in InviteStaffInput) (*domain.TenantInvitation, string, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, "", err
	}
	if err := authz.Authorize(ctx, domain.PermissionStaffInvite); err != nil {
		return nil, "", ErrForbidden
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if in.TenantID == uuid.Nil || in.InvitedBy == uuid.Nil || !strings.Contains(email, "@") {
		return nil, "", ErrValidation
	}
	if !isInvitableRole(in.Role) {
		return nil, "", ErrValidation
	}
	if existingUser, userErr := uc.Users.GetByEmail(ctx, in.TenantID, email); userErr == nil && existingUser != nil {
		if membership, membershipErr := uc.Staff.GetTenantUserByUserID(ctx, in.TenantID, existingUser.ID); membershipErr == nil && membership != nil {
			return nil, "", errors.New("user is already a staff member")
		}
	}
	rawToken := uuid.NewString() + uuid.NewString()
	item := &domain.TenantInvitation{
		ID:        uuid.New(),
		TenantID:  in.TenantID,
		Email:     email,
		Role:      in.Role,
		TokenHash: invitationTokenHash(rawToken),
		InvitedBy: &in.InvitedBy,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := uc.Staff.CreateInvitation(ctx, item); err != nil {
		return nil, "", err
	}
	if uc.EmailOutbox != nil {
		subject := "You're invited to join the team"
		link := strings.TrimSpace(in.InviteBase)
		if link != "" {
			link = strings.TrimRight(link, "/") + "?token=" + rawToken
		}
		institutionName := strings.TrimSpace(in.InstitutionName)
		if institutionName == "" {
			institutionName = "your institution"
		}
		html := fmt.Sprintf("<p>%s invited you to join as %s.</p><p>%s</p>", institutionName, strings.ToLower(string(in.Role)), link)
		text := fmt.Sprintf("%s invited you to join as %s. %s", institutionName, strings.ToLower(string(in.Role)), link)
		_, _ = uc.EmailOutbox.Queue(ctx, QueueEmailInput{
			TenantID: in.TenantID,
			ToEmail:  email,
			Subject:  subject,
			HTMLBody: html,
			TextBody: text,
			Type:     "staff_invite",
			Critical: true,
		})
	}
	uc.appendAudit(ctx, in.TenantID, "STAFF_INVITE_CREATED", &item.ID, map[string]any{"email": item.Email, "role": item.Role})
	return item, rawToken, nil
}

func (uc *StaffUseCase) AcceptInvitation(ctx context.Context, in AcceptInvitationInput) (string, *domain.User, error) {
	if in.TenantID == uuid.Nil || strings.TrimSpace(in.Token) == "" {
		return "", nil, ErrValidation
	}
	invite, err := uc.Staff.GetInvitationByTokenHash(ctx, in.TenantID, invitationTokenHash(in.Token))
	if err != nil {
		return "", nil, err
	}
	if invite.AcceptedAt != nil {
		return "", nil, errors.New("invitation already accepted")
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		return "", nil, errors.New("invitation expired")
	}
	if !isInvitableRole(invite.Role) {
		return "", nil, ErrValidation
	}

	var user *domain.User
	user, err = uc.Users.GetByEmail(ctx, in.TenantID, invite.Email)
	if err != nil {
		if len(in.Password) < 8 {
			return "", nil, ErrValidation
		}
		hash, hashErr := uc.Hasher.Hash(in.Password)
		if hashErr != nil {
			return "", nil, hashErr
		}
		user = &domain.User{
			ID:             uuid.New(),
			TenantID:       in.TenantID,
			Email:          invite.Email,
			PasswordHash:   hash,
			Locale:         domain.NormalizeLocale(in.Locale),
			Role:           invite.Role,
			Status:         domain.UserStatusActive,
			EmailVerified:  true,
			ApprovalStatus: domain.ApprovalApproved,
			KYCStatus:      domain.KYCNone,
		}
		if user.Locale == "" {
			user.Locale = domain.LocaleEnglish
		}
		if err := uc.Users.Create(ctx, in.TenantID, user); err != nil {
			return "", nil, err
		}
	} else {
		membership, membershipErr := uc.Staff.GetTenantUserByUserID(ctx, in.TenantID, user.ID)
		if membershipErr == nil && membership != nil {
			return "", nil, errors.New("user is already a staff member")
		}
	}

	item := &domain.TenantUser{
		ID:        uuid.New(),
		TenantID:  in.TenantID,
		UserID:    user.ID,
		Role:      invite.Role,
		InvitedBy: invite.InvitedBy,
		CreatedAt: time.Now().UTC(),
	}
	if err := uc.Staff.CreateTenantUser(ctx, item); err != nil {
		return "", nil, err
	}
	if err := uc.Users.UpdateRole(ctx, in.TenantID, user.ID, invite.Role); err != nil {
		return "", nil, err
	}
	if err := uc.Staff.MarkInvitationAccepted(ctx, in.TenantID, invite.ID, time.Now().UTC()); err != nil {
		return "", nil, err
	}
	uc.appendAudit(ctx, in.TenantID, "STAFF_INVITE_ACCEPTED", &item.ID, map[string]any{"user_id": user.ID, "email": invite.Email, "role": invite.Role})
	fresh, err := uc.Users.GetByID(ctx, in.TenantID, user.ID)
	if err != nil {
		return "", nil, err
	}
	token, err := uc.Tokens.Issue(fresh.ID, fresh.TenantID, fresh.Role, fresh.Locale)
	if err != nil {
		return "", nil, err
	}
	return token, fresh, nil
}

func (uc *StaffUseCase) UpdateRole(ctx context.Context, tenantID, tenantUserID uuid.UUID, role domain.Role) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionRoleManage); err != nil {
		return ErrForbidden
	}
	if err := requireOwnerRole(ctx); err != nil {
		return err
	}
	if !isInvitableRole(role) {
		return ErrValidation
	}
	current, err := uc.Staff.GetTenantUserByID(ctx, tenantID, tenantUserID)
	if err != nil {
		return err
	}
	if current.Role == domain.RoleOwner {
		return errors.New("owner cannot be changed without ownership transfer")
	}
	if current.IsPrimaryOwner {
		return errors.New("primary owner cannot be changed without ownership transfer")
	}
	if err := uc.Staff.UpdateTenantUserRole(ctx, tenantID, tenantUserID, role); err != nil {
		return err
	}
	if err := uc.Users.UpdateRole(ctx, tenantID, current.UserID, role); err != nil {
		return err
	}
	uc.appendAudit(ctx, tenantID, "STAFF_ROLE_CHANGED", &current.ID, map[string]any{"user_id": current.UserID, "role": role})
	return nil
}

func (uc *StaffUseCase) RevokeInvitation(ctx context.Context, tenantID, invitationID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionStaffManage); err != nil {
		return ErrForbidden
	}
	invite, err := uc.Staff.GetInvitationByID(ctx, tenantID, invitationID)
	if err != nil {
		return err
	}
	if err := uc.Staff.DeleteInvitation(ctx, tenantID, invitationID); err != nil {
		return err
	}
	uc.appendAudit(ctx, tenantID, "STAFF_INVITE_REVOKED", &invite.ID, map[string]any{"email": invite.Email, "role": invite.Role})
	return nil
}

func (uc *StaffUseCase) RemoveUser(ctx context.Context, tenantID, tenantUserID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionStaffManage); err != nil {
		return ErrForbidden
	}
	if err := requireOwnerRole(ctx); err != nil {
		return err
	}
	current, err := uc.Staff.GetTenantUserByID(ctx, tenantID, tenantUserID)
	if err != nil {
		return err
	}
	if current.Role == domain.RoleOwner || current.IsPrimaryOwner {
		return errors.New("owner cannot be removed without ownership transfer")
	}
	if err := uc.Staff.DeleteTenantUser(ctx, tenantID, tenantUserID); err != nil {
		return err
	}
	if err := uc.Users.UpdateRole(ctx, tenantID, current.UserID, domain.RoleBidder); err != nil {
		return err
	}
	uc.appendAudit(ctx, tenantID, "STAFF_REMOVED", &current.ID, map[string]any{"user_id": current.UserID})
	return nil
}

func isInvitableRole(role domain.Role) bool {
	switch role {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleAccounting, domain.RoleViewer:
		return true
	default:
		return false
	}
}

func invitationTokenHash(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func requireOwnerRole(ctx context.Context) error {
	uc, ok := middleware.UserContextFromContext(ctx)
	if !ok || uc.Role != domain.RoleOwner {
		return ErrForbidden
	}
	return nil
}

func (uc *StaffUseCase) appendAudit(ctx context.Context, tenantID uuid.UUID, action string, entityID *uuid.UUID, metadata map[string]any) {
	if uc.Audit == nil {
		return
	}
	auditCtx := middleware.AuditContextFromContext(ctx)
	payload, _ := json.Marshal(metadata)
	_ = uc.Audit.Append(ctx, port.AuditEntry{
		TenantID:    &tenantID,
		ActorUserID: auditCtx.ActorUserID,
		ActorRole:   auditCtx.ActorRole,
		Action:      action,
		EntityType:  "STAFF",
		EntityID:    entityID,
		RequestID:   auditCtx.RequestID,
		IP:          auditCtx.IP,
		UserAgent:   auditCtx.UserAgent,
		Metadata:    payload,
	})
}
