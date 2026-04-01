package notifier

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSMSConfigInvalid    = errors.New("sms config is invalid")
	ErrSMSRecipientEmpty   = errors.New("sms recipient is empty")
	ErrSMSTemplateInvalid  = errors.New("sms template is invalid")
	ErrSMSProviderNotSupport = errors.New("sms provider not supported")
	ErrSMSSendFailed       = errors.New("sms send failed")
	ErrSMSRateLimitExceeded = errors.New("sms rate limit exceeded")
)

// SMSNotifier 短信通知器
type SMSNotifier struct {
	config       *NotificationConfig
	smsConfig    *SMSConfig
	httpClient   *http.Client
	rateLimiter  RateLimiter
	templateMgr  TemplateManager
}

// NewSMSNotifier 创建短信通知器
func NewSMSNotifier(config *NotificationConfig, templateMgr TemplateManager) (*SMSNotifier, error) {
	if config == nil || config.SMSConfig == nil {
		return nil, ErrSMSConfigInvalid
	}

	return &SMSNotifier{
		config:      config,
		smsConfig:   config.SMSConfig,
		httpClient:  &http.Client{Timeout: config.Timeout},
		rateLimiter: NewTokenBucketRateLimiter(config.RateLimit, config.BurstLimit),
		templateMgr: templateMgr,
	}, nil
}

// Channel 返回通知渠道类型
func (s *SMSNotifier) Channel() NotificationChannel {
	return ChannelSMS
}

// Send 发送短信通知
func (s *SMSNotifier) Send(ctx context.Context, notification *Notification) (*NotificationResult, error) {
	if err := s.Validate(notification); err != nil {
		return nil, err
	}

	// 限流检查
	key := fmt.Sprintf("sms:%s", notification.AlarmID)
	if !s.rateLimiter.Allow(key) {
		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusFailed,
			Message:        "rate limit exceeded",
			Error:          ErrSMSRateLimitExceeded,
		}, ErrSMSRateLimitExceeded
	}

	// 准备短信内容
	content, err := s.prepareContent(notification)
	if err != nil {
		return nil, err
	}

	// 获取手机号列表
	phones := s.getPhones(notification.Recipients)
	if len(phones) == 0 {
		return nil, ErrSMSRecipientEmpty
	}

	// 根据服务商发送
	var result *NotificationResult
	switch s.smsConfig.Provider {
	case "aliyun":
		result, err = s.sendViaAliyun(ctx, phones, content, notification)
	case "tencent":
		result, err = s.sendViaTencent(ctx, phones, content, notification)
	default:
		return nil, ErrSMSProviderNotSupport
	}

	if err != nil {
		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusFailed,
			Message:        err.Error(),
			Error:          err,
		}, err
	}

	return result, nil
}

// SendBatch 批量发送短信通知
func (s *SMSNotifier) SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, len(notifications))
	for i, notification := range notifications {
		result, err := s.Send(ctx, notification)
		if err != nil {
			results[i] = &NotificationResult{
				NotificationID: notification.ID,
				Success:        false,
				Status:         StatusFailed,
				Message:        err.Error(),
				Error:          err,
			}
		} else {
			results[i] = result
		}
	}
	return results, nil
}

// Validate 验证短信通知
func (s *SMSNotifier) Validate(notification *Notification) error {
	if notification == nil {
		return errors.New("notification is nil")
	}

	if len(notification.Recipients) == 0 {
		return ErrSMSRecipientEmpty
	}

	// 验证手机号
	for _, r := range notification.Recipients {
		if r.Phone == "" {
			return fmt.Errorf("recipient %s has no phone number", r.Name)
		}
		if !isValidPhone(r.Phone) {
			return fmt.Errorf("invalid phone number: %s", r.Phone)
		}
	}

	return nil
}

// HealthCheck 健康检查
func (s *SMSNotifier) HealthCheck(ctx context.Context) error {
	// 简单检查配置是否有效
	if s.smsConfig.AccessKey == "" || s.smsConfig.AccessSecret == "" {
		return ErrSMSConfigInvalid
	}
	return nil
}

// Close 关闭短信通知器
func (s *SMSNotifier) Close() error {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
	}
	return nil
}

// prepareContent 准备短信内容
func (s *SMSNotifier) prepareContent(notification *Notification) (string, error) {
	// 如果有模板，使用模板渲染
	if notification.TemplateID != "" && s.templateMgr != nil {
		rendered, err := s.templateMgr.Render(notification.TemplateID, notification.TemplateData)
		if err != nil {
			return "", err
		}
		return rendered, nil
	}

	// 否则直接使用内容
	return notification.Content, nil
}

// getPhones 获取手机号列表
func (s *SMSNotifier) getPhones(recipients []Recipient) []string {
	phones := make([]string, 0, len(recipients))
	for _, r := range recipients {
		if r.Phone != "" {
			phones = append(phones, r.Phone)
		}
	}
	return phones
}

// sendViaAliyun 通过阿里云发送短信
func (s *SMSNotifier) sendViaAliyun(ctx context.Context, phones []string, content string, notification *Notification) (*NotificationResult, error) {
	// 构建请求参数
	params := map[string]string{
		"Action":        "SendSms",
		"Version":       "2017-05-25",
		"RegionId":      s.smsConfig.Region,
		"PhoneNumbers":  strings.Join(phones, ","),
		"SignName":      s.smsConfig.SignName,
		"TemplateCode":  notification.TemplateID,
	}

	// 如果有模板参数，添加模板参数
	if notification.TemplateData != nil {
		templateParam, err := json.Marshal(notification.TemplateData)
		if err == nil {
			params["TemplateParam"] = string(templateParam)
		}
	}

	// 发送请求
	resp, err := s.sendAliyunRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var aliyunResp AliyunSMSResponse
	if err := json.Unmarshal(resp, &aliyunResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	now := time.Now()
	if aliyunResp.Code == "OK" {
		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        true,
			Status:         StatusSent,
			Message:        "sms sent successfully",
			ExternalID:     aliyunResp.BizId,
			DeliveredAt:    &now,
		}, nil
	}

	return nil, fmt.Errorf("sms send failed: %s - %s", aliyunResp.Code, aliyunResp.Message)
}

// AliyunSMSResponse 阿里云短信响应
type AliyunSMSResponse struct {
	Code     string `json:"Code"`
	Message  string `json:"Message"`
	BizId    string `json:"BizId"`
	RequestId string `json:"RequestId"`
}

// sendAliyunRequest 发送阿里云请求
func (s *SMSNotifier) sendAliyunRequest(ctx context.Context, params map[string]string) ([]byte, error) {
	// 添加公共参数
	params["Format"] = "JSON"
	params["AccessKeyId"] = s.smsConfig.AccessKey
	params["SignatureMethod"] = "HMAC-SHA256"
	params["SignatureVersion"] = "1.0"
	params["SignatureNonce"] = uuid.New().String()
	params["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// 计算签名
	signature := s.calculateAliyunSignature(params)
	params["Signature"] = signature

	// 构建URL
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	endpoint := s.smsConfig.Endpoint
	if endpoint == "" {
		endpoint = "https://dysmsapi.aliyuncs.com"
	}

	reqURL := fmt.Sprintf("%s/?%s", endpoint, values.Encode())

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// calculateAliyunSignature 计算阿里云签名
func (s *SMSNotifier) calculateAliyunSignature(params map[string]string) string {
	// 构建规范化请求字符串
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	canonicalizedQueryString := values.Encode()

	// 构建待签名字符串
	stringToSign := "GET&%2F&" + url.QueryEscape(canonicalizedQueryString)

	// HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(s.smsConfig.AccessSecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

// sendViaTencent 通过腾讯云发送短信
func (s *SMSNotifier) sendViaTencent(ctx context.Context, phones []string, content string, notification *Notification) (*NotificationResult, error) {
	// 构建请求参数
	params := map[string]interface{}{
		"PhoneNumberSet":   phones,
		"SmsSdkAppId":      s.smsConfig.AccessKey,
		"SignName":         s.smsConfig.SignName,
		"TemplateId":       notification.TemplateID,
	}

	if notification.TemplateData != nil {
		// 转换模板参数为腾讯云格式
		templateParams := make([]string, 0)
		for _, v := range notification.TemplateData {
			templateParams = append(templateParams, fmt.Sprintf("%v", v))
		}
		params["TemplateParamSet"] = templateParams
	}

	// 发送请求
	resp, err := s.sendTencentRequest(ctx, "SendSms", params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var tencentResp TencentSMSResponse
	if err := json.Unmarshal(resp, &tencentResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	now := time.Now()
	if tencentResp.Response != nil && tencentResp.Response.SendStatusSet != nil {
		// 检查所有发送状态
		for _, status := range tencentResp.Response.SendStatusSet {
			if status.Code != "Ok" {
				return nil, fmt.Errorf("sms send failed for %s: %s", status.PhoneNumber, status.Message)
			}
		}

		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        true,
			Status:         StatusSent,
			Message:        "sms sent successfully",
			ExternalID:     tencentResp.Response.RequestId,
			DeliveredAt:    &now,
		}, nil
	}

	return nil, fmt.Errorf("invalid response from tencent cloud")
}

// TencentSMSResponse 腾讯云短信响应
type TencentSMSResponse struct {
	Response *struct {
		SendStatusSet []*struct {
			SerialNo       string `json:"SerialNo"`
			PhoneNumber    string `json:"PhoneNumber"`
			Fee            int    `json:"Fee"`
			SessionContext string `json:"SessionContext"`
			Code           string `json:"Code"`
			Message        string `json:"Message"`
			IsoCode        string `json:"IsoCode"`
		} `json:"SendStatusSet"`
		RequestId string `json:"RequestId"`
	} `json:"Response"`
}

// sendTencentRequest 发送腾讯云请求
func (s *SMSNotifier) sendTencentRequest(ctx context.Context, action string, params map[string]interface{}) ([]byte, error) {
	// 构建请求体
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	endpoint := s.smsConfig.Endpoint
	if endpoint == "" {
		endpoint = "https://sms.tencentcloudapi.com"
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", "2021-01-11")
	req.Header.Set("X-TC-Region", s.smsConfig.Region)

	// 计算签名并设置Authorization头
	authorization := s.calculateTencentSignature(req, body)
	req.Header.Set("Authorization", authorization)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// calculateTencentSignature 计算腾讯云签名
func (s *SMSNotifier) calculateTencentSignature(req *http.Request, body []byte) string {
	// 简化实现，实际需要按照腾讯云签名规范实现
	timestamp := time.Now().Unix()
	date := time.Now().Format("2006-01-02")

	// 构建签名字符串
	service := "sms"
	host := req.URL.Host
	httpRequestMethod := "POST"
	canonicalUri := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:%s\n", host)
	signedHeaders := "content-type;host"

	hasher := sha256.New()
	hasher.Write(body)
	hashedRequestPayload := fmt.Sprintf("%x", hasher.Sum(nil))

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod, canonicalUri, canonicalQueryString,
		canonicalHeaders, signedHeaders, hashedRequestPayload)

	// 构建待签名字符串
	algorithm := "TC3-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)

	hasher = sha256.New()
	hasher.Write([]byte(canonicalRequest))
	hashedCanonicalRequest := fmt.Sprintf("%x", hasher.Sum(nil))

	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm, timestamp, credentialScope, hashedCanonicalRequest)

	// 计算签名
	secretDate := hmacSha256([]byte("TC3"+s.smsConfig.AccessSecret), date)
	secretService := hmacSha256(secretDate, service)
	secretSigning := hmacSha256(secretService, "tc3_request")
	signature := hmacSha256Hex(secretSigning, stringToSign)

	// 构建Authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, s.smsConfig.AccessKey, credentialScope, signedHeaders, signature)

	return authorization
}

// hmacSha256 HMAC-SHA256
func hmacSha256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// hmacSha256Hex HMAC-SHA256返回十六进制字符串
func hmacSha256Hex(key []byte, data string) string {
	return fmt.Sprintf("%x", hmacSha256(key, data))
}

// isValidPhone 验证手机号格式
func isValidPhone(phone string) bool {
	// 简单验证：中国手机号
	if len(phone) != 11 {
		return false
	}
	if phone[0] != '1' {
		return false
	}
	for _, c := range phone {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// TokenBucketRateLimiter 令牌桶限流器
type TokenBucketRateLimiter struct {
	rate      int
	burst     int
	tokens    map[string]*tokenBucket
}

type tokenBucket struct {
	tokens     int
	lastUpdate time.Time
}

// NewTokenBucketRateLimiter 创建令牌桶限流器
func NewTokenBucketRateLimiter(rate, burst int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		rate:   rate,
		burst:  burst,
		tokens: make(map[string]*tokenBucket),
	}
}

// Allow 是否允许
func (r *TokenBucketRateLimiter) Allow(key string) bool {
	now := time.Now()
	bucket, exists := r.tokens[key]
	if !exists {
		r.tokens[key] = &tokenBucket{
			tokens:     r.burst - 1,
			lastUpdate: now,
		}
		return true
	}

	// 计算新增的令牌
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	newTokens := int(elapsed * float64(r.rate))
	bucket.tokens += newTokens
	if bucket.tokens > r.burst {
		bucket.tokens = r.burst
	}
	bucket.lastUpdate = now

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// Wait 等待直到可以发送
func (r *TokenBucketRateLimiter) Wait(ctx context.Context, key string) error {
	for {
		if r.Allow(key) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second / time.Duration(r.rate)):
			continue
		}
	}
}

// Reset 重置限流器
func (r *TokenBucketRateLimiter) Reset(key string) {
	delete(r.tokens, key)
}
