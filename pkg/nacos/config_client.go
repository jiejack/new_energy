package nacos

import (
	"context"
	"fmt"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// ConfigClient Nacos配置中心客户端
type ConfigClient struct {
	options    *Options
	configCli  config_client.IConfigClient
	listeners  map[string]*ConfigListener
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// ConfigListener 配置监听器
type ConfigListener struct {
	DataId    string
	Group     string
	Namespace string
	Callback  ConfigChangeCallback
	Cancel    context.CancelFunc
}

// ConfigChangeCallback 配置变更回调函数
type ConfigChangeCallback func(namespace, group, dataId, content string)

// NewConfigClient 创建新的配置中心客户端
func NewConfigClient(opts ...Option) (*ConfigClient, error) {
	options := ApplyOptions(opts...)

	// 构建服务器配置
	serverConfigs := make([]constant.ServerConfig, 0, len(options.ServerConfigs))
	for _, sc := range options.ServerConfigs {
		serverConfigs = append(serverConfigs, *constant.NewServerConfig(
			sc.IpAddr,
			sc.Port,
			constant.WithContextPath(sc.ContextPath),
			constant.WithScheme(sc.Scheme),
		))
	}

	// 构建客户端配置
	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId(options.ClientConfig.NamespaceId),
		constant.WithTimeoutMs(options.ClientConfig.TimeoutMs),
		constant.WithNotLoadCacheAtStart(options.ClientConfig.NotLoadCacheAtStart),
		constant.WithUpdateCacheWhenEmpty(options.ClientConfig.UpdateCacheWhenEmpty),
		constant.WithUsername(options.ClientConfig.Username),
		constant.WithPassword(options.ClientConfig.Password),
		constant.WithLogLevel(options.ClientConfig.LogLevel),
	)

	// 创建配置客户端
	configCli, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConfigClient{
		options:   options,
		configCli: configCli,
		listeners: make(map[string]*ConfigListener),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// GetConfig 获取配置
func (c *ConfigClient) GetConfig(dataId, group string) (string, error) {
	if group == "" {
		group = c.options.Group
	}

	content, err := c.configCli.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get config [dataId: %s, group: %s]: %w", dataId, group, err)
	}

	return content, nil
}

// GetConfigWithNamespace 获取指定命名空间的配置
func (c *ConfigClient) GetConfigWithNamespace(dataId, group, namespace string) (string, error) {
	if group == "" {
		group = c.options.Group
	}

	// 临时创建指定命名空间的客户端
	serverConfigs := make([]constant.ServerConfig, 0, len(c.options.ServerConfigs))
	for _, sc := range c.options.ServerConfigs {
		serverConfigs = append(serverConfigs, *constant.NewServerConfig(
			sc.IpAddr,
			sc.Port,
			constant.WithContextPath(sc.ContextPath),
			constant.WithScheme(sc.Scheme),
		))
	}

	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId(namespace),
		constant.WithTimeoutMs(c.options.ClientConfig.TimeoutMs),
		constant.WithNotLoadCacheAtStart(c.options.ClientConfig.NotLoadCacheAtStart),
		constant.WithUpdateCacheWhenEmpty(c.options.ClientConfig.UpdateCacheWhenEmpty),
		constant.WithUsername(c.options.ClientConfig.Username),
		constant.WithPassword(c.options.ClientConfig.Password),
		constant.WithLogLevel(c.options.ClientConfig.LogLevel),
	)

	tempCli, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create temp config client: %w", err)
	}

	content, err := tempCli.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get config [dataId: %s, group: %s, namespace: %s]: %w", dataId, group, namespace, err)
	}

	return content, nil
}

// PublishConfig 发布配置
func (c *ConfigClient) PublishConfig(dataId, group, content string) (bool, error) {
	if group == "" {
		group = c.options.Group
	}

	success, err := c.configCli.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
	})

	if err != nil {
		return false, fmt.Errorf("failed to publish config [dataId: %s, group: %s]: %w", dataId, group, err)
	}

	return success, nil
}

// DeleteConfig 删除配置
func (c *ConfigClient) DeleteConfig(dataId, group string) (bool, error) {
	if group == "" {
		group = c.options.Group
	}

	success, err := c.configCli.DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		return false, fmt.Errorf("failed to delete config [dataId: %s, group: %s]: %w", dataId, group, err)
	}

	return success, nil
}

// ListenConfig 监听配置变更
func (c *ConfigClient) ListenConfig(dataId, group string, callback ConfigChangeCallback) error {
	if group == "" {
		group = c.options.Group
	}

	listenerKey := fmt.Sprintf("%s@%s", dataId, group)

	c.mu.Lock()
	if _, exists := c.listeners[listenerKey]; exists {
		c.mu.Unlock()
		return fmt.Errorf("config listener already exists for [dataId: %s, group: %s]", dataId, group)
	}

	ctx, cancel := context.WithCancel(c.ctx)
	listener := &ConfigListener{
		DataId:    dataId,
		Group:     group,
		Namespace: c.options.Namespace,
		Callback:  callback,
		Cancel:    cancel,
	}
	c.listeners[listenerKey] = listener
	c.mu.Unlock()

	err := c.configCli.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, content string) {
			select {
			case <-ctx.Done():
				return
			default:
				if callback != nil {
					callback(namespace, group, dataId, content)
				}
			}
		},
	})

	if err != nil {
		c.mu.Lock()
		delete(c.listeners, listenerKey)
		c.mu.Unlock()
		return fmt.Errorf("failed to listen config [dataId: %s, group: %s]: %w", dataId, group, err)
	}

	return nil
}

// CancelListenConfig 取消监听配置变更
func (c *ConfigClient) CancelListenConfig(dataId, group string) error {
	if group == "" {
		group = c.options.Group
	}

	listenerKey := fmt.Sprintf("%s@%s", dataId, group)

	c.mu.Lock()
	listener, exists := c.listeners[listenerKey]
	if !exists {
		c.mu.Unlock()
		return fmt.Errorf("config listener not found for [dataId: %s, group: %s]", dataId, group)
	}
	delete(c.listeners, listenerKey)
	c.mu.Unlock()

	if listener.Cancel != nil {
		listener.Cancel()
	}

	err := c.configCli.CancelListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		return fmt.Errorf("failed to cancel listen config [dataId: %s, group: %s]: %w", dataId, group, err)
	}

	return nil
}

// SearchConfig 搜索配置
func (c *ConfigClient) SearchConfig(dataId, group string, pageNo, pageSize int) (*ConfigSearchResult, error) {
	if group == "" {
		group = c.options.Group
	}

	result, err := c.configCli.SearchConfig(vo.SearchConfigParam{
		DataId:   dataId,
		Group:    group,
		PageNo:   pageNo,
		PageSize: pageSize,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search config: %w", err)
	}

	items := make([]ConfigItem, 0, len(result.PageItems))
	for _, item := range result.PageItems {
		items = append(items, ConfigItem{
			DataId:  item.DataId,
			Group:   item.Group,
			Content: item.Content,
			Tenant:  item.Tenant,
		})
	}

	return &ConfigSearchResult{
		TotalCount:     int64(result.TotalCount),
		PageNumber:     int32(result.PageNumber),
		PagesAvailable: int32(result.PagesAvailable),
		PageItems:      items,
	}, nil
}

// GetConfigAndListen 获取配置并监听变更
func (c *ConfigClient) GetConfigAndListen(dataId, group string, callback ConfigChangeCallback) (string, error) {
	// 先获取配置
	content, err := c.GetConfig(dataId, group)
	if err != nil {
		return "", err
	}

	// 再监听变更
	err = c.ListenConfig(dataId, group, callback)
	if err != nil {
		return "", err
	}

	return content, nil
}

// Close 关闭配置中心客户端
func (c *ConfigClient) Close() error {
	c.cancel()

	// 取消所有监听
	c.mu.Lock()
	listeners := make([]*ConfigListener, 0, len(c.listeners))
	for _, listener := range c.listeners {
		listeners = append(listeners, listener)
	}
	c.mu.Unlock()

	for _, listener := range listeners {
		_ = c.CancelListenConfig(listener.DataId, listener.Group)
	}

	return nil
}

// ConfigItem 配置项
type ConfigItem struct {
	DataId  string
	Group   string
	Content string
	Tenant  string
}

// ConfigSearchResult 配置搜索结果
type ConfigSearchResult struct {
	TotalCount     int64
	PageNumber     int32
	PagesAvailable int32
	PageItems      []ConfigItem
}

// ConfigClientOption 配置客户端选项函数
type ConfigClientOption func(*ConfigClientOptions)

// ConfigClientOptions 配置客户端选项
type ConfigClientOptions struct {
	DataId    string
	Group     string
	Namespace string
}

// WithConfigDataId 设置配置ID
func WithConfigDataId(dataId string) ConfigClientOption {
	return func(o *ConfigClientOptions) {
		o.DataId = dataId
	}
}

// WithConfigGroup 设置配置分组
func WithConfigGroup(group string) ConfigClientOption {
	return func(o *ConfigClientOptions) {
		o.Group = group
	}
}

// WithConfigNamespace 设置配置命名空间
func WithConfigNamespace(namespace string) ConfigClientOption {
	return func(o *ConfigClientOptions) {
		o.Namespace = namespace
	}
}
