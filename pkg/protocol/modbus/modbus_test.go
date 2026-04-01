package modbus

import (
	"bytes"
	"math"
	"testing"
)

// ========== CRC校验测试 ==========

func TestCalculateCRC(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{
			name:     "Test case 1: Read holding registers request",
			data:     []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A},
			expected: 0xC5CD,
		},
		{
			name:     "Test case 2: Write single register request",
			data:     []byte{0x01, 0x06, 0x00, 0x01, 0x00, 0x03},
			expected: 0x988B,
		},
		{
			name:     "Test case 3: Simple data",
			data:     []byte{0x01, 0x02},
			expected: 0x4141,
		},
		{
			name:     "Test case 4: Empty data",
			data:     []byte{},
			expected: 0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCRC(tt.data)
			if result != tt.expected {
				t.Errorf("CalculateCRC() = 0x%04X, expected 0x%04X", result, tt.expected)
			}
		})
	}
}

func TestVerifyCRC(t *testing.T) {
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	crc := CalculateCRC(data)

	if !VerifyCRC(data, crc) {
		t.Error("VerifyCRC() returned false for valid CRC")
	}

	if VerifyCRC(data, 0x0000) {
		t.Error("VerifyCRC() returned true for invalid CRC")
	}
}

// ========== LRC校验测试 ==========

func TestCalculateLRC(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{
			name:     "Test case 1: Read holding registers",
			data:     []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A},
			expected: 0xF2,
		},
		{
			name:     "Test case 2: Write single register",
			data:     []byte{0x01, 0x06, 0x00, 0x01, 0x00, 0x03},
			expected: 0xF5,
		},
		{
			name:     "Test case 3: Simple data",
			data:     []byte{0x01, 0x02, 0x03},
			expected: 0xFA,
		},
		{
			name:     "Test case 4: Empty data",
			data:     []byte{},
			expected: 0x00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLRC(tt.data)
			if result != tt.expected {
				t.Errorf("CalculateLRC() = 0x%02X, expected 0x%02X", result, tt.expected)
			}
		})
	}
}

func TestVerifyLRC(t *testing.T) {
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	lrc := CalculateLRC(data)

	if !VerifyLRC(data, lrc) {
		t.Error("VerifyLRC() returned false for valid LRC")
	}

	if VerifyLRC(data, 0x00) {
		t.Error("VerifyLRC() returned true for invalid LRC")
	}
}

// ========== 数据类型转换测试 ==========

func TestUint16Conversion(t *testing.T) {
	tests := []struct {
		name  string
		value uint16
		order ByteOrder
	}{
		{"BigEndian 0x1234", 0x1234, BigEndian},
		{"LittleEndian 0x1234", 0x1234, LittleEndian},
		{"BigEndian 0xFFFF", 0xFFFF, BigEndian},
		{"LittleEndian 0x0000", 0x0000, LittleEndian},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Uint16ToBytes(tt.value, tt.order)
			result, err := BytesToUint16(bytes, tt.order)
			if err != nil {
				t.Fatalf("BytesToUint16() error: %v", err)
			}
			if result != tt.value {
				t.Errorf("Conversion failed: got 0x%04X, expected 0x%04X", result, tt.value)
			}
		})
	}
}

func TestInt16Conversion(t *testing.T) {
	tests := []struct {
		name  string
		value int16
		order ByteOrder
	}{
		{"Positive value", 12345, BigEndian},
		{"Negative value", -12345, BigEndian},
		{"Zero", 0, LittleEndian},
		{"Max value", 32767, BigEndian},
		{"Min value", -32768, LittleEndian},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Int16ToBytes(tt.value, tt.order)
			result, err := BytesToInt16(bytes, tt.order)
			if err != nil {
				t.Fatalf("BytesToInt16() error: %v", err)
			}
			if result != tt.value {
				t.Errorf("Conversion failed: got %d, expected %d", result, tt.value)
			}
		})
	}
}

func TestUint32Conversion(t *testing.T) {
	tests := []struct {
		name      string
		value     uint32
		byteOrder ByteOrder
		wordOrder WordOrder
	}{
		{"BigEndian HighWordFirst", 0x12345678, BigEndian, HighWordFirst},
		{"BigEndian LowWordFirst", 0x12345678, BigEndian, LowWordFirst},
		{"LittleEndian HighWordFirst", 0x12345678, LittleEndian, HighWordFirst},
		{"LittleEndian LowWordFirst", 0x12345678, LittleEndian, LowWordFirst},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Uint32ToBytes(tt.value, tt.byteOrder, tt.wordOrder)
			result, err := BytesToUint32(bytes, tt.byteOrder, tt.wordOrder)
			if err != nil {
				t.Fatalf("BytesToUint32() error: %v", err)
			}
			if result != tt.value {
				t.Errorf("Conversion failed: got 0x%08X, expected 0x%08X", result, tt.value)
			}
		})
	}
}

func TestFloat32Conversion(t *testing.T) {
	tests := []struct {
		name      string
		value     float32
		byteOrder ByteOrder
		wordOrder WordOrder
	}{
		{"Positive value", 123.456, BigEndian, HighWordFirst},
		{"Negative value", -123.456, BigEndian, HighWordFirst},
		{"Zero", 0.0, LittleEndian, LowWordFirst},
		{"Max value", math.MaxFloat32, BigEndian, HighWordFirst},
		{"Small value", 0.0001, LittleEndian, LowWordFirst},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Float32ToBytes(tt.value, tt.byteOrder, tt.wordOrder)
			result, err := BytesToFloat32(bytes, tt.byteOrder, tt.wordOrder)
			if err != nil {
				t.Fatalf("BytesToFloat32() error: %v", err)
			}
			// 使用容差比较浮点数
			if math.Abs(float64(result-tt.value)) > 0.001 {
				t.Errorf("Conversion failed: got %f, expected %f", result, tt.value)
			}
		})
	}
}

func TestFloat64Conversion(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		byteOrder ByteOrder
		wordOrder WordOrder
	}{
		{"Positive value", 123.456789, BigEndian, HighWordFirst},
		{"Negative value", -123.456789, BigEndian, HighWordFirst},
		{"Zero", 0.0, LittleEndian, LowWordFirst},
		{"Max value", math.MaxFloat64, BigEndian, HighWordFirst},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Float64ToBytes(tt.value, tt.byteOrder, tt.wordOrder)
			result, err := BytesToFloat64(bytes, tt.byteOrder, tt.wordOrder)
			if err != nil {
				t.Fatalf("BytesToFloat64() error: %v", err)
			}
			// 使用容差比较浮点数
			if math.Abs(result-tt.value) > 0.0000001 {
				t.Errorf("Conversion failed: got %f, expected %f", result, tt.value)
			}
		})
	}
}

// ========== BCD码转换测试 ==========

func TestBCDConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    uint8
		bcd      byte
	}{
		{"Zero", 0, 0x00},
		{"Single digit", 5, 0x05},
		{"Double digit", 12, 0x12},
		{"Max value", 99, 0x99},
	}

	for _, tt := range tests {
		t.Run("Uint8ToBCD_"+tt.name, func(t *testing.T) {
			result := Uint8ToBCD(tt.value)
			if result != tt.bcd {
				t.Errorf("Uint8ToBCD() = 0x%02X, expected 0x%02X", result, tt.bcd)
			}
		})

		t.Run("BCDToUint8_"+tt.name, func(t *testing.T) {
			result := BCDToUint8(tt.bcd)
			if result != tt.value {
				t.Errorf("BCDToUint8() = %d, expected %d", result, tt.value)
			}
		})
	}
}

func TestBCDUint16Conversion(t *testing.T) {
	tests := []struct {
		value uint16
		bcd   uint16
	}{
		{0, 0x0000},
		{1234, 0x1234},
		{9999, 0x9999},
	}

	for _, tt := range tests {
		t.Run("Uint16ToBCD", func(t *testing.T) {
			result := Uint16ToBCD(tt.value)
			if result != tt.bcd {
				t.Errorf("Uint16ToBCD() = 0x%04X, expected 0x%04X", result, tt.bcd)
			}
		})

		t.Run("BCDToUint16", func(t *testing.T) {
			result := BCDToUint16(tt.bcd)
			if result != tt.value {
				t.Errorf("BCDToUint16() = %d, expected %d", result, tt.value)
			}
		})
	}
}

// ========== RTU帧测试 ==========

func TestRTUFrame(t *testing.T) {
	tests := []struct {
		name         string
		slaveID      byte
		functionCode FunctionCode
		data         []byte
	}{
		{"Read holding registers", 0x01, FuncReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x0A}},
		{"Write single register", 0x01, FuncWriteSingleRegister, []byte{0x00, 0x01, 0x00, 0x03}},
		{"Read coils", 0x02, FuncReadCoils, []byte{0x00, 0x00, 0x00, 0x08}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建帧
			frame := NewRTUFrame(tt.slaveID, tt.functionCode, tt.data)
			frameBytes := frame.Bytes()

			// 解析帧
			parsed, err := ParseRTUFrame(frameBytes)
			if err != nil {
				t.Fatalf("ParseRTUFrame() error: %v", err)
			}

			// 验证解析结果
			if parsed.SlaveID != tt.slaveID {
				t.Errorf("SlaveID mismatch: got 0x%02X, expected 0x%02X", parsed.SlaveID, tt.slaveID)
			}
			if parsed.FunctionCode != tt.functionCode {
				t.Errorf("FunctionCode mismatch: got 0x%02X, expected 0x%02X", parsed.FunctionCode, tt.functionCode)
			}
			if !bytes.Equal(parsed.Data, tt.data) {
				t.Errorf("Data mismatch: got %v, expected %v", parsed.Data, tt.data)
			}
		})
	}
}

func TestRTUFrameCRCValidation(t *testing.T) {
	frame := NewRTUFrame(0x01, FuncReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x0A})
	frameBytes := frame.Bytes()

	// 正确的CRC
	_, err := ParseRTUFrame(frameBytes)
	if err != nil {
		t.Errorf("ParseRTUFrame() failed with valid CRC: %v", err)
	}

	// 错误的CRC
	frameBytes[len(frameBytes)-1] ^= 0xFF
	_, err = ParseRTUFrame(frameBytes)
	if err == nil {
		t.Error("ParseRTUFrame() should fail with invalid CRC")
	}
}

// ========== ASCII帧测试 ==========

func TestASCIIFrame(t *testing.T) {
	tests := []struct {
		name         string
		slaveID      byte
		functionCode FunctionCode
		data         []byte
	}{
		{"Read holding registers", 0x01, FuncReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x0A}},
		{"Write single register", 0x01, FuncWriteSingleRegister, []byte{0x00, 0x01, 0x00, 0x03}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建帧
			frame := NewASCIIFrame(tt.slaveID, tt.functionCode, tt.data)
			frameBytes := frame.Bytes()

			// 验证起始符
			if frameBytes[0] != ':' {
				t.Errorf("Invalid start character: got '%c', expected ':'", frameBytes[0])
			}

			// 验证结束符
			if frameBytes[len(frameBytes)-2] != 0x0D || frameBytes[len(frameBytes)-1] != 0x0A {
				t.Error("Invalid end characters: expected CR LF")
			}

			// 解析帧
			parsed, err := ParseASCIIFrame(frameBytes)
			if err != nil {
				t.Fatalf("ParseASCIIFrame() error: %v", err)
			}

			// 验证解析结果
			if parsed.SlaveID != tt.slaveID {
				t.Errorf("SlaveID mismatch: got 0x%02X, expected 0x%02X", parsed.SlaveID, tt.slaveID)
			}
			if parsed.FunctionCode != tt.functionCode {
				t.Errorf("FunctionCode mismatch: got 0x%02X, expected 0x%02X", parsed.FunctionCode, tt.functionCode)
			}
			if !bytes.Equal(parsed.Data, tt.data) {
				t.Errorf("Data mismatch: got %v, expected %v", parsed.Data, tt.data)
			}
		})
	}
}

func TestASCIIFrameLRCValidation(t *testing.T) {
	frame := NewASCIIFrame(0x01, FuncReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x0A})
	frameBytes := frame.Bytes()

	// 正确的LRC
	_, err := ParseASCIIFrame(frameBytes)
	if err != nil {
		t.Errorf("ParseASCIIFrame() failed with valid LRC: %v", err)
	}

	// 错误的LRC（修改数据部分）
	frameBytes[3] ^= 0xFF
	_, err = ParseASCIIFrame(frameBytes)
	if err == nil {
		t.Error("ParseASCIIFrame() should fail with invalid LRC")
	}
}

// ========== PDU测试 ==========

func TestPDU(t *testing.T) {
	pdu := NewPDU(FuncReadHoldingRegisters, []byte{0x00, 0x01, 0x00, 0x0A})
	bytes := pdu.Bytes()

	if len(bytes) != 5 {
		t.Errorf("PDU length mismatch: got %d, expected 5", len(bytes))
	}

	if bytes[0] != byte(FuncReadHoldingRegisters) {
		t.Errorf("Function code mismatch: got 0x%02X, expected 0x%02X", bytes[0], FuncReadHoldingRegisters)
	}
}

func TestPDUException(t *testing.T) {
	// 正常响应
	normalPDU := &PDU{
		FunctionCode: FuncReadHoldingRegisters,
		Data:         []byte{0x02, 0x12, 0x34},
	}

	if normalPDU.IsException() {
		t.Error("Normal PDU should not be an exception")
	}

	// 异常响应
	exceptionPDU := &PDU{
		FunctionCode: FuncReadHoldingRegisters | 0x80,
		Data:         []byte{0x02},
	}

	if !exceptionPDU.IsException() {
		t.Error("Exception PDU should be an exception")
	}

	if exceptionPDU.GetExceptionCode() != ExcIllegalDataAddress {
		t.Errorf("Exception code mismatch: got 0x%02X, expected 0x%02X",
			exceptionPDU.GetExceptionCode(), ExcIllegalDataAddress)
	}
}

// ========== MBAP头测试 ==========

func TestMBAPHeader(t *testing.T) {
	header := &MBAPHeader{
		TransactionID: 0x1234,
		ProtocolID:    0x0000,
		Length:        0x0006,
		UnitID:        0x01,
	}

	bytes := header.Bytes()

	if len(bytes) != 7 {
		t.Errorf("MBAP header length mismatch: got %d, expected 7", len(bytes))
	}

	// 解析
	parsed, err := ParseMBAPHeader(bytes)
	if err != nil {
		t.Fatalf("ParseMBAPHeader() error: %v", err)
	}

	if parsed.TransactionID != header.TransactionID {
		t.Errorf("TransactionID mismatch: got 0x%04X, expected 0x%04X",
			parsed.TransactionID, header.TransactionID)
	}

	if parsed.UnitID != header.UnitID {
		t.Errorf("UnitID mismatch: got 0x%02X, expected 0x%02X", parsed.UnitID, header.UnitID)
	}
}

// ========== 位操作测试 ==========

func TestBitOperations(t *testing.T) {
	// Test GetBit
	data := byte(0x55) // 01010101
	if !GetBit(data, 0) {
		t.Error("Bit 0 should be 1")
	}
	if GetBit(data, 1) {
		t.Error("Bit 1 should be 0")
	}

	// Test SetBit
	result := SetBit(data, 1, true)
	if result != 0x57 { // 01010111
		t.Errorf("SetBit() = 0x%02X, expected 0x57", result)
	}

	result = SetBit(data, 0, false)
	if result != 0x54 { // 01010100
		t.Errorf("SetBit() = 0x%02X, expected 0x54", result)
	}
}

func TestBitsOperations(t *testing.T) {
	data := []byte{0x55, 0xAA} // 01010101 10101010
	bits := GetBits(data, 0, 16)

	expected := []bool{true, false, true, false, true, false, true, false,
		false, true, false, true, false, true, false, true}

	for i, bit := range bits {
		if bit != expected[i] {
			t.Errorf("Bit %d mismatch: got %v, expected %v", i, bit, expected[i])
		}
	}

	// Test SetBits
	values := []bool{false, true, false, true, false, true, false, true,
		true, false, true, false, true, false, true, false}
	result := SetBits(data, 0, values)

	// 结果应该是反过来的：0xAA, 0x55
	if result[0] != 0xAA || result[1] != 0x55 {
		t.Errorf("SetBits() = %v, expected [0xAA, 0x55]", result)
	}
}

// ========== 寄存器转换测试 ==========

func TestRegistersConversion(t *testing.T) {
	registers := []uint16{0x1234, 0x5678, 0x9ABC}

	// BigEndian
	bytes := RegistersToBytes(registers, BigEndian)
	result, err := BytesToRegisters(bytes, BigEndian)
	if err != nil {
		t.Fatalf("BytesToRegisters() error: %v", err)
	}

	for i, reg := range result {
		if reg != registers[i] {
			t.Errorf("Register %d mismatch: got 0x%04X, expected 0x%04X", i, reg, registers[i])
		}
	}

	// LittleEndian
	bytes = RegistersToBytes(registers, LittleEndian)
	result, err = BytesToRegisters(bytes, LittleEndian)
	if err != nil {
		t.Fatalf("BytesToRegisters() error: %v", err)
	}

	for i, reg := range result {
		if reg != registers[i] {
			t.Errorf("Register %d mismatch: got 0x%04X, expected 0x%04X", i, reg, registers[i])
		}
	}
}

// ========== 功能码和异常码测试 ==========

func TestFunctionCodeString(t *testing.T) {
	tests := []struct {
		fc       FunctionCode
		expected string
	}{
		{FuncReadCoils, "Read Coils (0x01)"},
		{FuncReadHoldingRegisters, "Read Holding Registers (0x03)"},
		{FuncWriteSingleRegister, "Write Single Register (0x06)"},
		{FunctionCode(0xFF), "Unknown Function Code (0xFF)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if result := tt.fc.String(); result != tt.expected {
				t.Errorf("FunctionCode.String() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestExceptionCodeString(t *testing.T) {
	tests := []struct {
		ec       ExceptionCode
		expected string
	}{
		{ExcIllegalFunction, "Illegal Function"},
		{ExcIllegalDataAddress, "Illegal Data Address"},
		{ExcSlaveDeviceFailure, "Slave Device Failure"},
		{ExceptionCode(0xFF), "Unknown Exception Code (0xFF)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if result := tt.ec.String(); result != tt.expected {
				t.Errorf("ExceptionCode.String() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestModbusError(t *testing.T) {
	err := NewModbusError(FuncReadHoldingRegisters, ExcIllegalDataAddress)
	expected := "Modbus exception: function=Read Holding Registers (0x03), exception=Illegal Data Address"

	if err.Error() != expected {
		t.Errorf("ModbusError.Error() = %s, expected %s", err.Error(), expected)
	}
}

// ========== 基准测试 ==========

func BenchmarkCalculateCRC(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateCRC(data)
	}
}

func BenchmarkCalculateLRC(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateLRC(data)
	}
}

func BenchmarkFloat32ToBytes(b *testing.B) {
	value := float32(123.456)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Float32ToBytes(value, BigEndian, HighWordFirst)
	}
}

func BenchmarkBytesToFloat32(b *testing.B) {
	data := Float32ToBytes(123.456, BigEndian, HighWordFirst)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BytesToFloat32(data, BigEndian, HighWordFirst)
	}
}
