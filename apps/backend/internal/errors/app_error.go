package errors

type AppError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Debug   string `json:"debug,omitempty"`
}

func New(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

const (
	ErrInternal     = "internal_error"
	ErrValidation   = "validation_error"
	ErrUnauthorized = "unauthorized"
	ErrForbidden    = "forbidden"
	ErrNotFound     = "not_found"
	ErrConflict     = "conflict"
	ErrRateLimited  = "rate_limited"
)
