package emailtpl

import (
	"strings"
	"testing"

	"github.com/dinesh-88/elevare-auction/internal/domain"
)

func TestTemplateRendererRendersEnglishTemplate(t *testing.T) {
	r := NewTemplateRenderer()
	out, err := r.Render("invoice_ready", domain.LocaleEnglish, BrandContext{TenantName: "Elevare"}, map[string]any{
		"invoice_number": "INV-001",
		"total_due":      int64(1200),
		"link":           "https://example.com/invoices/1",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if out.Subject == "" || !strings.Contains(out.HTML, "INV-001") || !strings.Contains(out.Text, "INV-001") {
		t.Fatalf("unexpected render output: %+v", out)
	}
}

func TestTemplateRendererUsesLocaleWhenAvailable(t *testing.T) {
	r := NewTemplateRenderer()
	out, err := r.Render("outbid", domain.LocaleFrench, BrandContext{TenantName: "Elevare"}, map[string]any{
		"lot_title": "Lot X",
		"amount":    int64(2000),
		"link":      "https://example.com/lots/x",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out.Subject, "surench") {
		t.Fatalf("expected french subject, got %q", out.Subject)
	}
}

func TestTemplateRendererFallsBackToEnglishWhenLocaleTemplateMissing(t *testing.T) {
	r := NewTemplateRenderer()
	out, err := r.Render("invoice_ready", domain.LocaleFrench, BrandContext{TenantName: "Elevare"}, map[string]any{
		"invoice_number": "INV-001",
		"total_due":      int64(1200),
		"link":           "https://example.com/invoices/1",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out.HTML, "View invoice") {
		t.Fatalf("expected english fallback html, got %q", out.HTML)
	}
}

func TestTemplateRendererEscapesHTML(t *testing.T) {
	r := NewTemplateRenderer()
	out, err := r.Render("outbid", domain.LocaleEnglish, BrandContext{TenantName: "Elevare"}, map[string]any{
		"lot_title": "<script>alert(1)</script>",
		"amount":    int64(2000),
		"link":      "https://example.com/lots/x",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(out.HTML, "<script>alert(1)</script>") {
		t.Fatalf("expected escaped html, got %q", out.HTML)
	}
	if !strings.Contains(out.HTML, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped content, got %q", out.HTML)
	}
}
