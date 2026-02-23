package postgres

import (
	"testing"

	"github.com/dinesh-88/elevare-auction/internal/domain"
)

func TestCanMarkPaid(t *testing.T) {
	if err := canMarkPaid(domain.InvoicePaymentCancelled); err == nil {
		t.Fatal("expected error for cancelled invoice")
	}
	if err := canMarkPaid(domain.InvoicePaymentUnpaid); err != nil {
		t.Fatalf("unexpected error for unpaid: %v", err)
	}
	if err := canMarkPaid(domain.InvoicePaymentPaid); err != nil {
		t.Fatalf("unexpected error for paid: %v", err)
	}
}

func TestCanCancel(t *testing.T) {
	if err := canCancel(domain.InvoicePaymentPaid); err == nil {
		t.Fatal("expected error for paid invoice")
	}
	if err := canCancel(domain.InvoicePaymentUnpaid); err != nil {
		t.Fatalf("unexpected error for unpaid: %v", err)
	}
	if err := canCancel(domain.InvoicePaymentCancelled); err != nil {
		t.Fatalf("unexpected error for cancelled: %v", err)
	}
}
