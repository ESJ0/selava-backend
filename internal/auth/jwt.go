package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenInvalido = errors.New("token inválido o expirado")

const tokenTTL = 8 * time.Hour // duración típica de un turno de caja

type Claims struct {
	UsuarioID int `json:"usuario_id"`
	RolID     int `json:"rol_id"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, usuarioID, rolID int) (string, error) {
	claims := Claims{
		UsuarioID: usuarioID,
		RolID:     rolID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalido
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrTokenInvalido
	}
	return claims, nil
}
