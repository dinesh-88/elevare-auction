package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
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

type AuctionMode string

const (
	AuctionModeTimed      AuctionMode = "TIMED"
	AuctionModeSequential AuctionMode = "SEQUENTIAL"
)

type BidVisibility string

const (
	BidVisibilityPublic    BidVisibility = "PUBLIC"
	BidVisibilityAnonymous BidVisibility = "ANONYMOUS"
	BidVisibilitySealed    BidVisibility = "SEALED"
)

type TenantStatus string

const (
	TenantPending  TenantStatus = "PENDING"
	TenantActive   TenantStatus = "ACTIVE"
	TenantRejected TenantStatus = "REJECTED"
)

type Tenant struct {
	ID                     uuid.UUID
	Name                   string
	Subdomain              string
	CustomDomain           *string
	Status                 TenantStatus
	Country                string
	EstimatedAuctionVolume *int
	Branding               json.RawMessage
	RequireAdminApproval   bool
	RequireKYC             bool
	CreatedAt              time.Time
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
	AuctionMode             AuctionMode
	AutoProgressEnabled     bool
	LotDurationSeconds      int
	SnipingWindowSeconds    int
	SnipingExtensionSeconds int
	BidVisibility           BidVisibility
	CreatedAt               time.Time
}

type LotStatus string

const (
	LotPending   LotStatus = "PENDING"
	LotLive      LotStatus = "LIVE"
	LotPaused    LotStatus = "PAUSED"
	LotEnded     LotStatus = "ENDED"
	LotCancelled LotStatus = "CANCELLED"
)

type Lot struct {
	ID            uuid.UUID
	AuctionID     uuid.UUID
	LotNumber     string
	OrderIndex    int
	Title         string
	Description   string
	StartingPrice int64
	CurrentPrice  int64
	ReservePrice  *int64
	MinIncrement  int64
	Metadata      json.RawMessage
	Status        LotStatus
	HammerPrice   *int64
	TotalDue      *int64
	WinnerUserID  *uuid.UUID
	FinancialData json.RawMessage
	StartedAt     *time.Time
	EndedAt       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type FeeType string

const (
	FeeTypePercentage FeeType = "PERCENTAGE"
	FeeTypeFixed      FeeType = "FIXED"
)

type FinancialFee struct {
	Name      string  `json:"name"`
	Type      FeeType `json:"type"`
	Value     float64 `json:"value"`
	Amount    int64   `json:"amount"`
	AppliesTo string  `json:"applies_to,omitempty"`
	Taxable   bool    `json:"taxable"`
}

type LotFinancialSnapshot struct {
	HammerPrice int64          `json:"hammer_price"`
	Fees        []FinancialFee `json:"fees"`
	TotalDue    int64          `json:"total_due"`
}

type LotImage struct {
	ID       uuid.UUID
	LotID    uuid.UUID
	URL      string
	Position int
}

type Bid struct {
	ID        uuid.UUID
	AuctionID uuid.UUID
	LotID     *uuid.UUID
	UserID    uuid.UUID
	Amount    int64
	CreatedAt time.Time
}
