package modbus

import (
	"context"
	"encoding/binary"
	"io"
	"math"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockModbusClient struct {
	coils       []byte
	discrete    []byte
	holdRegs    []byte
	inputRegs   []byte
	connected   bool
	connectErr  error
	disconnErr  error
}

func TestTCPClient_WithLocalServer(t *testing.T) {
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
			request := buf[:n]
			if len(request) < 7 {
				return
			}
			header, _ := ParseMBAPHeader(request[:7])
			pduData := request[7:]
			fc := FunctionCode(pduData[0])
			if fc == FuncReadHoldingRegisters {
				respData := []byte{0x02, 0x00, 0x64}
				respPDU := append([]byte{byte(fc)}, respData...)
				respFrame := make([]byte, 7+len(respPDU))
				binary.BigEndian.PutUint16(respFrame[0:2], header.TransactionID)
				binary.BigEndian.PutUint16(respFrame[2:4], 0)
				binary.BigEndian.PutUint16(respFrame[4:6], uint16(len(respPDU)+1))
				respFrame[6] = header.UnitID
				copy(respFrame[7:], respPDU)
				conn.Write(respFrame)
			} else if fc == FuncWriteSingleCoil {
				respFrame := make([]byte, len(request))
				copy(respFrame, request)
				conn.Write(respFrame)
			} else if fc == FuncWriteSingleRegister {
				respFrame := make([]byte, len(request))
				copy(respFrame, request)
				conn.Write(respFrame)
			} else if fc == FuncWriteMultipleCoils {
				respData := pduData[:4]
				respPDU := append([]byte{byte(fc)}, respData...)
				respFrame := make([]byte, 7+len(respPDU))
				binary.BigEndian.PutUint16(respFrame[0:2], header.TransactionID)
				binary.BigEndian.PutUint16(respFrame[2:4], 0)
				binary.BigEndian.PutUint16(respFrame[4:6], uint16(len(respPDU)+1))
				respFrame[6] = header.UnitID
				copy(respFrame[7:], respPDU)
				conn.Write(respFrame)
			} else if fc == FuncWriteMultipleRegisters {
				respData := pduData[:4]
				respPDU := append([]byte{byte(fc)}, respData...)
				respFrame := make([]byte, 7+len(respPDU))
				binary.BigEndian.PutUint16(respFrame[0:2], header.TransactionID)
				binary.BigEndian.PutUint16(respFrame[2:4], 0)
				binary.BigEndian.PutUint16(respFrame[4:6], uint16(len(respPDU)+1))
				respFrame[6] = header.UnitID
				copy(respFrame[7:], respPDU)
				conn.Write(respFrame)
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	cfg := Config{Host: "127.0.0.1", Port: addr.Port, SlaveID: 1, Timeout: 5 * time.Second}
	client := NewTCPClient(cfg)

	err = client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	data, err := client.ReadHoldingRegisters(0, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x64}, data)

	err = client.WriteSingleCoil(0, true)
	assert.NoError(t, err)

	err = client.WriteSingleRegister(0, 100)
	assert.NoError(t, err)

	err = client.WriteMultipleCoils(0, []byte{0xFF})
	assert.NoError(t, err)

	err = client.WriteMultipleRegisters(0, []byte{0x00, 0x64})
	assert.NoError(t, err)
}

func TestTCPClient_ReadCoils_WithLocalServer(t *testing.T) {
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
			request := buf[:n]
			if len(request) < 7 {
				return
			}
			header, _ := ParseMBAPHeader(request[:7])
			pduData := request[7:]
			fc := FunctionCode(pduData[0])
			if fc == FuncReadCoils {
				respData := []byte{0x01, 0x05}
				respPDU := append([]byte{byte(fc)}, respData...)
				respFrame := make([]byte, 7+len(respPDU))
				binary.BigEndian.PutUint16(respFrame[0:2], header.TransactionID)
				binary.BigEndian.PutUint16(respFrame[2:4], 0)
				binary.BigEndian.PutUint16(respFrame[4:6], uint16(len(respPDU)+1))
				respFrame[6] = header.UnitID
				copy(respFrame[7:], respPDU)
				conn.Write(respFrame)
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	cfg := Config{Host: "127.0.0.1", Port: addr.Port, SlaveID: 1, Timeout: 5 * time.Second}
	client := NewTCPClient(cfg)

	err = client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	data, err := client.ReadCoils(0, 5)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x05}, data)
}

func TestTCPClient_ReadDiscreteInputs_WithLocalServer(t *testing.T) {
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
			request := buf[:n]
			if len(request) < 7 {
				return
			}
			header, _ := ParseMBAPHeader(request[:7])
			pduData := request[7:]
			fc := FunctionCode(pduData[0])
			if fc == FuncReadDiscreteInputs {
				respData := []byte{0x01, 0x03}
				respPDU := append([]byte{byte(fc)}, respData...)
				respFrame := make([]byte, 7+len(respPDU))
				binary.BigEndian.PutUint16(respFrame[0:2], header.TransactionID)
				binary.BigEndian.PutUint16(respFrame[2:4], 0)
				binary.BigEndian.PutUint16(respFrame[4:6], uint16(len(respPDU)+1))
				respFrame[6] = header.UnitID
				copy(respFrame[7:], respPDU)
				conn.Write(respFrame)
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	cfg := Config{Host: "127.0.0.1", Port: addr.Port, SlaveID: 1, Timeout: 5 * time.Second}
	client := NewTCPClient(cfg)

	err = client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	data, err := client.ReadDiscreteInputs(0, 3)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x03}, data)
}

func TestTCPClient_ReadInputRegisters_WithLocalServer(t *testing.T) {
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
			request := buf[:n]
			if len(request) < 7 {
				return
			}
			header, _ := ParseMBAPHeader(request[:7])
			pduData := request[7:]
			fc := FunctionCode(pduData[0])
			if fc == FuncReadInputRegisters {
				respData := []byte{0x02, 0x00, 0xC8}
				respPDU := append([]byte{byte(fc)}, respData...)
				respFrame := make([]byte, 7+len(respPDU))
				binary.BigEndian.PutUint16(respFrame[0:2], header.TransactionID)
				binary.BigEndian.PutUint16(respFrame[2:4], 0)
				binary.BigEndian.PutUint16(respFrame[4:6], uint16(len(respPDU)+1))
				respFrame[6] = header.UnitID
				copy(respFrame[7:], respPDU)
				conn.Write(respFrame)
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	cfg := Config{Host: "127.0.0.1", Port: addr.Port, SlaveID: 1, Timeout: 5 * time.Second}
	client := NewTCPClient(cfg)

	err = client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	data, err := client.ReadInputRegisters(0, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0xC8}, data)
}

func newMockClient() *mockModbusClient {
	return &mockModbusClient{
		coils:     []byte{0x05},
		discrete:  []byte{0x03},
		holdRegs:  []byte{0x00, 0x01, 0x00, 0x02},
		inputRegs: []byte{0x00, 0x03, 0x00, 0x04},
		connected: false,
	}
}

func (m *mockModbusClient) Connect() error {
	if m.connectErr != nil {
		return m.connectErr
	}
	m.connected = true
	return nil
}

func (m *mockModbusClient) Disconnect() error {
	if m.disconnErr != nil {
		return m.disconnErr
	}
	m.connected = false
	return nil
}

func (m *mockModbusClient) ReadCoils(address, quantity uint16) ([]byte, error) {
	return m.coils, nil
}

func (m *mockModbusClient) ReadDiscreteInputs(address, quantity uint16) ([]byte, error) {
	return m.discrete, nil
}

func (m *mockModbusClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
	return m.holdRegs, nil
}

func (m *mockModbusClient) ReadInputRegisters(address, quantity uint16) ([]byte, error) {
	return m.inputRegs, nil
}

func (m *mockModbusClient) WriteSingleCoil(address uint16, value bool) error {
	return nil
}

func (m *mockModbusClient) WriteSingleRegister(address uint16, value uint16) error {
	return nil
}

func (m *mockModbusClient) WriteMultipleCoils(address uint16, values []byte) error {
	return nil
}

func (m *mockModbusClient) WriteMultipleRegisters(address uint16, values []byte) error {
	return nil
}

func TestMaster_New(t *testing.T) {
	cfg := Config{
		Protocol: ProtocolTCP,
		Host:     "127.0.0.1",
		Port:     502,
		SlaveID:  1,
		Timeout:  5 * time.Second,
	}
	master := NewMaster(cfg)
	assert.NotNil(t, master)
	assert.NotNil(t, master.converter)
}

func TestMaster_NewWithSerial(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU}
	serialCfg := SerialPortConfig{
		PortName: "/dev/ttyUSB0",
		BaudRate: 9600,
		DataBits: 8,
		Parity:   "none",
		StopBits: 1,
	}
	master := NewMasterWithSerial(cfg, serialCfg)
	assert.NotNil(t, master)
	assert.Equal(t, "/dev/ttyUSB0", master.serialConfig.PortName)
}

func TestMaster_Connect_UnsupportedProtocol(t *testing.T) {
	cfg := Config{Protocol: Protocol("unsupported")}
	master := NewMaster(cfg)
	err := master.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol")
}

func TestMaster_Disconnect_NoClient(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	err := master.Disconnect()
	assert.NoError(t, err)
}

func TestMaster_SetConverter(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.SetConverter(LittleEndian, LowWordFirst)
	conv := master.GetConverter()
	assert.NotNil(t, conv)
	assert.Equal(t, LittleEndian, conv.byteOrder)
	assert.Equal(t, LowWordFirst, conv.wordOrder)
}

func TestMaster_ReadCoils(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	result, err := master.ReadCoils(0, 5)
	assert.NoError(t, err)
	assert.Len(t, result, 5)
	assert.True(t, result[0])
	assert.True(t, result[2])
}

func TestMaster_ReadDiscreteInputs(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	result, err := master.ReadDiscreteInputs(0, 3)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.True(t, result[0])
	assert.True(t, result[1])
}

func TestMaster_ReadHoldingRegisters(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	result, err := master.ReadHoldingRegisters(0, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestMaster_ReadInputRegisters(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	result, err := master.ReadInputRegisters(0, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestMaster_WriteSingleCoil(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	err := master.WriteSingleCoil(0, true)
	assert.NoError(t, err)
}

func TestMaster_WriteMultipleCoils(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	err := master.WriteMultipleCoils(0, []bool{true, false, true})
	assert.NoError(t, err)
}

func TestMaster_WriteSingleRegister(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	err := master.WriteSingleRegister(0, 100)
	assert.NoError(t, err)
}

func TestMaster_WriteMultipleRegisters(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	err := master.WriteMultipleRegisters(0, []uint16{100, 200})
	assert.NoError(t, err)
}

func TestMaster_WriteMultipleRegistersFromInt16(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	err := master.WriteMultipleRegistersFromInt16(0, []int16{-100, 200})
	assert.NoError(t, err)
}

func TestMaster_WriteMultipleRegistersFromFloat32(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	err := master.WriteMultipleRegistersFromFloat32(0, []float32{3.14, 2.71})
	assert.NoError(t, err)
}

func TestMaster_BatchRead(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []ReadRequest{
		{Address: 0, Quantity: 5, Type: DataTypeCoil},
		{Address: 0, Quantity: 3, Type: DataTypeDiscreteInput},
		{Address: 0, Quantity: 2, Type: DataTypeHoldingRegister},
		{Address: 0, Quantity: 2, Type: DataTypeInputRegister},
	}

	responses, err := master.BatchRead(requests)
	assert.NoError(t, err)
	assert.Len(t, responses, 4)
	assert.NoError(t, responses[0].Error)
	assert.NoError(t, responses[1].Error)
	assert.NoError(t, responses[2].Error)
	assert.NoError(t, responses[3].Error)
}

func TestMaster_BatchRead_UnknownType(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []ReadRequest{
		{Address: 0, Quantity: 1, Type: DataType(99)},
	}

	responses, err := master.BatchRead(requests)
	assert.NoError(t, err)
	assert.Error(t, responses[0].Error)
}

func TestMaster_BatchWrite_Coil(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeCoil, Data: []bool{true}},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_MultipleCoils(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeCoil, Data: []bool{true, false, true}},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_InvalidCoilData(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeCoil, Data: "invalid"},
	}

	err := master.BatchWrite(requests)
	assert.Error(t, err)
}

type mockPort struct {
	readData  []byte
	writeData []byte
	readPos   int
	writePos  int
	closed    bool
	closeErr  error
}

func (m *mockPort) Read(p []byte) (n int, err error) {
	if m.readPos >= len(m.readData) {
		return 0, io.EOF
	}
	n = copy(p, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockPort) Write(p []byte) (n int, err error) {
	m.writeData = append(m.writeData, p...)
	return len(p), nil
}

func (m *mockPort) Close() error {
	m.closed = true
	return m.closeErr
}

func TestRTUClient_ReadCoils_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncReadCoils, []byte{0x01, 0x05})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadCoils(0, 5)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x05}, result)
}

func TestRTUClient_ReadDiscreteInputs_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncReadDiscreteInputs, []byte{0x01, 0x03})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadDiscreteInputs(0, 3)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x03}, result)
}

func TestRTUClient_ReadHoldingRegisters_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncReadHoldingRegisters, []byte{0x02, 0x00, 0x64})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadHoldingRegisters(0, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x64}, result)
}

func TestRTUClient_ReadInputRegisters_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncReadInputRegisters, []byte{0x02, 0x00, 0xC8})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadInputRegisters(0, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0xC8}, result)
}

func TestRTUClient_WriteSingleCoil_On(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncWriteSingleCoil, []byte{0x00, 0x00, 0xFF, 0x00})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteSingleCoil(0, true)
	require.NoError(t, err)
}

func TestRTUClient_WriteSingleCoil_Off(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncWriteSingleCoil, []byte{0x00, 0x00, 0x00, 0x00})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteSingleCoil(0, false)
	require.NoError(t, err)
}

func TestRTUClient_WriteSingleRegister_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncWriteSingleRegister, []byte{0x00, 0x00, 0x00, 0x64})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteSingleRegister(0, 100)
	require.NoError(t, err)
}

func TestRTUClient_WriteMultipleCoils_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncWriteMultipleCoils, []byte{0x00, 0x00, 0x00, 0x08})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteMultipleCoils(0, []byte{0xFF})
	require.NoError(t, err)
}

func TestRTUClient_WriteMultipleRegisters_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncWriteMultipleRegisters, []byte{0x00, 0x00, 0x00, 0x01})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteMultipleRegisters(0, []byte{0x00, 0x64})
	require.NoError(t, err)
}

func TestRTUClient_Disconnect_WithPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	port := &mockPort{}
	client.SetPort(port)
	assert.True(t, client.IsConnected())

	err := client.Disconnect()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
	assert.True(t, port.closed)
}

func TestRTUClient_ReadResponse_SlaveIDMismatch(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 2, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncReadCoils, []byte{0x01, 0x05})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slave ID mismatch")
}

func TestASCIIClient_ReadCoils_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadCoils, []byte{0x01, 0x05})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadCoils(0, 5)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x05}, result)
}

func TestASCIIClient_ReadDiscreteInputs_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadDiscreteInputs, []byte{0x01, 0x03})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadDiscreteInputs(0, 3)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x03}, result)
}

func TestASCIIClient_ReadHoldingRegisters_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadHoldingRegisters, []byte{0x02, 0x00, 0x64})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadHoldingRegisters(0, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x64}, result)
}

func TestASCIIClient_ReadInputRegisters_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadInputRegisters, []byte{0x02, 0x00, 0xC8})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	result, err := client.ReadInputRegisters(0, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0xC8}, result)
}

func TestASCIIClient_WriteSingleCoil_On(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncWriteSingleCoil, []byte{0x00, 0x00, 0xFF, 0x00})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteSingleCoil(0, true)
	require.NoError(t, err)
}

func TestASCIIClient_WriteSingleCoil_Off(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncWriteSingleCoil, []byte{0x00, 0x00, 0x00, 0x00})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteSingleCoil(0, false)
	require.NoError(t, err)
}

func TestASCIIClient_WriteSingleRegister_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncWriteSingleRegister, []byte{0x00, 0x00, 0x00, 0x64})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteSingleRegister(0, 100)
	require.NoError(t, err)
}

func TestASCIIClient_WriteMultipleCoils_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncWriteMultipleCoils, []byte{0x00, 0x00, 0x00, 0x08})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteMultipleCoils(0, []byte{0xFF})
	require.NoError(t, err)
}

func TestASCIIClient_WriteMultipleRegisters_WithMockPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncWriteMultipleRegisters, []byte{0x00, 0x00, 0x00, 0x01})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	err := client.WriteMultipleRegisters(0, []byte{0x00, 0x64})
	require.NoError(t, err)
}

func TestASCIIClient_Disconnect_WithPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	port := &mockPort{}
	client.SetPort(port)
	assert.True(t, client.IsConnected())

	err := client.Disconnect()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
	assert.True(t, port.closed)
}

func TestASCIIClient_ReadResponse_SlaveIDMismatch(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 2, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadCoils, []byte{0x01, 0x05})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 5)
	assert.Error(t, err)
}

func TestRTUClient_ReadResponse_ExceptionResponse(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	exceptionData := []byte{1, byte(FuncReadCoils | 0x80), byte(ExcIllegalDataAddress)}
	crc := CalculateCRC(exceptionData)
	fullFrame := append(exceptionData, byte(crc), byte(crc>>8))
	port := &mockPort{readData: fullFrame}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_ReadResponse_WriteError(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	port := &mockPort{readData: []byte{}}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadResponse_InvalidStart(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	port := &mockPort{readData: []byte("X0103FF000DCRLF")}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestMaster_BatchWrite_RegisterUint16(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: uint16(100)},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_RegisterUint16Slice(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: []uint16{100, 200}},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_RegisterInt16(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: int16(-100)},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_RegisterInt16Slice(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: []int16{-100, 200}},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_RegisterFloat32(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: float32(3.14)},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_RegisterFloat32Slice(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: []float32{3.14, 2.71}},
	}

	err := master.BatchWrite(requests)
	assert.NoError(t, err)
}

func TestMaster_BatchWrite_InvalidRegisterData(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeHoldingRegister, Data: "invalid"},
	}

	err := master.BatchWrite(requests)
	assert.Error(t, err)
}

func TestMaster_BatchWrite_UnknownType(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()

	requests := []WriteRequest{
		{Address: 0, Type: DataType(99), Data: uint16(100)},
	}

	err := master.BatchWrite(requests)
	assert.Error(t, err)
}

func TestRegisterValue(t *testing.T) {
	rv := NewRegisterValue(100, 0x1234)
	assert.NotNil(t, rv)
	assert.NotEmpty(t, rv.ID)
	assert.Equal(t, uint16(100), rv.Address)
	assert.Equal(t, uint16(0x1234), rv.Value)
	assert.NotEmpty(t, rv.RawValue)
	assert.False(t, rv.Timestamp.IsZero())
}

func TestRegisterValue_ToFloat32(t *testing.T) {
	rv := NewRegisterValue(0, 100)
	assert.Equal(t, float32(100), rv.ToFloat32())
}

func TestRegisterValue_ToInt16(t *testing.T) {
	rv := NewRegisterValue(0, 0xFFFF)
	assert.Equal(t, int16(-1), rv.ToInt16())
}

func TestRegisterValue_ToBool(t *testing.T) {
	rv1 := NewRegisterValue(0, 0)
	assert.False(t, rv1.ToBool())

	rv2 := NewRegisterValue(0, 1)
	assert.True(t, rv2.ToBool())
}

func TestPoller_New(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	poller := NewPoller(master, 1*time.Second)
	assert.NotNil(t, poller)
	assert.NotNil(t, poller.responses)
}

func TestPoller_AddRequest(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	poller := NewPoller(master, 1*time.Second)

	poller.AddRequest(ReadRequest{Address: 0, Quantity: 10, Type: DataTypeCoil})
	assert.Len(t, poller.requests, 1)

	poller.AddRequest(ReadRequest{Address: 10, Quantity: 5, Type: DataTypeHoldingRegister})
	assert.Len(t, poller.requests, 2)
}

func TestPoller_SetRequests(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	poller := NewPoller(master, 1*time.Second)

	requests := []ReadRequest{
		{Address: 0, Quantity: 10, Type: DataTypeCoil},
		{Address: 10, Quantity: 5, Type: DataTypeHoldingRegister},
	}
	poller.SetRequests(requests)
	assert.Len(t, poller.requests, 2)
}

func TestPoller_Responses(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	poller := NewPoller(master, 1*time.Second)

	ch := poller.Responses()
	assert.NotNil(t, ch)
}

func TestPoller_StartStop(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()
	poller := NewPoller(master, 100*time.Millisecond)
	poller.SetRequests([]ReadRequest{{Address: 0, Quantity: 1, Type: DataTypeCoil}})

	poller.Start()
	assert.True(t, poller.running)

	time.Sleep(250 * time.Millisecond)

	poller.Stop()
	assert.False(t, poller.running)
}

func TestPoller_StartAlreadyRunning(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()
	poller := NewPoller(master, 1*time.Second)

	poller.Start()
	poller.Start()
	assert.True(t, poller.running)
	poller.Stop()
}

func TestPoller_StopNotRunning(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	poller := NewPoller(master, 1*time.Second)

	poller.Stop()
	assert.False(t, poller.running)
}

func TestTCPClient_New(t *testing.T) {
	cfg := Config{
		Protocol: ProtocolTCP,
		Host:     "127.0.0.1",
		Port:     502,
		SlaveID:  1,
		Timeout:  5 * time.Second,
	}
	client := NewTCPClient(cfg)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestTCPClient_Disconnect_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestTCPClient_Connect_Failed(t *testing.T) {
	cfg := Config{
		Host:     "127.0.0.1",
		Port:     19999,
		Timeout:  1 * time.Second,
		SlaveID:  1,
	}
	client := NewTCPClient(cfg)
	err := client.Connect()
	assert.Error(t, err)
}

func TestTCPClient_ReadCoils_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestTCPClient_ReadDiscreteInputs_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadDiscreteInputs(0, 1)
	assert.Error(t, err)
}

func TestTCPClient_ReadHoldingRegisters_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadHoldingRegisters(0, 1)
	assert.Error(t, err)
}

func TestTCPClient_ReadInputRegisters_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadInputRegisters(0, 1)
	assert.Error(t, err)
}

func TestTCPClient_WriteSingleCoil_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.WriteSingleCoil(0, true)
	assert.Error(t, err)
}

func TestTCPClient_WriteSingleRegister_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.WriteSingleRegister(0, 100)
	assert.Error(t, err)
}

func TestTCPClient_WriteMultipleCoils_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.WriteMultipleCoils(0, []byte{0xFF})
	assert.Error(t, err)
}

func TestTCPClient_WriteMultipleRegisters_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.WriteMultipleRegisters(0, []byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestTCPClient_ReadCoils_InvalidQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadCoils(0, 0)
	assert.Error(t, err)
	_, err = client.ReadCoils(0, 2001)
	assert.Error(t, err)
}

func TestTCPClient_ReadDiscreteInputs_InvalidQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadDiscreteInputs(0, 0)
	assert.Error(t, err)
}

func TestTCPClient_ReadHoldingRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadHoldingRegisters(0, 0)
	assert.Error(t, err)
	_, err = client.ReadHoldingRegisters(0, 126)
	assert.Error(t, err)
}

func TestTCPClient_ReadInputRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadInputRegisters(0, 0)
	assert.Error(t, err)
}

func TestTCPClient_WriteMultipleCoils_InvalidQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.WriteMultipleCoils(0, []byte{})
	assert.Error(t, err)
}

func TestTCPClient_WriteMultipleRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.WriteMultipleRegisters(0, []byte{})
	assert.Error(t, err)
}

func TestTCPClient_buildTCPFrame(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502, SlaveID: 1}
	client := NewTCPClient(cfg)

	pdu := NewPDU(FuncReadHoldingRegisters, []byte{0x00, 0x01, 0x00, 0x0A})
	frame := client.buildTCPFrame(1, 1, pdu)

	assert.Len(t, frame, 12)
	assert.Equal(t, uint16(1), uint16(frame[0])<<8|uint16(frame[1]))
	assert.Equal(t, uint16(0), uint16(frame[2])<<8|uint16(frame[3]))
	assert.Equal(t, byte(1), frame[6])
	assert.Equal(t, byte(FuncReadHoldingRegisters), frame[7])
}

func TestTCPClient_nextTransactionID(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)

	id1 := client.nextTransactionID()
	id2 := client.nextTransactionID()
	assert.NotEqual(t, id1, id2)
	assert.Equal(t, uint16(1), id1)
	assert.Equal(t, uint16(2), id2)
}

func TestTCPClientPool_New(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	pool := NewTCPClientPool(cfg, 5)
	assert.NotNil(t, pool)
}

func TestTCPClientPool_Get_Failed(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 19999, Timeout: 1 * time.Second}
	pool := NewTCPClientPool(cfg, 5)
	_, err := pool.Get()
	assert.Error(t, err)
}

func TestTCPClientPool_Close(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	pool := NewTCPClientPool(cfg, 5)
	pool.Close()
}

func TestParseMBAPHeader_InsufficientData(t *testing.T) {
	_, err := ParseMBAPHeader([]byte{0x01, 0x02, 0x03})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestParseRTUFrame_InsufficientData(t *testing.T) {
	_, err := ParseRTUFrame([]byte{0x01, 0x02})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestParseASCIIFrame_InsufficientData(t *testing.T) {
	_, err := ParseASCIIFrame([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestParseASCIIFrame_InvalidStart(t *testing.T) {
	data := []byte{0x30, 0x30, 0x31, 0x30, 0x33, 0x46, 0x46, 0x0D, 0x0A}
	_, err := ParseASCIIFrame(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start character")
}

func TestParseASCIIFrame_InvalidEnd(t *testing.T) {
	data := []byte{':', '0', '0', 'F', 'F', '0', '0', 0x0D, 0x0A}
	_, err := ParseASCIIFrame(data)
	assert.Error(t, err)
}

func TestPDU_GetExceptionCode_NoException(t *testing.T) {
	pdu := &PDU{FunctionCode: FuncReadCoils, Data: []byte{0x01}}
	assert.Equal(t, ExceptionCode(0), pdu.GetExceptionCode())
}

func TestPDU_GetExceptionCode_EmptyData(t *testing.T) {
	pdu := &PDU{FunctionCode: FuncReadCoils | 0x80, Data: []byte{}}
	assert.Equal(t, ExceptionCode(0), pdu.GetExceptionCode())
}

func TestConverter_ConvertToUint16(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	val, err := conv.ConvertToUint16([]byte{0x00, 0x64})
	assert.NoError(t, err)
	assert.Equal(t, uint16(100), val)
}

func TestConverter_ConvertToInt16(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	val, err := conv.ConvertToInt16([]byte{0xFF, 0x9C})
	assert.NoError(t, err)
	assert.Equal(t, int16(-100), val)
}

func TestConverter_ConvertToUint32(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	val, err := conv.ConvertToUint32([]byte{0x00, 0x00, 0x01, 0x00})
	assert.NoError(t, err)
	assert.Equal(t, uint32(256), val)
}

func TestConverter_ConvertToInt32(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	val, err := conv.ConvertToInt32([]byte{0x00, 0x00, 0x01, 0x00})
	assert.NoError(t, err)
	assert.Equal(t, int32(256), val)
}

func TestConverter_ConvertToFloat32(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	bytes := Float32ToBytes(3.14, BigEndian, HighWordFirst)
	val, err := conv.ConvertToFloat32(bytes)
	assert.NoError(t, err)
	assert.InDelta(t, float32(3.14), val, 0.001)
}

func TestConverter_ConvertToFloat64(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	bytes := Float64ToBytes(3.14159, BigEndian, HighWordFirst)
	val, err := conv.ConvertToFloat64(bytes)
	assert.NoError(t, err)
	assert.InDelta(t, 3.14159, val, 0.00001)
}

func TestBytesToUint16_InsufficientData(t *testing.T) {
	_, err := BytesToUint16([]byte{0x01}, BigEndian)
	assert.Error(t, err)
}

func TestBytesToUint32_InsufficientData(t *testing.T) {
	_, err := BytesToUint32([]byte{0x01, 0x02}, BigEndian, HighWordFirst)
	assert.Error(t, err)
}

func TestBytesToUint64_InsufficientData(t *testing.T) {
	_, err := BytesToUint64([]byte{0x01, 0x02}, BigEndian, HighWordFirst)
	assert.Error(t, err)
}

func TestBytesToRegisters_OddLength(t *testing.T) {
	_, err := BytesToRegisters([]byte{0x01, 0x02, 0x03}, BigEndian)
	assert.Error(t, err)
}

func TestUint8ToBCD_Overflow(t *testing.T) {
	result := Uint8ToBCD(100)
	assert.Equal(t, byte(0xFF), result)
}

func TestUint16ToBCD_Overflow(t *testing.T) {
	result := Uint16ToBCD(10000)
	assert.Equal(t, uint16(0xFFFF), result)
}

func TestUint32ToBCD_Overflow(t *testing.T) {
	result := Uint32ToBCD(100000000)
	assert.Equal(t, uint32(0xFFFFFFFF), result)
}

func TestBCDToUint32(t *testing.T) {
	result := BCDToUint32(0x00000012)
	assert.Equal(t, uint32(12), result)
}

func TestInt32Conversion(t *testing.T) {
	tests := []struct {
		name      string
		value     int32
		byteOrder ByteOrder
		wordOrder WordOrder
	}{
		{"Positive BigEndian", 12345, BigEndian, HighWordFirst},
		{"Zero", 0, BigEndian, HighWordFirst},
		{"LittleEndian", 12345, LittleEndian, LowWordFirst},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Int32ToBytes(tt.value, tt.byteOrder, tt.wordOrder)
			result, err := BytesToInt32(bytes, tt.byteOrder, tt.wordOrder)
			require.NoError(t, err)
			assert.Equal(t, tt.value, result)
		})
	}
}

func TestInt64Conversion(t *testing.T) {
	tests := []struct {
		name      string
		value     int64
		byteOrder ByteOrder
		wordOrder WordOrder
	}{
		{"Positive BigEndian", 123456789, BigEndian, HighWordFirst},
		{"Negative BigEndian", -123456789, BigEndian, HighWordFirst},
		{"LittleEndian LowWordFirst", 123456789, LittleEndian, LowWordFirst},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := Int64ToBytes(tt.value, tt.byteOrder, tt.wordOrder)
			result, err := BytesToInt64(bytes, tt.byteOrder, tt.wordOrder)
			require.NoError(t, err)
			assert.Equal(t, tt.value, result)
		})
	}
}

func TestUint64Conversion(t *testing.T) {
	value := uint64(0x123456789ABCDEF0)
	bytes := Uint64ToBytes(value, BigEndian, HighWordFirst)
	result, err := BytesToUint64(bytes, BigEndian, HighWordFirst)
	require.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestUint64Conversion_LowWordFirst(t *testing.T) {
	value := uint64(0x123456789ABCDEF0)
	bytes := Uint64ToBytes(value, BigEndian, LowWordFirst)
	result, err := BytesToUint64(bytes, BigEndian, LowWordFirst)
	require.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestGetBit_OutOfRange(t *testing.T) {
	assert.False(t, GetBit(0xFF, 8))
}

func TestSetBit_OutOfRange(t *testing.T) {
	assert.Equal(t, byte(0x55), SetBit(0x55, 8, true))
}

func TestGetBits_PartialData(t *testing.T) {
	data := []byte{0xFF}
	bits := GetBits(data, 0, 16)
	assert.Len(t, bits, 16)
	for i := 0; i < 8; i++ {
		assert.True(t, bits[i])
	}
	for i := 8; i < 16; i++ {
		assert.False(t, bits[i])
	}
}

func TestSetBits_EmptyValues(t *testing.T) {
	data := []byte{0x00, 0x00}
	result := SetBits(data, 0, []bool{})
	assert.Equal(t, []byte{0x00, 0x00}, result)
}

func TestFunctionCode_AllCodes(t *testing.T) {
	codes := []FunctionCode{
		FuncReadCoils, FuncReadDiscreteInputs, FuncReadHoldingRegisters,
		FuncReadInputRegisters, FuncWriteSingleCoil, FuncWriteSingleRegister,
		FuncWriteMultipleCoils, FuncWriteMultipleRegisters,
		FuncReadWriteMultipleRegisters, FuncMaskWriteRegister, FuncReadFIFOQueue,
	}
	for _, code := range codes {
		s := code.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Unknown")
	}
}

func TestExceptionCode_AllCodes(t *testing.T) {
	codes := []ExceptionCode{
		ExcIllegalFunction, ExcIllegalDataAddress, ExcIllegalDataValue,
		ExcSlaveDeviceFailure, ExcAcknowledge, ExcSlaveDeviceBusy,
		ExcMemoryParityError, ExcGatewayPathUnavailable,
		ExcGatewayTargetDeviceFailedToRespond,
	}
	for _, code := range codes {
		s := code.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Unknown")
	}
}

func TestMaster_ReadHoldingRegistersAsInt16(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	mc := newMockClient()
	mc.holdRegs = []byte{0xFF, 0x9C, 0x00, 0x64}
	master.client = mc

	result, err := master.ReadHoldingRegistersAsInt16(0, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int16(-100), result[0])
	assert.Equal(t, int16(100), result[1])
}

func TestMaster_ReadInputRegistersAsInt16(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	mc := newMockClient()
	mc.inputRegs = []byte{0xFF, 0x9C, 0x00, 0x64}
	master.client = mc

	result, err := master.ReadInputRegistersAsInt16(0, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestMaster_ReadHoldingRegistersAsFloat32(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	mc := newMockClient()
	floatBytes := Float32ToBytes(3.14, BigEndian, HighWordFirst)
	mc.holdRegs = floatBytes
	master.client = mc

	result, err := master.ReadHoldingRegistersAsFloat32(0, 1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.InDelta(t, float32(3.14), result[0], 0.001)
}

func TestMaster_ReadInputRegistersAsFloat32(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	mc := newMockClient()
	floatBytes := Float32ToBytes(2.71, BigEndian, HighWordFirst)
	mc.inputRegs = floatBytes
	master.client = mc

	result, err := master.ReadInputRegistersAsFloat32(0, 1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.InDelta(t, float32(2.71), result[0], 0.001)
}

type errorMockClient struct {
	mockModbusClient
}

func (m *errorMockClient) ReadCoils(address, quantity uint16) ([]byte, error) {
	return nil, assert.AnError
}

func (m *errorMockClient) ReadDiscreteInputs(address, quantity uint16) ([]byte, error) {
	return nil, assert.AnError
}

func (m *errorMockClient) ReadHoldingRegisters(address, quantity uint16) ([]byte, error) {
	return nil, assert.AnError
}

func (m *errorMockClient) ReadInputRegisters(address, quantity uint16) ([]byte, error) {
	return nil, assert.AnError
}

func (m *errorMockClient) WriteSingleCoil(address uint16, value bool) error {
	return assert.AnError
}

func (m *errorMockClient) WriteSingleRegister(address uint16, value uint16) error {
	return assert.AnError
}

func (m *errorMockClient) WriteMultipleCoils(address uint16, values []byte) error {
	return assert.AnError
}

func (m *errorMockClient) WriteMultipleRegisters(address uint16, values []byte) error {
	return assert.AnError
}

func TestMaster_ReadCoils_ClientError(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	_, err := master.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestMaster_ReadDiscreteInputs_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	_, err := master.ReadDiscreteInputs(0, 1)
	assert.Error(t, err)
}

func TestMaster_ReadInputRegisters_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	_, err := master.ReadInputRegisters(0, 1)
	assert.Error(t, err)
}

func TestMaster_WriteSingleCoil_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	err := master.WriteSingleCoil(0, true)
	assert.Error(t, err)
}

func TestMaster_WriteMultipleCoils_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	err := master.WriteMultipleCoils(0, []bool{true})
	assert.Error(t, err)
}

func TestMaster_WriteSingleRegister_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	err := master.WriteSingleRegister(0, 100)
	assert.Error(t, err)
}

func TestMaster_BatchWrite_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}

	requests := []WriteRequest{
		{Address: 0, Type: DataTypeCoil, Data: []bool{true}},
	}

	err := master.BatchWrite(requests)
	assert.Error(t, err)
}

func TestTCPClientPool_Put_Full(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	pool := NewTCPClientPool(cfg, 1)

	client := NewTCPClient(cfg)
	pool.Put(client)
}

func TestBytesToFloat32_InsufficientData(t *testing.T) {
	_, err := BytesToFloat32([]byte{0x01, 0x02}, BigEndian, HighWordFirst)
	assert.Error(t, err)
}

func TestBytesToFloat64_InsufficientData(t *testing.T) {
	_, err := BytesToFloat64([]byte{0x01, 0x02}, BigEndian, HighWordFirst)
	assert.Error(t, err)
}

func TestBytesToInt32_Overflow(t *testing.T) {
	_, err := BytesToInt32([]byte{0xFF, 0xFF, 0xFF, 0xFF}, BigEndian, HighWordFirst)
	assert.Error(t, err)
}

func TestRTUClient_New(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestRTUClient_NewWithSerial(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	serialCfg := SerialConfig{PortName: "/dev/ttyUSB0", BaudRate: 9600}
	client := NewRTUClientWithSerial(cfg, serialCfg)
	assert.NotNil(t, client)
	assert.Equal(t, "/dev/ttyUSB0", client.serialConfig.PortName)
}

func TestRTUClient_Connect_NoPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serial port not configured")
}

func TestRTUClient_Connect_AlreadyConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	client.connected = true
	err := client.Connect()
	assert.NoError(t, err)
}

func TestRTUClient_Disconnect_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestRTUClient_ReadCoils_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_ReadCoils_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadCoils(0, 0)
	assert.Error(t, err)
	_, err = client.ReadCoils(0, 2001)
	assert.Error(t, err)
}

func TestRTUClient_ReadDiscreteInputs_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadDiscreteInputs(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_ReadDiscreteInputs_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadDiscreteInputs(0, 0)
	assert.Error(t, err)
}

func TestRTUClient_ReadHoldingRegisters_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadHoldingRegisters(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_ReadHoldingRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadHoldingRegisters(0, 0)
	assert.Error(t, err)
}

func TestRTUClient_ReadInputRegisters_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadInputRegisters(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_ReadInputRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	_, err := client.ReadInputRegisters(0, 0)
	assert.Error(t, err)
}

func TestRTUClient_WriteSingleCoil_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.WriteSingleCoil(0, true)
	assert.Error(t, err)
}

func TestRTUClient_WriteSingleRegister_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.WriteSingleRegister(0, 100)
	assert.Error(t, err)
}

func TestRTUClient_WriteMultipleCoils_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.WriteMultipleCoils(0, []byte{0xFF})
	assert.Error(t, err)
}

func TestRTUClient_WriteMultipleCoils_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.WriteMultipleCoils(0, []byte{})
	assert.Error(t, err)
}

func TestRTUClient_WriteMultipleRegisters_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.WriteMultipleRegisters(0, []byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestRTUClient_WriteMultipleRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	err := client.WriteMultipleRegisters(0, []byte{})
	assert.Error(t, err)
}

func TestRTUClient_SetPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	assert.False(t, client.IsConnected())
	client.SetPort(nil)
	assert.True(t, client.IsConnected())
}

func TestASCIIClient_New(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestASCIIClient_NewWithSerial(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	serialCfg := SerialConfig{PortName: "/dev/ttyUSB0", BaudRate: 9600}
	client := NewASCIIClientWithSerial(cfg, serialCfg)
	assert.NotNil(t, client)
	assert.Equal(t, "/dev/ttyUSB0", client.serialConfig.PortName)
}

func TestASCIIClient_Connect_NoPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serial port not configured")
}

func TestASCIIClient_Connect_AlreadyConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	client.connected = true
	err := client.Connect()
	assert.NoError(t, err)
}

func TestASCIIClient_Disconnect_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestASCIIClient_ReadCoils_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadCoils_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadCoils(0, 0)
	assert.Error(t, err)
}

func TestASCIIClient_ReadDiscreteInputs_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadDiscreteInputs(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadDiscreteInputs_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadDiscreteInputs(0, 0)
	assert.Error(t, err)
}

func TestASCIIClient_ReadHoldingRegisters_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadHoldingRegisters(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadHoldingRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadHoldingRegisters(0, 0)
	assert.Error(t, err)
}

func TestASCIIClient_ReadInputRegisters_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadInputRegisters(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadInputRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	_, err := client.ReadInputRegisters(0, 0)
	assert.Error(t, err)
}

func TestASCIIClient_WriteSingleCoil_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.WriteSingleCoil(0, true)
	assert.Error(t, err)
}

func TestASCIIClient_WriteSingleRegister_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.WriteSingleRegister(0, 100)
	assert.Error(t, err)
}

func TestASCIIClient_WriteMultipleCoils_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.WriteMultipleCoils(0, []byte{0xFF})
	assert.Error(t, err)
}

func TestASCIIClient_WriteMultipleCoils_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.WriteMultipleCoils(0, []byte{})
	assert.Error(t, err)
}

func TestASCIIClient_WriteMultipleRegisters_NotConnected(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.WriteMultipleRegisters(0, []byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestASCIIClient_WriteMultipleRegisters_InvalidQuantity(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	err := client.WriteMultipleRegisters(0, []byte{})
	assert.Error(t, err)
}

func TestASCIIClient_SetPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	assert.False(t, client.IsConnected())
	client.SetPort(nil)
	assert.True(t, client.IsConnected())
}

func TestCalculateFrameTimeout(t *testing.T) {
	timeout := CalculateFrameTimeout(9600, 8, 1, "none")
	assert.GreaterOrEqual(t, timeout, 10*time.Millisecond)

	timeoutWithParity := CalculateFrameTimeout(9600, 8, 1, "even")
	assert.GreaterOrEqual(t, timeoutWithParity, 10*time.Millisecond)

	highBaudTimeout := CalculateFrameTimeout(115200, 8, 1, "none")
	assert.GreaterOrEqual(t, highBaudTimeout, 10*time.Millisecond)
}

func TestTCPClient_ReadWriteMultipleRegisters_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadWriteMultipleRegisters(0, 1, 0, 1, []byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestTCPClient_ReadWriteMultipleRegisters_InvalidReadQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadWriteMultipleRegisters(0, 0, 0, 1, []byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestTCPClient_ReadWriteMultipleRegisters_InvalidWriteQuantity(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	_, err := client.ReadWriteMultipleRegisters(0, 1, 0, 0, []byte{0x00, 0x01})
	assert.Error(t, err)
}

func TestTCPClient_MaskWriteRegister_NotConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	err := client.MaskWriteRegister(0, 0xFFFF, 0x0000)
	assert.Error(t, err)
}

func TestMaster_Connect_RTU(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	master := NewMaster(cfg)
	err := master.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serial port not configured")
}

func TestMaster_Connect_ASCII(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	master := NewMaster(cfg)
	err := master.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serial port not configured")
}

func TestMaster_ReadHoldingRegisters_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	_, err := master.ReadHoldingRegisters(0, 1)
	assert.Error(t, err)
}

func TestMaster_WriteMultipleRegisters_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	err := master.WriteMultipleRegisters(0, []uint16{100})
	assert.Error(t, err)
}

func TestMaster_WriteMultipleRegistersFromInt16_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	err := master.WriteMultipleRegistersFromInt16(0, []int16{-100})
	assert.Error(t, err)
}

func TestMaster_WriteMultipleRegistersFromFloat32_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	err := master.WriteMultipleRegistersFromFloat32(0, []float32{3.14})
	assert.Error(t, err)
}

func TestMaster_ReadHoldingRegistersAsInt16_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	_, err := master.ReadHoldingRegistersAsInt16(0, 1)
	assert.Error(t, err)
}

func TestMaster_ReadInputRegistersAsInt16_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	_, err := master.ReadInputRegistersAsInt16(0, 1)
	assert.Error(t, err)
}

func TestMaster_ReadHoldingRegistersAsFloat32_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	_, err := master.ReadHoldingRegistersAsFloat32(0, 1)
	assert.Error(t, err)
}

func TestMaster_ReadInputRegistersAsFloat32_Error(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = &errorMockClient{}
	_, err := master.ReadInputRegistersAsFloat32(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_WriteMultipleCoils_TooLarge(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	largeValues := make([]byte, math.MaxUint16/8+1)
	err := client.WriteMultipleCoils(0, largeValues)
	assert.Error(t, err)
}

func TestASCIIClient_WriteMultipleCoils_TooLarge(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	largeValues := make([]byte, math.MaxUint16/8+1)
	err := client.WriteMultipleCoils(0, largeValues)
	assert.Error(t, err)
}

func TestASCIIClient_WriteMultipleRegisters_TooLarge(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	largeValues := make([]byte, math.MaxUint16*2+1)
	err := client.WriteMultipleRegisters(0, largeValues)
	assert.Error(t, err)
}

func TestConverter_ConvertToUint16_InsufficientData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	_, err := conv.ConvertToUint16([]byte{0x01})
	assert.Error(t, err)
}

func TestConverter_ConvertToInt16_InsufficientData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	_, err := conv.ConvertToInt16([]byte{0x01})
	assert.Error(t, err)
}

func TestConverter_ConvertToUint32_InsufficientData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	_, err := conv.ConvertToUint32([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestConverter_ConvertToInt32_InsufficientData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	_, err := conv.ConvertToInt32([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestConverter_ConvertToFloat32_InsufficientData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	_, err := conv.ConvertToFloat32([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestConverter_ConvertToFloat64_InsufficientData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	_, err := conv.ConvertToFloat64([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestConverter_ConvertRegisters(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	data := []byte{0x00, 0x64, 0x00, 0xC8}
	regs, err := conv.ConvertRegisters(data)
	assert.NoError(t, err)
	assert.Len(t, regs, 2)
	assert.Equal(t, uint16(100), regs[0])
	assert.Equal(t, uint16(200), regs[1])
}

func TestConverter_ConvertRegisters_OddData(t *testing.T) {
	conv := NewConverter(BigEndian, HighWordFirst)
	data := []byte{0x00, 0x64, 0x00}
	_, err := conv.ConvertRegisters(data)
	assert.Error(t, err)
}

func TestTCPClient_Connect_AlreadyConnected(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	client.connected = true
	err := client.Connect()
	assert.NoError(t, err)
}

func TestTCPClient_Connect_IPv6(t *testing.T) {
	cfg := Config{Host: "::1", Port: 50202, Timeout: 1 * time.Second}
	client := NewTCPClient(cfg)
	err := client.Connect()
	assert.Error(t, err)
}

func TestUint32ToBCD_Normal(t *testing.T) {
	result := Uint32ToBCD(12345678)
	assert.Equal(t, uint32(0x01234578), result)
}

func TestUint32ToBCD_Zero(t *testing.T) {
	result := Uint32ToBCD(0)
	assert.Equal(t, uint32(0x00000000), result)
}

func TestUint32ToBCD_SmallValue(t *testing.T) {
	result := Uint32ToBCD(12)
	assert.Equal(t, uint32(0x00000012), result)
}

func TestBytesToInt64_InsufficientData(t *testing.T) {
	_, err := BytesToInt64([]byte{0x01, 0x02}, BigEndian, HighWordFirst)
	assert.Error(t, err)
}

func TestRTUClient_Connect_WithPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	port := &mockPort{}
	client.port = port
	err := client.Connect()
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
}

func TestASCIIClient_Connect_WithPort(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	port := &mockPort{}
	client.port = port
	err := client.Connect()
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
}

func TestRTUClient_sendRequest_WriteError(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1}
	client := NewRTUClient(cfg)
	client.connected = true
	client.port = &mockPort{readData: []byte{}}
	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_sendRequest_WriteError(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1}
	client := NewASCIIClient(cfg)
	client.connected = true
	client.port = &mockPort{readData: []byte{}}
	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestRTUClient_ReadResponse_UnsupportedFC(t *testing.T) {
	cfg := Config{Protocol: ProtocolRTU, SlaveID: 1, Timeout: time.Second}
	client := NewRTUClient(cfg)

	frame := NewRTUFrame(1, FuncReadFIFOQueue, []byte{0x00, 0x01})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	_, err := client.ReadHoldingRegisters(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadResponse_ExceptionResponse(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadCoils|0x80, []byte{0x02})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestASCIIClient_ReadResponse_NoExceptionCode(t *testing.T) {
	cfg := Config{Protocol: ProtocolASCII, SlaveID: 1, Timeout: time.Second}
	client := NewASCIIClient(cfg)

	frame := NewASCIIFrame(1, FuncReadCoils|0x80, []byte{})
	port := &mockPort{readData: frame.Bytes()}
	client.SetPort(port)

	_, err := client.ReadCoils(0, 1)
	assert.Error(t, err)
}

func TestParseASCIIFrame_ShortData(t *testing.T) {
	_, err := ParseASCIIFrame([]byte{':', '0', '1', 0x0D, 0x0A})
	assert.Error(t, err)
}

func TestParseASCIIFrame_NoCRLF(t *testing.T) {
	_, err := ParseASCIIFrame([]byte{':', '0', '1', '0', '3', 'F', 'F'})
	assert.Error(t, err)
}

func TestParseRTUFrame_TooShort(t *testing.T) {
	_, err := ParseRTUFrame([]byte{0x01})
	assert.Error(t, err)
}

func TestMaster_Connect_TCPFailed(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP, Host: "127.0.0.1", Port: 19999, Timeout: 1 * time.Second}
	master := NewMaster(cfg)
	err := master.Connect(context.Background())
	assert.Error(t, err)
}

func TestMaster_Disconnect_WithClient(t *testing.T) {
	cfg := Config{Protocol: ProtocolTCP}
	master := NewMaster(cfg)
	master.client = newMockClient()
	err := master.Disconnect()
	assert.NoError(t, err)
}

func TestTCPClient_Disconnect_WithConnection(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 502}
	client := NewTCPClient(cfg)
	client.conn = nil
	client.connected = true
	err := client.Disconnect()
	assert.NoError(t, err)
}
