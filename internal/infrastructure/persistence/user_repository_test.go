package persistence

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword")
	user.ID = uuid.New().String()
	user.Email = "test@example.com"
	user.RealName = "Test User"

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", found.Username)
	assert.Equal(t, "test@example.com", found.Email)

	found, err = repo.GetByUsername(ctx, "testuser")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)

	found, err = repo.GetByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)

	user.RealName = "Updated User"
	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	found, err = repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated User", found.RealName)

	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
}

func TestUserRepository_RealDB_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u1 := entity.NewUser("user1", "hash1")
	u1.ID = uuid.New().String()
	u1.Email = "user1@test.com"
	u2 := entity.NewUser("user2", "hash2")
	u2.ID = uuid.New().String()
	u2.Email = "user2@test.com"
	err := repo.Create(ctx, u1)
	require.NoError(t, err)
	err = repo.Create(ctx, u2)
	require.NoError(t, err)

	users, total, err := repo.List(ctx, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)

	active := entity.UserStatusActive
	users, total, err = repo.List(ctx, &active, 1, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(0))
}

func TestUserRepository_RealDB_ExistsByUsername(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("existuser", "hash")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	exists, err := repo.ExistsByUsername(ctx, "existuser")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByUsername(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_RealDB_ExistsByEmail(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("emailuser", "hash")
	user.ID = uuid.New().String()
	user.Email = "email@test.com"
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	exists, err := repo.ExistsByEmail(ctx, "email@test.com")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByEmail(ctx, "noemail@test.com")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_RealDB_UpdateLastLogin(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("loginuser", "hash")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	err = repo.UpdateLastLogin(ctx, user.ID)
	assert.NoError(t, err)
}

func TestUserRepository_RealDB_RoleAssignment(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	user := entity.NewUser("roleuser", "hash")
	user.ID = uuid.New().String()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	role := entity.NewRole("test_role", "Test Role")
	role.ID = uuid.New().String()
	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	err = userRepo.AssignRole(ctx, user.ID, role.ID)
	assert.NoError(t, err)

	found, err := userRepo.GetWithRoles(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Roles, 1)
	assert.Equal(t, "test_role", found.Roles[0].Code)

	err = userRepo.RemoveRole(ctx, user.ID, role.ID)
	assert.NoError(t, err)

	found, err = userRepo.GetWithRoles(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Roles, 0)
}

func TestUserRepository_RealDB_GetWithPermissions(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)
	ctx := context.Background()

	user := entity.NewUser("permuser", "hash")
	user.ID = uuid.New().String()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	role := entity.NewRole("perm_role", "Perm Role")
	role.ID = uuid.New().String()
	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	perm := entity.NewPermission("test:read", "Test Read")
	perm.ID = uuid.New().String()
	err = permRepo.Create(ctx, perm)
	require.NoError(t, err)

	err = userRepo.AssignRole(ctx, user.ID, role.ID)
	require.NoError(t, err)

	err = roleRepo.AssignPermission(ctx, role.ID, perm.ID)
	require.NoError(t, err)

	foundUser, permissions, err := userRepo.GetWithPermissions(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.GreaterOrEqual(t, len(permissions), 1)
}

func TestRoleRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := entity.NewRole("crud_role", "CRUD Role")
	role.ID = uuid.New().String()
	role.Description = "Test role"

	err := repo.Create(ctx, role)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, role.ID)
	assert.NoError(t, err)
	assert.Equal(t, "crud_role", found.Code)

	found, err = repo.GetByCode(ctx, "crud_role")
	assert.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)

	role.Name = "Updated Role"
	err = repo.Update(ctx, role)
	assert.NoError(t, err)

	roles, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 1)

	exists, err := repo.ExistsByCode(ctx, "crud_role")
	assert.NoError(t, err)
	assert.True(t, exists)

	err = repo.Delete(ctx, role.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, role.ID)
	assert.Error(t, err)
}

func TestRoleRepository_RealDB_PermissionAssignment(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)
	ctx := context.Background()

	role := entity.NewRole("perm_role_test", "Perm Role Test")
	role.ID = uuid.New().String()
	err := roleRepo.Create(ctx, role)
	require.NoError(t, err)

	perm := entity.NewPermission("test:write", "Test Write")
	perm.ID = uuid.New().String()
	err = permRepo.Create(ctx, perm)
	require.NoError(t, err)

	err = roleRepo.AssignPermission(ctx, role.ID, perm.ID)
	assert.NoError(t, err)

	found, err := roleRepo.GetWithPermissions(ctx, role.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Permissions, 1)

	err = roleRepo.RemovePermission(ctx, role.ID, perm.ID)
	assert.NoError(t, err)

	found, err = roleRepo.GetWithPermissions(ctx, role.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Permissions, 0)
}

func TestPermissionRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	perm := entity.NewPermission("perm:crud", "Perm CRUD")
	perm.ID = uuid.New().String()

	err := repo.Create(ctx, perm)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, perm.ID)
	assert.NoError(t, err)
	assert.Equal(t, "perm:crud", found.Code)

	found, err = repo.GetByCode(ctx, "perm:crud")
	assert.NoError(t, err)
	assert.Equal(t, perm.ID, found.ID)

	perms, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(perms), 1)

	exists, err := repo.ExistsByCode(ctx, "perm:crud")
	assert.NoError(t, err)
	assert.True(t, exists)

	err = repo.Delete(ctx, perm.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, perm.ID)
	assert.Error(t, err)
}

func TestPermissionRepository_RealDB_BatchCreate(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	p1 := entity.NewPermission("batch:1", "Batch 1")
	p1.ID = uuid.New().String()
	p2 := entity.NewPermission("batch:2", "Batch 2")
	p2.ID = uuid.New().String()

	err := repo.BatchCreate(ctx, []*entity.Permission{p1, p2})
	assert.NoError(t, err)

	found, err := repo.GetByCode(ctx, "batch:1")
	assert.NoError(t, err)
	assert.NotNil(t, found)
}

func TestPermissionRepository_RealDB_GetByRoleID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)
	ctx := context.Background()

	role := entity.NewRole("role_perm_test", "Role Perm Test")
	role.ID = uuid.New().String()
	err := roleRepo.Create(ctx, role)
	require.NoError(t, err)

	perm := entity.NewPermission("role:perm", "Role Perm")
	perm.ID = uuid.New().String()
	err = permRepo.Create(ctx, perm)
	require.NoError(t, err)

	err = roleRepo.AssignPermission(ctx, role.ID, perm.ID)
	require.NoError(t, err)

	perms, err := permRepo.GetByRoleID(ctx, role.ID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(perms), 1)
}

func TestPermissionRepository_RealDB_GetByUserID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)
	ctx := context.Background()

	user := entity.NewUser("uperm_user", "hash")
	user.ID = uuid.New().String()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	role := entity.NewRole("uperm_role", "User Perm Role")
	role.ID = uuid.New().String()
	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	perm := entity.NewPermission("user:perm", "User Perm")
	perm.ID = uuid.New().String()
	err = permRepo.Create(ctx, perm)
	require.NoError(t, err)

	err = userRepo.AssignRole(ctx, user.ID, role.ID)
	require.NoError(t, err)
	err = roleRepo.AssignPermission(ctx, role.ID, perm.ID)
	require.NoError(t, err)

	perms, err := permRepo.GetByUserID(ctx, user.ID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(perms), 1)
}

func TestPermissionRepository_RealDB_ListWithResourceType(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	perm := entity.NewPermission("res:perm", "Resource Perm")
	perm.ID = uuid.New().String()
	perm.ResourceType = "alarm"
	err := repo.Create(ctx, perm)
	require.NoError(t, err)

	resType := "alarm"
	perms, err := repo.List(ctx, &resType)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(perms), 1)
}
