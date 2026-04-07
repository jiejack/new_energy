package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
)

// SerialConfig 串口配置
type SerialConfig struct {
	PortName     string
	BaudRate     int
	DataBits     int
	Parity       string // "none", "odd", "even"
	StopBits     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// RTUClient Modbus RTU客户端
type RTUClient struct {
	config        Config
	serialConfig  SerialConfig
	port          io.ReadWriteCloser
	mu            sync.Mutex
	connected     bool
	lastSendTime  time.Time
	frameInterval time.Duration // RTU帧间隔时间
}

// NewRTUClient 创建RTU客户端
func NewRTUClient(config Config) *RTUClient {
	return &RTUClient{
		config:        config,
		frameInterval: 1750 * time.Microsecond, // Modbus RTU帧间隔至少1.75ms
	}
}

// NewRTUClientWithSerial 创建带串口配置的RTU客户端
func NewRTUClientWithSerial(config Config, serialConfig SerialConfig) *RTUClient {
	return &RTUClient{
		config:        config,
		serialConfig:  serialConfig,
		frameInterval: 1750 * time.Microsecond,
	}
}

// SetPort 设置串口（用于依赖注入）
func (c *RTUClient) SetPort(port io.ReadWriteCloser) {
	c.port = port
	c.connected = true
}

// Connect 连接串口
func (c *RTUClient) Connect() error {
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
	// 注意：实际使用时需要引入串口库，如 go.bug.st/serial
	// 这里提供一个接口，实际实现需要根据具体串口库来完成
	return errors.New("serial port not configured, use SetPort() or implement serial.OpenPort()")
}

// Disconnect 断开连接
func (c *RTUClient) Disconnect() error {
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
func (c *RTUClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// sendRequest 发送请求并接收响应
func (c *RTUClient) sendRequest(pdu *PDU) (*PDU, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	// 确保帧间隔
	if !c.lastSendTime.IsZero() {
		elapsed := time.Since(c.lastSendTime)
		if elapsed < c.frameInterval {
			time.Sleep(c.frameInterval - elapsed)
		}
	}

	// 构建RTU帧
	frame := NewRTUFrame(c.config.SlaveID, pdu.FunctionCode, pdu.Data)
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

	c.lastSendTime = time.Now()

	// 读取响应
	response, err := c.readResponse(pdu.FunctionCode)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// readResponse 读取响应
func (c *RTUClient) readResponse(expectedFC FunctionCode) (*PDU, error) {
	// 设置读取超时
	if c.config.Timeout > 0 {
		// 这里需要根据实际串口库来设置超时
		// 大多数串口库支持设置读取超时
	}

	// 读取响应头（从站地址 + 功能码）
	header := make([]byte, 2)
	_, err := io.ReadFull(c.port, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	slaveID := header[0]
	functionCode := FunctionCode(header[1])

	// 验证从站地址
	if slaveID != c.config.SlaveID {
		return nil, fmt.Errorf("slave ID mismatch: expected %d, got %d", c.config.SlaveID, slaveID)
	}

	// 检查异常响应
	if functionCode&0x80 != 0 {
		// 异常响应：读取异常码和CRC
		exceptionData := make([]byte, 3) // 异常码(1) + CRC(2)
		_, err := io.ReadFull(c.port, exceptionData)
		if err != nil {
			return nil, fmt.Errorf("failed to read exception response: %w", err)
		}

		// 验证CRC
		fullFrame := make([]byte, 5)
		fullFrame[0] = slaveID
		fullFrame[1] = byte(functionCode)
		fullFrame[2] = exceptionData[0]
		copy(fullFrame[3:], exceptionData[1:3])

		if _, err := ParseRTUFrame(fullFrame); err != nil {
			return nil, fmt.Errorf("CRC validation failed: %w", err)
		}

		return nil, NewModbusError(expectedFC, ExceptionCode(exceptionData[0]))
	}

	// 正常响应：根据功能码读取数据
	var data []byte
	switch functionCode {
	case FuncReadCoils, FuncReadDiscreteInputs, FuncReadHoldingRegisters, FuncReadInputRegisters:
		// 读取字节计数
		byteCountByte := make([]byte, 1)
		if _, err := io.ReadFull(c.port, byteCountByte); err != nil {
			return nil, fmt.Errorf("failed to read byte count: %w", err)
		}
		byteCount := int(byteCountByte[0])

		// 读取数据 + CRC
		remaining := make([]byte, byteCount+2)
		if _, err := io.ReadFull(c.port, remaining); err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}

		// 验证CRC
		fullFrame := make([]byte, 3+byteCount+2)
		fullFrame[0] = slaveID
		fullFrame[1] = byte(functionCode)
		fullFrame[2] = byteCountByte[0]
		copy(fullFrame[3:], remaining)

		if _, err := ParseRTUFrame(fullFrame); err != nil {
			return nil, fmt.Errorf("CRC validation failed: %w", err)
		}

		data = remaining[:byteCount]

	case FuncWriteSingleCoil, FuncWriteSingleRegister:
		// 写单个值响应：地址(2) + 值(2) + CRC(2)
		remaining := make([]byte, 6)
		if _, err := io.ReadFull(c.port, remaining); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// 验证CRC
		fullFrame := make([]byte, 8)
		fullFrame[0] = slaveID
		fullFrame[1] = byte(functionCode)
		copy(fullFrame[2:], remaining)

		if _, err := ParseRTUFrame(fullFrame); err != nil {
			return nil, fmt.Errorf("CRC validation failed: %w", err)
		}

		data = remaining[:4]

	case FuncWriteMultipleCoils, FuncWriteMultipleRegisters:
		// 写多个值响应：地址(2) + 数量(2) + CRC(2)
		remaining := make([]byte, 6)
		if _, err := io.ReadFull(c.port, remaining); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// 验证CRC
		fullFrame := make([]byte, 8)
		fullFrame[0] = slaveID
		fullFrame[1] = byte(functionCode)
		copy(fullFrame[2:], remaining)

		if _, err := ParseRTUFrame(fullFrame); err != nil {
			return nil, fmt.Errorf("CRC validation failed: %w", err)
		}

		data = remaining[:4]

	default:
		return nil, fmt.Errorf("unsupported function code: 0x%02X", functionCode)
	}

	return &PDU{
		FunctionCode: functionCode,
		Data:         data,
	}, nil
}

// ReadCoils 读取线圈状态（功能码 0x01）
func (c *RTUClient) ReadCoils(address, quantity uint16) ([]byte, error) {
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

	return response.Data, nil
}

// ReadDiscreteInputs 读取离散输入（功能码 0x02）
func (c *RTUClient) ReadDiscreteInputs(address, quantity uint16) ([]byte, error) {
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

	return response.Data, nil
}

// ReadHoldingRegisters 读取保持寄存器（功能码 0x03）
func (c *RTUClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
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

	return response.Data, nil
}

// ReadInputRegisters 读取输入寄存器（功能码 0x04）
func (c *RTUClient) ReadInputRegisters(address, quantity uint16) ([]byte, error) {
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

	return response.Data, nil
}

// WriteSingleCoil 写单个线圈（功能码 0x05）
func (c *RTUClient) WriteSingleCoil(address uint16, value bool) error {
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
func (c *RTUClient) WriteSingleRegister(address uint16, value uint16) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], value)
	pdu := NewPDU(FuncWriteSingleRegister, data)

	_, err := c.sendRequest(pdu)
	return err
}

// WriteMultipleCoils 写多个线圈（功能码 0x0F）
func (c *RTUClient) WriteMultipleCoils(address uint16, values []byte) error {
	// 安全转换：确保长度不会导致溢出
	if len(values) > math.MaxUint16/8 {
		return fmt.Errorf("values length too large")
	}
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
func (c *RTUClient) WriteMultipleRegisters(address uint16, values []byte) error {
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

// CalculateFrameTimeout 计算帧超时时间
// 根据波特率和数据位数计算接收完整帧所需的最小时间
func CalculateFrameTimeout(baudRate int, dataBits, stopBits int, parity string) time.Duration {
	// 计算每个字符的位数
	charBits := 1 + dataBits + stopBits // 起始位 + 数据位 + 停止位
	if parity != "none" {
		charBits++ // 校验位
	}

	// 计算每个字符的时间（微秒）
	charTime := time.Duration(1000000*charBits/baudRate) * time.Microsecond

	// RTU帧最大长度为256字节，加上安全余量
	// 帧超时 = 3.5个字符时间
	frameTimeout := charTime * 4

	// 最小超时为10ms
	if frameTimeout < 10*time.Millisecond {
		frameTimeout = 10 * time.Millisecond
	}

	return frameTimeout
}
