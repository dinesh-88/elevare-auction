package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCalculateFinancialsSeparatesFeesAndTax(t *testing.T) {
	now := time.Now().UTC()
	computation := calculateFinancials(10000, []rawFeeRule{
		{
			ID:        uuid.New(),
			Name:      "Buyer Premium",
			Type:      "PERCENTAGE",
			Value:     10,
			AppliesTo: "HAMMER",
			Taxable:   true,
			CreatedAt: now,
		},
		{
			ID:        uuid.New(),
			Name:      "Handling",
			Type:      "FIXED",
			Value:     250,
			AppliesTo: "SUBTOTAL",
			Taxable:   false,
			CreatedAt: now.Add(time.Second),
		},
	}, []rawTaxRule{
		{
			ID:          uuid.New(),
			Name:        "Sales Tax",
			RatePercent: 8,
			CreatedAt:   now,
		},
	})

	if computation.FeeSubtotal != 1250 {
		t.Fatalf("expected fee subtotal 1250, got %d", computation.FeeSubtotal)
	}
	if computation.TaxableBase != 1000 {
		t.Fatalf("expected taxable base 1000, got %d", computation.TaxableBase)
	}
	if computation.TaxSubtotal != 80 {
		t.Fatalf("expected tax subtotal 80, got %d", computation.TaxSubtotal)
	}
	if computation.TotalDue != 11330 {
		t.Fatalf("expected total due 11330, got %d", computation.TotalDue)
	}
	if len(computation.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(computation.Lines))
	}
	if computation.Lines[0].Category != "FEE" || computation.Lines[1].Category != "FEE" || computation.Lines[2].Category != "TAX" {
		t.Fatalf("unexpected categories: %#v", computation.Lines)
	}
}

func TestSnapshotTotalsUsesCategoryAndLegacyFallback(t *testing.T) {
	payload, err := json.Marshal(map[string]any{
		"hammer_price": 10000,
		"total_due":    11330,
		"fees": []map[string]any{
			{"name": "Buyer Premium", "amount": 1000, "taxable": true, "category": "FEE"},
			{"name": "Handling", "amount": 250, "category": "FEE"},
			{"name": "Sales Tax", "amount": 80, "category": "TAX"},
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	hammer, fees, tax, total := snapshotTotals(payload, 0)
	if hammer != 10000 || fees != 1250 || tax != 80 || total != 11330 {
		t.Fatalf("unexpected totals: hammer=%d fees=%d tax=%d total=%d", hammer, fees, tax, total)
	}

	legacy, err := json.Marshal(map[string]any{
		"hammer_price": 10000,
		"total_due":    10800,
		"fees": []map[string]any{
			{"name": "Buyer Premium", "amount": 1000, "taxable": true},
			{"name": "VAT", "amount": 800},
		},
	})
	if err != nil {
		t.Fatalf("marshal legacy payload: %v", err)
	}

	_, legacyFees, legacyTax, _ := snapshotTotals(legacy, 0)
	if legacyFees != 1000 || legacyTax != 800 {
		t.Fatalf("unexpected legacy totals: fees=%d tax=%d", legacyFees, legacyTax)
	}
}
