package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"valid active", UserStatusActive, true},
		{"valid inactive", UserStatusInactive, true},
		{"valid locked", UserStatusLocked, true},
		{"valid pending", UserStatusPending, true},
		{"valid suspended", UserStatusSuspended, true},
		{"invalid", UserStatus("invalid"), false},
		{"empty", UserStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewUser(t *testing.T) {
	tenantID := uuid.New()
	email := "test@example.com"
	password := "password123"
	firstName := "John"
	lastName := "Doe"

	user, err := NewUser(tenantID, email, password, firstName, lastName)

	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, tenantID, user.TenantID)
	assert.Equal(t, email, user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	assert.Equal(t, firstName, user.FirstName)
	assert.Equal(t, lastName, user.LastName)
	assert.Equal(t, UserStatusActive, user.Status)
	assert.Equal(t, "user", user.TenantRole)
	assert.Empty(t, user.Permissions)
	assert.False(t, user.MFAEnabled)
	assert.Equal(t, 0, user.LoginAttempts)
}

func TestUserVerifyPassword(t *testing.T) {
	password := "mysecurepassword"
	user, _ := NewUser(uuid.New(), "test@example.com", password, "John", "Doe")

	valid := user.VerifyPassword(password)
	assert.True(t, valid)

	invalid := user.VerifyPassword("wrongpassword")
	assert.False(t, invalid)
}

func TestUserSetPassword(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "oldpassword", "John", "Doe")
	oldHash := user.PasswordHash

	err := user.SetPassword("newpassword")
	assert.NoError(t, err)
	assert.NotEqual(t, oldHash, user.PasswordHash)
	assert.True(t, user.VerifyPassword("newpassword"))
}

func TestUserUpdateProfile(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.UpdateProfile("Jane", "Smith", "555-1234")

	assert.Equal(t, "Jane", user.FirstName)
	assert.Equal(t, "Smith", user.LastName)
	assert.Equal(t, "555-1234", user.Phone)
}

func TestUserAddLoginAttempt(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	initial := user.LoginAttempts

	user.AddLoginAttempt()

	assert.Equal(t, initial+1, user.LoginAttempts)
}

func TestUserResetLoginAttempts(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.LoginAttempts = 5

	user.ResetLoginAttempts()

	assert.Equal(t, 0, user.LoginAttempts)
}

func TestUserLock(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.Lock(30 * time.Minute)

	assert.Equal(t, UserStatusLocked, user.Status)
	assert.NotNil(t, user.LockedUntil)
	assert.True(t, user.LockedUntil.After(time.Now()))
}

func TestUserUnlock(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.Lock(30 * time.Minute)
	user.LoginAttempts = 3

	user.Unlock()

	assert.Equal(t, UserStatusActive, user.Status)
	assert.Nil(t, user.LockedUntil)
	assert.Equal(t, 0, user.LoginAttempts)
}

func TestUserIsLocked(t *testing.T) {
	t.Run("not locked", func(t *testing.T) {
		user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
		assert.False(t, user.IsLocked())
	})

	t.Run("locked but expired", func(t *testing.T) {
		user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
		pastTime := time.Now().Add(-1 * time.Hour)
		user.Status = UserStatusLocked
		user.LockedUntil = &pastTime
		assert.False(t, user.IsLocked())
	})

	t.Run("locked and current", func(t *testing.T) {
		user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
		user.Lock(30 * time.Minute)
		assert.True(t, user.IsLocked())
	})
}

func TestUserRecordLogin(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.LoginAttempts = 3

	user.RecordLogin()

	assert.NotNil(t, user.LastLoginAt)
	assert.Equal(t, 0, user.LoginAttempts)
}

func TestUserFullName(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	assert.Equal(t, "John Doe", user.FullName())
}

func TestUserSetRole(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.SetRole("admin")

	assert.Equal(t, "admin", user.Role)
}

func TestUserSetTenantRole(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.SetTenantRole("manager")

	assert.Equal(t, "manager", user.TenantRole)
}

func TestUserAddPermission(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.AddPermission("read:users")
	user.AddPermission("write:users")
	user.AddPermission("read:users")

	assert.Len(t, user.Permissions, 2)
	assert.Contains(t, user.Permissions, "read:users")
	assert.Contains(t, user.Permissions, "write:users")
}

func TestUserRemovePermission(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.Permissions = []string{"read:users", "write:users", "delete:users"}

	user.RemovePermission("write:users")

	assert.Len(t, user.Permissions, 2)
	assert.Contains(t, user.Permissions, "read:users")
	assert.Contains(t, user.Permissions, "delete:users")
	assert.NotContains(t, user.Permissions, "write:users")
}

func TestUserRemovePermissionNotFound(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.Permissions = []string{"read:users"}

	user.RemovePermission("write:users")

	assert.Len(t, user.Permissions, 1)
}

func TestUserEnableMFA(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.EnableMFA("secret123")

	assert.True(t, user.MFAEnabled)
	assert.Equal(t, "secret123", user.MFASecret)
}

func TestUserDisableMFA(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.EnableMFA("secret123")

	user.DisableMFA()

	assert.False(t, user.MFAEnabled)
	assert.Empty(t, user.MFASecret)
}

func TestUserDeactivate(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.Deactivate()

	assert.Equal(t, UserStatusInactive, user.Status)
}

func TestUserReactivate(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")
	user.Status = UserStatusInactive

	user.Reactivate()

	assert.Equal(t, UserStatusActive, user.Status)
}

func TestUserSuspend(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password", "John", "Doe")

	user.Suspend()

	assert.Equal(t, UserStatusSuspended, user.Status)
}

func TestUserLifecycle(t *testing.T) {
	tenantID := uuid.New()
	user, err := NewUser(tenantID, "user@example.com", "password", "John", "Doe")
	assert.NoError(t, err)

	assert.Equal(t, UserStatusActive, user.Status)
	assert.Equal(t, 0, user.LoginAttempts)

	user.AddLoginAttempt()
	user.AddLoginAttempt()
	assert.Equal(t, 2, user.LoginAttempts)

	user.RecordLogin()
	assert.NotNil(t, user.LastLoginAt)
	assert.Equal(t, 0, user.LoginAttempts)

	user.Lock(1 * time.Hour)
	assert.True(t, user.IsLocked())

	user.Unlock()
	assert.False(t, user.IsLocked())

	user.SetRole("admin")
	user.AddPermission("read:all")
	user.AddPermission("write:all")
	assert.Len(t, user.Permissions, 2)

	user.Suspend()
	assert.Equal(t, UserStatusSuspended, user.Status)

	user.Reactivate()
	assert.Equal(t, UserStatusActive, user.Status)
}
