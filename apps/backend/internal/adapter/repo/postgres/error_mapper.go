package postgres

import (
	"errors"
	"log"
	"net/http"

	apperrors "github.com/dinesh-88/elevare-auction/internal/errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapDBError(err error, conflictMessage string) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Printf(
			"DB ERROR code=%s constraint=%s detail=%s message=%s",
			pgErr.Code,
			pgErr.ConstraintName,
			pgErr.Detail,
			pgErr.Message,
		)
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			message := conflictMessage
			if message == "" {
				message = "Resource already exists."
			}
			return apperrors.New(apperrors.ErrConflict, message, http.StatusConflict)
		default:
			return apperrors.New(apperrors.ErrInternal, "An unexpected error occurred.", http.StatusInternalServerError)
		}
	}

	log.Printf("DB ERROR: %+v", err)
	return apperrors.New(apperrors.ErrInternal, "An unexpected error occurred.", http.StatusInternalServerError)
}
