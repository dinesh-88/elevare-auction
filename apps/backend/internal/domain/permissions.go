package domain

import "github.com/google/uuid"

type Permission string

const (
	PermissionAuctionCreate Permission = "auction:create"
	PermissionAuctionUpdate Permission = "auction:update"
	PermissionAuctionStart  Permission = "auction:start"
	PermissionAuctionEnd    Permission = "auction:end"

	PermissionLotCreate  Permission = "lot:create"
	PermissionLotUpdate  Permission = "lot:update"
	PermissionLotReorder Permission = "lot:reorder"
	PermissionLotStart   Permission = "lot:start"
	PermissionLotEnd     Permission = "lot:end"

	PermissionFeesManage      Permission = "fees:manage"
	PermissionInvoiceView     Permission = "invoice:view"
	PermissionInvoiceMarkPaid Permission = "invoice:mark_paid"

	PermissionBidderApprove Permission = "bidder:approve"
	PermissionBidderSuspend Permission = "bidder:suspend"

	PermissionSiteEdit    Permission = "site:edit"
	PermissionSitePublish Permission = "site:publish"

	PermissionAuditView Permission = "audit:view"
)

var allTenantPermissions = []Permission{
	PermissionAuctionCreate,
	PermissionAuctionUpdate,
	PermissionAuctionStart,
	PermissionAuctionEnd,
	PermissionLotCreate,
	PermissionLotUpdate,
	PermissionLotReorder,
	PermissionLotStart,
	PermissionLotEnd,
	PermissionFeesManage,
	PermissionInvoiceView,
	PermissionInvoiceMarkPaid,
	PermissionBidderApprove,
	PermissionBidderSuspend,
	PermissionSiteEdit,
	PermissionSitePublish,
	PermissionAuditView,
}

var RolePermissions = map[Role]map[Permission]struct{}{
	RoleOwner:      permissionSet(allTenantPermissions...),
	RoleAdmin:      permissionSet(allTenantPermissions...),
	RoleStaff:      permissionSet(PermissionAuctionCreate, PermissionAuctionUpdate, PermissionAuctionStart, PermissionAuctionEnd, PermissionLotCreate, PermissionLotUpdate, PermissionLotReorder, PermissionLotStart, PermissionLotEnd, PermissionBidderApprove, PermissionSiteEdit),
	RoleAccounting: permissionSet(PermissionInvoiceView, PermissionInvoiceMarkPaid, PermissionAuditView),
	RoleBidder:     permissionSet(),
}

func permissionSet(perms ...Permission) map[Permission]struct{} {
	out := make(map[Permission]struct{}, len(perms))
	for _, perm := range perms {
		out[perm] = struct{}{}
	}
	return out
}

func HasPermission(role Role, permission Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	_, ok = perms[permission]
	return ok
}

func PermissionsForRole(role Role) []Permission {
	perms := RolePermissions[role]
	out := make([]Permission, 0, len(perms))
	for perm := range perms {
		out = append(out, perm)
	}
	return out
}

func IsStaffRole(role Role) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleStaff, RoleAccounting:
		return true
	default:
		return false
	}
}

type UserContext struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Role     Role
}
