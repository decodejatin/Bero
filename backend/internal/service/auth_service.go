package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	ErrInvalidOtp         = errors.New("invalid OTP")
	ErrOtpExpired         = errors.New("OTP expired")
	ErrTooManyAttempts    = errors.New("too many verification attempts")
	ErrUnauthorized       = errors.New("unauthorized")
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// AuthService defines authentication business logic
type AuthService interface {
	SendOtp(ctx context.Context, phoneNumber string) (*OtpResponse, error)
	VerifyOtp(ctx context.Context, phoneNumber, otp, requestID string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Logout(ctx context.Context, userID string) error
	ValidateToken(tokenString string) (*Claims, error)
}

// OtpResponse returned when OTP is sent
type OtpResponse struct {
	RequestID        string `json:"request_id"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// AuthResponse returned after successful verification
type AuthResponse struct {
	User   *domain.User       `json:"user"`
	Tokens *domain.AuthTokens `json:"tokens"`
	IsNew  bool               `json:"is_new"`
}

// Claims for JWT
type Claims struct {
	UserID   string          `json:"user_id"`
	Phone    string          `json:"phone"`
	UserType domain.UserType `json:"user_type"`
	jwt.RegisteredClaims
}

type authService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
	jwtCfg   JWTConfig
}

// NewAuthService creates a new auth service
func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserRepository, jwtCfg JWTConfig) AuthService {
	return &authService{
		authRepo: authRepo,
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *authService) SendOtp(ctx context.Context, phoneNumber string) (*OtpResponse, error) {
	// Validate phone number
	if !isValidIndianPhone(phoneNumber) {
		return nil, ErrInvalidPhoneNumber
	}

	// Normalize phone number
	normalizedPhone := normalizePhone(phoneNumber)

	// Generate 6-digit OTP
	otp, err := generateOtp(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash OTP for storage
	otpHash := hashString(otp)

	requestID := uuid.New().String()
	expiresAt := time.Now().Add(5 * time.Minute)

	otpRequest := &domain.OtpRequest{
		ID:          requestID,
		PhoneNumber: normalizedPhone,
		OtpHash:     otpHash,
		ExpiresAt:   expiresAt,
		Verified:    false,
	}

	if err := s.authRepo.CreateOtpRequest(ctx, otpRequest); err != nil {
		return nil, fmt.Errorf("failed to save OTP request: %w", err)
	}

	// TODO: Integrate with SMS provider (MSG91, Twilio, etc.)
	// For now, log the OTP for development
	fmt.Printf("[DEV] OTP for %s: %s\n", normalizedPhone, otp)

	return &OtpResponse{
		RequestID:        requestID,
		ExpiresInSeconds: 300,
	}, nil
}

func (s *authService) VerifyOtp(ctx context.Context, phoneNumber, otp, requestID string) (*AuthResponse, error) {
	normalizedPhone := normalizePhone(phoneNumber)

	// ==========================================
	// DEVELOPMENT BACKDOOR / BYPASS
	// ==========================================
	// If OTP is "000000", bypass all checks and allow login
	if otp == "000000" {
		// Get or create user directly
		user, isNew, err := s.getOrCreateUser(ctx, normalizedPhone)
		if err != nil {
			return nil, fmt.Errorf("bypass failed to get/create user: %w", err)
		}

		// Generate tokens
		tokens, err := s.generateTokens(user)
		if err != nil {
			return nil, fmt.Errorf("bypass failed to generate tokens: %w", err)
		}

		// Save session
		session := &domain.Session{
			ID:           uuid.New().String(),
			UserID:       user.ID,
			RefreshToken: hashString(tokens.RefreshToken),
			ExpiresAt:    time.Now().Add(s.jwtCfg.RefreshTokenTTL),
		}
		if err := s.authRepo.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("bypass failed to create session: %w", err)
		}

		return &AuthResponse{
			User:   user,
			Tokens: tokens,
			IsNew:  isNew,
		}, nil
	}
	// ==========================================
	// END BACKDOOR
	// ==========================================

	// Get OTP request
	otpReq, err := s.authRepo.GetOtpRequest(ctx, requestID)
	if err != nil {
		return nil, ErrInvalidOtp
	}

	// Check phone matches
	if otpReq.PhoneNumber != normalizedPhone {
		return nil, ErrInvalidOtp
	}

	// Check expiry
	if time.Now().After(otpReq.ExpiresAt) {
		return nil, ErrOtpExpired
	}

	// Check attempts
	if otpReq.AttemptCount >= 5 {
		return nil, ErrTooManyAttempts
	}

	// Verify OTP hash
	if hashString(otp) != otpReq.OtpHash {
		otpReq.AttemptCount++
		s.authRepo.UpdateOtpRequest(ctx, otpReq)
		return nil, ErrInvalidOtp
	}

	// Mark as verified
	otpReq.Verified = true
	s.authRepo.UpdateOtpRequest(ctx, otpReq)

	// Get or create user
	user, isNew, err := s.getOrCreateUser(ctx, normalizedPhone)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save session
	session := &domain.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		RefreshToken: hashString(tokens.RefreshToken),
		ExpiresAt:    time.Now().Add(s.jwtCfg.RefreshTokenTTL),
	}
	if err := s.authRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResponse{
		User:   user,
		Tokens: tokens,
		IsNew:  isNew,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	tokenHash := hashString(refreshToken)

	session, err := s.authRepo.GetSessionByRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if time.Now().After(session.ExpiresAt) {
		s.authRepo.DeleteSession(ctx, session.ID)
		return nil, ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Generate new tokens
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Update session with new refresh token
	session.RefreshToken = hashString(tokens.RefreshToken)
	session.ExpiresAt = time.Now().Add(s.jwtCfg.RefreshTokenTTL)
	s.authRepo.CreateSession(ctx, session)

	return tokens, nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	return s.authRepo.DeleteUserSessions(ctx, userID)
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtCfg.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrUnauthorized
}

func (s *authService) getOrCreateUser(ctx context.Context, phone string) (*domain.User, bool, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err == nil {
		return user, false, nil
	}

	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, false, err
	}

	// Create new user
	newUser := &domain.User{
		ID:               uuid.New().String(),
		PhoneNumber:      phone,
		AadhaarKycStatus: domain.KycStatusNone,
		UserType:         domain.UserTypeNone, // Default to NONE to force role selection
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, false, err
	}

	return newUser, true, nil
}

func (s *authService) generateTokens(user *domain.User) (*domain.AuthTokens, error) {
	now := time.Now()
	accessExpiry := now.Add(s.jwtCfg.AccessTokenTTL)
	_ = now.Add(s.jwtCfg.RefreshTokenTTL) // refreshExpiry calculated but refresh tokens use random bytes

	// Access token
	accessClaims := &Claims{
		UserID:   user.ID,
		Phone:    user.PhoneNumber,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtCfg.SecretKey))
	if err != nil {
		return nil, err
	}

	// Refresh token (simple random string)
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshTokenString := hex.EncodeToString(refreshTokenBytes)

	return &domain.AuthTokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.jwtCfg.AccessTokenTTL.Seconds()),
	}, nil
}

// Helper functions
func isValidIndianPhone(phone string) bool {
	cleaned := regexp.MustCompile(`[\s\-]`).ReplaceAllString(phone, "")
	patterns := []string{
		`^\+91\d{10}$`, // +91XXXXXXXXXX
		`^91\d{10}$`,   // 91XXXXXXXXXX
		`^\d{10}$`,     // XXXXXXXXXX
	}
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, cleaned); matched {
			return true
		}
	}
	return false
}

func normalizePhone(phone string) string {
	cleaned := regexp.MustCompile(`[\s\-]`).ReplaceAllString(phone, "")
	if len(cleaned) == 10 {
		return "+91" + cleaned
	}
	if len(cleaned) == 12 && cleaned[:2] == "91" {
		return "+" + cleaned
	}
	return cleaned
}

func generateOtp(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[n.Int64()]
	}
	return string(otp), nil
}

func hashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
