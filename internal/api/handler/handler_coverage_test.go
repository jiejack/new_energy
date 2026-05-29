package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type covMockAlarmRepo struct{ mock.Mock }

func (m *covMockAlarmRepo) Create(ctx context.Context, a *entity.Alarm) error { return m.Called(ctx, a).Error(0) }
func (m *covMockAlarmRepo) Update(ctx context.Context, a *entity.Alarm) error { return m.Called(ctx, a).Error(0) }
func (m *covMockAlarmRepo) GetByID(ctx context.Context, id string) (*entity.Alarm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}
func (m *covMockAlarmRepo) GetActiveAlarms(ctx context.Context, s *string, l *entity.AlarmLevel) ([]*entity.Alarm, error) {
	args := m.Called(ctx, s, l)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}
func (m *covMockAlarmRepo) GetHistoryAlarms(ctx context.Context, s *string, st, et int64) ([]*entity.Alarm, error) { return nil, nil }
func (m *covMockAlarmRepo) Acknowledge(ctx context.Context, id, by string) error { return m.Called(ctx, id, by).Error(0) }
func (m *covMockAlarmRepo) Clear(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *covMockAlarmRepo) CountByLevel(ctx context.Context, s *string) (map[entity.AlarmLevel]int64, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(map[entity.AlarmLevel]int64), args.Error(1)
}

type covMockConfigRepo struct{ mock.Mock }

func (m *covMockConfigRepo) Create(ctx context.Context, c *entity.SystemConfig) error { return m.Called(ctx, c).Error(0) }
func (m *covMockConfigRepo) Update(ctx context.Context, c *entity.SystemConfig) error { return m.Called(ctx, c).Error(0) }
func (m *covMockConfigRepo) Delete(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *covMockConfigRepo) GetByID(ctx context.Context, id string) (*entity.SystemConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SystemConfig), args.Error(1)
}
func (m *covMockConfigRepo) GetByKey(ctx context.Context, cat, key string) (*entity.SystemConfig, error) {
	args := m.Called(ctx, cat, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SystemConfig), args.Error(1)
}
func (m *covMockConfigRepo) GetByCategory(ctx context.Context, cat string) ([]*entity.SystemConfig, error) {
	args := m.Called(ctx, cat)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Error(1)
}
func (m *covMockConfigRepo) GetAll(ctx context.Context) ([]*entity.SystemConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SystemConfig), args.Error(1)
}
func (m *covMockConfigRepo) List(ctx context.Context, f *entity.SystemConfigFilter) ([]*entity.SystemConfig, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*entity.SystemConfig), args.Get(1).(int64), args.Error(2)
}
func (m *covMockConfigRepo) BatchUpdate(ctx context.Context, c []*entity.SystemConfig) error { return m.Called(ctx, c).Error(0) }
func (m *covMockConfigRepo) ExistsByKey(ctx context.Context, cat, key string) (bool, error) {
	args := m.Called(ctx, cat, key)
	return args.Bool(0), args.Error(1)
}

type covMockLogRepo struct{ mock.Mock }

func (m *covMockLogRepo) Create(ctx context.Context, l *entity.OperationLog) error { return m.Called(ctx, l).Error(0) }
func (m *covMockLogRepo) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}
func (m *covMockLogRepo) List(ctx context.Context, q *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}
func (m *covMockLogRepo) DeleteBefore(ctx context.Context, b int64) (int64, error) {
	args := m.Called(ctx, b)
	return args.Get(0).(int64), args.Error(1)
}

type covMockDeviceRepo struct{ mock.Mock }

func (m *covMockDeviceRepo) Create(ctx context.Context, d *entity.Device) error { return m.Called(ctx, d).Error(0) }
func (m *covMockDeviceRepo) Update(ctx context.Context, d *entity.Device) error { return m.Called(ctx, d).Error(0) }
func (m *covMockDeviceRepo) Delete(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *covMockDeviceRepo) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}
func (m *covMockDeviceRepo) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}
func (m *covMockDeviceRepo) List(ctx context.Context, s *string, t *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, s, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}
func (m *covMockDeviceRepo) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}
func (m *covMockDeviceRepo) GetOnlineDevices(ctx context.Context, s string) ([]*entity.Device, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

type covMockPointRepo struct{ mock.Mock }

func (m *covMockPointRepo) Create(ctx context.Context, p *entity.Point) error { return m.Called(ctx, p).Error(0) }
func (m *covMockPointRepo) BatchCreate(ctx context.Context, p []*entity.Point) error { return m.Called(ctx, p).Error(0) }
func (m *covMockPointRepo) Update(ctx context.Context, p *entity.Point) error { return m.Called(ctx, p).Error(0) }
func (m *covMockPointRepo) Delete(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *covMockPointRepo) GetByID(ctx context.Context, id string) (*entity.Point, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}
func (m *covMockPointRepo) GetByCode(ctx context.Context, code string) (*entity.Point, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}
func (m *covMockPointRepo) ListByDeviceID(ctx context.Context, did string) ([]*entity.Point, error) {
	args := m.Called(ctx, did)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}
func (m *covMockPointRepo) GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}
func (m *covMockPointRepo) ListAll(ctx context.Context) ([]*entity.Point, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}
func (m *covMockPointRepo) List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	args := m.Called(ctx, deviceID, pointType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}
func (m *covMockPointRepo) GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	args := m.Called(ctx, protocol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

type covMockAlarmRuleRepo struct{ mock.Mock }

func (m *covMockAlarmRuleRepo) Create(ctx context.Context, r *entity.AlarmRule) error { return m.Called(ctx, r).Error(0) }
func (m *covMockAlarmRuleRepo) Update(ctx context.Context, r *entity.AlarmRule) error { return m.Called(ctx, r).Error(0) }
func (m *covMockAlarmRuleRepo) Delete(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *covMockAlarmRuleRepo) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}
func (m *covMockAlarmRuleRepo) GetByName(ctx context.Context, n string) (*entity.AlarmRule, error) {
	args := m.Called(ctx, n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AlarmRule), args.Error(1)
}
func (m *covMockAlarmRuleRepo) List(ctx context.Context, q *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]*entity.AlarmRule), args.Get(1).(int64), args.Error(2)
}
func (m *covMockAlarmRuleRepo) GetRulesByPointID(ctx context.Context, pid string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, pid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}
func (m *covMockAlarmRuleRepo) GetRulesByDeviceID(ctx context.Context, did string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, did)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}
func (m *covMockAlarmRuleRepo) GetRulesByStationID(ctx context.Context, sid string) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx, sid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}
func (m *covMockAlarmRuleRepo) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AlarmRule), args.Error(1)
}

type covMockNotificationConfigRepo struct{ mock.Mock }

func (m *covMockNotificationConfigRepo) Create(ctx context.Context, c *entity.NotificationConfig) error { return m.Called(ctx, c).Error(0) }
func (m *covMockNotificationConfigRepo) Update(ctx context.Context, c *entity.NotificationConfig) error { return m.Called(ctx, c).Error(0) }
func (m *covMockNotificationConfigRepo) GetByType(ctx context.Context, t entity.NotificationType) (*entity.NotificationConfig, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.NotificationConfig), args.Error(1)
}
func (m *covMockNotificationConfigRepo) GetAll(ctx context.Context) ([]*entity.NotificationConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.NotificationConfig), args.Error(1)
}

func newCovRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func assertErrResp(t *testing.T, w *httptest.ResponseRecorder, code int) {
	t.Helper()
	assert.Equal(t, code, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["message"])
	assert.Equal(t, float64(code), resp["code"])
}

func assertOKResp(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
	assert.Equal(t, "success", resp["message"])
}

func TestCovAlarmHandler(t *testing.T) {
	t.Run("GetAlarm_NotFound", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		mockRepo.On("GetByID", mock.Anything, "bad-id").Return(nil, assert.AnError)
		r.GET("/alarms/:id", h.GetAlarm)
		w := doRequest(r, http.MethodGet, "/alarms/bad-id", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("GetAlarm_Success", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
		mockRepo.On("GetByID", mock.Anything, "good-id").Return(alarm, nil)
		r.GET("/alarms/:id", h.GetAlarm)
		w := doRequest(r, http.MethodGet, "/alarms/good-id", nil)
		assertOKResp(t, w)
	})

	t.Run("ListAlarms_Success", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		mockRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).Return([]*entity.Alarm{}, nil)
		r.GET("/alarms", h.ListAlarms)
		w := doRequest(r, http.MethodGet, "/alarms", nil)
		assertOKResp(t, w)
	})

	t.Run("ListAlarms_Error", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		mockRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).Return(nil, assert.AnError)
		r.GET("/alarms", h.ListAlarms)
		w := doRequest(r, http.MethodGet, "/alarms", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("AcknowledgeAlarm_Error", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		mockRepo.On("GetByID", mock.Anything, "bad-id").Return(nil, assert.AnError)
		r.PUT("/alarms/:id/ack", h.AcknowledgeAlarm)
		w := doRequest(r, http.MethodPut, "/alarms/bad-id/ack", nil)
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("AcknowledgeAlarm_Success", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
		mockRepo.On("GetByID", mock.Anything, "ok-id").Return(alarm, nil)
		mockRepo.On("Acknowledge", mock.Anything, "ok-id", "").Return(nil)
		r.PUT("/alarms/:id/ack", h.AcknowledgeAlarm)
		w := doRequest(r, http.MethodPut, "/alarms/ok-id/ack", nil)
		assertOKResp(t, w)
	})

	t.Run("ClearAlarm_Error", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		mockRepo.On("GetByID", mock.Anything, "bad-id").Return(nil, assert.AnError)
		r.PUT("/alarms/:id/clear", h.ClearAlarm)
		w := doRequest(r, http.MethodPut, "/alarms/bad-id/clear", nil)
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("ClearAlarm_Success", func(t *testing.T) {
		mockRepo := new(covMockAlarmRepo)
		svc := service.NewAlarmService(mockRepo)
		h := NewAlarmHandler(svc)
		r := newCovRouter()
		alarm := entity.NewAlarm("p1", "d1", "s1", entity.AlarmTypeLimit, entity.AlarmLevelWarning, "Test", "msg")
		mockRepo.On("GetByID", mock.Anything, "ok-id").Return(alarm, nil)
		mockRepo.On("Clear", mock.Anything, "ok-id").Return(nil)
		r.PUT("/alarms/:id/clear", h.ClearAlarm)
		w := doRequest(r, http.MethodPut, "/alarms/ok-id/clear", nil)
		assertOKResp(t, w)
	})
}

func TestCovConfigHandler(t *testing.T) {
	t.Run("GetAllConfigs_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetAll", mock.Anything).Return([]*entity.SystemConfig{}, nil)
		r.GET("/configs", h.GetAllConfigs)
		w := doRequest(r, http.MethodGet, "/configs", nil)
		assertOKResp(t, w)
	})

	t.Run("GetAllConfigs_Error", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetAll", mock.Anything).Return(nil, assert.AnError)
		r.GET("/configs", h.GetAllConfigs)
		w := doRequest(r, http.MethodGet, "/configs", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("GetConfigsByCategory_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetByCategory", mock.Anything, "basic").Return([]*entity.SystemConfig{}, nil)
		r.GET("/configs/:category", h.GetConfigsByCategory)
		w := doRequest(r, http.MethodGet, "/configs/basic", nil)
		assertOKResp(t, w)
	})

	t.Run("GetConfigsByCategory_Error", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetByCategory", mock.Anything, "basic").Return(nil, assert.AnError)
		r.GET("/configs/:category", h.GetConfigsByCategory)
		w := doRequest(r, http.MethodGet, "/configs/basic", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("GetConfig_NotFound", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetByKey", mock.Anything, "basic", "nonexistent").Return(nil, service.ErrConfigNotFound)
		r.GET("/configs/:category/:key", h.GetConfig)
		w := doRequest(r, http.MethodGet, "/configs/basic/nonexistent", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("GetConfig_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfg := entity.NewSystemConfig("basic", "k", "v", entity.SystemConfigValueTypeString, "")
		cfgRepo.On("GetByKey", mock.Anything, "basic", "k").Return(cfg, nil)
		r.GET("/configs/:category/:key", h.GetConfig)
		w := doRequest(r, http.MethodGet, "/configs/basic/k", nil)
		assertOKResp(t, w)
	})

	t.Run("UpdateConfig_InvalidJSON", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		r.PUT("/configs/:category/:key", h.UpdateConfig)
		w := doRequest(r, http.MethodPut, "/configs/basic/k", "invalid json")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("UpdateConfig_NotFound", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetByKey", mock.Anything, "basic", "missing").Return(nil, service.ErrConfigNotFound)
		r.PUT("/configs/:category/:key", h.UpdateConfig)
		body := service.UpdateConfigRequest{Value: "new"}
		w := doRequest(r, http.MethodPut, "/configs/basic/missing", body)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("UpdateConfig_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfg := entity.NewSystemConfig("basic", "k", "old", entity.SystemConfigValueTypeString, "")
		cfgRepo.On("GetByKey", mock.Anything, "basic", "k").Return(cfg, nil)
		cfgRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		r.PUT("/configs/:category/:key", h.UpdateConfig)
		body := service.UpdateConfigRequest{Value: "new_val"}
		w := doRequest(r, http.MethodPut, "/configs/basic/k", body)
		assertOKResp(t, w)
	})

	t.Run("BatchUpdate_InvalidJSON", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		r.POST("/configs/batch", h.BatchUpdateConfigs)
		w := doRequest(r, http.MethodPost, "/configs/batch", "bad")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("CreateConfig_InvalidJSON", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		r.POST("/configs", h.CreateConfig)
		w := doRequest(r, http.MethodPost, "/configs", "bad")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("CreateConfig_KeyExists", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("ExistsByKey", mock.Anything, "basic", "dup").Return(true, nil)
		r.POST("/configs", h.CreateConfig)
		body := service.CreateConfigRequest{Category: "basic", Key: "dup", Value: "v", ValueType: "string"}
		w := doRequest(r, http.MethodPost, "/configs", body)
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("CreateConfig_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("ExistsByKey", mock.Anything, "basic", "new_k").Return(false, nil)
		cfgRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		r.POST("/configs", h.CreateConfig)
		body := service.CreateConfigRequest{Category: "basic", Key: "new_k", Value: "v", ValueType: "string"}
		w := doRequest(r, http.MethodPost, "/configs", body)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("DeleteConfig_NotFound", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("GetByKey", mock.Anything, "basic", "del_missing").Return(nil, service.ErrConfigNotFound)
		r.DELETE("/configs/:category/:key", h.DeleteConfig)
		w := doRequest(r, http.MethodDelete, "/configs/basic/del_missing", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("DeleteConfig_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfg := entity.NewSystemConfig("basic", "del_ok", "v", entity.SystemConfigValueTypeString, "")
		cfgRepo.On("GetByKey", mock.Anything, "basic", "del_ok").Return(cfg, nil)
		cfgRepo.On("Delete", mock.Anything, cfg.ID).Return(nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		r.DELETE("/configs/:category/:key", h.DeleteConfig)
		w := doRequest(r, http.MethodDelete, "/configs/basic/del_ok", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("ListConfigs_Success", func(t *testing.T) {
		cfgRepo := new(covMockConfigRepo)
		logRepo := new(covMockLogRepo)
		svc := service.NewConfigService(cfgRepo, logRepo)
		h := NewConfigHandler(svc)
		r := newCovRouter()
		cfgRepo.On("List", mock.Anything, mock.AnythingOfType("*entity.SystemConfigFilter")).Return([]*entity.SystemConfig{}, int64(0), nil)
		r.GET("/configs/list", h.ListConfigs)
		w := doRequest(r, http.MethodGet, "/configs/list?cat=basic&page=1&page_size=10", nil)
		assertOKResp(t, w)
	})
}

func TestCovDeviceHandler(t *testing.T) {
	t.Run("CreateDevice_InvalidJSON", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		r.POST("/devices", h.CreateDevice)
		w := doRequest(r, http.MethodPost, "/devices", "invalid")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("GetDevice_NotFound", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		devRepo.On("GetByID", mock.Anything, "no-dev").Return(nil, assert.AnError)
		r.GET("/devices/:id", h.GetDevice)
		w := doRequest(r, http.MethodGet, "/devices/no-dev", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("GetDevice_Success", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		d := entity.NewDevice("D1", "Dev", entity.DeviceTypeInverter, "st1")
		devRepo.On("GetByID", mock.Anything, "dev-ok").Return(d, nil)
		r.GET("/devices/:id", h.GetDevice)
		w := doRequest(r, http.MethodGet, "/devices/dev-ok", nil)
		assertOKResp(t, w)
	})

	t.Run("ListDevices_Success", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		devRepo.On("List", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).Return([]*entity.Device{}, nil)
		r.GET("/devices", h.ListDevices)
		w := doRequest(r, http.MethodGet, "/devices", nil)
		assertOKResp(t, w)
	})

	t.Run("ListDevices_Error", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		devRepo.On("List", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).Return(nil, assert.AnError)
		r.GET("/devices", h.ListDevices)
		w := doRequest(r, http.MethodGet, "/devices", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("UpdateDevice_InvalidJSON", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		r.PUT("/devices/:id", h.UpdateDevice)
		w := doRequest(r, http.MethodPut, "/devices/d1", "bad")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("DeleteDevice_NotFound", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		devRepo.On("GetByID", mock.Anything, "no-del").Return(nil, assert.AnError)
		r.DELETE("/devices/:id", h.DeleteDevice)
		w := doRequest(r, http.MethodDelete, "/devices/no-del", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("DeleteDevice_Success", func(t *testing.T) {
		devRepo := new(covMockDeviceRepo)
		ptRepo := new(covMockPointRepo)
		svc := service.NewDeviceService(devRepo, ptRepo)
		h := NewDeviceHandler(svc)
		r := newCovRouter()
		d := entity.NewDevice("D1", "Dev", entity.DeviceTypeInverter, "st1")
		devRepo.On("GetByID", mock.Anything, "del-ok").Return(d, nil)
		ptRepo.On("List", mock.Anything, &d.ID, (*entity.PointType)(nil)).Return([]*entity.Point{}, nil)
		devRepo.On("Delete", mock.Anything, "del-ok").Return(nil)
		r.DELETE("/devices/:id", h.DeleteDevice)
		w := doRequest(r, http.MethodDelete, "/devices/del-ok", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestCovAlarmRuleHandler(t *testing.T) {
	newH := func() (*AlarmRuleHandler, *covMockAlarmRuleRepo, *gin.Engine) {
		repo := new(covMockAlarmRuleRepo)
		svc := service.NewAlarmRuleService(repo)
		h := NewAlarmRuleHandler(svc)
		r := newCovRouter()
		return h, repo, r
	}

	t.Run("CreateAlarmRule_InvalidJSON", func(t *testing.T) {
		h, _, r := newH()
		r.POST("/alarm-rules", h.CreateAlarmRule)
		w := doRequest(r, http.MethodPost, "/alarm-rules", "bad")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("GetAlarmRule_NotFound", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("GetByID", mock.Anything, "no-rule").Return(nil, assert.AnError)
		r.GET("/alarm-rules/:id", h.GetAlarmRule)
		w := doRequest(r, http.MethodGet, "/alarm-rules/no-rule", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("GetAlarmRule_Success", func(t *testing.T) {
		h, repo, r := newH()
		rule := entity.NewAlarmRule("r1", entity.AlarmRuleTypeLimit, entity.AlarmLevelWarning, "expr")
		repo.On("GetByID", mock.Anything, "rule-ok").Return(rule, nil)
		r.GET("/alarm-rules/:id", h.GetAlarmRule)
		w := doRequest(r, http.MethodGet, "/alarm-rules/rule-ok", nil)
		assertOKResp(t, w)
	})

	t.Run("ListAlarmRules_Default", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("List", mock.Anything, mock.AnythingOfType("*repository.AlarmRuleQuery")).Return([]*entity.AlarmRule{}, int64(0), nil)
		r.GET("/alarm-rules", h.ListAlarmRules)
		w := doRequest(r, http.MethodGet, "/alarm-rules", nil)
		assertOKResp(t, w)
	})

	t.Run("ListAlarmRules_WithFilters", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("List", mock.Anything, mock.Anything).Return([]*entity.AlarmRule{}, int64(0), nil)
		r.GET("/alarm-rules", h.ListAlarmRules)
		w := doRequest(r, http.MethodGet, "/alarm-rules?type=limit&level=2&status=1&station_id=st1&page=2&page_size=5", nil)
		assertOKResp(t, w)
	})

	t.Run("UpdateAlarmRule_InvalidJSON", func(t *testing.T) {
		h, _, r := newH()
		r.PUT("/alarm-rules/:id", h.UpdateAlarmRule)
		w := doRequest(r, http.MethodPut, "/alarm-rules/r1", "bad")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("DeleteAlarmRule_NotFound", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("Delete", mock.Anything, "no-del").Return(assert.AnError)
		r.DELETE("/alarm-rules/:id", h.DeleteAlarmRule)
		w := doRequest(r, http.MethodDelete, "/alarm-rules/no-del", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("EnableAlarmRule_NotFound", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("GetByID", mock.Anything, "no-en").Return(nil, assert.AnError)
		r.PUT("/alarm-rules/:id/enable", h.EnableAlarmRule)
		w := doRequest(r, http.MethodPut, "/alarm-rules/no-en/enable", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("DisableAlarmRule_NotFound", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("GetByID", mock.Anything, "no-dis").Return(nil, assert.AnError)
		r.PUT("/alarm-rules/:id/disable", h.DisableAlarmRule)
		w := doRequest(r, http.MethodPut, "/alarm-rules/no-dis/disable", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("GetRulesByPoint_Success", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("GetRulesByPointID", mock.Anything, "pt1").Return([]*entity.AlarmRule{}, nil)
		r.GET("/alarm-rules/by-point/:point_id", h.GetRulesByPoint)
		w := doRequest(r, http.MethodGet, "/alarm-rules/by-point/pt1", nil)
		assertOKResp(t, w)
	})

	t.Run("GetRulesByDevice_Success", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("GetRulesByDeviceID", mock.Anything, "dv1").Return([]*entity.AlarmRule{}, nil)
		r.GET("/alarm-rules/by-device/:device_id", h.GetRulesByDevice)
		w := doRequest(r, http.MethodGet, "/alarm-rules/by-device/dv1", nil)
		assertOKResp(t, w)
	})

	t.Run("GetRulesByStation_Success", func(t *testing.T) {
		h, repo, r := newH()
		repo.On("GetRulesByStationID", mock.Anything, "st1").Return([]*entity.AlarmRule{}, nil)
		r.GET("/alarm-rules/by-station/:station_id", h.GetRulesByStation)
		w := doRequest(r, http.MethodGet, "/alarm-rules/by-station/st1", nil)
		assertOKResp(t, w)
	})
}

func TestCovNotificationConfigHandler(t *testing.T) {
	newH := func() (*NotificationConfigHandler, *covMockNotificationConfigRepo) {
		repo := new(covMockNotificationConfigRepo)
		svc := service.NewNotificationConfigService(repo)
		return NewNotificationConfigHandler(svc), repo
	}

	t.Run("GetAllConfigs_Success", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetAll", mock.Anything).Return([]*entity.NotificationConfig{}, nil)
		r.GET("/notification-configs", h.GetAllConfigs)
		w := doRequest(r, http.MethodGet, "/notification-configs", nil)
		assertOKResp(t, w)
	})

	t.Run("GetAllConfigs_Error", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetAll", mock.Anything).Return(nil, assert.AnError)
		r.GET("/notification-configs", h.GetAllConfigs)
		w := doRequest(r, http.MethodGet, "/notification-configs", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("GetConfigByType_NotFound", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetByType", mock.Anything, mock.Anything).Return(nil, assert.AnError)
		r.GET("/notification-configs/:type", h.GetConfigByType)
		w := doRequest(r, http.MethodGet, "/notification-configs/email", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("GetConfigByType_Success", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		cfg := &entity.NotificationConfig{ID: "c1", Type: entity.NotificationTypeEmail, Name: "Email", Enabled: true}
		repo.On("GetByType", mock.Anything, mock.Anything).Return(cfg, nil)
		r.GET("/notification-configs/:type", h.GetConfigByType)
		w := doRequest(r, http.MethodGet, "/notification-configs/email", nil)
		assertOKResp(t, w)
	})

	t.Run("UpdateConfig_InvalidJSON", func(t *testing.T) {
		h, _ := newH()
		r := newCovRouter()
		r.PUT("/notification-configs/:type", h.UpdateConfig)
		w := doRequest(r, http.MethodPut, "/notification-configs/email", "bad")
		assertErrResp(t, w, http.StatusBadRequest)
	})

	t.Run("UpdateConfig_NotFound", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetByType", mock.Anything, mock.Anything).Return(nil, assert.AnError)
		repo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
		r.PUT("/notification-configs/:type", h.UpdateConfig)
		body := map[string]interface{}{"enabled": true}
		w := doRequest(r, http.MethodPut, "/notification-configs/email", body)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("UpdateConfig_Success", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		cfg := &entity.NotificationConfig{ID: "c1", Type: entity.NotificationTypeEmail, Name: "Email", Enabled: false}
		repo.On("GetByType", mock.Anything, mock.Anything).Return(cfg, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		r.PUT("/notification-configs/:type", h.UpdateConfig)
		body := map[string]interface{}{"enabled": true, "config": map[string]string{"host": "smtp"}}
		w := doRequest(r, http.MethodPut, "/notification-configs/email", body)
		assertOKResp(t, w)
	})

	t.Run("EnableConfig_NotFound", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetByType", mock.Anything, mock.Anything).Return(nil, assert.AnError)
		r.PUT("/notification-configs/:type/enable", h.EnableConfig)
		w := doRequest(r, http.MethodPut, "/notification-configs/sms/enable", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})

	t.Run("DisableConfig_NotFound", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetByType", mock.Anything, mock.Anything).Return(nil, assert.AnError)
		r.PUT("/notification-configs/:type/disable", h.DisableConfig)
		w := doRequest(r, http.MethodPut, "/notification-configs/webhook/disable", nil)
		assertErrResp(t, w, http.StatusInternalServerError)
	})
}

func TestCovOperationLogHandler(t *testing.T) {
	newH := func() (*OperationLogHandler, *covMockLogRepo) {
		repo := new(covMockLogRepo)
		svc := service.NewOperationLogService(repo)
		return NewOperationLogHandler(svc), repo
	}

	t.Run("ListLogs_Default", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("List", mock.Anything, mock.Anything).Return([]*entity.OperationLog{}, int64(0), nil)
		r.GET("/operation-logs", h.ListLogs)
		w := doRequest(r, http.MethodGet, "/operation-logs?page=1&page_size=20", nil)
		assertOKResp(t, w)
	})

	t.Run("ListLogs_WithFilters", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("List", mock.Anything, mock.Anything).Return([]*entity.OperationLog{}, int64(0), nil)
		r.GET("/operation-logs", h.ListLogs)
		w := doRequest(r, http.MethodGet, "/operation-logs?user_id=u1&action=login&resource_type=config&start=100&end=200&page=1&page_size=10", nil)
		assertOKResp(t, w)
	})

	t.Run("GetLog_NotFound", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
		r.GET("/operation-logs/:id", h.GetLog)
		w := doRequest(r, http.MethodGet, "/operation-logs/no-log", nil)
		assertErrResp(t, w, http.StatusNotFound)
	})

	t.Run("GetLog_Success", func(t *testing.T) {
		h, repo := newH()
		r := newCovRouter()
		l := entity.NewOperationLog("u1", "admin", "login")
		repo.On("GetByID", mock.Anything, mock.Anything).Return(l, nil)
		r.GET("/operation-logs/:id", h.GetLog)
		w := doRequest(r, http.MethodGet, "/operation-logs/log-ok", nil)
		assertOKResp(t, w)
	})
}
