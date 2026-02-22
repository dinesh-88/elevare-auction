package httptransport

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/google/uuid"
)

const maxUploadSize = 10 << 20 // 10MB

var allowedImageTypes = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

func (h *Handler) UploadAuctionImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "auctions")
}

func (h *Handler) UploadLotImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, "lots")
}

func (h *Handler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant missing", http.StatusInternalServerError)
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	if kind != "logo" && kind != "banner" && kind != "image" {
		http.Error(w, "invalid kind", http.StatusBadRequest)
		return
	}
	fileURL, sizeBytes, mime, err := h.storeImageFile(w, r, "assets/"+kind)
	if err != nil {
		return
	}
	if kind != "logo" && mime == "image/svg+xml" {
		http.Error(w, "svg is allowed only for logo", http.StatusBadRequest)
		return
	}
	item, err := h.Asset.Register(r.Context(), usecase.RegisterAssetInput{
		TenantID:  tenant.ID,
		Kind:      kind,
		URL:       fileURL,
		SizeBytes: &sizeBytes,
		MIME:      mime,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request, scope string) {
	urlPath, _, _, err := h.storeImageFile(w, r, scope)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"url": urlPath})
}

func (h *Handler) storeImageFile(w http.ResponseWriter, r *http.Request, scope string) (string, int64, string, error) {
	tenant, ok := middleware.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant missing", http.StatusInternalServerError)
		return "", 0, "", errors.New("tenant missing")
	}
	if h.Storage == nil {
		http.Error(w, "storage not configured", http.StatusInternalServerError)
		return "", 0, "", errors.New("storage not configured")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "invalid multipart payload", http.StatusBadRequest)
		return "", 0, "", err
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return "", 0, "", err
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return "", 0, "", err
	}
	contentType := http.DetectContentType(header[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		http.Error(w, "unsupported file type", http.StatusBadRequest)
		return "", 0, "", errors.New("unsupported file type")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to process file", http.StatusBadRequest)
		return "", 0, "", err
	}

	filename := strings.ToLower(uuid.NewString()) + ext
	key := tenant.ID.String() + "/" + scope + "/" + filename
	urlPath, err := h.Storage.Put(r.Context(), key, contentType, file, fileHeader.Size)
	if err != nil {
		log.Printf("upload failed (scope=%s, key=%s): %v", scope, key, err)
		http.Error(w, "failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return "", 0, "", err
	}
	return urlPath, fileHeader.Size, contentType, nil
}
