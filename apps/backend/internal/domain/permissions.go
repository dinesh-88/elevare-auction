package domain

import "github.com/google/uuid"

type Permission string

const (
	PermissionStaffManage       Permission = "staff.manage"
	PermissionAuctionCreate     Permission = "auction.create"
	PermissionAuctionEdit       Permission = "auction.edit"
	PermissionAuctionPublish    Permission = "auction.publish"
	PermissionAuctionViewAdmin  Permission = "auction.view_admin"
	PermissionBidderApprove     Permission = "bidder.approve"
	PermissionFinancialView     Permission = "financial.view"
	PermissionSettlementApprove Permission = "settlement.approve"
	PermissionInvoiceSend       Permission = "invoice.send"
	PermissionVATManage         Permission = "vat.manage"
	PermissionFeesManage        Permission = "fees.manage"
	PermissionAuditView         Permission = "audit.view"
	PermissionCMSManage         Permission = "cms.manage"
	PermissionAnalyticsView     Permission = "analytics.view"

	PermissionAuctionTermsAccept Permission = "auction_terms.accept"
	PermissionBidPlace           Permission = "bid.place"

	PermissionAuctionUpdate   Permission = PermissionAuctionEdit
	PermissionAuctionStart    Permission = PermissionAuctionPublish
	PermissionAuctionEnd      Permission = PermissionAuctionPublish
	PermissionLotCreate       Permission = PermissionAuctionEdit
	PermissionLotUpdate       Permission = PermissionAuctionEdit
	PermissionLotReorder      Permission = PermissionAuctionEdit
	PermissionLotStart        Permission = PermissionAuctionPublish
	PermissionLotEnd          Permission = PermissionAuctionPublish
	PermissionInvoiceView     Permission = PermissionFinancialView
	PermissionInvoiceMarkPaid Permission = PermissionSettlementApprove
	PermissionBidderSuspend   Permission = PermissionBidderApprove
	PermissionSiteEdit        Permission = PermissionCMSManage
	PermissionSitePublish     Permission = PermissionCMSManage
	PermissionStaffInvite     Permission = PermissionStaffManage
	PermissionRoleManage      Permission = PermissionStaffManage
)

var ownerTenantPermissions = []Permission{
	PermissionStaffManage,
	PermissionAuctionCreate,
	PermissionAuctionEdit,
	PermissionAuctionPublish,
	PermissionAuctionViewAdmin,
	PermissionFeesManage,
	PermissionVATManage,
	PermissionFinancialView,
	PermissionSettlementApprove,
	PermissionInvoiceSend,
	PermissionBidderApprove,
	PermissionAuditView,
	PermissionCMSManage,
	PermissionAnalyticsView,
}

var adminTenantPermissions = []Permission{
	PermissionStaffManage,
	PermissionAuctionCreate,
	PermissionAuctionEdit,
	PermissionAuctionPublish,
	PermissionAuctionViewAdmin,
	PermissionFeesManage,
	PermissionVATManage,
	PermissionFinancialView,
	PermissionSettlementApprove,
	PermissionInvoiceSend,
	PermissionBidderApprove,
	PermissionAuditView,
	PermissionCMSManage,
	PermissionAnalyticsView,
}

var RolePermissions = map[Role]map[Permission]struct{}{
	RoleOwner:      permissionSet(ownerTenantPermissions...),
	RoleAdmin:      permissionSet(adminTenantPermissions...),
	RoleStaff:      permissionSet(PermissionAuctionCreate, PermissionAuctionEdit, PermissionAuctionViewAdmin, PermissionAnalyticsView),
	RoleAccounting: permissionSet(PermissionFinancialView, PermissionSettlementApprove, PermissionInvoiceSend, PermissionAnalyticsView),
	RoleViewer:     permissionSet(PermissionAuctionViewAdmin, PermissionAnalyticsView),
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
	case RoleOwner, RoleAdmin, RoleStaff, RoleAccounting, RoleViewer:
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
