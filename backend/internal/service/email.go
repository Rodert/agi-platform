package service

import (
	"crypto/tls"
	"fmt"

	"github.com/javapub/agi-platform-backend/internal/repository"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	configRepo *repository.ConfigRepository
}

func NewEmailService(configRepo *repository.ConfigRepository) *EmailService {
	return &EmailService{
		configRepo: configRepo,
	}
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(email, code string) error {
	// 从数据库读取邮箱配置
	emailConfig, err := s.configRepo.GetEmailConfig()
	if err != nil {
		return fmt.Errorf("获取邮箱配置失败: %w", err)
	}

	if !emailConfig.IsActive {
		return fmt.Errorf("邮箱服务未启用")
	}

	// 构建邮件内容
	subject := "【潮汐AI】验证码"
	body := fmt.Sprintf(`
		<div style="padding: 20px; font-family: Arial, sans-serif;">
			<h2 style="color: #54d6cf;">潮汐AI</h2>
			<p>您的验证码是：</p>
			<h1 style="color: #54d6cf; letter-spacing: 5px;">%s</h1>
			<p style="color: #666;">验证码将在 5 分钟内有效，请及时使用。</p>
			<p style="color: #999; font-size: 12px;">如果这不是您的操作，请忽略此邮件。</p>
		</div>
	`, code)

	// 创建邮件
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// 发送邮件
	d := gomail.NewDialer(
		emailConfig.SMTPHost,
		emailConfig.SMTPPort,
		emailConfig.SMTPUser,
		emailConfig.SMTPPassword,
	)

	// 根据配置决定是否使用SSL
	if emailConfig.SMTPSSL {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}
