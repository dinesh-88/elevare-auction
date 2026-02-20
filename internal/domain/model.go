package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleBidder Role = "BIDDER"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "PENDING"
	ApprovalApproved ApprovalStatus = "APPROVED"
	ApprovalRejected ApprovalStatus = "REJECTED"
)

type KYCStatus string

const (
	KYCNone     KYCStatus = "NONE"
	KYCPending  KYCStatus = "PENDING"
	KYCVerified KYCStatus = "VERIFIED"
	KYCRejected KYCStatus = "REJECTED"
)

type AuctionStatus string

const (
	AuctionDraft     AuctionStatus = "DRAFT"
	AuctionScheduled AuctionStatus = "SCHEDULED"
	AuctionLive      AuctionStatus = "LIVE"
	AuctionPaused    AuctionStatus = "PAUSED"
	AuctionEnded     AuctionStatus = "ENDED"
	AuctionCancelled AuctionStatus = "CANCELLED"
)

type BidVisibility string

const (
	BidVisibilityPublic    BidVisibility = "PUBLIC"
	BidVisibilityAnonymous BidVisibility = "ANONYMOUS"
	BidVisibilitySealed    BidVisibility = "SEALED"
)

type Tenant struct {
	ID                   uuid.UUID
	Name                 string
	Subdomain            string
	CustomDomain         *string
	Branding             json.RawMessage
	RequireAdminApproval bool
	RequireKYC           bool
	CreatedAt            time.Time
}

type User struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Email          string
	PasswordHash   string
	Role           Role
	EmailVerified  bool
	ApprovalStatus ApprovalStatus
	KYCStatus      KYCStatus
	CreatedAt      time.Time
}

type Auction struct {
	ID                      uuid.UUID
	TenantID                uuid.UUID
	Title                   string
	Description             string
	Metadata                json.RawMessage
	StartingPrice           int64
	CurrentPrice            int64
	ReservePrice            *int64
	MinIncrement            int64
	StartTime               time.Time
	EndTime                 time.Time
	Status                  AuctionStatus
	AllowManualStart        bool
	AllowManualPause        bool
	AllowManualEnd          bool
	AutoExtendEnabled       bool
	SnipingWindowSeconds    int
	SnipingExtensionSeconds int
	BidVisibility           BidVisibility
	CreatedAt               time.Time
}

type Bid struct {
	ID        uuid.UUID
	AuctionID uuid.UUID
	UserID    uuid.UUID
	Amount    int64
	CreatedAt time.Time
}
