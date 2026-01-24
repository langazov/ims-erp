package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusLocked    UserStatus = "locked"
	UserStatusPending   UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusLocked, UserStatusPending, UserStatusSuspended:
		return true
	}
	return false
}

type User struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Email         string
	PasswordHash  string
	FirstName     string
	LastName      string
	Phone         string
	Role          string
	Status        UserStatus
	TenantRole    string
	Permissions   []string
	MFAEnabled    bool
	MFASecret     string
	LastLoginAt   *time.Time
	LoginAttempts int
	LockedUntil   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewUser(tenantID uuid.UUID, email, password, firstName, lastName string) (*User, error) {
	now := time.Now().UTC()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         email,
		PasswordHash:  string(hash),
		FirstName:     firstName,
		LastName:      lastName,
		Status:        UserStatusActive,
		TenantRole:    "user",
		Permissions:   []string{},
		MFAEnabled:    false,
		LoginAttempts: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) UpdateProfile(firstName, lastName, phone string) {
	u.FirstName = firstName
	u.LastName = lastName
	u.Phone = phone
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) AddLoginAttempt() {
	u.LoginAttempts++
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) ResetLoginAttempts() {
	u.LoginAttempts = 0
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Lock(lockDuration time.Duration) {
	u.Status = UserStatusLocked
	lockedUntil := time.Now().UTC().Add(lockDuration)
	u.LockedUntil = &lockedUntil
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Unlock() {
	u.Status = UserStatusActive
	u.LockedUntil = nil
	u.LoginAttempts = 0
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked && u.LockedUntil != nil {
		return u.LockedUntil.After(time.Now().UTC())
	}
	return u.Status == UserStatusLocked
}

func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	u.LoginAttempts = 0
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) SetRole(role string) {
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) SetTenantRole(role string) {
	u.TenantRole = role
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) AddPermission(permission string) {
	for _, p := range u.Permissions {
		if p == permission {
			return
		}
	}
	u.Permissions = append(u.Permissions, permission)
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) RemovePermission(permission string) {
	for i, p := range u.Permissions {
		if p == permission {
			u.Permissions = append(u.Permissions[:i], u.Permissions[i+1:]...)
			u.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

func (u *User) EnableMFA(secret string) {
	u.MFAEnabled = true
	u.MFASecret = secret
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) DisableMFA() {
	u.MFAEnabled = false
	u.MFASecret = ""
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Reactivate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now().UTC()
}
