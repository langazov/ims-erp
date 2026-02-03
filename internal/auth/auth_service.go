package auth

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/domain"
	"github.com/ims-erp/system/internal/repository"
	apperr "github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	userStore      UserStore
	tokenService   *TokenService
	sessionService *SessionService
	rateLimiter    RateLimiter
	logger         *logger.Logger
	config         *config.AuthConfig
}

type UserStore interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email, tenantID string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByTenant(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.User, int64, error)
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
}

type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

type LoginRequest struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginResponse struct {
	User      *domain.User `json:"user"`
	Tokens    *TokenPair   `json:"tokens"`
	SessionID string       `json:"sessionId"`
}

func NewAuthService(
	userStore UserStore,
	tokenService *TokenService,
	sessionService *SessionService,
	rateLimiter RateLimiter,
	cfg *config.AuthConfig,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userStore:      userStore,
		tokenService:   tokenService,
		sessionService: sessionService,
		rateLimiter:    rateLimiter,
		logger:         log,
		config:         cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, tenantID, requestID string, req *RegisterRequest) (*domain.User, error) {
	if err := s.validatePassword(req.Password); err != nil {
		return nil, apperr.InvalidArgument("invalid password: %s", err)
	}

	existing, err := s.userStore.FindByEmail(ctx, req.Email, tenantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperr.AlreadyExists("email already registered")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, apperr.InvalidArgument("invalid tenant ID")
	}

	user, err := domain.NewUser(tenantUUID, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, apperr.InternalError("failed to create user")
	}

	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.userStore.Create(ctx, user); err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternalError, "failed to create user")
	}

	s.logger.Info("User registered",
		"user_id", user.ID.String(),
		"tenant_id", tenantID,
		"email", user.Email,
	)

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, tenantID, requestID string, req *LoginRequest) (*LoginResponse, error) {
	allowed, current, err := s.rateLimiter.Allow(ctx, fmt.Sprintf("login:%s:%s", tenantID, req.Email), s.config.MaxLoginAttempts, s.config.LockoutDuration)
	if err != nil {
		s.logger.Error("Rate limiter error", "error", err)
	}

	if !allowed {
		return nil, apperr.TooManyRequests("too many login attempts. try again later. current attempts: %d", current)
	}

	user, err := s.userStore.FindByEmail(ctx, req.Email, tenantID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		s.logger.Warn("Login attempt for non-existent user",
			"email", req.Email,
			"tenant_id", tenantID,
			"ip", req.IPAddress,
		)
		return nil, apperr.Unauthorized("invalid email or password")
	}

	if user.IsLocked() {
		return nil, apperr.Forbidden("account is locked. try again later")
	}

	if user.Status != domain.UserStatusActive {
		return nil, apperr.Forbidden("account is not active")
	}

	if !user.VerifyPassword(req.Password) {
		user.AddLoginAttempt()
		if user.LoginAttempts >= s.config.MaxLoginAttempts {
			user.Lock(s.config.LockoutDuration)
			s.logger.Warn("User account locked due to too many failed attempts",
				"user_id", user.ID.String(),
				"tenant_id", tenantID,
				"ip", req.IPAddress,
			)
		}
		s.userStore.Update(ctx, user)
		return nil, apperr.Unauthorized("invalid email or password")
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(user)
	if err != nil {
		return nil, apperr.InternalError("failed to generate tokens")
	}

	session, err := s.sessionService.CreateSession(ctx, user.ID.String(), tenantID, tokenPair.AccessToken, req.IPAddress, req.UserAgent)
	if err != nil {
		s.logger.Error("Failed to create session", "error", err)
	}

	user.RecordLogin()
	s.userStore.Update(ctx, user)

	s.logger.Info("User logged in",
		"user_id", user.ID.String(),
		"tenant_id", tenantID,
		"email", user.Email,
		"ip", req.IPAddress,
	)

	return &LoginResponse{
		User:      user,
		Tokens:    tokenPair,
		SessionID: session.SessionID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID, sessionID string) error {
	if err := s.sessionService.DeleteSession(ctx, sessionID); err != nil {
		s.logger.Error("Failed to delete session", "error", err)
	}

	if err := s.tokenService.RevokeRefreshToken(ctx, userID); err != nil {
		s.logger.Error("Failed to revoke refresh token", "error", err)
	}

	s.logger.Info("User logged out", "user_id", userID)
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	if err := s.tokenService.RevokeAllTokens(ctx, userID); err != nil {
		s.logger.Error("Failed to revoke all tokens", "error", err)
	}

	s.logger.Info("User logged out from all devices", "user_id", userID)
	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	return s.tokenService.RefreshTokens(ctx, refreshToken)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.NotFound("user not found")
	}

	if !user.VerifyPassword(currentPassword) {
		return apperr.InvalidArgument("current password is incorrect")
	}

	if err := s.validatePassword(newPassword); err != nil {
		return apperr.InvalidArgument("invalid new password: %s", err)
	}

	if err := user.SetPassword(newPassword); err != nil {
		return apperr.InternalError("failed to set new password")
	}

	if err := s.tokenService.RevokeAllTokens(ctx, userID); err != nil {
		s.logger.Error("Failed to revoke tokens after password change", "error", err)
	}

	if err := s.userStore.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("Password changed", "user_id", userID)
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, tenantID string) error {
	user, err := s.userStore.FindByEmail(ctx, email, tenantID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.NotFound("user not found")
	}

	newPassword := generateRandomPassword(12)
	if err := user.SetPassword(newPassword); err != nil {
		return apperr.InternalError("failed to reset password")
	}

	if err := s.userStore.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("Password reset", "user_id", user.ID.String(), "tenant_id", tenantID)
	return nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userStore.FindByID(ctx, userID)
}

func (s *AuthService) ListUsers(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.User, int64, error) {
	return s.userStore.FindByTenant(ctx, tenantID, page, pageSize)
}

func (s *AuthService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.userStore.Update(ctx, user)
}

func (s *AuthService) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.NotFound("user not found")
	}

	user.Deactivate()
	if err := s.userStore.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("User deactivated", "user_id", userID)
	return nil
}

func (s *AuthService) ActivateUser(ctx context.Context, userID string) error {
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.NotFound("user not found")
	}

	user.Reactivate()
	if err := s.userStore.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("User activated", "user_id", userID)
	return nil
}

func (s *AuthService) LockUser(ctx context.Context, userID string) error {
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.NotFound("user not found")
	}

	user.Lock(s.config.LockoutDuration)
	if err := s.userStore.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("User locked", "user_id", userID)
	return nil
}

func (s *AuthService) UnlockUser(ctx context.Context, userID string) error {
	user, err := s.userStore.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.NotFound("user not found")
	}

	user.Unlock()
	if err := s.userStore.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("User unlocked", "user_id", userID)
	return nil
}

func (s *AuthService) validatePassword(password string) error {
	if len(password) < s.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", s.config.PasswordMinLength)
	}

	if s.config.PasswordRequireUpper {
		if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	if s.config.PasswordRequireLower {
		if !regexp.MustCompile(`[a-z]`).MatchString(password) {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
	}

	if s.config.PasswordRequireNumber {
		if !regexp.MustCompile(`[0-9]`).MatchString(password) {
			return fmt.Errorf("password must contain at least one number")
		}
	}

	if s.config.PasswordRequireSpecial {
		if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

type UserRepository struct {
	collection *repository.ReadModelStore
}

func NewUserRepository(readModelStore *repository.ReadModelStore) *UserRepository {
	return &UserRepository{
		collection: readModelStore,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.collection.Save(ctx, user)
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	filter := map[string]interface{}{"_id": user.ID.String()}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"email":         user.Email,
			"firstName":     user.FirstName,
			"lastName":      user.LastName,
			"phone":         user.Phone,
			"role":          user.Role,
			"status":        string(user.Status),
			"tenantRole":    user.TenantRole,
			"permissions":   user.Permissions,
			"mfaEnabled":    user.MFAEnabled,
			"lastLoginAt":   user.LastLoginAt,
			"loginAttempts": user.LoginAttempts,
			"lockedUntil":   user.LockedUntil,
			"updatedAt":     user.UpdatedAt,
		},
	}
	if user.PasswordHash != "" {
		update["$set"].(map[string]interface{})["passwordHash"] = user.PasswordHash
	}
	return r.collection.Update(ctx, filter, update)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email, tenantID string) (*domain.User, error) {
	filter := map[string]interface{}{
		"email":    email,
		"tenantid": tenantID,
	}
	result, err := r.collection.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	userData, ok := result.(bson.M)
	if !ok {
		return nil, fmt.Errorf("invalid user data: got %T", result)
	}
	return mapToUser(userData)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	filter := map[string]interface{}{"_id": id}
	result, err := r.collection.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	userData, ok := result.(bson.M)
	if !ok {
		return nil, fmt.Errorf("invalid user data")
	}
	return mapToUser(userData)
}

func (r *UserRepository) FindByTenant(ctx context.Context, tenantID string, page, pageSize int) ([]*domain.User, int64, error) {
	filter := map[string]interface{}{"tenantid": tenantID}
	results, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, 0, len(results))
	for _, result := range results {
		userData, ok := result.(bson.M)
		if !ok {
			continue
		}
		user, err := mapToUser(userData)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, int64(len(users)), nil
}

func mapToUser(data bson.M) (*domain.User, error) {
	user := &domain.User{}

	if idStr, ok := data["id"].(string); ok {
		user.ID = uuid.MustParse(idStr)
	} else if idBinary, ok := data["id"].(primitive.Binary); ok {
		if idBinary.Subtype == 0x03 || len(idBinary.Data) == 16 {
			if u, err := uuid.FromBytes(idBinary.Data); err == nil {
				user.ID = u
			}
		}
	}

	if tenantIDStr, ok := data["tenantid"].(string); ok {
		user.TenantID = uuid.MustParse(tenantIDStr)
	} else if tenantBinary, ok := data["tenantid"].(primitive.Binary); ok {
		if tenantBinary.Subtype == 0x03 || len(tenantBinary.Data) == 16 {
			if u, err := uuid.FromBytes(tenantBinary.Data); err == nil {
				user.TenantID = u
			}
		}
	}

	if email, ok := data["email"].(string); ok {
		user.Email = email
	}
	if passwordHash, ok := data["passwordhash"].(string); ok {
		user.PasswordHash = passwordHash
	}
	if firstName, ok := data["firstname"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := data["lastname"].(string); ok {
		user.LastName = lastName
	}
	if phone, ok := data["phone"].(string); ok {
		user.Phone = phone
	}
	if role, ok := data["role"].(string); ok {
		user.Role = role
	}
	if status, ok := data["status"].(string); ok {
		user.Status = domain.UserStatus(status)
	}
	if tenantRole, ok := data["tenantrole"].(string); ok {
		user.TenantRole = tenantRole
	}
	if permissions, ok := data["permissions"].([]interface{}); ok {
		user.Permissions = make([]string, len(permissions))
		for i, p := range permissions {
			if s, ok := p.(string); ok {
				user.Permissions[i] = s
			}
		}
	}
	if mfaEnabled, ok := data["mfaenabled"].(bool); ok {
		user.MFAEnabled = mfaEnabled
	}
	if loginAttempts, ok := data["loginattempts"].(int64); ok {
		user.LoginAttempts = int(loginAttempts)
	}

	return user, nil
}
