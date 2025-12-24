package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"sync"
	"time"

	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
	"cold-backend/internal/timeutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jung-kurt/gofpdf/v2"
)

// CustomerReportData holds all data for a customer report
type CustomerReportData struct {
	Customer    *models.Customer
	Entries     []*models.Entry
	RoomEntries []*models.RoomEntry
	Payments    []*models.RentPayment
	TotalQty    int
	TotalRent   float64
	TotalPaid   float64
	Balance     float64
	ThockCount  int
}

// DailySummaryData holds data for daily summary report
type DailySummaryData struct {
	Date        time.Time
	Entries     []*models.Entry
	TotalSeed   int
	TotalSell   int
	SeedQty     int
	SellQty     int
	TotalQty    int
	TotalTrucks int
}

// ReportService handles report generation
type ReportService struct {
	DB              *pgxpool.Pool
	CustomerRepo    *repositories.CustomerRepository
	EntryRepo       *repositories.EntryRepository
	RoomEntryRepo   *repositories.RoomEntryRepository
	RentPaymentRepo *repositories.RentPaymentRepository
	SettingsRepo    *repositories.SystemSettingRepository
}

// NewReportService creates a new report service
func NewReportService(
	db *pgxpool.Pool,
	customerRepo *repositories.CustomerRepository,
	entryRepo *repositories.EntryRepository,
	roomEntryRepo *repositories.RoomEntryRepository,
	rentPaymentRepo *repositories.RentPaymentRepository,
	settingsRepo *repositories.SystemSettingRepository,
) *ReportService {
	return &ReportService{
		DB:              db,
		CustomerRepo:    customerRepo,
		EntryRepo:       entryRepo,
		RoomEntryRepo:   roomEntryRepo,
		RentPaymentRepo: rentPaymentRepo,
		SettingsRepo:    settingsRepo,
	}
}

// GetRentRate fetches the current rent rate from settings
func (s *ReportService) GetRentRate(ctx context.Context) (float64, error) {
	setting, err := s.SettingsRepo.Get(ctx, "rent_rate_per_bag")
	if err != nil {
		return 2.50, nil // Default rate
	}
	rate, err := strconv.ParseFloat(setting.SettingValue, 64)
	if err != nil {
		return 2.50, nil
	}
	return rate, nil
}

// GetCustomerReportData fetches all data for a customer
func (s *ReportService) GetCustomerReportData(ctx context.Context, phone string, rentRate float64) (*CustomerReportData, error) {
	customer, err := s.CustomerRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Get entries for this customer
	entries, err := s.EntryRepo.ListByCustomer(ctx, customer.ID)
	if err != nil {
		entries = []*models.Entry{}
	}

	// Get all room entries and payments
	var roomEntries []*models.RoomEntry
	var payments []*models.RentPayment
	var totalQty int
	var totalPaid float64
	thockSet := make(map[string]bool)

	for _, entry := range entries {
		// Get room entries for this entry
		reList, err := s.RoomEntryRepo.ListByThockNumber(ctx, entry.ThockNumber)
		if err == nil {
			roomEntries = append(roomEntries, reList...)
			for _, re := range reList {
				totalQty += re.Quantity
			}
		}
		thockSet[entry.ThockNumber] = true

		// Get payments for this customer phone
		pList, err := s.RentPaymentRepo.GetByPhone(ctx, phone)
		if err == nil && payments == nil {
			payments = pList
			for _, p := range pList {
				totalPaid += p.AmountPaid
			}
		}
	}

	totalRent := float64(totalQty) * rentRate
	balance := totalRent - totalPaid

	return &CustomerReportData{
		Customer:    customer,
		Entries:     entries,
		RoomEntries: roomEntries,
		Payments:    payments,
		TotalQty:    totalQty,
		TotalRent:   totalRent,
		TotalPaid:   totalPaid,
		Balance:     balance,
		ThockCount:  len(thockSet),
	}, nil
}

// GetAllCustomerReportData fetches data for all customers
func (s *ReportService) GetAllCustomerReportData(ctx context.Context, filter string) ([]*CustomerReportData, error) {
	rentRate, _ := s.GetRentRate(ctx)

	customers, err := s.CustomerRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Use parallel fetching with worker pool
	type result struct {
		index int
		data  *CustomerReportData
		err   error
	}

	results := make(chan result, len(customers))
	jobs := make(chan struct {
		index    int
		customer *models.Customer
	}, len(customers))

	// Start 10 workers
	var wg sync.WaitGroup
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				data, err := s.GetCustomerReportData(ctx, job.customer.Phone, rentRate)
				results <- result{index: job.index, data: data, err: err}
			}
		}()
	}

	// Send jobs
	for i, c := range customers {
		jobs <- struct {
			index    int
			customer *models.Customer
		}{index: i, customer: c}
	}
	close(jobs)

	// Wait and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	reportData := make([]*CustomerReportData, len(customers))
	for r := range results {
		if r.err == nil {
			reportData[r.index] = r.data
		}
	}

	// Filter out nil entries and apply filter
	var filtered []*CustomerReportData
	for _, data := range reportData {
		if data == nil {
			continue
		}
		switch filter {
		case "outstanding":
			if data.Balance > 0 {
				filtered = append(filtered, data)
			}
		case "paid":
			if data.Balance <= 0 {
				filtered = append(filtered, data)
			}
		default:
			filtered = append(filtered, data)
		}
	}

	return filtered, nil
}

// GenerateCustomerPDF generates a PDF for a single customer
func (s *ReportService) GenerateCustomerPDF(data *CustomerReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "Cold Storage - Customer Report", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(190, 6, fmt.Sprintf("Generated: %s", timeutil.Now().Format("02-Jan-2006 03:04 PM")), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Customer Info Box
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "Customer Information", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(95, 7, fmt.Sprintf("Name: %s", data.Customer.Name), "LB", 0, "L", false, 0, "")
	pdf.CellFormat(95, 7, fmt.Sprintf("Phone: %s", data.Customer.Phone), "RB", 1, "L", false, 0, "")
	pdf.CellFormat(95, 7, fmt.Sprintf("Village: %s", data.Customer.Village), "LB", 0, "L", false, 0, "")
	if data.Customer.SO != "" {
		pdf.CellFormat(95, 7, fmt.Sprintf("S/O: %s", data.Customer.SO), "RB", 1, "L", false, 0, "")
	} else {
		pdf.CellFormat(95, 7, "", "RB", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	// Storage Summary
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "Storage Details", "1", 1, "L", true, 0, "")

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(40, 7, "Thock No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Room", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Floor", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Gate", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, "Variety", "1", 1, "C", true, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 10)
	for _, re := range data.RoomEntries {
		pdf.CellFormat(40, 6, re.ThockNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, re.RoomNo, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, re.Floor, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, re.GateNo, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", re.Quantity), "1", 0, "C", false, 0, "")
		variety := re.Variety
		if len(variety) > 20 {
			variety = variety[:17] + "..."
		}
		pdf.CellFormat(50, 6, variety, "1", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	// Financial Summary
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "Financial Summary", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(63, 8, fmt.Sprintf("Total Quantity: %d bags", data.TotalQty), "1", 0, "C", false, 0, "")
	pdf.CellFormat(63, 8, fmt.Sprintf("Total Rent: Rs. %.2f", data.TotalRent), "1", 0, "C", false, 0, "")
	pdf.CellFormat(64, 8, fmt.Sprintf("Total Paid: Rs. %.2f", data.TotalPaid), "1", 1, "C", false, 0, "")

	// Balance - highlight if outstanding
	if data.Balance > 0 {
		pdf.SetFillColor(255, 200, 200) // Light red for outstanding
	} else {
		pdf.SetFillColor(200, 255, 200) // Light green for paid
	}
	pdf.SetFont("Arial", "B", 14)
	balanceText := fmt.Sprintf("Balance Due: Rs. %.2f", data.Balance)
	if data.Balance <= 0 {
		balanceText = "FULLY PAID"
	}
	pdf.CellFormat(190, 10, balanceText, "1", 1, "C", true, 0, "")

	// Payment History if any
	if len(data.Payments) > 0 {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(190, 8, "Payment History", "1", 1, "L", true, 0, "")

		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(200, 200, 200)
		pdf.CellFormat(40, 7, "Receipt #", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(55, 7, "Amount", "1", 0, "C", true, 0, "")
		pdf.CellFormat(55, 7, "Notes", "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 10)
		for _, p := range data.Payments {
			pdf.CellFormat(40, 6, p.ReceiptNumber, "1", 0, "C", false, 0, "")
			pdf.CellFormat(40, 6, p.PaymentDate.Format("02-Jan-2006"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(55, 6, fmt.Sprintf("Rs. %.2f", p.AmountPaid), "1", 0, "R", false, 0, "")
			notes := p.Notes
			if len(notes) > 25 {
				notes = notes[:22] + "..."
			}
			pdf.CellFormat(55, 6, notes, "1", 1, "L", false, 0, "")
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateBulkCustomerPDFs generates PDFs for all customers in parallel
func (s *ReportService) GenerateBulkCustomerPDFs(ctx context.Context, filter string) (map[string][]byte, error) {
	customers, err := s.GetAllCustomerReportData(ctx, filter)
	if err != nil {
		return nil, err
	}

	type pdfResult struct {
		phone string
		name  string
		data  []byte
		err   error
	}

	results := make(chan pdfResult, len(customers))
	jobs := make(chan *CustomerReportData, len(customers))

	// Start 5 workers for PDF generation
	var wg sync.WaitGroup
	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				pdfData, err := s.GenerateCustomerPDF(c)
				results <- pdfResult{
					phone: c.Customer.Phone,
					name:  c.Customer.Name,
					data:  pdfData,
					err:   err,
				}
			}
		}()
	}

	// Send jobs
	for _, c := range customers {
		jobs <- c
	}
	close(jobs)

	// Wait and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect PDFs
	pdfs := make(map[string][]byte)
	for r := range results {
		if r.err == nil && r.data != nil {
			// Use phone_name as filename
			filename := fmt.Sprintf("%s_%s", r.phone, r.name)
			pdfs[filename] = r.data
		}
	}

	return pdfs, nil
}

// CreateBulkPDFZip creates a ZIP file containing all customer PDFs
func (s *ReportService) CreateBulkPDFZip(pdfs map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for filename, pdfData := range pdfs {
		// Clean filename
		cleanName := fmt.Sprintf("customer_%s.pdf", filename)
		fw, err := zw.Create(cleanName)
		if err != nil {
			continue
		}
		fw.Write(pdfData)
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateCustomersCSV generates a CSV file with all customer data
func (s *ReportService) GenerateCustomersCSV(ctx context.Context, filter string) ([]byte, error) {
	customers, err := s.GetAllCustomerReportData(ctx, filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	w.Write([]string{
		"#", "Name", "Phone", "Village", "S/O",
		"Thocks", "Total Qty", "Total Rent", "Total Paid", "Balance", "Status",
	})

	// Data rows
	for i, c := range customers {
		status := "PAID"
		if c.Balance > 0 {
			status = "DUE"
		}
		w.Write([]string{
			fmt.Sprintf("%d", i+1),
			c.Customer.Name,
			c.Customer.Phone,
			c.Customer.Village,
			c.Customer.SO,
			fmt.Sprintf("%d", c.ThockCount),
			fmt.Sprintf("%d", c.TotalQty),
			fmt.Sprintf("%.2f", c.TotalRent),
			fmt.Sprintf("%.2f", c.TotalPaid),
			fmt.Sprintf("%.2f", c.Balance),
			status,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

// GetDailySummaryData fetches data for daily summary report
func (s *ReportService) GetDailySummaryData(ctx context.Context, date time.Time) (*DailySummaryData, error) {
	// Get all entries for the date
	allEntries, err := s.EntryRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	startOfDay := timeutil.StartOfDay(date)
	endOfDay := timeutil.EndOfDay(date)

	var dateEntries []*models.Entry
	var seedCount, sellCount, seedQty, sellQty int

	for _, e := range allEntries {
		entryTime := timeutil.ToIST(e.CreatedAt)
		if entryTime.After(startOfDay) && entryTime.Before(endOfDay) {
			dateEntries = append(dateEntries, e)
			if e.ThockCategory == "seed" {
				seedCount++
				seedQty += e.ExpectedQuantity
			} else {
				sellCount++
				sellQty += e.ExpectedQuantity
			}
		}
	}

	return &DailySummaryData{
		Date:        date,
		Entries:     dateEntries,
		TotalSeed:   seedCount,
		TotalSell:   sellCount,
		SeedQty:     seedQty,
		SellQty:     sellQty,
		TotalQty:    seedQty + sellQty,
		TotalTrucks: seedCount + sellCount,
	}, nil
}

// GenerateDailySummaryPDF generates a PDF for daily summary
func (s *ReportService) GenerateDailySummaryPDF(data *DailySummaryData) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for more columns
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(277, 12, "Cold Storage - Daily Summary Report", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(277, 8, fmt.Sprintf("Date: %s", data.Date.Format("02-Jan-2006 (Monday)")), "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(277, 6, fmt.Sprintf("Generated: %s", timeutil.Now().Format("02-Jan-2006 03:04 PM")), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Summary Box
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(277, 8, "Summary", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(69, 8, fmt.Sprintf("Total Trucks: %d", data.TotalTrucks), "1", 0, "C", false, 0, "")
	pdf.CellFormat(69, 8, fmt.Sprintf("Seed: %d trucks (%d bags)", data.TotalSeed, data.SeedQty), "1", 0, "C", false, 0, "")
	pdf.CellFormat(69, 8, fmt.Sprintf("Sell: %d trucks (%d bags)", data.TotalSell, data.SellQty), "1", 0, "C", false, 0, "")
	pdf.CellFormat(70, 8, fmt.Sprintf("Total Quantity: %d bags", data.TotalQty), "1", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Entries Table
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(277, 8, "Entries", "1", 1, "L", true, 0, "")

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(12, 7, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Thock No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, "Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Phone", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Village", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Time", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Variety", "1", 1, "C", true, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 9)
	for i, e := range data.Entries {
		// Alternate row colors
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(245, 245, 245)
		}

		category := "SEED"
		if e.ThockCategory == "sell" {
			category = "SELL"
		}

		name := e.Name
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		village := e.Village
		if len(village) > 18 {
			village = village[:15] + "..."
		}
		variety := e.Remark
		if len(variety) > 15 {
			variety = variety[:12] + "..."
		}

		pdf.CellFormat(12, 6, fmt.Sprintf("%d", i+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 6, e.ThockNumber, "1", 0, "C", true, 0, "")
		pdf.CellFormat(20, 6, category, "1", 0, "C", true, 0, "")
		pdf.CellFormat(50, 6, name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 6, e.Phone, "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 6, village, "1", 0, "L", true, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", e.ExpectedQuantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 6, timeutil.ToIST(e.CreatedAt).Format("03:04 PM"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 6, variety, "1", 1, "L", true, 0, "")
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateDailySummaryCSV generates a CSV file for daily summary
func (s *ReportService) GenerateDailySummaryCSV(ctx context.Context, date time.Time) ([]byte, error) {
	data, err := s.GetDailySummaryData(ctx, date)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header info
	w.Write([]string{"Daily Summary Report", data.Date.Format("02-Jan-2006")})
	w.Write([]string{""})
	w.Write([]string{"Total Trucks", fmt.Sprintf("%d", data.TotalTrucks)})
	w.Write([]string{"Seed Trucks", fmt.Sprintf("%d", data.TotalSeed)})
	w.Write([]string{"Sell Trucks", fmt.Sprintf("%d", data.TotalSell)})
	w.Write([]string{"Total Quantity", fmt.Sprintf("%d", data.TotalQty)})
	w.Write([]string{""})

	// Entries header
	w.Write([]string{"#", "Thock No", "Type", "Name", "Phone", "Village", "S/O", "Qty", "Time", "Variety"})

	// Entries data
	for i, e := range data.Entries {
		category := "SEED"
		if e.ThockCategory == "sell" {
			category = "SELL"
		}
		w.Write([]string{
			fmt.Sprintf("%d", i+1),
			e.ThockNumber,
			category,
			e.Name,
			e.Phone,
			e.Village,
			e.SO,
			fmt.Sprintf("%d", e.ExpectedQuantity),
			timeutil.ToIST(e.CreatedAt).Format("03:04 PM"),
			e.Remark,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}
