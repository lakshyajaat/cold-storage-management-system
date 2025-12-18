package g

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

// Authentication

func (s *Service) VerifyPins(ctx context.Context, pin1, pin2, deviceHash, ip string) (*AuthResponse, error) {
	// Check rate limiting (3 failed attempts = 1 hour lockout)
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	failedAttempts, err := s.Repo.GetFailedAttempts(ctx, deviceHash, oneHourAgo)
	if err == nil && failedAttempts >= 3 {
		s.Repo.LogAccess(ctx, deviceHash, ip, false, "rate_limited")
		return nil, errors.New("too many failed attempts")
	}

	// Get stored PIN hashes
	pin1Hash, err := s.Repo.GetConfig(ctx, "p1h")
	if err != nil {
		s.Repo.LogAccess(ctx, deviceHash, ip, false, "config_error")
		return nil, errors.New("system not configured")
	}

	pin2Hash, err := s.Repo.GetConfig(ctx, "p2h")
	if err != nil {
		s.Repo.LogAccess(ctx, deviceHash, ip, false, "config_error")
		return nil, errors.New("system not configured")
	}

	// Verify PIN 1
	if err := bcrypt.CompareHashAndPassword([]byte(pin1Hash), []byte(pin1)); err != nil {
		s.Repo.LogAccess(ctx, deviceHash, ip, false, "invalid_p1")
		return nil, errors.New("invalid credentials")
	}

	// Verify PIN 2
	if err := bcrypt.CompareHashAndPassword([]byte(pin2Hash), []byte(pin2)); err != nil {
		s.Repo.LogAccess(ctx, deviceHash, ip, false, "invalid_p2")
		return nil, errors.New("invalid credentials")
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return nil, errors.New("failed to generate session")
	}

	// Session expires in 15 minutes
	expiresAt := time.Now().Add(15 * time.Minute)

	if err := s.Repo.CreateSession(ctx, token, deviceHash, expiresAt); err != nil {
		return nil, errors.New("failed to create session")
	}

	s.Repo.LogAccess(ctx, deviceHash, ip, true, "")

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (s *Service) ValidateSession(ctx context.Context, token, deviceHash string) (bool, error) {
	session, err := s.Repo.GetSession(ctx, token)
	if err != nil {
		return false, err
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		s.Repo.DeleteSession(ctx, token)
		return false, nil
	}

	// Check device hash
	if session.DeviceHash != deviceHash {
		return false, nil
	}

	// Extend session by 15 minutes on activity
	newExpiry := time.Now().Add(15 * time.Minute)
	s.Repo.ExtendSession(ctx, token, newExpiry)

	return true, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.Repo.DeleteSession(ctx, token)
}

// Items

func (s *Service) AddItem(ctx context.Context, req *AddItemRequest) (*Item, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Floor < 0 || req.Floor > 4 {
		return nil, errors.New("floor must be 0-4")
	}
	if req.Qty < 0 {
		return nil, errors.New("quantity cannot be negative")
	}

	item := &Item{
		Name:       req.Name,
		SKU:        req.SKU,
		Floor:      req.Floor,
		CurrentQty: req.Qty,
		UnitCost:   req.UnitCost,
	}

	if err := s.Repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	// Record initial stock as 'in' transaction if qty > 0
	if req.Qty > 0 {
		txn := &Txn{
			ItemID:    item.ID,
			Type:      "in",
			Qty:       req.Qty,
			UnitPrice: req.UnitCost,
			Total:     float64(req.Qty) * req.UnitCost,
			Reason:    "initial_stock",
		}
		s.Repo.CreateTxn(ctx, txn)
	}

	return item, nil
}

func (s *Service) GetItem(ctx context.Context, id int) (*Item, error) {
	return s.Repo.GetItem(ctx, id)
}

func (s *Service) ListItems(ctx context.Context) ([]*Item, error) {
	return s.Repo.ListItems(ctx)
}

func (s *Service) UpdateItem(ctx context.Context, id int, req *UpdateItemRequest) error {
	item, err := s.Repo.GetItem(ctx, id)
	if err != nil {
		return err
	}

	name := item.Name
	sku := item.SKU
	floor := item.Floor
	unitCost := item.UnitCost

	if req.Name != "" {
		name = req.Name
	}
	if req.SKU != "" {
		sku = req.SKU
	}
	if req.Floor >= 0 && req.Floor <= 4 {
		floor = req.Floor
	}
	if req.UnitCost > 0 {
		unitCost = req.UnitCost
	}

	return s.Repo.UpdateItem(ctx, id, name, sku, floor, unitCost)
}

func (s *Service) DeleteItem(ctx context.Context, id int) error {
	return s.Repo.DeleteItem(ctx, id)
}

// Stock movements

func (s *Service) StockIn(ctx context.Context, req *StockInRequest) (*Txn, error) {
	if req.ItemID <= 0 {
		return nil, errors.New("item_id is required")
	}
	if req.Qty <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	// Verify item exists
	item, err := s.Repo.GetItem(ctx, req.ItemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	unitPrice := req.UnitPrice
	if unitPrice <= 0 {
		unitPrice = item.UnitCost
	}

	txn := &Txn{
		ItemID:    req.ItemID,
		Type:      "in",
		Qty:       req.Qty,
		UnitPrice: unitPrice,
		Total:     float64(req.Qty) * unitPrice,
		Reason:    req.Reason,
	}

	if err := s.Repo.CreateTxn(ctx, txn); err != nil {
		return nil, err
	}

	if err := s.Repo.UpdateItemQty(ctx, req.ItemID, req.Qty); err != nil {
		return nil, err
	}

	return txn, nil
}

func (s *Service) StockOut(ctx context.Context, req *StockOutRequest) (*Txn, error) {
	if req.ItemID <= 0 {
		return nil, errors.New("item_id is required")
	}
	if req.Qty <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}

	// Verify item exists and has enough stock
	item, err := s.Repo.GetItem(ctx, req.ItemID)
	if err != nil {
		return nil, errors.New("item not found")
	}
	if item.CurrentQty < req.Qty {
		return nil, errors.New("insufficient stock")
	}

	unitPrice := req.SalePrice
	if unitPrice <= 0 {
		unitPrice = item.UnitCost
	}

	txn := &Txn{
		ItemID:    req.ItemID,
		Type:      "out",
		Qty:       req.Qty,
		UnitPrice: unitPrice,
		Total:     float64(req.Qty) * unitPrice,
		Reason:    req.Reason,
	}

	if err := s.Repo.CreateTxn(ctx, txn); err != nil {
		return nil, err
	}

	if err := s.Repo.UpdateItemQty(ctx, req.ItemID, -req.Qty); err != nil {
		return nil, err
	}

	return txn, nil
}

func (s *Service) ListTxns(ctx context.Context, limit int) ([]*Txn, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.Repo.ListTxns(ctx, limit)
}

func (s *Service) GetSummary(ctx context.Context) (*Summary, error) {
	return s.Repo.GetSummary(ctx)
}

// PIN management (for initial setup)

func (s *Service) SetPins(ctx context.Context, pin1, pin2 string) error {
	if len(pin1) != 6 {
		return errors.New("PIN 1 must be 6 digits")
	}
	if len(pin2) != 8 {
		return errors.New("PIN 2 must be 8 digits")
	}

	hash1, err := bcrypt.GenerateFromPassword([]byte(pin1), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hash2, err := bcrypt.GenerateFromPassword([]byte(pin2), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.Repo.SetConfig(ctx, "p1h", string(hash1)); err != nil {
		return err
	}

	return s.Repo.SetConfig(ctx, "p2h", string(hash2))
}

// Helper functions

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateDeviceHash(userAgent, screenRes string) string {
	data := userAgent + "|" + screenRes
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
