package nacos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockConfigClient struct {
	mock.Mock
}

func (m *mockConfigClient) GetConfig(param vo.ConfigParam) (string, error) {
	args := m.Called(param)
	return args.String(0), args.Error(1)
}

func (m *mockConfigClient) PublishConfig(param vo.ConfigParam) (bool, error) {
	args := m.Called(param)
	return args.Bool(0), args.Error(1)
}

func (m *mockConfigClient) DeleteConfig(param vo.ConfigParam) (bool, error) {
	args := m.Called(param)
	return args.Bool(0), args.Error(1)
}

func (m *mockConfigClient) ListenConfig(params vo.ConfigParam) error {
	args := m.Called(params)
	return args.Error(0)
}

func (m *mockConfigClient) CancelListenConfig(params vo.ConfigParam) error {
	args := m.Called(params)
	return args.Error(0)
}

func (m *mockConfigClient) SearchConfig(param vo.SearchConfigParam) (*model.ConfigPage, error) {
	args := m.Called(param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ConfigPage), args.Error(1)
}

func (m *mockConfigClient) CloseClient() {
	m.Called()
}

type mockNamingClient struct {
	mock.Mock
}

func (m *mockNamingClient) RegisterInstance(param vo.RegisterInstanceParam) (bool, error) {
	args := m.Called(param)
	return args.Bool(0), args.Error(1)
}

func (m *mockNamingClient) BatchRegisterInstance(param vo.BatchRegisterInstanceParam) (bool, error) {
	args := m.Called(param)
	return args.Bool(0), args.Error(1)
}

func (m *mockNamingClient) DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error) {
	args := m.Called(param)
	return args.Bool(0), args.Error(1)
}

func (m *mockNamingClient) UpdateInstance(param vo.UpdateInstanceParam) (bool, error) {
	args := m.Called(param)
	return args.Bool(0), args.Error(1)
}

func (m *mockNamingClient) GetService(param vo.GetServiceParam) (model.Service, error) {
	args := m.Called(param)
	return args.Get(0).(model.Service), args.Error(1)
}

func (m *mockNamingClient) SelectAllInstances(param vo.SelectAllInstancesParam) ([]model.Instance, error) {
	args := m.Called(param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Instance), args.Error(1)
}

func (m *mockNamingClient) SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error) {
	args := m.Called(param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Instance), args.Error(1)
}

func (m *mockNamingClient) SelectOneHealthyInstance(param vo.SelectOneHealthInstanceParam) (*model.Instance, error) {
	args := m.Called(param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Instance), args.Error(1)
}

func (m *mockNamingClient) Subscribe(param *vo.SubscribeParam) error {
	args := m.Called(param)
	return args.Error(0)
}

func (m *mockNamingClient) Unsubscribe(param *vo.SubscribeParam) error {
	args := m.Called(param)
	return args.Error(0)
}

func (m *mockNamingClient) GetAllServicesInfo(param vo.GetAllServiceInfoParam) (model.ServiceList, error) {
	args := m.Called(param)
	return args.Get(0).(model.ServiceList), args.Error(1)
}

func (m *mockNamingClient) ServerHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockNamingClient) CloseClient() {
	m.Called()
}

func newTestConfigClient(cli config_client.IConfigClient) *ConfigClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConfigClient{
		options:   ApplyOptions(),
		configCli: cli,
		listeners: make(map[string]*ConfigListener),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func newTestRegistry(cli naming_client.INamingClient) *Registry {
	ctx, cancel := context.WithCancel(context.Background())
	return &Registry{
		options:   ApplyOptions(),
		namingCli: cli,
		instances: make(map[string]*ServiceInstance),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func TestConfigClientOptions_Funcs(t *testing.T) {
	cco := &ConfigClientOptions{}
	WithConfigDataId("test-dataId")(cco)
	assert.Equal(t, "test-dataId", cco.DataId)

	WithConfigGroup("test-group")(cco)
	assert.Equal(t, "test-group", cco.Group)

	WithConfigNamespace("test-namespace")(cco)
	assert.Equal(t, "test-namespace", cco.Namespace)
}

func TestDiscoveryOptions_Funcs(t *testing.T) {
	opts := &DiscoveryOptions{}
	WithDiscoveryGroup("test-group")(opts)
	assert.Equal(t, "test-group", opts.Group)

	WithDiscoveryClusters([]string{"cluster1", "cluster2"})(opts)
	assert.Equal(t, []string{"cluster1", "cluster2"}, opts.Clusters)

	WithDiscoveryHealthyOnly(false)(opts)
	assert.False(t, opts.HealthyOnly)
}

func TestConfigClient_GetConfig(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("GetConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return("test-content", nil)

	content, err := client.GetConfig("test-data", "DEFAULT_GROUP")
	assert.NoError(t, err)
	assert.Equal(t, "test-content", content)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_GetConfig_DefaultGroup(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("GetConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return("content", nil)

	content, err := client.GetConfig("test-data", "")
	assert.NoError(t, err)
	assert.Equal(t, "content", content)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_GetConfig_Error(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("GetConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return("", fmt.Errorf("not found"))

	content, err := client.GetConfig("test-data", "DEFAULT_GROUP")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get config")
	assert.Empty(t, content)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_PublishConfig(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("PublishConfig", vo.ConfigParam{
		DataId:  "test-data",
		Group:   "DEFAULT_GROUP",
		Content: "test-content",
	}).Return(true, nil)

	success, err := client.PublishConfig("test-data", "DEFAULT_GROUP", "test-content")
	assert.NoError(t, err)
	assert.True(t, success)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_PublishConfig_EmptyGroup(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("PublishConfig", vo.ConfigParam{
		DataId:  "test-data",
		Group:   "DEFAULT_GROUP",
		Content: "test-content",
	}).Return(true, nil)

	success, err := client.PublishConfig("test-data", "", "test-content")
	assert.NoError(t, err)
	assert.True(t, success)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_PublishConfig_Error(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("PublishConfig", vo.ConfigParam{
		DataId:  "test-data",
		Group:   "DEFAULT_GROUP",
		Content: "test-content",
	}).Return(false, fmt.Errorf("publish error"))

	success, err := client.PublishConfig("test-data", "DEFAULT_GROUP", "test-content")
	assert.Error(t, err)
	assert.False(t, success)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_DeleteConfig(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("DeleteConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return(true, nil)

	success, err := client.DeleteConfig("test-data", "DEFAULT_GROUP")
	assert.NoError(t, err)
	assert.True(t, success)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_DeleteConfig_Error(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("DeleteConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return(false, fmt.Errorf("delete error"))

	success, err := client.DeleteConfig("test-data", "DEFAULT_GROUP")
	assert.Error(t, err)
	assert.False(t, success)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_ListenConfig(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	var capturedOnChange vo.Listener
	mockCli.On("ListenConfig", mock.Anything).Run(func(args mock.Arguments) {
		param := args.Get(0).(vo.ConfigParam)
		capturedOnChange = param.OnChange
	}).Return(nil)

	callbackCalled := false
	err := client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {
		callbackCalled = true
	})
	assert.NoError(t, err)

	assert.NotNil(t, capturedOnChange)
	capturedOnChange("ns", "DEFAULT_GROUP", "test-data", "new-content")
	assert.True(t, callbackCalled)

	assert.Contains(t, client.listeners, "test-data@DEFAULT_GROUP")
	mockCli.AssertExpectations(t)
}

func TestConfigClient_ListenConfig_EmptyGroup(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("ListenConfig", mock.Anything).Return(nil)

	err := client.ListenConfig("test-data", "", func(namespace, group, dataId, content string) {})
	assert.NoError(t, err)
	assert.Contains(t, client.listeners, "test-data@DEFAULT_GROUP")
	mockCli.AssertExpectations(t)
}

func TestConfigClient_ListenConfig_Duplicate(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("ListenConfig", mock.Anything).Return(nil)

	err := client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.NoError(t, err)

	err = client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfigClient_ListenConfig_Error(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("ListenConfig", mock.Anything).Return(fmt.Errorf("listen error"))

	err := client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to listen config")
	assert.NotContains(t, client.listeners, "test-data@DEFAULT_GROUP")
	mockCli.AssertExpectations(t)
}

func TestConfigClient_CancelListenConfig(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("ListenConfig", mock.Anything).Return(nil)
	mockCli.On("CancelListenConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return(nil)

	err := client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.NoError(t, err)

	err = client.CancelListenConfig("test-data", "DEFAULT_GROUP")
	assert.NoError(t, err)
	assert.NotContains(t, client.listeners, "test-data@DEFAULT_GROUP")
	mockCli.AssertExpectations(t)
}

func TestConfigClient_CancelListenConfig_NotFound(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	err := client.CancelListenConfig("nonexistent", "DEFAULT_GROUP")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConfigClient_CancelListenConfig_Error(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("ListenConfig", mock.Anything).Return(nil)
	mockCli.On("CancelListenConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return(fmt.Errorf("cancel error"))

	err := client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.NoError(t, err)

	err = client.CancelListenConfig("test-data", "DEFAULT_GROUP")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to cancel listen config")
	mockCli.AssertExpectations(t)
}

func TestConfigClient_SearchConfig(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("SearchConfig", vo.SearchConfigParam{
		DataId:   "test-data",
		Group:    "DEFAULT_GROUP",
		PageNo:   1,
		PageSize: 10,
	}).Return(&model.ConfigPage{
		TotalCount:     1,
		PageNumber:     1,
		PagesAvailable: 1,
		PageItems: []model.ConfigItem{
			{DataId: "test-data", Group: "DEFAULT_GROUP", Content: "content", Tenant: "public"},
		},
	}, nil)

	result, err := client.SearchConfig("test-data", "DEFAULT_GROUP", 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.TotalCount)
	assert.Equal(t, int32(1), result.PageNumber)
	assert.Len(t, result.PageItems, 1)
	assert.Equal(t, "test-data", result.PageItems[0].DataId)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_SearchConfig_EmptyGroup(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("SearchConfig", vo.SearchConfigParam{
		DataId:   "test-data",
		Group:    "DEFAULT_GROUP",
		PageNo:   1,
		PageSize: 10,
	}).Return(&model.ConfigPage{}, nil)

	result, err := client.SearchConfig("test-data", "", 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_SearchConfig_Error(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("SearchConfig", mock.Anything).Return(nil, fmt.Errorf("search error"))

	result, err := client.SearchConfig("test-data", "DEFAULT_GROUP", 1, 10)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to search config")
	mockCli.AssertExpectations(t)
}

func TestConfigClient_SearchConfig_EmptyPageItems(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("SearchConfig", mock.Anything).Return(&model.ConfigPage{
		TotalCount:     0,
		PageNumber:     1,
		PagesAvailable: 0,
		PageItems:      []model.ConfigItem{},
	}, nil)

	result, err := client.SearchConfig("test-data", "DEFAULT_GROUP", 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.PageItems)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_GetConfigAndListen(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("GetConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return("initial-content", nil)

	mockCli.On("ListenConfig", mock.Anything).Return(nil)

	content, err := client.GetConfigAndListen("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.NoError(t, err)
	assert.Equal(t, "initial-content", content)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_GetConfigAndListen_GetError(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("GetConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return("", fmt.Errorf("get error"))

	content, err := client.GetConfigAndListen("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.Error(t, err)
	assert.Empty(t, content)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_GetConfigAndListen_ListenError(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("GetConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return("content", nil)

	mockCli.On("ListenConfig", mock.Anything).Return(fmt.Errorf("listen error"))

	content, err := client.GetConfigAndListen("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.Error(t, err)
	assert.Empty(t, content)
	mockCli.AssertExpectations(t)
}

func TestConfigClient_Close(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	mockCli.On("ListenConfig", mock.Anything).Return(nil)
	mockCli.On("CancelListenConfig", vo.ConfigParam{
		DataId: "test-data",
		Group:  "DEFAULT_GROUP",
	}).Return(nil)

	err := client.ListenConfig("test-data", "DEFAULT_GROUP", func(namespace, group, dataId, content string) {})
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
	assert.Error(t, client.ctx.Err())
	mockCli.AssertExpectations(t)
}

func TestConfigClient_Close_NoListeners(t *testing.T) {
	mockCli := new(mockConfigClient)
	client := newTestConfigClient(mockCli)

	err := client.Close()
	assert.NoError(t, err)
	assert.Error(t, client.ctx.Err())
}

func TestRegistry_Register(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{"version": "1.0"},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}

	mockCli.On("RegisterInstance", vo.RegisterInstanceParam{
		Ip:          "192.168.1.1",
		Port:        uint64(8080),
		ServiceName: "test-service",
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{"version": "1.0"},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
	}).Return(true, nil)

	err := registry.Register(instance)
	assert.NoError(t, err)
	assert.True(t, registry.IsRegistered())
	assert.Contains(t, registry.instances, "test-service")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Register_NilInstance(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	err := registry.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance cannot be nil")
}

func TestRegistry_Register_AutoIP(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)

	err := registry.Register(instance)
	assert.NoError(t, err)
	assert.NotEmpty(t, instance.Ip)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Register_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(false, fmt.Errorf("register error"))

	err := registry.Register(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register instance")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Register_FalseResult(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(false, nil)

	err := registry.Register(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Deregister(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("DeregisterInstance", vo.DeregisterInstanceParam{
		Ip:          "192.168.1.1",
		Port:        uint64(8080),
		ServiceName: "test-service",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}).Return(true, nil)

	registry.Register(instance)
	err := registry.Deregister("test-service")
	assert.NoError(t, err)
	assert.NotContains(t, registry.instances, "test-service")
	assert.False(t, registry.IsRegistered())
	mockCli.AssertExpectations(t)
}

func TestRegistry_Deregister_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	err := registry.Deregister("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_Deregister_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("DeregisterInstance", mock.Anything).Return(false, fmt.Errorf("deregister error"))

	registry.Register(instance)
	err := registry.Deregister("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deregister instance")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Deregister_FalseResult(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("DeregisterInstance", mock.Anything).Return(false, nil)

	registry.Register(instance)
	err := registry.Deregister("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Discover(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("SelectInstances", vo.SelectInstancesParam{
		ServiceName: "test-service",
		GroupName:   "DEFAULT_GROUP",
		Clusters:    []string{"DEFAULT"},
		HealthyOnly: true,
	}).Return([]model.Instance{
		{
			ServiceName: "test-service",
			Ip:          "192.168.1.1",
			Port:        uint64(8080),
			Weight:      1.0,
			Enable:      true,
			Healthy:     true,
			Metadata:    map[string]string{"version": "1.0"},
			Ephemeral:   true,
		},
	}, nil)

	instances, err := registry.Discover("test-service")
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "test-service", instances[0].ServiceName)
	assert.Equal(t, "192.168.1.1", instances[0].Ip)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Discover_WithOptions(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("SelectInstances", vo.SelectInstancesParam{
		ServiceName: "test-service",
		GroupName:   "custom-group",
		Clusters:    []string{"cluster1"},
		HealthyOnly: false,
	}).Return([]model.Instance{}, nil)

	instances, err := registry.Discover("test-service",
		WithDiscoveryGroup("custom-group"),
		WithDiscoveryClusters([]string{"cluster1"}),
		WithDiscoveryHealthyOnly(false),
	)
	assert.NoError(t, err)
	assert.Len(t, instances, 0)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Discover_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("SelectInstances", mock.Anything).Return(nil, fmt.Errorf("discover error"))

	instances, err := registry.Discover("test-service")
	assert.Error(t, err)
	assert.Nil(t, instances)
	assert.Contains(t, err.Error(), "failed to discover service")
	mockCli.AssertExpectations(t)
}

func TestRegistry_DiscoverOne(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("SelectOneHealthyInstance", vo.SelectOneHealthInstanceParam{
		ServiceName: "test-service",
		GroupName:   "DEFAULT_GROUP",
		Clusters:    []string{"DEFAULT"},
	}).Return(&model.Instance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        uint64(8080),
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
	}, nil)

	instance, err := registry.DiscoverOne("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "test-service", instance.ServiceName)
	mockCli.AssertExpectations(t)
}

func TestRegistry_DiscoverOne_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("SelectOneHealthyInstance", mock.Anything).Return(nil, fmt.Errorf("no instance"))

	instance, err := registry.DiscoverOne("test-service")
	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.Contains(t, err.Error(), "failed to discover one instance")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Subscribe(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	var capturedCallback func(services []model.Instance, err error)
	mockCli.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		param := args.Get(0).(*vo.SubscribeParam)
		capturedCallback = param.SubscribeCallback
	}).Return(nil)

	callbackCalled := false
	err := registry.Subscribe("test-service", func(instances []*ServiceInstance) {
		callbackCalled = true
		assert.Len(t, instances, 1)
	})
	assert.NoError(t, err)

	assert.NotNil(t, capturedCallback)
	capturedCallback([]model.Instance{
		{ServiceName: "test-service", Ip: "192.168.1.1", Port: 8080, Healthy: true},
	}, nil)
	assert.True(t, callbackCalled)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Subscribe_CallbackError(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	var capturedCallback func(services []model.Instance, err error)
	mockCli.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		param := args.Get(0).(*vo.SubscribeParam)
		capturedCallback = param.SubscribeCallback
	}).Return(nil)

	callbackCalled := false
	err := registry.Subscribe("test-service", func(instances []*ServiceInstance) {
		callbackCalled = true
	})
	assert.NoError(t, err)

	capturedCallback(nil, fmt.Errorf("service error"))
	assert.False(t, callbackCalled)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Subscribe_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("Subscribe", mock.Anything).Return(fmt.Errorf("subscribe error"))

	err := registry.Subscribe("test-service", func(instances []*ServiceInstance) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to subscribe service")
	mockCli.AssertExpectations(t)
}

func TestRegistry_Unsubscribe(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("Unsubscribe", mock.Anything).Return(nil)

	err := registry.Unsubscribe("test-service")
	assert.NoError(t, err)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Unsubscribe_WithOptions(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("Unsubscribe", mock.MatchedBy(func(param *vo.SubscribeParam) bool {
		return param.ServiceName == "test-service" && param.GroupName == "custom-group"
	})).Return(nil)

	err := registry.Unsubscribe("test-service", WithDiscoveryGroup("custom-group"))
	assert.NoError(t, err)
	mockCli.AssertExpectations(t)
}

func TestRegistry_Unsubscribe_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("Unsubscribe", mock.Anything).Return(fmt.Errorf("unsubscribe error"))

	err := registry.Unsubscribe("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unsubscribe service")
	mockCli.AssertExpectations(t)
}

func TestRegistry_GetAllServices(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	services, err := registry.GetAllServices()
	assert.NoError(t, err)
	assert.NotNil(t, services)
}

func TestRegistry_Close(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	}

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("DeregisterInstance", mock.Anything).Return(true, nil)

	registry.Register(instance)
	err := registry.Close()
	assert.NoError(t, err)
	assert.Error(t, registry.ctx.Err())
	mockCli.AssertExpectations(t)
}

func TestRegistry_Close_NoInstances(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	err := registry.Close()
	assert.NoError(t, err)
}

func TestRegistry_IsRegistered(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	assert.False(t, registry.IsRegistered())

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	}
	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(instance)
	assert.True(t, registry.IsRegistered())
}

func TestGetLocalIP(t *testing.T) {
	ip, err := getLocalIP()
	if err != nil {
		assert.Equal(t, "no valid local IP address found", err.Error())
	} else {
		assert.NotEmpty(t, ip)
	}
}

func TestHealthChecker_New(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, err := NewHealthChecker(registry)
	assert.NoError(t, err)
	assert.NotNil(t, checker)
	assert.Equal(t, registry, checker.registry)
}

func TestHealthChecker_New_NilRegistry(t *testing.T) {
	checker, err := NewHealthChecker(nil)
	assert.Error(t, err)
	assert.Nil(t, checker)
	assert.Contains(t, err.Error(), "registry cannot be nil")
}

func TestHealthChecker_New_WithOptions(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, err := NewHealthChecker(registry,
		WithHeartbeat(true, 10*time.Second),
	)
	assert.NoError(t, err)
	assert.NotNil(t, checker)
	assert.Equal(t, 10*time.Second, checker.options.HeartbeatInterval)
}

func TestHealthChecker_StartHeartbeat(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry, WithHeartbeat(true, 5*time.Second))

	err := checker.StartHeartbeat(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})
	assert.NoError(t, err)
	assert.Contains(t, checker.GetHeartbeatTasks(), "test-service")

	checker.Close()
}

func TestHealthChecker_StartHeartbeat_NilInstance(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.StartHeartbeat(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance cannot be nil")
}

func TestHealthChecker_StartHeartbeat_Duplicate(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry, WithHeartbeat(true, 5*time.Second))

	instance := &ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	}
	err := checker.StartHeartbeat(instance)
	assert.NoError(t, err)

	err = checker.StartHeartbeat(instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	checker.Close()
}

func TestHealthChecker_StopHeartbeat(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry, WithHeartbeat(true, 5*time.Second))

	checker.StartHeartbeat(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	err := checker.StopHeartbeat("test-service")
	assert.NoError(t, err)
	assert.NotContains(t, checker.GetHeartbeatTasks(), "test-service")
}

func TestHealthChecker_StopHeartbeat_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.StopHeartbeat("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_CheckInstance(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("SelectInstances", mock.Anything).Return([]model.Instance{
		{
			ServiceName: "test-service",
			Ip:          "192.168.1.1",
			Port:        uint64(8080),
			Healthy:     true,
		},
	}, nil)

	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry)

	status, err := checker.CheckInstance("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Healthy)
	assert.Equal(t, "test-service", status.ServiceName)
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_CheckInstance_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	status, err := checker.CheckInstance("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_CheckInstance_DiscoverError(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("SelectInstances", mock.Anything).Return(nil, fmt.Errorf("discover error"))

	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry)

	status, err := checker.CheckInstance("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.Healthy)
	assert.NotNil(t, status.Error)
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_CheckInstance_NotInServiceList(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("SelectInstances", mock.Anything).Return([]model.Instance{
		{
			ServiceName: "test-service",
			Ip:          "192.168.1.2",
			Port:        uint64(8080),
			Healthy:     true,
		},
	}, nil)

	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry)

	status, err := checker.CheckInstance("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.Healthy)
	assert.Contains(t, status.Error.Error(), "instance not found in service list")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_CheckAllInstances(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	mockCli.On("SelectInstances", mock.Anything).Return([]model.Instance{
		{
			ServiceName: "svc1",
			Ip:          "192.168.1.1",
			Port:        uint64(8080),
			Healthy:     true,
		},
	}, nil)

	registry.Register(&ServiceInstance{
		ServiceName: "svc1",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry)

	statuses, err := checker.CheckAllInstances()
	assert.NoError(t, err)
	assert.Len(t, statuses, 1)
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_UpdateInstanceStatus(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.MatchedBy(func(param vo.RegisterInstanceParam) bool {
		return param.Healthy
	})).Return(true, nil)

	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("RegisterInstance", mock.MatchedBy(func(param vo.RegisterInstanceParam) bool {
		return param.ServiceName == "test-service" && !param.Healthy
	})).Return(true, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.UpdateInstanceStatus("test-service", false)
	assert.NoError(t, err)
	assert.False(t, registry.instances["test-service"].Healthy)
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_UpdateInstanceStatus_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.UpdateInstanceStatus("nonexistent", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_UpdateInstanceStatus_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.MatchedBy(func(param vo.RegisterInstanceParam) bool {
		return param.Healthy
	})).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("RegisterInstance", mock.MatchedBy(func(param vo.RegisterInstanceParam) bool {
		return !param.Healthy
	})).Return(false, fmt.Errorf("update error"))

	checker, _ := NewHealthChecker(registry)

	err := checker.UpdateInstanceStatus("test-service", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update instance status")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_UpdateInstanceStatus_FalseResult(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.MatchedBy(func(param vo.RegisterInstanceParam) bool {
		return param.Healthy
	})).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("RegisterInstance", mock.MatchedBy(func(param vo.RegisterInstanceParam) bool {
		return !param.Healthy
	})).Return(false, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.UpdateInstanceStatus("test-service", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceWeight(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(true, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceWeight("test-service", 2.0)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, registry.instances["test-service"].Weight)
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceWeight_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceWeight("nonexistent", 2.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_SetInstanceWeight_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(false, fmt.Errorf("update error"))

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceWeight("test-service", 2.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update instance weight")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceWeight_FalseResult(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(false, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceWeight("test-service", 2.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceMetadata(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    map[string]string{"existing": "value"},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(true, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"new-key": "new-value"})
	assert.NoError(t, err)
	assert.Equal(t, "value", registry.instances["test-service"].Metadata["existing"])
	assert.Equal(t, "new-value", registry.instances["test-service"].Metadata["new-key"])
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceMetadata_NilMetadata(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    nil,
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(true, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Equal(t, "value", registry.instances["test-service"].Metadata["key"])
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceMetadata_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceMetadata("nonexistent", map[string]string{"key": "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_SetInstanceMetadata_Error(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(false, fmt.Errorf("update error"))

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"key": "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update instance metadata")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_SetInstanceMetadata_FalseResult(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	mockCli.On("UpdateInstance", mock.Anything).Return(false, nil)

	checker, _ := NewHealthChecker(registry)

	err := checker.SetInstanceMetadata("test-service", map[string]string{"key": "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation returned false")
	mockCli.AssertExpectations(t)
}

func TestHealthChecker_Beat(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry, WithHeartbeat(true, 5*time.Second))
	checker.logger = zap.NewNop()
	checker.StartHeartbeat(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
		Ephemeral:   true,
	})

	err := checker.Beat("test-service")
	assert.NoError(t, err)

	checker.Close()
}

func TestHealthChecker_Beat_NotFound(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.Beat("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_GetHeartbeatTasks(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{ServiceName: "svc1", Ip: "192.168.1.1", Port: 8080})
	registry.Register(&ServiceInstance{ServiceName: "svc2", Ip: "192.168.1.2", Port: 8080})

	checker, _ := NewHealthChecker(registry, WithHeartbeat(true, 5*time.Second))
	checker.StartHeartbeat(&ServiceInstance{ServiceName: "svc1", Ip: "192.168.1.1", Port: 8080})
	checker.StartHeartbeat(&ServiceInstance{ServiceName: "svc2", Ip: "192.168.1.2", Port: 8080})

	tasks := checker.GetHeartbeatTasks()
	assert.Len(t, tasks, 2)
	assert.Contains(t, tasks, "svc1")
	assert.Contains(t, tasks, "svc2")

	checker.Close()
}

func TestHealthChecker_Close(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	mockCli.On("RegisterInstance", mock.Anything).Return(true, nil)
	registry.Register(&ServiceInstance{
		ServiceName: "test-service",
		Ip:          "192.168.1.1",
		Port:        8080,
	})

	checker, _ := NewHealthChecker(registry, WithHeartbeat(true, 5*time.Second))
	checker.StartHeartbeat(&ServiceInstance{ServiceName: "test-service", Ip: "192.168.1.1", Port: 8080})

	err := checker.Close()
	assert.NoError(t, err)
	assert.Empty(t, checker.GetHeartbeatTasks())
}

func TestHealthChecker_Close_NoTasks(t *testing.T) {
	mockCli := new(mockNamingClient)
	registry := newTestRegistry(mockCli)

	checker, _ := NewHealthChecker(registry)

	err := checker.Close()
	assert.NoError(t, err)
}

func TestHealthCheckOptions_Funcs(t *testing.T) {
	opts := &HealthCheckOptions{}
	WithHealthCheckInterval(10 * time.Second)(opts)
	assert.Equal(t, 10*time.Second, opts.Interval)

	WithHealthCheckTimeout(5 * time.Second)(opts)
	assert.Equal(t, 5*time.Second, opts.Timeout)

	WithHealthCheckRetry(3, 1*time.Second)(opts)
	assert.Equal(t, 3, opts.RetryCount)
	assert.Equal(t, 1*time.Second, opts.RetryDelay)
}

func TestConfigItem_Struct(t *testing.T) {
	item := ConfigItem{
		DataId:  "test-data",
		Group:   "test-group",
		Content: "test-content",
		Tenant:  "test-tenant",
	}
	assert.Equal(t, "test-data", item.DataId)
	assert.Equal(t, "test-group", item.Group)
	assert.Equal(t, "test-content", item.Content)
	assert.Equal(t, "test-tenant", item.Tenant)
}

func TestConfigSearchResult_Struct(t *testing.T) {
	result := ConfigSearchResult{
		TotalCount:     10,
		PageNumber:     1,
		PagesAvailable: 2,
		PageItems:      []ConfigItem{{DataId: "test"}},
	}
	assert.Equal(t, int64(10), result.TotalCount)
	assert.Equal(t, int32(1), result.PageNumber)
	assert.Equal(t, int32(2), result.PagesAvailable)
	assert.Len(t, result.PageItems, 1)
}

func TestHealthStatus_Struct(t *testing.T) {
	status := HealthStatus{
		ServiceName: "test-service",
		Instance:    &ServiceInstance{ServiceName: "test-service"},
		Healthy:     true,
		LastCheck:   time.Now(),
		Error:       nil,
	}
	assert.Equal(t, "test-service", status.ServiceName)
	assert.True(t, status.Healthy)
}

func TestConfigListener_Struct(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener := &ConfigListener{
		DataId:    "test-data",
		Group:     "test-group",
		Namespace: "test-ns",
		Callback:  func(namespace, group, dataId, content string) {},
		Cancel:    cancel,
	}
	assert.Equal(t, "test-data", listener.DataId)
	assert.Equal(t, "test-group", listener.Group)
	assert.Equal(t, "test-ns", listener.Namespace)
	assert.NotNil(t, listener.Callback)
	assert.NotNil(t, listener.Cancel)
}

func TestHeartbeatTask_Struct(t *testing.T) {
	task := &HeartbeatTask{
		ServiceName: "svc",
		Instance:    &ServiceInstance{ServiceName: "svc"},
		Interval:    5 * time.Second,
	}
	assert.Equal(t, "svc", task.ServiceName)
	assert.Equal(t, 5*time.Second, task.Interval)
}

func TestConfigClientOptions_Struct(t *testing.T) {
	cco := ConfigClientOptions{
		DataId:    "data1",
		Group:     "group1",
		Namespace: "ns1",
	}
	assert.Equal(t, "data1", cco.DataId)
	assert.Equal(t, "group1", cco.Group)
	assert.Equal(t, "ns1", cco.Namespace)
}

func TestHealthCheckOptions_Struct(t *testing.T) {
	opts := HealthCheckOptions{
		Interval:   10 * time.Second,
		Timeout:    5 * time.Second,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	}
	assert.Equal(t, 10*time.Second, opts.Interval)
	assert.Equal(t, 5*time.Second, opts.Timeout)
	assert.Equal(t, 3, opts.RetryCount)
	assert.Equal(t, 1*time.Second, opts.RetryDelay)
}
