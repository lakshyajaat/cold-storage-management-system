package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cold-backend/internal/models"
)

// SMSProvider is an interface for sending SMS messages
type SMSProvider interface {
	SendOTP(phone, otp string) error
	SendSMS(phone, message, messageType string, customerID int) error
	SendBulkSMS(phones []string, message string, customerIDs []int) (int, int, error)
	SetLogRepository(repo SMSLogRepo)
	SetConfig(config *SMSConfig)
}

// SMSLogRepo interface for logging
type SMSLogRepo interface {
	Create(ctx context.Context, log *models.SMSLog) error
}

// SMSConfig holds SMS configuration
type SMSConfig struct {
	Route       string // "q" (quick/expensive), "dlt" (cheap/production), "v3" (promotional)
	SenderID    string // For DLT route (e.g., "COLDST")
	TemplateID  string // For DLT route
	EntityID    string // For DLT route (PEID)
	CostPerSMS  float64
}

// Fast2SMSService implements SMSProvider for Fast2SMS (India)
type Fast2SMSService struct {
	APIKey  string
	Config  *SMSConfig
	LogRepo SMSLogRepo
}

// NewFast2SMSService creates a new Fast2SMS service
func NewFast2SMSService(apiKey string) *Fast2SMSService {
	return &Fast2SMSService{
		APIKey: apiKey,
		Config: &SMSConfig{
			Route:      "q", // Default to quick route, can be changed via settings
			CostPerSMS: 5.0, // Quick route cost
		},
	}
}

// SetLogRepository sets the SMS log repository
func (s *Fast2SMSService) SetLogRepository(repo SMSLogRepo) {
	s.LogRepo = repo
}

// SetConfig sets the SMS configuration
func (s *Fast2SMSService) SetConfig(config *SMSConfig) {
	if config != nil {
		s.Config = config
	}
}

// SendOTP sends an OTP code via Fast2SMS
func (s *Fast2SMSService) SendOTP(phone, otp string) error {
	message := fmt.Sprintf("Your Cold Storage OTP is %s. Valid for 5 minutes. Do not share this code with anyone.", otp)
	return s.SendSMS(phone, message, models.SMSTypeOTP, 0)
}

// SendSMS sends a single SMS message
func (s *Fast2SMSService) SendSMS(phone, message, messageType string, customerID int) error {
	// Build API URL based on route
	var apiURL string

	switch s.Config.Route {
	case "dlt":
		// DLT route (cheaper, requires registration)
		apiURL = fmt.Sprintf(
			"https://www.fast2sms.com/dev/bulkV2?authorization=%s&route=dlt&sender_id=%s&message=%s&variables_values=%s&flash=0&numbers=%s",
			url.QueryEscape(s.APIKey),
			url.QueryEscape(s.Config.SenderID),
			url.QueryEscape(s.Config.TemplateID),
			url.QueryEscape(message),
			url.QueryEscape(phone),
		)
	case "v3":
		// Promotional route (cheapest, 9am-9pm only)
		apiURL = fmt.Sprintf(
			"https://www.fast2sms.com/dev/bulkV2?authorization=%s&route=v3&sender_id=%s&message=%s&language=english&numbers=%s",
			url.QueryEscape(s.APIKey),
			url.QueryEscape(s.Config.SenderID),
			url.QueryEscape(message),
			url.QueryEscape(phone),
		)
	default:
		// Quick route (expensive but works immediately)
		apiURL = fmt.Sprintf(
			"https://www.fast2sms.com/dev/bulkV2?authorization=%s&route=q&message=%s&language=english&flash=0&numbers=%s",
			url.QueryEscape(s.APIKey),
			url.QueryEscape(message),
			url.QueryEscape(phone),
		)
	}

	// Create log entry
	smsLog := &models.SMSLog{
		CustomerID:  customerID,
		Phone:       phone,
		MessageType: messageType,
		Message:     message,
		Status:      models.SMSStatusPending,
		Cost:        s.Config.CostPerSMS,
	}

	// Send request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		smsLog.Status = models.SMSStatusFailed
		smsLog.ErrorMessage = err.Error()
		s.logSMS(smsLog)
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		smsLog.Status = models.SMSStatusFailed
		smsLog.ErrorMessage = err.Error()
		s.logSMS(smsLog)
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse response
	var apiResp map[string]interface{}
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode != http.StatusOK {
		smsLog.Status = models.SMSStatusFailed
		smsLog.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		s.logSMS(smsLog)
		return fmt.Errorf("SMS API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Check for API-level errors
	if strings.Contains(string(body), "\"return\":false") {
		smsLog.Status = models.SMSStatusFailed
		smsLog.ErrorMessage = string(body)
		s.logSMS(smsLog)
		return fmt.Errorf("SMS API error: %s", string(body))
	}

	// Success
	smsLog.Status = models.SMSStatusSent
	if requestID, ok := apiResp["request_id"].(string); ok {
		smsLog.ReferenceID = requestID
	}
	s.logSMS(smsLog)

	return nil
}

// SendBulkSMS sends SMS to multiple phones
func (s *Fast2SMSService) SendBulkSMS(phones []string, message string, customerIDs []int) (int, int, error) {
	success := 0
	failed := 0

	for i, phone := range phones {
		customerID := 0
		if i < len(customerIDs) {
			customerID = customerIDs[i]
		}

		err := s.SendSMS(phone, message, models.SMSTypeBulk, customerID)
		if err != nil {
			failed++
		} else {
			success++
		}

		// Rate limit: 1 SMS per 100ms to avoid API throttling
		time.Sleep(100 * time.Millisecond)
	}

	return success, failed, nil
}

// logSMS logs the SMS to database
func (s *Fast2SMSService) logSMS(log *models.SMSLog) {
	if s.LogRepo == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.LogRepo.Create(ctx, log)
	}()
}

// MockSMSService is a mock implementation for testing (prints OTP to console)
type MockSMSService struct {
	LogRepo SMSLogRepo
	Config  *SMSConfig
}

// NewMockSMSService creates a new mock SMS service
func NewMockSMSService() *MockSMSService {
	return &MockSMSService{
		Config: &SMSConfig{CostPerSMS: 0},
	}
}

// SetLogRepository sets the SMS log repository
func (s *MockSMSService) SetLogRepository(repo SMSLogRepo) {
	s.LogRepo = repo
}

// SetConfig sets the SMS configuration
func (s *MockSMSService) SetConfig(config *SMSConfig) {
	if config != nil {
		s.Config = config
	}
}

// SendOTP prints the OTP to console instead of sending SMS (for testing)
func (s *MockSMSService) SendOTP(phone, otp string) error {
	message := fmt.Sprintf("Your Cold Storage OTP is %s. Valid for 5 minutes.", otp)
	return s.SendSMS(phone, message, models.SMSTypeOTP, 0)
}

// SendSMS logs the SMS to console
func (s *MockSMSService) SendSMS(phone, message, messageType string, customerID int) error {
	fmt.Printf("\n========== MOCK SMS ==========\n")
	fmt.Printf("To: %s\n", phone)
	fmt.Printf("Type: %s\n", messageType)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("==============================\n\n")

	// Log to database
	if s.LogRepo != nil {
		smsLog := &models.SMSLog{
			CustomerID:  customerID,
			Phone:       phone,
			MessageType: messageType,
			Message:     message,
			Status:      models.SMSStatusSent,
			Cost:        0,
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.LogRepo.Create(ctx, smsLog)
		}()
	}

	return nil
}

// SendBulkSMS sends bulk SMS (mock)
func (s *MockSMSService) SendBulkSMS(phones []string, message string, customerIDs []int) (int, int, error) {
	for i, phone := range phones {
		customerID := 0
		if i < len(customerIDs) {
			customerID = customerIDs[i]
		}
		s.SendSMS(phone, message, models.SMSTypeBulk, customerID)
	}
	return len(phones), 0, nil
}

// =============================================================================
// UnifiedMessagingService - WhatsApp-first with SMS fallback
// =============================================================================

// WhatsAppConfig holds WhatsApp provider configuration
type WhatsAppConfig struct {
	Enabled       bool
	Provider      string // "generic", "aisensy", "interakt", "gupshup"
	APIKey        string
	PhoneNumberID string // Required for generic/gupshup
	CostPerMsg    float64
}

// UnifiedMessagingService handles WhatsApp-first with SMS fallback
type UnifiedMessagingService struct {
	smsProvider     SMSProvider
	whatsappConfig  *WhatsAppConfig
	whatsappClient  *http.Client
	LogRepo         SMSLogRepo
}

// NewUnifiedMessagingService creates a new unified messaging service
func NewUnifiedMessagingService(smsProvider SMSProvider) *UnifiedMessagingService {
	return &UnifiedMessagingService{
		smsProvider:    smsProvider,
		whatsappConfig: &WhatsAppConfig{Enabled: false},
		whatsappClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetLogRepository sets the SMS log repository
func (s *UnifiedMessagingService) SetLogRepository(repo SMSLogRepo) {
	s.LogRepo = repo
	// Also set for underlying SMS provider
	if s.smsProvider != nil {
		s.smsProvider.SetLogRepository(repo)
	}
}

// SetConfig sets the SMS configuration
func (s *UnifiedMessagingService) SetConfig(config *SMSConfig) {
	if s.smsProvider != nil {
		s.smsProvider.SetConfig(config)
	}
}

// SetWhatsAppConfig configures WhatsApp settings
func (s *UnifiedMessagingService) SetWhatsAppConfig(config *WhatsAppConfig) {
	if config != nil {
		s.whatsappConfig = config
	}
}

// SendOTP sends OTP via WhatsApp first, falls back to SMS
func (s *UnifiedMessagingService) SendOTP(phone, otp string) error {
	message := fmt.Sprintf("Your Cold Storage OTP is %s. Valid for 5 minutes. Do not share this code with anyone.", otp)
	return s.SendSMS(phone, message, models.SMSTypeOTP, 0)
}

// SendSMS tries WhatsApp first, falls back to SMS
func (s *UnifiedMessagingService) SendSMS(phone, message, messageType string, customerID int) error {
	// If WhatsApp is enabled, try it first
	if s.whatsappConfig != nil && s.whatsappConfig.Enabled && s.whatsappConfig.APIKey != "" {
		err := s.sendWhatsApp(phone, message, messageType, customerID)
		if err == nil {
			return nil // WhatsApp succeeded
		}
		log.Printf("[UnifiedMessaging] WhatsApp failed for %s, falling back to SMS: %v", phone, err)
	}

	// Fall back to SMS
	return s.smsProvider.SendSMS(phone, message, messageType, customerID)
}

// SendBulkSMS sends to multiple recipients
func (s *UnifiedMessagingService) SendBulkSMS(phones []string, message string, customerIDs []int) (int, int, error) {
	success := 0
	failed := 0

	for i, phone := range phones {
		customerID := 0
		if i < len(customerIDs) {
			customerID = customerIDs[i]
		}

		err := s.SendSMS(phone, message, models.SMSTypeBulk, customerID)
		if err != nil {
			failed++
		} else {
			success++
		}

		// Rate limit
		time.Sleep(100 * time.Millisecond)
	}

	return success, failed, nil
}

// sendWhatsApp sends message via WhatsApp
func (s *UnifiedMessagingService) sendWhatsApp(phone, message, messageType string, customerID int) error {
	cfg := s.whatsappConfig

	// Log entry
	smsLog := &models.SMSLog{
		CustomerID:  customerID,
		Phone:       phone,
		MessageType: messageType,
		Message:     message,
		Status:      models.SMSStatusPending,
		Cost:        cfg.CostPerMsg,
		Channel:     "whatsapp",
	}

	var err error
	switch cfg.Provider {
	case "generic", "meta", "cloud", "":
		err = s.sendGenericWhatsApp(phone, message)
	case "aisensy":
		err = s.sendAiSensyWhatsApp(phone, message)
	case "interakt":
		err = s.sendInteraktWhatsApp(phone, message)
	case "gupshup":
		err = s.sendGupshupWhatsApp(phone, message)
	default:
		err = s.sendGenericWhatsApp(phone, message)
	}

	if err != nil {
		smsLog.Status = models.SMSStatusFailed
		smsLog.ErrorMessage = err.Error()
		s.logMessage(smsLog)
		return err
	}

	smsLog.Status = models.SMSStatusSent
	s.logMessage(smsLog)
	return nil
}

// sendGenericWhatsApp sends via Meta Cloud API
func (s *UnifiedMessagingService) sendGenericWhatsApp(phone, message string) error {
	cfg := s.whatsappConfig

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                formatPhone(phone),
		"type":              "text",
		"text": map[string]string{
			"preview_url": "false",
			"body":        message,
		},
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", cfg.PhoneNumberID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := s.whatsappClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("WhatsApp API error: %s", string(body))
	}

	return nil
}

// sendAiSensyWhatsApp sends via AiSensy
func (s *UnifiedMessagingService) sendAiSensyWhatsApp(phone, message string) error {
	cfg := s.whatsappConfig

	payload := map[string]interface{}{
		"apiKey":      cfg.APIKey,
		"destination": formatPhone(phone),
		"message":     message,
	}

	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://backend.aisensy.com/campaign/t1/api/v2", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.whatsappClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AiSensy error: %s", string(body))
	}

	return nil
}

// sendInteraktWhatsApp sends via Interakt
func (s *UnifiedMessagingService) sendInteraktWhatsApp(phone, message string) error {
	cfg := s.whatsappConfig

	payload := map[string]interface{}{
		"countryCode":  "+91",
		"phoneNumber":  formatPhone(phone),
		"callbackData": "cold_storage",
		"type":         "Text",
		"data": map[string]string{
			"message": message,
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.interakt.ai/v1/public/message/", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+cfg.APIKey)

	resp, err := s.whatsappClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Interakt error: %s", string(body))
	}

	return nil
}

// sendGupshupWhatsApp sends via Gupshup
func (s *UnifiedMessagingService) sendGupshupWhatsApp(phone, message string) error {
	cfg := s.whatsappConfig

	formData := fmt.Sprintf("channel=whatsapp&source=%s&destination=%s&message=%s&src.name=ColdStorage",
		cfg.PhoneNumberID,
		formatPhone(phone),
		url.QueryEscape(message),
	)

	req, err := http.NewRequest("POST", "https://api.gupshup.io/sm/api/v1/msg", strings.NewReader(formData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", cfg.APIKey)

	resp, err := s.whatsappClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Gupshup error: %s", string(body))
	}

	return nil
}

// logMessage logs to database
func (s *UnifiedMessagingService) logMessage(log *models.SMSLog) {
	if s.LogRepo == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.LogRepo.Create(ctx, log)
	}()
}

// formatPhone formats phone number for WhatsApp
func formatPhone(phone string) string {
	cleaned := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			cleaned += string(c)
		}
	}
	if len(cleaned) == 10 {
		return "91" + cleaned
	}
	return cleaned
}
