package service

import (
	"blog-system/internal/auth"
	"blog-system/internal/models"
	"blog-system/internal/repository"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtService *auth.JWTService
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo:   repository.NewUserRepository(),
		jwtService: auth.NewJWTService(),
	}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// Check if username exists
	existingUser, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already taken")
	}

	// Check if email exists
	existingUser, err = s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
		Status:       "active",
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, *models.User, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, nil, errors.New("account is disabled or banned")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, nil, err
	}

	// Save refresh token to database
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days
	err = s.userRepo.SaveRefreshToken(user.ID, refreshToken, refreshExpiry)
	if err != nil {
		return nil, nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, user, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	// Validate refresh token
	userID, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify refresh token exists in database
	storedUserID, err := s.userRepo.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	if storedUserID != userID {
		return nil, errors.New("invalid refresh token")
	}

	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status != "active" {
		return nil, errors.New("account is disabled or banned")
	}

	// Generate new tokens
	newAccessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token
	err = s.userRepo.DeleteRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Save new refresh token
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	err = s.userRepo.SaveRefreshToken(user.ID, newRefreshToken, refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.userRepo.DeleteRefreshToken(refreshToken)
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*auth.Claims, error) {
	return s.jwtService.ValidateToken(tokenString)
}
