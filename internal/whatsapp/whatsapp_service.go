package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WhatsAppProvider defines the interface for WhatsApp API providers
type WhatsAppProvider interface {
	SendMessage(phone, message string, templateName string, params map[string]string) error
	SendTemplateMessage(phone, templateName string, params []string) error
	CheckNumberExists(phone string) (bool, error)
	GetName() string
}

// WhatsAppConfig holds configuration for WhatsApp providers
type WhatsAppConfig struct {
	Provider     string // "aisensy", "interakt", "gupshup"
	APIKey       string
	APISecret    string // Some providers need this
	BusinessID   string // WhatsApp Business Account ID
	PhoneNumberID string // WhatsApp Phone Number ID
	BaseURL      string
}

// AiSensyService implements WhatsApp via AiSensy
type AiSensyService struct {
	config  *WhatsAppConfig
	client  *http.Client
}

// NewAiSensyService creates a new AiSensy WhatsApp service
func NewAiSensyService(apiKey string) *AiSensyService {
	return &AiSensyService{
		config: &WhatsAppConfig{
			Provider: "aisensy",
			APIKey:   apiKey,
			BaseURL:  "https://backend.aisensy.com/campaign/t1/api/v2",
		},
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMessage sends a text message via AiSensy
func (s *AiSensyService) SendMessage(phone, message string, templateName string, params map[string]string) error {
	// AiSensy requires template-based messages for Business API
	// For now, we'll use the template approach
	return s.SendTemplateMessage(phone, templateName, mapToSlice(params))
}

// SendTemplateMessage sends a template message via AiSensy
func (s *AiSensyService) SendTemplateMessage(phone, templateName string, params []string) error {
	payload := map[string]interface{}{
		"apiKey":       s.config.APIKey,
		"campaignName": templateName,
		"destination":  formatPhoneNumber(phone),
		"userName":     "Customer",
		"templateParams": params,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AiSensy API error: %s", string(body))
	}

	return nil
}

// CheckNumberExists checks if a number is on WhatsApp (AiSensy doesn't support this directly)
func (s *AiSensyService) CheckNumberExists(phone string) (bool, error) {
	// AiSensy doesn't provide a direct API to check if number is on WhatsApp
	// We'll assume it exists and handle the error if it doesn't
	return true, nil
}

// GetName returns the provider name
func (s *AiSensyService) GetName() string {
	return "AiSensy"
}

// InteraktService implements WhatsApp via Interakt
type InteraktService struct {
	config *WhatsAppConfig
	client *http.Client
}

// NewInteraktService creates a new Interakt WhatsApp service
func NewInteraktService(apiKey string) *InteraktService {
	return &InteraktService{
		config: &WhatsAppConfig{
			Provider: "interakt",
			APIKey:   apiKey,
			BaseURL:  "https://api.interakt.ai/v1/public",
		},
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMessage sends a message via Interakt
func (s *InteraktService) SendMessage(phone, message string, templateName string, params map[string]string) error {
	return s.SendTemplateMessage(phone, templateName, mapToSlice(params))
}

// SendTemplateMessage sends a template message via Interakt
func (s *InteraktService) SendTemplateMessage(phone, templateName string, params []string) error {
	// Build template params
	bodyParams := make([]map[string]string, len(params))
	for i, param := range params {
		bodyParams[i] = map[string]string{"type": "text", "text": param}
	}

	payload := map[string]interface{}{
		"countryCode":  "+91",
		"phoneNumber":  formatPhoneNumber(phone),
		"callbackData": "cold_storage_msg",
		"type":         "Template",
		"template": map[string]interface{}{
			"name":         templateName,
			"languageCode": "en",
			"bodyValues":   params,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.BaseURL+"/message/", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Interakt API error: %s", string(body))
	}

	return nil
}

// CheckNumberExists checks if a number is on WhatsApp
func (s *InteraktService) CheckNumberExists(phone string) (bool, error) {
	// Interakt doesn't provide direct number check
	return true, nil
}

// GetName returns the provider name
func (s *InteraktService) GetName() string {
	return "Interakt"
}

// MessagingService handles WhatsApp-first with SMS fallback
type MessagingService struct {
	whatsapp    WhatsAppProvider
	smsProvider SMSProvider
	logRepo     MessageLogRepo
	enabled     bool
}

// SMSProvider interface for SMS fallback
type SMSProvider interface {
	SendSMS(phone, message, messageType string, customerID int) error
}

// MessageLogRepo interface for logging
type MessageLogRepo interface {
	Create(ctx context.Context, log *MessageLog) error
}

// MessageLog represents a message log entry
type MessageLog struct {
	CustomerID   int
	Phone        string
	Message      string
	Channel      string // "whatsapp" or "sms"
	Status       string
	ErrorMessage string
	Cost         float64
}

// NewMessagingService creates a new messaging service with WhatsApp-first fallback
func NewMessagingService(whatsapp WhatsAppProvider, sms SMSProvider) *MessagingService {
	return &MessagingService{
		whatsapp:    whatsapp,
		smsProvider: sms,
		enabled:     whatsapp != nil,
	}
}

// SendMessage tries WhatsApp first, falls back to SMS
func (s *MessagingService) SendMessage(phone, message, templateName string, params map[string]string, messageType string, customerID int) error {
	// If WhatsApp is not configured, use SMS directly
	if s.whatsapp == nil || !s.enabled {
		return s.smsProvider.SendSMS(phone, message, messageType, customerID)
	}

	// Try WhatsApp first
	err := s.whatsapp.SendMessage(phone, message, templateName, params)
	if err == nil {
		// WhatsApp sent successfully
		return nil
	}

	// WhatsApp failed, fallback to SMS
	return s.smsProvider.SendSMS(phone, message, messageType, customerID)
}

// SendBulkMessage sends to multiple recipients with WhatsApp-first fallback
func (s *MessagingService) SendBulkMessage(phones []string, message, templateName string, params map[string]string, messageType string, customerIDs []int) (whatsappSuccess, smsSuccess, failed int) {
	for i, phone := range phones {
		customerID := 0
		if i < len(customerIDs) {
			customerID = customerIDs[i]
		}

		err := s.SendMessage(phone, message, templateName, params, messageType, customerID)
		if err != nil {
			failed++
		} else {
			// We don't know which channel succeeded without more tracking
			// For now, assume WhatsApp if enabled
			if s.whatsapp != nil && s.enabled {
				whatsappSuccess++
			} else {
				smsSuccess++
			}
		}
	}
	return
}

// SetWhatsAppEnabled enables/disables WhatsApp
func (s *MessagingService) SetWhatsAppEnabled(enabled bool) {
	s.enabled = enabled
}

// Helper functions

func formatPhoneNumber(phone string) string {
	// Remove any non-digit characters
	cleaned := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			cleaned += string(c)
		}
	}

	// Add country code if not present
	if len(cleaned) == 10 {
		return "91" + cleaned
	}
	if len(cleaned) == 12 && cleaned[:2] == "91" {
		return cleaned
	}
	return cleaned
}

func mapToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

// GenericWhatsAppService implements WhatsApp via Meta Cloud API (works with any BSP)
// This is the standard WhatsApp Business Cloud API that most providers support
type GenericWhatsAppService struct {
	config *WhatsAppConfig
	client *http.Client
}

// NewGenericWhatsAppService creates a new Generic WhatsApp service
// apiKey: Access Token from Meta Business Suite or BSP
// phoneNumberID: WhatsApp Business Phone Number ID
func NewGenericWhatsAppService(apiKey, phoneNumberID string) *GenericWhatsAppService {
	return &GenericWhatsAppService{
		config: &WhatsAppConfig{
			Provider:      "generic",
			APIKey:        apiKey,
			PhoneNumberID: phoneNumberID,
			BaseURL:       "https://graph.facebook.com/v18.0",
		},
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetBaseURL allows overriding the API base URL (for BSP proxies)
func (s *GenericWhatsAppService) SetBaseURL(url string) {
	s.config.BaseURL = url
}

// SendMessage sends a text message via WhatsApp Cloud API
func (s *GenericWhatsAppService) SendMessage(phone, message string, templateName string, params map[string]string) error {
	// If template is specified, use template message
	if templateName != "" {
		return s.SendTemplateMessage(phone, templateName, mapToSlice(params))
	}

	// Otherwise send a regular text message (only works within 24hr window)
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                formatPhoneNumber(phone),
		"type":              "text",
		"text": map[string]string{
			"preview_url": "false",
			"body":        message,
		},
	}

	return s.sendRequest(payload)
}

// SendTemplateMessage sends a template message via WhatsApp Cloud API
func (s *GenericWhatsAppService) SendTemplateMessage(phone, templateName string, params []string) error {
	// Build template components
	components := []map[string]interface{}{}

	if len(params) > 0 {
		bodyParams := make([]map[string]string, len(params))
		for i, param := range params {
			bodyParams[i] = map[string]string{"type": "text", "text": param}
		}
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": bodyParams,
		})
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                formatPhoneNumber(phone),
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]string{
				"code": "en",
			},
			"components": components,
		},
	}

	return s.sendRequest(payload)
}

// sendRequest sends the API request
func (s *GenericWhatsAppService) sendRequest(payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", s.config.BaseURL, s.config.PhoneNumberID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Parse error response
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if errObj, ok := errResp["error"].(map[string]interface{}); ok {
				if msg, ok := errObj["message"].(string); ok {
					return fmt.Errorf("WhatsApp API error: %s", msg)
				}
			}
		}
		return fmt.Errorf("WhatsApp API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// CheckNumberExists checks if a number is registered on WhatsApp
func (s *GenericWhatsAppService) CheckNumberExists(phone string) (bool, error) {
	// The Cloud API doesn't have a direct number check
	// We assume it exists and handle errors gracefully
	return true, nil
}

// GetName returns the provider name
func (s *GenericWhatsAppService) GetName() string {
	return "Generic (Meta Cloud API)"
}

// GupshupService implements WhatsApp via Gupshup
type GupshupService struct {
	config *WhatsAppConfig
	client *http.Client
}

// NewGupshupService creates a new Gupshup WhatsApp service
func NewGupshupService(apiKey, sourcePhone string) *GupshupService {
	return &GupshupService{
		config: &WhatsAppConfig{
			Provider:      "gupshup",
			APIKey:        apiKey,
			PhoneNumberID: sourcePhone, // Source phone number
			BaseURL:       "https://api.gupshup.io/sm/api/v1",
		},
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMessage sends a message via Gupshup
func (s *GupshupService) SendMessage(phone, message string, templateName string, params map[string]string) error {
	if templateName != "" {
		return s.SendTemplateMessage(phone, templateName, mapToSlice(params))
	}

	// Send session message
	formData := fmt.Sprintf("channel=whatsapp&source=%s&destination=%s&message=%s&src.name=ColdStorage",
		s.config.PhoneNumberID,
		formatPhoneNumber(phone),
		message,
	)

	req, err := http.NewRequest("POST", s.config.BaseURL+"/msg", bytes.NewBufferString(formData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Gupshup API error: %s", string(body))
	}

	return nil
}

// SendTemplateMessage sends a template message via Gupshup
func (s *GupshupService) SendTemplateMessage(phone, templateName string, params []string) error {
	// Gupshup template format
	templateData := map[string]interface{}{
		"id":     templateName,
		"params": params,
	}
	templateJSON, _ := json.Marshal(templateData)

	formData := fmt.Sprintf("channel=whatsapp&source=%s&destination=%s&template=%s&src.name=ColdStorage",
		s.config.PhoneNumberID,
		formatPhoneNumber(phone),
		string(templateJSON),
	)

	req, err := http.NewRequest("POST", s.config.BaseURL+"/template/msg", bytes.NewBufferString(formData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Gupshup API error: %s", string(body))
	}

	return nil
}

// CheckNumberExists checks if number is on WhatsApp via Gupshup
func (s *GupshupService) CheckNumberExists(phone string) (bool, error) {
	return true, nil
}

// GetName returns the provider name
func (s *GupshupService) GetName() string {
	return "Gupshup"
}

// CreateWhatsAppProvider creates a WhatsApp provider based on provider name
func CreateWhatsAppProvider(provider, apiKey, phoneNumberID string) WhatsAppProvider {
	switch provider {
	case "aisensy":
		return NewAiSensyService(apiKey)
	case "interakt":
		return NewInteraktService(apiKey)
	case "gupshup":
		return NewGupshupService(apiKey, phoneNumberID)
	case "generic", "meta", "cloud":
		return NewGenericWhatsAppService(apiKey, phoneNumberID)
	default:
		// Default to generic if provider not recognized
		if apiKey != "" && phoneNumberID != "" {
			return NewGenericWhatsAppService(apiKey, phoneNumberID)
		}
		return nil
	}
}
