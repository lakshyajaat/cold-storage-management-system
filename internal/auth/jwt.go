package auth

import (
	"errors"
	"time"

	"cold-backend/internal/config"
	"cold-backend/internal/models"
	"cold-backend/internal/timeutil"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID              int    `json:"user_id"`
	Email               string `json:"email"`
	Role                string `json:"role"`
	HasAccountantAccess bool   `json:"has_accountant_access"`
	IsActive            bool   `json:"is_active"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	cfg *config.Config
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{cfg: cfg}
}

// GenerateToken creates a new JWT token for a user
func (j *JWTManager) GenerateToken(user *models.User) (string, error) {
	now := timeutil.Now()
	expirationTime := now.Add(time.Duration(j.cfg.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID:              user.ID,
		Email:               user.Email,
		Role:                user.Role,
		HasAccountantAccess: user.HasAccountantAccess,
		IsActive:            user.IsActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.cfg.JWT.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.cfg.JWT.Secret))
}

// ValidateToken verifies a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
