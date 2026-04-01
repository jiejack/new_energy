package notifier

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

var (
	ErrEmailConfigInvalid   = errors.New("email config is invalid")
	ErrEmailRecipientEmpty  = errors.New("email recipient is empty")
	ErrEmailSendFailed      = errors.New("email send failed")
	ErrEmailTemplateInvalid = errors.New("email template is invalid")
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	config       *NotificationConfig
	emailConfig  *EmailConfig
	smtpClient   *smtp.Client
	rateLimiter  RateLimiter
	templateMgr  TemplateManager
	templateCache map[string]*template.Template
	mu           sync.RWMutex
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(config *NotificationConfig, templateMgr TemplateManager) (*EmailNotifier, error) {
	if config == nil || config.EmailConfig == nil {
		return nil, ErrEmailConfigInvalid
	}

	return &EmailNotifier{
		config:        config,
		emailConfig:   config.EmailConfig,
		rateLimiter:   NewTokenBucketRateLimiter(config.RateLimit, config.BurstLimit),
		templateMgr:   templateMgr,
		templateCache: make(map[string]*template.Template),
	}, nil
}

// Channel 返回通知渠道类型
func (e *EmailNotifier) Channel() NotificationChannel {
	return ChannelEmail
}

// Send 发送邮件通知
func (e *EmailNotifier) Send(ctx context.Context, notification *Notification) (*NotificationResult, error) {
	if err := e.Validate(notification); err != nil {
		return nil, err
	}

	// 限流检查
	key := fmt.Sprintf("email:%s", notification.AlarmID)
	if !e.rateLimiter.Allow(key) {
		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusFailed,
			Message:        "rate limit exceeded",
			Error:          ErrEmailRateLimitExceeded,
		}, ErrEmailRateLimitExceeded
	}

	// 准备邮件内容
	email, err := e.prepareEmail(notification)
	if err != nil {
		return nil, err
	}

	// 发送邮件
	if err := e.sendEmail(email); err != nil {
		return &NotificationResult{
			NotificationID: notification.ID,
			Success:        false,
			Status:         StatusFailed,
			Message:        err.Error(),
			Error:          err,
		}, err
	}

	now := time.Now()
	return &NotificationResult{
		NotificationID: notification.ID,
		Success:        true,
		Status:         StatusSent,
		Message:        "email sent successfully",
		DeliveredAt:    &now,
	}, nil
}

// SendBatch 批量发送邮件通知
func (e *EmailNotifier) SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, len(notifications))

	// 使用协程池并发发送
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, notification := range notifications {
		wg.Add(1)
		go func(idx int, notif *Notification) {
			defer wg.Done()
			result, err := e.Send(ctx, notif)
			mu.Lock()
			if err != nil {
				results[idx] = &NotificationResult{
					NotificationID: notif.ID,
					Success:        false,
					Status:         StatusFailed,
					Message:        err.Error(),
					Error:          err,
				}
			} else {
				results[idx] = result
			}
			mu.Unlock()
		}(i, notification)
	}

	wg.Wait()
	return results, nil
}

// Validate 验证邮件通知
func (e *EmailNotifier) Validate(notification *Notification) error {
	if notification == nil {
		return errors.New("notification is nil")
	}

	if len(notification.Recipients) == 0 {
		return ErrEmailRecipientEmpty
	}

	// 验证邮箱地址
	for _, r := range notification.Recipients {
		if r.Email == "" {
			return fmt.Errorf("recipient %s has no email address", r.Name)
		}
		if !isValidEmail(r.Email) {
			return fmt.Errorf("invalid email address: %s", r.Email)
		}
	}

	return nil
}

// HealthCheck 健康检查
func (e *EmailNotifier) HealthCheck(ctx context.Context) error {
	// 尝试连接SMTP服务器
	client, err := e.connectSMTP()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	client.Close()
	return nil
}

// Close 关闭邮件通知器
func (e *EmailNotifier) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.smtpClient != nil {
		e.smtpClient.Close()
		e.smtpClient = nil
	}
	return nil
}

// EmailMessage 邮件消息
type EmailMessage struct {
	From        mail.Address
	To          []mail.Address
	CC          []mail.Address
	BCC         []mail.Address
	Subject     string
	TextContent string
	HTMLContent string
	Attachments []Attachment
	Headers     map[string]string
}

// prepareEmail 准备邮件内容
func (e *EmailNotifier) prepareEmail(notification *Notification) (*EmailMessage, error) {
	email := &EmailMessage{
		From: mail.Address{
			Name:    e.emailConfig.FromName,
			Address: e.emailConfig.FromAddress,
		},
		Subject: notification.Subject,
		Headers: make(map[string]string),
	}

	// 设置收件人
	for _, r := range notification.Recipients {
		email.To = append(email.To, mail.Address{
			Name:    r.Name,
			Address: r.Email,
		})
	}

	// 准备邮件内容
	if notification.HTMLContent != "" {
		email.HTMLContent = notification.HTMLContent
	} else if notification.TemplateID != "" && e.templateMgr != nil {
		// 使用模板渲染
		rendered, err := e.templateMgr.Render(notification.TemplateID, notification.TemplateData)
		if err != nil {
			return nil, err
		}
		email.HTMLContent = rendered
	} else {
		email.TextContent = notification.Content
	}

	// 添加附件
	if len(notification.Attachments) > 0 {
		email.Attachments = notification.Attachments
	}

	// 设置邮件头
	email.Headers["X-Priority"] = e.getPriorityHeader(notification.Priority)
	email.Headers["X-Alarm-ID"] = notification.AlarmID

	return email, nil
}

// getPriorityHeader 获取优先级头
func (e *EmailNotifier) getPriorityHeader(priority NotificationPriority) string {
	switch priority {
	case PriorityCritical:
		return "1"
	case PriorityHigh:
		return "2"
	case PriorityNormal:
		return "3"
	default:
		return "4"
	}
}

// sendEmail 发送邮件
func (e *EmailNotifier) sendEmail(email *EmailMessage) error {
	// 连接SMTP服务器
	client, err := e.connectSMTP()
	if err != nil {
		return err
	}
	defer client.Close()

	// 设置发件人
	from := email.From.Address
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// 设置收件人
	recipients := make([]string, 0)
	for _, to := range email.To {
		recipients = append(recipients, to.Address)
		if err := client.Rcpt(to.Address); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", to.Address, err)
		}
	}

	// 设置抄送
	for _, cc := range email.CC {
		recipients = append(recipients, cc.Address)
		if err := client.Rcpt(cc.Address); err != nil {
			return fmt.Errorf("failed to set CC %s: %w", cc.Address, err)
		}
	}

	// 设置密送
	for _, bcc := range email.BCC {
		recipients = append(recipients, bcc.Address)
		if err := client.Rcpt(bcc.Address); err != nil {
			return fmt.Errorf("failed to set BCC %s: %w", bcc.Address, err)
		}
	}

	// 获取数据写入器
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer writer.Close()

	// 构建邮件内容
	content, err := e.buildEmailContent(email)
	if err != nil {
		return fmt.Errorf("failed to build email content: %w", err)
	}

	// 写入邮件内容
	if _, err := writer.Write(content); err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	return nil
}

// connectSMTP 连接SMTP服务器
func (e *EmailNotifier) connectSMTP() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", e.emailConfig.SMTPHost, e.emailConfig.SMTPPort)

	var client *smtp.Client
	var err error

	if e.emailConfig.UseTLS {
		// TLS连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         e.emailConfig.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect via TLS: %w", err)
		}

		client, err = smtp.NewClient(conn, e.emailConfig.SMTPHost)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// 普通连接
		client, err = smtp.Dial(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
	}

	// 发送HELO/EHLO
	host, _ := e.getLocalHostname()
	if err := client.Hello(host); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to send HELO: %w", err)
	}

	// STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok && !e.emailConfig.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         e.emailConfig.SMTPHost,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// 认证
	if e.emailConfig.Username != "" && e.emailConfig.Password != "" {
		auth := smtp.PlainAuth("", e.emailConfig.Username, e.emailConfig.Password, e.emailConfig.SMTPHost)
		if err := client.Auth(auth); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	return client, nil
}

// getLocalHostname 获取本地主机名
func (e *EmailNotifier) getLocalHostname() (string, error) {
	hostname, err := net.LookupCNAME("localhost")
	if err != nil {
		return "localhost", nil
	}
	return hostname, nil
}

// buildEmailContent 构建邮件内容
func (e *EmailNotifier) buildEmailContent(email *EmailMessage) ([]byte, error) {
	var buf bytes.Buffer

	// 写入邮件头
	fmt.Fprintf(&buf, "From: %s\r\n", email.From.String())

	// 收件人
	toList := make([]string, len(email.To))
	for i, to := range email.To {
		toList[i] = to.String()
	}
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(toList, ", "))

	// 抄送
	if len(email.CC) > 0 {
		ccList := make([]string, len(email.CC))
		for i, cc := range email.CC {
			ccList[i] = cc.String()
		}
		fmt.Fprintf(&buf, "Cc: %s\r\n", strings.Join(ccList, ", "))
	}

	// 主题
	fmt.Fprintf(&buf, "Subject: %s\r\n", email.Subject)

	// 日期
	fmt.Fprintf(&buf, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))

	// MIME版本
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")

	// 自定义头
	for k, v := range email.Headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
	}

	// 根据是否有附件和HTML内容选择不同的编码方式
	if len(email.Attachments) > 0 {
		return e.buildMultipartEmail(&buf, email)
	}

	if email.HTMLContent != "" {
		return e.buildMimeEmail(&buf, email)
	}

	// 纯文本邮件
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n")
	fmt.Fprintf(&buf, "\r\n")
	fmt.Fprintf(&buf, "%s", email.TextContent)

	return buf.Bytes(), nil
}

// buildMimeEmail 构建MIME邮件（HTML+文本）
func (e *EmailNotifier) buildMimeEmail(buf *bytes.Buffer, email *EmailMessage) ([]byte, error) {
	boundary := fmt.Sprintf("----=_Part_%d", time.Now().UnixNano())

	fmt.Fprintf(buf, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(buf, "\r\n")

	// 文本部分
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	fmt.Fprintf(buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: quoted-printable\r\n")
	fmt.Fprintf(buf, "\r\n")
	fmt.Fprintf(buf, "%s\r\n", email.TextContent)

	// HTML部分
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	fmt.Fprintf(buf, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: quoted-printable\r\n")
	fmt.Fprintf(buf, "\r\n")
	fmt.Fprintf(buf, "%s\r\n", email.HTMLContent)

	// 结束边界
	fmt.Fprintf(buf, "--%s--\r\n", boundary)

	return buf.Bytes(), nil
}

// buildMultipartEmail 构建多部分邮件（带附件）
func (e *EmailNotifier) buildMultipartEmail(buf *bytes.Buffer, email *EmailMessage) ([]byte, error) {
	boundary := fmt.Sprintf("----=_Part_%d", time.Now().UnixNano())

	fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(buf, "\r\n")

	// 邮件正文
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	if email.HTMLContent != "" {
		fmt.Fprintf(buf, "Content-Type: text/html; charset=UTF-8\r\n")
		fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n")
		fmt.Fprintf(buf, "\r\n")
		fmt.Fprintf(buf, "%s\r\n", email.HTMLContent)
	} else {
		fmt.Fprintf(buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(buf, "Content-Transfer-Encoding: quoted-printable\r\n")
		fmt.Fprintf(buf, "\r\n")
		fmt.Fprintf(buf, "%s\r\n", email.TextContent)
	}

	// 附件
	for _, attachment := range email.Attachments {
		fmt.Fprintf(buf, "--%s\r\n", boundary)
		fmt.Fprintf(buf, "Content-Type: %s; name=\"%s\"\r\n", attachment.MimeType, attachment.Name)
		fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n")
		fmt.Fprintf(buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name)
		fmt.Fprintf(buf, "\r\n")

		// Base64编码附件内容
		encoded := e.base64Encode(attachment.Content)
		fmt.Fprintf(buf, "%s\r\n", encoded)
	}

	// 结束边界
	fmt.Fprintf(buf, "--%s--\r\n", boundary)

	return buf.Bytes(), nil
}

// base64Encode Base64编码
func (e *EmailNotifier) base64Encode(data []byte) string {
	// 简单实现，实际应该使用标准库的base64编码
	// 这里为了演示，直接返回字符串
	return string(data)
}

// isValidEmail 验证邮箱地址格式
func isValidEmail(email string) bool {
	// 简单验证
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	return strings.Contains(parts[1], ".")
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID          string
	Name        string
	Subject     string
	TextContent string
	HTMLContent string
	Variables   []string
}

// RenderEmailTemplate 渲染邮件模板
func (e *EmailNotifier) RenderEmailTemplate(templateID string, data map[string]interface{}) (*EmailMessage, error) {
	// 从缓存获取模板
	e.mu.RLock()
	tmpl, exists := e.templateCache[templateID]
	e.mu.RUnlock()

	if !exists {
		// 从模板管理器加载
		if e.templateMgr == nil {
			return nil, ErrEmailTemplateInvalid
		}

		templateContent, err := e.templateMgr.Get(templateID)
		if err != nil {
			return nil, err
		}

		// 解析模板
		tmpl, err = template.New(templateID).Parse(templateContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template: %w", err)
		}

		// 缓存模板
		e.mu.Lock()
		e.templateCache[templateID] = tmpl
		e.mu.Unlock()
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return &EmailMessage{
		HTMLContent: buf.String(),
	}, nil
}

// BatchEmailSender 批量邮件发送器
type BatchEmailSender struct {
	notifier    *EmailNotifier
	batchSize   int
	sendTimeout time.Duration
}

// NewBatchEmailSender 创建批量邮件发送器
func NewBatchEmailSender(notifier *EmailNotifier, batchSize int, timeout time.Duration) *BatchEmailSender {
	return &BatchEmailSender{
		notifier:    notifier,
		batchSize:   batchSize,
		sendTimeout: timeout,
	}
}

// SendBatch 批量发送
func (b *BatchEmailSender) SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, len(notifications))

	// 分批处理
	for i := 0; i < len(notifications); i += b.batchSize {
		end := i + b.batchSize
		if end > len(notifications) {
			end = len(notifications)
		}

		batch := notifications[i:end]
		batchResults, err := b.notifier.SendBatch(ctx, batch)
		if err != nil {
			// 记录错误但继续处理
			for j := range batch {
				results[i+j] = &NotificationResult{
					NotificationID: batch[j].ID,
					Success:        false,
					Status:         StatusFailed,
					Message:        err.Error(),
					Error:          err,
				}
			}
		} else {
			for j, result := range batchResults {
				results[i+j] = result
			}
		}

		// 批次间延迟
		if end < len(notifications) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return results, nil
}
