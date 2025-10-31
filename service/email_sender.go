package service

import (
	"fmt"
	"net/smtp"

	"github.com/ekjyotshinh/email-service/config"
	"github.com/ekjyotshinh/email-service/model"
)

func SendEmail(e model.Email) error {
	auth := smtp.PlainAuth("", config.AppConfig.SMTPUsername, config.AppConfig.SMTPPassword, config.AppConfig.SMTPHost)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", e.To, e.Subject, e.Body))
	addr := fmt.Sprintf("%s:%d", config.AppConfig.SMTPHost, config.AppConfig.SMTPPort)

	err := smtp.SendMail(addr, auth, config.AppConfig.SMTPUsername, []string{e.To}, msg)
	if err != nil {
		return err
	}
	return nil
}
