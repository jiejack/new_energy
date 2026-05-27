package nacos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, 1, len(opts.ServerConfigs))
	assert.Equal(t, "127.0.0.1", opts.ServerConfigs[0].IpAddr)
	assert.Equal(t, uint64(8848), opts.ServerConfigs[0].Port)
	assert.Equal(t, "/nacos", opts.ServerConfigs[0].ContextPath)
	assert.Equal(t, "http", opts.ServerConfigs[0].Scheme)
	assert.Equal(t, uint64(5000), opts.ClientConfig.TimeoutMs)
	assert.True(t, opts.ClientConfig.NotLoadCacheAtStart)
	assert.Equal(t, "public", opts.Namespace)
	assert.Equal(t, "DEFAULT_GROUP", opts.Group)
	assert.Equal(t, 1.0, opts.ServiceWeight)
	assert.Equal(t, "DEFAULT", opts.ClusterName)
	assert.True(t, opts.EnableHeartbeat)
	assert.Equal(t, 5*time.Second, opts.HeartbeatInterval)
	assert.Equal(t, 3*time.Second, opts.ConfigListenInterval)
}

func TestApplyOptions_Defaults(t *testing.T) {
	opts := ApplyOptions()
	assert.Equal(t, "public", opts.Namespace)
}

func TestWithServerConfigs(t *testing.T) {
	servers := []ServerConfig{{IpAddr: "10.0.0.1", Port: 8848}}
	opts := ApplyOptions(WithServerConfigs(servers))
	assert.Equal(t, "10.0.0.1", opts.ServerConfigs[0].IpAddr)
}

func TestWithClientConfig(t *testing.T) {
	cfg := ClientConfig{TimeoutMs: 10000, LogLevel: "debug"}
	opts := ApplyOptions(WithClientConfig(cfg))
	assert.Equal(t, uint64(10000), opts.ClientConfig.TimeoutMs)
	assert.Equal(t, "debug", opts.ClientConfig.LogLevel)
}

func TestWithNamespace(t *testing.T) {
	opts := ApplyOptions(WithNamespace("production"))
	assert.Equal(t, "production", opts.Namespace)
	assert.Equal(t, "production", opts.ClientConfig.NamespaceId)
}

func TestWithGroup(t *testing.T) {
	opts := ApplyOptions(WithGroup("MY_GROUP"))
	assert.Equal(t, "MY_GROUP", opts.Group)
}

func TestWithServiceName(t *testing.T) {
	opts := ApplyOptions(WithServiceName("my-service"))
	assert.Equal(t, "my-service", opts.ServiceName)
}

func TestWithServicePort(t *testing.T) {
	opts := ApplyOptions(WithServicePort(8080))
	assert.Equal(t, uint64(8080), opts.ServicePort)
}

func TestWithServiceWeight(t *testing.T) {
	opts := ApplyOptions(WithServiceWeight(2.5))
	assert.Equal(t, 2.5, opts.ServiceWeight)
}

func TestWithServiceMetadata(t *testing.T) {
	meta := map[string]string{"version": "1.0", "env": "prod"}
	opts := ApplyOptions(WithServiceMetadata(meta))
	assert.Equal(t, "1.0", opts.ServiceMetadata["version"])
	assert.Equal(t, "prod", opts.ServiceMetadata["env"])
}

func TestWithClusterName(t *testing.T) {
	opts := ApplyOptions(WithClusterName("BJ"))
	assert.Equal(t, "BJ", opts.ClusterName)
}

func TestWithHeartbeat(t *testing.T) {
	opts := ApplyOptions(WithHeartbeat(false, 10*time.Second))
	assert.False(t, opts.EnableHeartbeat)
	assert.Equal(t, 10*time.Second, opts.HeartbeatInterval)
}

func TestWithConfigListenInterval(t *testing.T) {
	opts := ApplyOptions(WithConfigListenInterval(5 * time.Second))
	assert.Equal(t, 5*time.Second, opts.ConfigListenInterval)
}

func TestWithUsername(t *testing.T) {
	opts := ApplyOptions(WithUsername("admin"))
	assert.Equal(t, "admin", opts.ClientConfig.Username)
}

func TestWithPassword(t *testing.T) {
	opts := ApplyOptions(WithPassword("secret"))
	assert.Equal(t, "secret", opts.ClientConfig.Password)
}

func TestApplyOptions_Multiple(t *testing.T) {
	opts := ApplyOptions(
		WithNamespace("prod"),
		WithServiceName("api-gateway"),
		WithServicePort(9090),
		WithGroup("GATEWAY"),
	)
	assert.Equal(t, "prod", opts.Namespace)
	assert.Equal(t, "api-gateway", opts.ServiceName)
	assert.Equal(t, uint64(9090), opts.ServicePort)
	assert.Equal(t, "GATEWAY", opts.Group)
}

func TestServerConfig_Struct(t *testing.T) {
	sc := ServerConfig{IpAddr: "10.0.0.1", Port: 8848, ContextPath: "/nacos", Scheme: "https"}
	assert.Equal(t, "10.0.0.1", sc.IpAddr)
	assert.Equal(t, uint64(8848), sc.Port)
	assert.Equal(t, "https", sc.Scheme)
}

func TestClientConfig_Struct(t *testing.T) {
	cc := ClientConfig{
		NamespaceId:         "test",
		TimeoutMs:           3000,
		NotLoadCacheAtStart: true,
		Username:            "user",
		Password:            "pass",
		LogLevel:            "warn",
	}
	assert.Equal(t, "test", cc.NamespaceId)
	assert.Equal(t, uint64(3000), cc.TimeoutMs)
}

func TestServiceInstance_Struct(t *testing.T) {
	si := ServiceInstance{
		ServiceName: "test-svc",
		Ip:         "10.0.0.1",
		Port:       8080,
		Weight:     1.0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{"version": "1.0"},
		ClusterName: "DEFAULT",
		GroupName:  "DEFAULT_GROUP",
		Ephemeral:  true,
	}
	assert.Equal(t, "test-svc", si.ServiceName)
	assert.True(t, si.Enable)
	assert.True(t, si.Healthy)
}

func TestConfigOptions_Struct(t *testing.T) {
	co := ConfigOptions{
		DataId:    "application.yaml",
		Group:     "DEFAULT_GROUP",
		Namespace: "public",
		Content:   "key: value",
		Tag:       "v1",
		AppName:   "test-app",
	}
	assert.Equal(t, "application.yaml", co.DataId)
	assert.Equal(t, "key: value", co.Content)
}

func TestRegistryOptions_Struct(t *testing.T) {
	ro := RegistryOptions{
		ServiceName: "my-service",
		Group:       "DEFAULT_GROUP",
		Namespace:   "public",
		Clusters:    []string{"DEFAULT"},
		HealthyOnly: true,
		Enable:      true,
	}
	assert.Equal(t, "my-service", ro.ServiceName)
	assert.True(t, ro.HealthyOnly)
}

func TestOptions_Struct(t *testing.T) {
	opts := &Options{
		ServerConfigs:       []ServerConfig{{IpAddr: "10.0.0.1", Port: 8848}},
		ClientConfig:        ClientConfig{TimeoutMs: 5000},
		Namespace:           "test",
		Group:              "TEST_GROUP",
		ServiceName:        "test-service",
		ServicePort:        8080,
		ServiceWeight:      1.0,
		ServiceMetadata:    map[string]string{"env": "test"},
		ClusterName:        "DEFAULT",
		EnableHeartbeat:    true,
		HeartbeatInterval:  5 * time.Second,
		ConfigListenInterval: 3 * time.Second,
	}
	assert.Equal(t, "test", opts.Namespace)
	assert.Equal(t, "test-service", opts.ServiceName)
}
