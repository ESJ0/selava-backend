package config

import "testing"

func TestLoadQuitaEspaciosDeContrasenaDeAplicacionGmail(t *testing.T) {
	for key, value := range map[string]string{"DB_HOST": "h", "DB_USER": "u", "DB_NAME": "n", "JWT_SECRET": "s", "SMTP_HOST": " smtp.gmail.com ", "SMTP_PASSWORD": "abcd efgh ijkl mnop", "SMTP_USER": " cuenta@gmail.com "} {
		t.Setenv(key, value)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SMTPPassword != "abcdefghijklmnop" || cfg.SMTPHost != "smtp.gmail.com" || cfg.SMTPUser != "cuenta@gmail.com" {
		t.Fatalf("configuracion SMTP no normalizada: %+v", cfg)
	}
}
