package port

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TenantRepository interface {
	GetByHost(ctx context.Context, host string) (*domain.Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error)
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	CreatePendingTenantWithAdmin(ctx context.Context, in CreatePendingTenantWithAdminInput) error
	ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error)
	UpdateStatusAndAdminApproval(ctx context.Context, tenantID uuid.UUID, tenantStatus domain.TenantStatus, adminApproval domain.ApprovalStatus) error
	UpdateProfile(ctx context.Context, tenantID uuid.UUID, name, country string) error
	UpdateDefaultCurrency(ctx context.Context, tenantID uuid.UUID, currency string) error
	UpdateSiteDraft(ctx context.Context, tenantID uuid.UUID, brandingDraft, contentDraft json.RawMessage) error
	PublishSiteDraft(ctx context.Context, tenantID uuid.UUID) (beforeVersion int, afterVersion int, err error)
	UpdateOnboardingState(ctx context.Context, tenantID uuid.UUID, status domain.OnboardingStatus, step int, completedAt *time.Time) error
}

type AssetRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, asset *domain.TenantAsset) error
	ListByKind(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TenantAsset, error)
}

type DemoRequestRepository interface {
	Create(ctx context.Context, item *domain.DemoRequest) error
}

type UserRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error)
	GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error)
	UpdateLocale(ctx context.Context, tenantID, userID uuid.UUID, locale domain.Locale) error
	UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error
	UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error
	UpdateRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.Role) error
}

type TenantLoginCandidate struct {
	TenantID     uuid.UUID
	TenantName   string
	Subdomain    string
	CustomDomain *string
}

type UserTenantLookupRepository interface {
	ListTenantLoginCandidatesByEmail(ctx context.Context, email string) ([]TenantLoginCandidate, error)
}

type PermissionRepository interface {
	HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission domain.Permission) (bool, error)
	GetEffectiveRole(ctx context.Context, tenantID, userID uuid.UUID) (domain.Role, error)
}

type StaffRepository interface {
	PermissionRepository
	CreateTenantUser(ctx context.Context, item *domain.TenantUser) error
	ListTenantUsers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantUser, error)
	GetTenantUserByID(ctx context.Context, tenantID, tenantUserID uuid.UUID) (*domain.TenantUser, error)
	GetTenantUserByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantUser, error)
	UpdateTenantUserRole(ctx context.Context, tenantID, tenantUserID uuid.UUID, role domain.Role) error
	DeleteTenantUser(ctx context.Context, tenantID, tenantUserID uuid.UUID) error

	CreateInvitation(ctx context.Context, item *domain.TenantInvitation) error
	ListInvitations(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantInvitation, error)
	GetInvitationByID(ctx context.Context, tenantID, invitationID uuid.UUID) (*domain.TenantInvitation, error)
	GetInvitationByTokenHash(ctx context.Context, tenantID uuid.UUID, tokenHash string) (*domain.TenantInvitation, error)
	MarkInvitationAccepted(ctx context.Context, tenantID, invitationID uuid.UUID, acceptedAt time.Time) error
	DeleteInvitation(ctx context.Context, tenantID, invitationID uuid.UUID) error
}

type AuctionRepository interface {
	Create(ctx context.Context, auction *domain.Auction) error
	ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error)
	ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error)
	GetByID(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error)
	Update(ctx context.Context, auction *domain.Auction) error
	Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error
	UpdateStatus(ctx context.Context, tenantID, auctionID uuid.UUID, status domain.AuctionStatus) error
	UpdateRequireBidderApproval(ctx context.Context, tenantID, auctionID uuid.UUID, required bool) error
	RunSchedulerTick(ctx context.Context, now time.Time) error
}

type AuctionTermsRepository interface {
	ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID, locale domain.Locale) ([]domain.AuctionTerms, error)
	GetActiveByAuction(ctx context.Context, tenantID, auctionID uuid.UUID, locale domain.Locale) (*domain.AuctionTerms, error)
	GetAcceptance(ctx context.Context, tenantID, auctionID, userID, auctionTermsID uuid.UUID) (*domain.BidderTermsAcceptance, error)
	CreateVersion(ctx context.Context, item *domain.AuctionTerms) error
	PublishVersion(ctx context.Context, tenantID, auctionID, auctionTermsID uuid.UUID, locale domain.Locale) error
	CreateAcceptance(ctx context.Context, item *domain.BidderTermsAcceptance) error
}

type LotRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error
	Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error
	Delete(ctx context.Context, tenantID, lotID uuid.UUID) error
	Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error
	ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error)
	GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error)
	Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error
	End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*LotSettlementResult, error)
	OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error
	PlaceBidTx(ctx context.Context, in PlaceLotBidTxInput) (*PlaceLotBidTxResult, error)
	ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error)
	ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error)
	RunSchedulerTick(ctx context.Context, now time.Time) error
}

type BidderRepository interface {
	EnsureTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error)
	GetTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error)
	UpdateTenantBidder(ctx context.Context, tenantID, userID uuid.UUID, in UpdateTenantBidderInput) (*domain.TenantBidderProfile, error)
	RequestAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, bool, error)
	GetAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID) (*domain.AuctionBidderApproval, error)
	SetAuctionApproval(ctx context.Context, tenantID, auctionID, userID, decidedBy uuid.UUID, status domain.ApprovalStatus, now time.Time) (*domain.AuctionBidderApproval, error)
	UpdateAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, in UpdateAuctionBidderApprovalInput, now time.Time) (*domain.AuctionBidderApproval, error)
	ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, filter ListAuctionBiddersFilter) ([]domain.AuctionBidderAdminRow, error)
}

type UpdateTenantBidderInput struct {
	Status        domain.TenantBidderStatus
	RiskLevel     domain.TenantBidderRiskLevel
	InternalNotes *string
	Phone         *string
	CompanyName   *string
	AddressLine1  *string
	AddressLine2  *string
	City          *string
	PostalCode    *string
	Country       *string
}

type UpdateAuctionBidderApprovalInput struct {
	RiskLevel          domain.AuctionBidderRiskLevel
	MaxBidLimitMinor   *int64
	DepositRequired    bool
	DepositAmountMinor *int64
	DepositStatus      domain.AuctionBidderDepositStatus
	InternalNotes      *string
}

type ListAuctionBiddersFilter struct {
	Status          *domain.ApprovalStatus
	RiskLevel       *domain.AuctionBidderRiskLevel
	DepositRequired *bool
	DepositStatus   *domain.AuctionBidderDepositStatus
}

type FeeRepository interface {
	CreateComponent(ctx context.Context, component *domain.FeeComponent) error
	UpdateComponent(ctx context.Context, component *domain.FeeComponent) error
	ListComponents(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]domain.FeeComponent, error)
	GetComponentByID(ctx context.Context, tenantID, componentID uuid.UUID) (*domain.FeeComponent, error)
	DeactivateComponent(ctx context.Context, tenantID, componentID uuid.UUID) error

	CreateAssignment(ctx context.Context, tenantID uuid.UUID, assignment *domain.FeeAssignment) error
	ListAssignments(ctx context.Context, tenantID uuid.UUID, scopeType domain.FeeScopeType, scopeID uuid.UUID) ([]domain.FeeAssignment, error)
	DeleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error
}

type TaxRepository interface {
	CreateRule(ctx context.Context, rule *domain.TaxRule) error
	UpdateRule(ctx context.Context, rule *domain.TaxRule) error
	ListRules(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]domain.TaxRule, error)
	GetRuleByID(ctx context.Context, tenantID, ruleID uuid.UUID) (*domain.TaxRule, error)
	DeactivateRule(ctx context.Context, tenantID, ruleID uuid.UUID) error

	CreateAssignment(ctx context.Context, tenantID uuid.UUID, assignment *domain.TaxRuleAssignment) error
	ListAssignments(ctx context.Context, tenantID uuid.UUID, scopeType domain.TaxRuleScopeType, scopeID uuid.UUID) ([]domain.TaxRuleAssignment, error)
	DeleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error
}

type InvoiceRepository interface {
	GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error)
	GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error)
	GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error)
	UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error
	GetAuctionSettlementSummary(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionSettlementSummary, bool, bool, error)
	ListAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error)
	MarkPaid(ctx context.Context, tenantID, invoiceID uuid.UUID, paymentReference *string, now time.Time) (*domain.Invoice, error)
	Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, reason string, now time.Time) (*domain.Invoice, error)
	StreamAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter StreamAuctionInvoicesFilter, fn func(domain.SettlementInvoiceRow) error) error
}

type ListAuctionInvoicesFilter struct {
	Status *domain.InvoicePaymentStatus
	Search string
	Limit  int
	Offset int
}

type StreamAuctionInvoicesFilter struct {
	Status *domain.InvoicePaymentStatus
	Search string
}

type PlaceLotBidTxInput struct {
	TenantID uuid.UUID
	LotID    uuid.UUID
	UserID   uuid.UUID
	Amount   *int64
	MaxBid   *int64
	Now      time.Time
}

type PlaceLotBidTxResult struct {
	Lot                   domain.Lot
	Bid                   domain.Bid
	AuctionTitle          string
	PreviousHighestBidder *uuid.UUID
	WasExtended           bool
	ExtensionEndTime      *time.Time
}

type LotSettlementResult struct {
	LotID        uuid.UUID
	AuctionID    uuid.UUID
	AuctionTitle string
	LotTitle     string
	WinnerUserID *uuid.UUID
	HammerPrice  *int64
}

type CreatePendingTenantWithAdminInput struct {
	TenantID               uuid.UUID
	TenantName             string
	Subdomain              string
	Country                string
	EstimatedAuctionVolume *int
	RequireAdminApproval   bool
	RequireKYC             bool
	AdminUserID            uuid.UUID
	AdminEmail             string
	AdminPasswordHash      string
	AdminEmailVerified     bool
	AdminApprovalStatus    domain.ApprovalStatus
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

type TokenService interface {
	Issue(userID, tenantID uuid.UUID, role domain.Role, locale domain.Locale) (string, error)
	Parse(token string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     domain.Role
	Locale   domain.Locale
	Exp      int64
}

type Broadcaster interface {
	BroadcastLotBid(ctx context.Context, lotID uuid.UUID, bid domain.Bid, currentPrice int64, endedAt *time.Time)
	BroadcastAuctionBid(ctx context.Context, auctionID uuid.UUID, lot domain.Lot, bid domain.Bid, currentPrice int64, endedAt *time.Time)
	BroadcastLotExtended(ctx context.Context, lotID uuid.UUID, newEndTime time.Time)
	BroadcastLotStatus(ctx context.Context, lotID uuid.UUID, status domain.LotStatus)
	BroadcastAuctionStatus(ctx context.Context, auctionID uuid.UUID, status domain.AuctionStatus)
}

type UserNotificationBroadcaster interface {
	BroadcastUserNotification(ctx context.Context, tenantID, userID uuid.UUID, notification domain.Notification)
}

type NotificationRepository interface {
	Insert(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType, title, body string, link *string, metadata json.RawMessage) (uuid.UUID, error)
	ListForUser(ctx context.Context, tenantID, userID uuid.UUID, limit int, cursorCreatedAt *time.Time, cursorID *uuid.UUID) ([]domain.Notification, error)
	MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error
}

type NotificationJobRepository interface {
	Enqueue(ctx context.Context, job domain.NotificationJob) (bool, error)
	EnqueueTx(ctx context.Context, tx pgx.Tx, job domain.NotificationJob) (bool, error)
	FetchBatchForProcessing(ctx context.Context, limit int) ([]domain.NotificationJob, error)
	MarkSending(ctx context.Context, tenantID, jobID uuid.UUID) error
	MarkSent(ctx context.Context, tenantID, jobID uuid.UUID) error
	MarkFailedOrRetry(ctx context.Context, tenantID, jobID uuid.UUID, attempts int, nextRunAt time.Time, lastError string, finalFailed bool) error
}

type NotificationPreferenceRepository interface {
	Get(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType) (domain.NotificationPreference, error)
	Upsert(ctx context.Context, tenantID, userID uuid.UUID, eventType domain.NotificationEventType, emailEnabled, inAppEnabled bool) error
	List(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.NotificationPreference, error)
}

type EmailOutboxRepository interface {
	Create(ctx context.Context, item *domain.EmailOutbox) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailOutbox, error)
	MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID, attempts int, nextAttemptAt *time.Time, lastError string) error
	ListRetryable(ctx context.Context, now time.Time, limit int) ([]domain.EmailOutbox, error)
	CountByRecipientSince(ctx context.Context, tenantID uuid.UUID, toEmail string, since time.Time) (int64, error)
	CountByTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error)
	ExistsRecentByScope(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, toEmail, kind string, entityType string, entityID *uuid.UUID, since time.Time) (bool, error)
}

type EmailMessage struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
	From     string
}

type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) error
}
