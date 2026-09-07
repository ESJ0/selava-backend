package auth

import "golang.org/x/crypto/bcrypt"

// bcrypt.DefaultCost (10) balancea seguridad y velocidad. El login no
// ocurre por cada venta (solo al iniciar turno en caja), así que el costo
// extra de bcrypt aquí no afecta la velocidad del flujo de venta.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
