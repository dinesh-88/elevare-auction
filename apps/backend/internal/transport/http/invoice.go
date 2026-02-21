package httptransport

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
)

func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}
	inv, err := h.Invoice.GetAuthorized(r.Context(), tenant.ID, invoiceID, claims.UserID, claims.Role)
	if err != nil {
		if err == usecase.ErrForbidden {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (h *Handler) DownloadInvoicePDF(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "claims missing", http.StatusUnauthorized)
		return
	}
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}
	inv, err := h.Invoice.GetAuthorized(r.Context(), tenant.ID, invoiceID, claims.UserID, claims.Role)
	if err != nil {
		if err == usecase.ErrForbidden {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	pdfBytes, err := buildInvoicePDF(tenant.Name, inv)
	if err != nil {
		http.Error(w, "failed to generate pdf", http.StatusInternalServerError)
		return
	}
	filename := strings.ReplaceAll(inv.InvoiceNumber, " ", "_") + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func (h *Handler) UpdateInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status domain.InvoiceStatus `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.Invoice.UpdateStatus(r.Context(), tenant.ID, invoiceID, req.Status); err != nil {
		status := http.StatusBadRequest
		if err == usecase.ErrValidation {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (h *Handler) GenerateLotInvoice(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	lotID, err := uuid.Parse(chi.URLParam(r, "lotId"))
	if err != nil {
		http.Error(w, "invalid lot id", http.StatusBadRequest)
		return
	}
	inv, err := h.Invoice.GenerateForLot(r.Context(), tenant.ID, lotID, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

func buildInvoicePDF(tenantName string, inv *domain.Invoice) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	pdf.SetAutoPageBreak(true, 15)

	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 10, "Invoice", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, tenantName, "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(45, 7, "Invoice Number:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, inv.InvoiceNumber, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(45, 7, "Issue Date:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, inv.IssuedAt.Format(time.DateOnly), "", 1, "L", false, 0, "")
	if inv.DueDate != nil {
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(45, 7, "Due Date:", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(0, 7, inv.DueDate.Format(time.DateOnly), "", 1, "L", false, 0, "")
	}
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(45, 7, "Lot:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("%s - %s", inv.LotNumber, inv.LotTitle), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(45, 7, "Auction:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, inv.AuctionTitle, "", 1, "L", false, 0, "")
	pdf.Ln(3)

	snap, err := usecase.ParseInvoiceSnapshot(inv.Snapshot)
	if err != nil {
		return nil, err
	}
	currency := inv.Currency + " "
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(130, 8, "Line Item", "1", 0, "L", false, 0, "")
	pdf.CellFormat(0, 8, "Amount", "1", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(130, 8, "Hammer Price", "1", 0, "L", false, 0, "")
	pdf.CellFormat(0, 8, currency+formatInvoiceAmount(snap.HammerPrice), "1", 1, "R", false, 0, "")
	for _, fee := range snap.Fees {
		label := fee.Name
		if fee.Type == domain.FeeTypePercentage {
			label = fmt.Sprintf("%s (%.2f%%)", fee.Name, fee.Value)
		}
		pdf.CellFormat(130, 8, label, "1", 0, "L", false, 0, "")
		pdf.CellFormat(0, 8, currency+formatInvoiceAmount(fee.Amount), "1", 1, "R", false, 0, "")
	}
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(130, 9, "Total Due", "1", 0, "L", false, 0, "")
	pdf.CellFormat(0, 9, currency+formatInvoiceAmount(inv.TotalAmount), "1", 1, "R", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 6, "Payment instructions: Please contact the institution for settlement details.", "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatInvoiceAmount(v int64) string {
	s := fmt.Sprintf("%d", v)
	if len(s) <= 3 {
		return s
	}
	n := len(s) % 3
	if n == 0 {
		n = 3
	}
	out := s[:n]
	for i := n; i < len(s); i += 3 {
		out += "," + s[i:i+3]
	}
	return out
}
