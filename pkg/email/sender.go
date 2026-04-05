package email

import (
	"fmt"

	resend "github.com/resend/resend-go/v2"
)

// Sender отправляет письма через Resend API.
type Sender struct {
	client  *resend.Client
	from    string
	logoURL string // публичный URL логотипа (PNG), опционален
}

func NewSender(apiKey, from, logoURL string) *Sender {
	return &Sender{
		client:  resend.NewClient(apiKey),
		from:    from,
		logoURL: logoURL,
	}
}

// Enabled возвращает true, если Resend настроен (есть ключ и from).
func (s *Sender) Enabled() bool {
	return s.client != nil && s.from != ""
}

func (s *Sender) send(to, subject, html, text string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	}
	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}
	return nil
}

// SendVerificationCode отправляет 6-значный код подтверждения на указанный email.
func (s *Sender) SendVerificationCode(to, code string) error {
	subject := "Подтверждение email — Oqyrman"
	html := s.verificationHTML(code)
	text := fmt.Sprintf(
		"Добро пожаловать в Oqyrman!\n\nВаш код подтверждения: %s\n\nКод действителен 3 минуты.\nЕсли вы не регистрировались — проигнорируйте это письмо.",
		code,
	)
	return s.send(to, subject, html, text)
}

// SendPasswordResetCode отправляет 6-значный код сброса пароля.
func (s *Sender) SendPasswordResetCode(to, code string) error {
	subject := "Сброс пароля — Oqyrman"
	html := s.resetHTML(code)
	text := fmt.Sprintf(
		"Вы запросили сброс пароля.\n\nВаш код: %s\n\nКод действителен 5 минут.\nЕсли вы не запрашивали сброс — проигнорируйте это письмо.",
		code,
	)
	return s.send(to, subject, html, text)
}

// ── HTML templates ────────────────────────────────────────────────────────────

func (s *Sender) verificationHTML(code string) string {
	return s.buildEmail(
		"Подтверждение email",
		"Добро пожаловать!",
		"Для завершения регистрации введите код подтверждения:",
		code,
		"Код действителен <strong>3 минуты</strong>.",
		"Если вы не регистрировались в Oqyrman — просто проигнорируйте это письмо.",
	)
}

func (s *Sender) resetHTML(code string) string {
	return s.buildEmail(
		"Сброс пароля",
		"Запрос на сброс пароля",
		"Мы получили запрос на смену пароля. Используйте код ниже:",
		code,
		"Код действителен <strong>5 минут</strong>.",
		"Если вы не запрашивали сброс пароля — проигнорируйте это письмо. Ваш аккаунт в безопасности.",
	)
}

// headerContent возвращает содержимое хедера: логотип-картинку или текстовый fallback.
func (s *Sender) headerContent() string {
	if s.logoURL != "" {
		return fmt.Sprintf(
			`<img src="%s" alt="Oqyrman" style="height:44px;display:block;margin:0 auto 8px;">
              <span style="display:block;color:#A8D5C2;font-size:15px;letter-spacing:0.5px;">Oqyrman</span>`,
			s.logoURL,
		)
	}
	return `<span style="color:#FFFFFF;font-size:24px;font-weight:700;letter-spacing:0.5px;">Oqyrman</span>`
}

func (s *Sender) buildEmail(title, heading, description, code, expiry, disclaimer string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background-color:#F2F7F5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#F2F7F5;padding:48px 16px;">
    <tr>
      <td align="center">

        <!-- Card -->
        <table width="100%%" cellpadding="0" cellspacing="0" role="presentation"
               style="max-width:480px;background:#FFFFFF;border-radius:16px;overflow:hidden;box-shadow:0 2px 12px rgba(30,89,69,0.10);">

          <!-- Header -->
          <tr>
            <td style="background:#1E5945;padding:30px 40px;text-align:center;">
              %s
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding:40px 40px 32px;">

              <h1 style="margin:0 0 12px;font-size:20px;font-weight:600;color:#1E5945;">%s</h1>
              <p style="margin:0 0 28px;font-size:15px;color:#4A5568;line-height:1.6;">%s</p>

              <!-- Code block -->
              <table width="100%%" cellpadding="0" cellspacing="0" role="presentation"
                     style="background:#EDF7F2;border:1px solid #C6E8D8;border-radius:12px;margin-bottom:28px;">
                <tr>
                  <td style="padding:28px 16px;text-align:center;">
                    <span style="font-size:44px;font-weight:700;letter-spacing:14px;color:#1E5945;font-family:'Courier New',Courier,monospace;">%s</span>
                  </td>
                </tr>
              </table>

              <p style="margin:0 0 10px;font-size:14px;color:#4A5568;">%s</p>
              <p style="margin:0;font-size:13px;color:#9CA3AF;line-height:1.5;">%s</p>

            </td>
          </tr>

          <!-- Divider -->
          <tr>
            <td style="padding:0 40px;">
              <div style="height:1px;background:#E8F2EE;"></div>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:20px 40px;text-align:center;">
              <p style="margin:0;font-size:12px;color:#9CA3AF;">© 2026 Oqyrman · Не отвечайте на это письмо</p>
            </td>
          </tr>

        </table>
        <!-- /Card -->

      </td>
    </tr>
  </table>
</body>
</html>`, title, s.headerContent(), heading, description, code, expiry, disclaimer)
}
