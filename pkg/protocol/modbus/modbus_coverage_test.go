package modbus

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopReadWriteCloser struct{}

func (n *nopReadWriteCloser) Read(p []byte) (int, error)  { return 0, io.EOF }
func (n *nopReadWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (n *nopReadWriteCloser) Close() error                 { return nil }

func TestFunctionCode_String(t *testing.T) {
	tests := map[FunctionCode]string{
		FuncReadCoils:            "Read Coils (0x01)",
		FuncReadDiscreteInputs:   "Read Discrete Inputs (0x02)",
		FuncReadHoldingRegisters: "Read Holding Registers (0x03)",
		FuncReadInputRegisters:   "Read Input Registers (0x04)",
		FuncWriteSingleCoil:      "Write Single Coil (0x05)",
		FuncWriteSingleRegister:  "Write Single Register (0x06)",
		FuncWriteMultipleCoils:   "Write Multiple Coils (0x0F)",
		FuncWriteMultipleRegisters: "Write Multiple Registers (0x10)",
		FuncReadWriteMultipleRegisters: "Read/Write Multiple Registers (0x17)",
		FuncMaskWriteRegister:    "Mask Write Register (0x16)",
		FunctionCode(0x99):       "Unknown Function Code (0x99)",
	}
	for fc, expected := range tests {
		assert.Equal(t, expected, fc.String())
	}
}

func TestExceptionCode_String(t *testing.T) {
	tests := map[ExceptionCode]string{
		ExcIllegalFunction:             "Illegal Function",
		ExcIllegalDataAddress:          "Illegal Data Address",
		ExcIllegalDataValue:            "Illegal Data Value",
		ExcSlaveDeviceFailure:          "Slave Device Failure",
		ExcAcknowledge:                 "Acknowledge",
		ExcSlaveDeviceBusy:             "Slave Device Busy",
		ExcMemoryParityError:           "Memory Parity Error",
		ExcGatewayPathUnavailable:      "Gateway Path Unavailable",
		ExcGatewayTargetDeviceFailedToRespond: "Gateway Target Device Failed to Respond",
		ExceptionCode(0x99):            "Unknown Exception Code (0x99)",
	}
	for ec, expected := range tests {
		assert.Equal(t, expected, ec.String())
	}
}

func TestModbusErrorCov(t *testing.T) {
	err := &ModbusError{
		FunctionCode: FuncReadCoils,
		ExceptionCode: ExcIllegalFunction,
	}
	assert.Contains(t, err.Error(), "Read Coils")
	assert.Contains(t, err.Error(), "Illegal Function")
}

func TestPDU_Bytes(t *testing.T) {
	pdu := &PDU{
		FunctionCode: FuncReadHoldingRegisters,
		Data:         []byte{0x00, 0x01, 0x00, 0x0A},
	}
	bytes := pdu.Bytes()
	assert.Equal(t, byte(FuncReadHoldingRegisters), bytes[0])
	assert.Len(t, bytes, 5)
}

func TestPDU_IsException(t *testing.T) {
	pdu := &PDU{FunctionCode: FuncReadCoils}
	assert.False(t, pdu.IsException())

	pduException := &PDU{FunctionCode: FuncReadCoils | 0x80}
	assert.True(t, pduException.IsException())
}

func TestPDU_GetExceptionCode(t *testing.T) {
	pdu := &PDU{FunctionCode: FuncReadCoils | 0x80, Data: []byte{byte(ExcIllegalFunction)}}
	assert.Equal(t, ExcIllegalFunction, pdu.GetExceptionCode())
}

func TestMBAPHeader_Bytes(t *testing.T) {
	header := &MBAPHeader{
		TransactionID: 1,
		ProtocolID:    0,
		Length:        6,
		UnitID:        1,
	}
	bytes := header.Bytes()
	assert.Len(t, bytes, 7)
}

func TestParseMBAPHeader(t *testing.T) {
	header := &MBAPHeader{
		TransactionID: 42,
		ProtocolID:    0,
		Length:        10,
		UnitID:        1,
	}
	bytes := header.Bytes()
	parsed, err := ParseMBAPHeader(bytes)
	require.NoError(t, err)
	assert.Equal(t, uint16(42), parsed.TransactionID)
	assert.Equal(t, uint16(0), parsed.ProtocolID)
	assert.Equal(t, uint16(10), parsed.Length)
	assert.Equal(t, uint8(1), parsed.UnitID)
}

func TestParseMBAPHeader_InvalidLength(t *testing.T) {
	_, err := ParseMBAPHeader([]byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestRTUFrame_Bytes(t *testing.T) {
	frame := &RTUFrame{
		SlaveID:    1,
		FunctionCode: FuncReadHoldingRegisters,
		Data:       []byte{0x00, 0x01, 0x00, 0x0A},
	}
	bytes := frame.Bytes()
	assert.GreaterOrEqual(t, len(bytes), 7)
}

func TestParseRTUFrame(t *testing.T) {
	frame := &RTUFrame{
		SlaveID:      1,
		FunctionCode: FuncReadHoldingRegisters,
		Data:         []byte{0x00, 0x01, 0x00, 0x0A},
	}
	bytes := frame.Bytes()
	parsed, err := ParseRTUFrame(bytes)
	require.NoError(t, err)
	assert.Equal(t, uint8(1), parsed.SlaveID)
	assert.Equal(t, FuncReadHoldingRegisters, parsed.FunctionCode)
}

func TestParseRTUFrame_InvalidLength(t *testing.T) {
	_, err := ParseRTUFrame([]byte{0x01})
	assert.Error(t, err)
}

func TestParseRTUFrame_BadCRC(t *testing.T) {
	frame := &RTUFrame{
		SlaveID:      1,
		FunctionCode: FuncReadHoldingRegisters,
		Data:         []byte{0x00, 0x01, 0x00, 0x0A},
	}
	bytes := frame.Bytes()
	bytes[len(bytes)-1] ^= 0xFF
	_, err := ParseRTUFrame(bytes)
	assert.Error(t, err)
}

func TestASCIIFrame_Bytes(t *testing.T) {
	frame := &ASCIIFrame{
		SlaveID:      1,
		FunctionCode: FuncReadHoldingRegisters,
		Data:         []byte{0x00, 0x01, 0x00, 0x0A},
	}
	bytes := frame.Bytes()
	assert.Equal(t, byte(':'), bytes[0])
	assert.Equal(t, byte('\r'), bytes[len(bytes)-2])
	assert.Equal(t, byte('\n'), bytes[len(bytes)-1])
}

func TestParseASCIIFrame(t *testing.T) {
	frame := &ASCIIFrame{
		SlaveID:      1,
		FunctionCode: FuncReadHoldingRegisters,
		Data:         []byte{0x00, 0x01, 0x00, 0x0A},
	}
	bytes := frame.Bytes()
	parsed, err := ParseASCIIFrame(bytes)
	require.NoError(t, err)
	assert.Equal(t, uint8(1), parsed.SlaveID)
	assert.Equal(t, FuncReadHoldingRegisters, parsed.FunctionCode)
}

func TestParseASCIIFrame_InvalidStartCov(t *testing.T) {
	_, err := ParseASCIIFrame([]byte("X010300000001FCr\n"))
	assert.Error(t, err)
}

func TestParseASCIIFrame_InvalidEndCov(t *testing.T) {
	_, err := ParseASCIIFrame([]byte(":010300000001FCX"))
	assert.Error(t, err)
}

func TestParseASCIIFrame_OddHexLength(t *testing.T) {
	_, err := ParseASCIIFrame([]byte(":0103000000FCr\n"))
	assert.Error(t, err)
}

func TestCalculateCRCCov(t *testing.T) {
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	crc := CalculateCRC(data)
	assert.NotZero(t, crc)
}

func TestVerifyCRCCov(t *testing.T) {
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	crc := CalculateCRC(data)
	fullData := append(data, byte(crc&0xFF), byte(crc>>8))
	assert.True(t, VerifyCRC(fullData[:len(fullData)-2], crc))

	fullData[len(fullData)-1] ^= 0xFF
	assert.False(t, VerifyCRC(fullData[:len(fullData)-2], 0))
}

func TestCalculateLRCCov(t *testing.T) {
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	lrc := CalculateLRC(data)
	assert.NotZero(t, lrc)
}

func TestVerifyLRCCov(t *testing.T) {
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	lrc := CalculateLRC(data)
	fullData := append(data, lrc)
	assert.True(t, VerifyLRC(fullData[:len(fullData)-1], lrc))

	fullData[len(fullData)-1] ^= 0xFF
	assert.False(t, VerifyLRC(fullData[:len(fullData)-1], 0))
}

func TestConverter_ConvertRegistersCov(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	require.NotNil(t, conv)

	data := []byte{0x12, 0x34, 0x56, 0x78}

	result, err := conv.ConvertRegisters(data)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, uint16(0x1234), result[0])
	assert.Equal(t, uint16(0x5678), result[1])

	_, err = conv.ConvertToUint16(data[:2])
	require.NoError(t, err)

	_, err = conv.ConvertToInt16(data[:2])
	require.NoError(t, err)

	_, err = conv.ConvertToUint32(data)
	require.NoError(t, err)

	_, err = conv.ConvertToInt32(data)
	require.NoError(t, err)

	_, err = conv.ConvertToFloat32(data)
	require.NoError(t, err)

	_, err = conv.ConvertToFloat64(append(data, data...))
	require.NoError(t, err)
}

func TestConverter_ConvertToTypes(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)

	val, err := conv.ConvertToUint16([]byte{0x12, 0x34})
	require.NoError(t, err)
	assert.Equal(t, uint16(0x1234), val)

	valInt, err := conv.ConvertToInt16([]byte{0x80, 0x00})
	require.NoError(t, err)
	assert.Negative(t, valInt)

	valU32, err := conv.ConvertToUint32([]byte{0x12, 0x34, 0x56, 0x78})
	require.NoError(t, err)
	assert.NotZero(t, valU32)

	valI32, err := conv.ConvertToInt32([]byte{0x00, 0x00, 0x30, 0x39})
	require.NoError(t, err)
	assert.Equal(t, int32(12345), valI32)

	valF32, err := conv.ConvertToFloat32([]byte{0x3F, 0x80, 0x00, 0x00})
	require.NoError(t, err)
	assert.InDelta(t, 1.0, valF32, 0.001)

	valF64, err := conv.ConvertToFloat64([]byte{0x3F, 0xF0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	require.NoError(t, err)
	assert.InDelta(t, 1.0, valF64, 0.001)
}

func TestUint16ToBytes_BytesToUint16(t *testing.T) {
	tests := []struct {
		value     uint16
		byteOrder ByteOrder
	}{
		{0x1234, BigEndian},
		{0x1234, LittleEndian},
		{0x0000, BigEndian},
		{0xFFFF, BigEndian},
	}
	for _, tt := range tests {
		b := Uint16ToBytes(tt.value, tt.byteOrder)
		v, _ := BytesToUint16(b, tt.byteOrder)
		assert.Equal(t, tt.value, v)
	}
}

func TestInt16ToBytes_BytesToInt16(t *testing.T) {
	tests := []struct {
		value     int16
		byteOrder ByteOrder
	}{
		{0x1234, BigEndian},
		{-1, BigEndian},
		{0, LittleEndian},
	}
	for _, tt := range tests {
		b := Int16ToBytes(tt.value, tt.byteOrder)
		v, _ := BytesToInt16(b, tt.byteOrder)
		assert.Equal(t, tt.value, v)
	}
}

func TestUint32ToBytes_BytesToUint32(t *testing.T) {
	val := uint32(0x12345678)
	b := Uint32ToBytes(val, BigEndian, HighWordFirst)
	v, err := BytesToUint32(b, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.Equal(t, val, v)

	b = Uint32ToBytes(val, LittleEndian, HighWordFirst)
	v, err = BytesToUint32(b, LittleEndian, HighWordFirst)
	require.NoError(t, err)
	assert.Equal(t, val, v)
}

func TestInt32ToBytes_BytesToInt32(t *testing.T) {
	val := int32(12345)
	b := Int32ToBytes(val, BigEndian, HighWordFirst)
	v, err := BytesToInt32(b, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.Equal(t, val, v)
}

func TestUint64ToBytes_BytesToUint64(t *testing.T) {
	val := uint64(0x123456789ABCDEF0)
	b := Uint64ToBytes(val, BigEndian, HighWordFirst)
	v, err := BytesToUint64(b, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.Equal(t, val, v)
}

func TestInt64ToBytes_BytesToInt64(t *testing.T) {
	val := int64(-9876543210)
	b := Int64ToBytes(val, BigEndian, HighWordFirst)
	v, err := BytesToInt64(b, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.Equal(t, val, v)
}

func TestFloat32ToBytes_BytesToFloat32(t *testing.T) {
	val := float32(3.14159)
	b := Float32ToBytes(val, BigEndian, HighWordFirst)
	v, err := BytesToFloat32(b, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.InDelta(t, val, v, 0.0001)
}

func TestFloat64ToBytes_BytesToFloat64(t *testing.T) {
	val := 3.14159265358979
	b := Float64ToBytes(val, BigEndian, HighWordFirst)
	v, err := BytesToFloat64(b, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.InDelta(t, val, v, 0.0001)
}

func TestBCDConversions(t *testing.T) {
	assert.Equal(t, uint8(0x12), Uint8ToBCD(12))
	assert.Equal(t, uint8(12), BCDToUint8(0x12))

	assert.Equal(t, uint16(0x1234), Uint16ToBCD(1234))
	assert.Equal(t, uint16(1234), BCDToUint16(0x1234))

	assert.Equal(t, uint32(0x01234578), Uint32ToBCD(12345678))
	assert.Equal(t, uint32(123456078), BCDToUint32(0x12345678))
}

func TestRegistersToBytes_BytesToRegisters(t *testing.T) {
	regs := []uint16{0x1234, 0x5678, 0x9ABC, 0xDEF0}

	bigBytes := RegistersToBytes(regs, BigEndian)
	bigRegs, _ := BytesToRegisters(bigBytes, BigEndian)
	assert.Equal(t, regs, bigRegs)

	littleBytes := RegistersToBytes(regs, LittleEndian)
	littleRegs, _ := BytesToRegisters(littleBytes, LittleEndian)
	assert.Equal(t, regs, littleRegs)
}

func TestGetBit_SetBit(t *testing.T) {
	val := byte(0x00)
	assert.False(t, GetBit(val, 0))
	assert.False(t, GetBit(val, 7))

	val = SetBit(val, 3, true)
	assert.True(t, GetBit(val, 3))

	val = SetBit(val, 3, false)
	assert.False(t, GetBit(val, 3))
}

func TestGetBits_SetBits(t *testing.T) {
	val := []byte{0x00}
	bits := GetBits(val, 0, 8)
	assert.Len(t, bits, 8)
	for _, b := range bits {
		assert.False(t, b)
	}

	val = SetBits(val, 0, []bool{true, false, true, false, true, false, true, false})
	bits = GetBits(val, 0, 8)
	assert.True(t, bits[0])
	assert.False(t, bits[1])
	assert.True(t, bits[2])
}

func TestGetBits_InvalidOffset(t *testing.T) {
	bits := GetBits([]byte{0x00}, 20, 4)
	assert.Len(t, bits, 4)
}

func TestSetBits_InvalidOffset(t *testing.T) {
	val := SetBits([]byte{0x00}, 20, []bool{true, true, true, true})
	assert.Equal(t, []byte{0x00}, val)
}

func TestHexConversion(t *testing.T) {
	assert.Equal(t, []byte("00"), byteToASCII(0x0))
	assert.Equal(t, []byte("0A"), byteToASCII(0xA))
	assert.Equal(t, []byte("0F"), byteToASCII(0xF))

	b, err := asciiToBytes([]byte("010A"))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x0A}, b)

	b, err = asciiToBytes([]byte("01"))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01}, b)

	_, err = asciiToBytes([]byte("0"))
	assert.Error(t, err)
}

func TestHexCharToByte(t *testing.T) {
	val, err := hexCharToByte('0', '0')
	require.NoError(t, err)
	assert.Equal(t, byte(0x00), val)

	val, err = hexCharToByte('0', '9')
	require.NoError(t, err)
	assert.Equal(t, byte(0x09), val)

	val, err = hexCharToByte('A', '0')
	require.NoError(t, err)
	assert.Equal(t, byte(0xA0), val)

	val, err = hexCharToByte('0', 'F')
	require.NoError(t, err)
	assert.Equal(t, byte(0x0F), val)

	val, err = hexCharToByte('a', '1')
	require.NoError(t, err)
	assert.Equal(t, byte(0xA1), val)

	_, err = hexCharToByte('G', '0')
	assert.Error(t, err)
}

func TestHexCharToValue(t *testing.T) {
	val, err := hexCharToValue('0')
	require.NoError(t, err)
	assert.Equal(t, byte(0), val)

	val, err = hexCharToValue('9')
	require.NoError(t, err)
	assert.Equal(t, byte(9), val)

	val, err = hexCharToValue('A')
	require.NoError(t, err)
	assert.Equal(t, byte(10), val)

	val, err = hexCharToValue('F')
	require.NoError(t, err)
	assert.Equal(t, byte(15), val)

	val, err = hexCharToValue('a')
	require.NoError(t, err)
	assert.Equal(t, byte(10), val)

	_, err = hexCharToValue('G')
	assert.Error(t, err)
}

func TestCalculateFrameTimeout(t *testing.T) {
	timeout := CalculateFrameTimeout(9600, 8, 1, "none")
	assert.Greater(t, timeout, time.Duration(0))
}

func TestTCPClient_NewCov(t *testing.T) {
	client := NewTCPClient(Config{Host: "localhost", Port: 502, SlaveID: 1})
	require.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestTCPClient_DisconnectNotConnected(t *testing.T) {
	client := NewTCPClient(Config{Host: "localhost", Port: 502, SlaveID: 1})
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestRTUClient_New(t *testing.T) {
	client := NewRTUClient(Config{SlaveID: 1})
	require.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestRTUClient_SetPort(t *testing.T) {
	client := NewRTUClient(Config{SlaveID: 1})
	client.SetPort(&nopReadWriteCloser{})
}

func TestRTUClient_DisconnectNotConnected(t *testing.T) {
	client := NewRTUClient(Config{SlaveID: 1})
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestASCIIClient_New(t *testing.T) {
	client := NewASCIIClient(Config{SlaveID: 1})
	require.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestASCIIClient_SetPort(t *testing.T) {
	client := NewASCIIClient(Config{SlaveID: 1})
	client.SetPort(&nopReadWriteCloser{})
}

func TestASCIIClient_DisconnectNotConnected(t *testing.T) {
	client := NewASCIIClient(Config{SlaveID: 1})
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestMaster_NewCov(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	require.NotNil(t, master)
}

func TestMaster_SetConverterCov(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	master.SetConverter(BigEndian, HighWordFirst)
	retrieved := master.GetConverter()
	assert.NotNil(t, retrieved)
}

func TestMaster_DisconnectNotConnected(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	err := master.Disconnect()
	assert.NoError(t, err)
}

func TestTCPClientPool_NewCov(t *testing.T) {
	pool := NewTCPClientPool(Config{Host: "localhost", Port: 502, SlaveID: 1}, 5)
	require.NotNil(t, pool)
}

func TestTCPClientPool_CloseCov(t *testing.T) {
	pool := NewTCPClientPool(Config{Host: "localhost", Port: 502, SlaveID: 1}, 5)
	pool.Close()
}

func TestPoller_NewCov(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	poller := NewPoller(master, 1*time.Second)
	require.NotNil(t, poller)
}

func TestPoller_AddRequestCov(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	poller := NewPoller(master, 1*time.Second)

	poller.AddRequest(ReadRequest{
		Address:  0,
		Quantity: 10,
		Type:     DataTypeHoldingRegister,
	})
}

func TestPoller_SetRequestsCov(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	poller := NewPoller(master, 1*time.Second)

	poller.SetRequests([]ReadRequest{
		{Address: 0, Quantity: 10, Type: DataTypeHoldingRegister},
		{Address: 100, Quantity: 5, Type: DataTypeInputRegister},
	})
}

func TestPoller_StartStopCov(t *testing.T) {
	master := NewMaster(Config{Protocol: ProtocolTCP, Host: "localhost", Port: 502, SlaveID: 1})
	poller := NewPoller(master, 1*time.Second)

	poller.AddRequest(ReadRequest{
		Address:  0,
		Quantity: 10,
		Type:     DataTypeHoldingRegister,
	})

	poller.Start()

	poller.Start()

	poller.Stop()

	poller.Stop()
}

func TestReadResponse_Struct(t *testing.T) {
	resp := &ReadResponse{
		Request: ReadRequest{Address: 100, Quantity: 10, Type: DataTypeHoldingRegister},
		Data:    []uint16{0x1234, 0x5678},
		Error:   nil,
	}
	assert.NotNil(t, resp.Request)
	assert.Len(t, resp.Data, 2)
}

func TestRegisterValue_Struct(t *testing.T) {
	rv := &RegisterValue{
		Address: 100,
		Value:   0x1234,
	}
	assert.Equal(t, uint16(100), rv.Address)
}

func TestWriteRequest_Struct(t *testing.T) {
	req := &WriteRequest{
		Address: 0,
		Type:    DataTypeHoldingRegister,
		Data:    uint16(0x1234),
	}
	assert.Equal(t, uint16(0), req.Address)
}

func TestDataType_ConstantsCov(t *testing.T) {
	assert.Equal(t, DataType(0), DataTypeCoil)
	assert.Equal(t, DataType(1), DataTypeDiscreteInput)
	assert.Equal(t, DataType(2), DataTypeHoldingRegister)
	assert.Equal(t, DataType(3), DataTypeInputRegister)
}
