package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type InvoiceUseCase struct {
	Invoices port.InvoiceRepository
	Audit    port.AuditRepository
}

func (uc *InvoiceUseCase) GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return uc.Invoices.GenerateForLot(ctx, tenantID, lotID, now)
}

func (uc *InvoiceUseCase) GetAuthorized(ctx context.Context, tenantID, invoiceID, userID uuid.UUID, role domain.Role) (*domain.Invoice, error) {
	inv, err := uc.Invoices.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	if role == domain.RoleAdmin || inv.WinnerUserID == userID {
		return inv, nil
	}
	return nil, ErrForbidden
}

func (uc *InvoiceUseCase) GetByLotAuthorized(ctx context.Context, tenantID, lotID, userID uuid.UUID, role domain.Role) (*domain.Invoice, error) {
	inv, err := uc.Invoices.GetByLotID(ctx, tenantID, lotID)
	if err != nil {
		return nil, err
	}
	if role == domain.RoleAdmin || inv.WinnerUserID == userID {
		return inv, nil
	}
	return nil, ErrForbidden
}

func (uc *InvoiceUseCase) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	if status != domain.InvoicePaid && status != domain.InvoiceCancelled && status != domain.InvoiceIssued {
		return ErrValidation
	}
	if err := uc.Invoices.UpdateStatus(ctx, tenantID, invoiceID, status); err != nil {
		return err
	}
	if uc.Audit != nil {
		action := "INVOICE_STATUS_UPDATED"
		if status == domain.InvoicePaid {
			action = "INVOICE_MARKED_PAID"
		} else if status == domain.InvoiceCancelled {
			action = "INVOICE_CANCELLED"
		}
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"status": status})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      action,
			EntityType:  "INVOICE",
			EntityID:    &invoiceID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return nil
}

func ParseInvoiceSnapshot(raw json.RawMessage) (domain.LotFinancialSnapshot, error) {
	var snap domain.LotFinancialSnapshot
	if len(raw) == 0 {
		return snap, nil
	}
	err := json.Unmarshal(raw, &snap)
	return snap, err
}
