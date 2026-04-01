package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ByteOrder 字节序类型
type ByteOrder int

const (
	// BigEndian 大端序（高位在前）
	BigEndian ByteOrder = iota
	// LittleEndian 小端序（低位在前）
	LittleEndian
)

// WordOrder 字顺序类型
type WordOrder int

const (
	// HighWordFirst 高字在前
	HighWordFirst WordOrder = iota
	// LowWordFirst 低字在前
	LowWordFirst
)

// Converter 数据类型转换器
type Converter struct {
	byteOrder ByteOrder
	wordOrder WordOrder
}

// NewConverter 创建转换器
func NewConverter(byteOrder ByteOrder, wordOrder WordOrder) *Converter {
	return &Converter{
		byteOrder: byteOrder,
		wordOrder: wordOrder,
	}
}

// ========== 16位整数转换 ==========

// Uint16ToBytes 将uint16转换为字节切片
func Uint16ToBytes(value uint16, order ByteOrder) []byte {
	data := make([]byte, 2)
	if order == BigEndian {
		binary.BigEndian.PutUint16(data, value)
	} else {
		binary.LittleEndian.PutUint16(data, value)
	}
	return data
}

// BytesToUint16 将字节切片转换为uint16
func BytesToUint16(data []byte, order ByteOrder) (uint16, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("insufficient data: need 2 bytes, got %d", len(data))
	}
	if order == BigEndian {
		return binary.BigEndian.Uint16(data), nil
	}
	return binary.LittleEndian.Uint16(data), nil
}

// Int16ToBytes 将int16转换为字节切片
func Int16ToBytes(value int16, order ByteOrder) []byte {
	return Uint16ToBytes(uint16(value), order)
}

// BytesToInt16 将字节切片转换为int16
func BytesToInt16(data []byte, order ByteOrder) (int16, error) {
	val, err := BytesToUint16(data, order)
	if err != nil {
		return 0, err
	}
	return int16(val), nil
}

// ========== 32位整数转换 ==========

// Uint32ToBytes 将uint32转换为字节切片（2个寄存器）
func Uint32ToBytes(value uint32, byteOrder ByteOrder, wordOrder WordOrder) []byte {
	data := make([]byte, 4)
	if byteOrder == BigEndian {
		binary.BigEndian.PutUint32(data, value)
	} else {
		binary.LittleEndian.PutUint32(data, value)
	}

	// 如果字顺序需要交换
	if wordOrder == LowWordFirst {
		data[0], data[1], data[2], data[3] = data[2], data[3], data[0], data[1]
	}

	return data
}

// BytesToUint32 将字节切片转换为uint32
func BytesToUint32(data []byte, byteOrder ByteOrder, wordOrder WordOrder) (uint32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("insufficient data: need 4 bytes, got %d", len(data))
	}

	// 如果字顺序需要交换
	processedData := make([]byte, 4)
	if wordOrder == LowWordFirst {
		processedData[0], processedData[1] = data[2], data[3]
		processedData[2], processedData[3] = data[0], data[1]
	} else {
		copy(processedData, data)
	}

	if byteOrder == BigEndian {
		return binary.BigEndian.Uint32(processedData), nil
	}
	return binary.LittleEndian.Uint32(processedData), nil
}

// Int32ToBytes 将int32转换为字节切片
func Int32ToBytes(value int32, byteOrder ByteOrder, wordOrder WordOrder) []byte {
	return Uint32ToBytes(uint32(value), byteOrder, wordOrder)
}

// BytesToInt32 将字节切片转换为int32
func BytesToInt32(data []byte, byteOrder ByteOrder, wordOrder WordOrder) (int32, error) {
	val, err := BytesToUint32(data, byteOrder, wordOrder)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// ========== 64位整数转换 ==========

// Uint64ToBytes 将uint64转换为字节切片（4个寄存器）
func Uint64ToBytes(value uint64, byteOrder ByteOrder, wordOrder WordOrder) []byte {
	data := make([]byte, 8)
	if byteOrder == BigEndian {
		binary.BigEndian.PutUint64(data, value)
	} else {
		binary.LittleEndian.PutUint64(data, value)
	}

	// 字顺序处理
	if wordOrder == LowWordFirst {
		// 交换字顺序：word3 word2 word1 word0 -> word0 word1 word2 word3
		result := make([]byte, 8)
		result[0], result[1] = data[6], data[7]
		result[2], result[3] = data[4], data[5]
		result[4], result[5] = data[2], data[3]
		result[6], result[7] = data[0], data[1]
		return result
	}

	return data
}

// BytesToUint64 将字节切片转换为uint64
func BytesToUint64(data []byte, byteOrder ByteOrder, wordOrder WordOrder) (uint64, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("insufficient data: need 8 bytes, got %d", len(data))
	}

	processedData := make([]byte, 8)
	if wordOrder == LowWordFirst {
		processedData[0], processedData[1] = data[6], data[7]
		processedData[2], processedData[3] = data[4], data[5]
		processedData[4], processedData[5] = data[2], data[3]
		processedData[6], processedData[7] = data[0], data[1]
	} else {
		copy(processedData, data)
	}

	if byteOrder == BigEndian {
		return binary.BigEndian.Uint64(processedData), nil
	}
	return binary.LittleEndian.Uint64(processedData), nil
}

// Int64ToBytes 将int64转换为字节切片
func Int64ToBytes(value int64, byteOrder ByteOrder, wordOrder WordOrder) []byte {
	return Uint64ToBytes(uint64(value), byteOrder, wordOrder)
}

// BytesToInt64 将字节切片转换为int64
func BytesToInt64(data []byte, byteOrder ByteOrder, wordOrder WordOrder) (int64, error) {
	val, err := BytesToUint64(data, byteOrder, wordOrder)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// ========== 32位浮点数转换 ==========

// Float32ToBytes 将float32转换为字节切片（2个寄存器）
func Float32ToBytes(value float32, byteOrder ByteOrder, wordOrder WordOrder) []byte {
	bits := math.Float32bits(value)
	return Uint32ToBytes(bits, byteOrder, wordOrder)
}

// BytesToFloat32 将字节切片转换为float32
func BytesToFloat32(data []byte, byteOrder ByteOrder, wordOrder WordOrder) (float32, error) {
	bits, err := BytesToUint32(data, byteOrder, wordOrder)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

// ========== 64位浮点数转换 ==========

// Float64ToBytes 将float64转换为字节切片（4个寄存器）
func Float64ToBytes(value float64, byteOrder ByteOrder, wordOrder WordOrder) []byte {
	bits := math.Float64bits(value)
	return Uint64ToBytes(bits, byteOrder, wordOrder)
}

// BytesToFloat64 将字节切片转换为float64
func BytesToFloat64(data []byte, byteOrder ByteOrder, wordOrder WordOrder) (float64, error) {
	bits, err := BytesToUint64(data, byteOrder, wordOrder)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// ========== BCD码转换 ==========

// BCDToUint8 将BCD码转换为uint8
// 例如：0x12 -> 12
func BCDToUint8(bcd byte) uint8 {
	return (bcd>>4)*10 + (bcd & 0x0F)
}

// Uint8ToBCD 将uint8转换为BCD码
// 例如：12 -> 0x12
func Uint8ToBCD(value uint8) byte {
	if value > 99 {
		return 0xFF // 超出范围
	}
	return ((value / 10) << 4) | (value % 10)
}

// BCDToUint16 将BCD码转换为uint16（4位BCD）
// 例如：0x1234 -> 1234
func BCDToUint16(bcd uint16) uint16 {
	result := uint16(0)
	result += uint16(BCDToUint8(byte(bcd >> 8))) * 100
	result += uint16(BCDToUint8(byte(bcd)))
	return result
}

// Uint16ToBCD 将uint16转换为BCD码
// 例如：1234 -> 0x1234
func Uint16ToBCD(value uint16) uint16 {
	if value > 9999 {
		return 0xFFFF // 超出范围
	}
	high := Uint8ToBCD(byte(value / 100))
	low := Uint8ToBCD(byte(value % 100))
	return (uint16(high) << 8) | uint16(low)
}

// BCDToUint32 将BCD码转换为uint32（8位BCD）
// 例如：0x12345678 -> 12345678
func BCDToUint32(bcd uint32) uint32 {
	result := uint32(0)
	result += uint32(BCDToUint8(byte(bcd >> 24))) * 10000000
	result += uint32(BCDToUint8(byte(bcd >> 16))) * 100000
	result += uint32(BCDToUint8(byte(bcd >> 8))) * 1000
	result += uint32(BCDToUint8(byte(bcd))) * 1
	return result
}

// Uint32ToBCD 将uint32转换为BCD码
// 例如：12345678 -> 0x12345678
func Uint32ToBCD(value uint32) uint32 {
	if value > 99999999 {
		return 0xFFFFFFFF // 超出范围
	}
	b1 := Uint8ToBCD(byte(value / 10000000))
	b2 := Uint8ToBCD(byte((value / 100000) % 100))
	b3 := Uint8ToBCD(byte((value / 1000) % 100))
	b4 := Uint8ToBCD(byte(value % 100))
	return (uint32(b1) << 24) | (uint32(b2) << 16) | (uint32(b3) << 8) | uint32(b4)
}

// ========== 寄存器数组转换 ==========

// RegistersToBytes 将寄存器数组转换为字节切片
func RegistersToBytes(registers []uint16, order ByteOrder) []byte {
	data := make([]byte, len(registers)*2)
	for i, reg := range registers {
		if order == BigEndian {
			binary.BigEndian.PutUint16(data[i*2:], reg)
		} else {
			binary.LittleEndian.PutUint16(data[i*2:], reg)
		}
	}
	return data
}

// BytesToRegisters 将字节切片转换为寄存器数组
func BytesToRegisters(data []byte, order ByteOrder) ([]uint16, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("data length must be even, got %d", len(data))
	}

	registers := make([]uint16, len(data)/2)
	for i := 0; i < len(registers); i++ {
		if order == BigEndian {
			registers[i] = binary.BigEndian.Uint16(data[i*2:])
		} else {
			registers[i] = binary.LittleEndian.Uint16(data[i*2:])
		}
	}
	return registers, nil
}

// ========== 位操作 ==========

// GetBit 从字节中获取指定位
func GetBit(data byte, bitIndex uint) bool {
	if bitIndex > 7 {
		return false
	}
	return (data & (1 << bitIndex)) != 0
}

// SetBit 设置字节中的指定位
func SetBit(data byte, bitIndex uint, value bool) byte {
	if bitIndex > 7 {
		return data
	}
	if value {
		return data | (1 << bitIndex)
	}
	return data & ^(1 << bitIndex)
}

// GetBits 从字节切片中获取多个位
func GetBits(data []byte, startBit, count uint) []bool {
	bits := make([]bool, count)
	for i := uint(0); i < count; i++ {
		byteIndex := (startBit + i) / 8
		bitIndex := (startBit + i) % 8
		if byteIndex < uint(len(data)) {
			bits[i] = GetBit(data[byteIndex], bitIndex)
		}
	}
	return bits
}

// SetBits 设置字节切片中的多个位
func SetBits(data []byte, startBit uint, values []bool) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	for i, value := range values {
		byteIndex := (startBit + uint(i)) / 8
		bitIndex := (startBit + uint(i)) % 8
		if byteIndex < uint(len(result)) {
			result[byteIndex] = SetBit(result[byteIndex], bitIndex, value)
		}
	}

	return result
}

// ========== Converter方法 ==========

// ConvertRegisters 转换寄存器数据
func (c *Converter) ConvertRegisters(data []byte) ([]uint16, error) {
	return BytesToRegisters(data, c.byteOrder)
}

// ConvertToUint16 转换为uint16
func (c *Converter) ConvertToUint16(data []byte) (uint16, error) {
	return BytesToUint16(data, c.byteOrder)
}

// ConvertToInt16 转换为int16
func (c *Converter) ConvertToInt16(data []byte) (int16, error) {
	return BytesToInt16(data, c.byteOrder)
}

// ConvertToUint32 转换为uint32
func (c *Converter) ConvertToUint32(data []byte) (uint32, error) {
	return BytesToUint32(data, c.byteOrder, c.wordOrder)
}

// ConvertToInt32 转换为int32
func (c *Converter) ConvertToInt32(data []byte) (int32, error) {
	return BytesToInt32(data, c.byteOrder, c.wordOrder)
}

// ConvertToFloat32 转换为float32
func (c *Converter) ConvertToFloat32(data []byte) (float32, error) {
	return BytesToFloat32(data, c.byteOrder, c.wordOrder)
}

// ConvertToFloat64 转换为float64
func (c *Converter) ConvertToFloat64(data []byte) (float64, error) {
	return BytesToFloat64(data, c.byteOrder, c.wordOrder)
}
