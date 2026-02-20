package port

import (
	"context"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
)

type TenantRepository interface {
	GetByHost(ctx context.Context, host string) (*domain.Tenant, error)
	CreatePendingTenantWithAdmin(ctx context.Context, in CreatePendingTenantWithAdminInput) error
	ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error)
	UpdateStatusAndAdminApproval(ctx context.Context, tenantID uuid.UUID, tenantStatus domain.TenantStatus, adminApproval domain.ApprovalStatus) error
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
	PlaceBidTx(ctx context.Context, in PlaceBidTxInput) (*PlaceBidTxResult, error)
	RunSchedulerTick(ctx context.Context, now time.Time) error
}

type PlaceBidTxInput struct {
	TenantID  uuid.UUID
	AuctionID uuid.UUID
	UserID    uuid.UUID
	Amount    int64
	Now       time.Time
}

type PlaceBidTxResult struct {
	Auction domain.Auction
	Bid     domain.Bid
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
	BroadcastBid(ctx context.Context, auctionID uuid.UUID, bid domain.Bid, currentPrice int64, endTime time.Time)
}
