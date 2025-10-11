package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/utils"
)

// Service handles authentication operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new authentication service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Register creates a new user account
func (s *Service) Register(req *RegisterRequest) (*User, error) {
	// Check if user already exists
	var existingUser User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this username or email already exists")
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user
	user := User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Login authenticates a user and returns JWT tokens
func (s *Service) Login(req *LoginRequest) (*AuthResponse, error) {
	// Find user by username
	var user User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user,
		Message:      "Login successful",
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *Service) ValidateToken(tokenString string) (*User, error) {
	// Validate token using JWT utility
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Extract user ID from claims
	userId, ok := claims["userId"].(float64)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	// Get user from database
	var user User
	if err := s.db.First(&user, uint(userId)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// RefreshToken generates new access token using refresh token
func (s *Service) RefreshToken(refreshToken string) (string, error) {
	newAccessToken, err := utils.RefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	return newAccessToken, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID uint) (*User, error) {
	var user User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

// GetUserFromToken extracts user information from JWT token
func (s *Service) GetUserFromToken(tokenString string) (*User, error) {
	userID, err := utils.GetUserIDFromToken(tokenString)
	if err != nil {
		return nil, err
	}

	return s.GetUserByID(userID)
}
