package services

import (
	"context"
	"errors"

	"cold-backend/internal/auth"
	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
)

type UserService struct {
	Repo       *repositories.UserRepository
	JWTManager *auth.JWTManager
}

func NewUserService(repo *repositories.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		Repo:       repo,
		JWTManager: jwtManager,
	}
}

func (s *UserService) CreateUser(ctx context.Context, u *models.User) error {
	// Hash password if provided
	if u.PasswordHash != "" {
		hashedPassword, err := auth.HashPassword(u.PasswordHash)
		if err != nil {
			return err
		}
		u.PasswordHash = hashedPassword
	}
	return s.Repo.Create(ctx, u)
}

func (s *UserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	return s.Repo.Get(ctx, id)
}

// ListUsers returns all users
func (s *UserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return s.Repo.List(ctx)
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	// If password is provided, hash it
	if user.PasswordHash != "" {
		hashedPassword, err := auth.HashPassword(user.PasswordHash)
		if err != nil {
			return err
		}
		user.PasswordHash = hashedPassword
	}
	return s.Repo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	return s.Repo.Delete(ctx, id)
}

// Signup creates a new user with hashed password
func (s *UserService) Signup(ctx context.Context, req *models.SignupRequest) (*models.AuthResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, errors.New("name, email, and password are required")
	}

	// Check if user already exists
	existingUser, _ := s.Repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.Repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.JWTManager.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Get user by email
	user, err := s.Repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if !auth.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.JWTManager.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}
