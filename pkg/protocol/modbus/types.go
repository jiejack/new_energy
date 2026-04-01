package modbus

import (
	"encoding/binary"
	"fmt"
)

// FunctionCode Modbus功能码定义
type FunctionCode byte

const (
	// 读线圈状态
	FuncReadCoils FunctionCode = 0x01
	// 读离散输入
	FuncReadDiscreteInputs FunctionCode = 0x02
	// 读保持寄存器
	FuncReadHoldingRegisters FunctionCode = 0x03
	// 读输入寄存器
	FuncReadInputRegisters FunctionCode = 0x04
	// 写单个线圈
	FuncWriteSingleCoil FunctionCode = 0x05
	// 写单个寄存器
	FuncWriteSingleRegister FunctionCode = 0x06
	// 写多个线圈
	FuncWriteMultipleCoils FunctionCode = 0x0F
	// 写多个寄存器
	FuncWriteMultipleRegisters FunctionCode = 0x10
	// 读/写多个寄存器
	FuncReadWriteMultipleRegisters FunctionCode = 0x17
	// 掩码写寄存器
	FuncMaskWriteRegister FunctionCode = 0x16
	// 读FIFO队列
	FuncReadFIFOQueue FunctionCode = 0x18
)

// String 返回功能码的字符串表示
func (fc FunctionCode) String() string {
	switch fc {
	case FuncReadCoils:
		return "Read Coils (0x01)"
	case FuncReadDiscreteInputs:
		return "Read Discrete Inputs (0x02)"
	case FuncReadHoldingRegisters:
		return "Read Holding Registers (0x03)"
	case FuncReadInputRegisters:
		return "Read Input Registers (0x04)"
	case FuncWriteSingleCoil:
		return "Write Single Coil (0x05)"
	case FuncWriteSingleRegister:
		return "Write Single Register (0x06)"
	case FuncWriteMultipleCoils:
		return "Write Multiple Coils (0x0F)"
	case FuncWriteMultipleRegisters:
		return "Write Multiple Registers (0x10)"
	case FuncReadWriteMultipleRegisters:
		return "Read/Write Multiple Registers (0x17)"
	case FuncMaskWriteRegister:
		return "Mask Write Register (0x16)"
	case FuncReadFIFOQueue:
		return "Read FIFO Queue (0x18)"
	default:
		return fmt.Sprintf("Unknown Function Code (0x%02X)", byte(fc))
	}
}

// ExceptionCode Modbus异常码定义
type ExceptionCode byte

const (
	// 非法功能码
	ExcIllegalFunction ExceptionCode = 0x01
	// 非法数据地址
	ExcIllegalDataAddress ExceptionCode = 0x02
	// 非法数据值
	ExcIllegalDataValue ExceptionCode = 0x03
	// 从站设备故障
	ExcSlaveDeviceFailure ExceptionCode = 0x04
	// 确认
	ExcAcknowledge ExceptionCode = 0x05
	// 从站设备忙
	ExcSlaveDeviceBusy ExceptionCode = 0x06
	// 存储奇偶校验错误
	ExcMemoryParityError ExceptionCode = 0x08
	// 网关路径不可用
	ExcGatewayPathUnavailable ExceptionCode = 0x0A
	// 网关目标设备无响应
	ExcGatewayTargetDeviceFailedToRespond ExceptionCode = 0x0B
)

// String 返回异常码的字符串表示
func (ec ExceptionCode) String() string {
	switch ec {
	case ExcIllegalFunction:
		return "Illegal Function"
	case ExcIllegalDataAddress:
		return "Illegal Data Address"
	case ExcIllegalDataValue:
		return "Illegal Data Value"
	case ExcSlaveDeviceFailure:
		return "Slave Device Failure"
	case ExcAcknowledge:
		return "Acknowledge"
	case ExcSlaveDeviceBusy:
		return "Slave Device Busy"
	case ExcMemoryParityError:
		return "Memory Parity Error"
	case ExcGatewayPathUnavailable:
		return "Gateway Path Unavailable"
	case ExcGatewayTargetDeviceFailedToRespond:
		return "Gateway Target Device Failed to Respond"
	default:
		return fmt.Sprintf("Unknown Exception Code (0x%02X)", byte(ec))
	}
}

// ModbusError Modbus异常错误
type ModbusError struct {
	FunctionCode FunctionCode
	ExceptionCode ExceptionCode
}

func (e *ModbusError) Error() string {
	return fmt.Sprintf("Modbus exception: function=%s, exception=%s",
		e.FunctionCode.String(), e.ExceptionCode.String())
}

// NewModbusError 创建Modbus异常错误
func NewModbusError(fc FunctionCode, ec ExceptionCode) *ModbusError {
	return &ModbusError{
		FunctionCode: fc,
		ExceptionCode: ec,
	}
}

// PDU (Protocol Data Unit) 协议数据单元
type PDU struct {
	FunctionCode FunctionCode
	Data         []byte
}

// NewPDU 创建PDU
func NewPDU(fc FunctionCode, data []byte) *PDU {
	return &PDU{
		FunctionCode: fc,
		Data:         data,
	}
}

// Bytes 将PDU转换为字节切片
func (p *PDU) Bytes() []byte {
	result := make([]byte, 1+len(p.Data))
	result[0] = byte(p.FunctionCode)
	copy(result[1:], p.Data)
	return result
}

// IsException 检查是否为异常响应
func (p *PDU) IsException() bool {
	return p.FunctionCode&0x80 != 0
}

// GetExceptionCode 获取异常码
func (p *PDU) GetExceptionCode() ExceptionCode {
	if p.IsException() && len(p.Data) > 0 {
		return ExceptionCode(p.Data[0])
	}
	return 0
}

// MBAPHeader Modbus应用协议头（仅用于TCP）
type MBAPHeader struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        byte
}

// Bytes 将MBAP头转换为字节切片
func (h *MBAPHeader) Bytes() []byte {
	result := make([]byte, 7)
	binary.BigEndian.PutUint16(result[0:2], h.TransactionID)
	binary.BigEndian.PutUint16(result[2:4], h.ProtocolID)
	binary.BigEndian.PutUint16(result[4:6], h.Length)
	result[6] = h.UnitID
	return result
}

// ParseMBAPHeader 从字节切片解析MBAP头
func ParseMBAPHeader(data []byte) (*MBAPHeader, error) {
	if len(data) < 7 {
		return nil, fmt.Errorf("insufficient data for MBAP header: got %d bytes, need 7", len(data))
	}
	return &MBAPHeader{
		TransactionID: binary.BigEndian.Uint16(data[0:2]),
		ProtocolID:    binary.BigEndian.Uint16(data[2:4]),
		Length:        binary.BigEndian.Uint16(data[4:6]),
		UnitID:        data[6],
	}, nil
}

// RTUFrame RTU帧结构
type RTUFrame struct {
	SlaveID    byte
	FunctionCode FunctionCode
	Data       []byte
	CRC        uint16
}

// NewRTUFrame 创建RTU帧
func NewRTUFrame(slaveID byte, fc FunctionCode, data []byte) *RTUFrame {
	return &RTUFrame{
		SlaveID:      slaveID,
		FunctionCode: fc,
		Data:         data,
	}
}

// Bytes 将RTU帧转换为字节切片（包含CRC）
func (f *RTUFrame) Bytes() []byte {
	result := make([]byte, 2+len(f.Data)+2)
	result[0] = f.SlaveID
	result[1] = byte(f.FunctionCode)
	copy(result[2:], f.Data)
	crc := CalculateCRC(result[:len(result)-2])
	result[len(result)-2] = byte(crc)
	result[len(result)-1] = byte(crc >> 8)
	return result
}

// ParseRTUFrame 从字节切片解析RTU帧
func ParseRTUFrame(data []byte) (*RTUFrame, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("insufficient data for RTU frame: got %d bytes, need at least 4", len(data))
	}

	// 验证CRC
	crc := binary.LittleEndian.Uint16(data[len(data)-2:])
	calculatedCRC := CalculateCRC(data[:len(data)-2])
	if crc != calculatedCRC {
		return nil, fmt.Errorf("CRC mismatch: got 0x%04X, calculated 0x%04X", crc, calculatedCRC)
	}

	return &RTUFrame{
		SlaveID:      data[0],
		FunctionCode: FunctionCode(data[1]),
		Data:         data[2 : len(data)-2],
		CRC:          crc,
	}, nil
}

// ASCIIFrame ASCII帧结构
type ASCIIFrame struct {
	Start      byte // ':'
	SlaveID    byte
	FunctionCode FunctionCode
	Data       []byte
	LRC        byte
	End        []byte // CR LF
}

// NewASCIIFrame 创建ASCII帧
func NewASCIIFrame(slaveID byte, fc FunctionCode, data []byte) *ASCIIFrame {
	return &ASCIIFrame{
		Start:        ':',
		SlaveID:      slaveID,
		FunctionCode: fc,
		Data:         data,
		End:          []byte{0x0D, 0x0A},
	}
}

// Bytes 将ASCII帧转换为字节切片（ASCII编码）
func (f *ASCIIFrame) Bytes() []byte {
	// 计算LRC
	lrcData := make([]byte, 2+len(f.Data))
	lrcData[0] = f.SlaveID
	lrcData[1] = byte(f.FunctionCode)
	copy(lrcData[2:], f.Data)
	lrc := CalculateLRC(lrcData)

	// 转换为ASCII字符串
	result := make([]byte, 1) // 起始符':'
	result[0] = ':'
	result = append(result, byteToASCII(f.SlaveID)...)
	result = append(result, byteToASCII(byte(f.FunctionCode))...)
	for _, b := range f.Data {
		result = append(result, byteToASCII(b)...)
	}
	result = append(result, byteToASCII(lrc)...)
	result = append(result, 0x0D, 0x0A) // CR LF

	return result
}

// ParseASCIIFrame 从字节切片解析ASCII帧
func ParseASCIIFrame(data []byte) (*ASCIIFrame, error) {
	if len(data) < 9 { // : + 2(ID) + 2(FC) + 2(LRC) + CR + LF
		return nil, fmt.Errorf("insufficient data for ASCII frame: got %d bytes, need at least 9", len(data))
	}

	if data[0] != ':' {
		return nil, fmt.Errorf("invalid start character: expected ':', got '%c'", data[0])
	}

	if data[len(data)-2] != 0x0D || data[len(data)-1] != 0x0A {
		return nil, fmt.Errorf("invalid end characters: expected CR LF")
	}

	// 解析ASCII数据
	asciiData := data[1 : len(data)-2]
	binaryData, err := asciiToBytes(asciiData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ASCII data: %w", err)
	}

	if len(binaryData) < 3 {
		return nil, fmt.Errorf("insufficient binary data after ASCII decode")
	}

	// 验证LRC
	lrc := binaryData[len(binaryData)-1]
	lrcData := binaryData[:len(binaryData)-1]
	calculatedLRC := CalculateLRC(lrcData)
	if lrc != calculatedLRC {
		return nil, fmt.Errorf("LRC mismatch: got 0x%02X, calculated 0x%02X", lrc, calculatedLRC)
	}

	return &ASCIIFrame{
		Start:        ':',
		SlaveID:      binaryData[0],
		FunctionCode: FunctionCode(binaryData[1]),
		Data:         binaryData[2 : len(binaryData)-1],
		LRC:          lrc,
		End:          []byte{0x0D, 0x0A},
	}, nil
}

// byteToASCII 将字节转换为ASCII字符串（2个字符）
func byteToASCII(b byte) []byte {
	hex := fmt.Sprintf("%02X", b)
	return []byte(hex)
}

// asciiToBytes 将ASCII字符串转换为字节切片
func asciiToBytes(data []byte) ([]byte, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("ASCII data length must be even")
	}

	result := make([]byte, len(data)/2)
	for i := 0; i < len(result); i++ {
		b, err := hexCharToByte(data[i*2], data[i*2+1])
		if err != nil {
			return nil, err
		}
		result[i] = b
	}
	return result, nil
}

// hexCharToByte 将两个十六进制字符转换为一个字节
func hexCharToByte(hi, lo byte) (byte, error) {
	hiVal, err := hexCharToValue(hi)
	if err != nil {
		return 0, err
	}
	loVal, err := hexCharToValue(lo)
	if err != nil {
		return 0, err
	}
	return hiVal<<4 | loVal, nil
}

// hexCharToValue 将十六进制字符转换为数值
func hexCharToValue(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	default:
		return 0, fmt.Errorf("invalid hex character: '%c'", c)
	}
}

// CoilValue 线圈值常量
const (
	CoilOff uint16 = 0x0000
	CoilOn  uint16 = 0xFF00
)

// MaxQuantity 定义最大读取数量
const (
	MaxCoilsReadQuantity            uint16 = 2000
	MaxDiscreteInputsReadQuantity   uint16 = 2000
	MaxRegistersReadQuantity        uint16 = 125
	MaxCoilsWriteQuantity           uint16 = 1968
	MaxRegistersWriteQuantity       uint16 = 123
	MaxReadWriteRegistersQuantity   uint16 = 121
)
