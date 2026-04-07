package modbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Protocol Modbus协议类型
type Protocol string

const (
	ProtocolTCP   Protocol = "tcp"
	ProtocolRTU   Protocol = "rtu"
	ProtocolASCII Protocol = "ascii"
)

// Config Modbus配置
type Config struct {
	Protocol      Protocol
	Host          string
	Port          int
	SlaveID       byte
	Timeout       time.Duration
	RetryCount    int
	RetryInterval time.Duration
}

// SerialPortConfig 串口配置（用于RTU和ASCII）
type SerialPortConfig struct {
	PortName string
	BaudRate int
	DataBits int
	Parity   string // "none", "odd", "even"
	StopBits int
}

// Master Modbus主站
type Master struct {
	config        Config
	serialConfig  SerialPortConfig
	client        ModbusClient
	converter     *Converter
}

// ModbusClient Modbus客户端接口
type ModbusClient interface {
	Connect() error
	Disconnect() error
	ReadCoils(address, quantity uint16) ([]byte, error)
	ReadDiscreteInputs(address, quantity uint16) ([]byte, error)
	ReadHoldingRegisters(address, quantity uint16) ([]byte, error)
	ReadInputRegisters(address, quantity uint16) ([]byte, error)
	WriteSingleCoil(address uint16, value bool) error
	WriteSingleRegister(address uint16, value uint16) error
	WriteMultipleCoils(address uint16, values []byte) error
	WriteMultipleRegisters(address uint16, values []byte) error
}

// NewMaster 创建Modbus主站
func NewMaster(cfg Config) *Master {
	return &Master{
		config:    cfg,
		converter: NewConverter(BigEndian, HighWordFirst),
	}
}

// NewMasterWithSerial 创建带串口配置的Modbus主站
func NewMasterWithSerial(cfg Config, serialCfg SerialPortConfig) *Master {
	return &Master{
		config:       cfg,
		serialConfig: serialCfg,
		converter:    NewConverter(BigEndian, HighWordFirst),
	}
}

// Connect 连接到Modbus从站
func (m *Master) Connect(ctx context.Context) error {
	switch m.config.Protocol {
	case ProtocolTCP:
		m.client = NewTCPClient(m.config)
	case ProtocolRTU:
		m.client = NewRTUClientWithSerial(m.config, SerialConfig{
			PortName: m.serialConfig.PortName,
			BaudRate: m.serialConfig.BaudRate,
			DataBits: m.serialConfig.DataBits,
			Parity:   m.serialConfig.Parity,
			StopBits: m.serialConfig.StopBits,
		})
	case ProtocolASCII:
		m.client = NewASCIIClientWithSerial(m.config, SerialConfig{
			PortName: m.serialConfig.PortName,
			BaudRate: m.serialConfig.BaudRate,
			DataBits: m.serialConfig.DataBits,
			Parity:   m.serialConfig.Parity,
			StopBits: m.serialConfig.StopBits,
		})
	default:
		return fmt.Errorf("unsupported protocol: %s", m.config.Protocol)
	}

	return m.client.Connect()
}

// Disconnect 断开连接
func (m *Master) Disconnect() error {
	if m.client != nil {
		return m.client.Disconnect()
	}
	return nil
}

// SetConverter 设置数据转换器
func (m *Master) SetConverter(byteOrder ByteOrder, wordOrder WordOrder) {
	m.converter = NewConverter(byteOrder, wordOrder)
}

// GetConverter 获取数据转换器
func (m *Master) GetConverter() *Converter {
	return m.converter
}

// ========== 线圈操作 ==========

// ReadCoils 读取线圈状态
func (m *Master) ReadCoils(address, quantity uint16) ([]bool, error) {
	data, err := m.client.ReadCoils(address, quantity)
	if err != nil {
		return nil, err
	}

	result := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		result[i] = (data[i/8]>>(i%8))&0x01 == 1
	}
	return result, nil
}

// WriteSingleCoil 写单个线圈
func (m *Master) WriteSingleCoil(address uint16, value bool) error {
	return m.client.WriteSingleCoil(address, value)
}

// WriteMultipleCoils 写多个线圈
func (m *Master) WriteMultipleCoils(address uint16, values []bool) error {
	// 将bool数组转换为字节数组
	byteCount := (len(values) + 7) / 8
	data := make([]byte, byteCount)

	for i, v := range values {
		if v {
			data[i/8] |= 1 << (i % 8)
		}
	}

	return m.client.WriteMultipleCoils(address, data)
}

// ========== 离散输入操作 ==========

// ReadDiscreteInputs 读取离散输入
func (m *Master) ReadDiscreteInputs(address, quantity uint16) ([]bool, error) {
	data, err := m.client.ReadDiscreteInputs(address, quantity)
	if err != nil {
		return nil, err
	}

	result := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		result[i] = (data[i/8]>>(i%8))&0x01 == 1
	}
	return result, nil
}

// ========== 保持寄存器操作 ==========

// ReadHoldingRegisters 读取保持寄存器
func (m *Master) ReadHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	data, err := m.client.ReadHoldingRegisters(address, quantity)
	if err != nil {
		return nil, err
	}

	return m.converter.ConvertRegisters(data)
}

// ReadHoldingRegistersAsInt16 读取保持寄存器为int16数组
func (m *Master) ReadHoldingRegistersAsInt16(address, quantity uint16) ([]int16, error) {
	data, err := m.client.ReadHoldingRegisters(address, quantity)
	if err != nil {
		return nil, err
	}

	result := make([]int16, quantity)
	for i := uint16(0); i < quantity; i++ {
		val, err := m.converter.ConvertToInt16(data[i*2 : i*2+2])
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// ReadHoldingRegistersAsFloat32 读取保持寄存器为float32数组
func (m *Master) ReadHoldingRegistersAsFloat32(address, quantity uint16) ([]float32, error) {
	// 每个float32需要2个寄存器
	data, err := m.client.ReadHoldingRegisters(address, quantity*2)
	if err != nil {
		return nil, err
	}

	result := make([]float32, quantity)
	for i := uint16(0); i < quantity; i++ {
		val, err := m.converter.ConvertToFloat32(data[i*4 : i*4+4])
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// WriteSingleRegister 写单个寄存器
func (m *Master) WriteSingleRegister(address uint16, value uint16) error {
	return m.client.WriteSingleRegister(address, value)
}

// WriteMultipleRegisters 写多个寄存器
func (m *Master) WriteMultipleRegisters(address uint16, values []uint16) error {
	data := RegistersToBytes(values, BigEndian)
	return m.client.WriteMultipleRegisters(address, data)
}

// WriteMultipleRegistersFromInt16 写多个int16寄存器
func (m *Master) WriteMultipleRegistersFromInt16(address uint16, values []int16) error {
	data := make([]byte, len(values)*2)
	for i, v := range values {
		copy(data[i*2:], Int16ToBytes(v, BigEndian))
	}
	return m.client.WriteMultipleRegisters(address, data)
}

// WriteMultipleRegistersFromFloat32 写多个float32寄存器
func (m *Master) WriteMultipleRegistersFromFloat32(address uint16, values []float32) error {
	data := make([]byte, len(values)*4)
	for i, v := range values {
		copy(data[i*4:], Float32ToBytes(v, BigEndian, HighWordFirst))
	}
	return m.client.WriteMultipleRegisters(address, data)
}

// ========== 输入寄存器操作 ==========

// ReadInputRegisters 读取输入寄存器
func (m *Master) ReadInputRegisters(address, quantity uint16) ([]uint16, error) {
	data, err := m.client.ReadInputRegisters(address, quantity)
	if err != nil {
		return nil, err
	}

	return m.converter.ConvertRegisters(data)
}

// ReadInputRegistersAsInt16 读取输入寄存器为int16数组
func (m *Master) ReadInputRegistersAsInt16(address, quantity uint16) ([]int16, error) {
	data, err := m.client.ReadInputRegisters(address, quantity)
	if err != nil {
		return nil, err
	}

	result := make([]int16, quantity)
	for i := uint16(0); i < quantity; i++ {
		val, err := m.converter.ConvertToInt16(data[i*2 : i*2+2])
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// ReadInputRegistersAsFloat32 读取输入寄存器为float32数组
func (m *Master) ReadInputRegistersAsFloat32(address, quantity uint16) ([]float32, error) {
	data, err := m.client.ReadInputRegisters(address, quantity*2)
	if err != nil {
		return nil, err
	}

	result := make([]float32, quantity)
	for i := uint16(0); i < quantity; i++ {
		val, err := m.converter.ConvertToFloat32(data[i*4 : i*4+4])
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// ========== 批量操作 ==========

// BatchRead 批量读取多个地址范围
func (m *Master) BatchRead(requests []ReadRequest) ([]ReadResponse, error) {
	responses := make([]ReadResponse, len(requests))

	for i, req := range requests {
		var data interface{}
		var err error

		switch req.Type {
		case DataTypeCoil:
			data, err = m.ReadCoils(req.Address, req.Quantity)
		case DataTypeDiscreteInput:
			data, err = m.ReadDiscreteInputs(req.Address, req.Quantity)
		case DataTypeHoldingRegister:
			data, err = m.ReadHoldingRegisters(req.Address, req.Quantity)
		case DataTypeInputRegister:
			data, err = m.ReadInputRegisters(req.Address, req.Quantity)
		default:
			err = fmt.Errorf("unknown data type: %d", req.Type)
		}

		responses[i] = ReadResponse{
			Request: req,
			Data:    data,
			Error:   err,
		}
	}

	return responses, nil
}

// BatchWrite 批量写入多个地址
func (m *Master) BatchWrite(requests []WriteRequest) error {
	for _, req := range requests {
		var err error

		switch req.Type {
		case DataTypeCoil:
			if values, ok := req.Data.([]bool); ok {
				if len(values) == 1 {
					err = m.WriteSingleCoil(req.Address, values[0])
				} else {
					err = m.WriteMultipleCoils(req.Address, values)
				}
			} else {
				err = fmt.Errorf("invalid data type for coil write")
			}
		case DataTypeHoldingRegister:
			switch v := req.Data.(type) {
			case uint16:
				err = m.WriteSingleRegister(req.Address, v)
			case []uint16:
				err = m.WriteMultipleRegisters(req.Address, v)
			case int16:
				// 安全转换：int16到uint16的位级转换
				// 注意：这会保留位模式，负数会变成大于32767的值
				err = m.WriteSingleRegister(req.Address, uint16(v))
			case []int16:
				err = m.WriteMultipleRegistersFromInt16(req.Address, v)
			case float32:
				data := Float32ToBytes(v, BigEndian, HighWordFirst)
				err = m.client.WriteMultipleRegisters(req.Address, data)
			case []float32:
				err = m.WriteMultipleRegistersFromFloat32(req.Address, v)
			default:
				err = fmt.Errorf("invalid data type for register write")
			}
		default:
			err = fmt.Errorf("unknown data type: %d", req.Type)
		}

		if err != nil {
			return fmt.Errorf("write failed at address %d: %w", req.Address, err)
		}
	}

	return nil
}

// ========== 数据类型定义 ==========

// DataType 数据类型
type DataType int

const (
	DataTypeCoil            DataType = iota
	DataTypeDiscreteInput   DataType = iota
	DataTypeHoldingRegister DataType = iota
	DataTypeInputRegister   DataType = iota
)

// ReadRequest 读请求
type ReadRequest struct {
	Address  uint16
	Quantity uint16
	Type     DataType
}

// ReadResponse 读响应
type ReadResponse struct {
	Request ReadRequest
	Data    interface{}
	Error   error
}

// WriteRequest 写请求
type WriteRequest struct {
	Address uint16
	Type    DataType
	Data    interface{}
}

// ========== 寄存器值 ==========

// RegisterValue 寄存器值
type RegisterValue struct {
	ID        string
	Address   uint16
	Value     uint16
	RawValue  []byte
	Timestamp time.Time
}

// NewRegisterValue 创建寄存器值
func NewRegisterValue(address uint16, value uint16) *RegisterValue {
	return &RegisterValue{
		ID:        uuid.New().String(),
		Address:   address,
		Value:     value,
		RawValue:  []byte{byte(value >> 8), byte(value)},
		Timestamp: time.Now(),
	}
}

// ToFloat32 转换为float32
func (r *RegisterValue) ToFloat32() float32 {
	return float32(r.Value)
}

// ToInt16 转换为int16
func (r *RegisterValue) ToInt16() int16 {
	// 安全转换：uint16到int16的位级转换
	return int16(r.Value)
}

// ToBool 转换为bool
func (r *RegisterValue) ToBool() bool {
	return r.Value != 0
}

// ========== 轮询器 ==========

// Poller Modbus轮询器
type Poller struct {
	master    *Master
	interval  time.Duration
	requests  []ReadRequest
	responses chan ReadResponse
	stopChan  chan struct{}
	running   bool
}

// NewPoller 创建轮询器
func NewPoller(master *Master, interval time.Duration) *Poller {
	return &Poller{
		master:    master,
		interval:  interval,
		responses: make(chan ReadResponse, 100),
		stopChan:  make(chan struct{}),
	}
}

// AddRequest 添加轮询请求
func (p *Poller) AddRequest(req ReadRequest) {
	p.requests = append(p.requests, req)
}

// SetRequests 设置轮询请求
func (p *Poller) SetRequests(requests []ReadRequest) {
	p.requests = requests
}

// Start 启动轮询
func (p *Poller) Start() {
	if p.running {
		return
	}

	p.running = true
	go p.pollLoop()
}

// Stop 停止轮询
func (p *Poller) Stop() {
	if !p.running {
		return
	}

	p.running = false
	p.stopChan <- struct{}{}
}

// Responses 获取响应通道
func (p *Poller) Responses() <-chan ReadResponse {
	return p.responses
}

func (p *Poller) pollLoop() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			responses, err := p.master.BatchRead(p.requests)
			if err != nil {
				continue
			}

			for _, resp := range responses {
				select {
				case p.responses <- resp:
				default:
					// 通道已满，丢弃旧数据
				}
			}
		}
	}
}
