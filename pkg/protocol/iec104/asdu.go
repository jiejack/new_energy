package iec104

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

// ASDU编解码器
type ASDUCoder struct{}

// 新建ASDU编解码器
func NewASDUCoder() *ASDUCoder {
	return &ASDUCoder{}
}

// ==================== ASDU编码 ====================

// 编码ASDU头部
func (c *ASDUCoder) EncodeASDUHeader(typeID uint8, vsq VSQ, cot CauseOfTransmission, commonAddr uint16) []byte {
	data := make([]byte, 6)
	data[0] = typeID
	data[1] = EncodeVSQ(vsq)
	data[2], data[3] = EncodeCOT(cot)
	data[4] = byte(commonAddr)
	data[5] = byte(commonAddr >> 8)
	return data
}

// 编码信息体地址 (3字节)
func (c *ASDUCoder) EncodeInfoAddress(addr uint32) []byte {
	return []byte{
		byte(addr),
		byte(addr >> 8),
		byte(addr >> 16),
	}
}

// 编码单点遥信
func (c *ASDUCoder) EncodeSinglePointInfo(objects []InformationObject, withTime bool, timeSize int) ([]byte, error) {
	if len(objects) == 0 {
		return nil, errors.New("no objects to encode")
	}

	typeID := TYPE_ID_SINGLE_POINT_INFO
	if withTime {
		if timeSize == 7 {
			typeID = TYPE_ID_SINGLE_POINT_INFO_TIME_CP56
		} else {
			typeID = TYPE_ID_SINGLE_POINT_INFO_TIME
		}
	}

	vsq := VSQ{
		Number:     uint8(len(objects)),
		IsSequence: false,
	}

	data := c.EncodeASDUHeader(typeID, vsq, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0)

	for _, obj := range objects {
		data = append(data, c.EncodeInfoAddress(obj.Address)...)

		spi, ok := obj.Value.(*SinglePointInfo)
		if !ok {
			return nil, fmt.Errorf("invalid value type for single point info")
		}

		var siq byte
		if spi.Value {
			siq = 0x01
		}
		siq |= EncodeQuality(spi.Quality)
		data = append(data, siq)

		if withTime {
			if timeSize == 7 {
				data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
			} else {
				data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
			}
		}
	}

	return data, nil
}

// 编码双点遥信
func (c *ASDUCoder) EncodeDoublePointInfo(objects []InformationObject, withTime bool, timeSize int) ([]byte, error) {
	if len(objects) == 0 {
		return nil, errors.New("no objects to encode")
	}

	typeID := TYPE_ID_DOUBLE_POINT_INFO
	if withTime {
		if timeSize == 7 {
			typeID = TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56
		} else {
			typeID = TYPE_ID_DOUBLE_POINT_INFO_TIME
		}
	}

	vsq := VSQ{
		Number:     uint8(len(objects)),
		IsSequence: false,
	}

	data := c.EncodeASDUHeader(typeID, vsq, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0)

	for _, obj := range objects {
		data = append(data, c.EncodeInfoAddress(obj.Address)...)

		dpi, ok := obj.Value.(*DoublePointInfo)
		if !ok {
			return nil, fmt.Errorf("invalid value type for double point info")
		}

		var diq byte = dpi.Value & 0x03
		diq |= EncodeQuality(dpi.Quality)
		data = append(data, diq)

		if withTime {
			if timeSize == 7 {
				data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
			} else {
				data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
			}
		}
	}

	return data, nil
}

// 编码归一化遥测值
func (c *ASDUCoder) EncodeNormalizedValue(objects []InformationObject, withTime bool, timeSize int) ([]byte, error) {
	if len(objects) == 0 {
		return nil, errors.New("no objects to encode")
	}

	typeID := TYPE_ID_MEASURE_VALUE_NORMAL
	if withTime {
		if timeSize == 7 {
			typeID = TYPE_ID_MEASURE_VALUE_NORMAL_TIME_CP56
		} else {
			typeID = TYPE_ID_MEASURE_VALUE_NORMAL_TIME
		}
	}

	vsq := VSQ{
		Number:     uint8(len(objects)),
		IsSequence: false,
	}

	data := c.EncodeASDUHeader(typeID, vsq, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0)

	for _, obj := range objects {
		data = append(data, c.EncodeInfoAddress(obj.Address)...)

		nv, ok := obj.Value.(*NormalizedValue)
		if !ok {
			return nil, fmt.Errorf("invalid value type for normalized value")
		}

		// 归一化值: -1.0 ~ 1.0 映射到 -32768 ~ 32767
		intVal := int16(nv.Value * 32767)
		data = append(data, byte(intVal), byte(intVal>>8))

		qds := EncodeMeasureQuality(nv.Quality)
		data = append(data, qds)

		if withTime {
			if timeSize == 7 {
				data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
			} else {
				data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
			}
		}
	}

	return data, nil
}

// 编码标度化遥测值
func (c *ASDUCoder) EncodeScaledValue(objects []InformationObject, withTime bool, timeSize int) ([]byte, error) {
	if len(objects) == 0 {
		return nil, errors.New("no objects to encode")
	}

	typeID := TYPE_ID_MEASURE_VALUE_SCALED
	if withTime {
		if timeSize == 7 {
			typeID = TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56
		} else {
			typeID = TYPE_ID_MEASURE_VALUE_SCALED_TIME
		}
	}

	vsq := VSQ{
		Number:     uint8(len(objects)),
		IsSequence: false,
	}

	data := c.EncodeASDUHeader(typeID, vsq, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0)

	for _, obj := range objects {
		data = append(data, c.EncodeInfoAddress(obj.Address)...)

		sv, ok := obj.Value.(*ScaledValue)
		if !ok {
			return nil, fmt.Errorf("invalid value type for scaled value")
		}

		data = append(data, byte(sv.Value), byte(sv.Value>>8))

		qds := EncodeMeasureQuality(sv.Quality)
		data = append(data, qds)

		if withTime {
			if timeSize == 7 {
				data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
			} else {
				data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
			}
		}
	}

	return data, nil
}

// 编码短浮点遥测值
func (c *ASDUCoder) EncodeFloatValue(objects []InformationObject, withTime bool, timeSize int) ([]byte, error) {
	if len(objects) == 0 {
		return nil, errors.New("no objects to encode")
	}

	typeID := TYPE_ID_MEASURE_VALUE_FLOAT
	if withTime {
		if timeSize == 7 {
			typeID = TYPE_ID_MEASURE_VALUE_FLOAT_TIME_CP56
		} else {
			typeID = TYPE_ID_MEASURE_VALUE_FLOAT_TIME
		}
	}

	vsq := VSQ{
		Number:     uint8(len(objects)),
		IsSequence: false,
	}

	data := c.EncodeASDUHeader(typeID, vsq, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0)

	for _, obj := range objects {
		data = append(data, c.EncodeInfoAddress(obj.Address)...)

		fv, ok := obj.Value.(*FloatValue)
		if !ok {
			return nil, fmt.Errorf("invalid value type for float value")
		}

		bits := math.Float32bits(fv.Value)
		data = append(data, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))

		qds := EncodeMeasureQuality(fv.Quality)
		data = append(data, qds)

		if withTime {
			if timeSize == 7 {
				data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
			} else {
				data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
			}
		}
	}

	return data, nil
}

// 编码电度值
func (c *ASDUCoder) EncodeIntegratedTotal(objects []InformationObject, withTime bool, timeSize int) ([]byte, error) {
	if len(objects) == 0 {
		return nil, errors.New("no objects to encode")
	}

	typeID := TYPE_ID_INTEGRITY_TOTAL
	if withTime {
		if timeSize == 7 {
			typeID = TYPE_ID_INTEGRITY_TOTAL_TIME_CP56
		} else {
			typeID = TYPE_ID_INTEGRITY_TOTAL_TIME
		}
	}

	vsq := VSQ{
		Number:     uint8(len(objects)),
		IsSequence: false,
	}

	data := c.EncodeASDUHeader(typeID, vsq, CauseOfTransmission{Cause: COT_SPONTANEOUS}, 0)

	for _, obj := range objects {
		data = append(data, c.EncodeInfoAddress(obj.Address)...)

		it, ok := obj.Value.(*IntegratedTotal)
		if !ok {
			return nil, fmt.Errorf("invalid value type for integrated total")
		}

		data = append(data, byte(it.Value), byte(it.Value>>8), byte(it.Value>>16), byte(it.Value>>24))
		data = append(data, it.Sequence&0x1F)
		qds := EncodeMeasureQuality(it.Quality)
		data = append(data, qds)

		if withTime {
			if timeSize == 7 {
				data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
			} else {
				data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
			}
		}
	}

	return data, nil
}

// 编码单命令
func (c *ASDUCoder) EncodeSingleCommand(commonAddr uint16, infoAddr uint32, cmd *SingleCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_SINGLE_COMMAND, vsq, cot, commonAddr)
	data = append(data, c.EncodeInfoAddress(infoAddr)...)

	var sco byte
	if cmd.On {
		sco = 0x01
	}
	if cmd.Select {
		sco |= SCS_SELECT
	}
	sco |= (cmd.QU & 0x1F) << 2
	data = append(data, sco)

	return data
}

// 编码双命令
func (c *ASDUCoder) EncodeDoubleCommand(commonAddr uint16, infoAddr uint32, cmd *DoubleCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_DOUBLE_COMMAND, vsq, cot, commonAddr)
	data = append(data, c.EncodeInfoAddress(infoAddr)...)

	var dco byte = cmd.State & 0x03
	if cmd.Select {
		dco |= DCS_SELECT
	}
	dco |= (cmd.QU & 0x1F) << 2
	data = append(data, dco)

	return data
}

// 编码升降命令
func (c *ASDUCoder) EncodeRegulatingStepCommand(commonAddr uint16, infoAddr uint32, cmd *RegulatingStepCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_REGULATING_STEP_COMMAND, vsq, cot, commonAddr)
	data = append(data, c.EncodeInfoAddress(infoAddr)...)

	var rco byte = cmd.State & 0x03
	if cmd.Select {
		rco |= RCS_SELECT
	}
	rco |= (cmd.QU & 0x1F) << 2
	data = append(data, rco)

	return data
}

// 编码设点命令-归一化值
func (c *ASDUCoder) EncodeSetPointCommandNormal(commonAddr uint16, infoAddr uint32, cmd *SetPointCommand, cot CauseOfTransmission) ([]byte, error) {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_SET_POINT_COMMAND_NORMAL, vsq, cot, commonAddr)
	data = append(data, c.EncodeInfoAddress(infoAddr)...)

	nv, ok := cmd.Value.(*NormalizedValue)
	if !ok {
		return nil, fmt.Errorf("invalid value type for set point command normal")
	}

	intVal := int16(nv.Value * 32767)
	data = append(data, byte(intVal), byte(intVal>>8))

	var qos byte
	if cmd.Select {
		qos = 0x80
	}
	data = append(data, qos)

	return data, nil
}

// 编码设点命令-标度化值
func (c *ASDUCoder) EncodeSetPointCommandScaled(commonAddr uint16, infoAddr uint32, cmd *SetPointCommand, cot CauseOfTransmission) ([]byte, error) {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_SET_POINT_COMMAND_SCALED, vsq, cot, commonAddr)
	data = append(data, c.EncodeInfoAddress(infoAddr)...)

	sv, ok := cmd.Value.(*ScaledValue)
	if !ok {
		return nil, fmt.Errorf("invalid value type for set point command scaled")
	}

	data = append(data, byte(sv.Value), byte(sv.Value>>8))

	var qos byte
	if cmd.Select {
		qos = 0x80
	}
	data = append(data, qos)

	return data, nil
}

// 编码设点命令-短浮点数
func (c *ASDUCoder) EncodeSetPointCommandFloat(commonAddr uint16, infoAddr uint32, cmd *SetPointCommand, cot CauseOfTransmission) ([]byte, error) {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_SET_POINT_COMMAND_FLOAT, vsq, cot, commonAddr)
	data = append(data, c.EncodeInfoAddress(infoAddr)...)

	fv, ok := cmd.Value.(*FloatValue)
	if !ok {
		return nil, fmt.Errorf("invalid value type for set point command float")
	}

	bits := math.Float32bits(fv.Value)
	data = append(data, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))

	var qos byte
	if cmd.Select {
		qos = 0x80
	}
	data = append(data, qos)

	return data, nil
}

// 编码总召唤命令
func (c *ASDUCoder) EncodeInterrogationCommand(commonAddr uint16, cmd *InterrogationCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_INTERROGATION_CMD, vsq, cot, commonAddr)
	data = append(data, 0x00, 0x00, 0x00) // 信息体地址为0
	data = append(data, cmd.QOI)

	return data
}

// 编码电度召唤命令
func (c *ASDUCoder) EncodeCounterInterrogationCommand(commonAddr uint16, cmd *CounterInterrogationCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_COUNTER_INTERROGATION_CMD, vsq, cot, commonAddr)
	data = append(data, 0x00, 0x00, 0x00) // 信息体地址为0
	data = append(data, cmd.QCC)

	return data
}

// 编码时钟同步命令
func (c *ASDUCoder) EncodeClockSyncCommand(commonAddr uint16, cmd *ClockSyncCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_CLOCK_SYNC_CMD, vsq, cot, commonAddr)
	data = append(data, 0x00, 0x00, 0x00) // 信息体地址为0
	data = append(data, EncodeCP56Time2a(cmd.Time)...)

	return data
}

// 编码测试命令
func (c *ASDUCoder) EncodeTestCommand(commonAddr uint16, cmd *TestCommand, cot CauseOfTransmission) []byte {
	vsq := VSQ{Number: 1, IsSequence: false}

	data := c.EncodeASDUHeader(TYPE_ID_TEST_COMMAND, vsq, cot, commonAddr)
	data = append(data, 0x00, 0x00, 0x00) // 信息体地址为0
	data = append(data, cmd.FFS1, cmd.FFS2)

	return data
}

// ==================== ASDU解码 ====================

// 解析ASDU
func (c *ASDUCoder) DecodeASDU(data []byte) (*ASDU, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("insufficient data for ASDU header: need 6, got %d", len(data))
	}

	asdu := &ASDU{
		TypeID:     data[0],
		VSQ:        ParseVSQ(data[1]),
		COT:        ParseCOT(data[2], data[3]),
		Origin:     data[3],
		CommonAddr: uint16(data[4]) | uint16(data[5])<<8,
	}

	// 解析信息体
	infoData := data[6:]
	objects, err := c.DecodeInformationObjects(asdu.TypeID, asdu.VSQ, infoData)
	if err != nil {
		return nil, err
	}

	asdu.Information = objects
	return asdu, nil
}

// 解析信息体
func (c *ASDUCoder) DecodeInformationObjects(typeID uint8, vsq VSQ, data []byte) ([]InformationObject, error) {
	if vsq.Number == 0 {
		return nil, nil
	}

	objects := make([]InformationObject, 0, vsq.Number)

	if vsq.IsSequence {
		// 序列模式: 第一个信息体有地址，后续信息体地址递增
		if len(data) < 3 {
			return nil, errors.New("insufficient data for first info address")
		}

		baseAddr := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
		data = data[3:]

		for i := uint8(0); i < vsq.Number; i++ {
			obj, remaining, err := c.DecodeSingleInformationObject(typeID, baseAddr+uint32(i), data, false)
			if err != nil {
				return nil, err
			}
			objects = append(objects, *obj)
			data = remaining
		}
	} else {
		// 非序列模式: 每个信息体都有地址
		for i := uint8(0); i < vsq.Number; i++ {
			if len(data) < 3 {
				return nil, fmt.Errorf("insufficient data for info object %d address", i)
			}

			addr := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
			data = data[3:]

			obj, remaining, err := c.DecodeSingleInformationObject(typeID, addr, data, false)
			if err != nil {
				return nil, err
			}
			objects = append(objects, *obj)
			data = remaining
		}
	}

	return objects, nil
}

// 解析单个信息体
func (c *ASDUCoder) DecodeSingleInformationObject(typeID uint8, addr uint32, data []byte, hasAddr bool) (*InformationObject, []byte, error) {
	obj := &InformationObject{
		Address: addr,
	}

	var err error
	var consumed int

	switch typeID {
	case TYPE_ID_SINGLE_POINT_INFO:
		consumed, err = c.decodeSinglePointInfo(obj, data, false, 0)
	case TYPE_ID_SINGLE_POINT_INFO_TIME:
		consumed, err = c.decodeSinglePointInfo(obj, data, true, 3)
	case TYPE_ID_SINGLE_POINT_INFO_TIME_CP56:
		consumed, err = c.decodeSinglePointInfo(obj, data, true, 7)
	case TYPE_ID_DOUBLE_POINT_INFO:
		consumed, err = c.decodeDoublePointInfo(obj, data, false, 0)
	case TYPE_ID_DOUBLE_POINT_INFO_TIME:
		consumed, err = c.decodeDoublePointInfo(obj, data, true, 3)
	case TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56:
		consumed, err = c.decodeDoublePointInfo(obj, data, true, 7)
	case TYPE_ID_STEP_POSITION_INFO:
		consumed, err = c.decodeStepPositionInfo(obj, data, false, 0)
	case TYPE_ID_STEP_POSITION_INFO_TIME:
		consumed, err = c.decodeStepPositionInfo(obj, data, true, 3)
	case TYPE_ID_STEP_POSITION_INFO_TIME_CP56:
		consumed, err = c.decodeStepPositionInfo(obj, data, true, 7)
	case TYPE_ID_BITSTRING32:
		consumed, err = c.decodeBitstring32(obj, data, false, 0)
	case TYPE_ID_BITSTRING32_TIME:
		consumed, err = c.decodeBitstring32(obj, data, true, 3)
	case TYPE_ID_BITSTRING32_TIME_CP56:
		consumed, err = c.decodeBitstring32(obj, data, true, 7)
	case TYPE_ID_MEASURE_VALUE_NORMAL:
		consumed, err = c.decodeNormalizedValue(obj, data, false, 0)
	case TYPE_ID_MEASURE_VALUE_NORMAL_TIME:
		consumed, err = c.decodeNormalizedValue(obj, data, true, 3)
	case TYPE_ID_MEASURE_VALUE_NORMAL_TIME_CP56:
		consumed, err = c.decodeNormalizedValue(obj, data, true, 7)
	case TYPE_ID_MEASURE_VALUE_SCALED:
		consumed, err = c.decodeScaledValue(obj, data, false, 0)
	case TYPE_ID_MEASURE_VALUE_SCALED_TIME:
		consumed, err = c.decodeScaledValue(obj, data, true, 3)
	case TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56:
		consumed, err = c.decodeScaledValue(obj, data, true, 7)
	case TYPE_ID_MEASURE_VALUE_FLOAT:
		consumed, err = c.decodeFloatValue(obj, data, false, 0)
	case TYPE_ID_MEASURE_VALUE_FLOAT_TIME:
		consumed, err = c.decodeFloatValue(obj, data, true, 3)
	case TYPE_ID_MEASURE_VALUE_FLOAT_TIME_CP56:
		consumed, err = c.decodeFloatValue(obj, data, true, 7)
	case TYPE_ID_INTEGRITY_TOTAL:
		consumed, err = c.decodeIntegratedTotal(obj, data, false, 0)
	case TYPE_ID_INTEGRITY_TOTAL_TIME:
		consumed, err = c.decodeIntegratedTotal(obj, data, true, 3)
	case TYPE_ID_INTEGRITY_TOTAL_TIME_CP56:
		consumed, err = c.decodeIntegratedTotal(obj, data, true, 7)
	case TYPE_ID_SINGLE_COMMAND:
		consumed, err = c.decodeSingleCommand(obj, data)
	case TYPE_ID_DOUBLE_COMMAND:
		consumed, err = c.decodeDoubleCommand(obj, data)
	case TYPE_ID_REGULATING_STEP_COMMAND:
		consumed, err = c.decodeRegulatingStepCommand(obj, data)
	case TYPE_ID_SET_POINT_COMMAND_NORMAL:
		consumed, err = c.decodeSetPointCommandNormal(obj, data)
	case TYPE_ID_SET_POINT_COMMAND_SCALED:
		consumed, err = c.decodeSetPointCommandScaled(obj, data)
	case TYPE_ID_SET_POINT_COMMAND_FLOAT:
		consumed, err = c.decodeSetPointCommandFloat(obj, data)
	case TYPE_ID_INTERROGATION_CMD:
		consumed, err = c.decodeInterrogationCommand(obj, data)
	case TYPE_ID_COUNTER_INTERROGATION_CMD:
		consumed, err = c.decodeCounterInterrogationCommand(obj, data)
	case TYPE_ID_CLOCK_SYNC_CMD:
		consumed, err = c.decodeClockSyncCommand(obj, data)
	case TYPE_ID_TEST_COMMAND:
		consumed, err = c.decodeTestCommand(obj, data)
	case TYPE_ID_END_OF_INITIALIZATION:
		consumed, err = c.decodeEndOfInitialization(obj, data)
	default:
		return nil, data, fmt.Errorf("unsupported type ID: %d", typeID)
	}

	if err != nil {
		return nil, data, err
	}

	return obj, data[consumed:], nil
}

// 解析单点遥信
func (c *ASDUCoder) decodeSinglePointInfo(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for single point info")
	}

	siq := data[0]
	spi := &SinglePointInfo{
		Value:   (siq & 0x01) != 0,
		Quality: ParseQuality(siq),
	}
	obj.Value = spi
	obj.Quality = spi.Quality

	consumed := 1

	if hasTime {
		if len(data) < 1+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[1 : 1+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[1 : 1+3])
			// CP24只有时分秒，需要补充日期
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析双点遥信
func (c *ASDUCoder) decodeDoublePointInfo(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for double point info")
	}

	diq := data[0]
	dpi := &DoublePointInfo{
		Value:   diq & 0x03,
		Quality: ParseQuality(diq),
	}
	obj.Value = dpi
	obj.Quality = dpi.Quality

	consumed := 1

	if hasTime {
		if len(data) < 1+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[1 : 1+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[1 : 1+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析步位置信息
func (c *ASDUCoder) decodeStepPositionInfo(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("insufficient data for step position info")
	}

	vti := data[0]
	qds := data[1]

	spi := &StepPositionInfo{
		Value:     int16(int8(vti & 0x7F)),
		Transient: (vti & 0x80) != 0,
		Quality:   ParseQuality(qds),
	}
	obj.Value = spi
	obj.Quality = spi.Quality

	consumed := 2

	if hasTime {
		if len(data) < 2+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[2 : 2+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[2 : 2+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析32位串
func (c *ASDUCoder) decodeBitstring32(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 5 {
		return 0, errors.New("insufficient data for bitstring32")
	}

	bs := &Bitstring32{
		Value:   binary.LittleEndian.Uint32(data[0:4]),
		Quality: ParseQuality(data[4]),
	}
	obj.Value = bs
	obj.Quality = bs.Quality

	consumed := 5

	if hasTime {
		if len(data) < 5+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[5 : 5+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[5 : 5+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析归一化遥测值
func (c *ASDUCoder) decodeNormalizedValue(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 3 {
		return 0, errors.New("insufficient data for normalized value")
	}

	intVal := int16(binary.LittleEndian.Uint16(data[0:2]))
	nv := &NormalizedValue{
		Value:   float32(intVal) / 32767.0,
		Quality: ParseMeasureQuality(data[2]),
	}
	obj.Value = nv
	obj.Quality = nv.Quality

	consumed := 3

	if hasTime {
		if len(data) < 3+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[3 : 3+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[3 : 3+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析标度化遥测值
func (c *ASDUCoder) decodeScaledValue(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 3 {
		return 0, errors.New("insufficient data for scaled value")
	}

	sv := &ScaledValue{
		Value:   int16(binary.LittleEndian.Uint16(data[0:2])),
		Quality: ParseMeasureQuality(data[2]),
	}
	obj.Value = sv
	obj.Quality = sv.Quality

	consumed := 3

	if hasTime {
		if len(data) < 3+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[3 : 3+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[3 : 3+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析短浮点遥测值
func (c *ASDUCoder) decodeFloatValue(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 5 {
		return 0, errors.New("insufficient data for float value")
	}

	bits := binary.LittleEndian.Uint32(data[0:4])
	fv := &FloatValue{
		Value:   math.Float32frombits(bits),
		Quality: ParseMeasureQuality(data[4]),
	}
	obj.Value = fv
	obj.Quality = fv.Quality

	consumed := 5

	if hasTime {
		if len(data) < 5+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[5 : 5+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[5 : 5+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析电度值
func (c *ASDUCoder) decodeIntegratedTotal(obj *InformationObject, data []byte, hasTime bool, timeSize int) (int, error) {
	if len(data) < 6 {
		return 0, errors.New("insufficient data for integrated total")
	}

	it := &IntegratedTotal{
		Value:    binary.LittleEndian.Uint32(data[0:4]),
		Sequence: data[4] & 0x1F,
		Quality:  ParseMeasureQuality(data[5]),
	}
	obj.Value = it
	obj.Quality = it.Quality

	consumed := 6

	if hasTime {
		if len(data) < 6+timeSize {
			return 0, errors.New("insufficient data for timestamp")
		}

		if timeSize == 7 {
			cp56, err := DecodeCP56Time2a(data[6 : 6+7])
			if err != nil {
				return 0, err
			}
			obj.Timestamp = cp56.ToTime()
		} else {
			cp24 := DecodeCP24Time2a(data[6 : 6+3])
			now := time.Now()
			obj.Timestamp = time.Date(now.Year(), now.Month(), now.Day(),
				now.Hour(), int(cp24.Minutes), int(cp24.Milliseconds/1000),
				int((cp24.Milliseconds%1000)*1000000), time.Local)
		}
		consumed += timeSize
	}

	return consumed, nil
}

// 解析单命令
func (c *ASDUCoder) decodeSingleCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for single command")
	}

	sco := data[0]
	cmd := &SingleCommand{
		Select: (sco & SCS_SELECT) != 0,
		QU:     (sco >> 2) & 0x1F,
		On:     (sco & 0x01) != 0,
	}
	obj.Value = cmd

	return 1, nil
}

// 解析双命令
func (c *ASDUCoder) decodeDoubleCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for double command")
	}

	dco := data[0]
	cmd := &DoubleCommand{
		Select: (dco & DCS_SELECT) != 0,
		QU:     (dco >> 2) & 0x1F,
		State:  dco & 0x03,
	}
	obj.Value = cmd

	return 1, nil
}

// 解析升降命令
func (c *ASDUCoder) decodeRegulatingStepCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for regulating step command")
	}

	rco := data[0]
	cmd := &RegulatingStepCommand{
		Select: (rco & RCS_SELECT) != 0,
		QU:     (rco >> 2) & 0x1F,
		State:  rco & 0x03,
	}
	obj.Value = cmd

	return 1, nil
}

// 解析设点命令-归一化值
func (c *ASDUCoder) decodeSetPointCommandNormal(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 3 {
		return 0, errors.New("insufficient data for set point command normal")
	}

	intVal := int16(binary.LittleEndian.Uint16(data[0:2]))
	cmd := &SetPointCommand{
		Select: (data[2] & 0x80) != 0,
		Value: &NormalizedValue{
			Value: float32(intVal) / 32767.0,
		},
	}
	obj.Value = cmd

	return 3, nil
}

// 解析设点命令-标度化值
func (c *ASDUCoder) decodeSetPointCommandScaled(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 3 {
		return 0, errors.New("insufficient data for set point command scaled")
	}

	cmd := &SetPointCommand{
		Select: (data[2] & 0x80) != 0,
		Value: &ScaledValue{
			Value: int16(binary.LittleEndian.Uint16(data[0:2])),
		},
	}
	obj.Value = cmd

	return 3, nil
}

// 解析设点命令-短浮点数
func (c *ASDUCoder) decodeSetPointCommandFloat(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 5 {
		return 0, errors.New("insufficient data for set point command float")
	}

	bits := binary.LittleEndian.Uint32(data[0:4])
	cmd := &SetPointCommand{
		Select: (data[4] & 0x80) != 0,
		Value: &FloatValue{
			Value: math.Float32frombits(bits),
		},
	}
	obj.Value = cmd

	return 5, nil
}

// 解析总召唤命令
func (c *ASDUCoder) decodeInterrogationCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for interrogation command")
	}

	cmd := &InterrogationCommand{
		QOI: data[0],
	}
	obj.Value = cmd

	return 1, nil
}

// 解析电度召唤命令
func (c *ASDUCoder) decodeCounterInterrogationCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for counter interrogation command")
	}

	cmd := &CounterInterrogationCommand{
		QCC: data[0],
	}
	obj.Value = cmd

	return 1, nil
}

// 解析时钟同步命令
func (c *ASDUCoder) decodeClockSyncCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 7 {
		return 0, errors.New("insufficient data for clock sync command")
	}

	cp56, err := DecodeCP56Time2a(data[0:7])
	if err != nil {
		return 0, err
	}

	cmd := &ClockSyncCommand{
		Time: cp56.ToTime(),
	}
	obj.Value = cmd

	return 7, nil
}

// 解析测试命令
func (c *ASDUCoder) decodeTestCommand(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("insufficient data for test command")
	}

	cmd := &TestCommand{
		FFS1: data[0],
		FFS2: data[1],
	}
	obj.Value = cmd

	return 2, nil
}

// 解析初始化结束
func (c *ASDUCoder) decodeEndOfInitialization(obj *InformationObject, data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("insufficient data for end of initialization")
	}

	eoi := &EndOfInitialization{
		COI: data[0],
	}
	obj.Value = eoi

	return 1, nil
}

// ==================== 辅助函数 ====================

// 从信息体获取值字符串
func GetInfoObjectValueString(obj InformationObject) string {
	switch v := obj.Value.(type) {
	case *SinglePointInfo:
		if v.Value {
			return "ON"
		}
		return "OFF"
	case *DoublePointInfo:
		switch v.Value {
		case DP_OFF:
			return "OFF"
		case DP_ON:
			return "ON"
		default:
			return "INDETERMINATE"
		}
	case *StepPositionInfo:
		return fmt.Sprintf("STEP:%d", v.Value)
	case *Bitstring32:
		return fmt.Sprintf("BITS:0x%08X", v.Value)
	case *NormalizedValue:
		return fmt.Sprintf("%.4f", v.Value)
	case *ScaledValue:
		return fmt.Sprintf("%d", v.Value)
	case *FloatValue:
		return fmt.Sprintf("%.4f", v.Value)
	case *IntegratedTotal:
		return fmt.Sprintf("TOTAL:%d", v.Value)
	case *SingleCommand:
		return fmt.Sprintf("CMD:%v", v.On)
	case *DoubleCommand:
		return fmt.Sprintf("CMD:%d", v.State)
	case *RegulatingStepCommand:
		return fmt.Sprintf("STEP_CMD:%d", v.State)
	case *InterrogationCommand:
		return fmt.Sprintf("GI:%d", v.QOI)
	case *CounterInterrogationCommand:
		return fmt.Sprintf("CI:%d", v.QCC)
	case *ClockSyncCommand:
		return fmt.Sprintf("TIME:%s", v.Time.Format(time.RFC3339))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 从信息体获取质量字符串
func GetInfoObjectQualityString(obj InformationObject) string {
	var q Quality
	switch v := obj.Value.(type) {
	case *SinglePointInfo:
		q = v.Quality
	case *DoublePointInfo:
		q = v.Quality
	case *StepPositionInfo:
		q = v.Quality
	case *Bitstring32:
		q = v.Quality
	case *NormalizedValue:
		q = v.Quality
	case *ScaledValue:
		q = v.Quality
	case *FloatValue:
		q = v.Quality
	case *IntegratedTotal:
		q = v.Quality
	default:
		return ""
	}

	var flags []string
	if q.Invalid {
		flags = append(flags, "IV")
	}
	if q.NotCurrent {
		flags = append(flags, "NT")
	}
	if q.Substituted {
		flags = append(flags, "SB")
	}
	if q.Blocked {
		flags = append(flags, "BL")
	}
	if q.Overflow {
		flags = append(flags, "OV")
	}

	if len(flags) == 0 {
		return "OK"
	}

	result := ""
	for i, f := range flags {
		if i > 0 {
			result += "|"
		}
		result += f
	}
	return result
}
