package iec104

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 主站配置
type MasterConfig struct {
	Host                 string        // 目标主机
	Port                 int           // 目标端口
	CommonAddress        uint16        // 公共地址
	Timeout              time.Duration // 操作超时
	ReconnectInterval    time.Duration // 重连间隔
	HeartbeatInterval    time.Duration // 心跳间隔
	HeartbeatTimeout     time.Duration // 心跳超时
	MaxReconnectAttempts int           // 最大重连次数
	BalancedMode         bool          // 平衡式传输模式
	DataBufferSize       int           // 数据缓冲区大小
}

// 默认主站配置
func DefaultMasterConfig() MasterConfig {
	return MasterConfig{
		Timeout:              10 * time.Second,
		ReconnectInterval:    5 * time.Second,
		HeartbeatInterval:    15 * time.Second,
		HeartbeatTimeout:     30 * time.Second,
		MaxReconnectAttempts: 0, // 无限重连
		BalancedMode:         false,
		DataBufferSize:       10000,
	}
}

// 主站状态
type MasterState int32

const (
	MASTER_STATE_STOPPED MasterState = iota
	MASTER_STATE_STARTING
	MASTER_STATE_RUNNING
	MASTER_STATE_STOPPING
)

// 遥控操作结果
type ControlResult struct {
	Success bool
	Error   error
	ASDU    *ASDU
}

// 主站实现
type Master struct {
	config MasterConfig
	conn   *Connection
	coder  *ASDUCoder
	state  int32 // atomic: MasterState

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 数据通道
	dataChan chan *ASDU
	errChan  chan error

	// 遥控响应等待
	controlWaiters map[uint16]chan *ControlResult
	controlMu      sync.RWMutex

	// 召唤响应等待
	interrogationWaiters map[uint8]chan *ASDU
	interrogationMu      sync.RWMutex

	// 统计
	stats MasterStats

	// 回调函数
	onDataHandler      func(*ASDU)
	onConnectHandler   func()
	onDisconnectHandler func()
	onErrorHandler     func(error)
}

// 主站统计
type MasterStats struct {
	TotalASDUReceived   uint64
	TotalASDUSent       uint64
	TotalControlSent    uint64
	TotalControlSuccess uint64
	TotalControlFailed  uint64
	TotalGISent         uint64
	TotalGISuccess      uint64
	TotalGIFailed       uint64
	LastGIStartTime     time.Time
	LastGIEndTime       time.Time
}

// 新建主站
func NewMaster(config MasterConfig) *Master {
	ctx, cancel := context.WithCancel(context.Background())

	return &Master{
		config:               config,
		coder:                NewASDUCoder(),
		state:                int32(MASTER_STATE_STOPPED),
		ctx:                  ctx,
		cancel:               cancel,
		dataChan:             make(chan *ASDU, config.DataBufferSize),
		errChan:              make(chan error, 100),
		controlWaiters:       make(map[uint16]chan *ControlResult),
		interrogationWaiters: make(map[uint8]chan *ASDU),
	}
}

// 启动主站
func (m *Master) Start() error {
	if !atomic.CompareAndSwapInt32(&m.state, int32(MASTER_STATE_STOPPED), int32(MASTER_STATE_STARTING)) {
		return errors.New("master already running")
	}

	connConfig := ConnectionConfig{
		Host:                 m.config.Host,
		Port:                 m.config.Port,
		CommonAddress:        m.config.CommonAddress,
		Timeout:              m.config.Timeout,
		ReconnectInterval:    m.config.ReconnectInterval,
		HeartbeatInterval:    m.config.HeartbeatInterval,
		HeartbeatTimeout:     m.config.HeartbeatTimeout,
		MaxReconnectAttempts: m.config.MaxReconnectAttempts,
		BalancedMode:         m.config.BalancedMode,
	}

	m.conn = NewConnection(connConfig)

	// 启动事件处理
	m.wg.Add(1)
	go m.eventLoop()

	// 启动帧处理
	m.wg.Add(1)
	go m.frameLoop()

	// 连接
	if err := m.conn.Connect(); err != nil {
		atomic.StoreInt32(&m.state, int32(MASTER_STATE_STOPPED))
		return fmt.Errorf("connection failed: %w", err)
	}

	atomic.StoreInt32(&m.state, int32(MASTER_STATE_RUNNING))
	return nil
}

// 停止主站
func (m *Master) Stop() error {
	if !atomic.CompareAndSwapInt32(&m.state, int32(MASTER_STATE_RUNNING), int32(MASTER_STATE_STOPPING)) {
		return nil
	}

	m.cancel()

	if m.conn != nil {
		_ = m.conn.Disconnect()
	}

	// 等待协程退出
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	atomic.StoreInt32(&m.state, int32(MASTER_STATE_STOPPED))
	return nil
}

// 获取状态
func (m *Master) GetState() MasterState {
	return MasterState(atomic.LoadInt32(&m.state))
}

// 是否运行中
func (m *Master) IsRunning() bool {
	return m.GetState() == MASTER_STATE_RUNNING
}

// 是否已连接
func (m *Master) IsConnected() bool {
	return m.conn != nil && m.conn.IsActive()
}

// 数据通道
func (m *Master) DataChannel() <-chan *ASDU {
	return m.dataChan
}

// 错误通道
func (m *Master) ErrorChannel() <-chan error {
	return m.errChan
}

// 设置数据回调
func (m *Master) OnData(handler func(*ASDU)) {
	m.onDataHandler = handler
}

// 设置连接回调
func (m *Master) OnConnect(handler func()) {
	m.onConnectHandler = handler
}

// 设置断开回调
func (m *Master) OnDisconnect(handler func()) {
	m.onDisconnectHandler = handler
}

// 设置错误回调
func (m *Master) OnError(handler func(error)) {
	m.onErrorHandler = handler
}

// 事件处理协程
func (m *Master) eventLoop() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case event := <-m.conn.EventChannel():
			switch event.Type {
			case EVENT_CONNECTED:
				if m.onConnectHandler != nil {
					m.onConnectHandler()
				}
			case EVENT_DISCONNECTED:
				if m.onDisconnectHandler != nil {
					m.onDisconnectHandler()
				}
			case EVENT_ERROR:
				m.emitError(event.Error)
			}
		case err := <-m.conn.ErrorChannel():
			m.emitError(err)
		}
	}
}

// 帧处理协程
func (m *Master) frameLoop() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case apdu := <-m.conn.FrameChannel():
			m.handleAPDU(apdu)
		}
	}
}

// 处理APDU
func (m *Master) handleAPDU(apdu *APDU) {
	if apdu.ASDU == nil {
		return
	}

	m.mu.Lock()
	m.stats.TotalASDUReceived++
	m.mu.Unlock()

	asdu := apdu.ASDU

	// 完整解析ASDU
	fullASDU, err := m.coder.DecodeASDU(apdu.RawBytes[6:])
	if err == nil && fullASDU != nil {
		asdu = fullASDU
	}

	// 检查是否为响应
	cot := asdu.COT.Cause

	// 处理遥控响应
	if asdu.TypeID == TYPE_ID_SINGLE_COMMAND ||
		asdu.TypeID == TYPE_ID_DOUBLE_COMMAND ||
		asdu.TypeID == TYPE_ID_REGULATING_STEP_COMMAND ||
		asdu.TypeID == TYPE_ID_SET_POINT_COMMAND_NORMAL ||
		asdu.TypeID == TYPE_ID_SET_POINT_COMMAND_SCALED ||
		asdu.TypeID == TYPE_ID_SET_POINT_COMMAND_FLOAT {

		if cot == COT_ACTIVATION_CON || cot == COT_DEACTIVATION_CON {
			m.handleControlResponse(asdu)
		}
	}

	// 处理召唤响应
	if asdu.TypeID == TYPE_ID_INTERROGATION_CMD {
		if cot == COT_ACTIVATION_CON || cot == COT_ACTIVATION_TERMINATION {
			m.handleInterrogationResponse(asdu)
		}
	}

	if asdu.TypeID == TYPE_ID_COUNTER_INTERROGATION_CMD {
		if cot == COT_ACTIVATION_CON || cot == COT_ACTIVATION_TERMINATION {
			m.handleInterrogationResponse(asdu)
		}
	}

	// 发送到数据通道
	select {
	case m.dataChan <- asdu:
	default:
		m.emitError(errors.New("data channel full, dropping ASDU"))
	}

	// 调用回调
	if m.onDataHandler != nil {
		m.onDataHandler(asdu)
	}
}

// 处理遥控响应
func (m *Master) handleControlResponse(asdu *ASDU) {
	m.controlMu.RLock()
	defer m.controlMu.RUnlock()

	// 通知所有等待者
	for _, ch := range m.controlWaiters {
		select {
		case ch <- &ControlResult{
			Success: asdu.COT.Cause == COT_ACTIVATION_CON && !asdu.COT.IsPN,
			ASDU:    asdu,
		}:
		default:
		}
	}
}

// 处理召唤响应
func (m *Master) handleInterrogationResponse(asdu *ASDU) {
	m.interrogationMu.RLock()
	defer m.interrogationMu.RUnlock()

	var qoi uint8
	if len(asdu.Information) > 0 {
		if cmd, ok := asdu.Information[0].Value.(*InterrogationCommand); ok {
			qoi = cmd.QOI
		}
	}

	if ch, ok := m.interrogationWaiters[qoi]; ok {
		select {
		case ch <- asdu:
		default:
		}
	}
}

// 发送错误
func (m *Master) emitError(err error) {
	select {
	case m.errChan <- err:
	default:
	}
	if m.onErrorHandler != nil {
		m.onErrorHandler(err)
	}
}

// ==================== 总召唤 ====================

// 总召唤
func (m *Master) Interrogation() error {
	return m.InterrogationWithQOI(QOI_STATION_INTERROGATION)
}

// 总召唤(指定组)
func (m *Master) InterrogationWithQOI(qoi uint8) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &InterrogationCommand{QOI: qoi}
	data := m.coder.EncodeInterrogationCommand(m.config.CommonAddress, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})

	m.mu.Lock()
	m.stats.TotalGISent++
	m.stats.LastGIStartTime = time.Now()
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 总召唤(等待完成)
func (m *Master) InterrogationWait(timeout time.Duration) error {
	return m.InterrogationWaitWithQOI(QOI_STATION_INTERROGATION, timeout)
}

// 总召唤(等待完成,指定组)
func (m *Master) InterrogationWaitWithQOI(qoi uint8, timeout time.Duration) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	// 创建等待通道
	ch := make(chan *ASDU, 1)
	m.interrogationMu.Lock()
	m.interrogationWaiters[qoi] = ch
	m.interrogationMu.Unlock()

	defer func() {
		m.interrogationMu.Lock()
		delete(m.interrogationWaiters, qoi)
		m.interrogationMu.Unlock()
	}()

	// 发送召唤
	if err := m.InterrogationWithQOI(qoi); err != nil {
		return err
	}

	// 等待响应
	select {
	case <-time.After(timeout):
		m.mu.Lock()
		m.stats.TotalGIFailed++
		m.mu.Unlock()
		return errors.New("interrogation timeout")
	case asdu := <-ch:
		if asdu.COT.Cause == COT_ACTIVATION_TERMINATION {
			m.mu.Lock()
			m.stats.TotalGISuccess++
			m.stats.LastGIEndTime = time.Now()
			m.mu.Unlock()
			return nil
		}
		if asdu.COT.IsPN {
			m.mu.Lock()
			m.stats.TotalGIFailed++
			m.mu.Unlock()
			return errors.New("interrogation negative confirmation")
		}
		return nil
	}
}

// ==================== 电度召唤 ====================

// 电度召唤
func (m *Master) CounterInterrogation() error {
	return m.CounterInterrogationWithQCC(QCC_GROUP_1)
}

// 电度召唤(指定组)
func (m *Master) CounterInterrogationWithQCC(qcc uint8) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &CounterInterrogationCommand{QCC: qcc}
	data := m.coder.EncodeCounterInterrogationCommand(m.config.CommonAddress, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})

	return m.conn.sendIFrame(data)
}

// 电度召唤(等待完成)
func (m *Master) CounterInterrogationWait(timeout time.Duration) error {
	return m.CounterInterrogationWaitWithQCC(QCC_GROUP_1, timeout)
}

// 电度召唤(等待完成,指定组)
func (m *Master) CounterInterrogationWaitWithQCC(qcc uint8, timeout time.Duration) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	ch := make(chan *ASDU, 1)
	m.interrogationMu.Lock()
	m.interrogationWaiters[qcc] = ch
	m.interrogationMu.Unlock()

	defer func() {
		m.interrogationMu.Lock()
		delete(m.interrogationWaiters, qcc)
		m.interrogationMu.Unlock()
	}()

	if err := m.CounterInterrogationWithQCC(qcc); err != nil {
		return err
	}

	select {
	case <-time.After(timeout):
		return errors.New("counter interrogation timeout")
	case asdu := <-ch:
		if asdu.COT.Cause == COT_ACTIVATION_TERMINATION {
			return nil
		}
		if asdu.COT.IsPN {
			return errors.New("counter interrogation negative confirmation")
		}
		return nil
	}
}

// ==================== 时钟同步 ====================

// 时钟同步
func (m *Master) ClockSync() error {
	return m.ClockSyncWithTime(time.Now())
}

// 时钟同步(指定时间)
func (m *Master) ClockSyncWithTime(t time.Time) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &ClockSyncCommand{Time: t}
	data := m.coder.EncodeClockSyncCommand(m.config.CommonAddress, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})

	return m.conn.sendIFrame(data)
}

// ==================== 遥控命令 ====================

// 单命令遥控
func (m *Master) SingleCommand(infoAddr uint32, on bool) error {
	return m.SingleCommandWithSelect(infoAddr, on, false, 0)
}

// 单命令遥控(带选择)
func (m *Master) SingleCommandWithSelect(infoAddr uint32, on bool, selectFlag bool, qu uint8) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &SingleCommand{
		Select: selectFlag,
		QU:     qu,
		On:     on,
	}

	cot := COT_ACTIVATION
	if selectFlag {
		cot = COT_ACTIVATION
	}

	data := m.coder.EncodeSingleCommand(m.config.CommonAddress, infoAddr, cmd,
		CauseOfTransmission{Cause: cot})

	m.mu.Lock()
	m.stats.TotalControlSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 单命令遥控(选择-执行流程)
func (m *Master) SingleCommandSelectExecute(infoAddr uint32, on bool, timeout time.Duration) error {
	// 选择阶段
	if err := m.SingleCommandWithSelect(infoAddr, on, true, 0); err != nil {
		return err
	}

	// 等待选择确认
	time.Sleep(100 * time.Millisecond)

	// 执行阶段
	return m.SingleCommandWithSelect(infoAddr, on, false, 0)
}

// 双命令遥控
func (m *Master) DoubleCommand(infoAddr uint32, state uint8) error {
	return m.DoubleCommandWithSelect(infoAddr, state, false, 0)
}

// 双命令遥控(带选择)
func (m *Master) DoubleCommandWithSelect(infoAddr uint32, state uint8, selectFlag bool, qu uint8) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &DoubleCommand{
		Select: selectFlag,
		QU:     qu,
		State:  state,
	}

	data := m.coder.EncodeDoubleCommand(m.config.CommonAddress, infoAddr, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})

	m.mu.Lock()
	m.stats.TotalControlSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 双命令遥控(选择-执行流程)
func (m *Master) DoubleCommandSelectExecute(infoAddr uint32, state uint8, timeout time.Duration) error {
	// 选择阶段
	if err := m.DoubleCommandWithSelect(infoAddr, state, true, 0); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	// 执行阶段
	return m.DoubleCommandWithSelect(infoAddr, state, false, 0)
}

// 升降命令
func (m *Master) RegulatingStepCommand(infoAddr uint32, state uint8) error {
	return m.RegulatingStepCommandWithSelect(infoAddr, state, false, 0)
}

// 升降命令(带选择)
func (m *Master) RegulatingStepCommandWithSelect(infoAddr uint32, state uint8, selectFlag bool, qu uint8) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &RegulatingStepCommand{
		Select: selectFlag,
		QU:     qu,
		State:  state,
	}

	data := m.coder.EncodeRegulatingStepCommand(m.config.CommonAddress, infoAddr, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})

	m.mu.Lock()
	m.stats.TotalControlSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 设点命令-归一化值
func (m *Master) SetPointCommandNormal(infoAddr uint32, value float32, selectFlag bool) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &SetPointCommand{
		Select: selectFlag,
		Value: &NormalizedValue{
			Value: value,
		},
	}

	data, err := m.coder.EncodeSetPointCommandNormal(m.config.CommonAddress, infoAddr, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.stats.TotalControlSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 设点命令-标度化值
func (m *Master) SetPointCommandScaled(infoAddr uint32, value int16, selectFlag bool) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &SetPointCommand{
		Select: selectFlag,
		Value: &ScaledValue{
			Value: value,
		},
	}

	data, err := m.coder.EncodeSetPointCommandScaled(m.config.CommonAddress, infoAddr, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.stats.TotalControlSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 设点命令-短浮点数
func (m *Master) SetPointCommandFloat(infoAddr uint32, value float32, selectFlag bool) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	cmd := &SetPointCommand{
		Select: selectFlag,
		Value: &FloatValue{
			Value: value,
		},
	}

	data, err := m.coder.EncodeSetPointCommandFloat(m.config.CommonAddress, infoAddr, cmd,
		CauseOfTransmission{Cause: COT_ACTIVATION})
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.stats.TotalControlSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// ==================== 数据转发 ====================

// 发送ASDU
func (m *Master) SendASDU(asdu *ASDU) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}

	data, err := m.encodeASDU(asdu)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.stats.TotalASDUSent++
	m.mu.Unlock()

	return m.conn.sendIFrame(data)
}

// 编码ASDU
func (m *Master) encodeASDU(asdu *ASDU) ([]byte, error) {
	header := m.coder.EncodeASDUHeader(asdu.TypeID, asdu.VSQ, asdu.COT, asdu.CommonAddr)
	data := header

	for _, obj := range asdu.Information {
		infoData, err := m.encodeInfoObject(asdu.TypeID, obj)
		if err != nil {
			return nil, err
		}
		data = append(data, infoData...)
	}

	return data, nil
}

// 编码信息体
func (m *Master) encodeInfoObject(typeID uint8, obj InformationObject) ([]byte, error) {
	data := m.coder.EncodeInfoAddress(obj.Address)

	switch v := obj.Value.(type) {
	case *SinglePointInfo:
		var siq byte
		if v.Value {
			siq = 0x01
		}
		siq |= EncodeQuality(v.Quality)
		data = append(data, siq)
	case *DoublePointInfo:
		var diq byte = v.Value & 0x03
		diq |= EncodeQuality(v.Quality)
		data = append(data, diq)
	case *NormalizedValue:
		intVal := int16(v.Value * 32767)
		data = append(data, byte(intVal), byte(intVal>>8))
		data = append(data, EncodeMeasureQuality(v.Quality))
	case *ScaledValue:
		data = append(data, byte(v.Value), byte(v.Value>>8))
		data = append(data, EncodeMeasureQuality(v.Quality))
	case *FloatValue:
		bits := uint32(0)
		if v.Value != 0 {
			bits = uint32(int32(v.Value * 1000))
		}
		data = append(data, byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24))
		data = append(data, EncodeMeasureQuality(v.Quality))
	default:
		return nil, fmt.Errorf("unsupported value type: %T", v)
	}

	// 添加时标
	if HasTimestamp(typeID) && !obj.Timestamp.IsZero() {
		if HasCP56Time(typeID) {
			data = append(data, EncodeCP56Time2a(obj.Timestamp)...)
		} else {
			data = append(data, EncodeCP24Time2a(obj.Timestamp)...)
		}
	}

	return data, nil
}

// ==================== 统计 ====================

// 获取统计
func (m *Master) GetStats() MasterStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// 重置统计
func (m *Master) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats = MasterStats{}
}

// ==================== 连接管理 ====================

// 获取连接状态
func (m *Master) GetConnectionState() ConnectionState {
	if m.conn == nil {
		return STATE_DISCONNECTED
	}
	return m.conn.GetState()
}

// 获取连接统计
func (m *Master) GetConnectionStats() ConnectionStats {
	if m.conn == nil {
		return ConnectionStats{}
	}
	return m.conn.GetStats()
}

// ==================== 辅助方法 ====================

// 获取公共地址
func (m *Master) GetCommonAddress() uint16 {
	return m.config.CommonAddress
}

// 设置公共地址
func (m *Master) SetCommonAddress(addr uint16) {
	m.config.CommonAddress = addr
	if m.conn != nil {
		m.conn.SetCommonAddress(addr)
	}
}

// 等待连接
func (m *Master) WaitForConnection(timeout time.Duration) error {
	if m.IsConnected() {
		return nil
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)
	for {
		select {
		case <-ticker.C:
			if m.IsConnected() {
				return nil
			}
			if time.Now().After(deadline) {
				return errors.New("wait for connection timeout")
			}
		case <-m.ctx.Done():
			return errors.New("master stopped")
		}
	}
}
