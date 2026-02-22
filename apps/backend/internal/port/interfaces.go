package port

import (
	"context"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
)

type TenantRepository interface {
	GetByHost(ctx context.Context, host string) (*domain.Tenant, error)
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	CreatePendingTenantWithAdmin(ctx context.Context, in CreatePendingTenantWithAdminInput) error
	ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error)
	UpdateStatusAndAdminApproval(ctx context.Context, tenantID uuid.UUID, tenantStatus domain.TenantStatus, adminApproval domain.ApprovalStatus) error
	UpdateDefaultCurrency(ctx context.Context, tenantID uuid.UUID, currency string) error
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error)
	GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error)
	UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error
	UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error
}

type AuctionRepository interface {
	Create(ctx context.Context, auction *domain.Auction) error
	ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error)
	ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error)
	GetByID(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error)
	Update(ctx context.Context, auction *domain.Auction) error
	Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error
	UpdateStatus(ctx context.Context, tenantID, auctionID uuid.UUID, status domain.AuctionStatus) error
	RunSchedulerTick(ctx context.Context, now time.Time) error
}

type LotRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error
	Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error
	Delete(ctx context.Context, tenantID, lotID uuid.UUID) error
	Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error
	ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error)
	GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error)
	Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error
	End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error
	OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error
	PlaceBidTx(ctx context.Context, in PlaceLotBidTxInput) (*PlaceLotBidTxResult, error)
	ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error)
	ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error)
	RunSchedulerTick(ctx context.Context, now time.Time) error
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

type InvoiceRepository interface {
	GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error)
	GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error)
	GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error)
	UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error
}

type PlaceLotBidTxInput struct {
	TenantID uuid.UUID
	LotID    uuid.UUID
	UserID   uuid.UUID
	Amount   int64
	Now      time.Time
}

type PlaceLotBidTxResult struct {
	Lot domain.Lot
	Bid domain.Bid
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
	Issue(userID, tenantID uuid.UUID, role domain.Role) (string, error)
	Parse(token string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     domain.Role
	Exp      int64
}

type Broadcaster interface {
	BroadcastLotBid(ctx context.Context, lotID uuid.UUID, bid domain.Bid, currentPrice int64, endedAt *time.Time)
	BroadcastLotStatus(ctx context.Context, lotID uuid.UUID, status domain.LotStatus)
	BroadcastAuctionStatus(ctx context.Context, auctionID uuid.UUID, status domain.AuctionStatus)
}
