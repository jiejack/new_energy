package iec104

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// APDU类型定义
const (
	// 启动字节
	START_BYTE = 0x68

	// U帧控制域类型
	U_FRAME_STARTDT_ACT = 0x07 // 启动数据传输激活
	U_FRAME_STARTDT_CON = 0x0B // 启动数据传输确认
	U_FRAME_STOPDT_ACT  = 0x13 // 停止数据传输激活
	U_FRAME_STOPDT_CON  = 0x23 // 停止数据传输确认
	U_FRAME_TESTFR_ACT  = 0x43 // 测试帧激活
	U_FRAME_TESTFR_CON  = 0x83 // 测试帧确认

	// I帧标识
	I_FRAME_MASK = 0x01

	// S帧标识
	S_FRAME_MASK = 0x03
)

// 类型标识 (Type Identification)
const (
	// 过程信息在监视方向
	TYPE_ID_SINGLE_POINT_INFO              = 1  // 单点遥信
	TYPE_ID_SINGLE_POINT_INFO_TIME         = 2  // 单点遥信带时标
	TYPE_ID_DOUBLE_POINT_INFO              = 3  // 双点遥信
	TYPE_ID_DOUBLE_POINT_INFO_TIME         = 4  // 双点遥信带时标
	TYPE_ID_STEP_POSITION_INFO             = 5  // 步位置信息
	TYPE_ID_STEP_POSITION_INFO_TIME        = 6  // 步位置信息带时标
	TYPE_ID_BITSTRING32                    = 7  // 32位串
	TYPE_ID_BITSTRING32_TIME               = 8  // 32位串带时标
	TYPE_ID_MEASURE_VALUE_NORMAL           = 9  // 测量值-归一化值
	TYPE_ID_MEASURE_VALUE_NORMAL_TIME      = 10 // 测量值-归一化值带时标
	TYPE_ID_MEASURE_VALUE_SCALED           = 11 // 测量值-标度化值
	TYPE_ID_MEASURE_VALUE_SCALED_TIME      = 12 // 测量值-标度化值带时标
	TYPE_ID_MEASURE_VALUE_FLOAT            = 13 // 测量值-短浮点数
	TYPE_ID_MEASURE_VALUE_FLOAT_TIME       = 14 // 测量值-短浮点数带时标
	TYPE_ID_INTEGRITY_TOTAL                = 15 // 累积量
	TYPE_ID_INTEGRITY_TOTAL_TIME           = 16 // 累积量带时标
	TYPE_ID_EVENT_OF_PROTECTION            = 17 // 保护装置事件
	TYPE_ID_EVENT_OF_PROTECTION_TIME       = 18 // 保护装置事件带时标
	TYPE_ID_PACKED_SINGLE_POINT_INFO       = 19 // 打包单点信息
	TYPE_ID_PACKED_SINGLE_POINT_INFO_TIME  = 20 // 打包单点信息带时标
	TYPE_ID_PACKED_DOUBLE_POINT_INFO       = 21 // 打包双点信息
	TYPE_ID_PACKED_DOUBLE_POINT_INFO_TIME  = 22 // 打包双点信息带时标
	TYPE_ID_PACKED_MEASURE_NORMAL          = 23 // 打包测量值-归一化值
	TYPE_ID_PACKED_MEASURE_NORMAL_TIME     = 24 // 打包测量值-归一化值带时标
	TYPE_ID_PACKED_MEASURE_SCALED          = 25 // 打包测量值-标度化值
	TYPE_ID_PACKED_MEASURE_SCALED_TIME     = 26 // 打包测量值-标度化值带时标
	TYPE_ID_PACKED_MEASURE_FLOAT           = 27 // 打包测量值-短浮点数
	TYPE_ID_PACKED_MEASURE_FLOAT_TIME      = 28 // 打包测量值-短浮点数带时标
	TYPE_ID_SINGLE_POINT_INFO_TIME_CP56    = 30 // 单点遥信带CP56时标
	TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56    = 31 // 双点遥信带CP56时标
	TYPE_ID_STEP_POSITION_INFO_TIME_CP56   = 32 // 步位置信息带CP56时标
	TYPE_ID_BITSTRING32_TIME_CP56          = 33 // 32位串带CP56时标
	TYPE_ID_MEASURE_VALUE_NORMAL_TIME_CP56 = 35 // 测量值-归一化值带CP56时标
	TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56 = 36 // 测量值-标度化值带CP56时标
	TYPE_ID_MEASURE_VALUE_FLOAT_TIME_CP56  = 37 // 测量值-短浮点数带CP56时标
	TYPE_ID_INTEGRITY_TOTAL_TIME_CP56      = 38 // 累积量带CP56时标

	// 系统信息在监视方向
	TYPE_ID_END_OF_INITIALIZATION = 70 // 初始化结束

	// 过程信息在控制方向
	TYPE_ID_SINGLE_COMMAND              = 45 // 单命令
	TYPE_ID_DOUBLE_COMMAND              = 46 // 双命令
	TYPE_ID_REGULATING_STEP_COMMAND     = 47 // 升降命令
	TYPE_ID_SET_POINT_COMMAND_NORMAL    = 48 // 设点命令-归一化值
	TYPE_ID_SET_POINT_COMMAND_SCALED    = 49 // 设点命令-标度化值
	TYPE_ID_SET_POINT_COMMAND_FLOAT     = 50 // 设点命令-短浮点数
	TYPE_ID_BITSTRING32_COMMAND         = 51 // 32位串命令

	// 系统信息在控制方向
	TYPE_ID_INTERROGATION_CMD            = 100 // 总召唤命令
	TYPE_ID_COUNTER_INTERROGATION_CMD    = 101 // 电度召唤命令
	TYPE_ID_READ_COMMAND                 = 102 // 读命令
	TYPE_ID_CLOCK_SYNC_CMD               = 103 // 时钟同步命令
	TYPE_ID_TEST_COMMAND                 = 104 // 测试命令
	TYPE_ID_RESET_PROCESS_CMD            = 105 // 复位进程命令
	TYPE_ID_DELAY_ACQUISITION_CMD        = 106 // 延迟获取命令
	TYPE_ID_PARAMETER_ACTIVATION         = 113 // 参数激活

	// 参数设置在控制方向
	TYPE_ID_MEASURE_VALUE_NORMAL_PARAM   = 110 // 测量值参数-归一化值
	TYPE_ID_MEASURE_VALUE_SCALED_PARAM   = 111 // 测量值参数-标度化值
	TYPE_ID_MEASURE_VALUE_FLOAT_PARAM    = 112 // 测量值参数-短浮点数

	// 文件传输
	TYPE_ID_FILE_READY                   = 120 // 文件准备就绪
	TYPE_ID_SECTION_READY                = 121 // 节准备就绪
	TYPE_ID_FILE_CALL_OR_SELECT          = 122 // 文件调用或选择
	TYPE_ID_FILE_SEGMENT                 = 123 // 文件段
	TYPE_ID_DIRECTORY_OR_SECTION_ACK     = 124 // 目录或节确认
	TYPE_ID_FILE_TRANSFER_ACK            = 125 // 文件传输确认
)

// 传输原因 (Cause of Transmission)
const (
	COT_PERIODIC_CYCLIC                    = 1  // 周期/循环
	COT_BACKGROUND_SCAN                    = 2  // 背景扫描
	COT_SPONTANEOUS                        = 3  // 突发(自发)
	COT_INITIALIZED                        = 4  // 初始化
	COT_REQUEST_OR_REQUESTED               = 5  // 请求或被请求
	COT_ACTIVATION                         = 6  // 激活
	COT_ACTIVATION_CON                     = 7  // 激活确认
	COT_DEACTIVATION                       = 8  // 停止激活
	COT_DEACTIVATION_CON                   = 9  // 停止激活确认
	COT_ACTIVATION_TERMINATION             = 10 // 激活终止
	COT_RETURN_INFO_CAUSE_REMOTE_COMMAND   = 11 // 远方命令引起的返送信息
	COT_RETURN_INFO_CAUSE_LOCAL_COMMAND    = 12 // 当地命令引起的返送信息
	COT_FILE_TRANSFER                      = 13 // 文件传输
	COT_RESERVED_14                        = 14 // 保留
	COT_RESERVED_15                        = 15 // 保留
	COT_INTERROGATED_BY_STATION            = 20 // 站召唤
	COT_INTERROGATED_BY_GROUP_1            = 21 // 第1组召唤
	COT_INTERROGATED_BY_GROUP_2            = 22 // 第2组召唤
	COT_INTERROGATED_BY_GROUP_3            = 23 // 第3组召唤
	COT_INTERROGATED_BY_GROUP_4            = 24 // 第4组召唤
	COT_INTERROGATED_BY_GROUP_5            = 25 // 第5组召唤
	COT_INTERROGATED_BY_GROUP_6            = 26 // 第6组召唤
	COT_INTERROGATED_BY_GROUP_7            = 27 // 第7组召唤
	COT_INTERROGATED_BY_GROUP_8            = 28 // 第8组召唤
	COT_INTERROGATED_BY_GROUP_9            = 29 // 第9组召唤
	COT_INTERROGATED_BY_GROUP_10           = 30 // 第10组召唤
	COT_INTERROGATED_BY_GROUP_11           = 31 // 第11组召唤
	COT_INTERROGATED_BY_GROUP_12           = 32 // 第12组召唤
	COT_INTERROGATED_BY_GROUP_13           = 33 // 第13组召唤
	COT_INTERROGATED_BY_GROUP_14           = 34 // 第14组召唤
	COT_INTERROGATED_BY_GROUP_15           = 35 // 第15组召唤
	COT_INTERROGATED_BY_GROUP_16           = 36 // 第16组召唤
	COT_COUNTER_INTERROGATED_BY_GROUP_1    = 37 // 计数量召唤第1组
	COT_COUNTER_INTERROGATED_BY_GROUP_2    = 38 // 计数量召唤第2组
	COT_COUNTER_INTERROGATED_BY_GROUP_3    = 39 // 计数量召唤第3组
	COT_COUNTER_INTERROGATED_BY_GROUP_4    = 40 // 计数量召唤第4组
)

// 总召唤限定词
const (
	QOI_STATION_INTERROGATION = 20 // 站召唤
	QOI_GROUP_1               = 21 // 第1组
	QOI_GROUP_2               = 22 // 第2组
	QOI_GROUP_3               = 23 // 第3组
	QOI_GROUP_4               = 24 // 第4组
)

// 电度召唤限定词
const (
	QCC_GROUP_1 = 1 // 第1组计数量
	QCC_GROUP_2 = 2 // 第2组计数量
	QCC_GROUP_3 = 3 // 第3组计数量
	QCC_GROUP_4 = 4 // 第4组计数量
)

// 质量描述词
const (
	QDS_IV = 0x80 // 无效
	QDS_NT = 0x40 // 非当前值
	QDS_SB = 0x20 // 被取代
	QDS_BL = 0x10 // 被闭锁
)

// 遥信质量描述
const (
	SIQ_IV = 0x80 // 无效
	SIQ_NT = 0x40 // 非当前值
	SIQ_SB = 0x20 // 被取代
	SIQ_BL = 0x10 // 被闭锁
)

// 双点遥信状态
const (
	DP_INDETERMINATE = 0 // 不确定
	DP_OFF           = 1 // 分
	DP_ON            = 2 // 合
	DP_INDETERMINATE2 = 3 // 不确定
)

// 遥测质量描述
const (
	QDS_OVERFLOW = 0x01 // 溢出
)

// 单命令限定词
const (
	SCS_SELECT   = 0x80 // 选择
	SCS_QU_0     = 0x00 // 无额外定义
	SCS_QU_1     = 0x01 // 短脉冲持续时间
	SCS_QU_2     = 0x02 // 持续脉冲
	SCS_QU_3     = 0x03 // 持续输出
)

// 双命令限定词
const (
	DCS_SELECT = 0x80 // 选择
	DCS_OFF    = 0x00 // 分
	DCS_ON     = 0x01 // 合
)

// 升降命令限定词
const (
	RCS_SELECT = 0x80 // 选择
	RCS_STOP   = 0x00 // 停止
	RCS_LOWER  = 0x01 // 降
	RCS_HIGHER = 0x02 // 升
)

// APDU类型枚举
type APDUType int

const (
	APDU_TYPE_I APDUType = iota // I帧 - 信息传输帧
	APDU_TYPE_S                 // S帧 - 监视帧
	APDU_TYPE_U                 // U帧 - 无编号控制帧
)

// 连接状态
type ConnectionState int

const (
	STATE_DISCONNECTED ConnectionState = iota
	STATE_CONNECTING
	STATE_CONNECTED
	STATE_STARTDT_SENT
	STATE_ACTIVE
	STATE_STOPPING
)

func (s ConnectionState) String() string {
	switch s {
	case STATE_DISCONNECTED:
		return "DISCONNECTED"
	case STATE_CONNECTING:
		return "CONNECTING"
	case STATE_CONNECTED:
		return "CONNECTED"
	case STATE_STARTDT_SENT:
		return "STARTDT_SENT"
	case STATE_ACTIVE:
		return "ACTIVE"
	case STATE_STOPPING:
		return "STOPPING"
	default:
		return "UNKNOWN"
	}
}

// APDU 应用协议数据单元
type APDU struct {
	Type     APDUType
	SendSeq  uint16 // 发送序号 (I帧)
	RecvSeq  uint16 // 接收序号 (I帧/S帧)
	Control  byte   // U帧控制域
	ASDU     *ASDU  // 应用服务数据单元 (I帧)
	RawBytes []byte // 原始字节
}

// ASDU 应用服务数据单元
type ASDU struct {
	TypeID       uint8               // 类型标识
	VSQ          VSQ                 // 可变结构限定词
	COT          CauseOfTransmission // 传输原因
	Origin       uint8               // 源发地址
	CommonAddr   uint16              // 公共地址
	Information  []InformationObject // 信息体
	IsSequence   bool                // 是否为序列
}

// VSQ 可变结构限定词
type VSQ struct {
	Number     uint8 // 信息体数目
	IsSequence bool  // 是否为序列 (SQ位)
}

// CauseOfTransmission 传输原因
type CauseOfTransmission struct {
	Cause   uint8 // 传输原因
	IsTest  bool  // 测试标志 (T)
	IsPN    bool  // 肯定/否定确认 (P/N)
}

// InformationObject 信息体
type InformationObject struct {
	Address   uint32      // 信息体地址
	Value     interface{} // 值
	Quality   Quality     // 质量描述
	Timestamp time.Time   // 时标 (可选)
}

// Quality 质量描述
type Quality struct {
	Invalid    bool // 无效 (IV)
	NotCurrent bool // 非当前值 (NT)
	Substituted bool // 被取代 (SB)
	Blocked    bool // 被闭锁 (BL)
	Overflow   bool // 溢出 (OV) - 遥测专用
}

// SinglePointInfo 单点遥信
type SinglePointInfo struct {
	Value   bool // 开关状态
	Quality Quality
}

// DoublePointInfo 双点遥信
type DoublePointInfo struct {
	Value   uint8 // 0:不确定 1:分 2:合 3:不确定
	Quality Quality
}

// StepPositionInfo 步位置信息
type StepPositionInfo struct {
	Value    int16 // 步位置值 (-64~+63)
	Transient bool // 瞬态标志
	Quality   Quality
}

// Bitstring32 32位串
type Bitstring32 struct {
	Value   uint32
	Quality Quality
}

// NormalizedValue 归一化值 (-1.0 ~ 1.0)
type NormalizedValue struct {
	Value   float32
	Quality Quality
}

// ScaledValue 标度化值 (-32768 ~ 32767)
type ScaledValue struct {
	Value   int16
	Quality Quality
}

// FloatValue 短浮点数
type FloatValue struct {
	Value   float32
	Quality Quality
}

// IntegratedTotal 累积量(电度值)
type IntegratedTotal struct {
	Value   uint32
	Sequence uint8 // 顺序号
	Quality  Quality
}

// SingleCommand 单命令
type SingleCommand struct {
	Select  bool  // 选择标志
	QU      uint8 // 限定词
	On      bool  // 合/分
}

// DoubleCommand 双命令
type DoubleCommand struct {
	Select bool  // 选择标志
	QU     uint8 // 限定词
	State  uint8 // 0:不允许 1:分 2:合 3:不允许
}

// RegulatingStepCommand 升降命令
type RegulatingStepCommand struct {
	Select bool  // 选择标志
	QU     uint8 // 限定词
	State  uint8 // 0:停止 1:降 2:升 3:禁止
}

// SetPointCommand 设点命令
type SetPointCommand struct {
	Select bool
	Value  interface{} // NormalizedValue/ScaledValue/FloatValue
}

// InterrogationCommand 总召唤命令
type InterrogationCommand struct {
	QOI uint8 // 限定词
}

// CounterInterrogationCommand 电度召唤命令
type CounterInterrogationCommand struct {
	QCC uint8 // 限定词
}

// ClockSyncCommand 时钟同步命令
type ClockSyncCommand struct {
	Time time.Time
}

// TestCommand 测试命令
type TestCommand struct {
	FFS1 uint8 // 固定值1
	FFS2 uint8 // 固定值2
}

// EndOfInitialization 初始化结束
type EndOfInitialization struct {
	COI uint8 // 初始化原因
}

// CP24Time2a 3字节二进制时间
type CP24Time2a struct {
	Milliseconds uint16 // 毫秒 (0-59999)
	Minutes      uint8  // 分钟 (0-59)
	Reserved     bool   // 备用位
	SummerTime   bool   // 夏时制标志
}

// CP56Time2a 7字节二进制时间
type CP56Time2a struct {
	Milliseconds uint16 // 毫秒 (0-59999)
	Minutes      uint8  // 分钟 (0-59)
	Hours        uint8  // 小时 (0-23)
	DayOfWeek    uint8  // 星期 (1-7)
	DayOfMonth   uint8  // 日 (1-31)
	Month        uint8  // 月 (1-12)
	Year         uint8  // 年 (0-99)
	SummerTime   bool   // 夏时制标志
}

// 编码CP24Time2a
func EncodeCP24Time2a(t time.Time) []byte {
	data := make([]byte, 3)
	
	// 安全转换：确保毫秒值在uint16范围内
	ms := uint16(min(t.Nanosecond()/1000000+int(t.Second())*1000, math.MaxUint16))
	binary.LittleEndian.PutUint16(data[0:2], ms)
	
	data[2] = uint8(t.Minute()) & 0x3F
	// 备用位和夏时制标志未实现
	
	return data
}

// 解码CP24Time2a
func DecodeCP24Time2a(data []byte) CP24Time2a {
	if len(data) < 3 {
		return CP24Time2a{}
	}
	
	ms := binary.LittleEndian.Uint16(data[0:2])
	
	return CP24Time2a{
		Milliseconds: ms,
		Minutes:      data[2] & 0x3F,
		Reserved:     (data[2] & 0x40) != 0,
		SummerTime:   (data[2] & 0x80) != 0,
	}
}

// 编码CP56Time2a
func EncodeCP56Time2a(t time.Time) []byte {
	data := make([]byte, 7)
	
	// 安全转换：确保毫秒值在uint16范围内
	ms := uint16(min(t.Nanosecond()/1000000+int(t.Second())*1000, math.MaxUint16))
	binary.LittleEndian.PutUint16(data[0:2], ms)
	
	data[2] = uint8(t.Minute()) & 0x3F
	data[3] = uint8(t.Hour()) & 0x1F
	data[4] = uint8(t.Weekday())
	if data[4] == 0 {
		data[4] = 7 // Go的Sunday是0，IEC104中Sunday是7
	}
	data[5] = uint8(t.Day()) & 0x1F
	data[6] = uint8(t.Month()) & 0x0F
	
	year := t.Year() % 100
	data[6] |= byte(year << 4)
	
	// 夏时制标志未实现
	
	return data
}

// 解码CP56Time2a
func DecodeCP56Time2a(data []byte) (CP56Time2a, error) {
	if len(data) < 7 {
		return CP56Time2a{}, fmt.Errorf("insufficient data for CP56Time2a: need 7, got %d", len(data))
	}
	
	ms := binary.LittleEndian.Uint16(data[0:2])
	year := int(data[6] >> 4)
	
	return CP56Time2a{
		Milliseconds: ms,
		Minutes:      data[2] & 0x3F,
		Hours:        data[3] & 0x1F,
		DayOfWeek:    data[4] & 0x07,
		DayOfMonth:   data[5] & 0x1F,
		Month:        data[6] & 0x0F,
		Year:         uint8(year),
		SummerTime:   (data[3] & 0x80) != 0,
	}, nil
}

// CP56Time2a转time.Time
func (cp CP56Time2a) ToTime() time.Time {
	year := int(cp.Year)
	if year < 70 {
		year += 2000
	} else {
		year += 1900
	}
	
	sec := int(cp.Milliseconds / 1000)
	nsec := (int(cp.Milliseconds) % 1000) * 1000000
	
	return time.Date(year, time.Month(cp.Month), int(cp.DayOfMonth),
		int(cp.Hours), int(cp.Minutes), sec, nsec, time.Local)
}

// 解析质量描述 (SIQ/DPI等)
func ParseQuality(qds byte) Quality {
	return Quality{
		Invalid:     (qds & QDS_IV) != 0,
		NotCurrent:  (qds & QDS_NT) != 0,
		Substituted: (qds & QDS_SB) != 0,
		Blocked:     (qds & QDS_BL) != 0,
	}
}

// 编码质量描述
func EncodeQuality(q Quality) byte {
	var b byte
	if q.Invalid {
		b |= QDS_IV
	}
	if q.NotCurrent {
		b |= QDS_NT
	}
	if q.Substituted {
		b |= QDS_SB
	}
	if q.Blocked {
		b |= QDS_BL
	}
	return b
}

// 解析遥测质量描述 (QDS)
func ParseMeasureQuality(qds byte) Quality {
	return Quality{
		Invalid:     (qds & QDS_IV) != 0,
		NotCurrent:  (qds & QDS_NT) != 0,
		Substituted: (qds & QDS_SB) != 0,
		Blocked:     (qds & QDS_BL) != 0,
		Overflow:    (qds & QDS_OVERFLOW) != 0,
	}
}

// 编码遥测质量描述
func EncodeMeasureQuality(q Quality) byte {
	var b byte
	if q.Invalid {
		b |= QDS_IV
	}
	if q.NotCurrent {
		b |= QDS_NT
	}
	if q.Substituted {
		b |= QDS_SB
	}
	if q.Blocked {
		b |= QDS_BL
	}
	if q.Overflow {
		b |= QDS_OVERFLOW
	}
	return b
}

// 解析VSQ
func ParseVSQ(vsq byte) VSQ {
	return VSQ{
		Number:     vsq & 0x7F,
		IsSequence: (vsq & 0x80) != 0,
	}
}

// 编码VSQ
func EncodeVSQ(vsq VSQ) byte {
	var b byte = vsq.Number & 0x7F
	if vsq.IsSequence {
		b |= 0x80
	}
	return b
}

// 解析传输原因
func ParseCOT(cot1, cot2 byte) CauseOfTransmission {
	return CauseOfTransmission{
		Cause:  cot1 & 0x3F,
		IsTest: (cot1 & 0x80) != 0,
		IsPN:   (cot1 & 0x40) != 0,
	}
}

// 编码传输原因
func EncodeCOT(cot CauseOfTransmission) (byte, byte) {
	var b byte = cot.Cause & 0x3F
	if cot.IsTest {
		b |= 0x80
	}
	if cot.IsPN {
		b |= 0x40
	}
	return b, 0 // cot2为源发地址，通常为0
}

// 判断是否为遥信类型
func IsSinglePointInfo(typeID uint8) bool {
	return typeID == TYPE_ID_SINGLE_POINT_INFO ||
		typeID == TYPE_ID_SINGLE_POINT_INFO_TIME ||
		typeID == TYPE_ID_SINGLE_POINT_INFO_TIME_CP56 ||
		typeID == TYPE_ID_PACKED_SINGLE_POINT_INFO ||
		typeID == TYPE_ID_PACKED_SINGLE_POINT_INFO_TIME
}

// 判断是否为双点遥信类型
func IsDoublePointInfo(typeID uint8) bool {
	return typeID == TYPE_ID_DOUBLE_POINT_INFO ||
		typeID == TYPE_ID_DOUBLE_POINT_INFO_TIME ||
		typeID == TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56 ||
		typeID == TYPE_ID_PACKED_DOUBLE_POINT_INFO ||
		typeID == TYPE_ID_PACKED_DOUBLE_POINT_INFO_TIME
}

// 判断是否为遥测类型
func IsMeasureValue(typeID uint8) bool {
	return typeID == TYPE_ID_MEASURE_VALUE_NORMAL ||
		typeID == TYPE_ID_MEASURE_VALUE_NORMAL_TIME ||
		typeID == TYPE_ID_MEASURE_VALUE_NORMAL_TIME_CP56 ||
		typeID == TYPE_ID_MEASURE_VALUE_SCALED ||
		typeID == TYPE_ID_MEASURE_VALUE_SCALED_TIME ||
		typeID == TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56 ||
		typeID == TYPE_ID_MEASURE_VALUE_FLOAT ||
		typeID == TYPE_ID_MEASURE_VALUE_FLOAT_TIME ||
		typeID == TYPE_ID_MEASURE_VALUE_FLOAT_TIME_CP56
}

// 判断是否为电度类型
func IsIntegratedTotal(typeID uint8) bool {
	return typeID == TYPE_ID_INTEGRITY_TOTAL ||
		typeID == TYPE_ID_INTEGRITY_TOTAL_TIME ||
		typeID == TYPE_ID_INTEGRITY_TOTAL_TIME_CP56
}

// 判断是否带CP24时标
func HasCP24Time(typeID uint8) bool {
	return typeID == TYPE_ID_SINGLE_POINT_INFO_TIME ||
		typeID == TYPE_ID_DOUBLE_POINT_INFO_TIME ||
		typeID == TYPE_ID_STEP_POSITION_INFO_TIME ||
		typeID == TYPE_ID_BITSTRING32_TIME ||
		typeID == TYPE_ID_MEASURE_VALUE_NORMAL_TIME ||
		typeID == TYPE_ID_MEASURE_VALUE_SCALED_TIME ||
		typeID == TYPE_ID_MEASURE_VALUE_FLOAT_TIME ||
		typeID == TYPE_ID_INTEGRITY_TOTAL_TIME
}

// 判断是否带CP56时标
func HasCP56Time(typeID uint8) bool {
	return typeID == TYPE_ID_SINGLE_POINT_INFO_TIME_CP56 ||
		typeID == TYPE_ID_DOUBLE_POINT_INFO_TIME_CP56 ||
		typeID == TYPE_ID_STEP_POSITION_INFO_TIME_CP56 ||
		typeID == TYPE_ID_BITSTRING32_TIME_CP56 ||
		typeID == TYPE_ID_MEASURE_VALUE_NORMAL_TIME_CP56 ||
		typeID == TYPE_ID_MEASURE_VALUE_SCALED_TIME_CP56 ||
		typeID == TYPE_ID_MEASURE_VALUE_FLOAT_TIME_CP56 ||
		typeID == TYPE_ID_INTEGRITY_TOTAL_TIME_CP56
}

// 判断是否需要时标
func HasTimestamp(typeID uint8) bool {
	return HasCP24Time(typeID) || HasCP56Time(typeID)
}

// 判断是否为控制方向类型
func IsControlDirection(typeID uint8) bool {
	return typeID >= TYPE_ID_SINGLE_COMMAND && typeID <= TYPE_ID_PARAMETER_ACTIVATION
}

// 判断是否为系统信息类型
func IsSystemInfo(typeID uint8) bool {
	return typeID == TYPE_ID_END_OF_INITIALIZATION
}
