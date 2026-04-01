package handler

import (
	"bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/new-energy-monitoring/internal/application/service"
    "github.com/new-energy-monitoring/internal/domain/entity"
    "github.com/new-energy-monitoring/internal/domain/repository"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockAlarmRuleRepository 告警规则仓储Mock
type MockAlarmRuleRepository struct {
    mock.Mock
}

func (m *MockAlarmRuleRepository) Create(ctx context.Context, rule *entity.AlarmRule) error {
    args := m.Called(ctx, rule)
    return args.Error(0)
}

func (m *MockAlarmRuleRepository) Update(ctx context.Context, rule *entity.AlarmRule) error {
    args := m.Called(ctx, rule)
    return args.Error(0)
}

func (m *MockAlarmRuleRepository) Delete(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockAlarmRuleRepository) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
    args := m.Called(ctx, query)
    if args.Get(0) == nil {
        return nil, args.Get(1).(int64), args.Error(2)
    }
    return args.Get(0).([]*entity.AlarmRule), args.Get(1).(int64), args.Error(2)
}

func (m *MockAlarmRuleRepository) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
    args := m.Called(ctx)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
    args := m.Called(ctx, pointID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
    args := m.Called(ctx, deviceID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func (m *MockAlarmRuleRepository) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
    args := m.Called(ctx, stationID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

func init() {
    gin.SetMode(gin.TestMode)
}

func TestAlarmRuleHandler_CreateAlarmRule(t *testing.T) {
    t.Run("成功创建告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        req := &service.CreateAlarmRuleRequest{
            Name:        "温度过高告警",
            Description: "逆变器温度超过阈值",
            Type:        entity.AlarmRuleTypeLimit,
            Level:       entity.AlarmLevelWarning,
            Condition:   "value > threshold",
            Threshold:   85.0,
            Duration:    60,
            CreatedBy:   "admin",
        }

        mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

        body, _ := json.Marshal(req)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPost, "/alarm-rules", bytes.NewReader(body))
        c.Request.Header.Set("Content-Type", "application/json")

        handler.CreateAlarmRule(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("无效的请求参数", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        body := []byte(`{"invalid": "data"}`)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPost, "/alarm-rules", bytes.NewReader(body))
        c.Request.Header.Set("Content-Type", "application/json")

        handler.CreateAlarmRule(c)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("服务返回错误", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        req := &service.CreateAlarmRuleRequest{
            Name:      "测试规则",
            Type:      entity.AlarmRuleTypeLimit,
            Level:     entity.AlarmLevelWarning,
            CreatedBy: "admin",
        }

        mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(errors.New("database error"))

        body, _ := json.Marshal(req)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPost, "/alarm-rules", bytes.NewReader(body))
        c.Request.Header.Set("Content-Type", "application/json")

        handler.CreateAlarmRule(c)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_UpdateAlarmRule(t *testing.T) {
    t.Run("成功更新告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        existingRule := entity.NewAlarmRule("原始规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
        req := &service.UpdateAlarmRuleRequest{
            Name:      "更新后的规则",
            Level:     entity.AlarmLevelMajor,
            Threshold: 90.0,
            UpdatedBy: "admin",
        }

        mockRepo.On("GetByID", mock.Anything, existingRule.ID).Return(existingRule, nil)
        mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

        body, _ := json.Marshal(req)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPut, "/alarm-rules/"+existingRule.ID, bytes.NewReader(body))
        c.Request.Header.Set("Content-Type", "application/json")
        c.Params = gin.Params{{Key: "id", Value: existingRule.ID}}

        handler.UpdateAlarmRule(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("告警规则不存在", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        req := &service.UpdateAlarmRuleRequest{
            Name:      "更新后的规则",
            Level:     entity.AlarmLevelMajor,
            UpdatedBy: "admin",
        }

        mockRepo.On("GetByID", mock.Anything, "non-existent-id").Return(nil, errors.New("not found"))

        body, _ := json.Marshal(req)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPut, "/alarm-rules/non-existent-id", bytes.NewReader(body))
        c.Request.Header.Set("Content-Type", "application/json")
        c.Params = gin.Params{{Key: "id", Value: "non-existent-id"}}

        handler.UpdateAlarmRule(c)

        assert.Equal(t, http.StatusNotFound, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_DeleteAlarmRule(t *testing.T) {
    t.Run("成功删除告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
        ruleID := rule.ID

        mockRepo.On("GetByID", mock.Anything, ruleID).Return(rule, nil)
        mockRepo.On("Delete", mock.Anything, ruleID).Return(nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodDelete, "/alarm-rules/"+ruleID, nil)
        c.Params = gin.Params{{Key: "id", Value: ruleID}}

        handler.DeleteAlarmRule(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("删除不存在的告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        mockRepo.On("GetByID", mock.Anything, "non-existent-id").Return(nil, errors.New("not found"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodDelete, "/alarm-rules/non-existent-id", nil)
        c.Params = gin.Params{{Key: "id", Value: "non-existent-id"}}

        handler.DeleteAlarmRule(c)

        assert.Equal(t, http.StatusNotFound, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_GetAlarmRule(t *testing.T) {
    t.Run("成功获取告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        expectedRule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
        mockRepo.On("GetByID", mock.Anything, expectedRule.ID).Return(expectedRule, nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/alarm-rules/"+expectedRule.ID, nil)
        c.Params = gin.Params{{Key: "id", Value: expectedRule.ID}}

        handler.GetAlarmRule(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("告警规则不存在", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        mockRepo.On("GetByID", mock.Anything, "non-existent-id").Return(nil, errors.New("not found"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/alarm-rules/non-existent-id", nil)
        c.Params = gin.Params{{Key: "id", Value: "non-existent-id"}}

        handler.GetAlarmRule(c)

        assert.Equal(t, http.StatusNotFound, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_ListAlarmRules(t *testing.T) {
    t.Run("成功获取告警规则列表", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        expectedRules := []*entity.AlarmRule{
            entity.NewAlarmRule("规则1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
            entity.NewAlarmRule("规则2", entity.AlarmRuleTypeTrend, entity.AlarmLevelMajor),
        }

        mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.AlarmRuleQuery")).Return(expectedRules, int64(2), nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/alarm-rules?page=1&page_size=20", nil)

        handler.ListAlarmRules(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("服务返回错误", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.AlarmRuleQuery")).Return(nil, int64(0), errors.New("database error"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/alarm-rules", nil)

        handler.ListAlarmRules(c)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_EnableAlarmRule(t *testing.T) {
    t.Run("成功启用告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
        rule.Disable()
        ruleID := rule.ID

        mockRepo.On("GetByID", mock.Anything, ruleID).Return(rule, nil)
        mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPut, "/alarm-rules/"+ruleID+"/enable", nil)
        c.Params = gin.Params{{Key: "id", Value: ruleID}}

        handler.EnableAlarmRule(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("启用不存在的告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        mockRepo.On("GetByID", mock.Anything, "non-existent-id").Return(nil, errors.New("not found"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPut, "/alarm-rules/non-existent-id/enable", nil)
        c.Params = gin.Params{{Key: "id", Value: "non-existent-id"}}

        handler.EnableAlarmRule(c)

        assert.Equal(t, http.StatusNotFound, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_DisableAlarmRule(t *testing.T) {
    t.Run("成功禁用告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        rule := entity.NewAlarmRule("测试规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning)
        ruleID := rule.ID

        mockRepo.On("GetByID", mock.Anything, ruleID).Return(rule, nil)
        mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlarmRule")).Return(nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPut, "/alarm-rules/"+ruleID+"/disable", nil)
        c.Params = gin.Params{{Key: "id", Value: ruleID}}

        handler.DisableAlarmRule(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("禁用不存在的告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        mockRepo.On("GetByID", mock.Anything, "non-existent-id").Return(nil, errors.New("not found"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodPut, "/alarm-rules/non-existent-id/disable", nil)
        c.Params = gin.Params{{Key: "id", Value: "non-existent-id"}}

        handler.DisableAlarmRule(c)

        assert.Equal(t, http.StatusNotFound, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_GetRulesByPointID(t *testing.T) {
    t.Run("成功获取采集点告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        pointID := "point-001"
        expectedRules := []*entity.AlarmRule{
            entity.NewAlarmRule("采集点规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
        }

        mockRepo.On("GetRulesByPointID", mock.Anything, pointID).Return(expectedRules, nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/points/"+pointID+"/alarm-rules", nil)
        c.Params = gin.Params{{Key: "point_id", Value: pointID}}

        handler.GetRulesByPointID(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })

    t.Run("服务返回错误", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        pointID := "point-001"
        mockRepo.On("GetRulesByPointID", mock.Anything, pointID).Return(nil, errors.New("database error"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/points/"+pointID+"/alarm-rules", nil)
        c.Params = gin.Params{{Key: "point_id", Value: pointID}}

        handler.GetRulesByPointID(c)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_GetRulesByDeviceID(t *testing.T) {
    t.Run("成功获取设备告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        deviceID := "device-001"
        expectedRules := []*entity.AlarmRule{
            entity.NewAlarmRule("设备规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
        }

        mockRepo.On("GetRulesByDeviceID", mock.Anything, deviceID).Return(expectedRules, nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/devices/"+deviceID+"/alarm-rules", nil)
        c.Params = gin.Params{{Key: "device_id", Value: deviceID}}

        handler.GetRulesByDeviceID(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })
}

func TestAlarmRuleHandler_GetRulesByStationID(t *testing.T) {
    t.Run("成功获取厂站告警规则", func(t *testing.T) {
        mockRepo := new(MockAlarmRuleRepository)
        ruleService := service.NewAlarmRuleService(mockRepo)
        handler := NewAlarmRuleHandler(ruleService)

        stationID := "station-001"
        expectedRules := []*entity.AlarmRule{
            entity.NewAlarmRule("厂站规则", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning),
        }

        mockRepo.On("GetRulesByStationID", mock.Anything, stationID).Return(expectedRules, nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = httptest.NewRequest(http.MethodGet, "/stations/"+stationID+"/alarm-rules", nil)
        c.Params = gin.Params{{Key: "station_id", Value: stationID}}

        handler.GetRulesByStationID(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockRepo.AssertExpectations(t)
    })
}
