package httptransport

import (
	"encoding/json"
	"net/http"

	errmiddleware "github.com/dinesh-88/elevare-auction/internal/http/middleware"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func WriteError(w http.ResponseWriter, err error) {
	errmiddleware.WriteError(w, err)
}

func writeErrorText(w http.ResponseWriter, message string, status int) {
	errmiddleware.WriteMessage(w, status, message)
}
