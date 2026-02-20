package httptransport

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed swagger/openapi.yaml
var swaggerFS embed.FS

func ServeOpenAPISpec() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(swaggerFS, "swagger/openapi.yaml")
		if err != nil {
			http.Error(w, "openapi spec not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
