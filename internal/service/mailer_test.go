package service

import (
	"bufio"
	"context"
	"io"
	"mime/quotedprintable"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeSMTP acepta una sola conversación SMTP sin TLS y devuelve los datos recibidos.
func fakeSMTP(t *testing.T) (string, <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan string, 1)
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		w := func(s string) { _, _ = io.WriteString(conn, s+"\r\n") }
		w("220 fake ESMTP")
		var data strings.Builder
		inData := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			if inData {
				if line == ".\r\n" {
					inData = false
					w("250 OK")
					continue
				}
				data.WriteString(line)
				continue
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				w("250 fake")
			case strings.HasPrefix(cmd, "DATA"):
				inData = true
				w("354 go ahead")
			case strings.HasPrefix(cmd, "QUIT"):
				w("221 bye")
				got <- data.String()
				return
			default:
				w("250 OK")
			}
		}
	}()
	return ln.Addr().String(), got
}

func TestSMTPMailerEnviaCorreoMultiparteConEnlace(t *testing.T) {
	addr, got := fakeSMTP(t)
	host, port, _ := net.SplitHostPort(addr)
	mailer := &SMTPMailer{Host: host, Port: port, From: "selava@example.com", Timeout: 5 * time.Second}
	link := "http://localhost:5174/restablecer-contrasena?token=abc123"
	if err := mailer.EnviarRecuperacion(context.Background(), "cliente@gmail.com", "Ana", link); err != nil {
		t.Fatal(err)
	}
	raw := <-got
	decoded, _ := io.ReadAll(quotedprintable.NewReader(strings.NewReader(raw)))
	body := string(decoded)
	for _, want := range []string{`From: "SELAVA" <selava@example.com>`, "To: <cliente@gmail.com>", "Subject: =?utf-8?q?Recupera_tu_contrase=C3=B1a_de_SELAVA?=", "multipart/alternative", "Hola Ana", link, "Crear nueva contraseña"} {
		if !strings.Contains(body, want) && !strings.Contains(raw, want) {
			t.Fatalf("falta %q en el correo:\n%s", want, raw)
		}
	}
}

func TestSMTPMailerNoEnviaCredencialesSinTLS(t *testing.T) {
	addr, _ := fakeSMTP(t)
	host, port, _ := net.SplitHostPort(addr)
	mailer := &SMTPMailer{Host: host, Port: port, User: "u@example.com", Password: "x", Timeout: 5 * time.Second}
	err := mailer.EnviarRecuperacion(context.Background(), "cliente@gmail.com", "Ana", "http://x")
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("esperaba error por falta de STARTTLS, obtuve %v", err)
	}
}

func TestSMTPMailerUsaUsuarioComoRemitentePorDefecto(t *testing.T) {
	addr, got := fakeSMTP(t)
	host, port, _ := net.SplitHostPort(addr)
	mailer := &SMTPMailer{Host: host, Port: port, From: "", User: "", Timeout: 5 * time.Second}
	if err := mailer.EnviarRecuperacion(context.Background(), "a@b.com", "Ana", "http://x"); err == nil {
		t.Fatal("sin remitente debe fallar")
	}
	mailer.From = "selava.gt@gmail.com"
	if err := mailer.EnviarRecuperacion(context.Background(), "a@b.com", "Ana", "http://x"); err != nil {
		t.Fatal(err)
	}
	if raw := <-got; !strings.Contains(raw, "selava.gt@gmail.com") {
		t.Fatalf("remitente no aplicado: %s", raw)
	}
}
