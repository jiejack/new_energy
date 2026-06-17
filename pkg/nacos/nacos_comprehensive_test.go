package nacos

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- Mock IConfigClient ---

type mockConfigClient struct {
	mu               sync.Mutex
	getConfigFunc    func(param vo.ConfigParam) (string, error)
	publishFunc      func(param vo.ConfigParam) (bool, error)
	deleteFunc       func(param vo.ConfigParam) (bool, error)
	listenFunc       func(params vo.ConfigParam) error
	cancelListenFunc func(params vo.ConfigParam) error
	searchFunc       func(param vo.SearchConfigParam) (*model.ConfigPage, error)
	getConfigCalled  bool
	publishCalled    bool
	deleteCalled     bool
	listenCalled     bool
	cancelCalled     bool
	searchCalled     bool
}

func (m *mockConfigClient) GetConfig(param vo.ConfigParam) (string, error) {
	m.mu.Lock()
	m.getConfigCalled = true
	m.mu.Unlock()
	if m.getConfigFunc != nil {
		return m.getConfigFunc(param)
	}
	return "", nil
}

func (m *mockConfigClient) PublishConfig(param vo.ConfigParam) (bool, error) {
	m.mu.Lock()
	m.publishCalled = true
	m.mu.Unlock()
	if m.publishFunc != nil {
		return m.publishFunc(param)
	}
	return true, nil
}

func (m *mockConfigClient) DeleteConfig(param vo.ConfigParam) (bool, error) {
	m.mu.Lock()
	m.deleteCalled = true
	m.mu.Unlock()
	if m.deleteFunc != nil {
		return m.deleteFunc(param)
	}
	return true, nil
}

func (m *mockConfigClient) ListenConfig(params vo.ConfigParam) error {
	m.mu.Lock()
	m.listenCalled = true
	m.mu.Unlock()
	if m.listenFunc != nil {
		return m.listenFunc(params)
	}
	return nil
}

func (m *mockConfigClient) CancelListenConfig(params vo.ConfigParam) error {
	m.mu.Lock()
	m.cancelCalled = true
	m.mu.Unlock()
	if m.cancelListenFunc != nil {
		return m.cancelListenFunc(params)
	}
	return nil
}

func (m *mockConfigClient) SearchConfig(param vo.SearchConfigParam) (*model.ConfigPage, error) {
	m.mu.Lock()
	m.searchCalled = true
	m.mu.Unlock()
	if m.searchFunc != nil {
		return m.searchFunc(param)
	}
	return &model.ConfigPage{
		TotalCount:     1,
		PageNumber:     1,
		PagesAvailable: 1,
		PageItems: []model.ConfigItem{
			{DataId: param.DataId, Group: param.Group, Content: "test-content", Tenant: "public"},
		},
	}, nil
}

func (m *mockConfigClient) CloseClient() {}

// --- Mock INamingClient ---

type mockNamingClient struct {
	mu                 sync.Mutex
	registerFunc       func(param vo.RegisterInstanceParam) (bool, error)
	batchRegisterFunc  func(param vo.BatchRegisterInstanceParam) (bool, error)
	deregisterFunc     func(param vo.DeregisterInstanceParam) (bool, error)
	updateFunc         func(param vo.UpdateInstanceParam) (bool, error)
	getServiceFunc     func(param vo.GetServiceParam) (model.Service, error)
	selectAllFunc      func(param vo.SelectAllInstancesParam) ([]model.Instance, error)
	selectInstances    func(param vo.SelectInstancesParam) ([]model.Instance, error)
	selectOneFunc      func(param vo.SelectOneHealthInstanceParam) (*model.Instance, error)
	subscribeFunc      func(param *vo.SubscribeParam) error
	unsubscribeFunc    func(param *vo.SubscribeParam) error
	getAllServicesFunc func(param vo.GetAllServiceInfoParam) (model.ServiceList, error)
	serverHealthyFunc  func() bool
	registerCalled     bool
	deregisterCalled   bool
	updateCalled       bool
	selectCalled       bool
	selectOneCalled    bool
	subscribeCalled    bool
	unsubscribeCalled  bool
}

func (m *mockNamingClient) RegisterInstance(param vo.RegisterInstanceParam) (bool, error) {
	m.mu.Lock()
	m.registerCalled = true
	m.mu.Unlock()
	if m.registerFunc != nil {
		return m.registerFunc(param)
	}
	return true, nil
}

func (m *mockNamingClient) BatchRegisterInstance(param vo.BatchRegisterInstanceParam) (bool, error) {
	if m.batchRegisterFunc != nil {
		return m.batchRegisterFunc(param)
	}
	return true, nil
}

func (m *mockNamingClient) DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error) {
	m.mu.Lock()
	m.deregisterCalled = true
	m.mu.Unlock()
	if m.deregisterFunc != nil {
		return m.deregisterFunc(param)
	}
	return true, nil
}

func (m *mockNamingClient) UpdateInstance(param vo.UpdateInstanceParam) (bool, error) {
	m.mu.Lock()
	m.updateCalled = true
	m.mu.Unlock()
	if m.updateFunc != nil {
		return m.updateFunc(param)
	}
	return true, nil
}

func (m *mockNamingClient) GetService(param vo.GetServiceParam) (model.Service, error) {
	if m.getServiceFunc != nil {
		return m.getServiceFunc(param)
	}
	return model.Service{}, nil
}

func (m *mockNamingClient) SelectAllInstances(param vo.SelectAllInstancesParam) ([]model.Instance, error) {
	if m.selectAllFunc != nil {
		return m.selectAllFunc(param)
	}
	return nil, nil
}

func (m *mockNamingClient) SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error) {
	m.mu.Lock()
	m.selectCalled = true
	m.mu.Unlock()
	if m.selectInstances != nil {
		return m.selectInstances(param)
	}
	return []model.Instance{
		{ServiceName: param.ServiceName, Ip: "10.0.0.1", Port: 8080, Weight: 1.0, Enable: true, Healthy: true, Ephemeral: true},
	}, nil
}

func (m *mockNamingClient) SelectOneHealthyInstance(param vo.SelectOneHealthInstanceParam) (*model.Instance, error) {
	m.mu.Lock()
	m.selectOneCalled = true
	m.mu.Unlock()
	if m.selectOneFunc != nil {
		return m.selectOneFunc(param)
	}
	return &model.Instance{ServiceName: param.ServiceName, Ip: "10.0.0.1", Port: 8080, Weight: 1.0, Enable: true, Healthy: true, Ephemeral: true}, nil
}

func (m *mockNamingClient) Subscribe(param *vo.SubscribeParam) error {
	m.mu.Lock()
	m.subscribeCalled = true
	m.mu.Unlock()
	if m.subscribeFunc != nil {
		return m.subscribeFunc(param)
	}
	return nil
}

func (m *mockNamingClient) Unsubscribe(param *vo.SubscribeParam) error {
	m.mu.Lock()
	m.unsubscribeCalled = true
	m.mu.Unlock()
	if m.unsubscribeFunc != nil {
		return m.unsubscribeFunc(param)
	}
	return nil
}

func (m *mockNamingClient) GetAllServicesInfo(param vo.GetAllServiceInfoParam) (model.ServiceList, error) {
	if m.getAllServicesFunc != nil {
		return m.getAllServicesFunc(param)
	}
	return model.ServiceList{}, nil
}

func (m *mockNamingClient) ServerHealthy() bool {
	if m.serverHealthyFunc != nil {
		return m.serverHealthyFunc()
	}
	return true
}

func (m *mockNamingClient) CloseClient() {}

// --- Helper functions ---

func newTestConfigClient(mock *mockConfigClient) *ConfigClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConfigClient{
		options:   ApplyOptions(WithNamespace("test-ns")),
		configCli: mock,
		listeners: make(map[string]*ConfigListener),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func newTestRegistry(mock *mockNamingClient) *Registry {
	ctx, cancel := context.WithCancel(context.Background())
	return &Registry{
		options:   ApplyOptions(WithNamespace("test-ns")),
		namingCli: mock,
		instances: make(map[string]*ServiceInstance),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func newTestHealthChecker(registry *Registry) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		registry:       registry,
		options:        ApplyOptions(WithHeartbeat(true, 100*time.Millisecond)),
		heartbeatTasks: make(map[string]*HeartbeatTask),
		ctx:            ctx,
		cancel:         cancel,
		logger:         zap.NewNop(),
	}
}

// --- Interface compliance ---

func TestMockConfigClient_ImplementsIConfigClient(t *testing.T) {
	var _ config_client.IConfigClient = (*mockConfigClient)(nil)
}

func TestMockNamingClient_ImplementsINamingClient(t *testing.T) {
	var _ naming_client.INamingClient = (*mockNamingClient)(nil)
}

// =============================================================================
// ConfigClient Tests
// =============================================================================

func TestConfigClient_GetConfig_Success(t *testing.T) {
	mock := &mockConfigClient{
		getConfigFunc: func(param vo.ConfigParam) (string, error) {
			return "config-value", nil
		},
	}
	client := newTestConfigClient(mock)

	content, err := client.GetConfig("test-data-id", "test-group")
	assert.NoError(t, err)
	assert.Equal(t, "config-value", content)
	assert.True(t, mock.getConfigCalled)
}

func TestConfigClient_GetConfig_UsesDefaultGroup(t *testing.T) {
	var capturedGroup string
	mock := &mockConfigClient{
		getConfigFunc: func(param vo.ConfigParam) (string, error) {
			capturedGroup = param.Group
			return "value", nil
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.GetConfig("data-id", "")
	assert.NoError(t, err)
	assert.Equal(t, "DEFAULT_GROUP", capturedGroup)
}

func TestConfigClient_GetConfig_Error(t *testing.T) {
	mock := &mockConfigClient{
		getConfigFunc: func(param vo.ConfigParam) (string, error) {
			return "", errors.New("config not found")
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.GetConfig("data-id", "group")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get config")
}

func TestConfigClient_PublishConfig_Success(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	success, err := client.PublishConfig("data-id", "group", "content")
	assert.NoError(t, err)
	assert.True(t, success)
	assert.True(t, mock.publishCalled)
}

func TestConfigClient_PublishConfig_UsesDefaultGroup(t *testing.T) {
	var capturedGroup string
	mock := &mockConfigClient{
		publishFunc: func(param vo.ConfigParam) (bool, error) {
			capturedGroup = param.Group
			return true, nil
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.PublishConfig("data-id", "", "content")
	assert.NoError(t, err)
	assert.Equal(t, "DEFAULT_GROUP", capturedGroup)
}

func TestConfigClient_PublishConfig_Error(t *testing.T) {
	mock := &mockConfigClient{
		publishFunc: func(param vo.ConfigParam) (bool, error) {
			return false, errors.New("publish failed")
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.PublishConfig("data-id", "group", "content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish config")
}

func TestConfigClient_DeleteConfig_Success(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	success, err := client.DeleteConfig("data-id", "group")
	assert.NoError(t, err)
	assert.True(t, success)
	assert.True(t, mock.deleteCalled)
}

func TestConfigClient_DeleteConfig_UsesDefaultGroup(t *testing.T) {
	var capturedGroup string
	mock := &mockConfigClient{
		deleteFunc: func(param vo.ConfigParam) (bool, error) {
			capturedGroup = param.Group
			return true, nil
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.DeleteConfig("data-id", "")
	assert.NoError(t, err)
	assert.Equal(t, "DEFAULT_GROUP", capturedGroup)
}

func TestConfigClient_DeleteConfig_Error(t *testing.T) {
	mock := &mockConfigClient{
		deleteFunc: func(param vo.ConfigParam) (bool, error) {
			return false, errors.New("delete failed")
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.DeleteConfig("data-id", "group")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete config")
}

func TestConfigClient_ListenConfig_Success(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	err := client.ListenConfig("data-id", "group", callback)
	assert.NoError(t, err)
	assert.True(t, mock.listenCalled)

	client.mu.RLock()
	_, exists := client.listeners["data-id@group"]
	client.mu.RUnlock()
	assert.True(t, exists)
}

func TestConfigClient_ListenConfig_AlreadyExists(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	err := client.ListenConfig("data-id", "group", callback)
	assert.NoError(t, err)

	err = client.ListenConfig("data-id", "group", callback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfigClient_ListenConfig_UsesDefaultGroup(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	err := client.ListenConfig("data-id", "", callback)
	assert.NoError(t, err)

	client.mu.RLock()
	_, exists := client.listeners["data-id@DEFAULT_GROUP"]
	client.mu.RUnlock()
	assert.True(t, exists)
}

func TestConfigClient_ListenConfig_Error(t *testing.T) {
	mock := &mockConfigClient{
		listenFunc: func(params vo.ConfigParam) error {
			return errors.New("listen failed")
		},
	}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	err := client.ListenConfig("data-id", "group", callback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to listen config")

	client.mu.RLock()
	_, exists := client.listeners["data-id@group"]
	client.mu.RUnlock()
	assert.False(t, exists)
}

func TestConfigClient_CancelListenConfig_Success(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	err := client.ListenConfig("data-id", "group", callback)
	assert.NoError(t, err)

	err = client.CancelListenConfig("data-id", "group")
	assert.NoError(t, err)
	assert.True(t, mock.cancelCalled)

	client.mu.RLock()
	_, exists := client.listeners["data-id@group"]
	client.mu.RUnlock()
	assert.False(t, exists)
}

func TestConfigClient_CancelListenConfig_NotFound(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	err := client.CancelListenConfig("nonexistent", "group")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConfigClient_CancelListenConfig_UsesDefaultGroup(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	err := client.ListenConfig("data-id", "", callback)
	assert.NoError(t, err)

	err = client.CancelListenConfig("data-id", "")
	assert.NoError(t, err)
}

func TestConfigClient_CancelListenConfig_Error(t *testing.T) {
	mock := &mockConfigClient{
		cancelListenFunc: func(params vo.ConfigParam) error {
			return errors.New("cancel failed")
		},
	}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	_ = client.ListenConfig("data-id", "group", callback)

	err := client.CancelListenConfig("data-id", "group")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to cancel listen config")
}

func TestConfigClient_SearchConfig_Success(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	result, err := client.SearchConfig("data-id", "group", 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.TotalCount)
	assert.Len(t, result.PageItems, 1)
	assert.Equal(t, "test-content", result.PageItems[0].Content)
}

func TestConfigClient_SearchConfig_UsesDefaultGroup(t *testing.T) {
	var capturedGroup string
	mock := &mockConfigClient{
		searchFunc: func(param vo.SearchConfigParam) (*model.ConfigPage, error) {
			capturedGroup = param.Group
			return &model.ConfigPage{TotalCount: 0}, nil
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.SearchConfig("data-id", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, "DEFAULT_GROUP", capturedGroup)
}

func TestConfigClient_SearchConfig_Error(t *testing.T) {
	mock := &mockConfigClient{
		searchFunc: func(param vo.SearchConfigParam) (*model.ConfigPage, error) {
			return nil, errors.New("search failed")
		},
	}
	client := newTestConfigClient(mock)

	_, err := client.SearchConfig("data-id", "group", 1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search config")
}

func TestConfigClient_GetConfigAndListen_Success(t *testing.T) {
	mock := &mockConfigClient{
		getConfigFunc: func(param vo.ConfigParam) (string, error) {
			return "initial-config", nil
		},
	}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	content, err := client.GetConfigAndListen("data-id", "group", callback)
	assert.NoError(t, err)
	assert.Equal(t, "initial-config", content)
}

func TestConfigClient_GetConfigAndListen_GetConfigError(t *testing.T) {
	mock := &mockConfigClient{
		getConfigFunc: func(param vo.ConfigParam) (string, error) {
			return "", errors.New("get failed")
		},
	}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	_, err := client.GetConfigAndListen("data-id", "group", callback)
	assert.Error(t, err)
}

func TestConfigClient_GetConfigAndListen_ListenError(t *testing.T) {
	mock := &mockConfigClient{
		getConfigFunc: func(param vo.ConfigParam) (string, error) {
			return "config", nil
		},
		listenFunc: func(params vo.ConfigParam) error {
			return errors.New("listen failed")
		},
	}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	_, err := client.GetConfigAndListen("data-id", "group", callback)
	assert.Error(t, err)
}

func TestConfigClient_Close(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)

	callback := func(namespace, group, dataId, content string) {}
	_ = client.ListenConfig("data-id-1", "group", callback)
	_ = client.ListenConfig("data-id-2", "group", callback)

	err := client.Close()
	assert.NoError(t, err)

	client.mu.RLock()
	assert.Empty(t, client.listeners)
	client.mu.RUnlock()
}

func TestConfigClient_Close_Empty(t *testing.T) {
	mock := &mockConfigClient{}
	client := newTestConfigClient(mock)
	err := client.Close()
	assert.NoError(t, err)
}

func TestWithConfigDataId(t *testing.T) {
	opts := &ConfigClientOptions{}
	WithConfigDataId("my-data-id")(opts)
	assert.Equal(t, "my-data-id", opts.DataId)
}

func TestWithConfigGroup(t *testing.T) {
	opts := &ConfigClientOptions{}
	WithConfigGroup("my-group")(opts)
	assert.Equal(t, "my-group", opts.Group)
}

func TestWithConfigNamespace(t *testing.T) {
	opts := &ConfigClientOptions{}
	WithConfigNamespace("my-namespace")(opts)
	assert.Equal(t, "my-namespace", opts.Namespace)
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestRegistry_Register_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	instance := &ServiceInstance{
		ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080,
		Weight: 1.0, Enable: true, Healthy: true,
		GroupName: "DEFAULT_GROUP", ClusterName: "DEFAULT", Ephemeral: true,
	}

	err := registry.Register(instance)
	assert.NoError(t, err)
	assert.True(t, mock.registerCalled)
	assert.True(t, registry.IsRegistered())
}

func TestRegistry_Register_NilInstance(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	err := registry.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance cannot be nil")
}

func TestRegistry_Register_AutoDetectIP(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	instance := &ServiceInstance{ServiceName: "test-service", Port: 8080, Weight: 1.0, Enable: true, Healthy: true}
	err := registry.Register(instance)
	assert.NoError(t, err)

	registry.mu.RLock()
	saved := registry.instances["test-service"]
	registry.mu.RUnlock()
	assert.NotEmpty(t, saved.Ip)
}

func TestRegistry_Register_Error(t *testing.T) {
	mock := &mockNamingClient{
		registerFunc: func(param vo.RegisterInstanceParam) (bool, error) {
			return false, errors.New("register failed")
		},
	}
	registry := newTestRegistry(mock)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	err := registry.Register(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register instance")
}

func TestRegistry_Register_ReturnsFalse(t *testing.T) {
	mock := &mockNamingClient{
		registerFunc: func(param vo.RegisterInstanceParam) (bool, error) {
			return false, nil
		},
	}
	registry := newTestRegistry(mock)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	err := registry.Register(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
}

func TestRegistry_Deregister_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080, GroupName: "DEFAULT_GROUP", Ephemeral: true}
	_ = registry.Register(instance)

	err := registry.Deregister("test-service")
	assert.NoError(t, err)
	assert.True(t, mock.deregisterCalled)
	assert.False(t, registry.IsRegistered())
}

func TestRegistry_Deregister_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	err := registry.Deregister("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_Deregister_Error(t *testing.T) {
	mock := &mockNamingClient{
		deregisterFunc: func(param vo.DeregisterInstanceParam) (bool, error) {
			return false, errors.New("deregister failed")
		},
	}
	registry := newTestRegistry(mock)
	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	err := registry.Deregister("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deregister instance")
}

func TestRegistry_Deregister_ReturnsFalse(t *testing.T) {
	mock := &mockNamingClient{
		deregisterFunc: func(param vo.DeregisterInstanceParam) (bool, error) {
			return false, nil
		},
	}
	registry := newTestRegistry(mock)
	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	err := registry.Deregister("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
}

func TestRegistry_Discover_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	instances, err := registry.Discover("test-service")
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "10.0.0.1", instances[0].Ip)
}

func TestRegistry_Discover_WithOptions(t *testing.T) {
	var capturedGroup string
	var capturedClusters []string
	var capturedHealthyOnly bool
	mock := &mockNamingClient{
		selectInstances: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
			capturedGroup = param.GroupName
			capturedClusters = param.Clusters
			capturedHealthyOnly = param.HealthyOnly
			return []model.Instance{}, nil
		},
	}
	registry := newTestRegistry(mock)

	_, err := registry.Discover("test-service",
		WithDiscoveryGroup("CUSTOM_GROUP"),
		WithDiscoveryClusters([]string{"cluster-1", "cluster-2"}),
		WithDiscoveryHealthyOnly(false),
	)
	assert.NoError(t, err)
	assert.Equal(t, "CUSTOM_GROUP", capturedGroup)
	assert.Equal(t, []string{"cluster-1", "cluster-2"}, capturedClusters)
	assert.False(t, capturedHealthyOnly)
}

func TestRegistry_Discover_Error(t *testing.T) {
	mock := &mockNamingClient{
		selectInstances: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
			return nil, errors.New("discover failed")
		},
	}
	registry := newTestRegistry(mock)

	_, err := registry.Discover("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover service")
}

func TestRegistry_DiscoverOne_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	instance, err := registry.DiscoverOne("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "10.0.0.1", instance.Ip)
}

func TestRegistry_DiscoverOne_WithOptions(t *testing.T) {
	var capturedGroup string
	mock := &mockNamingClient{
		selectOneFunc: func(param vo.SelectOneHealthInstanceParam) (*model.Instance, error) {
			capturedGroup = param.GroupName
			return &model.Instance{ServiceName: param.ServiceName}, nil
		},
	}
	registry := newTestRegistry(mock)

	_, err := registry.DiscoverOne("test-service", WithDiscoveryGroup("CUSTOM"))
	assert.NoError(t, err)
	assert.Equal(t, "CUSTOM", capturedGroup)
}

func TestRegistry_DiscoverOne_Error(t *testing.T) {
	mock := &mockNamingClient{
		selectOneFunc: func(param vo.SelectOneHealthInstanceParam) (*model.Instance, error) {
			return nil, errors.New("discover one failed")
		},
	}
	registry := newTestRegistry(mock)

	_, err := registry.DiscoverOne("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover one instance")
}

func TestRegistry_Subscribe_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	callback := func(instances []*ServiceInstance) {}
	err := registry.Subscribe("test-service", callback)
	assert.NoError(t, err)
}

func TestRegistry_Subscribe_WithOptions(t *testing.T) {
	var capturedGroup string
	mock := &mockNamingClient{
		subscribeFunc: func(param *vo.SubscribeParam) error {
			capturedGroup = param.GroupName
			return nil
		},
	}
	registry := newTestRegistry(mock)

	callback := func(instances []*ServiceInstance) {}
	err := registry.Subscribe("test-service", callback, WithDiscoveryGroup("CUSTOM"))
	assert.NoError(t, err)
	assert.Equal(t, "CUSTOM", capturedGroup)
}

func TestRegistry_Subscribe_Error(t *testing.T) {
	mock := &mockNamingClient{
		subscribeFunc: func(param *vo.SubscribeParam) error {
			return errors.New("subscribe failed")
		},
	}
	registry := newTestRegistry(mock)

	callback := func(instances []*ServiceInstance) {}
	err := registry.Subscribe("test-service", callback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to subscribe service")
}

func TestRegistry_Unsubscribe_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	err := registry.Unsubscribe("test-service")
	assert.NoError(t, err)
}

func TestRegistry_Unsubscribe_WithOptions(t *testing.T) {
	var capturedGroup string
	mock := &mockNamingClient{
		unsubscribeFunc: func(param *vo.SubscribeParam) error {
			capturedGroup = param.GroupName
			return nil
		},
	}
	registry := newTestRegistry(mock)

	err := registry.Unsubscribe("test-service", WithDiscoveryGroup("CUSTOM"))
	assert.NoError(t, err)
	assert.Equal(t, "CUSTOM", capturedGroup)
}

func TestRegistry_Unsubscribe_Error(t *testing.T) {
	mock := &mockNamingClient{
		unsubscribeFunc: func(param *vo.SubscribeParam) error {
			return errors.New("unsubscribe failed")
		},
	}
	registry := newTestRegistry(mock)

	err := registry.Unsubscribe("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unsubscribe service")
}

func TestRegistry_GetAllServices(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	services, err := registry.GetAllServices()
	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.Empty(t, services)
}

func TestRegistry_GetAllServices_WithOptions(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	services, err := registry.GetAllServices(WithDiscoveryGroup("CUSTOM"))
	assert.NoError(t, err)
	assert.NotNil(t, services)
}

func TestRegistry_Close(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	err := registry.Close()
	assert.NoError(t, err)
	assert.False(t, registry.IsRegistered())
}

func TestRegistry_Close_Empty(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	err := registry.Close()
	assert.NoError(t, err)
}

func TestRegistry_IsRegistered_Empty(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	assert.False(t, registry.IsRegistered())
}

func TestWithDiscoveryGroup(t *testing.T) {
	opts := &DiscoveryOptions{}
	WithDiscoveryGroup("MY_GROUP")(opts)
	assert.Equal(t, "MY_GROUP", opts.Group)
}

func TestWithDiscoveryClusters(t *testing.T) {
	opts := &DiscoveryOptions{}
	WithDiscoveryClusters([]string{"cluster-1", "cluster-2"})(opts)
	assert.Equal(t, []string{"cluster-1", "cluster-2"}, opts.Clusters)
}

func TestWithDiscoveryHealthyOnly(t *testing.T) {
	opts := &DiscoveryOptions{}
	WithDiscoveryHealthyOnly(false)(opts)
	assert.False(t, opts.HealthyOnly)
}

// =============================================================================
// HealthChecker Tests
// =============================================================================

func TestNewHealthChecker_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	checker, err := NewHealthChecker(registry)
	assert.NoError(t, err)
	assert.NotNil(t, checker)
}

func TestNewHealthChecker_NilRegistry(t *testing.T) {
	_, err := NewHealthChecker(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registry cannot be nil")
}

func TestNewHealthChecker_WithOptions(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	checker, err := NewHealthChecker(registry, WithHeartbeat(true, 10*time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Second, checker.options.HeartbeatInterval)
}

func TestHealthChecker_StartHeartbeat_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080, Ephemeral: true}
	err := checker.StartHeartbeat(instance)
	assert.NoError(t, err)

	tasks := checker.GetHeartbeatTasks()
	assert.Contains(t, tasks, "test-service")
	_ = checker.StopHeartbeat("test-service")
}

func TestHealthChecker_StartHeartbeat_NilInstance(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.StartHeartbeat(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance cannot be nil")
}

func TestHealthChecker_StartHeartbeat_AlreadyExists(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	err := checker.StartHeartbeat(instance)
	assert.NoError(t, err)

	err = checker.StartHeartbeat(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	_ = checker.StopHeartbeat("test-service")
}

func TestHealthChecker_StartHeartbeat_DefaultInterval(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)

	ctx, cancel := context.WithCancel(context.Background())
	checker := &HealthChecker{
		registry:       registry,
		options:        ApplyOptions(WithHeartbeat(true, 0)),
		heartbeatTasks: make(map[string]*HeartbeatTask),
		ctx:            ctx,
		cancel:         cancel,
		logger:         zap.NewNop(),
	}
	defer cancel()

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080, Ephemeral: true}
	err := checker.StartHeartbeat(instance)
	require.NoError(t, err)

	checker.mu.Lock()
	task := checker.heartbeatTasks["test-service"]
	checker.mu.Unlock()
	assert.NotNil(t, task)
	assert.Equal(t, 5*time.Second, task.Interval)
	_ = checker.StopHeartbeat("test-service")
}

func TestHealthChecker_StopHeartbeat_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = checker.StartHeartbeat(instance)

	err := checker.StopHeartbeat("test-service")
	assert.NoError(t, err)
	assert.NotContains(t, checker.GetHeartbeatTasks(), "test-service")
}

func TestHealthChecker_StopHeartbeat_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.StopHeartbeat("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_Beat_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080, Ephemeral: true}
	_ = registry.Register(instance)
	_ = checker.StartHeartbeat(instance)

	err := checker.Beat("test-service")
	assert.NoError(t, err)
	_ = checker.StopHeartbeat("test-service")
}

func TestHealthChecker_Beat_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.Beat("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_GetHeartbeatTasks(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance1 := &ServiceInstance{ServiceName: "svc-1", Ip: "10.0.0.1", Port: 8080}
	instance2 := &ServiceInstance{ServiceName: "svc-2", Ip: "10.0.0.2", Port: 8081}
	_ = checker.StartHeartbeat(instance1)
	_ = checker.StartHeartbeat(instance2)

	tasks := checker.GetHeartbeatTasks()
	assert.Len(t, tasks, 2)
	assert.Contains(t, tasks, "svc-1")
	assert.Contains(t, tasks, "svc-2")

	_ = checker.StopHeartbeat("svc-1")
	_ = checker.StopHeartbeat("svc-2")
}

func TestHealthChecker_CheckInstance_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	_, err := checker.CheckInstance("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_CheckInstance_DiscoverError(t *testing.T) {
	mock := &mockNamingClient{
		selectInstances: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
			return nil, errors.New("discover failed")
		},
	}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	status, err := checker.CheckInstance("test-service")
	assert.NoError(t, err)
	assert.False(t, status.Healthy)
	assert.Error(t, status.Error)
}

func TestHealthChecker_CheckInstance_InstanceNotFoundInList(t *testing.T) {
	mock := &mockNamingClient{
		selectInstances: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
			return []model.Instance{{Ip: "10.0.0.99", Port: 9999, ServiceName: param.ServiceName}}, nil
		},
	}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	status, err := checker.CheckInstance("test-service")
	assert.NoError(t, err)
	assert.False(t, status.Healthy)
	assert.Error(t, status.Error)
	assert.Contains(t, status.Error.Error(), "not found in service list")
}

func TestHealthChecker_CheckInstance_Found(t *testing.T) {
	mock := &mockNamingClient{
		selectInstances: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
			return []model.Instance{
				{Ip: "10.0.0.1", Port: 8080, ServiceName: param.ServiceName, Healthy: true, Enable: true, Weight: 1.0},
			}, nil
		},
	}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	status, err := checker.CheckInstance("test-service")
	assert.NoError(t, err)
	assert.True(t, status.Healthy)
	assert.NoError(t, status.Error)
}

func TestHealthChecker_CheckAllInstances_Empty(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	statuses, err := checker.CheckAllInstances()
	assert.NoError(t, err)
	assert.Empty(t, statuses)
}

func TestHealthChecker_CheckAllInstances_WithInstances(t *testing.T) {
	mock := &mockNamingClient{
		selectInstances: func(param vo.SelectInstancesParam) ([]model.Instance, error) {
			return []model.Instance{
				{Ip: "10.0.0.1", Port: 8080, ServiceName: param.ServiceName, Healthy: true},
			}, nil
		},
	}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	statuses, err := checker.CheckAllInstances()
	assert.NoError(t, err)
	assert.Len(t, statuses, 1)
	assert.True(t, statuses[0].Healthy)
}

func TestHealthChecker_UpdateInstanceStatus_Success(t *testing.T) {
	// Register with working mock
	regMock := &mockNamingClient{}
	registry := newTestRegistry(regMock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{
		ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080,
		GroupName: "DEFAULT_GROUP", ClusterName: "DEFAULT",
	}
	_ = registry.Register(instance)

	// Swap to verifying mock
	registry.mu.Lock()
	registry.namingCli = &mockNamingClient{registerCalled: true}
	registry.mu.Unlock()

	err := checker.UpdateInstanceStatus("test-service", false)
	assert.NoError(t, err)

	registry.mu.RLock()
	saved := registry.instances["test-service"]
	registry.mu.RUnlock()
	assert.False(t, saved.Healthy)
}

func TestHealthChecker_UpdateInstanceStatus_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.UpdateInstanceStatus("nonexistent", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_UpdateInstanceStatus_Error(t *testing.T) {
	regMock := &mockNamingClient{}
	registry := newTestRegistry(regMock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	errMock := &mockNamingClient{
		registerFunc: func(param vo.RegisterInstanceParam) (bool, error) {
			return false, errors.New("register failed")
		},
	}
	registry.mu.Lock()
	registry.namingCli = errMock
	registry.mu.Unlock()

	err := checker.UpdateInstanceStatus("test-service", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update instance status")
}

func TestHealthChecker_UpdateInstanceStatus_ReturnsFalse(t *testing.T) {
	regMock := &mockNamingClient{}
	registry := newTestRegistry(regMock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	falseMock := &mockNamingClient{
		registerFunc: func(param vo.RegisterInstanceParam) (bool, error) {
			return false, nil
		},
	}
	registry.mu.Lock()
	registry.namingCli = falseMock
	registry.mu.Unlock()

	err := checker.UpdateInstanceStatus("test-service", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
}

func TestHealthChecker_SetInstanceWeight_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{
		ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080,
		GroupName: "DEFAULT_GROUP", ClusterName: "DEFAULT",
	}
	_ = registry.Register(instance)

	err := checker.SetInstanceWeight("test-service", 5.0)
	assert.NoError(t, err)
	assert.True(t, mock.updateCalled)

	registry.mu.RLock()
	saved := registry.instances["test-service"]
	registry.mu.RUnlock()
	assert.Equal(t, 5.0, saved.Weight)
}

func TestHealthChecker_SetInstanceWeight_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.SetInstanceWeight("nonexistent", 5.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_SetInstanceWeight_Error(t *testing.T) {
	mock := &mockNamingClient{
		updateFunc: func(param vo.UpdateInstanceParam) (bool, error) {
			return false, errors.New("update failed")
		},
	}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	err := checker.SetInstanceWeight("test-service", 5.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update instance weight")
}

func TestHealthChecker_SetInstanceMetadata_Success(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{
		ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080,
		GroupName: "DEFAULT_GROUP", ClusterName: "DEFAULT",
		Metadata: map[string]string{"version": "1.0"},
	}
	_ = registry.Register(instance)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"env": "prod"})
	assert.NoError(t, err)

	registry.mu.RLock()
	saved := registry.instances["test-service"]
	registry.mu.RUnlock()
	assert.Equal(t, "1.0", saved.Metadata["version"])
	assert.Equal(t, "prod", saved.Metadata["env"])
}

func TestHealthChecker_SetInstanceMetadata_NilMap(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080, Metadata: nil}
	_ = registry.Register(instance)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"env": "prod"})
	assert.NoError(t, err)

	registry.mu.RLock()
	saved := registry.instances["test-service"]
	registry.mu.RUnlock()
	assert.Equal(t, "prod", saved.Metadata["env"])
}

func TestHealthChecker_SetInstanceMetadata_NotFound(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.SetInstanceMetadata("nonexistent", map[string]string{"env": "prod"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_SetInstanceMetadata_Error(t *testing.T) {
	mock := &mockNamingClient{
		updateFunc: func(param vo.UpdateInstanceParam) (bool, error) {
			return false, errors.New("update failed")
		},
	}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080}
	_ = registry.Register(instance)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"env": "prod"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update instance metadata")
}

func TestHealthChecker_Close(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "svc-1", Ip: "10.0.0.1", Port: 8080}
	_ = checker.StartHeartbeat(instance)

	err := checker.Close()
	assert.NoError(t, err)
	assert.Empty(t, checker.GetHeartbeatTasks())
}

func TestHealthChecker_Close_Empty(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)
	err := checker.Close()
	assert.NoError(t, err)
}

func TestWithHealthCheckInterval(t *testing.T) {
	opts := &HealthCheckOptions{}
	WithHealthCheckInterval(30 * time.Second)(opts)
	assert.Equal(t, 30*time.Second, opts.Interval)
}

func TestWithHealthCheckTimeout(t *testing.T) {
	opts := &HealthCheckOptions{}
	WithHealthCheckTimeout(10 * time.Second)(opts)
	assert.Equal(t, 10*time.Second, opts.Timeout)
}

func TestWithHealthCheckRetry(t *testing.T) {
	opts := &HealthCheckOptions{}
	WithHealthCheckRetry(3, 5*time.Second)(opts)
	assert.Equal(t, 3, opts.RetryCount)
	assert.Equal(t, 5*time.Second, opts.RetryDelay)
}

// =============================================================================
// Utility & Struct Tests
// =============================================================================

func TestGetLocalIP(t *testing.T) {
	ip, err := getLocalIP()
	if err == nil {
		assert.NotEmpty(t, ip)
	}
}

func TestConfigItem_Struct(t *testing.T) {
	item := ConfigItem{DataId: "test", Group: "DEFAULT_GROUP", Content: "content", Tenant: "public"}
	assert.Equal(t, "test", item.DataId)
	assert.Equal(t, "content", item.Content)
}

func TestConfigSearchResult_Struct(t *testing.T) {
	result := ConfigSearchResult{
		TotalCount:     10,
		PageNumber:     1,
		PagesAvailable: 5,
		PageItems:      []ConfigItem{{DataId: "test"}},
	}
	assert.Equal(t, int64(10), result.TotalCount)
	assert.Len(t, result.PageItems, 1)
}

func TestConfigListener_Struct(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener := &ConfigListener{
		DataId:    "test",
		Group:     "DEFAULT_GROUP",
		Namespace: "public",
		Callback:  func(namespace, group, dataId, content string) {},
		Cancel:    cancel,
	}
	assert.Equal(t, "test", listener.DataId)
	assert.NotNil(t, listener.Callback)
	assert.NotNil(t, listener.Cancel)
}

func TestHeartbeatTask_Struct(t *testing.T) {
	task := &HeartbeatTask{
		ServiceName: "test-service",
		Instance:    &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080},
		Interval:    5 * time.Second,
	}
	assert.Equal(t, "test-service", task.ServiceName)
	assert.NotNil(t, task.Instance)
}

func TestHealthStatus_Struct(t *testing.T) {
	status := &HealthStatus{
		ServiceName: "test-service",
		Instance:    &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080},
		Healthy:     true,
		LastCheck:   time.Now(),
		Error:       nil,
	}
	assert.True(t, status.Healthy)
	assert.NoError(t, status.Error)
}

func TestConfigClientOptions_Struct(t *testing.T) {
	opts := ConfigClientOptions{DataId: "test", Group: "DEFAULT_GROUP", Namespace: "public"}
	assert.Equal(t, "test", opts.DataId)
}

func TestDiscoveryOptions_Struct(t *testing.T) {
	opts := DiscoveryOptions{Group: "DEFAULT_GROUP", Clusters: []string{"DEFAULT"}, HealthyOnly: true}
	assert.True(t, opts.HealthyOnly)
}

func TestHealthCheckOptions_Struct(t *testing.T) {
	opts := HealthCheckOptions{Interval: 5 * time.Second, Timeout: 3 * time.Second, RetryCount: 3, RetryDelay: 1 * time.Second}
	assert.Equal(t, 3, opts.RetryCount)
}

func TestHealthChecker_SendHeartbeat_UnregisteredInstance(t *testing.T) {
	mock := &mockNamingClient{}
	registry := newTestRegistry(mock)
	checker := newTestHealthChecker(registry)

	instance := &ServiceInstance{ServiceName: "test-service", Ip: "10.0.0.1", Port: 8080, Ephemeral: true}
	_ = registry.Register(instance)
	_ = checker.StartHeartbeat(instance)

	_ = registry.Deregister("test-service")
	_ = checker.Beat("test-service")
}
