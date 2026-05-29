package iec104

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionState_StringCov(t *testing.T) {
	assert.Equal(t, "DISCONNECTED", STATE_DISCONNECTED.String())
	assert.Equal(t, "CONNECTING", STATE_CONNECTING.String())
	assert.Equal(t, "CONNECTED", STATE_CONNECTED.String())
	assert.Equal(t, "ACTIVE", STATE_ACTIVE.String())
	assert.Equal(t, "STOPPING", STATE_STOPPING.String())
	assert.Equal(t, "UNKNOWN", ConnectionState(6).String())
	assert.Equal(t, "UNKNOWN", ConnectionState(99).String())
}

func TestCP24Time2a_EncodeDecodeCoverage(t *testing.T) {
	now := time.Date(2025, 3, 15, 14, 30, 45, 123000000, time.UTC)
	encoded := EncodeCP24Time2a(now)
	decoded := DecodeCP24Time2a(encoded)
	assert.Equal(t, uint8(30), decoded.Minutes)
}

func TestCP56Time2a_EncodeDecodeCoverage(t *testing.T) {
	now := time.Date(2025, 3, 15, 14, 30, 45, 123000000, time.UTC)
	encoded := EncodeCP56Time2a(now)
	decoded, err := DecodeCP56Time2a(encoded)
	require.NoError(t, err)
	assert.Equal(t, uint8(9), decoded.Year)
	assert.Equal(t, uint8(3), decoded.Month)
	assert.Equal(t, uint8(15), decoded.DayOfMonth)
	assert.Equal(t, uint8(14), decoded.Hours)
	assert.Equal(t, uint8(30), decoded.Minutes)
}

func TestCP56Time2a_InvalidLengthCoverage(t *testing.T) {
	_, err := DecodeCP56Time2a([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestCP24Time2a_InvalidLengthCoverage(t *testing.T) {
	decoded := DecodeCP24Time2a([]byte{0x01})
	assert.Equal(t, CP24Time2a{}, decoded)
}

func TestParseQuality(t *testing.T) {
	q := ParseQuality(0x10)
	assert.True(t, q.Blocked)

	q = ParseQuality(0x20)
	assert.True(t, q.Substituted)

	q = ParseQuality(0x40)
	assert.True(t, q.NotCurrent)

	q = ParseQuality(0x80)
	assert.True(t, q.Invalid)

	mq := ParseMeasureQuality(0x01)
	assert.True(t, mq.Overflow)
}

func TestEncodeQuality(t *testing.T) {
	q := Quality{Invalid: true, Blocked: true}
	b := EncodeQuality(q)
	assert.True(t, b&0x80 != 0)
	assert.True(t, b&0x10 != 0)
}

func TestParseMeasureQuality(t *testing.T) {
	q := ParseMeasureQuality(0x01)
	assert.True(t, q.Overflow)

	q = ParseMeasureQuality(0x80)
	assert.True(t, q.Invalid)
}

func TestEncodeMeasureQuality(t *testing.T) {
	q := Quality{Invalid: true, Overflow: true}
	b := EncodeMeasureQuality(q)
	assert.True(t, b&0x80 != 0)
	assert.True(t, b&0x01 != 0)
}

func TestParseVSQ(t *testing.T) {
	vsq := ParseVSQ(0x01)
	assert.Equal(t, uint8(1), vsq.Number)
	assert.False(t, vsq.IsSequence)

	vsq = ParseVSQ(0x81)
	assert.Equal(t, uint8(1), vsq.Number)
	assert.True(t, vsq.IsSequence)
}

func TestEncodeVSQ(t *testing.T) {
	vsq := VSQ{Number: 5, IsSequence: true}
	b := EncodeVSQ(vsq)
	assert.Equal(t, uint8(0x85), b)

	vsq = VSQ{Number: 3, IsSequence: false}
	b = EncodeVSQ(vsq)
	assert.Equal(t, uint8(0x03), b)
}

func TestParseCOT(t *testing.T) {
	cot := ParseCOT(0x03, 0x00)
	assert.Equal(t, uint8(3), cot.Cause)
	assert.False(t, cot.IsTest)
	assert.False(t, cot.IsPN)

	cot = ParseCOT(0x83, 0x00)
	assert.Equal(t, uint8(3), cot.Cause)
	assert.True(t, cot.IsTest)
	assert.False(t, cot.IsPN)

	cot = ParseCOT(0x43, 0x00)
	assert.Equal(t, uint8(3), cot.Cause)
	assert.False(t, cot.IsTest)
	assert.True(t, cot.IsPN)
}

func TestEncodeCOT(t *testing.T) {
	cot := CauseOfTransmission{Cause: 3, IsTest: true, IsPN: true}
	b1, b2 := EncodeCOT(cot)
	assert.True(t, b1&0x80 != 0)
	assert.True(t, b1&0x40 != 0)
	_ = b2
}

func TestIsSinglePointInfo(t *testing.T) {
	assert.True(t, IsSinglePointInfo(TYPE_ID_SINGLE_POINT_INFO))
	assert.True(t, IsSinglePointInfo(TYPE_ID_SINGLE_POINT_INFO_TIME))
	assert.False(t, IsSinglePointInfo(TYPE_ID_DOUBLE_POINT_INFO))
}

func TestIsDoublePointInfo(t *testing.T) {
	assert.True(t, IsDoublePointInfo(TYPE_ID_DOUBLE_POINT_INFO))
	assert.True(t, IsDoublePointInfo(TYPE_ID_DOUBLE_POINT_INFO_TIME))
	assert.False(t, IsDoublePointInfo(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsMeasureValue(t *testing.T) {
	assert.True(t, IsMeasureValue(TYPE_ID_MEASURE_VALUE_NORMAL))
	assert.True(t, IsMeasureValue(TYPE_ID_MEASURE_VALUE_SCALED))
	assert.True(t, IsMeasureValue(TYPE_ID_MEASURE_VALUE_FLOAT))
	assert.False(t, IsMeasureValue(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsIntegratedTotal(t *testing.T) {
	assert.True(t, IsIntegratedTotal(TYPE_ID_INTEGRITY_TOTAL))
	assert.False(t, IsIntegratedTotal(TYPE_ID_SINGLE_POINT_INFO))
}

func TestHasCP24Time(t *testing.T) {
	assert.True(t, HasCP24Time(TYPE_ID_SINGLE_POINT_INFO_TIME))
	assert.False(t, HasCP24Time(TYPE_ID_SINGLE_POINT_INFO))
}

func TestHasCP56Time(t *testing.T) {
	assert.True(t, HasCP56Time(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56))
	assert.False(t, HasCP56Time(TYPE_ID_SINGLE_POINT_INFO))
}

func TestHasTimestampCov(t *testing.T) {
	assert.True(t, HasTimestamp(TYPE_ID_SINGLE_POINT_INFO_TIME))
	assert.True(t, HasTimestamp(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56))
	assert.False(t, HasTimestamp(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsControlDirectionCov(t *testing.T) {
	assert.True(t, IsControlDirection(TYPE_ID_SINGLE_COMMAND))
	assert.True(t, IsControlDirection(TYPE_ID_DOUBLE_COMMAND))
	assert.False(t, IsControlDirection(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsSystemInfoCov(t *testing.T) {
	assert.True(t, IsSystemInfo(TYPE_ID_END_OF_INITIALIZATION))
	assert.False(t, IsSystemInfo(TYPE_ID_INTERROGATION_CMD))
	assert.False(t, IsSystemInfo(TYPE_ID_SINGLE_POINT_INFO))
}

func TestASDUCoder_EncodeDecode_SinglePointInfo(t *testing.T) {
	coder := NewASDUCoder()
	require.NotNil(t, coder)

	objects := []InformationObject{
		{
			Address: 1,
			Value:   &SinglePointInfo{Value: true, Quality: Quality{Invalid: false}},
		},
		{
			Address: 2,
			Value:   &SinglePointInfo{Value: false, Quality: Quality{Invalid: true}},
		},
	}

	data, err := coder.EncodeSinglePointInfo(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeInfoAddress(t *testing.T) {
	coder := NewASDUCoder()
	data := coder.EncodeInfoAddress(0x123456)
	assert.Len(t, data, 3)
}

func TestASDUCoder_EncodeSinglePointInfo(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1, Value: &SinglePointInfo{Value: true, Quality: Quality{}}},
	}
	data, err := coder.EncodeSinglePointInfo(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeSinglePointInfo_NoObjects(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeSinglePointInfo([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeDoublePointInfo(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1, Value: &DoublePointInfo{Value: DP_ON, Quality: Quality{}}},
	}
	data, err := coder.EncodeDoublePointInfo(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeDoublePointInfo_NoObjects(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeDoublePointInfo([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeNormalizedValue(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1, Value: &NormalizedValue{Value: 0.5, Quality: Quality{}}},
	}
	data, err := coder.EncodeNormalizedValue(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeNormalizedValue_NoObjects(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeNormalizedValue([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeScaledValue(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1, Value: &ScaledValue{Value: 100, Quality: Quality{}}},
	}
	data, err := coder.EncodeScaledValue(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeScaledValue_NoObjects(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeScaledValue([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeFloatValue(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1, Value: &FloatValue{Value: 3.14, Quality: Quality{}}},
	}
	data, err := coder.EncodeFloatValue(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeFloatValue_NoObjects(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeFloatValue([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeIntegratedTotal(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1, Value: &IntegratedTotal{Value: 12345, Quality: Quality{}}},
	}
	data, err := coder.EncodeIntegratedTotal(objects, false, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeIntegratedTotal_NoObjects(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeIntegratedTotal([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeSingleCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SingleCommand{On: true, Select: false, QU: 0}
	data := coder.EncodeSingleCommand(1, 1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeDoubleCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &DoubleCommand{State: DCS_ON, Select: false, QU: 0}
	data := coder.EncodeDoubleCommand(1, 1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeRegulatingStepCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &RegulatingStepCommand{State: RCS_HIGHER, Select: false, QU: 0}
	data := coder.EncodeRegulatingStepCommand(1, 1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeSetPointCommandNormal(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &NormalizedValue{Value: 0.5, Quality: Quality{}}}
	data, err := coder.EncodeSetPointCommandNormal(1, 1, cmd, CauseOfTransmission{Cause: 3})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeSetPointCommandScaled(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &ScaledValue{Value: 100, Quality: Quality{}}}
	data, err := coder.EncodeSetPointCommandScaled(1, 1, cmd, CauseOfTransmission{Cause: 3})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeSetPointCommandFloat(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &FloatValue{Value: 3.14, Quality: Quality{}}}
	data, err := coder.EncodeSetPointCommandFloat(1, 1, cmd, CauseOfTransmission{Cause: 3})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeInterrogationCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &InterrogationCommand{QOI: 20}
	data := coder.EncodeInterrogationCommand(1, cmd, CauseOfTransmission{Cause: 6})
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeCounterInterrogationCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &CounterInterrogationCommand{QCC: 5}
	data := coder.EncodeCounterInterrogationCommand(1, cmd, CauseOfTransmission{Cause: 6})
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeClockSyncCommand(t *testing.T) {
	coder := NewASDUCoder()
	now := time.Now()
	cmd := &ClockSyncCommand{Time: now}
	data := coder.EncodeClockSyncCommand(1, cmd, CauseOfTransmission{Cause: 6})
	assert.NotEmpty(t, data)
}

func TestASDUCoder_EncodeTestCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &TestCommand{FFS1: 0xAA, FFS2: 0x55}
	data := coder.EncodeTestCommand(1, cmd, CauseOfTransmission{Cause: 6})
	assert.NotEmpty(t, data)
}

func TestDecodeASDU_SinglePointInfo(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{Address: 1, Value: &SinglePointInfo{Value: true, Quality: Quality{}}},
	}
	data, err := coder.EncodeSinglePointInfo(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_SINGLE_POINT_INFO), asdu.TypeID)
	assert.Len(t, asdu.Information, 1)
}

func TestDecodeASDU_DoublePointInfo(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{Address: 1, Value: &DoublePointInfo{Value: DP_ON, Quality: Quality{}}},
	}
	data, err := coder.EncodeDoublePointInfo(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_DOUBLE_POINT_INFO), asdu.TypeID)
}

func TestDecodeASDU_NormalizedValue(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{Address: 1, Value: &NormalizedValue{Value: 0.5, Quality: Quality{}}},
	}
	data, err := coder.EncodeNormalizedValue(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_NORMAL), asdu.TypeID)
}

func TestDecodeASDU_ScaledValue(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{Address: 1, Value: &ScaledValue{Value: 100, Quality: Quality{}}},
	}
	data, err := coder.EncodeScaledValue(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_SCALED), asdu.TypeID)
}

func TestDecodeASDU_FloatValue(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{Address: 1, Value: &FloatValue{Value: 3.14, Quality: Quality{}}},
	}
	data, err := coder.EncodeFloatValue(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_FLOAT), asdu.TypeID)
}

func TestDecodeASDU_IntegratedTotal(t *testing.T) {
	coder := NewASDUCoder()

	objects := []InformationObject{
		{Address: 1, Value: &IntegratedTotal{Value: 12345, Quality: Quality{}}},
	}
	data, err := coder.EncodeIntegratedTotal(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_INTEGRITY_TOTAL), asdu.TypeID)
}

func TestDecodeASDU_SingleCommand(t *testing.T) {
	coder := NewASDUCoder()

	cmd := &SingleCommand{On: true, Select: false, QU: 0}
	header := coder.EncodeASDUHeader(TYPE_ID_SINGLE_COMMAND, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_ACTIVATION}, 1)
	ioAddr := coder.EncodeInfoAddress(1)
	cmdData := coder.EncodeSingleCommand(1, 1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	fullData := append(header, ioAddr...)
	fullData = append(fullData, cmdData...)

	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_SINGLE_COMMAND), asdu.TypeID)
}

func TestDecodeASDU_DoubleCommand(t *testing.T) {
	coder := NewASDUCoder()

	cmd := &DoubleCommand{State: DCS_ON, Select: false, QU: 0}
	header := coder.EncodeASDUHeader(TYPE_ID_DOUBLE_COMMAND, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_ACTIVATION}, 1)
	ioAddr := coder.EncodeInfoAddress(1)
	cmdData := coder.EncodeDoubleCommand(1, 1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	fullData := append(header, ioAddr...)
	fullData = append(fullData, cmdData...)

	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_DOUBLE_COMMAND), asdu.TypeID)
}

func TestDecodeASDU_InterrogationCommand(t *testing.T) {
	coder := NewASDUCoder()

	header := coder.EncodeASDUHeader(TYPE_ID_INTERROGATION_CMD, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_ACTIVATION}, 1)
	ioAddr := coder.EncodeInfoAddress(0)
	cmdData := coder.EncodeInterrogationCommand(1, &InterrogationCommand{QOI: 20}, CauseOfTransmission{Cause: COT_ACTIVATION})

	fullData := append(header, ioAddr...)
	fullData = append(fullData, cmdData...)

	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_INTERROGATION_CMD), asdu.TypeID)
}

func TestDecodeASDU_CounterInterrogationCommand(t *testing.T) {
	coder := NewASDUCoder()

	header := coder.EncodeASDUHeader(TYPE_ID_COUNTER_INTERROGATION_CMD, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_ACTIVATION}, 1)
	ioAddr := coder.EncodeInfoAddress(0)
	cmdData := coder.EncodeCounterInterrogationCommand(1, &CounterInterrogationCommand{QCC: 5}, CauseOfTransmission{Cause: COT_ACTIVATION})

	fullData := append(header, ioAddr...)
	fullData = append(fullData, cmdData...)

	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_COUNTER_INTERROGATION_CMD), asdu.TypeID)
}

func TestDecodeASDU_ClockSyncCommand(t *testing.T) {
	coder := NewASDUCoder()

	header := coder.EncodeASDUHeader(TYPE_ID_CLOCK_SYNC_CMD, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_ACTIVATION}, 1)
	ioAddr := coder.EncodeInfoAddress(0)
	cmdData := coder.EncodeClockSyncCommand(1, &ClockSyncCommand{Time: time.Now()}, CauseOfTransmission{Cause: COT_ACTIVATION})

	fullData := append(header, ioAddr...)
	fullData = append(fullData, cmdData...)

	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_CLOCK_SYNC_CMD), asdu.TypeID)
}

func TestDecodeASDU_TestCommand(t *testing.T) {
	coder := NewASDUCoder()

	header := coder.EncodeASDUHeader(TYPE_ID_TEST_COMMAND, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 1)
	ioAddr := coder.EncodeInfoAddress(0)
	cmdData := coder.EncodeTestCommand(1, &TestCommand{FFS1: 0xAA, FFS2: 0x55}, CauseOfTransmission{Cause: COT_SPONTANEOUS})

	fullData := append(header, ioAddr...)
	fullData = append(fullData, cmdData...)

	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_TEST_COMMAND), asdu.TypeID)
}

func TestDecodeASDU_EndOfInitialization(t *testing.T) {
	coder := NewASDUCoder()

	header := coder.EncodeASDUHeader(TYPE_ID_END_OF_INITIALIZATION, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_INITIALIZED}, 1)
	ioAddr := coder.EncodeInfoAddress(0)
	ioData := append(ioAddr, byte(0x00))

	fullData := append(header, ioData...)
	asdu, err := coder.DecodeASDU(fullData)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_END_OF_INITIALIZATION), asdu.TypeID)
}

func TestDecodeASDU_UnknownType(t *testing.T) {
	coder := NewASDUCoder()

	header := coder.EncodeASDUHeader(0xFE, VSQ{Number: 1, IsSequence: true}, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 1)
	ioAddr := coder.EncodeInfoAddress(0)
	ioData := append(ioAddr, []byte{0x00, 0x00}...)

	fullData := append(header, ioData...)
	_, err := coder.DecodeASDU(fullData)
	assert.Error(t, err)
}

func TestDecodeASDU_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.DecodeASDU([]byte{0x01})
	assert.Error(t, err)
}

func TestGetInfoObjectValueStringCov(t *testing.T) {
	assert.Equal(t, "ON", GetInfoObjectValueString(InformationObject{Value: &SinglePointInfo{Value: true}}))
	assert.Equal(t, "OFF", GetInfoObjectValueString(InformationObject{Value: &SinglePointInfo{Value: false}}))
	assert.Equal(t, "ON", GetInfoObjectValueString(InformationObject{Value: &DoublePointInfo{Value: DP_ON}}))
	assert.Equal(t, "OFF", GetInfoObjectValueString(InformationObject{Value: &DoublePointInfo{Value: DP_OFF}}))
	assert.Equal(t, "100", GetInfoObjectValueString(InformationObject{Value: &ScaledValue{Value: 100}}))
	assert.Contains(t, GetInfoObjectValueString(InformationObject{Value: &FloatValue{Value: 3.14}}), "3.14")
	assert.Contains(t, GetInfoObjectValueString(InformationObject{Value: &NormalizedValue{Value: 0.5}}), "0.5")
	assert.Contains(t, GetInfoObjectValueString(InformationObject{Value: &IntegratedTotal{Value: 12345}}), "12345")
}

func TestGetInfoObjectQualityStringCov(t *testing.T) {
	assert.Equal(t, "OK", GetInfoObjectQualityString(InformationObject{Value: &SinglePointInfo{Quality: Quality{}}}))
	assert.Contains(t, GetInfoObjectQualityString(InformationObject{Value: &SinglePointInfo{Quality: Quality{Invalid: true}}}), "IV")
	assert.Equal(t, "OK", GetInfoObjectQualityString(InformationObject{Value: &FloatValue{Quality: Quality{}}}))
	assert.Contains(t, GetInfoObjectQualityString(InformationObject{Value: &FloatValue{Quality: Quality{Invalid: true}}}), "IV")
	assert.Equal(t, "", GetInfoObjectQualityString(InformationObject{}))
}

func TestConnectionConfig_Default(t *testing.T) {
	config := DefaultConnectionConfig()
	assert.Equal(t, 0, config.MaxReconnectAttempts)
	assert.Equal(t, 5*time.Second, config.ReconnectInterval)
}

func TestConnection_NewCov(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	require.NotNil(t, conn)
	assert.Equal(t, STATE_DISCONNECTED, conn.GetState())
	assert.False(t, conn.IsConnected())
	assert.False(t, conn.IsActive())
}

func TestConnection_SetCommonAddress(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	conn.SetCommonAddress(2)
}

func TestConnection_GetStatsCov(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	stats := conn.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, uint64(0), stats.FramesSent)
	assert.Equal(t, uint64(0), stats.FramesReceived)
}

func TestConnection_GetSeqNumbers(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	assert.Equal(t, uint16(0), conn.GetSendSeq())
	assert.Equal(t, uint16(0), conn.GetRecvSeq())
}

func TestConnection_UpdateRecvSeqCov(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	conn.UpdateRecvSeq(5)
	assert.Equal(t, uint16(5), conn.GetRecvSeq())
}

func TestConnection_DisconnectNotConnected(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	err := conn.Disconnect()
	assert.NoError(t, err)
}

func TestConnection_StopResumeReconnectCov(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	conn.StopReconnect()
	conn.ResumeReconnect()
}

func TestConnectionPool_NewCov(t *testing.T) {
	pool := NewConnectionPool()
	require.NotNil(t, pool)
}

func TestConnectionPool_AddConfig(t *testing.T) {
	pool := NewConnectionPool()
	config := DefaultConnectionConfig()
	pool.AddConfig("conn1", config)
}

func TestConnectionPool_DisconnectAllCov(t *testing.T) {
	pool := NewConnectionPool()
	pool.DisconnectAll()
}

func TestConnectionPool_RemoveCov(t *testing.T) {
	pool := NewConnectionPool()
	pool.Remove("nonexistent")
}

func TestAPDU_StructCov(t *testing.T) {
	apdu := &APDU{
		Type: APDU_TYPE_I,
		ASDU: &ASDU{
			TypeID:     TYPE_ID_SINGLE_POINT_INFO,
			VSQ:        VSQ{Number: 1},
			COT:        CauseOfTransmission{Cause: COT_SPONTANEOUS},
			CommonAddr: 1,
		},
	}
	assert.Equal(t, APDU_TYPE_I, apdu.Type)
}

func TestASDU_Struct(t *testing.T) {
	asdu := &ASDU{
		TypeID:     TYPE_ID_SINGLE_POINT_INFO,
		VSQ:        VSQ{Number: 1, IsSequence: true},
		COT:        CauseOfTransmission{Cause: COT_SPONTANEOUS, IsPN: true},
		CommonAddr: 1,
		Information: []InformationObject{
			{Address: 1, Value: &SinglePointInfo{Value: true}},
		},
	}
	assert.Equal(t, uint8(TYPE_ID_SINGLE_POINT_INFO), asdu.TypeID)
	assert.Len(t, asdu.Information, 1)
}

func TestInformationObject_Struct(t *testing.T) {
	io := InformationObject{
		Address: 1234,
		Value:   &SinglePointInfo{Value: true, Quality: Quality{}},
	}
	assert.Equal(t, uint32(1234), io.Address)
}

func TestQuality_Struct(t *testing.T) {
	q := Quality{
		Overflow:    true,
		Blocked:     false,
		Substituted: false,
		NotCurrent:  false,
		Invalid:     true,
	}
	assert.True(t, q.Overflow)
	assert.True(t, q.Invalid)
}

func TestSinglePointInfo_Struct(t *testing.T) {
	info := &SinglePointInfo{Value: true, Quality: Quality{Invalid: false}}
	assert.True(t, info.Value)
}

func TestDoublePointInfo_Values(t *testing.T) {
	assert.Equal(t, 0, DP_INDETERMINATE)
	assert.Equal(t, 1, DP_OFF)
	assert.Equal(t, 2, DP_ON)
	assert.Equal(t, 3, DP_INDETERMINATE2)
}

func TestStepPositionInfo_Struct(t *testing.T) {
	info := &StepPositionInfo{Value: 5, Transient: true, Quality: Quality{}}
	assert.Equal(t, int16(5), info.Value)
	assert.True(t, info.Transient)
}

func TestBitstring32_Struct(t *testing.T) {
	info := &Bitstring32{Value: 0x80000000, Quality: Quality{}}
	assert.Equal(t, uint32(0x80000000), info.Value)
}

func TestNormalizedValue_Struct(t *testing.T) {
	val := &NormalizedValue{Value: 0.75, Quality: Quality{}}
	assert.InDelta(t, 0.75, val.Value, 0.001)
}

func TestScaledValue_Struct(t *testing.T) {
	val := &ScaledValue{Value: -100, Quality: Quality{}}
	assert.Equal(t, int16(-100), val.Value)
}

func TestFloatValue_Struct(t *testing.T) {
	val := &FloatValue{Value: 3.14159, Quality: Quality{}}
	assert.InDelta(t, 3.14159, val.Value, 0.0001)
}

func TestIntegratedTotal_Struct(t *testing.T) {
	val := &IntegratedTotal{Value: 99999, Quality: Quality{}}
	assert.Equal(t, uint32(99999), val.Value)
}

func TestSingleCommand_Struct(t *testing.T) {
	cmd := &SingleCommand{On: true, Select: true, QU: 1}
	assert.True(t, cmd.On)
	assert.True(t, cmd.Select)
}

func TestDoubleCommand_Values(t *testing.T) {
	assert.Equal(t, uint8(0x00), uint8(DCS_OFF))
	assert.Equal(t, uint8(0x01), uint8(DCS_ON))
}

func TestRegulatingStepCommand_Values(t *testing.T) {
	assert.Equal(t, 0, RCS_STOP)
	assert.Equal(t, 1, RCS_LOWER)
	assert.Equal(t, 2, RCS_HIGHER)
}

func TestSetPointCommand_Struct(t *testing.T) {
	cmd := &SetPointCommand{
		Select: false,
		Value:  float32(42.5),
	}
	assert.Equal(t, float32(42.5), cmd.Value)
}

func TestInterrogationCommand_Struct(t *testing.T) {
	cmd := &InterrogationCommand{QOI: 20}
	assert.Equal(t, uint8(20), cmd.QOI)
}

func TestCounterInterrogationCommand_Struct(t *testing.T) {
	cmd := &CounterInterrogationCommand{QCC: 5}
	assert.Equal(t, uint8(5), cmd.QCC)
}

func TestClockSyncCommand_Struct(t *testing.T) {
	now := time.Now()
	cmd := &ClockSyncCommand{Time: now}
	assert.Equal(t, now, cmd.Time)
}

func TestTestCommand_Struct(t *testing.T) {
	cmd := &TestCommand{}
	assert.NotNil(t, cmd)
}

func TestEndOfInitialization_Struct(t *testing.T) {
	cmd := &EndOfInitialization{COI: 0}
	assert.Equal(t, uint8(0), cmd.COI)
}

func TestMeasureQuality_Struct(t *testing.T) {
	q := Quality{
		Overflow: true,
		Invalid:  true,
	}
	assert.True(t, q.Overflow)
	assert.True(t, q.Invalid)
}
