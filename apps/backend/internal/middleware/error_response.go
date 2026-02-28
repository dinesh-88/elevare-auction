package middleware

import (
	"net/http"

	errmiddleware "github.com/dinesh-88/elevare-auction/internal/http/middleware"
)

func writeError(w http.ResponseWriter, message string, status int) {
	errmiddleware.WriteMessage(w, status, message)
}
