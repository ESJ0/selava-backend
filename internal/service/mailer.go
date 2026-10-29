package service

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
)

type PasswordMailer interface {
	EnviarRecuperacion(context.Context, string, string, string) error
}

type SMTPMailer struct{ Host, Port, User, Password, From string }

func (m *SMTPMailer) EnviarRecuperacion(_ context.Context, to, nombre, link string) error {
	address := m.Host + ":" + m.Port
	var auth smtp.Auth
	if m.User != "" {
		auth = smtp.PlainAuth("", m.User, m.Password, m.Host)
	}
	subject := "Recuperacion de contrasena SeLava"
	body := fmt.Sprintf("Hola %s,\r\n\r\nUsa este enlace para crear una nueva contrasena:\r\n%s\r\n\r\nEl enlace vence en 30 minutos y solo puede usarse una vez.\r\n", nombre, link)
	message := []byte("From: " + m.From + "\r\nTo: " + to + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	if err := smtp.SendMail(address, auth, m.From, []string{to}, message); err != nil {
		return fmt.Errorf("enviando correo: %w", err)
	}
	return nil
}

type DevelopmentMailer struct{}

func (DevelopmentMailer) EnviarRecuperacion(_ context.Context, to, _, link string) error {
	log.Printf("correo de recuperacion local para %s: %s", to, link)
	return nil
}
