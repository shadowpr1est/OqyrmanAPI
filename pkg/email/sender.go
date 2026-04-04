package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

const dialTimeout = 10 * time.Second

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

func (s *Sender) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	// Порт 587 использует STARTTLS — апгрейдим соединение до TLS если сервер поддерживает
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: s.host}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	)
	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}

// SendVerificationCode отправляет 6-значный код подтверждения на указанный email.
func (s *Sender) SendVerificationCode(to, code string) error {
	subject := "Подтверждение email — Oqyrman"
	body := fmt.Sprintf(
		"Добро пожаловать в Oqyrman!\r\n\r\nВаш код подтверждения: %s\r\n\r\nКод действителен 3 минуты.\r\nЕсли вы не регистрировались — проигнорируйте это письмо.",
		code,
	)
	return s.send(to, subject, body)
}

// SendPasswordResetCode отправляет 6-значный код сброса пароля.
func (s *Sender) SendPasswordResetCode(to, code string) error {
	subject := "Сброс пароля — Oqyrman"
	body := fmt.Sprintf(
		"Вы запросили сброс пароля.\r\n\r\nВаш код: %s\r\n\r\nКод действителен 3 минуты.\r\nЕсли вы не запрашивали сброс — проигнорируйте это письмо.",
		code,
	)
	return s.send(to, subject, body)
}
