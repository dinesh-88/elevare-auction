package domain

import "testing"

func TestRolePermissions_SeparatesStaffAndOwnerControls(t *testing.T) {
	if !HasPermission(RoleOwner, PermissionRoleManage) {
		t.Fatal("expected owner to retain role management permission")
	}
	if HasPermission(RoleAdmin, PermissionRoleManage) {
		t.Fatal("expected admin to be blocked from role management permission")
	}
	if !HasPermission(RoleAdmin, PermissionStaffInvite) {
		t.Fatal("expected admin to be allowed to invite staff")
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
