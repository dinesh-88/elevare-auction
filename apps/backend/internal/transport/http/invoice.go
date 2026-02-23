package httptransport

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
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
		if err == usecase.ErrForbidden {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (h *Handler) MarkInvoicePaid(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}
	var req struct {
		PaymentReference *string `json:"payment_reference"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	inv, err := h.Invoice.MarkPaid(r.Context(), tenant.ID, invoiceID, req.PaymentReference, time.Now().UTC())
	if err != nil {
		status := http.StatusBadRequest
		if err == usecase.ErrForbidden {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (h *Handler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	inv, err := h.Invoice.Cancel(r.Context(), tenant.ID, invoiceID, req.Reason, time.Now().UTC())
	if err != nil {
		status := http.StatusBadRequest
		if err == usecase.ErrForbidden {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, inv)
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

func (h *Handler) GetAuctionSettlementSummary(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	summary, err := h.Invoice.GetAuctionSettlementSummary(r.Context(), tenant.ID, auctionID)
	if err != nil {
		status := http.StatusBadRequest
		if err == usecase.ErrForbidden {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) ListAuctionInvoices(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	filter, err := parseInvoiceListFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, total, err := h.Invoice.ListAuctionInvoices(r.Context(), tenant.ID, auctionID, filter)
	if err != nil {
		status := http.StatusBadRequest
		if err == usecase.ErrForbidden {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  rows,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *Handler) ExportAuctionSettlementCSV(w http.ResponseWriter, r *http.Request) {
	tenant, _ := middleware.TenantFromContext(r.Context())
	auctionID, err := uuid.Parse(chi.URLParam(r, "auctionId"))
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	statusFilter, err := parseInvoicePaymentStatus(strings.TrimSpace(r.URL.Query().Get("status")))
	if err != nil {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	filename := fmt.Sprintf("settlement-%s.csv", auctionID.String())
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	csvWriter := csv.NewWriter(w)
	header := []string{
		"invoice_number",
		"invoice_id",
		"bidder_name",
		"bidder_email",
		"lot_summary",
		"hammer",
		"fees",
		"tax",
		"total_due",
		"status",
		"issued_at",
		"paid_at",
		"payment_reference",
		"cancel_reason",
	}
	if err := csvWriter.Write(header); err != nil {
		http.Error(w, "failed writing csv", http.StatusInternalServerError)
		return
	}

	err = h.Invoice.StreamAuctionInvoices(r.Context(), tenant.ID, auctionID, usecase.InvoiceStreamFilter{
		Status: statusFilter,
		Search: search,
	}, func(row domain.SettlementInvoiceRow) error {
		paidAt := ""
		if row.PaidAt != nil {
			paidAt = row.PaidAt.UTC().Format(time.RFC3339)
		}
		paymentRef := ""
		if row.PaymentReference != nil {
			paymentRef = *row.PaymentReference
		}
		cancelReason := ""
		if row.CancelReason != nil {
			cancelReason = *row.CancelReason
		}
		record := []string{
			row.InvoiceNumber,
			row.ID.String(),
			row.BidderName,
			row.BidderEmail,
			row.LotSummary,
			formatMoneyFromMinor(row.Currency, row.HammerMinor),
			formatMoneyFromMinor(row.Currency, row.FeesMinor),
			formatMoneyFromMinor(row.Currency, row.TaxMinor),
			formatMoneyFromMinor(row.Currency, row.TotalDueMinor),
			string(row.PaymentStatus),
			row.IssuedAt.UTC().Format(time.RFC3339),
			paidAt,
			paymentRef,
			cancelReason,
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
		csvWriter.Flush()
		return csvWriter.Error()
	})
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == usecase.ErrForbidden {
			statusCode = http.StatusForbidden
		}
		http.Error(w, err.Error(), statusCode)
		return
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		http.Error(w, "failed writing csv", http.StatusInternalServerError)
		return
	}
	h.Invoice.AuditCSVExport(r.Context(), tenant.ID, auctionID)
}

func parseInvoiceListFilter(r *http.Request) (usecase.InvoiceListFilter, error) {
	limit := 50
	offset := 0
	var err error
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit <= 0 || limit > 200 {
			return usecase.InvoiceListFilter{}, fmt.Errorf("invalid limit")
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("offset")); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return usecase.InvoiceListFilter{}, fmt.Errorf("invalid offset")
		}
	}
	statusFilter, err := parseInvoicePaymentStatus(strings.TrimSpace(r.URL.Query().Get("status")))
	if err != nil {
		return usecase.InvoiceListFilter{}, err
	}
	return usecase.InvoiceListFilter{
		Status: statusFilter,
		Search: strings.TrimSpace(r.URL.Query().Get("search")),
		Limit:  limit,
		Offset: offset,
	}, nil
}

func parseInvoicePaymentStatus(raw string) (*domain.InvoicePaymentStatus, error) {
	if raw == "" {
		return nil, nil
	}
	status := domain.InvoicePaymentStatus(raw)
	switch status {
	case domain.InvoicePaymentUnpaid, domain.InvoicePaymentPaid, domain.InvoicePaymentCancelled:
		return &status, nil
	default:
		return nil, fmt.Errorf("invalid payment status")
	}
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
	pdf.CellFormat(0, 9, currency+formatInvoiceAmount(inv.TotalDueMinor), "1", 1, "R", false, 0, "")
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

func formatMoneyFromMinor(currency string, minor int64) string {
	decimals := currencyDecimals(currency)
	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}
	if decimals == 0 {
		return sign + strconv.FormatInt(minor, 10)
	}
	factor := int64(1)
	for i := 0; i < decimals; i++ {
		factor *= 10
	}
	major := minor / factor
	fraction := minor % factor
	return fmt.Sprintf("%s%d.%0*d", sign, major, decimals, fraction)
}

func currencyDecimals(currency string) int {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "JPY", "KRW", "VND":
		return 0
	default:
		return 2
	}
}
