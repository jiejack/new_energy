package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ASCIClient Modbus ASCII客户端
type ASCIClient struct {
	config       Config
	serialConfig SerialConfig
	port         io.ReadWriteCloser
	mu           sync.Mutex
	connected    bool
}

// NewASCIIClient 创建ASCII客户端
func NewASCIIClient(config Config) *ASCIClient {
	return &ASCIClient{
		config: config,
	}
}

// NewASCIIClientWithSerial 创建带串口配置的ASCII客户端
func NewASCIIClientWithSerial(config Config, serialConfig SerialConfig) *ASCIClient {
	return &ASCIClient{
		config:       config,
		serialConfig: serialConfig,
	}
}

// SetPort 设置串口（用于依赖注入）
func (c *ASCIClient) SetPort(port io.ReadWriteCloser) {
	c.port = port
	c.connected = true
}

// Connect 连接串口
func (c *ASCIClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// 如果已经设置了port（通过SetPort），直接返回
	if c.port != nil {
		c.connected = true
		return nil
	}

	// 否则需要打开串口
	return errors.New("serial port not configured, use SetPort() or implement serial.OpenPort()")
}

// Disconnect 断开连接
func (c *ASCIClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	if c.port != nil {
		err := c.port.Close()
		c.port = nil
		c.connected = false
		return err
	}

	return nil
}

// IsConnected 检查连接状态
func (c *ASCIClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// sendRequest 发送请求并接收响应
func (c *ASCIClient) sendRequest(pdu *PDU) (*PDU, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	// 构建ASCII帧
	frame := NewASCIIFrame(c.config.SlaveID, pdu.FunctionCode, pdu.Data)
	frameBytes := frame.Bytes()

	// 发送请求
	n, err := c.port.Write(frameBytes)
	if err != nil {
		c.connected = false
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if n != len(frameBytes) {
		return nil, fmt.Errorf("incomplete write: sent %d of %d bytes", n, len(frameBytes))
	}

	// 读取响应
	response, err := c.readResponse(pdu.FunctionCode)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// readResponse 读取响应
func (c *ASCIClient) readResponse(expectedFC FunctionCode) (*PDU, error) {
	// 设置读取超时
	if c.config.Timeout > 0 {
		// 根据实际串口库设置超时
	}

	// 读取起始符 ':'
	startChar := make([]byte, 1)
	if _, err := io.ReadFull(c.port, startChar); err != nil {
		return nil, fmt.Errorf("failed to read start character: %w", err)
	}

	if startChar[0] != ':' {
		return nil, fmt.Errorf("invalid start character: expected ':', got '%c'", startChar[0])
	}

	// 读取直到遇到 CR LF
	var asciiData []byte
	buf := make([]byte, 1)

	for {
		if _, err := io.ReadFull(c.port, buf); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// 检查是否到达结束符
		if buf[0] == 0x0D {
			// 读取 LF
			if _, err := io.ReadFull(c.port, buf); err != nil {
				return nil, fmt.Errorf("failed to read LF: %w", err)
			}
			if buf[0] != 0x0A {
				return nil, fmt.Errorf("expected LF after CR, got '%c'", buf[0])
			}
			break
		}

		asciiData = append(asciiData, buf[0])

		// 防止无限读取（ASCII帧最大长度为513字节）
		if len(asciiData) > 513 {
			return nil, errors.New("response too long")
		}
	}

	// 解析ASCII帧
	frame, err := ParseASCIIFrame(append([]byte{':'}, append(asciiData, 0x0D, 0x0A)...))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ASCII frame: %w", err)
	}

	// 验证从站地址
	if frame.SlaveID != c.config.SlaveID {
		return nil, fmt.Errorf("slave ID mismatch: expected %d, got %d", c.config.SlaveID, frame.SlaveID)
	}

	// 检查异常响应
	if frame.FunctionCode&0x80 != 0 {
		if len(frame.Data) < 1 {
			return nil, errors.New("invalid exception response: no exception code")
		}
		return nil, NewModbusError(expectedFC, ExceptionCode(frame.Data[0]))
	}

	return &PDU{
		FunctionCode: frame.FunctionCode,
		Data:         frame.Data,
	}, nil
}

// ReadCoils 读取线圈状态（功能码 0x01）
func (c *ASCIClient) ReadCoils(address, quantity uint16) ([]byte, error) {
	if quantity < 1 || quantity > MaxCoilsReadQuantity {
		return nil, fmt.Errorf("quantity must be between 1 and %d", MaxCoilsReadQuantity)
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	pdu := NewPDU(FuncReadCoils, data)

	response, err := c.sendRequest(pdu)
	if err != nil {
		return nil, err
	}

	if len(response.Data) < 1 {
		return nil, errors.New("invalid response: no data")
	}

	byteCount := int(response.Data[0])
	if len(response.Data) < byteCount+1 {
		return nil, errors.New("invalid response: insufficient data")
	}

	return response.Data[1 : byteCount+1], nil
}

// ReadDiscreteInputs 读取离散输入（功能码 0x02）
func (c *ASCIClient) ReadDiscreteInputs(address, quantity uint16) ([]byte, error) {
	if quantity < 1 || quantity > MaxDiscreteInputsReadQuantity {
		return nil, fmt.Errorf("quantity must be between 1 and %d", MaxDiscreteInputsReadQuantity)
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	pdu := NewPDU(FuncReadDiscreteInputs, data)

	response, err := c.sendRequest(pdu)
	if err != nil {
		return nil, err
	}

	if len(response.Data) < 1 {
		return nil, errors.New("invalid response: no data")
	}

	byteCount := int(response.Data[0])
	if len(response.Data) < byteCount+1 {
		return nil, errors.New("invalid response: insufficient data")
	}

	return response.Data[1 : byteCount+1], nil
}

// ReadHoldingRegisters 读取保持寄存器（功能码 0x03）
func (c *ASCIClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
	if quantity < 1 || quantity > MaxRegistersReadQuantity {
		return nil, fmt.Errorf("quantity must be between 1 and %d", MaxRegistersReadQuantity)
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	pdu := NewPDU(FuncReadHoldingRegisters, data)

	response, err := c.sendRequest(pdu)
	if err != nil {
		return nil, err
	}

	if len(response.Data) < 1 {
		return nil, errors.New("invalid response: no data")
	}

	byteCount := int(response.Data[0])
	if len(response.Data) < byteCount+1 {
		return nil, errors.New("invalid response: insufficient data")
	}

	return response.Data[1 : byteCount+1], nil
}

// ReadInputRegisters 读取输入寄存器（功能码 0x04）
func (c *ASCIClient) ReadInputRegisters(address, quantity uint16) ([]byte, error) {
	if quantity < 1 || quantity > MaxRegistersReadQuantity {
		return nil, fmt.Errorf("quantity must be between 1 and %d", MaxRegistersReadQuantity)
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	pdu := NewPDU(FuncReadInputRegisters, data)

	response, err := c.sendRequest(pdu)
	if err != nil {
		return nil, err
	}

	if len(response.Data) < 1 {
		return nil, errors.New("invalid response: no data")
	}

	byteCount := int(response.Data[0])
	if len(response.Data) < byteCount+1 {
		return nil, errors.New("invalid response: insufficient data")
	}

	return response.Data[1 : byteCount+1], nil
}

// WriteSingleCoil 写单个线圈（功能码 0x05）
func (c *ASCIClient) WriteSingleCoil(address uint16, value bool) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)

	if value {
		binary.BigEndian.PutUint16(data[2:4], CoilOn)
	} else {
		binary.BigEndian.PutUint16(data[2:4], CoilOff)
	}

	pdu := NewPDU(FuncWriteSingleCoil, data)
	_, err := c.sendRequest(pdu)
	return err
}

// WriteSingleRegister 写单个寄存器（功能码 0x06）
func (c *ASCIClient) WriteSingleRegister(address uint16, value uint16) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], value)
	pdu := NewPDU(FuncWriteSingleRegister, data)

	_, err := c.sendRequest(pdu)
	return err
}

// WriteMultipleCoils 写多个线圈（功能码 0x0F）
func (c *ASCIClient) WriteMultipleCoils(address uint16, values []byte) error {
	quantity := uint16(len(values) * 8)
	if quantity < 1 || quantity > MaxCoilsWriteQuantity {
		return fmt.Errorf("quantity must be between 1 and %d", MaxCoilsWriteQuantity)
	}

	byteCount := byte(len(values))
	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	data[4] = byteCount
	copy(data[5:], values)

	pdu := NewPDU(FuncWriteMultipleCoils, data)
	_, err := c.sendRequest(pdu)
	return err
}

// WriteMultipleRegisters 写多个寄存器（功能码 0x10）
func (c *ASCIClient) WriteMultipleRegisters(address uint16, values []byte) error {
	quantity := uint16(len(values) / 2)
	if quantity < 1 || quantity > MaxRegistersWriteQuantity {
		return fmt.Errorf("quantity must be between 1 and %d", MaxRegistersWriteQuantity)
	}

	byteCount := byte(len(values))
	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	data[4] = byteCount
	copy(data[5:], values)

	pdu := NewPDU(FuncWriteMultipleRegisters, data)
	_, err := c.sendRequest(pdu)
	return err
}

// ASCII帧特点说明：
// 1. 起始符：冒号 ':'
// 2. 结束符：CR LF (0x0D 0x0A)
// 3. 数据格式：十六进制ASCII字符
// 4. 校验方式：LRC（纵向冗余校验）
// 5. 最大帧长度：513字节（包括起始符和结束符）
// 6. 相比RTU模式，ASCII模式传输效率较低，但更适合文本环境
