package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewDeviceHarness(t *testing.T) {
	harness := NewDeviceHarness()
	assert.NotNil(t, harness)
	assert.NotNil(t, harness.harness)
}

func TestDeviceHarness_ValidateCreateDevice(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	t.Run("有效的创建设备请求", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:          "INV-001",
			Name:          "1号逆变器",
			Type:          entity.DeviceTypeInverter,
			StationID:     "station001",
			Manufacturer:  "华为",
			Model:         "SUN2000-100KTL",
			RatedPower:    100.0,
			RatedVoltage:  400.0,
			RatedCurrent:  144.3,
			Protocol:      "modbus-tcp",
			IPAddress:     "192.168.1.100",
			Port:          502,
			SlaveID:       1,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("无效的设备编码-为空", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "",
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device code cannot be empty")
	})

	t.Run("无效的设备编码-格式错误", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "测试设备编码!", // 包含中文和特殊字符
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device code must be 3-50 characters")
	})

	t.Run("无效的设备编码-长度不足", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "AB", // 长度不足
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device code must be 3-50 characters")
	})

	t.Run("无效的设备名称-为空", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device name cannot be empty")
	})

	t.Run("无效的设备类型", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "测试设备",
			Type:      entity.DeviceType("invalid_type"),
			StationID: "station001",
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid device type")
	})

	t.Run("负数的额定功率", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:        "INV-001",
			Name:        "测试设备",
			Type:        entity.DeviceTypeInverter,
			StationID:   "station001",
			RatedPower:  -100.0,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated power cannot be negative")
	})

	t.Run("负数的额定电压", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:          "INV-001",
			Name:          "测试设备",
			Type:          entity.DeviceTypeInverter,
			StationID:     "station001",
			RatedVoltage:  -400.0,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated voltage cannot be negative")
	})

	t.Run("负数的额定电流", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:          "INV-001",
			Name:          "测试设备",
			Type:          entity.DeviceTypeInverter,
			StationID:     "station001",
			RatedCurrent:  -100.0,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated current cannot be negative")
	})

	t.Run("无效的IP地址", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
			Protocol:  "modbus-tcp",
			IPAddress: "invalid-ip",
			Port:      502,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP address")
	})

	t.Run("无效的端口号-超出范围", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
			Protocol:  "modbus-tcp",
			IPAddress: "192.168.1.100",
			Port:      70000, // 超出范围
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port must be between 1 and 65535")
	})

	t.Run("无效的端口号-为零", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
			Protocol:  "modbus-tcp",
			IPAddress: "192.168.1.100",
			Port:      0, // 零值
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.NoError(t, err) // 零值表示未设置，应该通过
	})

	t.Run("无效的从站ID", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
			Protocol:  "modbus-tcp",
			IPAddress: "192.168.1.100",
			Port:      502,
			SlaveID:   300, // 超出范围
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "slave ID must be between 1 and 247")
	})

	t.Run("不支持的协议", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:      "INV-001",
			Name:      "测试设备",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
			Protocol:  "unsupported-protocol",
			IPAddress: "192.168.1.100",
			Port:      502,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol")
	})

	t.Run("电表设备", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:         "METER-001",
			Name:         "1号电表",
			Type:         entity.DeviceTypeMeter,
			StationID:    "station001",
			Manufacturer: "安科瑞",
			Model:        "ACR220EL",
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("储能设备", func(t *testing.T) {
		req := &CreateDeviceRequest{
			Code:         "ESS-001",
			Name:         "1号储能系统",
			Type:         entity.DeviceTypeESS,
			StationID:    "station001",
			Manufacturer: "比亚迪",
			RatedPower:   500.0,
		}

		err := harness.ValidateCreateDevice(ctx, req)
		assert.NoError(t, err)
	})
}

func TestDeviceHarness_ValidateUpdateDevice(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	t.Run("有效的更新请求", func(t *testing.T) {
		req := &UpdateDeviceRequest{
			Name:          "更新后的设备名称",
			Manufacturer:  "新制造商",
			Model:         "新型号",
			RatedPower:    150.0,
			RatedVoltage:  400.0,
			RatedCurrent:  216.5,
		}

		err := harness.ValidateUpdateDevice(ctx, "device001", req)
		assert.NoError(t, err)
	})

	t.Run("设备ID为空", func(t *testing.T) {
		req := &UpdateDeviceRequest{
			Name: "更新后的设备名称",
		}

		err := harness.ValidateUpdateDevice(ctx, "", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device ID cannot be empty")
	})

	t.Run("负数的额定功率", func(t *testing.T) {
		req := &UpdateDeviceRequest{
			RatedPower: -100.0,
		}

		err := harness.ValidateUpdateDevice(ctx, "device001", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated power cannot be negative")
	})
}

func TestDeviceHarness_ValidateDeviceStatus(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	tests := []struct {
		name     string
		status   entity.DeviceStatus
		hasError bool
	}{
		{"离线状态", entity.DeviceStatusOffline, false},
		{"在线状态", entity.DeviceStatusOnline, false},
		{"故障状态", entity.DeviceStatusFault, false},
		{"维护状态", entity.DeviceStatusMaintain, false},
		{"无效状态", entity.DeviceStatus(99), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := harness.ValidateDeviceStatus(ctx, tt.status)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid device status")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceHarness_ValidateDeviceQuery(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	t.Run("有效的查询请求", func(t *testing.T) {
		stationID := "station001"
		deviceType := entity.DeviceTypeInverter
		err := harness.ValidateDeviceQuery(ctx, &stationID, &deviceType)
		assert.NoError(t, err)
	})

	t.Run("无过滤条件的查询", func(t *testing.T) {
		err := harness.ValidateDeviceQuery(ctx, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("无效的设备类型", func(t *testing.T) {
		deviceType := entity.DeviceType("invalid_type")
		err := harness.ValidateDeviceQuery(ctx, nil, &deviceType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid device type")
	})
}

func TestDeviceHarness_ValidateCommunicationParams(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	t.Run("有效的Modbus-TCP参数", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "modbus-tcp", "192.168.1.100", 502, 1)
		assert.NoError(t, err)
	})

	t.Run("有效的MQTT参数", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "mqtt", "192.168.1.200", 1883, 0)
		assert.NoError(t, err)
	})

	t.Run("有效的HTTP参数", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "http", "192.168.1.50", 8080, 0)
		assert.NoError(t, err)
	})

	t.Run("有效的OPCUA参数", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "opcua", "192.168.1.150", 4840, 0)
		assert.NoError(t, err)
	})

	t.Run("无效的协议", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "invalid", "192.168.1.100", 502, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol")
	})

	t.Run("无效的IP地址", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "modbus-tcp", "999.999.999.999", 502, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP address")
	})

	t.Run("无效的端口", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "modbus-tcp", "192.168.1.100", 0, 1)
		assert.NoError(t, err) // 零值表示未设置
	})

	t.Run("无效的从站ID", func(t *testing.T) {
		err := harness.ValidateCommunicationParams(ctx, "modbus-rtu", "192.168.1.100", 502, 0)
		assert.NoError(t, err) // 零值表示未设置
	})
}

func TestDeviceHarness_VerifyDeviceOutput(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	t.Run("输出验证成功", func(t *testing.T) {
		expected := entity.NewDevice("INV-001", "逆变器", entity.DeviceTypeInverter, "station001")
		expected.Manufacturer = "华为"
		expected.Model = "SUN2000-100KTL"

		actual := entity.NewDevice("INV-001", "逆变器", entity.DeviceTypeInverter, "station001")
		actual.Manufacturer = "华为"
		actual.Model = "SUN2000-100KTL"

		match, err := harness.VerifyDeviceOutput(ctx, expected, actual)
		assert.NoError(t, err)
		assert.True(t, match)
	})

	t.Run("输出验证失败", func(t *testing.T) {
		expected := entity.NewDevice("INV-001", "逆变器1", entity.DeviceTypeInverter, "station001")
		actual := entity.NewDevice("INV-002", "逆变器2", entity.DeviceTypeMeter, "station002")

		match, err := harness.VerifyDeviceOutput(ctx, expected, actual)
		assert.NoError(t, err)
		assert.False(t, match)
	})
}

func TestDeviceHarness_CreateDeviceSnapshot(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	device := entity.NewDevice("INV-001", "逆变器", entity.DeviceTypeInverter, "station001")
	device.Manufacturer = "华为"
	device.Model = "SUN2000-100KTL"

	snapshot, err := harness.CreateDeviceSnapshot(ctx, device)
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.True(t, len(snapshot) > 0)
}

func TestDeviceHarness_RecordDeviceMetric(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	err := harness.RecordDeviceMetric(ctx, "device.power", 100.5)
	assert.NoError(t, err)

	err = harness.RecordDeviceMetric(ctx, "device.voltage", 400.2)
	assert.NoError(t, err)
}

func TestDeviceHarness_GetDeviceMetrics(t *testing.T) {
	ctx := context.Background()
	harness := NewDeviceHarness()

	// 记录一些指标
	_ = harness.RecordDeviceMetric(ctx, "device.power", 100.5)
	_ = harness.RecordDeviceMetric(ctx, "device.voltage", 400.2)
	_ = harness.RecordDeviceMetric(ctx, "device.current", 144.3)

	// 获取所有指标
	metrics, err := harness.GetDeviceMetrics(ctx, "")
	assert.NoError(t, err)
	assert.True(t, len(metrics) >= 3)

	// 获取特定模式的指标
	powerMetrics, err := harness.GetDeviceMetrics(ctx, "power")
	assert.NoError(t, err)
	assert.True(t, len(powerMetrics) >= 1)
}

func TestDeviceHarness_GetHarness(t *testing.T) {
	harness := NewDeviceHarness()

	h := harness.GetHarness()
	assert.NotNil(t, h)
}

func TestIsValidDeviceType(t *testing.T) {
	tests := []struct {
		name       string
		deviceType entity.DeviceType
		expected   bool
	}{
		{"逆变器", entity.DeviceTypeInverter, true},
		{"电表", entity.DeviceTypeMeter, true},
		{"变压器", entity.DeviceTypeTransformer, true},
		{"开关", entity.DeviceTypeSwitch, true},
		{"气象站", entity.DeviceTypeWeather, true},
		{"储能系统", entity.DeviceTypeESS, true},
		{"PCS", entity.DeviceTypePCS, true},
		{"BMS", entity.DeviceTypeBMS, true},
		{"无效类型", entity.DeviceType("invalid"), false},
		{"空类型", entity.DeviceType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDeviceType(tt.deviceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidDeviceStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   entity.DeviceStatus
		expected bool
	}{
		{"离线", entity.DeviceStatusOffline, true},
		{"在线", entity.DeviceStatusOnline, true},
		{"故障", entity.DeviceStatusFault, true},
		{"维护", entity.DeviceStatusMaintain, true},
		{"无效状态", entity.DeviceStatus(99), false},
		{"无效状态负数", entity.DeviceStatus(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDeviceStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateDeviceCode(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		hasError  bool
		errorMsg  string
	}{
		{"有效编码-字母数字", "INV001", false, ""},
		{"有效编码-带中划线", "INV-001", false, ""},
		{"有效编码-带下划线", "INV_001", false, ""},
		{"有效编码-混合", "INV-001_A", false, ""},
		{"空编码", "", true, "device code cannot be empty"},
		{"纯空格", "   ", true, "device code cannot be empty"},
		{"包含中文", "设备001", true, "device code must be 3-50 characters"},
		{"包含特殊字符", "INV@001", true, "device code must be 3-50 characters"},
		{"长度不足", "AB", true, "device code must be 3-50 characters"},
		{"长度刚好", "ABC", false, ""},
		{"长度超长", "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQRSTUV", true, "device code must be 3-50 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeviceCode(tt.code)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRatedParameters(t *testing.T) {
	t.Run("所有参数为零", func(t *testing.T) {
		err := validateRatedParameters(0, 0, 0)
		assert.NoError(t, err)
	})

	t.Run("所有参数为正数", func(t *testing.T) {
		err := validateRatedParameters(100.0, 400.0, 144.3)
		assert.NoError(t, err)
	})

	t.Run("功率为负数", func(t *testing.T) {
		err := validateRatedParameters(-100.0, 400.0, 144.3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated power cannot be negative")
	})

	t.Run("电压为负数", func(t *testing.T) {
		err := validateRatedParameters(100.0, -400.0, 144.3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated voltage cannot be negative")
	})

	t.Run("电流为负数", func(t *testing.T) {
		err := validateRatedParameters(100.0, 400.0, -144.3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rated current cannot be negative")
	})

	t.Run("部分参数为零", func(t *testing.T) {
		err := validateRatedParameters(100.0, 0, 0)
		assert.NoError(t, err)
	})
}

func TestValidateCommunicationParams(t *testing.T) {
	t.Run("空参数", func(t *testing.T) {
		err := validateCommunicationParams("", "", 0, 0)
		assert.NoError(t, err)
	})

	t.Run("有效的IPv4地址", func(t *testing.T) {
		err := validateCommunicationParams("modbus-tcp", "192.168.1.100", 502, 1)
		assert.NoError(t, err)
	})

	t.Run("有效的IPv6地址", func(t *testing.T) {
		err := validateCommunicationParams("modbus-tcp", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", 502, 1)
		assert.NoError(t, err)
	})

	t.Run("协议大小写不敏感", func(t *testing.T) {
		err := validateCommunicationParams("MODBUS-TCP", "192.168.1.100", 502, 1)
		assert.NoError(t, err)
	})

	t.Run("边界端口号", func(t *testing.T) {
		err := validateCommunicationParams("modbus-tcp", "192.168.1.100", 1, 1)
		assert.NoError(t, err)

		err = validateCommunicationParams("modbus-tcp", "192.168.1.100", 65535, 1)
		assert.NoError(t, err)
	})

	t.Run("边界从站ID", func(t *testing.T) {
		err := validateCommunicationParams("modbus-tcp", "192.168.1.100", 502, 1)
		assert.NoError(t, err)

		err = validateCommunicationParams("modbus-tcp", "192.168.1.100", 502, 247)
		assert.NoError(t, err)
	})
}
