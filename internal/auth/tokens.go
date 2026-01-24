package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/domain"
	apperr "github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
)

type JWTService struct {
	config    *config.AuthConfig
	logger    *logger.Logger
	jwtSecret []byte
	jwtParser *jwt.Parser
}

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"userId"`
	TenantID    string   `json:"tenantId"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	TenantRole  string   `json:"tenantRole"`
	Permissions []string `json:"permissions"`
}

type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresIn    int       `json:"expiresIn"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type TokenService struct {
	jwtService *JWTService
	redis      RedisClient
	logger     *logger.Logger
	config     *config.AuthConfig
}

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

func NewTokenService(cfg *config.AuthConfig, redisClient RedisClient, log *logger.Logger) *TokenService {
	return &TokenService{
		jwtService: NewJWTService(cfg, log),
		redis:      redisClient,
		logger:     log,
		config:     cfg,
	}
}

func NewJWTService(cfg *config.AuthConfig, log *logger.Logger) *JWTService {
	return &JWTService{
		config:    cfg,
		logger:    log,
		jwtSecret: []byte(cfg.JWT_SECRET),
		jwtParser: &jwt.Parser{},
	}
}

func (s *JWTService) GenerateAccessToken(user *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(s.config.AccessTokenExpiry)

	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    s.config.JWT_ISSUER,
			Audience:  jwt.ClaimStrings{user.TenantID.String()},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			ID:        uuid.New().String(),
		},
		UserID:      user.ID.String(),
		TenantID:    user.TenantID.String(),
		Email:       user.Email,
		Role:        user.Role,
		TenantRole:  user.TenantRole,
		Permissions: user.Permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (s *JWTService) GenerateRefreshToken(userID, tenantID string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(s.config.RefreshTokenExpiry)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    s.config.JWT_ISSUER,
		Audience:  jwt.ClaimStrings{tenantID},
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := s.jwtParser.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *TokenService) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
	accessToken, accessExpiresAt, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.jwtService.GenerateRefreshToken(user.ID.String(), user.TenantID.String())
	if err != nil {
		return nil, err
	}

	refreshKey := fmt.Sprintf("refresh:%s", user.ID.String())
	if err := s.redis.Set(context.Background(), refreshKey, refreshToken, s.config.RefreshTokenExpiry); err != nil {
		s.logger.Error("Failed to store refresh token", "error", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.AccessTokenExpiry.Seconds()),
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func (s *TokenService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, apperr.Unauthorized("invalid refresh token")
	}

	refreshKey := fmt.Sprintf("refresh:%s", claims.UserID)
	storedToken, err := s.redis.Get(ctx, refreshKey)
	if err != nil {
		return nil, apperr.Unauthorized("refresh token not found or expired")
	}

	if storedToken != refreshToken {
		return nil, apperr.Unauthorized("refresh token has been revoked")
	}

	user, err := s.GetUserByID(ctx, claims.UserID, claims.TenantID)
	if err != nil {
		return nil, err
	}

	if user.Status != domain.UserStatusActive {
		return nil, apperr.Unauthorized("user account is not active")
	}

	return s.GenerateTokenPair(user)
}

func (s *TokenService) RevokeRefreshToken(ctx context.Context, userID string) error {
	refreshKey := fmt.Sprintf("refresh:%s", userID)
	return s.redis.Del(ctx, refreshKey)
}

func (s *TokenService) RevokeAllTokens(ctx context.Context, userID string) error {
	accessKey := fmt.Sprintf("access:blacklist:%s", userID)
	refreshKey := fmt.Sprintf("refresh:%s", userID)
	return s.redis.Del(ctx, accessKey, refreshKey)
}

func (s *TokenService) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("access:blacklist:%s", tokenID)
	result, err := s.redis.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

func (s *TokenService) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("access:blacklist:%s", tokenID)
	return s.redis.Set(ctx, key, "true", ttl)
}

func (s *TokenService) GetUserByID(ctx context.Context, userID, tenantID string) (*domain.User, error) {
	return nil, fmt.Errorf("not implemented")
}

type Session struct {
	SessionID   string    `json:"sessionId"`
	UserID      string    `json:"userId"`
	TenantID    string    `json:"tenantId"`
	AccessToken string    `json:"accessToken"`
	IPAddress   string    `json:"ipAddress"`
	UserAgent   string    `json:"userAgent"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type SessionService struct {
	redis      RedisClient
	logger     *logger.Logger
	sessionTTL time.Duration
}

func NewSessionService(redisClient RedisClient, log *logger.Logger, sessionTTL time.Duration) *SessionService {
	return &SessionService{
		redis:      redisClient,
		logger:     log,
		sessionTTL: sessionTTL,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, userID, tenantID, accessToken, ipAddress, userAgent string) (*Session, error) {
	sessionID := generateSecureToken(32)
	expiresAt := time.Now().UTC().Add(s.sessionTTL)

	session := &Session{
		SessionID:   sessionID,
		UserID:      userID,
		TenantID:    tenantID,
		AccessToken: accessToken,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", sessionID)
	if err := s.redis.Set(ctx, key, string(data), s.sessionTTL); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return s.redis.Del(ctx, key)
}

func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		s.DeleteSession(ctx, sessionID)
		return nil, apperr.Unauthorized("session expired")
	}

	return session, nil
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("authorization header must be 'Bearer <token>'")
	}

	return parts[1], nil
}
