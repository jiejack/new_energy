package iec104

import (
	"math"
	"testing"
	"time"
)

// ==================== VSQ测试 ====================

func TestVSQEncodeDecode(t *testing.T) {
	tests := []struct {
		name      string
		vsq       VSQ
		expected  byte
	}{
		{"Single non-sequence", VSQ{Number: 1, IsSequence: false}, 0x01},
		{"Multiple non-sequence", VSQ{Number: 10, IsSequence: false}, 0x0A},
		{"Single sequence", VSQ{Number: 1, IsSequence: true}, 0x81},
		{"Multiple sequence", VSQ{Number: 10, IsSequence: true}, 0x8A},
		{"Max number", VSQ{Number: 127, IsSequence: false}, 0x7F},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeVSQ(tt.vsq)
			if encoded != tt.expected {
				t.Errorf("EncodeVSQ() = %02X, want %02X", encoded, tt.expected)
			}

			decoded := ParseVSQ(encoded)
			if decoded.Number != tt.vsq.Number || decoded.IsSequence != tt.vsq.IsSequence {
				t.Errorf("ParseVSQ() = %+v, want %+v", decoded, tt.vsq)
			}
		})
	}
}

// ==================== COT测试 ====================

func TestCOTEncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		cot      CauseOfTransmission
		expected byte
	}{
		{"Periodic", CauseOfTransmission{Cause: COT_PERIODIC_CYCLIC}, 0x01},
		{"Spontaneous", CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0x03},
		{"Activation", CauseOfTransmission{Cause: COT_ACTIVATION}, 0x06},
		{"Activation with test", CauseOfTransmission{Cause: COT_ACTIVATION, IsTest: true}, 0x86},
		{"Activation with PN", CauseOfTransmission{Cause: COT_ACTIVATION, IsPN: true}, 0x46},
		{"Interrogated by station", CauseOfTransmission{Cause: COT_INTERROGATED_BY_STATION}, 0x14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cot1, cot2 := EncodeCOT(tt.cot)
			decoded := ParseCOT(cot1, cot2)

			if decoded.Cause != tt.cot.Cause {
				t.Errorf("Cause mismatch: got %d, want %d", decoded.Cause, tt.cot.Cause)
			}
			if decoded.IsTest != tt.cot.IsTest {
				t.Errorf("IsTest mismatch: got %v, want %v", decoded.IsTest, tt.cot.IsTest)
			}
			if decoded.IsPN != tt.cot.IsPN {
				t.Errorf("IsPN mismatch: got %v, want %v", decoded.IsPN, tt.cot.IsPN)
			}
		})
	}
}

// ==================== 质量描述测试 ====================

func TestQualityEncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		quality  Quality
		expected byte
	}{
		{"Good", Quality{}, 0x00},
		{"Invalid", Quality{Invalid: true}, 0x80},
		{"Not current", Quality{NotCurrent: true}, 0x40},
		{"Substituted", Quality{Substituted: true}, 0x20},
		{"Blocked", Quality{Blocked: true}, 0x10},
		{"Invalid+Blocked", Quality{Invalid: true, Blocked: true}, 0x90},
		{"All flags", Quality{Invalid: true, NotCurrent: true, Substituted: true, Blocked: true}, 0xF0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeQuality(tt.quality)
			if encoded != tt.expected {
				t.Errorf("EncodeQuality() = %02X, want %02X", encoded, tt.expected)
			}

			decoded := ParseQuality(encoded)
			if decoded != tt.quality {
				t.Errorf("ParseQuality() = %+v, want %+v", decoded, tt.quality)
			}
		})
	}
}

func TestMeasureQualityEncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		quality  Quality
		expected byte
	}{
		{"Good", Quality{}, 0x00},
		{"Overflow", Quality{Overflow: true}, 0x01},
		{"Invalid+Overflow", Quality{Invalid: true, Overflow: true}, 0x81},
		{"All flags", Quality{Invalid: true, NotCurrent: true, Substituted: true, Blocked: true, Overflow: true}, 0xF1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeMeasureQuality(tt.quality)
			if encoded != tt.expected {
				t.Errorf("EncodeMeasureQuality() = %02X, want %02X", encoded, tt.expected)
			}

			decoded := ParseMeasureQuality(encoded)
			if decoded != tt.quality {
				t.Errorf("ParseMeasureQuality() = %+v, want %+v", decoded, tt.quality)
			}
		})
	}
}

// ==================== 时标测试 ====================

func TestCP24Time2aEncodeDecode(t *testing.T) {
	now := time.Now()

	encoded := EncodeCP24Time2a(now)
	decoded := DecodeCP24Time2a(encoded)

	// 验证毫秒和分钟
	// 安全转换：确保值在uint16范围内
	expectedMs := uint16(min(now.Nanosecond()/1000000+int(now.Second())*1000, math.MaxUint16))
	if decoded.Milliseconds != expectedMs {
		t.Errorf("Milliseconds mismatch: got %d, want %d", decoded.Milliseconds, expectedMs)
	}

	if decoded.Minutes != uint8(now.Minute()) {
		t.Errorf("Minutes mismatch: got %d, want %d", decoded.Minutes, uint8(now.Minute()))
	}
}

func TestCP56Time2aEncodeDecode(t *testing.T) {
	now := time.Now()

	encoded := EncodeCP56Time2a(now)
	decoded, err := DecodeCP56Time2a(encoded)
	if err != nil {
		t.Fatalf("DecodeCP56Time2a() error: %v", err)
	}

	// 验证各字段
	expectedMs := uint16(now.Nanosecond()/1000000) + uint16(now.Second())*1000
	if decoded.Milliseconds != expectedMs {
		t.Errorf("Milliseconds mismatch: got %d, want %d", decoded.Milliseconds, expectedMs)
	}

	if decoded.Minutes != uint8(now.Minute()) {
		t.Errorf("Minutes mismatch: got %d, want %d", decoded.Minutes, uint8(now.Minute()))
	}

	if decoded.Hours != uint8(now.Hour()) {
		t.Errorf("Hours mismatch: got %d, want %d", decoded.Hours, uint8(now.Hour()))
	}

	if decoded.DayOfMonth != uint8(now.Day()) {
		t.Errorf("DayOfMonth mismatch: got %d, want %d", decoded.DayOfMonth, uint8(now.Day()))
	}

	if decoded.Month != uint8(now.Month()) {
		t.Errorf("Month mismatch: got %d, want %d", decoded.Month, uint8(now.Month()))
	}
}

func TestCP56Time2aToTime(t *testing.T) {
	cp56 := CP56Time2a{
		Milliseconds: 12345,
		Minutes:      30,
		Hours:        14,
		DayOfWeek:    3,
		DayOfMonth:   15,
		Month:        6,
		Year:         25,
	}

	tm := cp56.ToTime()

	if tm.Hour() != 14 {
		t.Errorf("Hour mismatch: got %d, want 14", tm.Hour())
	}
	if tm.Minute() != 30 {
		t.Errorf("Minute mismatch: got %d, want 30", tm.Minute())
	}
	if tm.Second() != 12 {
		t.Errorf("Second mismatch: got %d, want 12", tm.Second())
	}
	if tm.Day() != 15 {
		t.Errorf("Day mismatch: got %d, want 15", tm.Day())
	}
	if tm.Month() != 6 {
		t.Errorf("Month mismatch: got %d, want 6", tm.Month())
	}
}

// ==================== ASDU编解码测试 ====================

func TestASDUEncodeDecodeHeader(t *testing.T) {
	coder := NewASDUCoder()

	typeID := uint8(TYPE_ID_SINGLE_POINT_INFO)
	vsq := VSQ{Number: 5, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	commonAddr := uint16(1)

	data := coder.EncodeASDUHeader(typeID, vsq, cot, commonAddr)

	if len(data) != 6 {
		t.Errorf("Header length mismatch: got %d, want 6", len(data))
	}

	if data[0] != typeID {
		t.Errorf("TypeID mismatch: got %d, want %d", data[0], typeID)
	}

	decodedVSQ := ParseVSQ(data[1])
	if decodedVSQ.Number != vsq.Number || decodedVSQ.IsSequence != vsq.IsSequence {
		t.Errorf("VSQ mismatch: got %+v, want %+v", decodedVSQ, vsq)
	}

	decodedCOT := ParseCOT(data[2], data[3])
	if decodedCOT.Cause != cot.Cause {
		t.Errorf("COT mismatch: got %d, want %d", decodedCOT.Cause, cot.Cause)
	}

	decodedAddr := uint16(data[4]) | uint16(data[5])<<8
	if decodedAddr != commonAddr {
		t.Errorf("CommonAddr mismatch: got %d, want %d", decodedAddr, commonAddr)
	}
}

func TestInfoAddressEncode(t *testing.T) {
	coder := NewASDUCoder()

	tests := []struct {
		name    string
		address uint32
	}{
		{"Zero", 0},
		{"Small", 100},
		{"Medium", 10000},
		{"Large", 0xFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := coder.EncodeInfoAddress(tt.address)
			if len(encoded) != 3 {
				t.Errorf("InfoAddress length mismatch: got %d, want 3", len(encoded))
			}

			decoded := uint32(encoded[0]) | uint32(encoded[1])<<8 | uint32(encoded[2])<<16
			if decoded != tt.address {
				t.Errorf("InfoAddress mismatch: got %d, want %d", decoded, tt.address)
			}
		})
	}
}

// ==================== 单点遥信测试 ====================

func TestSinglePointInfoEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{
			Address: 1001,
			Value: &SinglePointInfo{
				Value:   true,
				Quality: Quality{Invalid: false},
			},
		},
		{
			Address: 1002,
			Value: &SinglePointInfo{
				Value:   false,
				Quality: Quality{Invalid: true},
			},
		},
	}

	data, err := coder.EncodeSinglePointInfo(objects, false, 0)
	if err != nil {
		t.Fatalf("EncodeSinglePointInfo() error: %v", err)
	}

	// 验证基本结构
	if len(data) < 6 {
		t.Errorf("Data too short: %d", len(data))
	}

	// 解码验证
	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_SINGLE_POINT_INFO {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_SINGLE_POINT_INFO)
	}

	if asdu.VSQ.Number != 2 {
		t.Errorf("VSQ.Number mismatch: got %d, want 2", asdu.VSQ.Number)
	}

	if len(asdu.Information) != 2 {
		t.Errorf("Information count mismatch: got %d, want 2", len(asdu.Information))
	}

	// 验证第一个对象
	if asdu.Information[0].Address != 1001 {
		t.Errorf("Address[0] mismatch: got %d, want 1001", asdu.Information[0].Address)
	}

	spi0, ok := asdu.Information[0].Value.(*SinglePointInfo)
	if !ok {
		t.Fatalf("Value[0] type mismatch")
	}
	if !spi0.Value {
		t.Errorf("Value[0] mismatch: got %v, want true", spi0.Value)
	}

	// 验证第二个对象
	spi1, ok := asdu.Information[1].Value.(*SinglePointInfo)
	if !ok {
		t.Fatalf("Value[1] type mismatch")
	}
	if spi1.Value {
		t.Errorf("Value[1] mismatch: got %v, want false", spi1.Value)
	}
	if !spi1.Quality.Invalid {
		t.Errorf("Quality[1].Invalid mismatch: got %v, want true", spi1.Quality.Invalid)
	}
}

// ==================== 双点遥信测试 ====================

func TestDoublePointInfoEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{
			Address: 2001,
			Value: &DoublePointInfo{
				Value:   DP_ON,
				Quality: Quality{},
			},
		},
		{
			Address: 2002,
			Value: &DoublePointInfo{
				Value:   DP_OFF,
				Quality: Quality{Blocked: true},
			},
		},
	}

	data, err := coder.EncodeDoublePointInfo(objects, false, 0)
	if err != nil {
		t.Fatalf("EncodeDoublePointInfo() error: %v", err)
	}

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_DOUBLE_POINT_INFO {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_DOUBLE_POINT_INFO)
	}

	if len(asdu.Information) != 2 {
		t.Errorf("Information count mismatch: got %d, want 2", len(asdu.Information))
	}

	dpi0, ok := asdu.Information[0].Value.(*DoublePointInfo)
	if !ok {
		t.Fatalf("Value[0] type mismatch")
	}
	if dpi0.Value != DP_ON {
		t.Errorf("Value[0] mismatch: got %d, want %d", dpi0.Value, DP_ON)
	}

	dpi1, ok := asdu.Information[1].Value.(*DoublePointInfo)
	if !ok {
		t.Fatalf("Value[1] type mismatch")
	}
	if dpi1.Value != DP_OFF {
		t.Errorf("Value[1] mismatch: got %d, want %d", dpi1.Value, DP_OFF)
	}
}

// ==================== 遥测值测试 ====================

func TestNormalizedValueEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{
			Address: 3001,
			Value: &NormalizedValue{
				Value:   0.5,
				Quality: Quality{},
			},
		},
		{
			Address: 3002,
			Value: &NormalizedValue{
				Value:   -0.5,
				Quality: Quality{Overflow: true},
			},
		},
	}

	data, err := coder.EncodeNormalizedValue(objects, false, 0)
	if err != nil {
		t.Fatalf("EncodeNormalizedValue() error: %v", err)
	}

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_MEASURE_VALUE_NORMAL {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_MEASURE_VALUE_NORMAL)
	}

	nv0, ok := asdu.Information[0].Value.(*NormalizedValue)
	if !ok {
		t.Fatalf("Value[0] type mismatch")
	}
	// 允许一定的精度误差
	if nv0.Value < 0.49 || nv0.Value > 0.51 {
		t.Errorf("Value[0] mismatch: got %f, want ~0.5", nv0.Value)
	}

	nv1, ok := asdu.Information[1].Value.(*NormalizedValue)
	if !ok {
		t.Fatalf("Value[1] type mismatch")
	}
	if nv1.Value > -0.49 || nv1.Value < -0.51 {
		t.Errorf("Value[1] mismatch: got %f, want ~-0.5", nv1.Value)
	}
}

func TestScaledValueEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{
			Address: 4001,
			Value: &ScaledValue{
				Value:   1000,
				Quality: Quality{},
			},
		},
		{
			Address: 4002,
			Value: &ScaledValue{
				Value:   -500,
				Quality: Quality{Invalid: true},
			},
		},
	}

	data, err := coder.EncodeScaledValue(objects, false, 0)
	if err != nil {
		t.Fatalf("EncodeScaledValue() error: %v", err)
	}

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_MEASURE_VALUE_SCALED {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_MEASURE_VALUE_SCALED)
	}

	sv0, ok := asdu.Information[0].Value.(*ScaledValue)
	if !ok {
		t.Fatalf("Value[0] type mismatch")
	}
	if sv0.Value != 1000 {
		t.Errorf("Value[0] mismatch: got %d, want 1000", sv0.Value)
	}

	sv1, ok := asdu.Information[1].Value.(*ScaledValue)
	if !ok {
		t.Fatalf("Value[1] type mismatch")
	}
	if sv1.Value != -500 {
		t.Errorf("Value[1] mismatch: got %d, want -500", sv1.Value)
	}
}

func TestFloatValueEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{
			Address: 5001,
			Value: &FloatValue{
				Value:   123.456,
				Quality: Quality{},
			},
		},
		{
			Address: 5002,
			Value: &FloatValue{
				Value:   -789.012,
				Quality: Quality{Overflow: true},
			},
		},
	}

	data, err := coder.EncodeFloatValue(objects, false, 0)
	if err != nil {
		t.Fatalf("EncodeFloatValue() error: %v", err)
	}

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_MEASURE_VALUE_FLOAT {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_MEASURE_VALUE_FLOAT)
	}

	fv0, ok := asdu.Information[0].Value.(*FloatValue)
	if !ok {
		t.Fatalf("Value[0] type mismatch")
	}
	// 允许浮点精度误差
	if fv0.Value < 123.4 || fv0.Value > 123.5 {
		t.Errorf("Value[0] mismatch: got %f, want ~123.456", fv0.Value)
	}

	fv1, ok := asdu.Information[1].Value.(*FloatValue)
	if !ok {
		t.Fatalf("Value[1] type mismatch")
	}
	if fv1.Value > -788.9 || fv1.Value < -789.1 {
		t.Errorf("Value[1] mismatch: got %f, want ~-789.012", fv1.Value)
	}
}

// ==================== 遥控命令测试 ====================

func TestSingleCommandEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	cmd := &SingleCommand{
		Select: false,
		QU:     0,
		On:     true,
	}

	data := coder.EncodeSingleCommand(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_SINGLE_COMMAND {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_SINGLE_COMMAND)
	}

	if asdu.COT.Cause != COT_ACTIVATION {
		t.Errorf("COT mismatch: got %d, want %d", asdu.COT.Cause, COT_ACTIVATION)
	}

	if len(asdu.Information) != 1 {
		t.Fatalf("Information count mismatch: got %d, want 1", len(asdu.Information))
	}

	sc, ok := asdu.Information[0].Value.(*SingleCommand)
	if !ok {
		t.Fatalf("Value type mismatch")
	}

	if sc.On != true {
		t.Errorf("On mismatch: got %v, want true", sc.On)
	}
	if sc.Select != false {
		t.Errorf("Select mismatch: got %v, want false", sc.Select)
	}
}

func TestDoubleCommandEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	cmd := &DoubleCommand{
		Select: true,
		QU:     0,
		State:  DP_ON,
	}

	data := coder.EncodeDoubleCommand(1, 2001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	dc, ok := asdu.Information[0].Value.(*DoubleCommand)
	if !ok {
		t.Fatalf("Value type mismatch")
	}

	if dc.State != DP_ON {
		t.Errorf("State mismatch: got %d, want %d", dc.State, DP_ON)
	}
	if dc.Select != true {
		t.Errorf("Select mismatch: got %v, want true", dc.Select)
	}
}

// ==================== 总召唤命令测试 ====================

func TestInterrogationCommandEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	cmd := &InterrogationCommand{QOI: QOI_STATION_INTERROGATION}
	data := coder.EncodeInterrogationCommand(1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_INTERROGATION_CMD {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_INTERROGATION_CMD)
	}

	if asdu.COT.Cause != COT_ACTIVATION {
		t.Errorf("COT mismatch: got %d, want %d", asdu.COT.Cause, COT_ACTIVATION)
	}

	ic, ok := asdu.Information[0].Value.(*InterrogationCommand)
	if !ok {
		t.Fatalf("Value type mismatch")
	}

	if ic.QOI != QOI_STATION_INTERROGATION {
		t.Errorf("QOI mismatch: got %d, want %d", ic.QOI, QOI_STATION_INTERROGATION)
	}
}

// ==================== 时钟同步命令测试 ====================

func TestClockSyncCommandEncodeDecode(t *testing.T) {
	coder := NewASDUCoder()

	// 使用2010年的时间，因为CP56Time2a格式只支持2000-2015年
	now := time.Date(2010, 4, 18, 7, 10, 20, 673458881, time.UTC)
	cmd := &ClockSyncCommand{Time: now}
	data := coder.EncodeClockSyncCommand(1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	if err != nil {
		t.Fatalf("DecodeASDU() error: %v", err)
	}

	if asdu.TypeID != TYPE_ID_CLOCK_SYNC_CMD {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu.TypeID, TYPE_ID_CLOCK_SYNC_CMD)
	}

	csc, ok := asdu.Information[0].Value.(*ClockSyncCommand)
	if !ok {
		t.Fatalf("Value type mismatch")
	}

	// 验证时间(允许一定误差)
	diff := csc.Time.Sub(now)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Time mismatch: got %v, want ~%v", csc.Time, now)
	}
}

// ==================== 带时标数据测试 ====================

func TestSinglePointInfoWithTime(t *testing.T) {
	coder := NewASDUCoder()

	now := time.Now()
	objects := []InformationObject{
		{
			Address:   1001,
			Value:     &SinglePointInfo{Value: true, Quality: Quality{}},
			Timestamp: now,
		},
	}

	// CP24时标
	data24, err := coder.EncodeSinglePointInfo(objects, true, 3)
	if err != nil {
		t.Fatalf("EncodeSinglePointInfo(CP24) error: %v", err)
	}

	asdu24, err := coder.DecodeASDU(data24)
	if err != nil {
		t.Fatalf("DecodeASDU(CP24) error: %v", err)
	}

	if asdu24.TypeID != TYPE_ID_SINGLE_POINT_INFO_TIME {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu24.TypeID, TYPE_ID_SINGLE_POINT_INFO_TIME)
	}

	// CP56时标
	data56, err := coder.EncodeSinglePointInfo(objects, true, 7)
	if err != nil {
		t.Fatalf("EncodeSinglePointInfo(CP56) error: %v", err)
	}

	asdu56, err := coder.DecodeASDU(data56)
	if err != nil {
		t.Fatalf("DecodeASDU(CP56) error: %v", err)
	}

	if asdu56.TypeID != TYPE_ID_SINGLE_POINT_INFO_TIME_CP56 {
		t.Errorf("TypeID mismatch: got %d, want %d", asdu56.TypeID, TYPE_ID_SINGLE_POINT_INFO_TIME_CP56)
	}
}

// ==================== 连接状态测试 ====================

func TestConnectionStateString(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{STATE_DISCONNECTED, "DISCONNECTED"},
		{STATE_CONNECTING, "CONNECTING"},
		{STATE_CONNECTED, "CONNECTED"},
		{STATE_STARTDT_SENT, "STARTDT_SENT"},
		{STATE_ACTIVE, "ACTIVE"},
		{STATE_STOPPING, "STOPPING"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("String() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// ==================== 类型判断函数测试 ====================

func TestTypeIdentification(t *testing.T) {
	// 单点遥信
	if !IsSinglePointInfo(TYPE_ID_SINGLE_POINT_INFO) {
		t.Error("IsSinglePointInfo(TYPE_ID_SINGLE_POINT_INFO) should be true")
	}
	if !IsSinglePointInfo(TYPE_ID_SINGLE_POINT_INFO_TIME) {
		t.Error("IsSinglePointInfo(TYPE_ID_SINGLE_POINT_INFO_TIME) should be true")
	}

	// 双点遥信
	if !IsDoublePointInfo(TYPE_ID_DOUBLE_POINT_INFO) {
		t.Error("IsDoublePointInfo(TYPE_ID_DOUBLE_POINT_INFO) should be true")
	}

	// 遥测
	if !IsMeasureValue(TYPE_ID_MEASURE_VALUE_NORMAL) {
		t.Error("IsMeasureValue(TYPE_ID_MEASURE_VALUE_NORMAL) should be true")
	}
	if !IsMeasureValue(TYPE_ID_MEASURE_VALUE_FLOAT) {
		t.Error("IsMeasureValue(TYPE_ID_MEASURE_VALUE_FLOAT) should be true")
	}

	// 电度
	if !IsIntegratedTotal(TYPE_ID_INTEGRITY_TOTAL) {
		t.Error("IsIntegratedTotal(TYPE_ID_INTEGRITY_TOTAL) should be true")
	}

	// 时标
	if !HasCP24Time(TYPE_ID_SINGLE_POINT_INFO_TIME) {
		t.Error("HasCP24Time(TYPE_ID_SINGLE_POINT_INFO_TIME) should be true")
	}
	if !HasCP56Time(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56) {
		t.Error("HasCP56Time(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56) should be true")
	}

	// 控制方向
	if !IsControlDirection(TYPE_ID_SINGLE_COMMAND) {
		t.Error("IsControlDirection(TYPE_ID_SINGLE_COMMAND) should be true")
	}
}

// ==================== 辅助函数测试 ====================

func TestGetInfoObjectValueString(t *testing.T) {
	tests := []struct {
		name     string
		obj      InformationObject
		expected string
	}{
		{
			name:     "SinglePoint ON",
			obj:      InformationObject{Value: &SinglePointInfo{Value: true}},
			expected: "ON",
		},
		{
			name:     "SinglePoint OFF",
			obj:      InformationObject{Value: &SinglePointInfo{Value: false}},
			expected: "OFF",
		},
		{
			name:     "DoublePoint ON",
			obj:      InformationObject{Value: &DoublePointInfo{Value: DP_ON}},
			expected: "ON",
		},
		{
			name:     "DoublePoint OFF",
			obj:      InformationObject{Value: &DoublePointInfo{Value: DP_OFF}},
			expected: "OFF",
		},
		{
			name:     "ScaledValue",
			obj:      InformationObject{Value: &ScaledValue{Value: 100}},
			expected: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetInfoObjectValueString(tt.obj)
			if got != tt.expected {
				t.Errorf("GetInfoObjectValueString() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestGetInfoObjectQualityString(t *testing.T) {
	tests := []struct {
		name     string
		obj      InformationObject
		expected string
	}{
		{
			name:     "Good quality",
			obj:      InformationObject{Value: &SinglePointInfo{Quality: Quality{}}},
			expected: "OK",
		},
		{
			name:     "Invalid",
			obj:      InformationObject{Value: &SinglePointInfo{Quality: Quality{Invalid: true}}},
			expected: "IV",
		},
		{
			name:     "Invalid+Blocked",
			obj:      InformationObject{Value: &SinglePointInfo{Quality: Quality{Invalid: true, Blocked: true}}},
			expected: "IV|BL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetInfoObjectQualityString(tt.obj)
			if got != tt.expected {
				t.Errorf("GetInfoObjectQualityString() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// ==================== 基准测试 ====================

func BenchmarkEncodeSinglePointInfo(b *testing.B) {
	coder := NewASDUCoder()

	objects := make([]InformationObject, 100)
	for i := range objects {
		objects[i] = InformationObject{
			// 安全转换：确保地址在uint32范围内
			Address: uint32(min(1000+i, math.MaxUint32)),
			Value:   &SinglePointInfo{Value: i%2 == 0, Quality: Quality{}},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coder.EncodeSinglePointInfo(objects, false, 0)
	}
}

func BenchmarkDecodeASDU(b *testing.B) {
	coder := NewASDUCoder()

	objects := make([]InformationObject, 100)
	for i := range objects {
		objects[i] = InformationObject{
			Address: uint32(1000 + i),
			Value:   &SinglePointInfo{Value: i%2 == 0, Quality: Quality{}},
		}
	}

	data, _ := coder.EncodeSinglePointInfo(objects, false, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coder.DecodeASDU(data)
	}
}

func BenchmarkCP56Time2aEncode(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeCP56Time2a(now)
	}
}

func BenchmarkCP56Time2aDecode(b *testing.B) {
	data := EncodeCP56Time2a(time.Now())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeCP56Time2a(data)
	}
}
