package errors

import (
	"fmt"
	"runtime"
	"sync"
)

type ErrorCode int

const (
	ErrSuccess ErrorCode = 0
	ErrUnknown ErrorCode = 1

	ErrInvalidParam ErrorCode = 100
	ErrUnauthorized ErrorCode = 401
	ErrForbidden ErrorCode = 403
	ErrNotFound ErrorCode = 404
	ErrConflict ErrorCode = 409
	ErrInternalServer ErrorCode = 500
	ErrServiceUnavailable ErrorCode = 503
	ErrTimeout ErrorCode = 504

	ErrUserNotFound ErrorCode = 1001
	ErrUserAlreadyExists ErrorCode = 1002
	ErrInvalidPassword ErrorCode = 1003
	ErrInvalidToken ErrorCode = 1004
	ErrTokenExpired ErrorCode = 1005
	ErrInsufficientPermission ErrorCode = 1006

	ErrDeviceNotFound ErrorCode = 2001
	ErrDeviceOffline ErrorCode = 2002
	ErrDeviceCommunication ErrorCode = 2003
	ErrDeviceAlreadyExists ErrorCode = 2004
	ErrDeviceTypeMismatch ErrorCode = 2005

	ErrStationNotFound ErrorCode = 2101
	ErrStationAlreadyExists ErrorCode = 2102
	ErrStationOffline ErrorCode = 2103

	ErrPointNotFound ErrorCode = 3001
	ErrPointAlreadyExists ErrorCode = 3002
	ErrDataQuality ErrorCode = 3003
	ErrCollectTimeout ErrorCode = 3004
	ErrInvalidPointValue ErrorCode = 3005

	ErrAlarmNotFound ErrorCode = 4001
	ErrAlarmRuleInvalid ErrorCode = 4002
	ErrAlarmAlreadyAcknowledged ErrorCode = 4003
	ErrAlarmAlreadyCleared ErrorCode = 4004

	ErrRuleNotFound ErrorCode = 5001
	ErrRuleAlreadyExists ErrorCode = 5002
	ErrRuleInvalidExpression ErrorCode = 5003

	ErrConfigNotFound ErrorCode = 6001
	ErrConfigAlreadyExists ErrorCode = 6002
	ErrConfigInvalid ErrorCode = 6003

	ErrDatabaseError ErrorCode = 7001
	ErrCacheError ErrorCode = 7002
	ErrQueueError ErrorCode = 7003
	ErrNetworkError ErrorCode = 7004

	ErrRateLimitExceeded ErrorCode = 8001
	ErrResourceExhausted ErrorCode = 8002
)

var errorMessages = map[ErrorCode]string{
	ErrSuccess:                   "操作成功",
	ErrUnknown:                   "未知错误",
	ErrInvalidParam:              "无效参数",
	ErrUnauthorized:              "未授权",
	ErrForbidden:                 "禁止访问",
	ErrNotFound:                 "资源不存在",
	ErrConflict:                  "资源冲突",
	ErrInternalServer:            "服务器内部错误",
	ErrServiceUnavailable:         "服务不可用",
	ErrTimeout:                   "操作超时",
	ErrUserNotFound:              "用户不存在",
	ErrUserAlreadyExists:         "用户已存在",
	ErrInvalidPassword:           "密码错误",
	ErrInvalidToken:              "无效Token",
	ErrTokenExpired:              "Token已过期",
	ErrInsufficientPermission:    "权限不足",
	ErrDeviceNotFound:            "设备不存在",
	ErrDeviceOffline:              "设备离线",
	ErrDeviceCommunication:        "设备通信失败",
	ErrDeviceAlreadyExists:       "设备已存在",
	ErrDeviceTypeMismatch:         "设备类型不匹配",
	ErrStationNotFound:           "电站不存在",
	ErrStationAlreadyExists:       "电站已存在",
	ErrStationOffline:             "电站离线",
	ErrPointNotFound:             "采集点不存在",
	ErrPointAlreadyExists:        "采集点已存在",
	ErrDataQuality:               "数据质量异常",
	ErrCollectTimeout:            "采集超时",
	ErrInvalidPointValue:         "无效的测点值",
	ErrAlarmNotFound:             "告警不存在",
	ErrAlarmRuleInvalid:           "告警规则无效",
	ErrAlarmAlreadyAcknowledged:   "告警已确认",
	ErrAlarmAlreadyCleared:       "告警已清除",
	ErrRuleNotFound:              "规则不存在",
	ErrRuleAlreadyExists:          "规则已存在",
	ErrRuleInvalidExpression:       "规则表达式无效",
	ErrConfigNotFound:             "配置不存在",
	ErrConfigAlreadyExists:        "配置已存在",
	ErrConfigInvalid:              "配置无效",
	ErrDatabaseError:             "数据库错误",
	ErrCacheError:                "缓存错误",
	ErrQueueError:                "消息队列错误",
	ErrNetworkError:              "网络错误",
	ErrRateLimitExceeded:         "请求过于频繁",
	ErrResourceExhausted:         "资源耗尽",
}

func (c ErrorCode) String() string {
	if msg, ok := errorMessages[c]; ok {
		return msg
	}
	return "未知错误"
}

func (c ErrorCode) IsSuccess() bool {
	return c == ErrSuccess
}

func (c ErrorCode) IsClientError() bool {
	return c >= 100 && c < 500
}

func (c ErrorCode) IsServerError() bool {
	return c >= 500
}

func (c ErrorCode) IsBusinessError() bool {
	return c >= 1000 && c < 9000
}
