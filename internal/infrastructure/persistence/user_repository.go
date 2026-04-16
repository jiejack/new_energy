package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) GetWithRoles(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, nil, err
	}

	permissionMap := make(map[string]*entity.Permission)
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permissionMap[perm.ID] = perm
		}
	}

	permissions := make([]*entity.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return &user, permissions, nil
}

func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	userRole := &entity.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	return r.db.WithContext(ctx).
		Delete(&entity.UserRole{}, "user_id = ? AND role_id = ?", userID, roleID).Error
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login":  &now,
			"login_count": gorm.Expr("login_count + 1"),
		}).Error
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

type RoleRepository struct {
	db *Database
}

func NewRoleRepository(db *Database) repository.RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RoleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, "id = ?", id).Error
}

func (r *RoleRepository) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) GetWithPermissions(ctx context.Context, id string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	rolePermission := &entity.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).
		Delete(&entity.RolePermission{}, "role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

func (r *RoleRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Role{}).
		Where("code = ?", code).
		Count(&count).Error
	return count > 0, err
}

type PermissionRepository struct {
	db *Database
}

func NewPermissionRepository(db *Database) repository.PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *PermissionRepository) BatchCreate(ctx context.Context, permissions []*entity.Permission) error {
	return r.db.WithContext(ctx).CreateInBatches(permissions, 100).Error
}

func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, "id = ?", id).Error
}

func (r *PermissionRepository) GetByID(ctx context.Context, id string) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepository) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).First(&permission, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepository) List(ctx context.Context, resourceType *string) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	query := r.db.WithContext(ctx)

	if resourceType != nil {
		query = query.Where("resource_type = ?", *resourceType)
	}

	err := query.Order("created_at ASC").Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) GetByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().
		Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Permission{}).
		Where("code = ?", code).
		Count(&count).Error
	return count > 0, err
}


