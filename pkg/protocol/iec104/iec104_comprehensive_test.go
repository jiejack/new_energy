package iec104

import (
	"math"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionState_String_Unknown(t *testing.T) {
	s := ConnectionState(99)
	assert.Equal(t, "UNKNOWN", s.String())
}

func TestCP56Time2a_InsufficientData(t *testing.T) {
	_, err := DecodeCP56Time2a([]byte{0x01, 0x02, 0x03})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestCP56Time2a_ExactData(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	cp56, err := DecodeCP56Time2a(data)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0x0201), cp56.Milliseconds)
}

func TestCP56Time2a_ToTime_Year1900(t *testing.T) {
	cp56 := CP56Time2a{
		Milliseconds: 0,
		Minutes:      0,
		Hours:        0,
		DayOfMonth:   1,
		Month:        1,
		Year:         70,
	}
	tm := cp56.ToTime()
	assert.Equal(t, 1970, tm.Year())
}

func TestCP56Time2a_ToTime_Year2000(t *testing.T) {
	cp56 := CP56Time2a{
		Milliseconds: 0,
		Minutes:      0,
		Hours:        0,
		DayOfMonth:   1,
		Month:        1,
		Year:         15,
	}
	tm := cp56.ToTime()
	assert.Equal(t, 2015, tm.Year())
}

func TestCP24Time2a_InsufficientData(t *testing.T) {
	result := DecodeCP24Time2a([]byte{0x01})
	assert.Equal(t, CP24Time2a{}, result)
}

func TestCP24Time2a_ExactData(t *testing.T) {
	data := []byte{0x01, 0x02, 0x83}
	result := DecodeCP24Time2a(data)
	assert.Equal(t, uint16(0x0201), result.Milliseconds)
	assert.Equal(t, uint8(3), result.Minutes)
	assert.True(t, result.SummerTime)
}

func TestASDUCoder_DecodeASDU_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.DecodeASDU([]byte{0x01, 0x02, 0x03})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestASDUCoder_DecodeInformationObjects_ZeroNumber(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 0, IsSequence: false}
	objects, err := coder.DecodeInformationObjects(TYPE_ID_SINGLE_POINT_INFO, vsq, []byte{})
	assert.NoError(t, err)
	assert.Nil(t, objects)
}

func TestASDUCoder_DecodeInformationObjects_SequenceMode(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true, Quality: Quality{}}},
		{Address: 1002, Value: &SinglePointInfo{Value: false, Quality: Quality{}}},
	}
	data, err := coder.EncodeSinglePointInfo(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)

	vsq := VSQ{Number: asdu.VSQ.Number, IsSequence: true}
	infoData := data[6:]
	decoded, err := coder.DecodeInformationObjects(asdu.TypeID, vsq, infoData)
	assert.NoError(t, err)
	assert.Len(t, decoded, int(vsq.Number))
}

func TestASDUCoder_DecodeInformationObjects_SequenceInsufficientAddr(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: true}
	_, err := coder.DecodeInformationObjects(TYPE_ID_SINGLE_POINT_INFO, vsq, []byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestASDUCoder_DecodeInformationObjects_NonSequenceInsufficientAddr(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	_, err := coder.DecodeInformationObjects(TYPE_ID_SINGLE_POINT_INFO, vsq, []byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestASDUCoder_DecodeSingleInformationObject_UnsupportedType(t *testing.T) {
	coder := NewASDUCoder()
	_, _, err := coder.DecodeSingleInformationObject(255, 0, []byte{}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported type ID")
}

func TestASDUCoder_EncodeSinglePointInfo_Empty(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeSinglePointInfo([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeSinglePointInfo_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &FloatValue{Value: 1.0}},
	}
	_, err := coder.EncodeSinglePointInfo(objects, false, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value type")
}

func TestASDUCoder_EncodeDoublePointInfo_Empty(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeDoublePointInfo([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeDoublePointInfo_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &FloatValue{Value: 1.0}},
	}
	_, err := coder.EncodeDoublePointInfo(objects, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeNormalizedValue_Empty(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeNormalizedValue([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeNormalizedValue_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true}},
	}
	_, err := coder.EncodeNormalizedValue(objects, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeScaledValue_Empty(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeScaledValue([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeScaledValue_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true}},
	}
	_, err := coder.EncodeScaledValue(objects, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeFloatValue_Empty(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeFloatValue([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeFloatValue_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true}},
	}
	_, err := coder.EncodeFloatValue(objects, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeIntegratedTotal_Empty(t *testing.T) {
	coder := NewASDUCoder()
	_, err := coder.EncodeIntegratedTotal([]InformationObject{}, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeIntegratedTotal_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true}},
	}
	_, err := coder.EncodeIntegratedTotal(objects, false, 0)
	assert.Error(t, err)
}

func TestASDUCoder_EncodeSetPointCommandNormal_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &FloatValue{Value: 1.0}}
	_, err := coder.EncodeSetPointCommandNormal(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	assert.Error(t, err)
}

func TestASDUCoder_EncodeSetPointCommandScaled_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &FloatValue{Value: 1.0}}
	_, err := coder.EncodeSetPointCommandScaled(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	assert.Error(t, err)
}

func TestASDUCoder_EncodeSetPointCommandFloat_InvalidType(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &NormalizedValue{Value: 0.5}}
	_, err := coder.EncodeSetPointCommandFloat(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	assert.Error(t, err)
}

func TestASDUCoder_EncodeDecode_StepPositionInfo(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_STEP_POSITION_INFO, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	infoData := append(infoAddr, 0x0A, 0x00)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_STEP_POSITION_INFO), asdu.TypeID)

	spi, ok := asdu.Information[0].Value.(*StepPositionInfo)
	require.True(t, ok)
	assert.Equal(t, int16(10), spi.Value)
}

func TestASDUCoder_EncodeDecode_IntegratedTotal(t *testing.T) {
	coder := NewASDUCoder()
	objects := []InformationObject{
		{
			Address: 6001,
			Value: &IntegratedTotal{
				Value:    12345,
				Sequence: 1,
				Quality:  Quality{},
			},
		},
	}
	data, err := coder.EncodeIntegratedTotal(objects, false, 0)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_INTEGRITY_TOTAL), asdu.TypeID)

	it, ok := asdu.Information[0].Value.(*IntegratedTotal)
	require.True(t, ok)
	assert.Equal(t, uint32(12345), it.Value)
	assert.Equal(t, uint8(1), it.Sequence)
}

func TestASDUCoder_EncodeDecode_IntegratedTotal_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{
			Address:   6001,
			Value:     &IntegratedTotal{Value: 99999, Sequence: 2, Quality: Quality{Overflow: true}},
			Timestamp: ts,
		},
	}
	data, err := coder.EncodeIntegratedTotal(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_INTEGRITY_TOTAL_TIME_CP56), asdu.TypeID)
}

func TestASDUCoder_EncodeDecode_NormalizedValue_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{
			Address:   3001,
			Value:     &NormalizedValue{Value: 0.75, Quality: Quality{}},
			Timestamp: ts,
		},
	}
	data, err := coder.EncodeNormalizedValue(objects, true, 3)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_NORMAL_TIME), asdu.TypeID)
}

func TestASDUCoder_EncodeDecode_ScaledValue_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{
			Address:   4001,
			Value:     &ScaledValue{Value: 500, Quality: Quality{}},
			Timestamp: ts,
		},
	}
	data, err := coder.EncodeScaledValue(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56), asdu.TypeID)
}

func TestASDUCoder_EncodeDecode_FloatValue_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{
			Address:   5001,
			Value:     &FloatValue{Value: 3.14, Quality: Quality{}},
			Timestamp: ts,
		},
	}
	data, err := coder.EncodeFloatValue(objects, true, 3)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_FLOAT_TIME), asdu.TypeID)
}

func TestASDUCoder_EncodeDecode_DoublePointInfo_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{
			Address:   2001,
			Value:     &DoublePointInfo{Value: DP_ON, Quality: Quality{}},
			Timestamp: ts,
		},
	}
	data, err := coder.EncodeDoublePointInfo(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56), asdu.TypeID)
}

func TestASDUCoder_RegulatingStepCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &RegulatingStepCommand{Select: false, QU: 0, State: RCS_HIGHER}
	data := coder.EncodeRegulatingStepCommand(1, 3001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_REGULATING_STEP_COMMAND), asdu.TypeID)

	rsc, ok := asdu.Information[0].Value.(*RegulatingStepCommand)
	require.True(t, ok)
	assert.Equal(t, uint8(RCS_HIGHER), rsc.State)
	assert.False(t, rsc.Select)
}

func TestASDUCoder_CounterInterrogationCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &CounterInterrogationCommand{QCC: QCC_GROUP_1}
	data := coder.EncodeCounterInterrogationCommand(1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_COUNTER_INTERROGATION_CMD), asdu.TypeID)

	cic, ok := asdu.Information[0].Value.(*CounterInterrogationCommand)
	require.True(t, ok)
	assert.Equal(t, uint8(QCC_GROUP_1), cic.QCC)
}

func TestASDUCoder_TestCommand(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &TestCommand{FFS1: 0xAA, FFS2: 0x55}
	data := coder.EncodeTestCommand(1, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_TEST_COMMAND), asdu.TypeID)

	tc, ok := asdu.Information[0].Value.(*TestCommand)
	require.True(t, ok)
	assert.Equal(t, uint8(0xAA), tc.FFS1)
	assert.Equal(t, uint8(0x55), tc.FFS2)
}

func TestASDUCoder_SetPointCommandNormal(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: true, Value: &NormalizedValue{Value: 0.5}}
	data, err := coder.EncodeSetPointCommandNormal(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_SET_POINT_COMMAND_NORMAL), asdu.TypeID)

	spc, ok := asdu.Information[0].Value.(*SetPointCommand)
	require.True(t, ok)
	assert.True(t, spc.Select)
}

func TestASDUCoder_SetPointCommandScaled(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &ScaledValue{Value: 1000}}
	data, err := coder.EncodeSetPointCommandScaled(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_SET_POINT_COMMAND_SCALED), asdu.TypeID)
}

func TestASDUCoder_SetPointCommandFloat(t *testing.T) {
	coder := NewASDUCoder()
	cmd := &SetPointCommand{Select: false, Value: &FloatValue{Value: 123.456}}
	data, err := coder.EncodeSetPointCommandFloat(1, 1001, cmd, CauseOfTransmission{Cause: COT_ACTIVATION})
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_SET_POINT_COMMAND_FLOAT), asdu.TypeID)
}

func TestASDUCoder_Decode_StepPositionInfo(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_STEP_POSITION_INFO, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	infoData := append(infoAddr, 0x0A, 0x00)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_STEP_POSITION_INFO), asdu.TypeID)

	spi, ok := asdu.Information[0].Value.(*StepPositionInfo)
	require.True(t, ok)
	assert.Equal(t, int16(10), spi.Value)
}

func TestASDUCoder_Decode_Bitstring32(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_BITSTRING32, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(7001)
	infoData := append(infoAddr, 0x01, 0x02, 0x03, 0x04, 0x00)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_BITSTRING32), asdu.TypeID)

	bs, ok := asdu.Information[0].Value.(*Bitstring32)
	require.True(t, ok)
	assert.Equal(t, uint32(0x04030201), bs.Value)
}

func TestASDUCoder_Decode_EndOfInitialization(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_INITIALIZED}
	header := coder.EncodeASDUHeader(TYPE_ID_END_OF_INITIALIZATION, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(0)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_END_OF_INITIALIZATION), asdu.TypeID)

	eoi, ok := asdu.Information[0].Value.(*EndOfInitialization)
	require.True(t, ok)
	assert.Equal(t, uint8(1), eoi.COI)
}

func TestConnection_New(t *testing.T) {
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = 2404
	conn := NewConnection(config)
	assert.NotNil(t, conn)
	assert.Equal(t, STATE_DISCONNECTED, conn.GetState())
	assert.False(t, conn.IsConnected())
	assert.False(t, conn.IsActive())
}

func TestConnection_DefaultConfig(t *testing.T) {
	config := DefaultConnectionConfig()
	assert.Equal(t, 10*time.Second, config.Timeout)
	assert.Equal(t, 5*time.Second, config.ReconnectInterval)
	assert.Equal(t, 15*time.Second, config.HeartbeatInterval)
	assert.Equal(t, 30*time.Second, config.HeartbeatTimeout)
	assert.Equal(t, 0, config.MaxReconnectAttempts)
	assert.Equal(t, 4096, config.SendBufferSize)
	assert.Equal(t, 65535, config.RecvBufferSize)
	assert.False(t, config.BalancedMode)
}

func TestConnection_Connect_Failed(t *testing.T) {
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = 19999
	config.Timeout = 1 * time.Second
	conn := NewConnection(config)
	err := conn.Connect()
	assert.Error(t, err)
	assert.Equal(t, STATE_DISCONNECTED, conn.GetState())
}

func TestConnection_Connect_AlreadyConnected(t *testing.T) {
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = 19999
	conn := NewConnection(config)
	atomicState := int32(STATE_CONNECTED)
	conn.state = atomicState
	err := conn.Connect()
	assert.Error(t, err)
	conn.state = int32(STATE_DISCONNECTED)
}

func TestConnection_Disconnect_NotConnected(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	err := conn.Disconnect()
	assert.NoError(t, err)
}

func TestConnection_GetStats(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	stats := conn.GetStats()
	assert.Equal(t, uint64(0), stats.ConnectCount)
}

func TestConnection_Channels(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	assert.NotNil(t, conn.EventChannel())
	assert.NotNil(t, conn.FrameChannel())
	assert.NotNil(t, conn.ErrorChannel())
}

func TestConnection_StopResumeReconnect(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	conn.StopReconnect()
	conn.ResumeReconnect()
}

func TestConnection_UpdateRecvSeq(t *testing.T) {
	config := DefaultConnectionConfig()
	conn := NewConnection(config)
	conn.UpdateRecvSeq(42)
	assert.Equal(t, uint16(42), conn.GetRecvSeq())
}

func TestConnection_CommonAddress(t *testing.T) {
	config := DefaultConnectionConfig()
	config.CommonAddress = 100
	conn := NewConnection(config)
	assert.Equal(t, uint16(100), conn.GetCommonAddress())
	conn.SetCommonAddress(200)
	assert.Equal(t, uint16(200), conn.GetCommonAddress())
}

func TestConnection_parseAPDU_IFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x08, 0x02, 0x00, 0x04, 0x00, 0x01, 0x01, 0x06, 0x00}
	apdu := conn.parseAPDU(frame)
	assert.NotNil(t, apdu)
	assert.Equal(t, APDU_TYPE_I, apdu.Type)
	assert.Equal(t, uint16(1), apdu.SendSeq)
	assert.Equal(t, uint16(2), apdu.RecvSeq)
}

func TestConnection_parseAPDU_SFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x04, 0x01, 0x00, 0x06, 0x00}
	apdu := conn.parseAPDU(frame)
	assert.NotNil(t, apdu)
	assert.Equal(t, APDU_TYPE_S, apdu.Type)
	assert.Equal(t, uint16(3), apdu.RecvSeq)
}

func TestConnection_parseAPDU_UFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x04, U_FRAME_TESTFR_ACT, 0x00, 0x00, 0x00}
	apdu := conn.parseAPDU(frame)
	assert.NotNil(t, apdu)
	assert.Equal(t, APDU_TYPE_U, apdu.Type)
	assert.Equal(t, byte(U_FRAME_TESTFR_ACT), apdu.Control)
}

func TestConnection_parseAPDU_ShortFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	apdu := conn.parseAPDU([]byte{0x68, 0x04})
	assert.Nil(t, apdu)
}

func TestConnection_parseAPDU_IFrameWithASDU(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x0E, 0x00, 0x00, 0x02, 0x00, 0x01, 0x01, 0x06, 0x00, 0x01, 0x00, 0x01, 0x01, 0x00, 0x00}
	apdu := conn.parseAPDU(frame)
	assert.NotNil(t, apdu)
	assert.Equal(t, APDU_TYPE_I, apdu.Type)
	assert.NotNil(t, apdu.ASDU)
}

func TestConnectionPool_New(t *testing.T) {
	pool := NewConnectionPool()
	assert.NotNil(t, pool)
}

func TestConnectionPool_Get_NotFound(t *testing.T) {
	pool := NewConnectionPool()
	_, err := pool.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConnectionPool_AddConfig_Get(t *testing.T) {
	pool := NewConnectionPool()
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = 2404
	pool.AddConfig("test_conn", config)

	conn, err := pool.Get("test_conn")
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	conn2, err := pool.Get("test_conn")
	assert.NoError(t, err)
	assert.Equal(t, conn, conn2)
}

func TestConnectionPool_Remove(t *testing.T) {
	pool := NewConnectionPool()
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	pool.AddConfig("test_conn", config)

	pool.Remove("test_conn")
	_, err := pool.Get("test_conn")
	assert.Error(t, err)
}

func TestConnectionPool_DisconnectAll(t *testing.T) {
	pool := NewConnectionPool()
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	pool.AddConfig("test1", config)
	pool.AddConfig("test2", config)
	pool.DisconnectAll()
}

func TestMaster_New(t *testing.T) {
	config := DefaultMasterConfig()
	config.Host = "127.0.0.1"
	config.Port = 2404
	master := NewMaster(config)
	assert.NotNil(t, master)
	assert.Equal(t, MASTER_STATE_STOPPED, master.GetState())
	assert.False(t, master.IsRunning())
	assert.False(t, master.IsConnected())
}

func TestMaster_DefaultConfig(t *testing.T) {
	config := DefaultMasterConfig()
	assert.Equal(t, 10*time.Second, config.Timeout)
	assert.Equal(t, 5*time.Second, config.ReconnectInterval)
	assert.Equal(t, 15*time.Second, config.HeartbeatInterval)
	assert.Equal(t, 30*time.Second, config.HeartbeatTimeout)
	assert.Equal(t, 0, config.MaxReconnectAttempts)
	assert.False(t, config.BalancedMode)
	assert.Equal(t, 10000, config.DataBufferSize)
}

func TestMaster_Start_Failed(t *testing.T) {
	config := DefaultMasterConfig()
	config.Host = "127.0.0.1"
	config.Port = 19999
	config.Timeout = 1 * time.Second
	master := NewMaster(config)
	err := master.Start()
	assert.Error(t, err)
	assert.Equal(t, MASTER_STATE_STOPPED, master.GetState())
}

func TestMaster_Stop_NotRunning(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.Stop()
	assert.NoError(t, err)
}

func TestMaster_Channels(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	assert.NotNil(t, master.DataChannel())
	assert.NotNil(t, master.ErrorChannel())
}

func TestMaster_Callbacks(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	called := false
	master.OnData(func(asdu *ASDU) { called = true })
	master.OnConnect(func() { called = true })
	master.OnDisconnect(func() { called = true })
	master.OnError(func(err error) { called = true })
	assert.False(t, called)
}

func TestMaster_Stats(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	stats := master.GetStats()
	assert.Equal(t, uint64(0), stats.TotalASDUReceived)
	assert.Equal(t, uint64(0), stats.TotalASDUSent)
	master.ResetStats()
	stats = master.GetStats()
	assert.Equal(t, uint64(0), stats.TotalASDUReceived)
}

func TestMaster_CommonAddress(t *testing.T) {
	config := DefaultMasterConfig()
	config.CommonAddress = 100
	master := NewMaster(config)
	assert.Equal(t, uint16(100), master.GetCommonAddress())
	master.SetCommonAddress(200)
	assert.Equal(t, uint16(200), master.GetCommonAddress())
}

func TestMaster_ConnectionState_NoConn(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	assert.Equal(t, STATE_DISCONNECTED, master.GetConnectionState())
}

func TestMaster_ConnectionStats_NoConn(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	stats := master.GetConnectionStats()
	assert.Equal(t, ConnectionStats{}, stats)
}

func TestMaster_Interrogation_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.Interrogation()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestMaster_InterrogationWithQOI_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.InterrogationWithQOI(QOI_STATION_INTERROGATION)
	assert.Error(t, err)
}

func TestMaster_CounterInterrogation_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.CounterInterrogation()
	assert.Error(t, err)
}

func TestMaster_ClockSync_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.ClockSync()
	assert.Error(t, err)
}

func TestMaster_SingleCommand_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.SingleCommand(1001, true)
	assert.Error(t, err)
}

func TestMaster_DoubleCommand_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.DoubleCommand(2001, DP_ON)
	assert.Error(t, err)
}

func TestMaster_RegulatingStepCommand_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.RegulatingStepCommand(3001, RCS_HIGHER)
	assert.Error(t, err)
}

func TestMaster_SetPointCommandNormal_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.SetPointCommandNormal(1001, 0.5, false)
	assert.Error(t, err)
}

func TestMaster_SetPointCommandScaled_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.SetPointCommandScaled(1001, 1000, false)
	assert.Error(t, err)
}

func TestMaster_SetPointCommandFloat_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.SetPointCommandFloat(1001, 3.14, false)
	assert.Error(t, err)
}

func TestMaster_SendASDU_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	asdu := &ASDU{
		TypeID:      TYPE_ID_SINGLE_POINT_INFO,
		VSQ:         VSQ{Number: 1, IsSequence: false},
		COT:         CauseOfTransmission{Cause: COT_SPONTANEOUS},
		CommonAddr:  1,
		Information: []InformationObject{{Address: 1001, Value: &SinglePointInfo{Value: true}}},
	}
	err := master.SendASDU(asdu)
	assert.Error(t, err)
}

func TestIsSinglePointInfo_Negative(t *testing.T) {
	assert.False(t, IsSinglePointInfo(TYPE_ID_DOUBLE_POINT_INFO))
	assert.False(t, IsSinglePointInfo(TYPE_ID_MEASURE_VALUE_FLOAT))
}

func TestIsDoublePointInfo_Negative(t *testing.T) {
	assert.False(t, IsDoublePointInfo(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsMeasureValue_Negative(t *testing.T) {
	assert.False(t, IsMeasureValue(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsIntegratedTotal_Negative(t *testing.T) {
	assert.False(t, IsIntegratedTotal(TYPE_ID_SINGLE_POINT_INFO))
}

func TestHasCP24Time_Negative(t *testing.T) {
	assert.False(t, HasCP24Time(TYPE_ID_SINGLE_POINT_INFO))
}

func TestHasCP56Time_Negative(t *testing.T) {
	assert.False(t, HasCP56Time(TYPE_ID_SINGLE_POINT_INFO))
}

func TestHasTimestamp(t *testing.T) {
	assert.True(t, HasTimestamp(TYPE_ID_SINGLE_POINT_INFO_TIME))
	assert.True(t, HasTimestamp(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56))
	assert.False(t, HasTimestamp(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsControlDirection(t *testing.T) {
	assert.True(t, IsControlDirection(TYPE_ID_SINGLE_COMMAND))
	assert.True(t, IsControlDirection(TYPE_ID_DOUBLE_COMMAND))
	assert.True(t, IsControlDirection(TYPE_ID_INTERROGATION_CMD))
	assert.False(t, IsControlDirection(TYPE_ID_SINGLE_POINT_INFO))
}

func TestIsSystemInfo(t *testing.T) {
	assert.True(t, IsSystemInfo(TYPE_ID_END_OF_INITIALIZATION))
	assert.False(t, IsSystemInfo(TYPE_ID_SINGLE_POINT_INFO))
}

func TestGetInfoObjectValueString_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		obj      InformationObject
		contains string
	}{
		{"DoublePointIndeterminate", InformationObject{Value: &DoublePointInfo{Value: DP_INDETERMINATE}}, "INDETERMINATE"},
		{"StepPosition", InformationObject{Value: &StepPositionInfo{Value: 5}}, "STEP:5"},
		{"Bitstring32", InformationObject{Value: &Bitstring32{Value: 0xFF}}, "BITS:0x"},
		{"NormalizedValue", InformationObject{Value: &NormalizedValue{Value: 0.5}}, "0.5"},
		{"FloatValue", InformationObject{Value: &FloatValue{Value: 1.23}}, "1.23"},
		{"IntegratedTotal", InformationObject{Value: &IntegratedTotal{Value: 100}}, "TOTAL:100"},
		{"SingleCommand", InformationObject{Value: &SingleCommand{On: true}}, "CMD:true"},
		{"DoubleCommand", InformationObject{Value: &DoubleCommand{State: 2}}, "CMD:2"},
		{"RegulatingStepCommand", InformationObject{Value: &RegulatingStepCommand{State: 1}}, "STEP_CMD:1"},
		{"InterrogationCommand", InformationObject{Value: &InterrogationCommand{QOI: 20}}, "GI:20"},
		{"CounterInterrogationCommand", InformationObject{Value: &CounterInterrogationCommand{QCC: 1}}, "CI:1"},
		{"ClockSyncCommand", InformationObject{Value: &ClockSyncCommand{Time: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)}}, "TIME:"},
		{"Unknown", InformationObject{Value: "string_value"}, "string_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInfoObjectValueString(tt.obj)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestGetInfoObjectQualityString_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		obj      InformationObject
		expected string
	}{
		{"GoodQuality_SinglePoint", InformationObject{Value: &SinglePointInfo{Quality: Quality{}}}, "OK"},
		{"Invalid_DoublePoint", InformationObject{Value: &DoublePointInfo{Quality: Quality{Invalid: true}}}, "IV"},
		{"Overflow_Float", InformationObject{Value: &FloatValue{Quality: Quality{Overflow: true}}}, "OV"},
		{"NotCurrent_Normalized", InformationObject{Value: &NormalizedValue{Quality: Quality{NotCurrent: true}}}, "NT"},
		{"Substituted_Scaled", InformationObject{Value: &ScaledValue{Quality: Quality{Substituted: true}}}, "SB"},
		{"Blocked_IntegratedTotal", InformationObject{Value: &IntegratedTotal{Quality: Quality{Blocked: true}}}, "BL"},
		{"Unknown", InformationObject{Value: &SingleCommand{}}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInfoObjectQualityString(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPDU_Struct(t *testing.T) {
	apdu := &APDU{
		Type:     APDU_TYPE_I,
		SendSeq:  1,
		RecvSeq:  2,
		Control:  0,
		ASDU:     nil,
		RawBytes: []byte{0x68, 0x04},
	}
	assert.Equal(t, APDU_TYPE_I, apdu.Type)
	assert.Equal(t, uint16(1), apdu.SendSeq)
	assert.Equal(t, uint16(2), apdu.RecvSeq)
}

func TestCP24Time2a_Struct(t *testing.T) {
	cp24 := CP24Time2a{
		Milliseconds: 30000,
		Minutes:      30,
		Reserved:     false,
		SummerTime:   true,
	}
	assert.Equal(t, uint16(30000), cp24.Milliseconds)
	assert.Equal(t, uint8(30), cp24.Minutes)
	assert.True(t, cp24.SummerTime)
}

func TestConnection_processReceivedData_NoStartByte(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	conn.processReceivedData(data)
}

func TestConnection_processReceivedData_PartialFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	data := []byte{0x68, 0x04}
	conn.processReceivedData(data)
}

func TestConnection_processReceivedData_CompleteFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	data := []byte{0x68, 0x04, U_FRAME_TESTFR_ACT, 0x00, 0x00, 0x00}
	conn.processReceivedData(data)
}

func TestConnection_emitEvent(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.emitEvent(EVENT_CONNECTED, nil)
}

func TestConnection_emitError(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.emitError(assert.AnError)
}

func TestMaster_encodeASDU_UnsupportedType(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	asdu := &ASDU{
		TypeID:      TYPE_ID_SINGLE_POINT_INFO,
		VSQ:         VSQ{Number: 1, IsSequence: false},
		COT:         CauseOfTransmission{Cause: COT_SPONTANEOUS},
		CommonAddr:  1,
		Information: []InformationObject{{Address: 1001, Value: "invalid_type"}},
	}
	_, err := master.encodeASDU(asdu)
	assert.Error(t, err)
}

func TestMaster_encodeInfoObject_FloatValue(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address: 1001,
		Value:   &FloatValue{Value: 0},
	}
	data, err := master.encodeInfoObject(TYPE_ID_MEASURE_VALUE_FLOAT, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestMaster_encodeInfoObject_DoublePointInfo(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address: 1001,
		Value:   &DoublePointInfo{Value: DP_ON, Quality: Quality{}},
	}
	data, err := master.encodeInfoObject(TYPE_ID_DOUBLE_POINT_INFO, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestMaster_encodeInfoObject_WithTimestamp(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address:   1001,
		Value:     &SinglePointInfo{Value: true, Quality: Quality{}},
		Timestamp: time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local),
	}
	data, err := master.encodeInfoObject(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestMaster_encodeInfoObject_WithCP24Timestamp(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address:   1001,
		Value:     &SinglePointInfo{Value: true, Quality: Quality{}},
		Timestamp: time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local),
	}
	data, err := master.encodeInfoObject(TYPE_ID_SINGLE_POINT_INFO_TIME, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestDecodeStepPositionInfo_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_STEP_POSITION_INFO, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	infoData := append(infoAddr, 0x0A)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeBitstring32_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_BITSTRING32, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(7001)
	infoData := append(infoAddr, 0x01, 0x02, 0x03)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeNormalizedValue_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_MEASURE_VALUE_NORMAL, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(3001)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeScaledValue_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_MEASURE_VALUE_SCALED, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(4001)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeFloatValue_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_MEASURE_VALUE_FLOAT, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	infoData := append(infoAddr, 0x01, 0x02, 0x03)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeIntegratedTotal_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_INTEGRITY_TOTAL, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(6001)
	infoData := append(infoAddr, 0x01, 0x02, 0x03)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeSingleCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_SINGLE_COMMAND, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(1001)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeDoubleCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_DOUBLE_COMMAND, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(2001)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeRegulatingStepCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_REGULATING_STEP_COMMAND, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(3001)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeSetPointCommandNormal_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_SET_POINT_COMMAND_NORMAL, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(1001)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeSetPointCommandScaled_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_SET_POINT_COMMAND_SCALED, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(1001)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeSetPointCommandFloat_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_SET_POINT_COMMAND_FLOAT, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(1001)
	infoData := append(infoAddr, 0x01, 0x02, 0x03)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeInterrogationCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_INTERROGATION_CMD, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(0)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeCounterInterrogationCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_COUNTER_INTERROGATION_CMD, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(0)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeClockSyncCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_CLOCK_SYNC_CMD, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(0)
	infoData := append(infoAddr, 0x01, 0x02, 0x03)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeTestCommand_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_ACTIVATION}
	header := coder.EncodeASDUHeader(TYPE_ID_TEST_COMMAND, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(0)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeEndOfInitialization_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_INITIALIZED}
	header := coder.EncodeASDUHeader(TYPE_ID_END_OF_INITIALIZATION, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(0)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestMaster_handleAPDU_NilASDU(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	apdu := &APDU{ASDU: nil}
	master.handleAPDU(apdu)
}

func TestMaster_handleControlResponse(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	ch := make(chan *ControlResult, 1)
	master.controlWaiters[1] = ch
	asdu := &ASDU{
		TypeID: TYPE_ID_SINGLE_COMMAND,
		COT:    CauseOfTransmission{Cause: COT_ACTIVATION_CON},
	}
	master.handleControlResponse(asdu)
	select {
	case result := <-ch:
		assert.True(t, result.Success)
	default:
	}
}

func TestMaster_handleInterrogationResponse(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	ch := make(chan *ASDU, 1)
	master.interrogationWaiters[QOI_STATION_INTERROGATION] = ch
	asdu := &ASDU{
		TypeID: TYPE_ID_INTERROGATION_CMD,
		COT:    CauseOfTransmission{Cause: COT_ACTIVATION_TERMINATION},
		Information: []InformationObject{
			{Value: &InterrogationCommand{QOI: QOI_STATION_INTERROGATION}},
		},
	}
	master.handleInterrogationResponse(asdu)
}

func TestMaster_emitError(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	master.emitError(assert.AnError)
}

func TestMaster_WaitForConnection_NotConnected(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	err := master.WaitForConnection(100 * time.Millisecond)
	assert.Error(t, err)
}

func TestConnection_SendAck(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.SendAck()
}

func TestConnection_SendSeq(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	assert.Equal(t, uint16(0), conn.GetSendSeq())
}

func TestEncodeCP56Time2a_Sunday(t *testing.T) {
	sunday := time.Date(2015, 6, 14, 10, 30, 0, 0, time.UTC)
	data := EncodeCP56Time2a(sunday)
	assert.Equal(t, byte(7), data[4])
}

func TestEncodeCP56Time2a_Year(t *testing.T) {
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.UTC)
	data := EncodeCP56Time2a(ts)
	year := data[6] >> 4
	assert.Equal(t, uint8(15), year)
}

func TestASDUCoder_EncodeSinglePointInfo_WithCP24Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeSinglePointInfo(objects, true, 3)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestASDUCoder_EncodeDoublePointInfo_WithCP24Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 2001, Value: &DoublePointInfo{Value: DP_ON, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeDoublePointInfo(objects, true, 3)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestASDUCoder_EncodeNormalizedValue_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 3001, Value: &NormalizedValue{Value: 0.5, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeNormalizedValue(objects, true, 7)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestASDUCoder_EncodeScaledValue_WithCP24Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 4001, Value: &ScaledValue{Value: 500, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeScaledValue(objects, true, 3)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestASDUCoder_EncodeFloatValue_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 5001, Value: &FloatValue{Value: 3.14, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeFloatValue(objects, true, 7)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestASDUCoder_EncodeIntegratedTotal_WithCP24Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 6001, Value: &IntegratedTotal{Value: 12345, Sequence: 1, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeIntegratedTotal(objects, true, 3)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestConnection_processFrame_IFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x0E, 0x02, 0x00, 0x00, 0x00, 0x01, 0x01, 0x06, 0x00, 0x01, 0x00, 0x01, 0x01, 0x00, 0x00}
	conn.processFrame(frame)
}

func TestConnection_processFrame_SFrame(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x04, 0x01, 0x00, 0x00, 0x00}
	conn.processFrame(frame)
}

func TestConnection_processFrame_UFrame_STARTDT_CON(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.state = int32(STATE_STARTDT_SENT)
	frame := []byte{0x68, 0x04, U_FRAME_STARTDT_CON, 0x00, 0x00, 0x00}
	conn.processFrame(frame)
	assert.Equal(t, STATE_ACTIVE, conn.GetState())
}

func TestConnection_processFrame_UFrame_STOPDT_CON(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.state = int32(STATE_CONNECTED)
	frame := []byte{0x68, 0x04, U_FRAME_STOPDT_CON, 0x00, 0x00, 0x00}
	conn.processFrame(frame)
}

func TestConnection_processFrame_UFrame_TESTFR_ACT(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.state = int32(STATE_ACTIVE)
	frame := []byte{0x68, 0x04, U_FRAME_TESTFR_ACT, 0x00, 0x00, 0x00}
	conn.processFrame(frame)
}

func TestConnection_processFrame_UFrame_TESTFR_CON(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	conn.state = int32(STATE_ACTIVE)
	conn.testfrSent = true
	frame := []byte{0x68, 0x04, U_FRAME_TESTFR_CON, 0x00, 0x00, 0x00}
	conn.processFrame(frame)
	assert.False(t, conn.testfrSent)
}

func TestConnection_sendRaw(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	frame := []byte{0x68, 0x04, U_FRAME_TESTFR_ACT, 0x00, 0x00, 0x00}
	err := conn.sendRaw(frame)
	assert.NoError(t, err)
}

func TestConnection_sendIFrame_NoConn(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	err := conn.sendIFrame([]byte{0x01, 0x02})
	assert.NoError(t, err)
}

func TestConnection_writeDirect_NoConn(t *testing.T) {
	conn := NewConnection(DefaultConnectionConfig())
	err := conn.writeDirect([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestMasterState_String(t *testing.T) {
	assert.Equal(t, MasterState(0), MASTER_STATE_STOPPED)
	assert.Equal(t, MasterState(1), MASTER_STATE_STARTING)
	assert.Equal(t, MasterState(2), MASTER_STATE_RUNNING)
	assert.Equal(t, MasterState(3), MASTER_STATE_STOPPING)
}

func TestDecodeStepPositionInfo_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_STEP_POSITION_INFO_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	cp24 := EncodeCP24Time2a(ts)
	infoData := append(infoAddr, 0x0A, 0x00)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_STEP_POSITION_INFO_TIME), asdu.TypeID)

	spi, ok := asdu.Information[0].Value.(*StepPositionInfo)
	require.True(t, ok)
	assert.Equal(t, int16(10), spi.Value)
}

func TestDecodeStepPositionInfo_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_STEP_POSITION_INFO_TIME_CP56, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	cp56 := EncodeCP56Time2a(ts)
	infoData := append(infoAddr, 0x8A, 0x00)
	infoData = append(infoData, cp56...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_STEP_POSITION_INFO_TIME_CP56), asdu.TypeID)

	spi, ok := asdu.Information[0].Value.(*StepPositionInfo)
	require.True(t, ok)
	assert.Equal(t, int16(10), spi.Value)
	assert.True(t, spi.Transient)
}

func TestDecodeBitstring32_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_BITSTRING32_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(7001)
	cp24 := EncodeCP24Time2a(ts)
	infoData := append(infoAddr, 0x01, 0x02, 0x03, 0x04, 0x00)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_BITSTRING32_TIME), asdu.TypeID)
}

func TestDecodeBitstring32_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_BITSTRING32_TIME_CP56, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(7001)
	cp56 := EncodeCP56Time2a(ts)
	infoData := append(infoAddr, 0x01, 0x02, 0x03, 0x04, 0x00)
	infoData = append(infoData, cp56...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_BITSTRING32_TIME_CP56), asdu.TypeID)
}

func TestDecodeNormalizedValue_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_MEASURE_VALUE_NORMAL_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(3001)
	cp24 := EncodeCP24Time2a(ts)
	intVal := int16(16383)
	infoData := append(infoAddr, byte(intVal), byte(intVal>>8), 0x00)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_NORMAL_TIME), asdu.TypeID)
}

func TestDecodeFloatValue_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_MEASURE_VALUE_FLOAT_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(5001)
	cp24 := EncodeCP24Time2a(ts)
	bits := math.Float32bits(123.456)
	infoData := append(infoAddr, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24), 0x00)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_FLOAT_TIME), asdu.TypeID)
}

func TestDecodeScaledValue_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_MEASURE_VALUE_SCALED_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(4001)
	cp24 := EncodeCP24Time2a(ts)
	infoData := append(infoAddr, 0xE8, 0x03, 0x00)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_SCALED_TIME), asdu.TypeID)
}

func TestDecodeIntegratedTotal_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_INTEGRITY_TOTAL_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(6001)
	cp24 := EncodeCP24Time2a(ts)
	infoData := append(infoAddr, 0x39, 0x30, 0x00, 0x00, 0x01, 0x00)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_INTEGRITY_TOTAL_TIME), asdu.TypeID)
}

func TestDecodeDoublePointInfo_WithTime(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)

	header := coder.EncodeASDUHeader(TYPE_ID_DOUBLE_POINT_INFO_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(2001)
	cp24 := EncodeCP24Time2a(ts)
	infoData := append(infoAddr, 0x01)
	infoData = append(infoData, cp24...)
	data := append(header, infoData...)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_DOUBLE_POINT_INFO_TIME), asdu.TypeID)
}

func TestDecodeSinglePointInfo_InsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_SINGLE_POINT_INFO, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(1001)
	data := append(header, infoAddr...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestDecodeSinglePointInfo_TimeInsufficientData(t *testing.T) {
	coder := NewASDUCoder()
	vsq := VSQ{Number: 1, IsSequence: false}
	cot := CauseOfTransmission{Cause: COT_SPONTANEOUS}
	header := coder.EncodeASDUHeader(TYPE_ID_SINGLE_POINT_INFO_TIME, vsq, cot, 1)
	infoAddr := coder.EncodeInfoAddress(1001)
	infoData := append(infoAddr, 0x01)
	data := append(header, infoData...)

	_, err := coder.DecodeASDU(data)
	assert.Error(t, err)
}

func TestConnection_WithLocalServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			if n >= 6 && buf[0] == 0x68 {
				frameLen := int(buf[1])
				if buf[2] == U_FRAME_STARTDT_ACT {
					resp := []byte{0x68, 0x04, U_FRAME_STARTDT_CON, 0x00, 0x00, 0x00}
					conn.Write(resp)
				} else if buf[2] == U_FRAME_TESTFR_ACT {
					resp := []byte{0x68, 0x04, U_FRAME_TESTFR_CON, 0x00, 0x00, 0x00}
					conn.Write(resp)
				} else if buf[2] == U_FRAME_STOPDT_ACT {
					resp := []byte{0x68, 0x04, U_FRAME_STOPDT_CON, 0x00, 0x00, 0x00}
					conn.Write(resp)
				} else if frameLen > 4 {
					_ = frameLen
				}
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = addr.Port
	config.Timeout = 5 * time.Second
	config.HeartbeatInterval = 1 * time.Hour
	config.HeartbeatTimeout = 1 * time.Hour

	conn := NewConnection(config)
	err = conn.Connect()
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, conn.IsActive())

	stats := conn.GetStats()
	assert.Equal(t, uint64(1), stats.ConnectCount)
	assert.Greater(t, stats.BytesSent, uint64(0))

	err = conn.Disconnect()
	require.NoError(t, err)
	assert.Equal(t, STATE_DISCONNECTED, conn.GetState())
}

func TestConnection_BalancedMode(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			_ = n
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = addr.Port
	config.Timeout = 5 * time.Second
	config.BalancedMode = true
	config.HeartbeatInterval = 1 * time.Hour
	config.HeartbeatTimeout = 1 * time.Hour

	conn := NewConnection(config)
	err = conn.Connect()
	require.NoError(t, err)
	assert.Equal(t, STATE_ACTIVE, conn.GetState())

	err = conn.Disconnect()
	require.NoError(t, err)
}

func TestConnection_Reconnect(t *testing.T) {
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = 19999
	config.Timeout = 1 * time.Second
	config.MaxReconnectAttempts = 1
	config.ReconnectInterval = 100 * time.Millisecond

	conn := NewConnection(config)
	conn.StopReconnect()

	err := conn.Connect()
	assert.Error(t, err)
}

func TestConnectionPool_ConnectAll_Failed(t *testing.T) {
	pool := NewConnectionPool()
	config := DefaultConnectionConfig()
	config.Host = "127.0.0.1"
	config.Port = 19999
	config.Timeout = 1 * time.Second
	pool.AddConfig("fail_conn", config)

	err := pool.ConnectAll()
	assert.Error(t, err)
}

func TestMaster_WithLocalServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			if n >= 6 && buf[0] == 0x68 && buf[2] == U_FRAME_STARTDT_ACT {
				resp := []byte{0x68, 0x04, U_FRAME_STARTDT_CON, 0x00, 0x00, 0x00}
				conn.Write(resp)
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	config := DefaultMasterConfig()
	config.Host = "127.0.0.1"
	config.Port = addr.Port
	config.Timeout = 5 * time.Second
	config.HeartbeatInterval = 1 * time.Hour
	config.HeartbeatTimeout = 1 * time.Hour

	master := NewMaster(config)
	err = master.Start()
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, master.IsRunning())

	stats := master.GetStats()
	assert.Equal(t, uint64(0), stats.TotalASDUReceived)

	err = master.Stop()
	require.NoError(t, err)
}

func TestMaster_Start_AlreadyRunning(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	master.state = int32(MASTER_STATE_RUNNING)
	err := master.Start()
	assert.Error(t, err)
	master.state = int32(MASTER_STATE_STOPPED)
}

func TestDecodeSinglePointInfo_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 1001, Value: &SinglePointInfo{Value: true, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeSinglePointInfo(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_SINGLE_POINT_INFO_TIME_CP56), asdu.TypeID)
	assert.False(t, asdu.Information[0].Timestamp.IsZero())
}

func TestDecodeDoublePointInfo_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 2001, Value: &DoublePointInfo{Value: DP_ON, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeDoublePointInfo(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56), asdu.TypeID)
	assert.False(t, asdu.Information[0].Timestamp.IsZero())
}

func TestDecodeNormalizedValue_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 3001, Value: &NormalizedValue{Value: 0.75, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeNormalizedValue(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_NORMAL_TIME_CP56), asdu.TypeID)
	assert.False(t, asdu.Information[0].Timestamp.IsZero())
}

func TestDecodeScaledValue_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 4001, Value: &ScaledValue{Value: 500, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeScaledValue(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56), asdu.TypeID)
	assert.False(t, asdu.Information[0].Timestamp.IsZero())
}

func TestDecodeFloatValue_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 5001, Value: &FloatValue{Value: 3.14, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeFloatValue(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_MEASURE_VALUE_FLOAT_TIME_CP56), asdu.TypeID)
	assert.False(t, asdu.Information[0].Timestamp.IsZero())
}

func TestDecodeIntegratedTotal_WithCP56Time(t *testing.T) {
	coder := NewASDUCoder()
	ts := time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local)
	objects := []InformationObject{
		{Address: 6001, Value: &IntegratedTotal{Value: 12345, Sequence: 1, Quality: Quality{}}, Timestamp: ts},
	}
	data, err := coder.EncodeIntegratedTotal(objects, true, 7)
	require.NoError(t, err)

	asdu, err := coder.DecodeASDU(data)
	require.NoError(t, err)
	assert.Equal(t, uint8(TYPE_ID_INTEGRITY_TOTAL_TIME_CP56), asdu.TypeID)
	assert.False(t, asdu.Information[0].Timestamp.IsZero())
}

func TestMaster_encodeInfoObject_NormalizedValue(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address: 3001,
		Value:   &NormalizedValue{Value: 0.5, Quality: Quality{}},
	}
	data, err := master.encodeInfoObject(TYPE_ID_MEASURE_VALUE_NORMAL, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestMaster_encodeInfoObject_ScaledValue(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address: 4001,
		Value:   &ScaledValue{Value: 500, Quality: Quality{}},
	}
	data, err := master.encodeInfoObject(TYPE_ID_MEASURE_VALUE_SCALED, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestMaster_encodeInfoObject_FloatValueNonZero(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address: 5001,
		Value:   &FloatValue{Value: 3.14, Quality: Quality{}},
	}
	data, err := master.encodeInfoObject(TYPE_ID_MEASURE_VALUE_FLOAT, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestMaster_encodeInfoObject_SinglePointWithTimestamp(t *testing.T) {
	config := DefaultMasterConfig()
	master := NewMaster(config)
	obj := InformationObject{
		Address:   1001,
		Value:     &SinglePointInfo{Value: true, Quality: Quality{}},
		Timestamp: time.Date(2015, 6, 15, 10, 30, 0, 0, time.Local),
	}
	data, err := master.encodeInfoObject(TYPE_ID_SINGLE_POINT_INFO_TIME, obj)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}
