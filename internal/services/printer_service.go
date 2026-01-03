package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	PrinterURL = "http://192.168.15.101:5000"
)

type PrinterService struct {
	client  *http.Client
	baseURL string
}

type PrintFullRequest struct {
	Line1  string `json:"line1"`
	Line2  string `json:"line2"`
	Font1  string `json:"font1"`
	Font2  string `json:"font2"`
	Copies int    `json:"copies"`
}

type PrintResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewPrinterService() *PrinterService {
	return &PrinterService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: PrinterURL,
	}
}

// PrintRoomEntryLabel prints a label with thock number and customer name
func (s *PrinterService) PrintRoomEntryLabel(thockNumber, customerName string, copies int) error {
	if copies < 1 {
		copies = 1
	}
	req := PrintFullRequest{
		Line1:  thockNumber,
		Line2:  customerName,
		Font1:  "5", // XL font for thock number
		Font2:  "4", // L font for customer name
		Copies: copies,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal print request: %w", err)
	}

	resp, err := s.client.Post(
		s.baseURL+"/print-full",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to send print request: %w", err)
	}
	defer resp.Body.Close()

	var printResp PrintResponse
	if err := json.NewDecoder(resp.Body).Decode(&printResp); err != nil {
		return fmt.Errorf("failed to decode print response: %w", err)
	}

	if !printResp.Success {
		return fmt.Errorf("print failed: %s", printResp.Message)
	}

	return nil
}
