package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

type PasswordMailer interface {
	EnviarRecuperacion(context.Context, string, string, string) error
}

// SMTPMailer envía los correos de recuperación. Soporta TLS implícito (puerto 465)
// y STARTTLS (puerto 587, p. ej. Gmail con contraseña de aplicación).
type SMTPMailer struct {
	Host, Port, User, Password, From, FromName string
	// Timeout total de la conversación SMTP; 20 s si no se define.
	Timeout time.Duration
	// Solo para pruebas: permite STARTTLS contra servidores con certificados locales.
	InsecureSkipVerify bool
}

func (m *SMTPMailer) EnviarRecuperacion(ctx context.Context, to, nombre, link string) error {
	from := strings.TrimSpace(m.From)
	if from == "" {
		from = strings.TrimSpace(m.User)
	}
	if _, err := mail.ParseAddress(from); err != nil {
		return fmt.Errorf("remitente SMTP invalido %q: %w", from, err)
	}
	message, err := buildRecoveryMessage(from, m.FromName, to, nombre, link, time.Now())
	if err != nil {
		return err
	}
	return m.send(ctx, from, to, message)
}

func (m *SMTPMailer) send(ctx context.Context, from, to string, message []byte) error {
	timeout := m.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	address := net.JoinHostPort(m.Host, m.Port)
	dialer := &net.Dialer{Timeout: timeout}
	tlsConfig := &tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12, InsecureSkipVerify: m.InsecureSkipVerify} // #nosec G402 -- solo pruebas

	var conn net.Conn
	var err error
	if m.Port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return fmt.Errorf("conectando con SMTP %s: %w", address, err)
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))
	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("iniciando sesion SMTP: %w", err)
	}
	defer client.Close()

	if m.Port != "465" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("activando STARTTLS: %w", err)
			}
		} else if m.User != "" {
			return fmt.Errorf("el servidor SMTP %s no ofrece STARTTLS; no se enviaran credenciales sin cifrar", address)
		}
	}
	if m.User != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("el servidor SMTP %s no acepta autenticacion", address)
		}
		if err = client.Auth(smtp.PlainAuth("", m.User, m.Password, m.Host)); err != nil {
			return fmt.Errorf("autenticando en SMTP (revisa SMTP_USER/SMTP_PASSWORD; en Gmail usa una contrasena de aplicacion): %w", err)
		}
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("remitente rechazado: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("destinatario rechazado: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("abriendo cuerpo del correo: %w", err)
	}
	if _, err = writer.Write(message); err != nil {
		return fmt.Errorf("escribiendo correo: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("enviando correo: %w", err)
	}
	return client.Quit()
}

// buildRecoveryMessage arma un correo multiparte (texto + HTML) con cabeceras completas
// para que los proveedores como Gmail no lo marquen como sospechoso.
func buildRecoveryMessage(from, fromName, to, nombre, link string, now time.Time) ([]byte, error) {
	if fromName == "" {
		fromName = "SELAVA"
	}
	if strings.TrimSpace(nombre) == "" {
		nombre = "cliente"
	}
	boundary, err := randomHex(12)
	if err != nil {
		return nil, err
	}
	messageID, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	domain := "selava.local"
	if at := strings.LastIndex(from, "@"); at >= 0 {
		domain = from[at+1:]
	}
	subject := mime.QEncoding.Encode("utf-8", "Recupera tu contraseña de SELAVA")
	text := fmt.Sprintf("Hola %s,\n\nRecibimos una solicitud para restablecer la contraseña de tu cuenta SELAVA.\n\nCrea una nueva contraseña aquí:\n%s\n\nEl enlace vence en 30 minutos y solo puede usarse una vez.\nSi no fuiste tú, puedes ignorar este correo; tu contraseña no cambiará.\n\nSELAVA · Lavandería y Dry Clean\n", nombre, link)
	safeName, safeLink := html.EscapeString(nombre), html.EscapeString(link)
	htmlBody := `<!doctype html><html lang="es"><body style="margin:0;background:#F4F8FA;font-family:Arial,Helvetica,sans-serif;color:#0A2433">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#F4F8FA;padding:32px 12px"><tr><td align="center">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#ffffff;border-radius:20px;overflow:hidden;border:1px solid #DCE7EC">
<tr><td style="background:#0A2433;padding:24px 28px;color:#ffffff;font-size:22px;font-weight:bold">SE<span style="color:#5BC3EC">LAVA</span></td></tr>
<tr><td style="padding:28px">
<p style="margin:0 0 12px;font-size:20px;font-weight:bold">Hola ` + safeName + `,</p>
<p style="margin:0 0 20px;font-size:15px;line-height:1.6;color:#3E5866">Recibimos una solicitud para restablecer la contraseña de tu cuenta SELAVA. Pulsa el botón para crear una nueva.</p>
<p style="margin:0 0 24px"><a href="` + safeLink + `" style="display:inline-block;background:#FFC93C;color:#0A2433;text-decoration:none;font-weight:bold;padding:14px 22px;border-radius:12px">Crear nueva contraseña</a></p>
<p style="margin:0 0 8px;font-size:13px;color:#5A7280">El enlace vence en 30 minutos y solo puede usarse una vez.</p>
<p style="margin:0 0 8px;font-size:13px;color:#5A7280">Si el botón no funciona, copia este enlace en tu navegador:<br><a href="` + safeLink + `" style="color:#06709F;word-break:break-all">` + safeLink + `</a></p>
<p style="margin:16px 0 0;font-size:13px;color:#5A7280">Si no fuiste tú, ignora este correo; tu contraseña no cambiará.</p>
</td></tr></table></td></tr></table></body></html>`

	var buf bytes.Buffer
	header := func(key, value string) { fmt.Fprintf(&buf, "%s: %s\r\n", key, value) }
	header("From", (&mail.Address{Name: fromName, Address: from}).String())
	header("To", (&mail.Address{Address: to}).String())
	header("Subject", subject)
	header("Date", now.Format(time.RFC1123Z))
	header("Message-ID", "<"+messageID+"@"+domain+">")
	header("MIME-Version", "1.0")
	header("Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	buf.WriteString("\r\n")
	for _, part := range []struct{ kind, body string }{{"text/plain", text}, {"text/html", htmlBody}} {
		fmt.Fprintf(&buf, "--%s\r\nContent-Type: %s; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", boundary, part.kind)
		qp := quotedprintable.NewWriter(&buf)
		if _, err := qp.Write([]byte(strings.ReplaceAll(part.body, "\n", "\r\n"))); err != nil {
			return nil, err
		}
		if err := qp.Close(); err != nil {
			return nil, err
		}
		buf.WriteString("\r\n")
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	return buf.Bytes(), nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generando identificador de correo: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// DevelopmentMailer no envía correos: solo deja el enlace en el log del servidor.
type DevelopmentMailer struct{}

func (DevelopmentMailer) EnviarRecuperacion(_ context.Context, to, _, link string) error {
	log.Printf("SMTP no configurado: NO se envio correo a %s. Enlace de recuperacion: %s", to, link)
	return nil
}
