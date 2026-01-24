package rbac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/errors"
	"github.com/ims-erp/system/pkg/logger"
)

type Role string

const (
	RoleSuperAdmin  Role = "super_admin"
	RoleTenantAdmin Role = "tenant_admin"
	RoleModuleAdmin Role = "module_admin"
	RoleUserManager Role = "user_manager"
	RoleUser        Role = "user"
	RoleViewer      Role = "viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleTenantAdmin, RoleModuleAdmin, RoleUserManager, RoleUser, RoleViewer:
		return true
	}
	return false
}

type Permission struct {
	ID          string   `json:"id" bson:"_id"`
	Name        string   `json:"name" bson:"name"`
	DisplayName string   `json:"displayName" bson:"displayName"`
	Module      string   `json:"module" bson:"module"`
	Actions     []string `json:"actions" bson:"actions"`
	Description string   `json:"description" bson:"description"`
}

type RolePermission struct {
	RoleID      string   `json:"roleId" bson:"roleId"`
	Permissions []string `json:"permissions" bson:"permissions"`
	TenantID    string   `json:"tenantId" bson:"tenantId"`
	IsSystem    bool     `json:"isSystem" bson:"isSystem"`
}

type UserRole struct {
	ID        string     `json:"id" bson:"_id"`
	UserID    string     `json:"userId" bson:"userId"`
	Role      Role       `json:"role" bson:"role"`
	Scope     string     `json:"scope" bson:"scope"`   // "tenant", "module", "own"
	Module    string     `json:"module" bson:"module"` // for module-level roles
	TenantID  string     `json:"tenantId" bson:"tenantId"`
	GrantedBy string     `json:"grantedBy" bson:"grantedBy"`
	GrantedAt time.Time  `json:"grantedAt" bson:"grantedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty" bson:"expiresAt,omitempty"`
}

type RBACService struct {
	roleStore       RoleStore
	permissionStore PermissionStore
	userRoleStore   UserRoleStore
	logger          *logger.Logger
}

type RoleStore interface {
	CreateRole(ctx context.Context, role *RolePermission) error
	UpdateRole(ctx context.Context, role *RolePermission) error
	GetRole(ctx context.Context, roleID string) (*RolePermission, error)
	GetRoleByName(ctx context.Context, name string, tenantID string) (*RolePermission, error)
	ListRoles(ctx context.Context, tenantID string) ([]*RolePermission, error)
	DeleteRole(ctx context.Context, roleID string) error
}

type PermissionStore interface {
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)
	ListPermissions(ctx context.Context, module string) ([]*Permission, error)
	ListAllPermissions(ctx context.Context) ([]*Permission, error)
}

type UserRoleStore interface {
	AssignRole(ctx context.Context, userRole *UserRole) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*UserRole, error)
	GetUserEffectiveRoles(ctx context.Context, userID string) ([]*UserRole, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	HasPermission(ctx context.Context, userID, permission string) (bool, error)
}

func NewRBACService(
	roleStore RoleStore,
	permissionStore PermissionStore,
	userRoleStore UserRoleStore,
	log *logger.Logger,
) *RBACService {
	return &RBACService{
		roleStore:       roleStore,
		permissionStore: permissionStore,
		userRoleStore:   userRoleStore,
		logger:          log,
	}
}

func (s *RBACService) AssignRole(ctx context.Context, userID, roleName, scope, module, tenantID, grantedBy string) error {
	role, err := s.roleStore.GetRoleByName(ctx, roleName, tenantID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.NotFound("role not found: %s", roleName)
	}

	userRole := &UserRole{
		ID:        uuid.New().String(),
		UserID:    userID,
		Role:      Role(roleName),
		Scope:     scope,
		Module:    module,
		TenantID:  tenantID,
		GrantedBy: grantedBy,
		GrantedAt: time.Now().UTC(),
	}

	if err := s.userRoleStore.AssignRole(ctx, userRole); err != nil {
		return err
	}

	s.logger.Info("Role assigned",
		"user_id", userID,
		"role", roleName,
		"tenant_id", tenantID,
		"granted_by", grantedBy,
	)

	return nil
}

func (s *RBACService) RevokeRole(ctx context.Context, userID, roleID string) error {
	if err := s.userRoleStore.RevokeRole(ctx, userID, roleID); err != nil {
		return err
	}

	s.logger.Info("Role revoked",
		"user_id", userID,
		"role_id", roleID,
	)

	return nil
}

func (s *RBACService) GetUserRoles(ctx context.Context, userID string) ([]*UserRole, error) {
	return s.userRoleStore.GetUserRoles(ctx, userID)
}

func (s *RBACService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return s.userRoleStore.GetUserPermissions(ctx, userID)
}

func (s *RBACService) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	return s.userRoleStore.HasPermission(ctx, userID, permission)
}

func (s *RBACService) CreateRole(ctx context.Context, name, description string, permissions []string, tenantID string, isSystem bool) error {
	role := &RolePermission{
		RoleID:      uuid.New().String(),
		Permissions: permissions,
		TenantID:    tenantID,
		IsSystem:    isSystem,
	}

	return s.roleStore.CreateRole(ctx, role)
}

func (s *RBACService) UpdateRole(ctx context.Context, roleID string, permissions []string) error {
	role, err := s.roleStore.GetRole(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.NotFound("role not found")
	}

	role.Permissions = permissions
	return s.roleStore.UpdateRole(ctx, role)
}

func (s *RBACService) ListRoles(ctx context.Context, tenantID string) ([]*RolePermission, error) {
	return s.roleStore.ListRoles(ctx, tenantID)
}

func (s *RBACService) CreatePermission(ctx context.Context, name, displayName, module, description string, actions []string) error {
	permission := &Permission{
		ID:          uuid.New().String(),
		Name:        name,
		DisplayName: displayName,
		Module:      module,
		Actions:     actions,
		Description: description,
	}

	return s.permissionStore.CreatePermission(ctx, permission)
}

func (s *RBACService) ListPermissions(ctx context.Context, module string) ([]*Permission, error) {
	return s.permissionStore.ListPermissions(ctx, module)
}

func (s *RBACService) InitializeDefaultRoles(ctx context.Context, tenantID string) error {
	defaultRoles := []struct {
		name        string
		permissions []string
		description string
	}{
		{
			name:        "admin",
			permissions: []string{"*"},
			description: "Full admin access",
		},
		{
			name:        "user_manager",
			permissions: []string{"user:read", "user:write", "user:delete"},
			description: "Can manage users",
		},
		{
			name:        "viewer",
			permissions: []string{"*:read"},
			description: "Read-only access",
		},
	}

	for _, role := range defaultRoles {
		existing, err := s.roleStore.GetRoleByName(ctx, role.name, tenantID)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}

		if err := s.CreateRole(ctx, role.name, role.description, role.permissions, tenantID, true); err != nil {
			return err
		}
	}

	s.logger.Info("Default roles initialized", "tenant_id", tenantID)
	return nil
}

func (s *RBACService) InitializeDefaultPermissions(ctx context.Context) error {
	defaultPermissions := []Permission{
		{ID: uuid.New().String(), Name: "client:read", DisplayName: "Read Clients", Module: "client", Actions: []string{"read"}, Description: "View client information"},
		{ID: uuid.New().String(), Name: "client:write", DisplayName: "Write Clients", Module: "client", Actions: []string{"write"}, Description: "Create and update clients"},
		{ID: uuid.New().String(), Name: "client:delete", DisplayName: "Delete Clients", Module: "client", Actions: []string{"delete"}, Description: "Delete clients"},
		{ID: uuid.New().String(), Name: "invoice:read", DisplayName: "Read Invoices", Module: "invoice", Actions: []string{"read"}, Description: "View invoices"},
		{ID: uuid.New().String(), Name: "invoice:write", DisplayName: "Write Invoices", Module: "invoice", Actions: []string{"write"}, Description: "Create and update invoices"},
		{ID: uuid.New().String(), Name: "invoice:approve", DisplayName: "Approve Invoices", Module: "invoice", Actions: []string{"approve"}, Description: "Approve invoices"},
		{ID: uuid.New().String(), Name: "payment:process", DisplayName: "Process Payments", Module: "payment", Actions: []string{"process"}, Description: "Process payments"},
		{ID: uuid.New().String(), Name: "user:read", DisplayName: "Read Users", Module: "user", Actions: []string{"read"}, Description: "View users"},
		{ID: uuid.New().String(), Name: "user:write", DisplayName: "Write Users", Module: "user", Actions: []string{"write"}, Description: "Create and update users"},
		{ID: uuid.New().String(), Name: "user:delete", DisplayName: "Delete Users", Module: "user", Actions: []string{"delete"}, Description: "Delete users"},
	}

	for _, perm := range defaultPermissions {
		existing, err := s.permissionStore.GetPermissionByName(ctx, perm.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}

		if err := s.permissionStore.CreatePermission(ctx, &perm); err != nil {
			return err
		}
	}

	s.logger.Info("Default permissions initialized")
	return nil
}

func (s *RBACService) HasAccess(userPermissions []string, requiredPermission string) bool {
	for _, up := range userPermissions {
		if up == "*" {
			return true
		}
		if up == requiredPermission {
			return true
		}
		if isWildcardMatch(up, requiredPermission) {
			return true
		}
	}
	return false
}

func isWildcardMatch(pattern, permission string) bool {
	if pattern == "" {
		return false
	}

	parts := splitN(pattern, ":", 3)
	permParts := splitN(permission, ":", 3)

	if len(parts) != len(permParts) {
		return false
	}

	for i, part := range parts {
		if part == "*" || part == "#" {
			continue
		}
		if part != permParts[i] {
			return false
		}
	}

	return true
}

func splitN(s string, sep string, n int) []string {
	result := make([]string, 0, n)
	remaining := s
	for i := 0; i < n-1; i++ {
		idx := findIndex(remaining, sep)
		if idx == -1 {
			break
		}
		result = append(result, remaining[:idx])
		remaining = remaining[idx+len(sep):]
	}
	result = append(result, remaining)
	return result
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

type RBACRepository struct {
	roles       *repository.ReadModelStore
	permissions *repository.ReadModelStore
	userRoles   *repository.ReadModelStore
}

func NewRBACRepository(readModelStore *repository.ReadModelStore) *RBACRepository {
	return &RBACRepository{
		roles:       readModelStore,
		permissions: readModelStore,
		userRoles:   readModelStore,
	}
}

func (r *RBACRepository) CreateRole(ctx context.Context, role *RolePermission) error {
	return r.roles.Save(ctx, role)
}

func (r *RBACRepository) UpdateRole(ctx context.Context, role *RolePermission) error {
	filter := map[string]interface{}{"_id": role.RoleID}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"permissions": role.Permissions,
		},
	}
	return r.roles.Update(ctx, filter, update)
}

func (r *RBACRepository) GetRole(ctx context.Context, roleID string) (*RolePermission, error) {
	filter := map[string]interface{}{"_id": roleID}
	result, err := r.roles.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return mapToRole(result)
}

func (r *RBACRepository) GetRoleByName(ctx context.Context, name string, tenantID string) (*RolePermission, error) {
	filter := map[string]interface{}{"_id": name, "tenantId": tenantID}
	result, err := r.roles.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return mapToRole(result)
}

func (r *RBACRepository) ListRoles(ctx context.Context, tenantID string) ([]*RolePermission, error) {
	filter := map[string]interface{}{"tenantId": tenantID}
	results, err := r.roles.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	roles := make([]*RolePermission, 0, len(results))
	for _, result := range results {
		role, err := mapToRole(result)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *RBACRepository) DeleteRole(ctx context.Context, roleID string) error {
	filter := map[string]interface{}{"_id": roleID}
	return r.roles.Delete(ctx, filter)
}

func (r *RBACRepository) CreatePermission(ctx context.Context, permission *Permission) error {
	return r.permissions.Save(ctx, permission)
}

func (r *RBACRepository) GetPermission(ctx context.Context, permissionID string) (*Permission, error) {
	filter := map[string]interface{}{"_id": permissionID}
	result, err := r.permissions.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return mapToPermission(result)
}

func (r *RBACRepository) GetPermissionByName(ctx context.Context, name string) (*Permission, error) {
	filter := map[string]interface{}{"name": name}
	result, err := r.permissions.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return mapToPermission(result)
}

func (r *RBACRepository) ListPermissions(ctx context.Context, module string) ([]*Permission, error) {
	filter := map[string]interface{}{"module": module}
	results, err := r.permissions.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	permissions := make([]*Permission, 0, len(results))
	for _, result := range results {
		perm, err := mapToPermission(result)
		if err != nil {
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *RBACRepository) ListAllPermissions(ctx context.Context) ([]*Permission, error) {
	results, err := r.permissions.Find(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	permissions := make([]*Permission, 0, len(results))
	for _, result := range results {
		perm, err := mapToPermission(result)
		if err != nil {
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *RBACRepository) AssignRole(ctx context.Context, userRole *UserRole) error {
	return r.userRoles.Save(ctx, userRole)
}

func (r *RBACRepository) RevokeRole(ctx context.Context, userID, roleID string) error {
	filter := map[string]interface{}{"userId": userID, "_id": roleID}
	return r.userRoles.Delete(ctx, filter)
}

func (r *RBACRepository) GetUserRoles(ctx context.Context, userID string) ([]*UserRole, error) {
	filter := map[string]interface{}{"userId": userID}
	results, err := r.userRoles.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	userRoles := make([]*UserRole, 0, len(results))
	for _, result := range results {
		ur, err := mapToUserRole(result)
		if err != nil {
			continue
		}
		userRoles = append(userRoles, ur)
	}

	return userRoles, nil
}

func (r *RBACRepository) GetUserEffectiveRoles(ctx context.Context, userID string) ([]*UserRole, error) {
	return r.GetUserRoles(ctx, userID)
}

func (r *RBACRepository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionsSet := make(map[string]bool)
	for _, ur := range roles {
		role, err := r.GetRole(ctx, string(ur.Role))
		if err != nil {
			continue
		}
		if role != nil {
			for _, p := range role.Permissions {
				permissionsSet[p] = true
			}
		}
	}

	permissions := make([]string, 0, len(permissionsSet))
	for p := range permissionsSet {
		permissions = append(permissions, p)
	}

	return permissions, nil
}

func (r *RBACRepository) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	permissions, err := r.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == "*" || p == permission {
			return true, nil
		}
	}

	return false, nil
}

func mapToRole(data interface{}) (*RolePermission, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid role data")
	}

	role := &RolePermission{}

	if id, ok := m["_id"].(string); ok {
		role.RoleID = id
	}
	if permissions, ok := m["permissions"].([]interface{}); ok {
		role.Permissions = make([]string, len(permissions))
		for i, p := range permissions {
			if s, ok := p.(string); ok {
				role.Permissions[i] = s
			}
		}
	}
	if tenantID, ok := m["tenantId"].(string); ok {
		role.TenantID = tenantID
	}
	if isSystem, ok := m["isSystem"].(bool); ok {
		role.IsSystem = isSystem
	}

	return role, nil
}

func mapToPermission(data interface{}) (*Permission, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid permission data")
	}

	perm := &Permission{}

	if id, ok := m["_id"].(string); ok {
		perm.ID = id
	}
	if name, ok := m["name"].(string); ok {
		perm.Name = name
	}
	if displayName, ok := m["displayName"].(string); ok {
		perm.DisplayName = displayName
	}
	if module, ok := m["module"].(string); ok {
		perm.Module = module
	}
	if actions, ok := m["actions"].([]interface{}); ok {
		perm.Actions = make([]string, len(actions))
		for i, a := range actions {
			if s, ok := a.(string); ok {
				perm.Actions[i] = s
			}
		}
	}
	if description, ok := m["description"].(string); ok {
		perm.Description = description
	}

	return perm, nil
}

func mapToUserRole(data interface{}) (*UserRole, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user role data")
	}

	ur := &UserRole{}

	if id, ok := m["_id"].(string); ok {
		ur.ID = id
	}
	if userID, ok := m["userId"].(string); ok {
		ur.UserID = userID
	}
	if role, ok := m["role"].(string); ok {
		ur.Role = Role(role)
	}
	if scope, ok := m["scope"].(string); ok {
		ur.Scope = scope
	}
	if module, ok := m["module"].(string); ok {
		ur.Module = module
	}
	if tenantID, ok := m["tenantId"].(string); ok {
		ur.TenantID = tenantID
	}
	if grantedBy, ok := m["grantedBy"].(string); ok {
		ur.GrantedBy = grantedBy
	}
	if grantedAt, ok := m["grantedAt"].(time.Time); ok {
		ur.GrantedAt = grantedAt
	}

	return ur, nil
}
