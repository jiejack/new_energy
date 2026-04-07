package modbus

import (
	"context"
	"fmt"
	"time"
)

// ExampleUsage 展示Modbus客户端的使用示例
func ExampleUsage() {
	// ========== TCP客户端示例 ==========
	tcpConfig := Config{
		Protocol:   ProtocolTCP,
		Host:       "192.168.1.100",
		Port:       502,
		SlaveID:    1,
		Timeout:    5 * time.Second,
		RetryCount: 3,
	}

	tcpMaster := NewMaster(tcpConfig)
	ctx := context.Background()

	if err := tcpMaster.Connect(ctx); err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer tcpMaster.Disconnect()

	// 读取保持寄存器
	registers, err := tcpMaster.ReadHoldingRegisters(0, 10)
	if err != nil {
		fmt.Printf("读取寄存器失败: %v\n", err)
	} else {
		fmt.Printf("读取到的寄存器值: %v\n", registers)
	}

	// 读取线圈状态
	coils, err := tcpMaster.ReadCoils(0, 8)
	if err != nil {
		fmt.Printf("读取线圈失败: %v\n", err)
	} else {
		fmt.Printf("线圈状态: %v\n", coils)
	}

	// 写单个寄存器
	if err := tcpMaster.WriteSingleRegister(0, 100); err != nil {
		fmt.Printf("写寄存器失败: %v\n", err)
	}

	// 写多个寄存器
	values := []uint16{100, 200, 300}
	if err := tcpMaster.WriteMultipleRegisters(0, values); err != nil {
		fmt.Printf("写多个寄存器失败: %v\n", err)
	}

	// ========== RTU客户端示例 ==========
	rtuConfig := Config{
		Protocol:   ProtocolRTU,
		SlaveID:    1,
		Timeout:    2 * time.Second,
		RetryCount: 3,
	}

	serialConfig := SerialPortConfig{
		PortName: "COM1",
		BaudRate: 9600,
		DataBits: 8,
		Parity:   "none",
		StopBits: 1,
	}

	rtuMaster := NewMasterWithSerial(rtuConfig, serialConfig)
	_ = rtuMaster // 示例代码，实际使用时需要连接

	// 注意：实际使用时需要先打开串口
	// rtuMaster.Connect(ctx)

	// ========== ASCII客户端示例 ==========
	asciiConfig := Config{
		Protocol:   ProtocolASCII,
		SlaveID:    1,
		Timeout:    2 * time.Second,
	}

	asciiMaster := NewMasterWithSerial(asciiConfig, serialConfig)
	_ = asciiMaster // 示例代码，实际使用时需要连接

	// 注意：实际使用时需要先打开串口
	// asciiMaster.Connect(ctx)

	// ========== 数据类型转换示例 ==========
	// 设置转换器（大端序，高字在前）
	tcpMaster.SetConverter(BigEndian, HighWordFirst)

	// 读取寄存器并转换为int16
	int16Values, err := tcpMaster.ReadHoldingRegistersAsInt16(0, 10)
	if err != nil {
		fmt.Printf("读取失败: %v\n", err)
	} else {
		fmt.Printf("Int16值: %v\n", int16Values)
	}

	// 读取寄存器并转换为float32
	float32Values, err := tcpMaster.ReadHoldingRegistersAsFloat32(0, 5)
	if err != nil {
		fmt.Printf("读取失败: %v\n", err)
	} else {
		fmt.Printf("Float32值: %v\n", float32Values)
	}

	// ========== 批量操作示例 ==========
	readRequests := []ReadRequest{
		{Address: 0, Quantity: 10, Type: DataTypeHoldingRegister},
		{Address: 0, Quantity: 8, Type: DataTypeCoil},
		{Address: 100, Quantity: 5, Type: DataTypeInputRegister},
	}

	responses, err := tcpMaster.BatchRead(readRequests)
	if err != nil {
		fmt.Printf("批量读取失败: %v\n", err)
	}

	for i, resp := range responses {
		if resp.Error != nil {
			fmt.Printf("请求 %d 失败: %v\n", i, resp.Error)
		} else {
			fmt.Printf("请求 %d 成功: %v\n", i, resp.Data)
		}
	}

	// ========== 轮询示例 ==========
	poller := NewPoller(tcpMaster, 1*time.Second)
	poller.SetRequests(readRequests)
	poller.Start()

	// 处理轮询响应
	go func() {
		for resp := range poller.Responses() {
			if resp.Error != nil {
				fmt.Printf("轮询错误: %v\n", resp.Error)
			} else {
				fmt.Printf("轮询数据: %+v\n", resp.Data)
			}
		}
	}()

	// 运行一段时间后停止
	time.Sleep(10 * time.Second)
	poller.Stop()

	// ========== 连接池示例 ==========
	poolConfig := Config{
		Protocol: ProtocolTCP,
		Host:     "192.168.1.100",
		Port:     502,
		SlaveID:  1,
		Timeout:  5 * time.Second,
	}

	pool := NewTCPClientPool(poolConfig, 10)

	// 从连接池获取客户端
	client, err := pool.Get()
	if err != nil {
		fmt.Printf("获取连接失败: %v\n", err)
		return
	}

	// 使用客户端
	data, err := client.ReadHoldingRegisters(0, 10)
	if err != nil {
		fmt.Printf("读取失败: %v\n", err)
	} else {
		fmt.Printf("读取到的数据: %v\n", data)
	}

	// 将客户端放回连接池
	pool.Put(client)

	// 关闭连接池
	defer pool.Close()

	// ========== 错误处理示例 ==========
	// 读取不存在的地址，会收到异常响应
	_, err = tcpMaster.ReadHoldingRegisters(9999, 10)
	if err != nil {
		if modbusErr, ok := err.(*ModbusError); ok {
			fmt.Printf("Modbus异常 - 功能码: %s, 异常码: %s\n",
				modbusErr.FunctionCode.String(),
				modbusErr.ExceptionCode.String())
		} else {
			fmt.Printf("其他错误: %v\n", err)
		}
	}
}

// ExampleTCPClientWithRetry 展示带重试机制的TCP客户端使用
func ExampleTCPClientWithRetry() {
	config := Config{
		Protocol:      ProtocolTCP,
		Host:          "192.168.1.100",
		Port:          502,
		SlaveID:       1,
		Timeout:       5 * time.Second,
		RetryCount:    3,
		RetryInterval: 1 * time.Second,
	}

	master := NewMaster(config)
	ctx := context.Background()

	var lastErr error
	for i := 0; i < config.RetryCount; i++ {
		if err := master.Connect(ctx); err != nil {
			lastErr = err
			time.Sleep(config.RetryInterval)
			continue
		}

		// 连接成功，执行操作
		registers, err := master.ReadHoldingRegisters(0, 10)
		if err == nil {
			fmt.Printf("成功读取: %v\n", registers)
			master.Disconnect()
			return
		}

		lastErr = err
		master.Disconnect()
		time.Sleep(config.RetryInterval)
	}

	fmt.Printf("重试 %d 次后仍然失败: %v\n", config.RetryCount, lastErr)
}

// ExampleDataConversion 展示数据类型转换
func ExampleDataConversion() {
	// 创建转换器
	converter := NewConverter(BigEndian, HighWordFirst)

	// uint16转换
	uint16Bytes := Uint16ToBytes(0x1234, BigEndian)
	uint16Val, _ := BytesToUint16(uint16Bytes, BigEndian)
	fmt.Printf("Uint16: 0x%04X\n", uint16Val)

	// int16转换
	int16Bytes := Int16ToBytes(-12345, BigEndian)
	int16Val, _ := BytesToInt16(int16Bytes, BigEndian)
	fmt.Printf("Int16: %d\n", int16Val)

	// float32转换
	float32Bytes := Float32ToBytes(123.456, BigEndian, HighWordFirst)
	float32Val, _ := BytesToFloat32(float32Bytes, BigEndian, HighWordFirst)
	fmt.Printf("Float32: %f\n", float32Val)

	// BCD转换
	bcdValue := Uint8ToBCD(12) // 12 -> 0x12
	decimalValue := BCDToUint8(bcdValue)
	fmt.Printf("BCD: 0x%02X -> Decimal: %d\n", bcdValue, decimalValue)

	// 使用转换器
	data := []byte{0x00, 0x64, 0x00, 0xC8} // 100, 200
	registers, _ := converter.ConvertRegisters(data)
	fmt.Printf("寄存器: %v\n", registers)

	floatData := []byte{0x42, 0xF6, 0xE9, 0x79} // 123.45
	floatVal, _ := converter.ConvertToFloat32(floatData)
	fmt.Printf("Float32: %f\n", floatVal)
}

// ExampleFrameConstruction 展示帧构造
func ExampleFrameConstruction() {
	// RTU帧构造
	rtuFrame := NewRTUFrame(0x01, FuncReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x0A})
	rtuBytes := rtuFrame.Bytes()
	fmt.Printf("RTU帧: % X\n", rtuBytes)

	// ASCII帧构造
	asciiFrame := NewASCIIFrame(0x01, FuncReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x0A})
	asciiBytes := asciiFrame.Bytes()
	fmt.Printf("ASCII帧: %s\n", string(asciiBytes))

	// PDU构造
	pdu := NewPDU(FuncReadCoils, []byte{0x00, 0x00, 0x00, 0x08})
	pduBytes := pdu.Bytes()
	fmt.Printf("PDU: % X\n", pduBytes)

	// MBAP头构造
	mbap := &MBAPHeader{
		TransactionID: 0x0001,
		ProtocolID:    0x0000,
		Length:        0x0006,
		UnitID:        0x01,
	}
	mbapBytes := mbap.Bytes()
	fmt.Printf("MBAP头: % X\n", mbapBytes)
}
