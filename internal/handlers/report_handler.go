package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cold-backend/internal/services"
	"cold-backend/internal/timeutil"

	"github.com/gorilla/mux"
)

type ReportHandler struct {
	Service *services.ReportService
}

func NewReportHandler(service *services.ReportService) *ReportHandler {
	return &ReportHandler{Service: service}
}

// GetCustomersCSV handles GET /api/reports/customers/csv
// Query params: filter=all|outstanding|paid
func (h *ReportHandler) GetCustomersCSV(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	csvData, err := h.Service.GenerateCustomersCSV(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate CSV: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("customers_%s_%s.csv", filter, timeutil.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(csvData)
}

// GetCustomersPDFZip handles GET /api/reports/customers/pdf
// Query params: filter=all|outstanding|paid
// Returns a ZIP file containing individual PDFs for each customer
func (h *ReportHandler) GetCustomersPDFZip(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// Generate all PDFs in parallel
	pdfs, err := h.Service.GenerateBulkCustomerPDFs(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate PDFs: %v", err), http.StatusInternalServerError)
		return
	}

	if len(pdfs) == 0 {
		http.Error(w, "No customers found matching the filter", http.StatusNotFound)
		return
	}

	// Create ZIP
	zipData, err := h.Service.CreateBulkPDFZip(pdfs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create ZIP: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("customer_reports_%s_%s.zip", filter, timeutil.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(zipData)
}

// GetSingleCustomerPDF handles GET /api/reports/customers/{phone}/pdf
// Returns a single customer's PDF report
func (h *ReportHandler) GetSingleCustomerPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phone := vars["phone"]

	if phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	rentRate, _ := h.Service.GetRentRate(ctx)
	data, err := h.Service.GetCustomerReportData(ctx, phone, rentRate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Customer not found: %v", err), http.StatusNotFound)
		return
	}

	pdfData, err := h.Service.GenerateCustomerPDF(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate PDF: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("customer_%s_%s_%s.pdf", phone, data.Customer.Name, timeutil.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(pdfData)
}

// GetDailySummaryCSV handles GET /api/reports/daily-summary/csv
// Query params: date=YYYY-MM-DD (defaults to today)
func (h *ReportHandler) GetDailySummaryCSV(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr == "" {
		date = timeutil.StartOfDay(timeutil.Now())
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	csvData, err := h.Service.GenerateDailySummaryCSV(ctx, date)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate CSV: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("daily_summary_%s.csv", date.Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(csvData)
}

// GetDailySummaryPDF handles GET /api/reports/daily-summary/pdf
// Query params: date=YYYY-MM-DD (defaults to today)
func (h *ReportHandler) GetDailySummaryPDF(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr == "" {
		date = timeutil.StartOfDay(timeutil.Now())
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	data, err := h.Service.GetDailySummaryData(ctx, date)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get data: %v", err), http.StatusInternalServerError)
		return
	}

	pdfData, err := h.Service.GenerateDailySummaryPDF(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate PDF: %v", err), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("daily_summary_%s.pdf", date.Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(pdfData)
}

// GetReportStats handles GET /api/reports/stats
// Returns a summary of report data without generating files
func (h *ReportHandler) GetReportStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	customers, err := h.Service.GetAllCustomerReportData(ctx, "all")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get data: %v", err), http.StatusInternalServerError)
		return
	}

	var totalQty, totalPaid int
	var totalRent, totalBalance float64
	var outstanding, paid int

	for _, c := range customers {
		totalQty += c.TotalQty
		totalRent += c.TotalRent
		totalPaid += int(c.TotalPaid)
		totalBalance += c.Balance
		if c.Balance > 0 {
			outstanding++
		} else {
			paid++
		}
	}

	stats := map[string]interface{}{
		"total_customers":       len(customers),
		"outstanding_customers": outstanding,
		"paid_customers":        paid,
		"total_quantity":        totalQty,
		"total_rent":            totalRent,
		"total_paid":            totalPaid,
		"total_balance":         totalBalance,
		"generated_at":          timeutil.Now().Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
