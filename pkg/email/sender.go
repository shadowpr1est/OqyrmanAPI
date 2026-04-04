package email

import (
	"fmt"
	"net/smtp"
)

// Sender отправляет письма через SMTP.
type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSender(host string, port int, username, password, from string) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Enabled возвращает true, если SMTP настроен (есть хост и from).
func (s *Sender) Enabled() bool {
	return s.host != "" && s.from != ""
}

// SendPasswordResetCode отправляет 6-значный код сброса пароля.
func (s *Sender) SendPasswordResetCode(to, code string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	subject := "Сброс пароля — Oqyrman"
	body := fmt.Sprintf(
		"Вы запросили сброс пароля.\r\n\r\nВаш код: %s\r\n\r\nКод действителен 15 минут.\r\nЕсли вы не запрашивали сброс — проигнорируйте это письмо.",
		code,
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}

// SendVerificationCode отправляет 6-значный код подтверждения на указанный email.
func (s *Sender) SendVerificationCode(to, code string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	subject := "Подтверждение email — Oqyrman"
	body := fmt.Sprintf(
		"Добро пожаловать в Oqyrman!\r\n\r\nВаш код подтверждения: %s\r\n\r\nКод действителен 15 минут.\r\nЕсли вы не регистрировались — проигнорируйте это письмо.",
		code,
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
