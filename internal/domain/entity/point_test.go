package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPoint(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		pointName string
		pointType PointType
		want      *Point
	}{
		{
			name:      "创建遥信点",
			code:      "YX001",
			pointName: "开关状态",
			pointType: PointTypeYaoXin,
			want: &Point{
				Code:   "YX001",
				Name:   "开关状态",
				Type:   PointTypeYaoXin,
				Status: 1,
			},
		},
		{
			name:      "创建遥测点",
			code:      "YC001",
			pointName: "电压",
			pointType: PointTypeYaoCe,
			want: &Point{
				Code:   "YC001",
				Name:   "电压",
				Type:   PointTypeYaoCe,
				Status: 1,
			},
		},
		{
			name:      "创建遥控点",
			code:      "YK001",
			pointName: "开关控制",
			pointType: PointTypeYaoKong,
			want: &Point{
				Code:   "YK001",
				Name:   "开关控制",
				Type:   PointTypeYaoKong,
				Status: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPoint(tt.code, tt.pointName, tt.pointType)
			assert.Equal(t, tt.want.Code, got.Code)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestPoint_SetRange(t *testing.T) {
	point := NewPoint("YC001", "电压", PointTypeYaoCe)

	point.SetRange(0, 500)

	assert.Equal(t, 0.0, point.MinValue)
	assert.Equal(t, 500.0, point.MaxValue)
}

func TestPoint_SetAlarmThreshold(t *testing.T) {
	point := NewPoint("YC001", "电压", PointTypeYaoCe)

	point.SetAlarmThreshold(450, 10)

	assert.True(t, point.IsAlarm)
	assert.Equal(t, 450.0, point.AlarmHigh)
	assert.Equal(t, 10.0, point.AlarmLow)
}

func TestPoint_IsInRange(t *testing.T) {
	tests := []struct {
		name     string
		min      float64
		max      float64
		value    float64
		expected bool
	}{
		{
			name:     "值在范围内",
			min:      0,
			max:      100,
			value:    50,
			expected: true,
		},
		{
			name:     "值等于最小值",
			min:      0,
			max:      100,
			value:    0,
			expected: true,
		},
		{
			name:     "值等于最大值",
			min:      0,
			max:      100,
			value:    100,
			expected: true,
		},
		{
			name:     "值小于最小值",
			min:      0,
			max:      100,
			value:    -1,
			expected: false,
		},
		{
			name:     "值大于最大值",
			min:      0,
			max:      100,
			value:    101,
			expected: false,
		},
		{
			name:     "未设置范围",
			min:      0,
			max:      0,
			value:    1000,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := NewPoint("YC001", "电压", PointTypeYaoCe)
			if tt.min != 0 || tt.max != 0 {
				point.SetRange(tt.min, tt.max)
			}
			assert.Equal(t, tt.expected, point.IsInRange(tt.value))
		})
	}
}

func TestPoint_CheckAlarm(t *testing.T) {
	tests := []struct {
		name           string
		alarmHigh      float64
		alarmLow       float64
		value          float64
		expectedAlarm  bool
		expectedType   string
	}{
		{
			name:          "值正常",
			alarmHigh:     100,
			alarmLow:      10,
			value:         50,
			expectedAlarm: false,
			expectedType:  "",
		},
		{
			name:          "值超过高限",
			alarmHigh:     100,
			alarmLow:      10,
			value:         150,
			expectedAlarm: true,
			expectedType:  "high",
		},
		{
			name:          "值低于低限",
			alarmHigh:     100,
			alarmLow:      10,
			value:         5,
			expectedAlarm: true,
			expectedType:  "low",
		},
		{
			name:          "值等于高限",
			alarmHigh:     100,
			alarmLow:      10,
			value:         100,
			expectedAlarm: false,
			expectedType:  "",
		},
		{
			name:          "值等于低限",
			alarmHigh:     100,
			alarmLow:      10,
			value:         10,
			expectedAlarm: false,
			expectedType:  "",
		},
		{
			name:          "未设置告警阈值",
			alarmHigh:     0,
			alarmLow:      0,
			value:         1000,
			expectedAlarm: false,
			expectedType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := NewPoint("YC001", "电压", PointTypeYaoCe)
			if tt.alarmHigh > 0 || tt.alarmLow > 0 {
				point.SetAlarmThreshold(tt.alarmHigh, tt.alarmLow)
			}
			alarm, alarmType := point.CheckAlarm(tt.value)
			assert.Equal(t, tt.expectedAlarm, alarm)
			assert.Equal(t, tt.expectedType, alarmType)
		})
	}
}

func TestPoint_TableName(t *testing.T) {
	point := Point{}
	assert.Equal(t, "points", point.TableName())
}

func TestPointType_Constants(t *testing.T) {
	assert.Equal(t, PointType("yaoxin"), PointTypeYaoXin)
	assert.Equal(t, PointType("yaoc"), PointTypeYaoCe)
	assert.Equal(t, PointType("yaokong"), PointTypeYaoKong)
	assert.Equal(t, PointType("setpoint"), PointTypeSetPoint)
	assert.Equal(t, PointType("diandu"), PointTypeDianDu)
}
