// mailtest envía un correo de recuperación de prueba con la configuración SMTP del .env
// y muestra el error exacto si falla. Uso: go run ./cmd/mailtest tu-correo@gmail.com
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ESJ0/selava-backend/internal/config"
	"github.com/ESJ0/selava-backend/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("uso: go run ./cmd/mailtest destinatario@correo.com")
		os.Exit(2)
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("error cargando .env:", err)
		os.Exit(1)
	}
	if cfg.SMTPHost == "" {
		fmt.Println("SMTP_HOST esta vacio: el backend no enviara correos. Revisa que el .env este en la carpeta desde donde ejecutas el comando.")
		os.Exit(1)
	}
	fmt.Printf("SMTP %s:%s · usuario %q · remitente %q · contrasena de %d caracteres\n", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPFrom, len(cfg.SMTPPassword))
	fmt.Printf("Enlace de recuperacion configurado: %s\n", cfg.PasswordResetURL)
	mailer := &service.SMTPMailer{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Password: cfg.SMTPPassword, From: cfg.SMTPFrom, FromName: cfg.SMTPFromName}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := mailer.EnviarRecuperacion(ctx, os.Args[1], "prueba", cfg.PasswordResetURL+"?token=prueba"); err != nil {
		fmt.Println("FALLO el envio:", err)
		os.Exit(1)
	}
	fmt.Println("Correo de prueba enviado. Revisa la bandeja de entrada (y la carpeta de spam) de", os.Args[1])
}
