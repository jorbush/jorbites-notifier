package email

import (
	"bytes"
	"fmt"
	"net/smtp"

	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/models"
)

type EmailSender struct {
	config *config.Config
}

func NewEmailSender(cfg *config.Config) *EmailSender {
	return &EmailSender{
		config: cfg,
	}
}

func (s *EmailSender) SendNotificationEmail(notification models.Notification, language string) (bool, error) {
	if notification.Recipient == "" {
		return false, fmt.Errorf("no recipient specified")
	}

	if s.config.SMTPUser == "" || s.config.SMTPPassword == "" {
		return false, fmt.Errorf("SMTP credentials not configured")
	}

	subject, body, err := GetEmailTemplate(notification.Type, notification.Metadata, language)
	if err != nil {
		return false, fmt.Errorf("error preparing email template: %w", err)
	}

	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)

	message := bytes.NewBuffer(nil)
	message.WriteString(fmt.Sprintf("From: Jorbites <%s>\r\n", s.config.SMTPUser))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString(fmt.Sprintf("To: %s\r\n", notification.Recipient))
	message.WriteString("MIME-version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	message.WriteString(body)

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	err = smtp.SendMail(
		addr,
		auth,
		s.config.SMTPUser,
		[]string{notification.Recipient},
		message.Bytes(),
	)

	if err != nil {
		return false, fmt.Errorf("failed to send email: %w", err)
	}

	return true, nil
}
