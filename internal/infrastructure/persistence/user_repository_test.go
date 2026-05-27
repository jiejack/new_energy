package persistence

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "testuser", found.Username)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	found, err := repo.GetByUsername(ctx, "testuser")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByUsername(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	user.SetEmail("test@example.com")
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	found, err := repo.GetByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	user.SetRealName("Test User")
	user.SetPhone("13800138000")
	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test User", found.RealName)
	assert.Equal(t, "13800138000", found.Phone)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u1 := entity.NewUser("user1", "pass1")
	u1.ID = uuid.New().String()
	u1.SetEmail("user1@test.com")
	u2 := entity.NewUser("user2", "pass2")
	u2.ID = uuid.New().String()
	u2.SetEmail("user2@test.com")
	err := repo.Create(ctx, u1)
	require.NoError(t, err)
	err = repo.Create(ctx, u2)
	require.NoError(t, err)
	err = db.DB.Exec("UPDATE users SET status = 0 WHERE id = ?", u2.ID).Error
	require.NoError(t, err)

	users, total, err := repo.List(ctx, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)

	activeStatus := entity.UserStatusActive
	users, total, err = repo.List(ctx, &activeStatus, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Username)
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	exists, err := repo.ExistsByUsername(ctx, "testuser")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByUsername(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	user.SetEmail("test@example.com")
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	exists, err := repo.ExistsByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByEmail(ctx, "nonexistent@example.com")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	err = repo.UpdateLastLogin(ctx, user.ID)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.LastLogin)
	assert.Equal(t, 1, found.LoginCount)
}

func TestUserRepository_AssignAndRemoveRole(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	userRepo := NewUserRepository(db)
	roleRepo := NewRoleRepository(db)
	ctx := context.Background()

	user := entity.NewUser("testuser", "hashedpassword123")
	user.ID = uuid.New().String()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	role := entity.NewRole("admin", "Admin")
	role.ID = uuid.New().String()
	err = roleRepo.Create(ctx, role)
	require.NoError(t, err)

	err = userRepo.AssignRole(ctx, user.ID, role.ID)
	assert.NoError(t, err)

	found, err := userRepo.GetWithRoles(ctx, user.ID)
	assert.NoError(t, err)
	if assert.Len(t, found.Roles, 1) {
		assert.Equal(t, "admin", found.Roles[0].Code)
	}

	err = userRepo.RemoveRole(ctx, user.ID, role.ID)
	assert.NoError(t, err)

	found, err = userRepo.GetWithRoles(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Roles, 0)
}

func TestRoleRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := entity.NewRole("operator", "Operator")
	role.ID = uuid.New().String()
	err := repo.Create(ctx, role)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, role.ID)
	assert.NoError(t, err)
	assert.Equal(t, "operator", found.Code)

	found, err = repo.GetByCode(ctx, "operator")
	assert.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)

	role.SetDescription("Updated description")
	err = repo.Update(ctx, role)
	assert.NoError(t, err)

	roles, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 1)

	exists, err := repo.ExistsByCode(ctx, "operator")
	assert.NoError(t, err)
	assert.True(t, exists)

	err = repo.Delete(ctx, role.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, role.ID)
	assert.Error(t, err)
}

func TestPermissionRepository_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	perm := entity.NewPermission("device:read", "Read Device")
	perm.ID = uuid.New().String()
	err := repo.Create(ctx, perm)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, perm.ID)
	assert.NoError(t, err)
	assert.Equal(t, "device:read", found.Code)

	found, err = repo.GetByCode(ctx, "device:read")
	assert.NoError(t, err)
	assert.Equal(t, perm.ID, found.ID)

	perms, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(perms), 1)

	exists, err := repo.ExistsByCode(ctx, "device:read")
	assert.NoError(t, err)
	assert.True(t, exists)

	err = repo.Delete(ctx, perm.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, perm.ID)
	assert.Error(t, err)
}

func TestPermissionRepository_BatchCreate(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	perms := []*entity.Permission{
		entity.NewPermission("station:read", "Read Station"),
		entity.NewPermission("station:create", "Create Station"),
		entity.NewPermission("station:update", "Update Station"),
	}
	for _, p := range perms {
		p.ID = uuid.New().String()
	}
	err := repo.BatchCreate(ctx, perms)
	assert.NoError(t, err)

	found, err := repo.List(ctx, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 3)
}

func TestPermissionRepository_ListByResourceType(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	p1 := entity.NewPermission("station:read", "Read Station")
	p1.ID = uuid.New().String()
	p1.SetResource("station", "")
	p2 := entity.NewPermission("device:read", "Read Device")
	p2.ID = uuid.New().String()
	p2.SetResource("device", "")
	err := repo.Create(ctx, p1)
	require.NoError(t, err)
	err = repo.Create(ctx, p2)
	require.NoError(t, err)

	rt := "station"
	found, err := repo.List(ctx, &rt)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, "station:read", found[0].Code)
}

func TestRoleRepository_AssignAndRemovePermission(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)
	ctx := context.Background()

	role := entity.NewRole("admin", "Admin")
	role.ID = uuid.New().String()
	err := roleRepo.Create(ctx, role)
	require.NoError(t, err)

	perm := entity.NewPermission("device:read", "Read Device")
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
