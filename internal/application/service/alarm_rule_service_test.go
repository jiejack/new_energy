package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAlarmRuleRepo struct {
	mock.Mock
}

func (m *mockAlarmRuleRepo) Create(ctx context.Context, rule *entity.AlarmRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockAlarmRuleRepo) Update(ctx context.Context, rule *entity.AlarmRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockAlarmRuleRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAlarmRuleRepo) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepo) GetByName(ctx context.Context, name string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepo) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Get(1).(int64), args.Error(2)
}

func (m *mockAlarmRuleRepo) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepo) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, pointID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepo) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *mockAlarmRuleRepo) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func TestAlarmRuleService_CreateRule_Success(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Test Rule").Return(nil, errors.New("not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	req := &CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 85.0,
	}
	rule, err := svc.CreateRule(ctx, req, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "Test Rule", rule.Name)
	assert.Equal(t, "admin", rule.CreatedBy)
	repo.AssertExpectations(t)
}

func TestAlarmRuleService_CreateRule_DuplicateName(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	existing := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	repo.On("GetByName", ctx, "Test Rule").Return(existing, nil)

	req := &CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 85.0,
	}
	_, err := svc.CreateRule(ctx, req, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	repo.AssertExpectations(t)
}

func TestAlarmRuleService_CreateRule_InvalidType(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Test Rule").Return(nil, errors.New("not found"))

	req := &CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      "invalid",
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 85.0,
	}
	_, err := svc.CreateRule(ctx, req, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid alarm rule type")
}

func TestAlarmRuleService_CreateRule_EmptyCondition(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Test Rule").Return(nil, errors.New("not found"))

	req := &CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "",
		Threshold: 85.0,
	}
	_, err := svc.CreateRule(ctx, req, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition is required")
}

func TestAlarmRuleService_CreateRule_LimitZeroThreshold(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Test Rule").Return(nil, errors.New("not found"))

	req := &CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 0,
	}
	_, err := svc.CreateRule(ctx, req, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "threshold is required")
}

func TestAlarmRuleService_UpdateRule_Success(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	rule.ID = "rule-001"
	repo.On("GetByID", ctx, "rule-001").Return(rule, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	newName := "Updated Rule"
	req := &UpdateAlarmRuleRequest{Name: &newName}
	updated, err := svc.UpdateRule(ctx, "rule-001", req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Rule", updated.Name)
	repo.AssertExpectations(t)
}

func TestAlarmRuleService_UpdateRule_NotFound(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	req := &UpdateAlarmRuleRequest{}
	_, err := svc.UpdateRule(ctx, "nonexistent", req, "admin")
	assert.Error(t, err)
}

func TestAlarmRuleService_DeleteRule(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("Delete", ctx, "rule-001").Return(nil)

	err := svc.DeleteRule(ctx, "rule-001")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAlarmRuleService_GetRule(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "value > threshold")
	repo.On("GetByID", ctx, "rule-001").Return(rule, nil)

	found, err := svc.GetRule(ctx, "rule-001")
	assert.NoError(t, err)
	assert.Equal(t, "Test Rule", found.Name)
}

func TestAlarmRuleService_ListRules(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rules := []*entity.AlarmRule{
		entity.NewAlarmRule("Rule 1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1"),
	}
	query := &repository.AlarmRuleQuery{Page: 1, PageSize: 10}
	repo.On("List", ctx, query).Return(rules, int64(1), nil)

	found, total, err := svc.ListRules(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, found, 1)
}

func TestAlarmRuleService_EnableRule(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	rule.ID = "rule-001"
	rule.Status = entity.AlarmRuleStatusDisabled
	repo.On("GetByID", ctx, "rule-001").Return(rule, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	updated, err := svc.EnableRule(ctx, "rule-001", "admin")
	assert.NoError(t, err)
	assert.Equal(t, entity.AlarmRuleStatusEnabled, updated.Status)
}

func TestAlarmRuleService_EnableRule_AlreadyEnabled(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	rule.ID = "rule-001"
	rule.Status = entity.AlarmRuleStatusEnabled
	repo.On("GetByID", ctx, "rule-001").Return(rule, nil)

	updated, err := svc.EnableRule(ctx, "rule-001", "admin")
	assert.NoError(t, err)
	assert.Equal(t, entity.AlarmRuleStatusEnabled, updated.Status)
	repo.AssertNotCalled(t, "Update")
}

func TestAlarmRuleService_DisableRule(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	rule.ID = "rule-001"
	repo.On("GetByID", ctx, "rule-001").Return(rule, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	updated, err := svc.DisableRule(ctx, "rule-001", "admin")
	assert.NoError(t, err)
	assert.Equal(t, entity.AlarmRuleStatusDisabled, updated.Status)
}

func TestAlarmRuleService_GetRulesByPointID(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rules := []*entity.AlarmRule{
		entity.NewAlarmRule("Rule 1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1"),
	}
	repo.On("GetRulesByPointID", ctx, "point-001").Return(rules, nil)

	found, err := svc.GetRulesByPointID(ctx, "point-001")
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestAlarmRuleService_GetRulesByDeviceID(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rules := []*entity.AlarmRule{}
	repo.On("GetRulesByDeviceID", ctx, "device-001").Return(rules, nil)

	found, err := svc.GetRulesByDeviceID(ctx, "device-001")
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

func TestAlarmRuleService_GetRulesByStationID(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rules := []*entity.AlarmRule{}
	repo.On("GetRulesByStationID", ctx, "station-001").Return(rules, nil)

	found, err := svc.GetRulesByStationID(ctx, "station-001")
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

func TestAlarmRuleService_CreateRule_RepoError(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Test Rule").Return(nil, errors.New("not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(errors.New("db error"))

	req := &CreateAlarmRuleRequest{
		Name:      "Test Rule",
		Type:      entity.AlarmRuleTypeLimit,
		Level:     entity.AlarmLevelWarning,
		Condition: "value > threshold",
		Threshold: 85.0,
	}
	_, err := svc.CreateRule(ctx, req, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create alarm rule")
}

func TestAlarmRuleService_CreateRule_TrendType(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Trend Rule").Return(nil, errors.New("not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	req := &CreateAlarmRuleRequest{
		Name:      "Trend Rule",
		Type:      entity.AlarmRuleTypeTrend,
		Level:     entity.AlarmLevelMajor,
		Condition: "value decreasing",
	}
	rule, err := svc.CreateRule(ctx, req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, entity.AlarmRuleTypeTrend, rule.Type)
}

func TestAlarmRuleService_CreateRule_CustomType(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	repo.On("GetByName", ctx, "Custom Rule").Return(nil, errors.New("not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	req := &CreateAlarmRuleRequest{
		Name:      "Custom Rule",
		Type:      entity.AlarmRuleTypeCustom,
		Level:     entity.AlarmLevelCritical,
		Condition: "custom_condition()",
	}
	rule, err := svc.CreateRule(ctx, req, "admin")
	assert.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, entity.AlarmRuleTypeCustom, rule.Type)
}

func TestAlarmRuleService_UpdateRule_AllFields(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rule := entity.NewAlarmRule("Test Rule", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1")
	rule.ID = "rule-001"
	repo.On("GetByID", ctx, "rule-001").Return(rule, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

	newName := "Updated"
	newDesc := "New desc"
	newType := entity.AlarmRuleTypeTrend
	newLevel := entity.AlarmLevelCritical
	newCond := "new condition"
	newThreshold := 100.0
	newDuration := 120
	newStatus := entity.AlarmRuleStatusDisabled

	req := &UpdateAlarmRuleRequest{
		Name:           &newName,
		Description:    &newDesc,
		Type:           &newType,
		Level:          &newLevel,
		Condition:      &newCond,
		Threshold:      &newThreshold,
		Duration:       &newDuration,
		NotifyChannels: []string{"email"},
		NotifyUsers:    []string{"user1"},
		Status:         &newStatus,
	}
	updated, err := svc.UpdateRule(ctx, "rule-001", req, "admin")
	assert.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newDesc, updated.Description)
	assert.Equal(t, newType, updated.Type)
	assert.Equal(t, newLevel, updated.Level)
	assert.Equal(t, newCond, updated.Condition)
	assert.Equal(t, newThreshold, updated.Threshold)
	assert.Equal(t, newDuration, updated.Duration)
	assert.Equal(t, []string{"email"}, updated.NotifyChannels)
	assert.Equal(t, []string{"user1"}, updated.NotifyUsers)
	assert.Equal(t, newStatus, updated.Status)
}

func TestAlarmRuleService_GetEnabledRules(t *testing.T) {
	repo := new(mockAlarmRuleRepo)
	svc := NewAlarmRuleService(repo)
	ctx := context.Background()

	rules := []*entity.AlarmRule{
		entity.NewAlarmRule("R1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "c1"),
	}
	repo.On("GetEnabledRules", ctx).Return(rules, nil)

	found, err := svc.GetEnabledRules(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}
