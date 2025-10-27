package service
import (
	"fmt"
	"net/smtp"
	"os"
	"github.com/ekjyotshinh/email-service/model"
)
func SendEmail(e model.Email) error {
	auth := smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST"))
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", e.To, e.Subject, e.Body))
	addr := fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))

	err := smtp.SendMail(addr, auth, os.Getenv("SMTP_USERNAME"), []string{e.To}, msg)
	if err != nil {
		return err
	}
	return nil
}
