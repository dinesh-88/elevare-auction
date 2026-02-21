package httptransport

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/google/uuid"
)

const maxUploadSize = 8 << 20 // 8MB

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func (h *Handler) UploadAuctionImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "auctions")
}

func (h *Handler) UploadLotImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "lots")
}

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request, scope string) {
	tenant, ok := middleware.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant missing", http.StatusInternalServerError)
		return
	}
	if h.Storage == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "invalid multipart payload", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	contentType := http.DetectContentType(header[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		http.Error(w, "unsupported file type", http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to process file", http.StatusBadRequest)
		return
	}

	filename := strings.ToLower(uuid.NewString()) + ext
	key := tenant.ID.String() + "/" + scope + "/" + filename
	urlPath, err := h.Storage.Put(r.Context(), key, contentType, file, fileHeader.Size)
	if err != nil {
		log.Printf("upload failed (scope=%s, key=%s): %v", scope, key, err)
		http.Error(w, "failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"url": urlPath})
}
