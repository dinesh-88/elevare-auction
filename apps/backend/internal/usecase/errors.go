package usecase

import "errors"

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrValidation        = errors.New("validation failed")
	ErrNotFound          = errors.New("not found")
	ErrApprovalRequired  = errors.New("approval required")
	ErrPendingApproval   = errors.New("pending approval")
	ErrNotApproved       = errors.New("not approved")
	ErrBidderSuspended   = errors.New("you are not allowed to bid")
	ErrSimulatorBlocked  = errors.New("simulator bidder is restricted to the demo tenant")
)
