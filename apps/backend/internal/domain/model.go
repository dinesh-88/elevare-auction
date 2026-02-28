package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Role string

const (
	RoleOwner      Role = "OWNER"
	RoleAdmin      Role = "ADMIN"
	RoleStaff      Role = "STAFF"
	RoleAccounting Role = "ACCOUNTING"
	RoleBidder     Role = "BIDDER"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
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
	DefaultCurrency        string
	Status                 TenantStatus
	Country                string
	EstimatedAuctionVolume *int
	Branding               json.RawMessage
	BrandingPublished      json.RawMessage
	ContentPublished       json.RawMessage
	BrandingDraft          json.RawMessage
	ContentDraft           json.RawMessage
	ContentVersion         int
	ContentSchemaVersion   int
	RequireAdminApproval   bool
	RequireKYC             bool
	CreatedAt              time.Time
}

type DemoRequest struct {
	ID          uuid.UUID
	Name        string
	Institution string
	Email       string
	Message     *string
	CreatedAt   time.Time
}

type User struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Email          string
	PasswordHash   string
	Role           Role
	Status         UserStatus
	EmailVerified  bool
	ApprovalStatus ApprovalStatus
	KYCStatus      KYCStatus
	CreatedAt      time.Time
}

const DemoBidSimulatorEmail = "simulator@elevare.local"

type Auction struct {
	ID                      uuid.UUID
	TenantID                uuid.UUID
	Currency                string
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
	RequireBidderApproval   bool
	CreatedAt               time.Time
}

type AuctionTerms struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	AuctionID       uuid.UUID
	VersionNumber   int
	ContentMarkdown string
	ContentHash     string
	IsActive        bool
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
}

type BidderTermsAcceptance struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	AuctionID          uuid.UUID
	UserID             uuid.UUID
	AuctionTermsID     uuid.UUID
	TermsVersionNumber int
	ContentHash        string
	AcceptedAt         time.Time
	AcceptedIP         string
	AcceptedUserAgent  string
	CreatedAt          time.Time
}

type TenantBidderStatus string

const (
	TenantBidderActive    TenantBidderStatus = "ACTIVE"
	TenantBidderSuspended TenantBidderStatus = "SUSPENDED"
)

type TenantBidderRiskLevel string

const (
	TenantBidderRiskLow    TenantBidderRiskLevel = "LOW"
	TenantBidderRiskNormal TenantBidderRiskLevel = "NORMAL"
	TenantBidderRiskHigh   TenantBidderRiskLevel = "HIGH"
)

type TenantBidderProfile struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	UserID        uuid.UUID
	Phone         *string
	CompanyName   *string
	AddressLine1  *string
	AddressLine2  *string
	City          *string
	PostalCode    *string
	Country       *string
	Status        TenantBidderStatus
	RiskLevel     TenantBidderRiskLevel
	InternalNotes *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AuctionBidderApproval struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	AuctionID          uuid.UUID
	UserID             uuid.UUID
	Status             ApprovalStatus
	RiskLevel          AuctionBidderRiskLevel
	MaxBidLimitMinor   *int64
	DepositRequired    bool
	DepositAmountMinor *int64
	DepositStatus      AuctionBidderDepositStatus
	InternalNotes      *string
	DecidedBy          *uuid.UUID
	DecidedAt          *time.Time
	ApprovedByUserID   *uuid.UUID
	ApprovedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type AuctionBidderAdminRow struct {
	UserID             uuid.UUID                  `json:"user_id"`
	Email              string                     `json:"email"`
	Approval           ApprovalStatus             `json:"approval_status"`
	TenantStatus       TenantBidderStatus         `json:"tenant_status"`
	RiskLevel          AuctionBidderRiskLevel     `json:"risk_level"`
	MaxBidLimitMinor   *int64                     `json:"max_bid_limit_minor,omitempty"`
	DepositRequired    bool                       `json:"deposit_required"`
	DepositAmountMinor *int64                     `json:"deposit_amount_minor,omitempty"`
	DepositStatus      AuctionBidderDepositStatus `json:"deposit_status"`
	InternalNotes      *string                    `json:"internal_notes,omitempty"`
	RequestedAt        time.Time                  `json:"requested_at"`
	DecidedAt          *time.Time                 `json:"decided_at,omitempty"`
}

type AuctionBidderRiskLevel string

const (
	AuctionBidderRiskLow    AuctionBidderRiskLevel = "low"
	AuctionBidderRiskMedium AuctionBidderRiskLevel = "medium"
	AuctionBidderRiskHigh   AuctionBidderRiskLevel = "high"
)

type AuctionBidderDepositStatus string

const (
	AuctionBidderDepositNotRequired AuctionBidderDepositStatus = "not_required"
	AuctionBidderDepositRequired    AuctionBidderDepositStatus = "required"
	AuctionBidderDepositPending     AuctionBidderDepositStatus = "pending"
	AuctionBidderDepositReceived    AuctionBidderDepositStatus = "received"
	AuctionBidderDepositWaived      AuctionBidderDepositStatus = "waived"
)

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

type FeeCalculationType string

const (
	FeeCalculationPercentage FeeCalculationType = "PERCENTAGE"
	FeeCalculationFixed      FeeCalculationType = "FIXED"
)

type FeeAppliesTo string

const (
	FeeAppliesToHammer   FeeAppliesTo = "HAMMER"
	FeeAppliesToPremium  FeeAppliesTo = "PREMIUM"
	FeeAppliesToSubtotal FeeAppliesTo = "SUBTOTAL"
)

type FeeScopeType string

const (
	FeeScopeTenant  FeeScopeType = "TENANT"
	FeeScopeAuction FeeScopeType = "AUCTION"
	FeeScopeLot     FeeScopeType = "LOT"
)

type FeeComponent struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Name            string
	CalculationType FeeCalculationType
	Value           float64
	AppliesTo       FeeAppliesTo
	Taxable         bool
	Active          bool
	CreatedAt       time.Time
}

type FeeAssignment struct {
	ID             uuid.UUID
	ScopeType      FeeScopeType
	ScopeID        uuid.UUID
	FeeComponentID uuid.UUID
	CreatedAt      time.Time
}

type LotImage struct {
	ID       uuid.UUID
	LotID    uuid.UUID
	URL      string
	Position int
}

type TenantAsset struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Kind      string    `json:"kind"`
	URL       string    `json:"url"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	SizeBytes *int64    `json:"size_bytes,omitempty"`
	MIME      string    `json:"mime,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Bid struct {
	ID        uuid.UUID
	AuctionID uuid.UUID
	LotID     *uuid.UUID
	UserID    uuid.UUID
	Amount    int64
	MaxBid    *int64
	CreatedAt time.Time
}

type InvoiceStatus string

const (
	InvoiceIssued    InvoiceStatus = "ISSUED"
	InvoicePaid      InvoiceStatus = "PAID"
	InvoiceCancelled InvoiceStatus = "CANCELLED"
)

type InvoicePaymentStatus string

const (
	InvoicePaymentUnpaid    InvoicePaymentStatus = "UNPAID"
	InvoicePaymentPaid      InvoicePaymentStatus = "PAID"
	InvoicePaymentCancelled InvoicePaymentStatus = "CANCELLED"
)

type Invoice struct {
	ID            uuid.UUID
	LotID         uuid.UUID
	AuctionID     uuid.UUID
	TenantID      uuid.UUID
	WinnerUserID  uuid.UUID
	InvoiceNumber string
	IssuedAt      time.Time
	DueDate       *time.Time
	Currency      string
	TotalAmount   int64
	HammerMinor   int64
	FeesMinor     int64
	TaxMinor      int64
	TotalDueMinor int64
	PaymentStatus InvoicePaymentStatus
	PaidAt        *time.Time
	CancelledAt   *time.Time
	PaymentRef    *string
	CancelReason  *string
	Snapshot      json.RawMessage
	Status        InvoiceStatus
	CreatedAt     time.Time

	LotNumber    string
	LotTitle     string
	AuctionTitle string
}

type AuctionSettlementSummary struct {
	AuctionID        uuid.UUID
	Currency         string
	InvoiceCount     int64
	PaidCount        int64
	UnpaidCount      int64
	CancelledCount   int64
	TotalHammerMinor int64
	TotalFeesMinor   int64
	TotalTaxMinor    int64
	TotalDueMinor    int64
	TotalPaidMinor   int64
	OutstandingMinor int64
}

type SettlementInvoiceRow struct {
	ID               uuid.UUID
	InvoiceNumber    string
	BidderName       string
	BidderEmail      string
	LotSummary       string
	HammerMinor      int64
	FeesMinor        int64
	TaxMinor         int64
	TotalDueMinor    int64
	Currency         string
	PaymentStatus    InvoicePaymentStatus
	IssuedAt         time.Time
	PaidAt           *time.Time
	PaymentReference *string
	CancelReason     *string
}
