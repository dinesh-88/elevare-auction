package domain

import "testing"

func TestRolePermissions_OwnerAllowedForAllAdminPermissions(t *testing.T) {
	perms := []Permission{
		PermissionStaffManage,
		PermissionAuctionCreate,
		PermissionAuctionEdit,
		PermissionAuctionPublish,
		PermissionAuctionViewAdmin,
		PermissionBidderApprove,
		PermissionFinancialView,
		PermissionSettlementApprove,
		PermissionInvoiceSend,
		PermissionVATManage,
		PermissionFeesManage,
		PermissionAuditView,
		PermissionCMSManage,
		PermissionAnalyticsView,
	}

	for _, perm := range perms {
		if !HasPermission(RoleOwner, perm) {
			t.Fatalf("expected owner to have %s", perm)
		}
	}
}

func TestRolePermissions_AdminCanManageStaffInvites(t *testing.T) {
	if !HasPermission(RoleAdmin, PermissionStaffManage) {
		t.Fatal("expected admin to have staff.manage")
	}
}

func TestRolePermissions_AccountingCanApproveSettlementButCannotPublishAuction(t *testing.T) {
	if !HasPermission(RoleAccounting, PermissionSettlementApprove) {
		t.Fatal("expected accounting to approve settlement")
	}
	if HasPermission(RoleAccounting, PermissionAuctionPublish) {
		t.Fatal("expected accounting to be blocked from auction.publish")
	}
}

func TestRolePermissions_ViewerCannotMutate(t *testing.T) {
	if !HasPermission(RoleViewer, PermissionAuctionViewAdmin) {
		t.Fatal("expected viewer to retain auction.view_admin")
	}
	if HasPermission(RoleViewer, PermissionAuctionCreate) {
		t.Fatal("expected viewer to be blocked from auction.create")
	}
	if HasPermission(RoleViewer, PermissionStaffManage) {
		t.Fatal("expected viewer to be blocked from staff.manage")
	}
}

func TestRolePermissions_BidderActionsStayBidderOnly(t *testing.T) {
	if !HasPermission(RoleBidder, PermissionBidPlace) {
		t.Fatal("expected bidder to be able to place bids")
	}
	if !HasPermission(RoleBidder, PermissionAuctionTermsAccept) {
		t.Fatal("expected bidder to be able to accept auction terms")
	}
	if HasPermission(RoleViewer, PermissionBidPlace) {
		t.Fatal("expected viewer to be blocked from placing bids")
	}
}
