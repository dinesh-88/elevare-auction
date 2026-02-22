package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AssetUseCase struct {
	Assets port.AssetRepository
}

type RegisterAssetInput struct {
	TenantID   uuid.UUID
	Kind       string
	URL        string
	Width      *int
	Height     *int
	SizeBytes  *int64
	MIME       string
	UploadedAt time.Time
}

func (uc *AssetUseCase) Register(ctx context.Context, in RegisterAssetInput) (*domain.TenantAsset, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	switch kind {
	case "logo", "banner", "image":
	default:
		return nil, ErrValidation
	}
	if strings.TrimSpace(in.URL) == "" {
		return nil, ErrValidation
	}
	if in.SizeBytes != nil && *in.SizeBytes > 10*1024*1024 {
		return nil, ErrValidation
	}
	item := &domain.TenantAsset{
		ID:        uuid.New(),
		TenantID:  in.TenantID,
		Kind:      kind,
		URL:       strings.TrimSpace(in.URL),
		Width:     in.Width,
		Height:    in.Height,
		SizeBytes: in.SizeBytes,
		MIME:      strings.TrimSpace(in.MIME),
	}
	if err := uc.Assets.Create(ctx, in.TenantID, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *AssetUseCase) List(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TenantAsset, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != "" && kind != "logo" && kind != "banner" && kind != "image" {
		return nil, ErrValidation
	}
	return uc.Assets.ListByKind(ctx, tenantID, kind, limit)
}
