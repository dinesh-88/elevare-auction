package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type DemoRequestUseCase struct {
	Requests port.DemoRequestRepository
}

type SubmitDemoRequestInput struct {
	Name        string
	Institution string
	Email       string
	Message     string
}

func (uc *DemoRequestUseCase) Submit(ctx context.Context, in SubmitDemoRequestInput) (*domain.DemoRequest, error) {
	name := strings.TrimSpace(in.Name)
	institution := strings.TrimSpace(in.Institution)
	email := strings.TrimSpace(in.Email)
	message := strings.TrimSpace(in.Message)

	if name == "" || institution == "" || email == "" || !strings.Contains(email, "@") {
		return nil, ErrValidation
	}
	if len(name) > 160 || len(institution) > 200 || len(email) > 255 {
		return nil, ErrValidation
	}
	if len(message) > 4000 {
		message = message[:4000]
	}

	var messagePtr *string
	if message != "" {
		messagePtr = &message
	}

	item := &domain.DemoRequest{
		ID:          uuid.New(),
		Name:        name,
		Institution: institution,
		Email:       email,
		Message:     messagePtr,
		CreatedAt:   time.Now().UTC(),
	}
	if err := uc.Requests.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}
