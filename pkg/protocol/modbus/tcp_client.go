package modbus

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// TCPClient Modbus TCP客户端
type TCPClient struct {
	config        Config
	conn          net.Conn
	transactionID uint32
	mu            sync.Mutex
	connected     bool
	reader        *bufio.Reader
	writer        *bufio.Writer
}

// NewTCPClient 创建TCP客户端
func NewTCPClient(config Config) *TCPClient {
	return &TCPClient{
		config: config,
	}
}

// Connect 连接到Modbus TCP服务器
func (c *TCPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	dialer := &net.Dialer{
		Timeout: c.config.Timeout,
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	// 设置TCP连接选项
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	c.connected = true

	return nil
}

// Disconnect 断开连接
func (c *TCPClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.connected = false
		return err
	}

	return nil
}

// IsConnected 检查连接状态
func (c *TCPClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// nextTransactionID 获取下一个事务ID
func (c *TCPClient) nextTransactionID() uint16 {
	id := atomic.AddUint32(&c.transactionID, 1)
	// 安全转换：使用取模确保在uint16范围内
	return uint16(id % 65536)
}

// sendRequest 发送请求并接收响应
func (c *TCPClient) sendRequest(pdu *PDU) (*PDU, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, errors.New("not connected")
	}

	// 设置读写超时
	if c.config.Timeout > 0 {
		c.conn.SetDeadline(time.Now().Add(c.config.Timeout))
	}

	// 构建请求
	transactionID := c.nextTransactionID()
	request := c.buildTCPFrame(transactionID, c.config.SlaveID, pdu)

	// 发送请求
	if _, err := c.writer.Write(request); err != nil {
		c.connected = false
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if err := c.writer.Flush(); err != nil {
		c.connected = false
		return nil, fmt.Errorf("failed to flush request: %w", err)
	}

	// 读取响应头
	header, err := c.readBytes(7)
	if err != nil {
		c.connected = false
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	// 解析MBAP头
	mbap, err := ParseMBAPHeader(header)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MBAP header: %w", err)
	}

	// 验证事务ID
	if mbap.TransactionID != transactionID {
		return nil, fmt.Errorf("transaction ID mismatch: expected %d, got %d",
			transactionID, mbap.TransactionID)
	}

	// 验证协议ID
	if mbap.ProtocolID != 0 {
		return nil, fmt.Errorf("invalid protocol ID: %d", mbap.ProtocolID)
	}

	// 读取响应数据
	responseData, err := c.readBytes(int(mbap.Length - 1))
	if err != nil {
		c.connected = false
		return nil, fmt.Errorf("failed to read response data: %w", err)
	}

	// 构建响应PDU
	responsePDU := &PDU{
		FunctionCode: FunctionCode(responseData[0]),
		Data:         responseData[1:],
	}

	// 检查异常响应
	if responsePDU.IsException() {
		return nil, NewModbusError(pdu.FunctionCode, responsePDU.GetExceptionCode())
	}

	return responsePDU, nil
}

// buildTCPFrame 构建TCP帧
func (c *TCPClient) buildTCPFrame(transactionID uint16, unitID byte, pdu *PDU) []byte {
	pduBytes := pdu.Bytes()
	length := uint16(len(pduBytes) + 1) // +1 for unit ID

	// MBAP Header (7 bytes) + PDU
	frame := make([]byte, 7+len(pduBytes))

	// Transaction ID
	binary.BigEndian.PutUint16(frame[0:2], transactionID)
	// Protocol ID (0 for Modbus)
	binary.BigEndian.PutUint16(frame[2:4], 0)
	// Length
	binary.BigEndian.PutUint16(frame[4:6], length)
	// Unit ID
	frame[6] = unitID
	// PDU
	copy(frame[7:], pduBytes)

	return frame
}

// readBytes 读取指定字节数
func (c *TCPClient) readBytes(n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := io.ReadFull(c.reader, data)
	return data, err
}

// ReadCoils 读取线圈状态（功能码 0x01）
func (c *TCPClient) ReadCoils(address, quantity uint16) ([]byte, error) {
	if quantity < 1 || quantity > MaxCoilsReadQuantity {
		return nil, fmt.Errorf("quantity must be between 1 and %d", MaxCoilsReadQuantity)
	}

	// 构建请求PDU
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)
	pdu := NewPDU(FuncReadCoils, data)

	// 发送请求
	response, err := c.sendRequest(pdu)
	if err != nil {
		return nil, err
	}

	// 解析响应
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
func (c *TCPClient) ReadDiscreteInputs(address, quantity uint16) ([]byte, error) {
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
func (c *TCPClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
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
func (c *TCPClient) ReadInputRegisters(address, quantity uint16) ([]byte, error) {
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
func (c *TCPClient) WriteSingleCoil(address uint16, value bool) error {
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
func (c *TCPClient) WriteSingleRegister(address uint16, value uint16) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], value)
	pdu := NewPDU(FuncWriteSingleRegister, data)

	_, err := c.sendRequest(pdu)
	return err
}

// WriteMultipleCoils 写多个线圈（功能码 0x0F）
func (c *TCPClient) WriteMultipleCoils(address uint16, values []byte) error {
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
func (c *TCPClient) WriteMultipleRegisters(address uint16, values []byte) error {
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

// ReadWriteMultipleRegisters 读写多个寄存器（功能码 0x17）
func (c *TCPClient) ReadWriteMultipleRegisters(readAddress, readQuantity uint16,
	writeAddress, writeQuantity uint16, values []byte) ([]byte, error) {

	if readQuantity < 1 || readQuantity > MaxReadWriteRegistersQuantity {
		return nil, fmt.Errorf("read quantity must be between 1 and %d", MaxReadWriteRegistersQuantity)
	}

	if writeQuantity < 1 || writeQuantity > MaxReadWriteRegistersQuantity {
		return nil, fmt.Errorf("write quantity must be between 1 and %d", MaxReadWriteRegistersQuantity)
	}

	byteCount := byte(len(values))
	data := make([]byte, 9+byteCount)
	binary.BigEndian.PutUint16(data[0:2], readAddress)
	binary.BigEndian.PutUint16(data[2:4], readQuantity)
	binary.BigEndian.PutUint16(data[4:6], writeAddress)
	binary.BigEndian.PutUint16(data[6:8], writeQuantity)
	data[8] = byteCount
	copy(data[9:], values)

	pdu := NewPDU(FuncReadWriteMultipleRegisters, data)

	response, err := c.sendRequest(pdu)
	if err != nil {
		return nil, err
	}

	if len(response.Data) < 1 {
		return nil, errors.New("invalid response: no data")
	}

	respByteCount := int(response.Data[0])
	if len(response.Data) < respByteCount+1 {
		return nil, errors.New("invalid response: insufficient data")
	}

	return response.Data[1 : respByteCount+1], nil
}

// MaskWriteRegister 掩码写寄存器（功能码 0x16）
func (c *TCPClient) MaskWriteRegister(address, andMask, orMask uint16) error {
	data := make([]byte, 6)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], andMask)
	binary.BigEndian.PutUint16(data[4:6], orMask)

	pdu := NewPDU(FuncMaskWriteRegister, data)

	_, err := c.sendRequest(pdu)
	return err
}

// TCPClientPool TCP客户端连接池
type TCPClientPool struct {
	config    Config
	pool      chan *TCPClient
	mu        sync.Mutex
	created   int
	maxSize   int
}

// NewTCPClientPool 创建TCP客户端连接池
func NewTCPClientPool(config Config, maxSize int) *TCPClientPool {
	return &TCPClientPool{
		config:  config,
		pool:    make(chan *TCPClient, maxSize),
		maxSize: maxSize,
	}
}

// Get 从连接池获取客户端
func (p *TCPClientPool) Get() (*TCPClient, error) {
	select {
	case client := <-p.pool:
		if client.IsConnected() {
			return client, nil
		}
		// 连接已断开，创建新连接
		_ = client.Disconnect()
	default:
	}

	p.mu.Lock()
	if p.created < p.maxSize {
		p.created++
		p.mu.Unlock()
		client := NewTCPClient(p.config)
		if err := client.Connect(); err != nil {
			p.mu.Lock()
			p.created--
			p.mu.Unlock()
			return nil, err
		}
		return client, nil
	}
	p.mu.Unlock()

	// 等待可用连接
	select {
	case client := <-p.pool:
		if client.IsConnected() {
			return client, nil
		}
		_ = client.Disconnect()
		if err := client.Connect(); err != nil {
			return nil, err
		}
		return client, nil
	}
}

// Put 将客户端放回连接池
func (p *TCPClientPool) Put(client *TCPClient) {
	select {
	case p.pool <- client:
		// 成功放回连接池
	default:
		// 连接池已满，关闭连接
		_ = client.Disconnect()
		p.mu.Lock()
		p.created--
		p.mu.Unlock()
	}
}

// Close 关闭连接池
func (p *TCPClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.pool)
	for client := range p.pool {
		_ = client.Disconnect()
	}
	p.created = 0
}
