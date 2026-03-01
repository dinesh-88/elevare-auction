package postgres

import (
	"testing"

	"github.com/dinesh-88/elevare-auction/internal/domain"
)

func TestCanMarkPaid(t *testing.T) {
	if err := canMarkPaid(domain.InvoicePaymentRefunded, domain.InvoiceSettlementCancelled); err == nil {
		t.Fatal("expected error for cancelled invoice")
	}
	if err := canMarkPaid(domain.InvoicePaymentUnpaid, domain.InvoiceSettlementApproved); err != nil {
		t.Fatalf("unexpected error for unpaid: %v", err)
	}
	if err := canMarkPaid(domain.InvoicePaymentPaid, domain.InvoiceSettlementInvoiceSent); err != nil {
		t.Fatalf("unexpected error for paid: %v", err)
	}
}

func TestCanCancel(t *testing.T) {
	if err := canCancel(domain.InvoicePaymentPaid, domain.InvoiceSettlementPaid); err == nil {
		t.Fatal("expected error for paid invoice")
	}
	if err := canCancel(domain.InvoicePaymentUnpaid, domain.InvoiceSettlementDraft); err != nil {
		t.Fatalf("unexpected error for unpaid: %v", err)
	}
	if err := canCancel(domain.InvoicePaymentRefunded, domain.InvoiceSettlementCancelled); err != nil {
		t.Fatalf("unexpected error for cancelled: %v", err)
	}
}
