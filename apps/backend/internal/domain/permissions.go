package domain

import "github.com/google/uuid"

type Permission string

const (
	PermissionAuctionCreate      Permission = "auction:create"
	PermissionAuctionUpdate      Permission = "auction:update"
	PermissionAuctionStart       Permission = "auction:start"
	PermissionAuctionEnd         Permission = "auction:end"
	PermissionAuctionTermsAccept Permission = "auction_terms:accept"

	PermissionLotCreate  Permission = "lot:create"
	PermissionLotUpdate  Permission = "lot:update"
	PermissionLotReorder Permission = "lot:reorder"
	PermissionLotStart   Permission = "lot:start"
	PermissionLotEnd     Permission = "lot:end"
	PermissionBidPlace   Permission = "bid:place"

	PermissionFeesManage      Permission = "fees:manage"
	PermissionInvoiceView     Permission = "invoice:view"
	PermissionInvoiceMarkPaid Permission = "invoice:mark_paid"

	PermissionBidderApprove Permission = "bidder:approve"
	PermissionBidderSuspend Permission = "bidder:suspend"

	PermissionSiteEdit    Permission = "site:edit"
	PermissionSitePublish Permission = "site:publish"

	PermissionStaffInvite Permission = "staff:invite"
	PermissionAuditView   Permission = "audit:view"
	PermissionRoleManage  Permission = "role:manage"
)

var ownerTenantPermissions = []Permission{
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
	PermissionStaffInvite,
	PermissionAuditView,
	PermissionRoleManage,
}

var adminTenantPermissions = []Permission{
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
	PermissionStaffInvite,
	PermissionAuditView,
}

var RolePermissions = map[Role]map[Permission]struct{}{
	RoleOwner:      permissionSet(ownerTenantPermissions...),
	RoleAdmin:      permissionSet(adminTenantPermissions...),
	RoleStaff:      permissionSet(PermissionAuctionCreate, PermissionAuctionUpdate, PermissionAuctionStart, PermissionAuctionEnd, PermissionLotCreate, PermissionLotUpdate, PermissionLotReorder, PermissionLotStart, PermissionLotEnd, PermissionBidderApprove, PermissionSiteEdit),
	RoleAccounting: permissionSet(PermissionInvoiceView, PermissionInvoiceMarkPaid, PermissionAuditView),
	RoleViewer:     permissionSet(),
	RoleBidder:     permissionSet(PermissionAuctionTermsAccept, PermissionBidPlace),
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
