package usecase

import (
	"errors"
	"testing"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
)

func TestStaffInvite_AllowsAdminPermissionPath(t *testing.T) {
	tenantID := uuid.New()
	uc := &StaffUseCase{}

	_, _, err := uc.Invite(authCtx(tenantID, domain.RoleAdmin), InviteStaffInput{
		TenantID:  tenantID,
		Email:     "invalid-email",
		Role:      domain.RoleStaff,
		InvitedBy: uuid.New(),
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation after permission check, got %v", err)
	}
}

func TestStaffUpdateRole_RejectsAdminWithoutOwnerPermission(t *testing.T) {
	tenantID := uuid.New()
	uc := &StaffUseCase{}

	err := uc.UpdateRole(authCtx(tenantID, domain.RoleAdmin), tenantID, uuid.New(), domain.RoleStaff)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
