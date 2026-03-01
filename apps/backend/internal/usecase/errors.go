package usecase

import "errors"

var (
	ErrUnauthorized             = errors.New("unauthorized")
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidCredential        = errors.New("invalid credentials")
	ErrValidation               = errors.New("validation failed")
	ErrNotFound                 = errors.New("not found")
	ErrApprovalRequired         = errors.New("approval required")
	ErrRegistrationRequired     = errors.New("registration required")
	ErrPendingApproval          = errors.New("pending approval")
	ErrNotApproved              = errors.New("not approved")
	ErrBidderSuspended          = errors.New("you are not allowed to bid")
	ErrBidderOnboardingRequired = errors.New("bidder onboarding required")
	ErrSimulatorBlocked         = errors.New("simulator bidder is restricted to the demo tenant")
	ErrBidLimitExceeded         = errors.New("Your bid exceeds your authorised limit.")
	ErrDepositNeeded            = errors.New("A deposit is required before bidding.")
	ErrTermsUnavailable         = errors.New("active auction terms are not available")
	ErrTermsRequired            = errors.New("auction terms must be accepted before bidding")
	ErrOnboardingPending        = errors.New("institution onboarding must be completed before creating or publishing auctions")
)
