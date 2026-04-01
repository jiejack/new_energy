package iec104

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// 连接配置
type ConnectionConfig struct {
	Host                 string        // 目标主机
	Port                 int           // 目标端口
	CommonAddress        uint16        // 公共地址
	LinkAddress          uint16        // 链路地址
	Timeout              time.Duration // 连接超时
	ReconnectInterval    time.Duration // 重连间隔
	HeartbeatInterval    time.Duration // 心跳间隔
	HeartbeatTimeout     time.Duration // 心跳超时
	MaxReconnectAttempts int           // 最大重连次数 (0表示无限)
	SendBufferSize       int           // 发送缓冲区大小
	RecvBufferSize       int           // 接收缓冲区大小
	BalancedMode         bool          // 是否平衡式传输
}

// 默认连接配置
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		Timeout:              10 * time.Second,
		ReconnectInterval:    5 * time.Second,
		HeartbeatInterval:    15 * time.Second,
		HeartbeatTimeout:     30 * time.Second,
		MaxReconnectAttempts: 0, // 无限重连
		SendBufferSize:       4096,
		RecvBufferSize:       65535,
		BalancedMode:         false,
	}
}

// 连接事件类型
type ConnectionEvent int

const (
	EVENT_CONNECTED ConnectionEvent = iota
	EVENT_DISCONNECTED
	EVENT_STARTDT_CON
	EVENT_STOPDT_CON
	EVENT_TESTFR_TIMEOUT
	EVENT_RECONNECTING
	EVENT_ERROR
)

// 连接事件
type ConnEvent struct {
	Type  ConnectionEvent
	Error error
}

// 连接统计
type ConnectionStats struct {
	ConnectCount      uint64
	DisconnectCount   uint64
	BytesSent         uint64
	BytesReceived     uint64
	FramesSent        uint64
	FramesReceived    uint64
	IFramesSent       uint64
	IFramesReceived   uint64
	SFramesSent       uint64
	SFramesReceived   uint64
	UFramesSent       uint64
	UFramesReceived   uint64
	LastConnectTime   time.Time
	LastActivityTime  time.Time
	ReconnectAttempts uint64
}

// 连接管理器
type Connection struct {
	config     ConnectionConfig
	conn       net.Conn
	state      int32 // atomic: ConnectionState
	sendSeq    uint16
	recvSeq    uint16
	ackSeq     uint16 // 已确认的接收序号

	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	// 通道
	eventChan     chan ConnEvent
	frameChan     chan *APDU
	errChan       chan error

	// 心跳
	lastSendTime  time.Time
	lastRecvTime  time.Time
	testfrSent    bool
	testfrTimer   *time.Timer

	// 统计
	stats         ConnectionStats

	// 重连
	reconnectAttempts int
	stopReconnect     int32

	// 发送队列
	sendQueue     chan []byte
	sendMu        sync.Mutex

	// 接收缓冲
	recvBuffer    []byte
	recvBufMu     sync.Mutex
}

// 新建连接
func NewConnection(config ConnectionConfig) *Connection {
	ctx, cancel := context.WithCancel(context.Background())

	return &Connection{
		config:     config,
		state:      int32(STATE_DISCONNECTED),
		sendSeq:    0,
		recvSeq:    0,
		ackSeq:     0,
		ctx:        ctx,
		cancel:     cancel,
		eventChan:  make(chan ConnEvent, 100),
		frameChan:  make(chan *APDU, 10000),
		errChan:    make(chan error, 100),
		sendQueue:  make(chan []byte, 1000),
		recvBuffer: make([]byte, 0, config.RecvBufferSize),
	}
}

// 连接
func (c *Connection) Connect() error {
	return c.ConnectWithContext(c.ctx)
}

// 带上下文的连接
func (c *Connection) ConnectWithContext(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.state) != int32(STATE_DISCONNECTED) {
		return errors.New("already connected or connecting")
	}

	atomic.StoreInt32(&c.state, int32(STATE_CONNECTING))
	c.emitEvent(EVENT_RECONNECTING, nil)

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	dialer := &net.Dialer{Timeout: c.config.Timeout}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		atomic.StoreInt32(&c.state, int32(STATE_DISCONNECTED))
		c.reconnectAttempts++
		c.stats.ReconnectAttempts++
		c.emitEvent(EVENT_ERROR, fmt.Errorf("connection failed: %w", err))
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.conn = conn
	c.sendSeq = 0
	c.recvSeq = 0
	c.ackSeq = 0
	c.testfrSent = false
	c.recvBuffer = c.recvBuffer[:0]
	c.stats.ConnectCount++
	c.stats.LastConnectTime = time.Now()
	c.stats.LastActivityTime = time.Now()
	c.lastSendTime = time.Now()
	c.lastRecvTime = time.Now()

	atomic.StoreInt32(&c.state, int32(STATE_CONNECTED))
	c.emitEvent(EVENT_CONNECTED, nil)

	// 启动处理协程
	c.wg.Add(3)
	go c.readLoop()
	go c.writeLoop()
	go c.heartbeatLoop()

	// 发送STARTDT
	if !c.config.BalancedMode {
		if err := c.sendStartDT(); err != nil {
			_ = c.close()
			return fmt.Errorf("failed to send STARTDT: %w", err)
		}
		atomic.StoreInt32(&c.state, int32(STATE_STARTDT_SENT))
	} else {
		atomic.StoreInt32(&c.state, int32(STATE_ACTIVE))
	}

	return nil
}

// 断开连接
func (c *Connection) Disconnect() error {
	return c.close()
}

// 内部关闭
func (c *Connection) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.state) == int32(STATE_DISCONNECTED) {
		return nil
	}

	atomic.StoreInt32(&c.state, int32(STATE_STOPPING))

	if c.cancel != nil {
		c.cancel()
	}

	// 等待所有协程退出
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	atomic.StoreInt32(&c.state, int32(STATE_DISCONNECTED))
	c.stats.DisconnectCount++
	c.emitEvent(EVENT_DISCONNECTED, nil)

	return nil
}

// 获取状态
func (c *Connection) GetState() ConnectionState {
	return ConnectionState(atomic.LoadInt32(&c.state))
}

// 是否已连接
func (c *Connection) IsConnected() bool {
	state := c.GetState()
	return state == STATE_CONNECTED || state == STATE_STARTDT_SENT || state == STATE_ACTIVE
}

// 是否激活
func (c *Connection) IsActive() bool {
	return c.GetState() == STATE_ACTIVE
}

// 事件通道
func (c *Connection) EventChannel() <-chan ConnEvent {
	return c.eventChan
}

// 帧通道
func (c *Connection) FrameChannel() <-chan *APDU {
	return c.frameChan
}

// 错误通道
func (c *Connection) ErrorChannel() <-chan error {
	return c.errChan
}

// 获取统计
func (c *Connection) GetStats() ConnectionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// 发送事件
func (c *Connection) emitEvent(t ConnectionEvent, err error) {
	select {
	case c.eventChan <- ConnEvent{Type: t, Error: err}:
	default:
	}
}

// 发送错误
func (c *Connection) emitError(err error) {
	select {
	case c.errChan <- err:
	default:
	}
}

// 发送STARTDT_ACT
func (c *Connection) sendStartDT() error {
	frame := []byte{START_BYTE, 0x04, U_FRAME_STARTDT_ACT, 0x00, 0x00, 0x00}
	return c.sendRaw(frame)
}

// 发送STOPDT_ACT
func (c *Connection) sendStopDT() error {
	frame := []byte{START_BYTE, 0x04, U_FRAME_STOPDT_ACT, 0x00, 0x00, 0x00}
	return c.sendRaw(frame)
}

// 发送TESTFR_ACT
func (c *Connection) sendTestFR() error {
	frame := []byte{START_BYTE, 0x04, U_FRAME_TESTFR_ACT, 0x00, 0x00, 0x00}
	c.testfrSent = true
	c.lastSendTime = time.Now()
	return c.sendRaw(frame)
}

// 发送TESTFR_CON
func (c *Connection) sendTestFRCon() error {
	frame := []byte{START_BYTE, 0x04, U_FRAME_TESTFR_CON, 0x00, 0x00, 0x00}
	return c.sendRaw(frame)
}

// 发送S帧
func (c *Connection) sendSFrame() error {
	c.mu.Lock()
	recvSeq := c.recvSeq
	c.mu.Unlock()

	control1 := byte((recvSeq << 1) & 0xFF)
	control2 := byte((recvSeq >> 7) & 0xFF)

	frame := []byte{START_BYTE, 0x04, 0x01, control1, control2, 0x00}
	c.stats.SFramesSent++
	return c.sendRaw(frame)
}

// 发送I帧
func (c *Connection) sendIFrame(asdu []byte) error {
	c.mu.Lock()
	c.sendSeq++
	sendSeq := c.sendSeq
	recvSeq := c.recvSeq
	c.mu.Unlock()

	control1 := byte((sendSeq << 1) & 0xFF)
	control2 := byte((sendSeq >> 7) & 0xFF)
	control3 := byte((recvSeq << 1) & 0xFF)
	control4 := byte((recvSeq >> 7) & 0xFF)

	frame := make([]byte, 0, len(asdu)+6)
	frame = append(frame, START_BYTE)
	frame = append(frame, byte(len(asdu)+4))
	frame = append(frame, control1, control2, control3, control4)
	frame = append(frame, asdu...)

	c.stats.IFramesSent++
	c.lastSendTime = time.Now()
	return c.sendRaw(frame)
}

// 原始发送
func (c *Connection) sendRaw(data []byte) error {
	select {
	case c.sendQueue <- data:
		return nil
	default:
		return errors.New("send queue full")
	}
}

// 直接写入
func (c *Connection) writeDirect(data []byte) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("not connected")
	}

	n, err := conn.Write(data)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.stats.BytesSent += uint64(n)
	c.stats.FramesSent++
	c.mu.Unlock()

	return nil
}

// 写入协程
func (c *Connection) writeLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.sendQueue:
			if err := c.writeDirect(data); err != nil {
				c.emitError(fmt.Errorf("write error: %w", err))
				c.handleDisconnect()
				return
			}
		}
	}
}

// 读取协程
func (c *Connection) readLoop() {
	defer c.wg.Done()

	buf := make([]byte, c.config.RecvBufferSize)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 设置读取超时
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))

		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if atomic.LoadInt32(&c.state) == int32(STATE_DISCONNECTED) ||
				atomic.LoadInt32(&c.state) == int32(STATE_STOPPING) {
				return
			}
			c.emitError(fmt.Errorf("read error: %w", err))
			c.handleDisconnect()
			return
		}

		if n > 0 {
			c.mu.Lock()
			c.stats.BytesReceived += uint64(n)
			c.stats.LastActivityTime = time.Now()
			c.lastRecvTime = time.Now()
			c.mu.Unlock()

			c.processReceivedData(buf[:n])
		}
	}
}

// 处理接收数据
func (c *Connection) processReceivedData(data []byte) {
	c.recvBufMu.Lock()
	c.recvBuffer = append(c.recvBuffer, data...)
	c.recvBufMu.Unlock()

	for {
		c.recvBufMu.Lock()
		if len(c.recvBuffer) < 6 {
			c.recvBufMu.Unlock()
			break
		}

		if c.recvBuffer[0] != START_BYTE {
			// 寻找启动字节
			idx := -1
			for i, b := range c.recvBuffer {
				if b == START_BYTE {
					idx = i
					break
				}
			}
			if idx > 0 {
				c.recvBuffer = c.recvBuffer[idx:]
			} else {
				c.recvBuffer = c.recvBuffer[:0]
			}
			c.recvBufMu.Unlock()
			continue
		}

		frameLen := int(c.recvBuffer[1])
		totalLen := frameLen + 2

		if len(c.recvBuffer) < totalLen {
			c.recvBufMu.Unlock()
			break
		}

		frame := make([]byte, totalLen)
		copy(frame, c.recvBuffer[:totalLen])
		c.recvBuffer = c.recvBuffer[totalLen:]
		c.recvBufMu.Unlock()

		c.processFrame(frame)
	}
}

// 处理帧
func (c *Connection) processFrame(frame []byte) {
	c.mu.Lock()
	c.stats.FramesReceived++
	c.mu.Unlock()

	apdu := c.parseAPDU(frame)
	if apdu == nil {
		return
	}

	switch apdu.Type {
	case APDU_TYPE_I:
		c.handleIFrame(apdu)
	case APDU_TYPE_S:
		c.handleSFrame(apdu)
	case APDU_TYPE_U:
		c.handleUFrame(apdu)
	}
}

// 解析APDU
func (c *Connection) parseAPDU(frame []byte) *APDU {
	if len(frame) < 6 {
		return nil
	}

	apdu := &APDU{
		RawBytes: frame,
	}

	control1 := frame[2]
	control2 := frame[3]

	// 判断帧类型
	if control1&0x01 == 0 {
		// I帧
		apdu.Type = APDU_TYPE_I
		apdu.SendSeq = uint16(control1>>1) | uint16(control2)<<7

		control3 := frame[4]
		control4 := frame[5]
		apdu.RecvSeq = uint16(control3>>1) | uint16(control4)<<7

		// 解析ASDU
		if len(frame) > 6 {
			apdu.ASDU = c.parseASDU(frame[6:])
		}

		c.mu.Lock()
		c.stats.IFramesReceived++
		c.mu.Unlock()

	} else if control1&0x03 == 0x01 {
		// S帧
		apdu.Type = APDU_TYPE_S

		control3 := frame[4]
		control4 := frame[5]
		apdu.RecvSeq = uint16(control3>>1) | uint16(control4)<<7

		c.mu.Lock()
		c.stats.SFramesReceived++
		c.mu.Unlock()

	} else {
		// U帧
		apdu.Type = APDU_TYPE_U
		apdu.Control = control1

		c.mu.Lock()
		c.stats.UFramesReceived++
		c.mu.Unlock()
	}

	return apdu
}

// 解析ASDU
func (c *Connection) parseASDU(data []byte) *ASDU {
	if len(data) < 6 {
		return nil
	}

	asdu := &ASDU{
		TypeID: data[0],
		VSQ:    ParseVSQ(data[1]),
		COT:    ParseCOT(data[2], data[3]),
		Origin: data[3],
	}

	// 公共地址 (2字节)
	asdu.CommonAddr = uint16(data[4]) | uint16(data[5])<<8

	return asdu
}

// 处理I帧
func (c *Connection) handleIFrame(apdu *APDU) {
	c.mu.Lock()
	// 更新接收序号
	expectedSeq := c.recvSeq + 1
	if apdu.SendSeq == expectedSeq {
		c.recvSeq = apdu.SendSeq
	}
	c.mu.Unlock()

	// 发送到帧通道
	select {
	case c.frameChan <- apdu:
	default:
		c.emitError(errors.New("frame channel full, dropping frame"))
	}

	// 发送S帧确认
	// 可以优化为批量确认
}

// 处理S帧
func (c *Connection) handleSFrame(apdu *APDU) {
	c.mu.Lock()
	c.ackSeq = apdu.RecvSeq
	c.mu.Unlock()
}

// 处理U帧
func (c *Connection) handleUFrame(apdu *APDU) {
	switch apdu.Control {
	case U_FRAME_STARTDT_CON:
		atomic.StoreInt32(&c.state, int32(STATE_ACTIVE))
		c.emitEvent(EVENT_STARTDT_CON, nil)

	case U_FRAME_STOPDT_CON:
		atomic.StoreInt32(&c.state, int32(STATE_CONNECTED))
		c.emitEvent(EVENT_STOPDT_CON, nil)

	case U_FRAME_TESTFR_ACT:
		// 响应TESTFR_CON
		_ = c.sendTestFRCon()

	case U_FRAME_TESTFR_CON:
		c.mu.Lock()
		c.testfrSent = false
		c.mu.Unlock()
	}
}

// 心跳协程
func (c *Connection) heartbeatLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	timeoutTicker := time.NewTicker(time.Second)
	defer timeoutTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-ticker.C:
			if !c.IsActive() {
				continue
			}

			c.mu.Lock()
			testfrSent := c.testfrSent
			lastSend := c.lastSendTime
			lastRecv := c.lastRecvTime
			c.mu.Unlock()

			now := time.Now()

			// 检查心跳超时
			if testfrSent && now.Sub(lastRecv) > c.config.HeartbeatTimeout {
				c.emitEvent(EVENT_TESTFR_TIMEOUT, nil)
				c.emitError(errors.New("heartbeat timeout"))
				c.handleDisconnect()
				return
			}

			// 发送心跳
			if !testfrSent && now.Sub(lastSend) > c.config.HeartbeatInterval {
				if err := c.sendTestFR(); err != nil {
					c.emitError(fmt.Errorf("send testfr failed: %w", err))
					c.handleDisconnect()
					return
				}
			}

		case <-timeoutTicker.C:
			// 检查接收超时
			c.mu.Lock()
			lastRecv := c.lastRecvTime
			c.mu.Unlock()

			if time.Since(lastRecv) > c.config.HeartbeatTimeout*2 {
				c.emitError(errors.New("receive timeout"))
				c.handleDisconnect()
				return
			}
		}
	}
}

// 处理断开
func (c *Connection) handleDisconnect() {
	_ = c.close()

	// 自动重连
	if atomic.LoadInt32(&c.stopReconnect) == 0 {
		go c.reconnect()
	}
}

// 重连
func (c *Connection) reconnect() {
	if c.config.MaxReconnectAttempts > 0 && c.reconnectAttempts >= c.config.MaxReconnectAttempts {
		c.emitError(errors.New("max reconnect attempts reached"))
		return
	}

	ticker := time.NewTicker(c.config.ReconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt32(&c.stopReconnect) != 0 {
				return
			}

			c.emitEvent(EVENT_RECONNECTING, nil)

			if err := c.Connect(); err == nil {
				c.reconnectAttempts = 0
				return
			}
		}
	}
}

// 停止重连
func (c *Connection) StopReconnect() {
	atomic.StoreInt32(&c.stopReconnect, 1)
}

// 恢复重连
func (c *Connection) ResumeReconnect() {
	atomic.StoreInt32(&c.stopReconnect, 0)
}

// 获取发送序号
func (c *Connection) GetSendSeq() uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sendSeq
}

// 获取接收序号
func (c *Connection) GetRecvSeq() uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.recvSeq
}

// 更新接收序号
func (c *Connection) UpdateRecvSeq(seq uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recvSeq = seq
}

// 发送确认
func (c *Connection) SendAck() error {
	return c.sendSFrame()
}

// 获取公共地址
func (c *Connection) GetCommonAddress() uint16 {
	return c.config.CommonAddress
}

// 设置公共地址
func (c *Connection) SetCommonAddress(addr uint16) {
	c.config.CommonAddress = addr
}

// 连接池管理器
type ConnectionPool struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	configs     map[string]ConnectionConfig
}

// 新建连接池
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*Connection),
		configs:     make(map[string]ConnectionConfig),
	}
}

// 添加连接配置
func (p *ConnectionPool) AddConfig(name string, config ConnectionConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.configs[name] = config
}

// 获取连接
func (p *ConnectionPool) Get(name string) (*Connection, error) {
	p.mu.RLock()
	if conn, ok := p.connections[name]; ok {
		p.mu.RUnlock()
		return conn, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查
	if conn, ok := p.connections[name]; ok {
		return conn, nil
	}

	config, ok := p.configs[name]
	if !ok {
		return nil, fmt.Errorf("connection config not found: %s", name)
	}

	conn := NewConnection(config)
	p.connections[name] = config
	return conn, nil
}

// 连接所有
func (p *ConnectionPool) ConnectAll() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var errs []error
	for name, config := range p.configs {
		conn := NewConnection(config)
		if err := conn.Connect(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		p.connections[name] = conn
	}

	if len(errs) > 0 {
		return fmt.Errorf("connection errors: %v", errs)
	}
	return nil
}

// 断开所有
func (p *ConnectionPool) DisconnectAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.connections {
		_ = conn.Disconnect()
	}
}

// 移除连接
func (p *ConnectionPool) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.connections[name]; ok {
		_ = conn.Disconnect()
		delete(p.connections, name)
	}
	delete(p.configs, name)
}
