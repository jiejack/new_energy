package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, status *entity.UserStatus, page, pageSize int) ([]*entity.User, int64, error)
	GetWithRoles(ctx context.Context, id string) (*entity.User, error)
	GetWithPermissions(ctx context.Context, id string) (*entity.User, []*entity.Permission, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error
	UpdateLastLogin(ctx context.Context, id string) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Role, error)
	GetByCode(ctx context.Context, code string) (*entity.Role, error)
	List(ctx context.Context) ([]*entity.Role, error)
	GetWithPermissions(ctx context.Context, id string) (*entity.Role, error)
	AssignPermission(ctx context.Context, roleID, permissionID string) error
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	ExistsByCode(ctx context.Context, code string) (bool, error)
}

type PermissionRepository interface {
	Create(ctx context.Context, permission *entity.Permission) error
	BatchCreate(ctx context.Context, permissions []*entity.Permission) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.Permission, error)
	GetByCode(ctx context.Context, code string) (*entity.Permission, error)
	List(ctx context.Context, resourceType *string) ([]*entity.Permission, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error)
	GetByUserID(ctx context.Context, userID string) ([]*entity.Permission, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
}


