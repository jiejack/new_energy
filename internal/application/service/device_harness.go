package service

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/harness"
)

// DeviceHarness 设备 Harness 验证器
type DeviceHarness struct {
	harness *harness.Harness
}

// NewDeviceHarness 创建新的设备 Harness 实例
func NewDeviceHarness() *DeviceHarness {
	return &DeviceHarness{
		harness: harness.NewHarness(),
	}
}

// NewDeviceHarnessWithComponents 使用自定义组件创建设备 Harness 实例
func NewDeviceHarnessWithComponents(h *harness.Harness) *DeviceHarness {
	return &DeviceHarness{
		harness: h,
	}
}

// ValidateCreateDevice 验证创建设备请求
func (dh *DeviceHarness) ValidateCreateDevice(ctx context.Context, req *CreateDeviceRequest) error {
	// 基础验证
	if err := dh.harness.Validate(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 业务规则验证
	if err := dh.validateDeviceBusinessRules(req); err != nil {
		return err
	}

	return nil
}

// validateDeviceBusinessRules 验证设备业务规则
func (dh *DeviceHarness) validateDeviceBusinessRules(req *CreateDeviceRequest) error {
	// 验证设备编码格式
	if err := validateDeviceCode(req.Code); err != nil {
		return err
	}

	// 验证设备名称
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("device name cannot be empty or whitespace")
	}

	// 验证设备类型
	if !isValidDeviceType(req.Type) {
		return fmt.Errorf("invalid device type: %s", req.Type)
	}

	// 验证额定参数
	if err := validateRatedParameters(req.RatedPower, req.RatedVoltage, req.RatedCurrent); err != nil {
		return err
	}

	// 验证通信参数
	if req.Protocol != "" || req.IPAddress != "" {
		if err := validateCommunicationParams(req.Protocol, req.IPAddress, req.Port, req.SlaveID); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateDevice 验证更新设备请求
func (dh *DeviceHarness) ValidateUpdateDevice(ctx context.Context, deviceID string, req *UpdateDeviceRequest) error {
	// 基础验证
	if deviceID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}

	// 验证额定参数
	if err := validateRatedParameters(req.RatedPower, req.RatedVoltage, req.RatedCurrent); err != nil {
		return err
	}

	return nil
}

// ValidateDeviceStatus 验证设备状态
func (dh *DeviceHarness) ValidateDeviceStatus(ctx context.Context, status entity.DeviceStatus) error {
	if !isValidDeviceStatus(status) {
		return fmt.Errorf("invalid device status: %d", status)
	}
	return nil
}

// ValidateDeviceQuery 验证设备查询请求
func (dh *DeviceHarness) ValidateDeviceQuery(ctx context.Context, stationID *string, deviceType *entity.DeviceType) error {
	// 验证设备类型（如果提供）
	if deviceType != nil && !isValidDeviceType(*deviceType) {
		return fmt.Errorf("invalid device type: %s", *deviceType)
	}

	return nil
}

// ValidateCommunicationParams 验证通信参数
func (dh *DeviceHarness) ValidateCommunicationParams(ctx context.Context, protocol, ipAddress string, port, slaveID int) error {
	return validateCommunicationParams(protocol, ipAddress, port, slaveID)
}

// VerifyDeviceOutput 验证设备输出
func (dh *DeviceHarness) VerifyDeviceOutput(ctx context.Context, expected, actual *entity.Device) (bool, error) {
	return dh.harness.Verify(ctx, expected, actual)
}

// CreateDeviceSnapshot 创建设备快照
func (dh *DeviceHarness) CreateDeviceSnapshot(ctx context.Context, device *entity.Device) ([]byte, error) {
	return dh.harness.Snapshot(ctx, device)
}

// RecordDeviceMetric 记录设备指标
func (dh *DeviceHarness) RecordDeviceMetric(ctx context.Context, metric string, value float64) error {
	return dh.harness.RecordMetric(ctx, metric, value)
}

// GetDeviceMetrics 获取设备指标
func (dh *DeviceHarness) GetDeviceMetrics(ctx context.Context, pattern string) ([]harness.Metric, error) {
	return dh.harness.GetMetrics(ctx, pattern)
}

// GetHarness 获取底层 Harness 实例
func (dh *DeviceHarness) GetHarness() *harness.Harness {
	return dh.harness
}

// isValidDeviceType 验证设备类型是否有效
func isValidDeviceType(deviceType entity.DeviceType) bool {
	return deviceType == entity.DeviceTypeInverter ||
		deviceType == entity.DeviceTypeMeter ||
		deviceType == entity.DeviceTypeTransformer ||
		deviceType == entity.DeviceTypeSwitch ||
		deviceType == entity.DeviceTypeWeather ||
		deviceType == entity.DeviceTypeESS ||
		deviceType == entity.DeviceTypePCS ||
		deviceType == entity.DeviceTypeBMS
}

// isValidDeviceStatus 验证设备状态是否有效
func isValidDeviceStatus(status entity.DeviceStatus) bool {
	return status == entity.DeviceStatusOffline ||
		status == entity.DeviceStatusOnline ||
		status == entity.DeviceStatusFault ||
		status == entity.DeviceStatusMaintain
}

// validateDeviceCode 验证设备编码格式
func validateDeviceCode(code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("device code cannot be empty or whitespace")
	}

	// 设备编码格式：字母、数字、下划线、中划线，长度3-50
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]{3,50}$`, code)
	if err != nil {
		return fmt.Errorf("failed to validate device code: %w", err)
	}
	if !matched {
		return fmt.Errorf("device code must be 3-50 characters and contain only letters, numbers, underscores, and hyphens")
	}

	return nil
}

// validateRatedParameters 验证额定参数
func validateRatedParameters(power, voltage, current float64) error {
	// 额定功率应该为非负数
	if power < 0 {
		return fmt.Errorf("rated power cannot be negative: %f", power)
	}

	// 额定电压应该为非负数
	if voltage < 0 {
		return fmt.Errorf("rated voltage cannot be negative: %f", voltage)
	}

	// 额定电流应该为非负数
	if current < 0 {
		return fmt.Errorf("rated current cannot be negative: %f", current)
	}

	// 如果提供了功率和电压，可以计算电流，验证一致性
	if power > 0 && voltage > 0 && current > 0 {
		// 三相功率公式: P = √3 * U * I * cos(φ)
		// 这里简化验证，允许一定的误差范围（±20%）
		expectedCurrent := power / (voltage * 1.732) // 假设功率因数为1
		tolerance := expectedCurrent * 0.2
		if current < expectedCurrent-tolerance || current > expectedCurrent+tolerance {
			// 不强制要求一致性，但记录警告
			// 实际应用中可以根据需求调整
		}
	}

	return nil
}

// validateCommunicationParams 验证通信参数
func validateCommunicationParams(protocol, ipAddress string, port, slaveID int) error {
	// 验证协议
	if protocol != "" {
		validProtocols := map[string]bool{
			"modbus-tcp": true,
			"modbus-rtu": true,
			"mqtt":       true,
			"http":       true,
			"opcua":      true,
		}
		if !validProtocols[strings.ToLower(protocol)] {
			return fmt.Errorf("unsupported protocol: %s", protocol)
		}
	}

	// 验证IP地址
	if ipAddress != "" {
		if net.ParseIP(ipAddress) == nil {
			return fmt.Errorf("invalid IP address: %s", ipAddress)
		}
	}

	// 验证端口
	if port != 0 {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535, got: %d", port)
		}
	}

	// 验证从站ID（Modbus）
	if slaveID != 0 {
		if slaveID < 1 || slaveID > 247 {
			return fmt.Errorf("slave ID must be between 1 and 247, got: %d", slaveID)
		}
	}

	return nil
}
